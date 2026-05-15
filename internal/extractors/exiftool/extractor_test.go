package exiftool

import (
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

func TestParseImageMetadata(t *testing.T) {
	got, err := Parse("recovered/photo", []byte(`[
	  {
	    "SourceFile": "recovered/photo",
	    "File:FileType": "JPEG",
	    "File:FileTypeExtension": "jpg",
	    "File:MIMEType": "image/jpeg",
	    "File:ImageWidth": 4032,
	    "File:ImageHeight": 3024,
	    "EXIF:Make": "Apple",
	    "EXIF:Model": "iPhone 12",
	    "EXIF:DateTimeOriginal": "2021:08:14 16:22:09",
	    "XMP:Title": "Family Picnic"
	  }
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if got.DetectedMIME != "image/jpeg" || got.Extension != ".jpg" {
		t.Fatalf("mime/ext = %q/%q", got.DetectedMIME, got.Extension)
	}
	if len(got.FormatIDs) != 1 || got.FormatIDs[0].Source != evidence.SourceExifTool {
		t.Fatalf("FormatIDs = %+v", got.FormatIDs)
	}
	if got.Image == nil {
		t.Fatal("Image evidence missing")
	}
	if got.Image.Width != 4032 || got.Image.Height != 3024 || got.Image.CameraModel != "iPhone 12" {
		t.Fatalf("Image = %+v", got.Image)
	}
	if got.TextSignals[0] != "Family Picnic" {
		t.Fatalf("TextSignals = %+v", got.TextSignals)
	}
}

func TestParseMediaMetadata(t *testing.T) {
	got, err := Parse("recovered/video", []byte(`[
	  {
	    "SourceFile": "recovered/video",
	    "File:FileTypeExtension": "mp4",
	    "File:MIMEType": "video/mp4",
	    "QuickTime:Duration": 12.5,
	    "QuickTime:VideoCodec": "h264",
	    "QuickTime:ImageWidth": 1920,
	    "QuickTime:ImageHeight": 1080,
	    "QuickTime:Title": "Family Video"
	  }
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Media == nil {
		t.Fatal("Media evidence missing")
	}
	if got.Media.DurationSeconds != 12.5 || got.Media.Codec != "h264" || got.Media.Width != 1920 {
		t.Fatalf("Media = %+v", got.Media)
	}
	if got.TextSignals[0] != "Family Video" {
		t.Fatalf("TextSignals = %+v", got.TextSignals)
	}
}
