package pdf

import (
	"bytes"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/config"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/test/testutil"
	"path/filepath"
	"testing"
)

func init() {
	filesystem.SetMemMapFs()
	lo.Must0(config.Setup())
	viper.Set(key.FormatsUse, constant.FormatPDF)
	viper.Set(key.FormatsSkipUnsupportedImages, false)
}

func TestPDF(t *testing.T) {
	pdf := New()

	Convey("Given a FormatPDF converter", t, func() {
		Convey("When saving a chapter", func() {
			chapter := testutil.ChapterWithPages("chapter name", 3)
			result, err := pdf.Save(chapter)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
				Convey("And the result should be a path with .pdf extension", func() {
					So(result, ShouldNotBeEmpty)
					So(filepath.Ext(result), ShouldEqual, ".pdf")

					Convey("A path that can be read", func() {
						file, err := filesystem.Api().Open(result)
						So(err, ShouldBeNil)
						So(file, ShouldNotBeNil)

						Convey("And the file should not be empty", func() {
							info := lo.Must(file.Stat())
							So(info.Size(), ShouldBeGreaterThan, 0)
						})
					})
				})
			})
		})
	})
}

func TestPagesToPDFProducesExpectedPageCount(t *testing.T) {
	chapter := testutil.ChapterWithPages("pages", 3)

	var out bytes.Buffer
	if err := pagesToPDF(&out, chapter.Pages); err != nil {
		t.Fatalf("pagesToPDF failed: %v", err)
	}

	ctx, err := api.ReadAndValidate(bytes.NewReader(out.Bytes()), nil)
	if err != nil {
		t.Fatalf("generated PDF is not parseable: %v", err)
	}

	if ctx.PageCount != len(chapter.Pages) {
		t.Errorf("PageCount = %d, want %d", ctx.PageCount, len(chapter.Pages))
	}
}

func TestPagesToPDFSkipsUnsupportedImages(t *testing.T) {
	viper.Set(key.FormatsSkipUnsupportedImages, true)
	defer viper.Set(key.FormatsSkipUnsupportedImages, false)

	chapter := testutil.ChapterWithPages("pages", 2)

	// inject a corrupt "image" that pdfcpu cannot decode
	chapter.Pages = append(chapter.Pages, &source.Page{
		Contents: bytes.NewBuffer([]byte("this is not an image")),
	})

	var out bytes.Buffer
	if err := pagesToPDF(&out, chapter.Pages); err != nil {
		t.Fatalf("pagesToPDF should skip unsupported images: %v", err)
	}
}

func TestPagesToPDFRejectsUnsupportedImages(t *testing.T) {
	viper.Set(key.FormatsSkipUnsupportedImages, false)

	chapter := testutil.ChapterWithPages("pages", 1)
	chapter.Pages = append(chapter.Pages, &source.Page{
		Contents: bytes.NewBuffer([]byte("this is not an image")),
	})

	var out bytes.Buffer
	if err := pagesToPDF(&out, chapter.Pages); err == nil {
		t.Fatal("expected an error for an unsupported image")
	}
}

func TestPagesToPDFSkipsNilContents(t *testing.T) {
	chapter := testutil.ChapterWithPages("pages", 2)
	chapter.Pages[1].Contents = nil

	var out bytes.Buffer
	if err := pagesToPDF(&out, chapter.Pages); err != nil {
		t.Fatalf("pagesToPDF should skip nil contents: %v", err)
	}

	ctx, err := api.ReadAndValidate(bytes.NewReader(out.Bytes()), nil)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.PageCount != 1 {
		t.Errorf("PageCount = %d, want 1", ctx.PageCount)
	}
}
