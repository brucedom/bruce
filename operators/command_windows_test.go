//go:build windows
// +build windows

package operators

import (
	"testing"
)

func TestCommand_Execute(t *testing.T) {
	type fields struct {
		Name string
		Cmd  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "run directory output",
			fields:  fields{Name: "list dirs", Cmd: "dir"},
			wantErr: false,
		},
		{
			name:    "run invalid command",
			fields:  fields{Name: "invalid command", Cmd: "fubarz.exe"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Log("running windows tests")
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{
				Cmd: tt.fields.Cmd,
			}
			if err := c.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

	}
}
