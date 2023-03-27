package loader

import (
	"io"
	"reflect"
	"testing"
)

func readFromReader(t *testing.T, reader io.Reader) string {
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	return string(data)
}

func TestReaderFromHttp(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "read from http",
			args:    args{fileName: "https://raw.githubusercontent.com/Nitecon/bruce/main/test.txt"},
			want:    "HelloWorld",
			want1:   "test.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ReaderFromHttp(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReaderFromHttp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(readFromReader(t, got), tt.want) {
				t.Errorf("ReaderFromHttp() got = %s, want %s", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ReaderFromHttp() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
