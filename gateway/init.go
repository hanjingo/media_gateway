package gateway

import (
	"context"

	"github.com/hanjingo/media_gateway/gateway/util"
)

func (gate *Gate) InitHttpHandler() {
	// file
	gate.Http().SetHandler("GET", "/file/new.html", gate.onNewFileHtml)
	gate.Http().SetHandler("POST", "/file/new", gate.onNewFile)
	gate.Http().SetHandler("GET", "/file/search", gate.onSearch)

	// video
	gate.Http().SetHandler("GET", "/video/play.html", gate.onPlayHtml)
	gate.Http().SetHandler("GET", "/video/player.html", gate.onPlayerHtml)
	gate.Http().SetHandler("GET", "/video/play", gate.onPlay)
}

func (gate *Gate) InitCacheMgr() {
	gate.Cache().Init()

	gate.Cache().SetHandler(util.CachePulling, func(args ...interface{}) {
		if len(args) == 0 {
			return
		}
		f, ok := args[0].(*util.FileInfo)
		if !ok {
			return
		}

		Log().Debugf("onPulling:%v", f.Hash)
		f.SetStatus(util.Pulling)
		if gate.Record() != nil {
			gate.Record().Add(f)
		}
	})

	gate.Cache().SetHandler(util.CachePulled, func(args ...interface{}) {
		if len(args) != 2 {
			return
		}
		f, ok := args[0].(*util.FileInfo)
		if !ok {
			return
		}
		path, ok := args[1].(string)
		if !ok {
			return
		}

		Log().Debugf("onPulled:%v", f.Hash)
		f.SetStatus(util.Pulled)
		f.SetFilePath(path)
		sec := GetConf().Cache.FileExpire * 60 * 60
		if err := f.SetRdsExpire(context.Background(), util.RdsAddr, sec); err != nil {
			Log().Errorf("set file:%s with redis expire fail, err:%v", f.Hash, err)
		}
		if gate.Record() != nil {
			gate.Record().Add(f)
		}
	})

	gate.Cache().SetHandler(util.CachePullFail, func(args ...interface{}) {
		if len(args) == 0 {
			return
		}
		f, ok := args[0].(*util.FileInfo)
		if !ok {
			return
		}

		Log().Debugf("onPullFail:%v", f.Hash)
		f.SetStatus(util.PullFail)
		if gate.Record() != nil {
			gate.Record().Add(f)
		}
	})

	gate.Cache().SetHandler(util.CacheExpired, func(args ...interface{}) {
		if len(args) == 0 {
			return
		}
		hash, ok := args[0].(string)
		if !ok {
			return
		}

		Log().Debugf("onExpired:%s", hash)
		f := &util.FileInfo{Hash: hash}
		if err := f.Get(context.Background()); err != nil {
			Log().Errorf("get fileinfo:%s fail, err:%v", hash, err)
			return
		}

		f.SetStatus(util.Expired)
		if gate.Record() != nil {
			gate.Record().Add(f)
		}
	})

	gate.Cache().SetHandler(util.CacheDeleted, func(args ...interface{}) {
		if len(args) == 0 {
			return
		}
		hash, ok := args[0].(string)
		if !ok {
			return
		}

		Log().Debugf("onDeleted:%s", hash)
		f := &util.FileInfo{Hash: hash}
		if err := f.Get(context.Background()); err != nil {
			Log().Errorf("get fileinfo:%s fail, err:%v", hash, err)
			return
		}

		f.SetStatus(util.Deleted)
		if gate.Record() != nil {
			gate.Record().Add(f)
		}
	})
}

func (gate *Gate) InitRecordMgr() {
	gate.Record().Init()

}
