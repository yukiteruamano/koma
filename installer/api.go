package installer

import (
	"encoding/json"
	"fmt"
	"github.com/yukiteruamano/koma/network"
	"github.com/yukiteruamano/koma/util"
	"io"
	"net/http"
)

type GithubFile struct {
	Path string `json:"path"`
	Url  string `json:"url"`
}

type githubFilesCollector struct {
	user, repo, branch string

	Files []*GithubFile `json:"tree"`
}

func (g *githubFilesCollector) collect() error {
	if g.user == "" || g.repo == "" || g.branch == "" {
		return fmt.Errorf("user, repo and branch must be set")
	}

	if len(g.Files) > 0 {
		return nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", g.user, g.repo, g.branch)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := network.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get %s: %s", url, res.Status)
	}

	defer func() { _ = res.Body.Close() }()

	// decode the response
	var r []byte
	r, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	defer util.Ignore(res.Body.Close)

	err = json.Unmarshal(r, g)
	if err != nil {
		return err
	}

	return nil
}
