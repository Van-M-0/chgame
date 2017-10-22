package configs

import (
	"exportor/proto"
	"io/ioutil"
	"encoding/csv"
	"strings"
	"fmt"
)

var (
	gameLibs 		[]proto.GameLibItemP
)

func readGameLibs() {
	gameLibs = []proto.GameLibItemP{}
	gameConfig := configDir + "games.csv"
	content, err := ioutil.ReadFile(gameConfig)
	r := csv.NewReader(strings.NewReader(string(content)))
	ss,err := r.ReadAll()
	if err != nil {
		fmt.Println("read error, ", err)
		return
	}
	for i := 2; i < len(ss); i++ {
		record := ss[i]
		gameLibs = append(gameLibs, proto.GameLibItemP{
			GameLibItem: proto.GameLibItem{
				Id: atoi(record[0]),
				Name: record[1],
				Area: record[2],
				City: record[3],
				Province: record[4],
			},
			Pid: atoi(record[5]),
		})
	}
}

func GetGameLibs() [] proto.GameLibItemP {
	return gameLibs
}
