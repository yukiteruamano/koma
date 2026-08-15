package where

import (
	"sync"

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"os"
	"path/filepath"
)

const EnvConfigPath = "KOMA_CONFIG_PATH"

// created memoizes already-created directories so hot paths do not repeat
// MkdirAll syscalls for the same path.
var created sync.Map

// mkdir creates a directory and all parent directories if they don't exist
// will return the path of the directory
func mkdir(path string) string {
	if _, ok := created.Load(path); ok {
		// the directory may have been deleted at runtime (e.g. the startup
		// temp cleanup); re-verify and recreate instead of trusting the memo
		if _, err := filesystem.Api().Stat(path); err == nil {
			return path
		}
		created.Delete(path)
	}

	lo.Must0(filesystem.Api().MkdirAll(path, os.ModePerm))
	created.Store(path, struct{}{})
	return path
}

// Config path
// Will create the directory if it doesn't exist
func Config() string {
	var path string

	if customDir, present := os.LookupEnv(EnvConfigPath); present {
		path = customDir
	} else {
		path = filepath.Join(lo.Must(os.UserConfigDir()), constant.Koma)
	}

	return mkdir(path)
}

// Sources path
// Will create the directory if it doesn't exist
func Sources() string {
	return mkdir(filepath.Join(Config(), "sources"))
}

func AnilistBinds() string {
	return filepath.Join(Config(), "anilist.json")
}

// Logs path
// Will create the directory if it doesn't exist
func Logs() string {
	return mkdir(filepath.Join(Config(), "logs"))
}

// Queries path
// Will create the directory if it doesn't exist
func Queries() string {
	return filepath.Join(Cache(), "queries.json")
}

// History path to the file
// Will create the directory if it doesn't exist
func History() string {
	return filepath.Join(Config(), "history.json")
}

// Downloads path
// Will create the directory if it doesn't exist
func Downloads() string {
	path, err := filepath.Abs(viper.GetString(key.DownloaderPath))

	if err != nil {
		path, err = os.Getwd()
		if err != nil {
			path = "."
		}
	}

	return mkdir(path)
}

// Cache path
// Will create the directory if it doesn't exist
func Cache() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = filepath.Join(".", "cache")
	}

	cacheDir = filepath.Join(cacheDir, constant.Koma)
	return mkdir(cacheDir)
}

// Temp path
// Will create the directory if it doesn't exist
func Temp() string {
	tempDir := filepath.Join(os.TempDir(), constant.Koma)
	return mkdir(tempDir)
}
