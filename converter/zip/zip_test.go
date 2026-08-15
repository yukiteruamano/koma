package zip

import (
	"archive/zip"
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/config"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/test/testutil"
	"path/filepath"
	"testing"
)

func init() {
	filesystem.SetMemMapFs()
	lo.Must0(config.Setup())
	viper.Set(key.FormatsUse, constant.FormatZIP)
}

func TestZIP(t *testing.T) {
	z := New()

	Convey("Given a FormatZIP converter", t, func() {
		Convey("When saving a chapter", func() {
			chapter := testutil.ChapterWithPages("chapter name", 3)
			result, err := z.Save(chapter)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
				Convey("And the result should be a path with .zip extension", func() {
					So(result, ShouldNotBeEmpty)
					So(filepath.Ext(result), ShouldEqual, ".zip")

					Convey("A path that can be read", func() {
						file, err := filesystem.Api().Open(result)
						So(err, ShouldBeNil)
						So(file, ShouldNotBeNil)

						info := lo.Must(file.Stat())

						zipReader := lo.Must(zip.NewReader(file, info.Size()))

						Convey("And the number of files should be equal to the number of pages", func() {
							So(len(zipReader.File), ShouldEqual, len(chapter.Pages))
						})
					})
				})
			})
		})
	})
}
