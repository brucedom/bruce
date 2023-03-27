package exe

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestExecution_ContainsLC(t *testing.T) {
	type fields struct {
		input       string
		fields      []string
		useSudo     bool
		outputStr   string
		isError     bool
		cmnd        string
		args        []string
		regex       *regexp.Regexp
		regexString string
		err         error
	}
	type args struct {
		c string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "hasOutputMatch",
			want:   true,
			args:   args{c: "hello"},
			fields: fields{outputStr: "hello"},
		},
		{
			name:   "hasErrorMatch",
			want:   true,
			args:   args{c: "hello"},
			fields: fields{err: fmt.Errorf("hello")},
		},
		{
			name:   "hasNoMatch",
			want:   false,
			args:   args{c: "fubar"},
			fields: fields{err: fmt.Errorf("hello")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Execution{
				input:       tt.fields.input,
				fields:      tt.fields.fields,
				useSudo:     tt.fields.useSudo,
				outputStr:   tt.fields.outputStr,
				isError:     tt.fields.isError,
				cmnd:        tt.fields.cmnd,
				args:        tt.fields.args,
				regex:       tt.fields.regex,
				regexString: tt.fields.regexString,
				err:         tt.fields.err,
			}
			if got := e.ContainsLC(tt.args.c); got != tt.want {
				t.Errorf("ContainsLC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecution_GetErr(t *testing.T) {
	type fields struct {
		input       string
		fields      []string
		useSudo     bool
		outputStr   string
		isError     bool
		cmnd        string
		args        []string
		regex       *regexp.Regexp
		regexString string
		err         error
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "testIfErrSet",
			fields:  fields{err: fmt.Errorf("testme")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Execution{
				input:       tt.fields.input,
				fields:      tt.fields.fields,
				useSudo:     tt.fields.useSudo,
				outputStr:   tt.fields.outputStr,
				isError:     tt.fields.isError,
				cmnd:        tt.fields.cmnd,
				args:        tt.fields.args,
				regex:       tt.fields.regex,
				regexString: tt.fields.regexString,
				err:         tt.fields.err,
			}
			if err := e.GetErr(); (err != nil) != tt.wantErr {
				t.Errorf("GetErr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecution_SetRegex(t *testing.T) {
	type fields struct {
		input       string
		fields      []string
		useSudo     bool
		outputStr   string
		isError     bool
		cmnd        string
		args        []string
		regex       *regexp.Regexp
		regexString string
		err         error
	}
	type args struct {
		re string
	}
	regx, _ := regexp.Compile("p([a-z]+)ch")
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *regexp.Regexp
		wantErr bool
	}{
		{
			name:    "regexSuccess",
			fields:  fields{},
			wantErr: false,
			args:    args{re: "p([a-z]+)ch"},
			want:    regx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Execution{
				input:       tt.fields.input,
				fields:      tt.fields.fields,
				useSudo:     tt.fields.useSudo,
				outputStr:   tt.fields.outputStr,
				isError:     tt.fields.isError,
				cmnd:        tt.fields.cmnd,
				args:        tt.fields.args,
				regex:       tt.fields.regex,
				regexString: tt.fields.regexString,
				err:         tt.fields.err,
			}
			got, err := e.SetRegex(tt.args.re)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetRegex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecution_RegexMatch(t *testing.T) {
	type fields struct {
		input       string
		fields      []string
		useSudo     bool
		outputStr   string
		isError     bool
		cmnd        string
		args        []string
		regex       *regexp.Regexp
		regexString string
		err         error
	}
	r, _ := regexp.Compile("p([a-z]+)ch")
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "outputSuccess",
			fields: fields{outputStr: "peachy", regex: r},
			want:   true,
		},
		{
			name:   "regexNotCompiled",
			fields: fields{outputStr: "peachy", regex: nil},
			want:   false,
		},
		{
			name:   "errorMatched",
			fields: fields{err: fmt.Errorf("peachy"), regex: r},
			want:   true,
		},
		{
			name:   "noMatchCompile",
			fields: fields{err: fmt.Errorf("channels"), regex: r},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Execution{
				input:       tt.fields.input,
				fields:      tt.fields.fields,
				useSudo:     tt.fields.useSudo,
				outputStr:   tt.fields.outputStr,
				isError:     tt.fields.isError,
				cmnd:        tt.fields.cmnd,
				args:        tt.fields.args,
				regex:       tt.fields.regex,
				regexString: tt.fields.regexString,
				err:         tt.fields.err,
			}
			if got := e.RegexMatch(); got != tt.want {
				t.Errorf("RegexMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecution_Failed(t *testing.T) {
	type fields struct {
		input       string
		fields      []string
		useSudo     bool
		outputStr   string
		isError     bool
		cmnd        string
		args        []string
		regex       *regexp.Regexp
		regexString string
		err         error
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "success",
			want:   true,
			fields: fields{err: fmt.Errorf("hi"), isError: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Execution{
				input:       tt.fields.input,
				fields:      tt.fields.fields,
				useSudo:     tt.fields.useSudo,
				outputStr:   tt.fields.outputStr,
				isError:     tt.fields.isError,
				cmnd:        tt.fields.cmnd,
				args:        tt.fields.args,
				regex:       tt.fields.regex,
				regexString: tt.fields.regexString,
				err:         tt.fields.err,
			}
			if got := e.Failed(); got != tt.want {
				t.Errorf("Failed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on windows")
	}
	type args struct {
		c       string
		useSudo bool
	}
	tests := []struct {
		name string
		args args
		want *Execution
	}{
		{
			name: "success",
			args: args{c: "echo foo", useSudo: false},
			want: &Execution{
				input:     "echo foo",
				fields:    []string{"echo", "foo"},
				isError:   false,
				cmnd:      "echo",
				args:      []string{"foo"},
				outputStr: "foo",
				err:       nil,
			},
		},
		{
			name: "withSudo",
			args: args{c: "echo foo", useSudo: true},
			want: &Execution{
				input:     "echo foo",
				fields:    []string{"echo", "foo"},
				isError:   false,
				useSudo:   true,
				cmnd:      "sudo",
				args:      []string{"echo", "foo"},
				outputStr: "foo",
				err:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Run(tt.args.c, tt.args.useSudo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\n%#v\n want=\n%#v", got, tt.want)
			}
		})
	}
}

func TestHasExecInPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		type args struct {
			name string
		}
		tests := []struct {
			name string
			args args
			want string
		}{
			{
				name: "windows",
				args: args{name: "explorer"},
				want: "C:\\Windows\\explorer.exe",
			},
			{
				name: "windows NotFound",
				args: args{name: "explorers"},
				want: "",
			},
			{
				name: "windowsDotPath",
				args: args{name: "./explorers"},
				want: "",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := HasExecInPath(tt.args.name); got != tt.want {
					t.Errorf("HasExecInPath() = %v, want %v", got, tt.want)
				}
			})
		}
	}
	if runtime.GOOS == "linux" {
		type args struct {
			name string
		}
		tests := []struct {
			name string
			args args
			want string
		}{
			{
				name: "linux",
				args: args{name: "sh"},
				want: "/usr/bin/sh",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := HasExecInPath(tt.args.name); got != tt.want {
					t.Errorf("HasExecInPath() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

func TestGetFileChecksum(t *testing.T) {
	path := fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, "checksumTest")
	err := os.WriteFile(path, []byte("foo"), 0664)
	if err != nil {
		t.Fatal("could not create temp file for test")
	}
	defer os.Remove(path)
	type args struct {
		fname string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "ValidHash",
			args:    args{fname: path},
			want:    "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			wantErr: false,
		},
		{
			name:    "InvalidHash",
			args:    args{fname: "foobar"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileChecksum(tt.args.fname)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileChecksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFileChecksum() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	path := fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, "checksumTest")
	err := os.WriteFile(path, []byte("foo"), 0664)
	if err != nil {
		t.Fatal("could not create temp file for test")
	}
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "success delete temp file",
			args:    args{src: path},
			wantErr: false,
		},
		{
			name:    "fail on already deleted temp file",
			args:    args{src: path},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteFile(tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("DeleteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	path := fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, "checksumTest")
	err := os.WriteFile(path, []byte("foo"), 0664)
	if err != nil {
		t.Fatal("could not create temp file for test")
	}
	defer os.Remove(path)
	type args struct {
		src string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "fileExists",
			args: args{src: path},
			want: true,
		},
		{
			name: "fileDoesNotExists",
			args: args{src: "foothebarbazbono.foo"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.src); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEchoToFile(t *testing.T) {
	got := EchoToFile("echo hi", os.TempDir())
	if runtime.GOOS == "windows" {
		if !strings.Contains(got, ".bat") {
			t.Errorf("EchoToFile() | got [%s] wanted [%s]", got, ".bat")
		}
	} else {
		if !strings.Contains(got, ".sh") {
			t.Errorf("EchoToFile() | got [%s] wanted [%s]", got, ".sh")
		}
	}
}

func TestCopyFile(t *testing.T) {
	path := fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, "checksumTest.txt")
	err := os.WriteFile(path, []byte("foo"), 0664)
	if err != nil {
		t.Fatal("could not create temp file for test")
	}
	defer os.Remove(path)

	type args struct {
		src      string
		dst      string
		makedirs bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "existingValidCopy",
			args:    args{src: path, dst: fmt.Sprintf("%s.copy", path)},
			wantErr: false,
		},
		{
			name: "NotExist",
			args: args{
				src:      fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, "foo.test"),
				dst:      fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, "foo.test2"),
				makedirs: false,
			},
			wantErr: true,
		},
		{
			name: "NotExistMakeDir",
			args: args{
				src:      path,
				dst:      fmt.Sprintf("%s%c%s%c%s", os.TempDir(), os.PathSeparator, "deepdirtest", os.PathSeparator, "checksumTest.txt"),
				makedirs: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyFile(tt.args.src, tt.args.dst, tt.args.makedirs); (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		t.Logf("removing temp file: %s", tt.args.dst)
		os.Remove(tt.args.dst)
	}
}

// obType, opath, owner, group string, recursive bool
func TestSetOwnership(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on windows")
	}
	tf := path.Join(os.TempDir(), "testFile.txt")
	os.WriteFile(tf, []byte("stuffhere"), 0664)
	defer os.Remove(tf)
	type args struct {
		obType    string
		opath     string
		owner     string
		group     string
		recursive bool
	}
	u, err := user.Current()
	if err != nil {
		t.Errorf("couldnt' look up current user: %s", err)
	}
	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		t.Errorf("couldnt' look up group for the current user: %s", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "validTest",
			args:    args{obType: "file", owner: u.Username, group: g.Name, opath: tf, recursive: false},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetOwnership(tt.args.obType, tt.args.opath, tt.args.owner, tt.args.group, tt.args.recursive); (err != nil) != tt.wantErr {
				t.Errorf("SetOwnership() error = \n%#v\nwant\n%#v", err, tt.wantErr)
			}
		})
	}

}
