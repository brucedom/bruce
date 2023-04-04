package loader

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var s = &S3Manager{}

type S3Manager struct {
	Session *session.Session
	Timeout time.Duration
	Service s3iface.S3API
}

func ReaderFromS3(fileName string) (io.ReadCloser, string, error) {
	fn := path.Base(fileName)
	if s == nil {
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
		}

		sess, err := session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(region), Endpoint: aws.String(fmt.Sprintf("s3.%s.amazonaws.com", region))}})
		if err != nil {
			log.Debug().Msg("Could not initialize AWS session")
			return nil, fn, err
		}
		s.Service = s3.New(sess)
	}
	ctx := context.Background()
	var cancelFn func()
	if s.Timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, s.Timeout)
	}
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	defer cancelFn()
	pfxCut := fileName[5:]
	subIdx := strings.Index(pfxCut, "/")
	bucket := pfxCut[:subIdx]
	objKey := pfxCut[subIdx:]
	fd, err := s.Service.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objKey),
	})
	if err != nil {
		log.Debug().Err(err).Msg("fetching object from S3 failed")
		return nil, fn, err
	}
	return fd.Body, fn, nil
}

func downloadS3File(svc *s3.S3, bucket, key, aDest string, overwrite bool, wg *sync.WaitGroup, semaphore chan struct{}) {
	defer wg.Done()

	semaphore <- struct{}{}

	destPath := aDest
	if !overwrite {
		if _, err := os.Stat(destPath); !os.IsNotExist(err) {
			log.Warn().Str("destPath", destPath).Msg("destination file already exists, skipping")
			<-semaphore
			return
		}
	}
	// if the destination directory doesn't exist, create it
	destDir := filepath.Dir(destPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			log.Error().Err(err).Str("destDir", destDir).Msg("could not create destination directory")
			<-semaphore
			return
		}
	}
	file, err := os.Create(destPath)
	if err != nil {
		log.Error().Err(err).Str("destPath", destPath).Msg("could not create destination file")
		<-semaphore
		return
	}
	defer file.Close()

	downloader := s3manager.NewDownloaderWithClient(svc)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("could not download S3 file")
	}

	<-semaphore
}

func recursiveS3Copy(src string, baseDir, dest string, overwrite bool, ignores []string, isFlatCopy bool, maxDepth, maxConcurrent int) error {
	parsedURL, err := url.Parse(src)
	if err != nil {
		return fmt.Errorf("invalid S3 URL: %v", err)
	}

	if parsedURL.Scheme != "s3" {
		return fmt.Errorf("URL scheme must be 's3'")
	}

	bucket := parsedURL.Host
	prefix := strings.TrimPrefix(parsedURL.Path, "/")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region), // Update the region according to your S3 bucket
	})

	if err != nil {
		return err
	}

	svc := s3.New(sess)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent)

	var downloadFunc func(page *s3.ListObjectsV2Output, lastPage bool) bool
	downloadFunc = func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			objKey := *obj.Key
			aDest := isFlatCopyDest(objKey, baseDir, dest, isFlatCopy)

			// Check if the file or directory should be ignored
			shouldIgnore := false
			for _, ignore := range ignores {
				if strings.Contains(objKey, ignore) {
					shouldIgnore = true
					break
				}
			}
			if shouldIgnore {
				continue
			}

			wg.Add(1)
			go downloadS3File(svc, bucket, objKey, aDest, overwrite, &wg, semaphore)
		}

		wg.Wait()

		if maxDepth <= 0 {
			return false
		}
		if *page.IsTruncated {
			input := &s3.ListObjectsV2Input{
				Bucket:            aws.String(bucket),
				Prefix:            aws.String(prefix),
				ContinuationToken: page.NextContinuationToken,
			}
			err := svc.ListObjectsV2PagesWithContext(context.Background(), input, downloadFunc)
			if err != nil {
				log.Printf("error listing objects: %v", err)
			}
		}
		return true
	}
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	ctx := context.Background()
	err = svc.ListObjectsV2PagesWithContext(ctx, input, downloadFunc)
	if err != nil {
		log.Printf("error listing objects: %v", err)
	}
	return nil
}
func isFlatCopyDest(filename, baseDir, dest string, isFlatCopy bool) string {
	if isFlatCopy {
		return path.Join(baseDir, path.Base(filename))
	}
	return path.Join(dest, filename)
}
