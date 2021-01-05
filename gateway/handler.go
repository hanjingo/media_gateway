package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hanjingo/media_gateway/gateway/util"
)

// 注册回调
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

/************************* video ***************************/
// http Get >> 127.0.0.1:10086/play/htm
func (g *Gate) onPlayHtml(ctx *gin.Context) {
	Log().Debugf("onPlayHtml")

	ctx.HTML(http.StatusOK, "play.html", nil)
}

// http Get >> 127.0.0.1:10086/video/player/htm?hash=qmxxxx
func (g *Gate) onPlayerHtml(ctx *gin.Context) {
	//ctx.HTML(http.StatusOK, "video.html", nil)
	Log().Debugf("onPlayer")

	hash := ctx.DefaultQuery("hash", "")
	Log().Debugf("hash=%s", hash)
	ctx.HTML(http.StatusOK, "player.tmpl", gin.H{
		"src": "/video/play?hash=" + hash,
	})

}

// http Get >> 127.0.0.1:10086/play?hash=qmxxxx
func (g *Gate) onPlay(ctx *gin.Context) {
	Log().Debugf("onPlay")

	info := &util.FileInfo{
		Hash: ctx.DefaultQuery("hash", ""),
	}
	Log().Debugf("play with hash:%s", info.Hash)
	if err := g.Record().Get(ctx, info); err != nil {
		Log().Errorf("play unexist hash:%s", info.Hash)
		return
	}
	// 不存在就重新拉取
	if !info.InStatus(util.Pulled) || !info.IsExist() {
		g.Cache().Save(context.Background(), info)
	}
	if err := g.Player().Play(ctx, info); err != nil {
		Log().Errorf("play:%s fail, err:%v", info.Hash, err)
		return
	}
}

/********************** file ******************************/
// http Get >> 127.0.0.1:10086/file/new/htm
func (g *Gate) onNewFileHtml(ctx *gin.Context) {
	Log().Debugf("onNewFileHtml")

	ctx.HTML(http.StatusOK, "new_file.html", nil)
}

// http Get >> 127.0.0.1:10086/file/search
// json: {tag:["标签1", "标签2"], page:1, limit:30}
func (g *Gate) onSearch(ctx *gin.Context) {
	Log().Debugf("onSearchResultHtml")

	req := &SearchReq{Tag: []string{}}
	if err := ctx.ShouldBind(req); err != nil {
		Log().Errorf("parse search req fail, err:%v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, SearchRsp{
			Result: ParseParamFail,
		})
	}
	back := util.GetHashByTag(context.Background(), util.PgAddr, req.Tag...)
	start := req.Page * req.Limit
	end := len(back) - 1
	if start >= end {
		start = end
	}
	detail := &SearchDetail{
		Tag:     req.Tag,
		Page:    req.Page,
		Limit:   req.Limit,
		MaxPage: len(back)/req.Limit + 1,
		Results: back[start:end],
	}
	data, _ := json.Marshal(detail)
	ctx.HTML(http.StatusOK, "search.tmpl", gin.H{
		"result": string(data),
	})
}

// http Post >> 127.0.0.1:10086/file/new
// json: {hash:"hash1", tag:["标签1", "标签2"]}
func (g *Gate) onNewFile(ctx *gin.Context) {
	Log().Debugf("onNewFile")

	req := &NewFileReq{}
	if err := ctx.ShouldBind(req); err != nil {
		Log().Errorf("parse new file req fail, err:%v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewFileRsp{
			Result: ParseParamFail,
		})
		return
	}

	info := util.NewFileInfo()
	info.Hash = req.Hash
	info.RecordTime = time.Now()
	info.SetStatus(util.Tracked)
	g.Cache().Save(context.Background(), info)

	if req.Tag != nil && len(req.Tag) > 0 {
		for _, tag := range req.Tag {
			util.AddHashTag(context.Background(), util.PgAddr, info.Hash, tag)
		}
	}

	Log().Debugf("new file with hash:%s, tags:%v", req.Hash, req.Tag)
	ctx.JSON(http.StatusOK, NewFileRsp{
		Result: Ok,
	})
}
