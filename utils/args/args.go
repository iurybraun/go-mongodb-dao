package args

import "flag"

var (
	IsRebuild      bool
	ConfigFilePath string
)

func init() {
	flag.BoolVar(&IsRebuild, "r", false, "rebuild database")
	flag.StringVar(&ConfigFilePath, "c", "", "the config file path")
	flag.Parse()
}
