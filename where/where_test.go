package where

import (
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yukiteruamano/koma/filesystem"
	"testing"
)

func init() {
	filesystem.SetMemMapFs()
}

func TestConfig(t *testing.T) {
	Convey("When gettings config path", t, func() {
		path := Config()
		Convey("It should exist", func() {
			exists := lo.Must(filesystem.Api().Exists(path))
			So(exists, ShouldBeTrue)

			Convey("And it should be a directory", func() {
				isDir := lo.Must(filesystem.Api().IsDir(path))
				So(isDir, ShouldBeTrue)
			})
		})
	})
}

func TestLogs(t *testing.T) {
	Convey("When gettings logs path", t, func() {
		path := Logs()
		Convey("It should exist", func() {
			exists := lo.Must(filesystem.Api().Exists(path))
			So(exists, ShouldBeTrue)

			Convey("And it should be a directory", func() {
				isDir := lo.Must(filesystem.Api().IsDir(path))
				So(isDir, ShouldBeTrue)
			})
		})
	})
}

func TestTempDirRecreatedAfterDeletion(t *testing.T) {
	path := Temp()
	if _, err := filesystem.Api().Stat(path); err != nil {
		t.Fatalf("temp dir should exist after first call: %v", err)
	}

	// simulate the startup cleanup that deletes the memoized temp dir
	if err := filesystem.Api().RemoveAll(path); err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	recreated := Temp()
	if recreated != path {
		t.Fatalf("Temp() returned %q, want %q", recreated, path)
	}

	if _, err := filesystem.Api().Stat(path); err != nil {
		t.Errorf("temp dir should be recreated after deletion: %v", err)
	}
}
