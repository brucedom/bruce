package loader

import (
	"os"
	"testing"
)

func TestGetRemoteReader(t *testing.T) {
	// Test HTTP
	rc, filename, err := GetRemoteReader("http://example.com/file.txt")
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if filename != "file.txt" {
		t.Error("Expected filename to be file.txt, got", filename)
	}
	if rc == nil {
		t.Error("Expected non-nil ReadCloser, got nil")
	}
	rc.Close()

	// Test S3
	/*
		rc, filename, err = GetRemoteReader("s3://bucket/file.txt")
		if err != nil {
			t.Error("Expected no error, got", err)
		}
		if filename != "file.txt" {
			t.Error("Expected filename to be file.txt, got", filename)
		}
		if rc == nil {
			t.Error("Expected non-nil ReadCloser, got nil")
		}
		rc.Close()
	*/
	file, err := os.Create("file.txt")
	if err != nil {
		// Handle error
	}
	defer os.Remove("file.txt")

	_, err = file.WriteString("helloworld")
	if err != nil {
		// Handle error
	}

	// Test Local
	rc, filename, err = GetRemoteReader("file.txt")
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if filename != "file.txt" {
		t.Error("Expected filename to be file.txt, got", filename)
	}
	if rc == nil {
		t.Error("Expected non-nil ReadCloser, got nil")
	}
	rc.Close()

	file.Close()

	// Test Invalid URL
	rc, filename, err = GetRemoteReader("invalid")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if rc != nil {
		t.Error("Expected nil ReadCloser, got non-nil")
	}
}

func TestReadRemoteFile1(t *testing.T) {
	// Create test file with contents "helloworld"
	err := os.WriteFile("test.txt", []byte("helloworld"), 0644)
	if err != nil {
		t.Error("Error creating test file: ", err)
	}
	defer os.Remove("test.txt")

	// Test the ReadRemoteFile function
	d, fn, err := ReadRemoteFile("test.txt")
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if fn != "test.txt" {
		t.Error("Expected filename to be test.txt, got", fn)
	}
	if string(d) != "helloworld" {
		t.Error("Expected data to be 'helloworld', got", string(d))
	}

	// Test invalid file
	_, _, err = ReadRemoteFile("invalid.txt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
