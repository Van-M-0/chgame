package configs

var (
	configDir      string
)

func LoadConfigs(dir string) {
	configDir = dir
	readItemsConfig()
	readGameModules()
	readGameLibs()
}
