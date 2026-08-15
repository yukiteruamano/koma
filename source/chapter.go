package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/samber/mo"
	"github.com/sourcegraph/conc/pool"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/style"
	"github.com/yukiteruamano/koma/util"
)

// Chapter is a struct that represents a chapter of a manga.
type Chapter struct {
	// Name of the chapter
	Name string `json:"name" jsonschema:"description=Name of the chapter"`
	// URL of the chapter
	URL string `json:"url" jsonschema:"description=URL of the chapter"`
	// Index of the chapter in the manga.
	Index uint16 `json:"index" jsonschema:"description=Index of the chapter in the manga"`
	// ID of the chapter in the source.
	ID string `json:"id" jsonschema:"description=ID of the chapter in the source"`
	// Volume which the chapter belongs to.
	Volume string `json:"volume" jsonschema:"description=Volume which the chapter belongs to"`
	// PublishDate is the original publish date of the chapter from the source.
	PublishDate date `json:"publish_date" jsonschema:"description=Original publish date of the chapter"`
	// Manga that the chapter belongs to.
	Manga *Manga `json:"-"`
	// Pages of the chapter.
	Pages []*Page `json:"pages" jsonschema:"description=Pages of the chapter"`

	isDownloaded mo.Option[bool]
	size         uint64
}

func (c *Chapter) String() string {
	return c.Name
}

// DownloadPages downloads the Pages contents of the Chapter into memory.
// Pages needs to be set before calling this function.
func (c *Chapter) DownloadPages(temp bool, progress func(string)) error {
	return c.downloadAll(func(page *Page) error { return page.Download() }, temp, progress)
}

// DownloadPagesTo streams the Pages contents of the Chapter straight to the
// given directory on disk, keeping memory bounded.
// Pages needs to be set before calling this function.
func (c *Chapter) DownloadPagesTo(dir string, temp bool, progress func(string)) error {
	if err := filesystem.Api().MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return c.downloadAll(func(page *Page) error {
		return page.DownloadTo(filepath.Join(dir, page.Filename()))
	}, temp, progress)
}

// downloadAll downloads every page using the given download function.
// It runs sequentially when downloader.async is off, otherwise through a
// bounded worker pool sized by downloader.concurrency.
func (c *Chapter) downloadAll(download func(*Page) error, temp bool, progress func(string)) error {
	if len(c.Pages) == 0 {
		return fmt.Errorf("chapter %q has no pages", c.Name)
	}

	for i, page := range c.Pages {
		if page == nil {
			return fmt.Errorf("page #%d is nil, aborting download", i)
		}
	}

	c.size = 0
	status := func() string {
		return fmt.Sprintf(
			"Downloading %s %s",
			util.Quantify(len(c.Pages), "page", "pages"),
			style.Faint(c.SizeHuman()),
		)
	}

	// Sequential path (downloader.async off).
	if !viper.GetBool(key.DownloaderAsync) {
		progress(status())
		for _, page := range c.Pages {
			if err := download(page); err != nil {
				c.isDownloaded = mo.Some(false)
				return err
			}
			c.size += page.Size
			progress(status())
		}

		c.isDownloaded = mo.Some(!temp)
		return nil
	}

	concurrency := viper.GetInt(key.DownloaderConcurrency)
	if concurrency < 1 {
		concurrency = 1
	}

	progress(status())

	// A single goroutine applies size deltas and reports progress, so workers
	// never touch shared state concurrently and progress callbacks are serialized.
	sizes := make(chan uint64, concurrency)
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		for delta := range sizes {
			c.size += delta
			progress(status())
		}
	}()

	p := pool.New().WithMaxGoroutines(concurrency).WithErrors().WithFirstError()
	for _, page := range c.Pages {
		page := page
		p.Go(func() error {
			if err := download(page); err != nil {
				return err
			}
			sizes <- page.Size
			return nil
		})
	}

	err := p.Wait()
	close(sizes)
	<-progressDone

	if err != nil {
		c.isDownloaded = mo.Some(false)
		return err
	}

	c.isDownloaded = mo.Some(!temp)
	return nil
}

// formattedName of the chapter according to the template in the config.
func (c *Chapter) formattedName() (name string) {
	template := viper.GetString(key.DownloaderChapterNameTemplate)

	var sourceName string
	if c.Source() != nil {
		sourceName = c.Source().Name()
	}

	// Ordered, single-style replacements without a per-call map.
	name = strings.ReplaceAll(template, "{manga}", c.Manga.Name)
	name = strings.ReplaceAll(name, "{chapter}", c.Name)
	name = strings.ReplaceAll(name, "{index}", strconv.Itoa(int(c.Index)))
	name = strings.ReplaceAll(name, "{padded-index}", fmt.Sprintf("%04d", c.Index))
	name = strings.ReplaceAll(name, "{chapters-count}", strconv.Itoa(len(c.Manga.Chapters)))
	name = strings.ReplaceAll(name, "{volume}", c.Volume)
	name = strings.ReplaceAll(name, "{source}", sourceName)

	return
}

// SizeHuman is the same as Size but returns a human-readable string.
func (c *Chapter) SizeHuman() string {
	return humanize.Bytes(c.size)
}

func (c *Chapter) Filename() (filename string) {
	filename = util.SanitizeFilename(c.formattedName())

	// plain format assumes that chapter is a directory with images
	// rather than a single file. So no need to add extension to it
	if f := viper.GetString(key.FormatsUse); f != constant.FormatPlain {
		return filename + "." + f
	}

	return
}

func (c *Chapter) IsDownloaded() bool {
	if c.isDownloaded.IsPresent() {
		return c.isDownloaded.MustGet()
	}

	path, _ := c.path(c.Manga.peekPath(), false)
	exists, _ := filesystem.Api().Exists(path)
	c.isDownloaded = mo.Some(exists)
	return exists
}

func (c *Chapter) path(relativeTo string, createVolumeDir bool) (path string, err error) {
	path = relativeTo
	if createVolumeDir {
		path = filepath.Join(path, util.SanitizeFilename(c.Volume))
		err = filesystem.Api().MkdirAll(path, os.ModePerm)
		if err != nil {
			return
		}
	}

	path = filepath.Join(path, c.Filename())
	return
}

func (c *Chapter) Path(temp bool) (path string, err error) {
	var manga string
	manga, err = c.Manga.Path(temp)
	if err != nil {
		return
	}

	return c.path(manga, c.Volume != "" && viper.GetBool(key.DownloaderCreateVolumeDir))
}

// Release frees the memory held by page contents after conversion.
// Must be called once the chapter is no longer needed in memory.
func (c *Chapter) Release() {
	for _, page := range c.Pages {
		if page != nil {
			page.Contents = nil
		}
	}
}

func (c *Chapter) Source() Source {
	return c.Manga.Source
}

func (c *Chapter) ComicInfo() *ComicInfo {
	var (
		day, month, year int
	)

	if viper.GetBool(key.MetadataComicInfoXMLAddDate) {
		if viper.GetBool(key.MetadataComicInfoXMLAlternativeDate) {
			// get current date
			t := time.Now()
			day = t.Day()
			month = int(t.Month())
			year = t.Year()
		} else if c.PublishDate.Year != 0 {
			// use chapter publish date from source if available
			day = c.PublishDate.Day
			month = c.PublishDate.Month
			year = c.PublishDate.Year
		} else {
			day = c.Manga.Metadata.StartDate.Day
			month = c.Manga.Metadata.StartDate.Month
			year = c.Manga.Metadata.StartDate.Year
		}
	} // empty dates will be omitted

	return &ComicInfo{
		XmlnsXsd: "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance",

		Title:      c.Name,
		Series:     c.Manga.Name,
		Number:     int(c.Index),
		Web:        c.URL,
		Genre:      strings.Join(c.Manga.Metadata.Genres, ","),
		PageCount:  len(c.Pages),
		Summary:    c.Manga.Metadata.Summary,
		Count:      c.Manga.Metadata.Chapters,
		Characters: strings.Join(c.Manga.Metadata.Characters, ","),
		Year:       year,
		Month:      month,
		Day:        day,
		Writer:     strings.Join(c.Manga.Metadata.Staff.Story, ","),
		Penciller:  strings.Join(c.Manga.Metadata.Staff.Art, ","),
		Letterer:   strings.Join(c.Manga.Metadata.Staff.Lettering, ","),
		Translator: strings.Join(c.Manga.Metadata.Staff.Translation, ","),
		Tags:       strings.Join(c.Manga.Metadata.Tags, ","),
		Notes:      "Downloaded with Koma. https://github.com/yukiteruamano/koma",
		Manga:      "YesAndRightToLeft",
	}
}
