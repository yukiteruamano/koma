package open

import (
	"fmt"
	"github.com/yukiteruamano/koma/constant"
	"os/exec"
	"runtime"
)

var (
	errUnsupportedOS = fmt.Errorf("can't open on this OS: %s", runtime.GOOS)
)

func open(input string) (cmd *exec.Cmd, osSupported bool) {
	switch runtime.GOOS {
	case constant.Darwin:
		return exec.Command("open", input), true
	case constant.Linux:
		return exec.Command("xdg-open", input), true
	case constant.Android:
		return exec.Command("termux-open", input), true
	default:
		return nil, false
	}
}

func openWith(input, with string) (cmd *exec.Cmd, osSupported bool) {
	switch runtime.GOOS {
	case constant.Darwin:
		return exec.Command("open", "-a", with, input), true
	case constant.Linux:
		return exec.Command(with, input), true
	case constant.Android:
		return exec.Command("termux-open", "--choose", input), true
	default:
		return nil, false
	}
}
