package soundcloader

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name string
		args args
		want *URLInfo
	}{
		{
			name: "Old mobile URL",
			args: args{message: "https://m.soundcloud.com/cg5-beats/i-see-a-dreamer"},
			want: &URLInfo{User: "cg5-beats", Title: "i-see-a-dreamer", Kind: "song"},
		},
		{
			name: "New mobile URL",
			args: args{message: "https://soundcloud.app.goo.gl/8Mzh"},
			want: &URLInfo{User: "wfmn7igqznnw", Title: "bodiev-nutro", Kind: "song"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
