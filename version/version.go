package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yukiteruamano/gache"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/network"
	"github.com/yukiteruamano/koma/util"
	"github.com/yukiteruamano/koma/where"
	"net/http"
	"path/filepath"
	"time"
)

var versionCacher = gache.New[string](&gache.Options{
	Path:       filepath.Join(where.Cache(), "version.json"),
	Lifetime:   time.Hour * 24 * 2,
	FileSystem: &filesystem.GacheFs{},
})

// Latest returns the latest version of koma.
// It will fetch the latest version from the GitHub API.
func Latest() (version string, err error) {
	ver, expired, err := versionCacher.Get()
	if err != nil {
		return "", err
	}

	if !expired && ver != "" {
		return ver, nil
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/yukiteruamano/koma/releases/latest", nil)
	if err != nil {
		return "", err
	}

	resp, err := network.Do(req)
	if err != nil {
		return "", err
	}

	defer util.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return
	}

	// remove the v from the tag name
	if release.TagName == "" {
		err = errors.New("empty tag name")
		return
	}

	version = release.TagName[1:]
	_ = versionCacher.Set(version)
	return
}
