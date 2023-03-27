package mutation

import (
	"io"
	"os"
	"testing"
)

func TestWriteInlineTemplate(t *testing.T) {
	// Test valid template
	tpl := "{{ .Name }} {{ .Time }}"
	content := struct {
		Name string
		Time string
	}{"test", "* * * * *"}
	err := WriteInlineTemplate("test", tpl, content)
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	defer os.Remove("test")
	// Check if file has been written
	file, err := os.Open("test")
	if err != nil {
		t.Error("Expected file to exist, got", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if string(data) != "test * * * * *" {
		t.Error("Expected file to contain \n'test * * * * *', got\n", string(data))
	}

	// Test invalid template
	tpl = "{{ .Name } {{ .Time }}"
	err = WriteInlineTemplate("test", tpl, content)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Clean up
	os.Remove("/etc/cron.d/test")
}
