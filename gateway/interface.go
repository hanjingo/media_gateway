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

// http回调器
type HttpHandler interface {
	Service()
	SetHandler(typ string, url string, h func(ctx *gin.Context)) error
}

// 流播放器
type StreamPlayer interface {
	Play(ctx *gin.Context, f *util.FileInfo) error
}

// 缓存管理器
type CacheMgr interface {
	Init()                                         // 初始化
	Run(wg *sync.WaitGroup)                        // 跑起来
	Save(ctx context.Context, f *util.FileInfo)    // 拉取文件
	SetHandler(typ string, h func(...interface{})) // 设置回调函数
	Close()                                        // 关闭缓存器
}

// 记录器
type Recorder interface {
	Init()
	Run(wg *sync.WaitGroup)
	Add(f *util.FileInfo)
	Get(ctx context.Context, f *util.FileInfo) error
	Close()
}

// 预览器
type Previewer interface {
}

// 分片机
type Spliter interface {
}

// 文件简介管理器
type SummaryMgr interface {
}

// 聊天管理器
type IM interface {
}

// 直播流服务器
type LiveServer interface {
}
