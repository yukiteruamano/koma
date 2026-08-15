package pdf

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/util"

	_ "golang.org/x/image/webp"
)

type PDF struct{}

func New() *PDF {
	return &PDF{}
}

func (*PDF) Save(chapter *source.Chapter) (string, error) {
	return save(chapter, false)
}

func (*PDF) SaveTemp(chapter *source.Chapter) (string, error) {
	return save(chapter, true)
}

func save(chapter *source.Chapter, temp bool) (path string, err error) {
	path, err = chapter.Path(temp)
	if err != nil {
		return
	}

	file, err := filesystem.Api().Create(path)
	if err != nil {
		return
	}

	defer util.Ignore(file.Close)
	defer func() {
		// do not leave a partial file at the final path
		if err != nil {
			_ = filesystem.Api().Remove(path)
		}
	}()

	err = pagesToPDF(file, chapter.Pages)
	return
}

// pagesToPDF will convert images to PDF and write to w.
// The PDF context is built once and every image is appended to the same page
// tree, so the document is serialized a single time (O(n) instead of O(n^2)).
func pagesToPDF(w io.Writer, pages []*source.Page) error {
	conf := model.NewDefaultConfiguration()
	imp := pdfcpu.DefaultImportConfig()

	ctx, err := pdfcpu.CreateContextWithXRefTable(conf, imp.PageDim)
	if err != nil {
		return err
	}

	pagesIndRef, err := ctx.Pages()
	if err != nil {
		return err
	}

	pagesDict, err := ctx.DereferenceDict(*pagesIndRef)
	if err != nil {
		return err
	}

	for _, r := range pages {
		if r.Contents == nil {
			continue
		}

		// Read the page contents so we can decode image dimensions
		// and then set the page size to match the image,
		// preventing tall webtoon/manhwa pages from being clipped to A4.
		data := r.Contents.Bytes()

		if imgCfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
			imp.PageDim = &types.Dim{
				Width:  float64(imgCfg.Width),
				Height: float64(imgCfg.Height),
			}
		}

		indRefs, err := pdfcpu.NewPagesForImage(ctx.XRefTable, bytes.NewReader(data), pagesIndRef, imp)
		if err != nil {
			if viper.GetBool(key.FormatsSkipUnsupportedImages) {
				continue
			}

			return err
		}

		for _, indRef := range indRefs {
			if err := ctx.SetValid(*indRef); err != nil {
				return err
			}
			if err := model.AppendPageTree(indRef, 1, pagesDict); err != nil {
				return err
			}
			ctx.PageCount++
		}
	}

	return api.Write(ctx, w, conf)
}
