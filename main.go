package main

import (
	"github.com/samber/lo"
	"github.com/yukiteruamano/koma/cmd"
	"github.com/yukiteruamano/koma/config"
	"github.com/yukiteruamano/koma/log"
)

func main() {
	lo.Must0(config.Setup())
	lo.Must0(log.Setup())
	cmd.Execute()
}
