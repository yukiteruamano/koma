package history

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/source"
	"sync"
	"testing"
)

type testSource struct{}

func (testSource) Name() string {
	panic("")
}

func (testSource) Search(_ string) ([]*source.Manga, error) {
	panic("")
}

func (testSource) ChaptersOf(_ *source.Manga) ([]*source.Chapter, error) {
	panic("")
}

func (testSource) PagesOf(_ *source.Chapter) ([]*source.Page, error) {
	panic("")
}

func (testSource) ID() string {
	return "test source"
}

func init() {
	filesystem.SetMemMapFs()
}

func TestHistory(t *testing.T) {
	Convey("Given a chapter", t, func() {
		chapter := source.Chapter{
			Name:  "adwad",
			URL:   "dwaofa",
			Index: 42069,
			ID:    "fawfa",
			Pages: nil,
		}
		manga := source.Manga{
			Name:     "dawf",
			URL:      "fwa",
			Index:    1337,
			ID:       "wjakfkawgjj",
			Source:   testSource{},
			Chapters: []*source.Chapter{&chapter},
		}
		chapter.Manga = &manga

		Convey("When saving the chapter", func() {
			err := Save(&chapter)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)

				Convey("And the chapter should be saved", func() {
					chapters, err := Get()
					So(err, ShouldBeNil)
					So(len(chapters), ShouldBeGreaterThan, 0)
					So(chapters[fmt.Sprintf("%s (%s)", chapter.Manga.Name, chapter.Source().ID())].Name, ShouldEqual, chapter.Name)
				})
			})
		})
	})
}

func TestHistoryConcurrentSaves(t *testing.T) {
	const count = 20

	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			defer wg.Done()
			chapter := source.Chapter{
				Name:  fmt.Sprintf("chapter-%d", i),
				Index: uint16(i),
				ID:    fmt.Sprintf("id-%d", i),
				Manga: &source.Manga{Name: fmt.Sprintf("manga-%d", i), Source: testSource{}},
			}
			if err := Save(&chapter); err != nil {
				t.Errorf("Save(%d) failed: %v", i, err)
			}
		}()
	}
	wg.Wait()

	chapters, err := Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("manga-%d (test source)", i)
		if _, ok := chapters[key]; !ok {
			t.Errorf("concurrent save %d was lost", i)
		}
	}
}

func TestHistoryRemove(t *testing.T) {
	chapter := source.Chapter{
		Name:  "to-remove",
		Index: 1,
		ID:    "remove-id",
		Manga: &source.Manga{Name: "remove-manga", Source: testSource{}},
	}

	if err := Save(&chapter); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	chapters, err := Get()
	if err != nil {
		t.Fatal(err)
	}

	saved := chapters[fmt.Sprintf("%s (%s)", chapter.Manga.Name, chapter.Source().ID())]
	if saved == nil {
		t.Fatal("expected chapter to be saved before removal")
	}

	if err := Remove(saved); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	chapters, err = Get()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := chapters[fmt.Sprintf("%s (%s)", chapter.Manga.Name, chapter.Source().ID())]; ok {
		t.Error("chapter should have been removed")
	}
}
