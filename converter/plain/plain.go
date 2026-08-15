package plain

import (
	"github.com/sourcegraph/conc/pool"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/log"
	"github.com/yukiteruamano/koma/source"
	"io"
	"os"
	"path/filepath"
)

type Plain struct{}

func New() *Plain {
	return &Plain{}
}

func (*Plain) Save(chapter *source.Chapter) (string, error) {
	return save(chapter, false)
}

func (*Plain) SaveTemp(chapter *source.Chapter) (string, error) {
	return save(chapter, true)
}

func save(chapter *source.Chapter, temp bool) (path string, err error) {
	path, err = chapter.Path(temp)
	if err != nil {
		return
	}

	err = filesystem.Api().MkdirAll(path, os.ModePerm)
	if err != nil {
		return
	}

	p := pool.New().WithMaxGoroutines(8).WithErrors().WithFirstError()
	for _, page := range chapter.Pages {
		page := page
		p.Go(func() error {
			return savePage(page, path)
		})
	}

	err = p.Wait()
	return
}

func savePage(page *source.Page, to string) error {
	if page.Contents == nil {
		log.Warnf("Skipping page #%d: contents are nil", page.Index)
		return nil
	}

	dst := filepath.Join(to, page.Filename())
	file, err := filesystem.Api().Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if _, err = io.Copy(file, page.Contents); err != nil {
		// do not leave a partial page at the final path
		_ = filesystem.Api().Remove(dst)
		return err
	}

	_ = page.Close()
	return nil
}
