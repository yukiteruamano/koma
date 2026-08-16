package update

import (
	"bytes"
	"encoding/json"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/converter/cbz"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/log"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/util"
	"os"
	"path/filepath"
	"strings"
)

func Metadata(mangaPath string) error {
	log.Infof("extracting series name from %s", mangaPath)
	name, err := GetName(mangaPath)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("extracted name: %s", name)
	log.Infof("finding %s on anilist", name)
	manga := &source.Manga{
		Name: name,
	}

	// will set new metadata from anilist
	err = manga.PopulateMetadata(func(string) {})
	if err != nil {
		log.Error(err)
		return err
	}

	chapters, err := getChapters(mangaPath)
	if err != nil {
		log.Error(err)
		return err
	}

	manga.Chapters = make([]*source.Chapter, 0)
	chaptersPaths := make(map[*source.Chapter]string)
	for _, chapter := range chapters {
		// since we are trying to update ComicInfo.xml here, we do not care about any other formats other than FormatCBZ
		if chapter.format != constant.FormatCBZ {
			continue
		}

		log.Infof("getting ComicInfoXML from %s", chapter.path)
		comicInfo, err := getComicInfoXML(chapter.path)
		if err != nil {
			log.Error(err)
			continue
		}

		chap := &source.Chapter{
			Name:  comicInfo.Title,
			Manga: manga,
			URL:   comicInfo.Web,
			Index: uint16(comicInfo.Number),
		}
		manga.Chapters = append(manga.Chapters, chap)
		chaptersPaths[chap] = chapter.path
	}

	// okay, we're ready to regenerate series.json and ComicInfo.xml now
	seriesJSON := manga.SeriesJSON()
	buf, err := json.Marshal(seriesJSON)
	if err != nil {
		log.Error(err)
		return err
	}

	// update series.json
	log.Info("updating series json")
	err = filesystem.Api().WriteFile(filepath.Join(mangaPath, "series.json"), buf, os.ModePerm)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("downloading new cover")
	// remove old cover(s).
	// even though DownloadCover() will overwrite previous one
	// there may be a sitation when new cover has a different extension
	// which would result having duplicates
	files, err := filesystem.Api().ReadDir(mangaPath)
	if err == nil {
		for _, file := range files {
			if util.FileStem(file.Name()) == "cover" {
				_ = filesystem.Api().Remove(filepath.Join(mangaPath, file.Name()))
			}
		}
	}
	err = manga.DownloadCover(true, mangaPath, func(string) {})
	if err != nil {
		log.Error(err)
	}

	log.Infof("updating ComicInfo.xml for %d chapters", len(manga.Chapters))
	for _, chapter := range manga.Chapters {
		path := chaptersPaths[chapter]

		// extract the pages into memory (memmap fs), then restore the OS fs
		// before rewriting the archive on disk
		if err := extractChapterPages(chapter, path); err != nil {
			log.Error(err)
			continue
		}

		log.Debugf("removing old %s", path)
		err := filesystem.Api().Remove(path)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Debugf("saving to %s", path)
		err = cbz.SaveTo(chapter, path)
		if err != nil {
			log.Error(err)
			return err
		}

		// pages are no longer needed once the archive is rewritten
		chapter.Release()
	}

	return nil
}

// extractChapterPages unzips a chapter into the memory filesystem and loads its
// page contents into memory. The OS filesystem is restored on every exit path.
func extractChapterPages(chapter *source.Chapter, path string) error {
	file, err := filesystem.Api().Open(path)
	if err != nil {
		return err
	}
	defer util.Ignore(file.Close)

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	filesystem.SetMemMapFs()
	defer filesystem.SetOsFs()

	if err := util.Unzip(file, stat.Size(), chapter.Name); err != nil {
		return err
	}

	files, err := filesystem.Api().ReadDir(chapter.Name)
	if err != nil {
		return err
	}

	for _, file := range files {
		// skip ComicInfo.xml
		if strings.HasSuffix(file.Name(), ".xml") {
			continue
		}

		image, err := filesystem.Api().ReadFile(filepath.Join(chapter.Name, file.Name()))
		if err != nil {
			return err
		}

		chapter.Pages = append(chapter.Pages, &source.Page{
			Chapter:   chapter,
			Size:      uint64(file.Size()),
			Index:     uint16(len(chapter.Pages)),
			Extension: filepath.Ext(file.Name()),
			Contents:  bytes.NewBuffer(image),
		})
	}

	return nil
}
