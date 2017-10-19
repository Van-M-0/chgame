package configs

import (
	"exportor/proto"
	"encoding/csv"
	"strings"
	"fmt"
	"io/ioutil"
	"strconv"
)

var (
	configItemList []*proto.ItemConfig
)

func atoi(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func readItemsConfig() {
	configItemList = []*proto.ItemConfig{}
	itemConfigName := configDir + "items.csv"
	content, err := ioutil.ReadFile(itemConfigName)
	r := csv.NewReader(strings.NewReader(string(content)))
	ss,err := r.ReadAll()
	if err != nil {
		fmt.Println("read error, ", err)
		return
	}
	for i := 1; i < len(ss); i++ {
		record := ss[i]
		configItemList = append(configItemList, &proto.ItemConfig{
			Itemid: uint32(atoi(record[0])),
			Itemname: record[1],
			Category: int(atoi(record[2])),
			Nums: int(atoi(record[3])),
			Sell: int(atoi(record[4])),
			Buyvalue: int(atoi(record[5])),
			GameKind: int(atoi(record[6])),
			Description: record[7],
		})
	}
}

func GetItemsConfig() []proto.ItemConfig {
	l := []proto.ItemConfig{}
	for _, item := range configItemList {
		l = append(l, *item)
	}
	return l
}
