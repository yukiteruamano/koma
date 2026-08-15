package version

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/color"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/icon"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/style"
	"github.com/yukiteruamano/koma/util"
)

func Notify() {
	if !viper.GetBool(key.CliVersionCheck) {
		return
	}

	erase := util.PrintErasable(fmt.Sprintf("%s Checking if new version is available...", icon.Get(icon.Progress)))
	version, err := Latest()
	erase()
	if err == nil {
		comp, err := Compare(version, constant.Version)
		if err == nil && comp <= 0 {
			return
		}
	}

	fmt.Printf(`
%s New version is available %s %s
%s

`,
		style.Fg(color.Green)("▇▇▇"),
		style.Bold(version),
		style.Faint(fmt.Sprintf("(You're on %s)", constant.Version)),
		style.Faint("https://github.com/yukiteruamano/koma/releases/tag/v"+version),
	)

}
