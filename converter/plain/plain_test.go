package plain

import (
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yukiteruamano/koma/config"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/test/testutil"
	"os"
	"testing"
)

func init() {
	lo.Must0(config.Setup())
	filesystem.SetMemMapFs()
}

func Test(t *testing.T) {
	plain := New()

	Convey("Given a plain converter", t, func() {
		Convey("When saving a chapter", func() {
			chapter := testutil.ChapterWithPages("chapter name", 3)
			result, err := plain.Save(chapter)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
				Convey("And the result should be a path pointing to a directory", func() {
					So(result, ShouldNotBeEmpty)
					isDir, err := filesystem.Api().IsDir(result)

					if err != nil {
						t.Fatal(err)
					}

					So(isDir, ShouldBeTrue)

					Convey("And the directory should contain the chapter's pages", func() {
						files, err := filesystem.Api().ReadDir(result)
						if err != nil {
							t.Fatal(err)
						}

						So(len(files), ShouldEqual, len(chapter.Pages))

						lo.ForEach(files, func(file os.FileInfo, _ int) {
							So(file.Size(), ShouldBeGreaterThan, 0)
							So(file.IsDir(), ShouldBeFalse)
						})
					})
				})
			})
		})
	})
}
