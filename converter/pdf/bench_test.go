package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/source"
)

func makeTestPage(w, h int) *source.Page {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return &source.Page{Contents: &buf}
}

func BenchmarkPDFAssembly(b *testing.B) {
	viper.Set(key.FormatsSkipUnsupportedImages, false)

	const pageCount = 20
	pages := make([]*source.Page, pageCount)
	for i := range pages {
		pages[i] = makeTestPage(400, 600)
	}

	var out bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out.Reset()
		if err := pagesToPDF(&out, pages); err != nil {
			b.Fatal(err)
		}
	}
}
