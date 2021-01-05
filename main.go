package main

import (
	"flag"
	"sync"

	"github.com/hanjingo/media_gateway/gateway"
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			gateway.Log().Fatalf("%v", p)
		}
	}()

	wg := new(sync.WaitGroup)

	c := ""
	flag.StringVar(&c, "c", "conf.json", "configure file abs path")
	flag.Parse()

	if err := gateway.GetConf().Load(c); err != nil {
		panic(err)
	}
	gateway.App().Init()
	gateway.App().Run(wg)

	wg.Wait()
	gateway.Log().Infof("gate exit!!!")
}
