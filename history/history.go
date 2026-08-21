package history

import (
	"sync"

	"github.com/yukiteruamano/gache"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/integration"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/log"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/where"
)

var cacher = gache.New[map[string]*SavedChapter](
	&gache.Options{
		Path:       where.History(),
		FileSystem: &filesystem.GacheFs{},
	},
)

// mu serializes history reads and the Get-mutate-Set cycle so concurrent
// saves cannot lose updates or corrupt the history file.
var mu sync.RWMutex

// Get returns all chapters from the history file
func Get() (chapters map[string]*SavedChapter, err error) {
	mu.RLock()
	defer mu.RUnlock()

	return read()
}

// read returns the in-memory history map. Caller must hold the lock.
func read() (map[string]*SavedChapter, error) {
	cached, expired, err := cacher.Get()

	if err != nil {
		return nil, err
	}

	if expired || cached == nil {
		return make(map[string]*SavedChapter), nil
	}

	return cached, nil
}

// Save saves the chapter to the history file
func Save(chapter *source.Chapter) error {
	if viper.GetBool(key.AnilistEnable) {
		go func() {
			log.Info("Saving chapter to anilist")
			err := integration.Anilist.MarkRead(chapter)
			if err != nil {
				log.Warn("Saving chapter to anilist failed: " + err.Error())
			}
		}()
	}

	mu.Lock()
	defer mu.Unlock()

	saved, err := read()
	if err != nil {
		return err
	}

	savedChapter := newSavedChapter(chapter)
	saved[savedChapter.encode()] = savedChapter

	return cacher.Set(saved)
}

// Remove removes the chapter from the history file
func Remove(chapter *SavedChapter) error {
	mu.Lock()
	defer mu.Unlock()

	saved, err := read()
	if err != nil {
		return err
	}

	delete(saved, chapter.encode())

	return cacher.Set(saved)
}
