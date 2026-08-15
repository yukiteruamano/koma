package update

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/util"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func getAnyChapterComicInfo(mangaPath string) (*source.ComicInfo, error) {
	// recursively search for .cbz files
	// find the first one and get the name from it
	var cbzFiles []string
	err := filepath.Walk(mangaPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".cbz") {
			cbzFiles = append(cbzFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(cbzFiles) == 0 {
		return nil, fmt.Errorf("no .cbz files found")
	}

	comicInfo, err := getComicInfoXML(cbzFiles[0])
	if err != nil {
		return nil, err
	}

	return comicInfo, nil
}

func getComicInfoXML(chapter string) (*source.ComicInfo, error) {
	if !strings.HasSuffix(chapter, ".cbz") {
		return nil, fmt.Errorf("chapter must be a .cbz file")
	}

	file, err := filesystem.Api().Open(chapter)
	if err != nil {
		return nil, err
	}
	defer util.Ignore(file.Close)

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Read only the ComicInfo.xml entry directly from the archive, without
	// unzipping every page or switching the global filesystem.
	reader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, err
	}

	for _, entry := range reader.File {
		if filepath.Base(entry.Name) != "ComicInfo.xml" {
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			return nil, err
		}
		contents, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}

		var comicInfo source.ComicInfo
		if err := xml.Unmarshal(contents, &comicInfo); err != nil {
			return nil, err
		}

		return &comicInfo, nil
	}

	return nil, fmt.Errorf("no ComicInfo.xml found in %s", chapter)
}
