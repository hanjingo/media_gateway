package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GinPlayer struct{}

func NewGinPlayer() *GinPlayer {
	return &GinPlayer{}
}

func (gp *GinPlayer) Play(c *gin.Context, f *FileInfo) error {
	if !f.InStatus(Pulled) {
		return fmt.Errorf("file status err, curr status:%s", FStatToStr(f.GetStatus()))
	}
	Log().Debugf("play file:%s", f.filePathName)
	if !IsExist(f.filePathName) {
		return fmt.Errorf("play with unexist file:%s", f.filePathName)
	}
	defer c.File(f.filePathName)

	// 处理range; 视频是分段播放的，有个start和end；根据start和end解析出size
	var start int64 = 0                              // 开始
	var end int64 = 0                                // 结束
	var length int64 = GetFileLength(f.filePathName) // 文件长度
	var size int64 = 0                               // 要播放的视频长度
	var err error
	var acceptValue = "bytes"     // Accept-Ranges 字段赋值
	var ctnTypValue = "video/mp4" // Content-Type 字段赋值
	var ctnLenValue = ""          // Content-length 字段赋值
	var ctnRangeValue = ""        // Content-Range 字段赋值

	rangeStr := c.GetHeader("range")
	Log().Debugf("rangeStr replace before:%s", rangeStr)
	rangeStr = strings.ReplaceAll(rangeStr, "bytes=", "")
	Log().Debugf("rangeStr replace after:%s", rangeStr)
	ranges := strings.Split(rangeStr, "-")
	switch len(ranges) {
	case 0: // 如果传过来没有开始结束，把整个文件发他
		size = length
	case 1: // 如果只有开始, 把当前位置到最后的视频发他
		if start, err = strconv.ParseInt(ranges[0], 10, 64); err != nil || start > length {
			start = length
			Log().Errorf("parse range start fail, err:%v", err)
		}
		size = length - start
	default: // 开始结束都有
		if start, err = strconv.ParseInt(ranges[0], 10, 64); err != nil || start > length {
			start = length
			Log().Errorf("parse range start fail, err:%v", err)
		}
		if end, err = strconv.ParseInt(ranges[1], 10, 64); err != nil || end < start || end > length {
			end = length
			Log().Errorf("parse range end fail, err:%v", err)
		}
		size = end - start + 1
	}
	ctnLenValue = fmt.Sprintf("%d", size)
	ctnRangeValue = fmt.Sprintf("bytes %d-%d/%d", start, end, length)

	c.Header("Accept-Ranges", acceptValue)
	c.Header("Content-Type", ctnTypValue)
	c.Header("Content-length", ctnLenValue)
	c.Header("Content-Range", ctnRangeValue)

	Log().Debugf("req header range:%s, ranges:%s, start:%d, end:%d, size:%d", rangeStr, ranges, start, end, size)
	return nil
}
