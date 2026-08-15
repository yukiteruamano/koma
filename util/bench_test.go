package util

import (
	"testing"

	"github.com/yukiteruamano/koma/key"
	"github.com/spf13/viper"
)

func BenchmarkSanitizeFilename(b *testing.B) {
	viper.SetDefault(key.DownloaderEscapeWhitespace, true)
	sample := "Chapter 01 [English] - The / Awkward ? Title!!"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizeFilename(sample)
	}
}
