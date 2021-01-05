package gateway

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hanjingo/media_gateway/gateway/util"
)

var g *Gate
var gOnce = new(sync.Once)

func App() *Gate {
	gOnce.Do(func() {
		g = newGate(GetConf())
	})
	return g
}

type Gate struct {
	bRun bool

	http   HttpHandler
	cache  CacheMgr
	player StreamPlayer
	record Recorder
}

func newGate(conf *Config) *Gate {
	conf.Check()
	back := &Gate{
		http:   util.NewHttpHandler(conf.Http),
		cache:  util.NewCacheMgr(conf.Cache),
		player: util.NewGinPlayer(),
		record: util.NewRecorder(conf.Record),
	}

	return back
}

func (gate *Gate) Init() {
	// init http handler
	gate.InitHttpHandler()

	// init cache
	gate.InitCacheMgr()

	// init recorder
	gate.InitRecordMgr()

	Log().Infof("init gate success")
}

func (gate *Gate) Run(wg *sync.WaitGroup) {
	if gate.bRun {
		return
	}
	gate.bRun = true

	gate.Cache().Run(wg)
	gate.Record().Run(wg)

	gate.Http().Service()
	go gate.listenSignal()

	Log().Infof("run gate success")
}

func (gate *Gate) Shutdown() {
	if !gate.bRun {
		return
	}

	gate.Cache().Close()
	gate.Record().Close()

	Log().Warnf("gate shutdown")
}

func (gate *Gate) Http() HttpHandler {
	return gate.http
}

func (gate *Gate) Cache() CacheMgr {
	return gate.cache
}

func (gate *Gate) Player() StreamPlayer {
	return gate.player
}

func (gate *Gate) Record() Recorder {
	return gate.record
}

func (gate *Gate) listenSignal() {
	sigChan := make(chan os.Signal, 1)
	defer func() {
		signal.Stop(sigChan)
		close(sigChan)
	}()

	// 开始监听Ctrl+C信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT:
				Log().Infof("wait to exit...")
				gate.Shutdown()
				return

			default:
				continue
			}
		}
	}
}
