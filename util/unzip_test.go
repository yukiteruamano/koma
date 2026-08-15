package util

import (
	"archive/zip"
	"bytes"
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yukiteruamano/koma/filesystem"
	"path/filepath"
	"testing"
)

func TestUnzip(t *testing.T) {
	// build a zip archive in memory
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range []struct {
		name string
		body string
	}{
		{name: "zipdata/hey.jpeg", body: "jpegbytes"},
		{name: "zipdata/a/hello.txt", body: "hello"},
	} {
		w, err := zw.Create(f.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(f.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	filesystem.SetMemMapFs()

	zipPath := "zipdata.zip"
	if err := filesystem.Api().WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := filesystem.Api().Open(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	Convey("Given a zip file", t, func() {
		Convey("When unzipping it", func() {
			err := Unzip(file, int64(buf.Len()), "a")
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
				Convey("And the files should be extracted", func() {
					for _, info := range []lo.Tuple2[string, bool]{
						{filepath.Join("a", "zipdata", "hey.jpeg"), false},
						{filepath.Join("a", "zipdata", "a", "hello.txt"), false},
					} {
						filename := info.A

						exists := lo.Must(filesystem.Api().Exists(filename))
						So(exists, ShouldBeTrue)
					}
				})
			})
		})
	})
}
