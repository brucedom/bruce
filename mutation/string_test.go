package mutation

import "testing"

func TestStripNonAlnum(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure stripped",
			args: args{str: "The quick! Brown ~ Fox - Jumps #?!|"},
			want: "ThequickBrownFoxJumps",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripNonAlnum(tt.args.str); got != tt.want {
				t.Errorf("StripNonAlnum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripExtraWhitespace(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "doublespaces",
			args: args{str: "the  quick brown  fox    jumps"},
			want: "the quick brown fox jumps",
		}, {
			name: "tabs",
			args: args{str: "                foo bar baz"},
			want: " foo bar baz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripExtraWhitespace(tt.args.str); got != tt.want {
				t.Errorf("StripExtraWhitespace() = \nrecv:%#v \nwant:%#v", got, tt.want)
			}
		})
	}
}

func TestStripExtraWhitespaceFB(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "doublespaces",
			args: args{str: " the  quick brown  fox    jumps "},
			want: "the quick brown fox jumps",
		}, {
			name: "tabs",
			args: args{str: "                foo bar baz"},
			want: "foo bar baz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripExtraWhitespaceFB(tt.args.str); got != tt.want {
				t.Errorf("StripExtraWhitespaceFB() = \nrecv:%#v \nwant:%#v", got, tt.want)
			}
		})
	}
}
