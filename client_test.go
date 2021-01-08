package soundcloader

import (
	"log"
	"os"
	"testing"
)

func Test_Get(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Song", args{s: "https://soundcloud.com/unitasprima/yasuha-flyday-chinatown"}, false},
		{"Song from playlist", args{s: "https://soundcloud.com/iamtrevordaniel/falling?in=iamtrevordaniel/sets/homesick"}, false},
		{"Station", args{s: "https://soundcloud.com/stations/track/unitasprima/yasuha-flyday-chinatown"}, true},
		{"Playlist", args{s: "https://soundcloud.com/discover/sets/charts-top:all-music:ua"}, true},
		{"Playlist #2", args{s: "https://soundcloud.com/nitza-md/sets/piano-deep-concentration"}, true},
		{"User", args{s: "https://soundcloud.com/faceless1-7"}, true},
		{"song with some text", args{"Check this: https://soundcloud.com/unitasprima/yasuha-flyday-chinatown"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := DefaultClient
			//cl.SetDebug(true)
			var filename string
			result, err := cl.Get(tt.args.s)
			if err == nil {
				filename, err = result.GetNext()
			}

			if err != nil || result == nil {
				if tt.wantErr {
					return
				}
				log.Print(err)
				t.Fail()
				return
			}

			log.Print(result.Thumbnail)

			if _, err := os.Stat(filename); err != nil {
				t.Fail()
			}
		})
	}
}

// just to make sure gofmt wont throw errors
func _() {
	_, _ = Get("")
	_, _ = GetURL(nil)
	_ = ParseURL(nil)
}
