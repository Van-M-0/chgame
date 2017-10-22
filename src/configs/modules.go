package configs

import (
	"exportor/defines"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Mods struct {
	Modules		[]defines.GameModuleConfig
}

var (
	modules 	Mods
)

func readGameModules() {
	mods := configDir + "modules"
	content, err := ioutil.ReadFile(mods)
	if err != nil {
		panic("read config file error " + err.Error())
	}

	if err := json.Unmarshal(content, &modules); err != nil {
		panic(fmt.Errorf("config file invalid err %v", err).Error())
	}
}

func GetKindConfig(kind int) *defines.GameModuleConfig {
	for _, m := range modules.Modules {
		if m.Kind == kind {
			return &m
		}
	}
	return nil
}
