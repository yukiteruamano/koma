package cbz

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/log"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/util"
	"io"
)

type CBZ struct{}

func New() *CBZ {
	return &CBZ{}
}

func (*CBZ) Save(chapter *source.Chapter) (string, error) {
	return save(chapter, false)
}

func (*CBZ) SaveTemp(chapter *source.Chapter) (string, error) {
	return save(chapter, true)
}

func save(chapter *source.Chapter, temp bool) (path string, err error) {
	path, err = chapter.Path(temp)
	if err != nil {
		return
	}

	err = SaveTo(chapter, path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func SaveTo(chapter *source.Chapter, to string) (err error) {
	cbzFile, err := filesystem.Api().Create(to)
	if err != nil {
		return err
	}

	defer util.Ignore(cbzFile.Close)
	defer func() {
		// do not leave a partial archive at the final path
		if err != nil {
			_ = filesystem.Api().Remove(to)
		}
	}()

	zipWriter := zip.NewWriter(cbzFile)
	defer util.Ignore(zipWriter.Close)

	for _, page := range chapter.Pages {
		if page.Contents == nil {
			log.Warnf("Skipping page #%d: contents are nil", page.Index)
			continue
		}
		if err = addToZip(zipWriter, page.Contents, page.Filename()); err != nil {
			return err
		}
	}

	if viper.GetBool(key.MetadataComicInfoXML) {
		comicInfo := chapter.ComicInfo()
		marshalled, err := xml.MarshalIndent(comicInfo, "", "  ")
		if err == nil {
			buf := bytes.NewBuffer(marshalled)
			err = addToZip(zipWriter, buf, "ComicInfo.xml")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func addToZip(writer *zip.Writer, file io.Reader, name string) error {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	}

	headerWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(headerWriter, file)
	return err
}
