package cbz

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
	viper.Set(key.FormatsUse, constant.FormatCBZ)
	viper.Set(key.MetadataComicInfoXML, true)
}

func TestCBZ(t *testing.T) {
	cbz := New()

	Convey("Given a FormatCBZ converter", t, func() {
		Convey("When saving a chapter", func() {
			chapter := testutil.ChapterWithPages("chapter name", 3)
			result, err := cbz.Save(chapter)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
				Convey("And the result should be a path with .cbz extension", func() {
					So(result, ShouldNotBeEmpty)
					So(filepath.Ext(result), ShouldEqual, ".cbz")

					Convey("A path that can be read", func() {
						file, err := filesystem.Api().Open(result)
						So(err, ShouldBeNil)
						So(file, ShouldNotBeNil)

						info := lo.Must(file.Stat())

						zipReader := lo.Must(zip.NewReader(file, info.Size()))

						Convey("Zip file should contain ComicInfo.xml", func() {
							_, ok := lo.Find(zipReader.File, func(f *zip.File) bool {
								return f.Name == "ComicInfo.xml"
							})

							So(ok, ShouldBeTrue)
						})

						Convey("And the number of files should be equal to the number of pages + 1", func() {
							So(len(zipReader.File), ShouldEqual, len(chapter.Pages)+1)
						})
					})
				})
			})
		})
	})
}
