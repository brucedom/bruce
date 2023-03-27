package mutation

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func CreateTarball(tarballFilePath string, thePath string) error {
	file, err := os.Create(tarballFilePath)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not create tarball file '%s', got error '%s'", tarballFilePath, err.Error()))
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	err = filepath.Walk(thePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path != "." && path != ".." {
				err = addFileToTarWriter(path, tarWriter)
				if err != nil {
					return fmt.Errorf(fmt.Sprintf("Could not add file '%s', to tarball, got error '%s'", path, err.Error()))
				}
			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}

// Private methods

func addFileToTarWriter(filePath string, tarWriter *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not open file '%s', got error '%s'", filePath, err.Error()))
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not get stat for file '%s', got error '%s'", filePath, err.Error()))
	}

	if !stat.IsDir() {
		hpath := strings.Replace(filePath, os.TempDir()+"/", "", 1)
		//fmt.Printf("OldPath: %s\n", filePath)
		//fmt.Printf("NewPath: %s\n", hpath)
		header := &tar.Header{
			Name:    hpath,
			Size:    stat.Size(),
			Mode:    int64(stat.Mode()),
			ModTime: stat.ModTime(),
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Could not write header for file '%#v', got error '%s'", header, err.Error()))
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Could not copy the file '%s' data to the tarball, got error '%s'", filePath, err.Error()))
		}
	}
	return nil
}

func setupInitialTarball(t *testing.T) {
	createDir := path.Join(os.TempDir(), "test")
	err := os.MkdirAll(createDir, 0775)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(createDir)

	fn1 := path.Join(createDir, "test1.txt")
	err = os.WriteFile(fn1, []byte("test1"), 0644)
	if err != nil {
		t.Error("Error creating test file: ", err)
	}
	fn2 := path.Join(createDir, "test2.txt")
	err = os.WriteFile(fn2, []byte("test2"), 0644)
	if err != nil {
		t.Error("Error creating test file: ", err)
	}

	// Create the archive and write the output to the "out" Writer
	err = CreateTarball("test.tar.gz", createDir)
	if err != nil {
		log.Fatalln("Error creating archive:", err)
	}
}

func Test_useGzipReader(t *testing.T) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte("helloworld")); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	err := os.WriteFile("test.tgz", b.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove("test.tgz")
	// Test gzip file
	fileReader, _ := os.Open("test.tgz")
	defer fileReader.Close()
	gzr := useGzipReader("test.tgz", fileReader)
	data, err := io.ReadAll(gzr)
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if string(data) != "helloworld" {
		t.Error("Expected data to be 'helloworld', got", string(data))
	}

	// Test non-gzip file
	fileReader, _ = os.Open("test.txt")
	defer fileReader.Close()
	r := useGzipReader("test.txt", fileReader)
	if _, ok := r.(*gzip.Reader); ok {
		t.Error("Expected non-gzip reader, got gzip reader")
	}
}

func TestExtractTarball(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if _, err := os.Stat("test.tar.gz"); err == nil {
		os.Remove("test.tar.gz")
	}
	if _, err := os.Stat("extracted/"); err == nil {
		os.RemoveAll("extracted/")
	}
	setupInitialTarball(t)

	defer os.Remove("test.tar.gz")

	// Test extracting tarball
	err := ExtractTarball("test.tar.gz", "extracted", true, true)
	if err != nil {
		t.Error("Expected no error, got", err)
	}

	// Check if files have been extracted
	_, err = os.Stat("extracted/test1.txt")
	if err != nil {
		t.Error("Expected test1.txt to exist, got", err)
	}
	_, err = os.Stat("extracted/test2.txt")
	if err != nil {
		t.Error("Expected test2.txt to exist, got", err)
	}
	os.RemoveAll("extracted")
	// Test extracting to existing directory
	err = ExtractTarball("test.tar.gz", "extracted", false, true)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Clean up
	os.RemoveAll("extracted")
}
