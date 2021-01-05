package gateway

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hanjingo/media_gateway/gateway/util"
	"go.uber.org/zap"
)

func Log() *zap.SugaredLogger {
	return util.Log()
}

type HttpHandler interface {
	Service()
	SetHandler(typ string, url string, h func(ctx *gin.Context)) error
}

type StreamPlayer interface {
	Play(ctx *gin.Context, f *util.FileInfo) error
}

type CacheMgr interface {
	Init()                                         // 初始化
	Run(wg *sync.WaitGroup)                        // 跑起来
	Save(ctx context.Context, f *util.FileInfo)    // 拉取文件
	SetHandler(typ string, h func(...interface{})) // 设置回调函数
	Close()                                        // 关闭缓存器
}

type Recorder interface {
	Init()
	Run(wg *sync.WaitGroup)
	Add(f *util.FileInfo)
	Get(ctx context.Context, f *util.FileInfo) error
	Close()
}
