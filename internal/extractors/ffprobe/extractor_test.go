package ffprobe

import (
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

func TestParseVideoMetadata(t *testing.T) {
	got, err := Parse("recovered/video.mp4", []byte(`{
	  "format": {
	    "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
	    "format_long_name": "QuickTime / MOV",
	    "duration": "12.500000",
	    "tags": {
	      "title": "Family Video",
	      "creation_time": "2020-10-03T18:12:01.000000Z"
	    }
	  },
	  "streams": [
	    {"codec_type":"audio","codec_name":"aac"},
	    {"codec_type":"video","codec_name":"h264","width":1920,"height":1080}
	  ]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.FormatIDs) != 1 || got.FormatIDs[0].Source != evidence.SourceFFProbe {
		t.Fatalf("FormatIDs = %+v", got.FormatIDs)
	}
	if got.Media == nil {
		t.Fatal("Media evidence missing")
	}
	if got.Media.DurationSeconds != 12.5 || got.Media.Codec != "h264" || got.Media.Width != 1920 || got.Media.Height != 1080 {
		t.Fatalf("Media = %+v", got.Media)
	}
	if got.Metadata["creation_time"] != "2020-10-03T18:12:01.000000Z" {
		t.Fatalf("Metadata = %+v", got.Metadata)
	}
	if got.TextSignals[0] != "Family Video" {
		t.Fatalf("TextSignals = %+v", got.TextSignals)
	}
}

func TestParseAudioMetadata(t *testing.T) {
	got, err := Parse("recovered/audio.mp3", []byte(`{
	  "format": {
	    "format_name": "mp3",
	    "duration": "180.25",
	    "tags": {
	      "title": "Start",
	      "artist": "Retro Metro"
	    }
	  },
	  "streams": [
	    {"codec_type":"audio","codec_name":"mp3"}
	  ]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Media == nil || got.Media.Codec != "mp3" || got.Media.DurationSeconds != 180.25 {
		t.Fatalf("Media = %+v", got.Media)
	}
	if len(got.TextSignals) != 2 || got.TextSignals[0] != "Start" || got.TextSignals[1] != "Retro Metro" {
		t.Fatalf("TextSignals = %+v", got.TextSignals)
	}
}
