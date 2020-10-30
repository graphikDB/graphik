package main

import (
	"fmt"
	"github.com/autom8ter/graphik/graph/model"
	"time"
)

func main() {
	node := &model.Node{
		ID:   time.Now().String(),
		Type: "user",
	}
	var map1 = map[string]interface{}{
		"node": node,
	}
	node = nil
	for _, v := range map1 {
		fmt.Println(v)
	}

}
