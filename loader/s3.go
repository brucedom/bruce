package loader

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path"
	"strings"
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
