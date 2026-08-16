package anilist

import (
	"sync"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
)

type Anilist struct {
	tokenMu sync.RWMutex
	token   string
}

// New cereates a new Anilist integration instance
func New() *Anilist {
	return &Anilist{}
}

func (a *Anilist) getToken() string {
	a.tokenMu.RLock()
	defer a.tokenMu.RUnlock()
	return a.token
}

func (a *Anilist) setToken(token string) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	a.token = token
}

func (a *Anilist) id() string {
	return viper.GetString(key.AnilistID)
}

func (a *Anilist) secret() string {
	return viper.GetString(key.AnilistSecret)
}

func (a *Anilist) code() string {
	return viper.GetString(key.AnilistCode)
}

// AuthURL returns the URL to authenticate with Anilist
func (a *Anilist) AuthURL() string {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=" + a.id() + "&response_type=code&redirect_uri=https://anilist.co/api/v2/oauth/pin"
}
