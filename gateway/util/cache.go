package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	c "github.com/robfig/cron"
)

const (
	CacheMgrStatNew     uint32 = 1
	CacheMgrStatInit    uint32 = 2
	CacheMgrStatRunning uint32 = 3
	CacheMgrStatClosed  uint32 = 4
)

var (
	CachePulling  string = "Pulling"
	CachePulled   string = "Pulled"
	CachePullFail string = "PullFail"
	CacheExpired  string = "expired"
	CacheDeleted  string = "deleted"
)

type CacheMgr struct {
	mu   *sync.Mutex
	conf *CacheConfig

	currPath   string
	cache      chan *FileInfo
	stat       uint32
	cancel     context.CancelFunc
	cron       *c.Cron
	downloader *IpfsDownloader

	funcs map[string]func(...interface{})
}

func NewCacheMgr(conf *CacheConfig) *CacheMgr {
	back := &CacheMgr{
		mu:   new(sync.Mutex),
		conf: conf,

		cache:      make(chan *FileInfo, conf.Capa),
		stat:       CacheMgrStatNew,
		cron:       c.New(),
		downloader: NewIpfsDownloader(),

		funcs: make(map[string]func(...interface{})),
	}
	return back
}

func (s *CacheMgr) Init() {
	if !atomic.CompareAndSwapUint32(&s.stat, CacheMgrStatNew, CacheMgrStatInit) {
		return
	}
	s.getCurrPath()

	Log().Infof("init cache mgr success")
}

func (s *CacheMgr) Run(wg *sync.WaitGroup) {
	if !atomic.CompareAndSwapUint32(&s.stat, CacheMgrStatInit, CacheMgrStatRunning) {
		return
	}

	wg.Add(1)
	go func() {
		s.cron.Start()
		defer func() {
			s.cron.Stop()  // 停止计时
			close(s.cache) // 关闭管道
			Log().Infof("stop run cache mgr success")

			wg.Done()
		}()

		var ctx context.Context
		ctx, s.cancel = context.WithCancel(context.Background())
		for {
			select {
			case <-ctx.Done():
				for len(s.cache) > 0 {
					s.save(ctx, <-s.cache)
				}
				return
			case f := <-s.cache:
				if err := s.save(ctx, f); err != nil {
					Log().Errorf("save file:%s fail, err:%v", f.Hash, err)
				}
			}
		}
	}()

	if err := s.cron.AddFunc("0 0 */1 * * ?", s.gc); err != nil {
		Log().Errorf("add gc task fail, err:%v", err)
		return
	}

	Log().Infof("run cache mgr success")
}

func (s *CacheMgr) Save(ctx context.Context, f *FileInfo) {
	s.cache <- f
}

func (s *CacheMgr) SetHandler(typ string, h func(...interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.funcs[typ] = h
}

func (s *CacheMgr) save(ctx context.Context, f *FileInfo) error {
	return s.pull(ctx, f)
}

func (s *CacheMgr) pull(ctx context.Context, f *FileInfo) error {
	// 先从本地找
	if IsExist(f.filePathName) {
		return nil
	}

	// 再从ipfs找
	path := filepath.Join(s.getCurrPath(), f.Hash)
	s.call(CachePulling, f)
	if s.downloader == nil {
		s.call(CachePullFail, f)
		return fmt.Errorf("ipfs downloader not exist")
	}
	Log().Debugf("start to download file:%s with hash:%s", path, f.Hash)
	if err := s.downloader.Download(IpfsAddr, f.Hash, path); err != nil {
		s.call(CachePullFail, f)
		return fmt.Errorf("get hash:%v from ipfs addr:%s fail, err:%v", f.Hash, IpfsAddr, err)
	}
	s.call(CachePulled, f, path)
	return nil
}

func (s *CacheMgr) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *CacheMgr) call(typ string, args ...interface{}) {
	f, ok := s.funcs[typ]
	if !ok {
		Log().Errorf("fail to call:%s", typ)
		return
	}
	f(args...)
}

func (s *CacheMgr) getCurrPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	dirName := filepath.Base(s.currPath)
	old, err := time.Parse("20060102150405", dirName)
	if err != nil || !IsExist(s.currPath) || old.Add(time.Duration(1)*time.Hour).Before(now) {
		path := filepath.Join(s.conf.Path, time.Now().Format("20060102150405"))
		err := MustCreatePath(path)
		if err != nil {
			Log().Errorf("fail to create path:%v fail, err:%v", path, err)
		} else {
			s.currPath = path
			Log().Infof("create path:%s", s.currPath)
		}
	}
	return s.currPath
}

func (s *CacheMgr) gc() {
	s.getCurrPath()

	now := time.Now()
	// 扫描本地目录
	dirs := GetSubDirs(s.conf.Path)
	// 过期时长
	dur := time.Duration(s.conf.FileExpire) * time.Hour
	for _, dir := range dirs {
		t, err := time.Parse("20060102150405", dir)
		if err != nil {
			Log().Errorf("parse path:%s to time fail, err:%v", dir, err)
			continue
		}
		if t.Add(dur).Before(now) {
			target := filepath.Join(s.conf.Path, dir)
			// Expired
			files := GetFileNameByExt(target)
			for _, hash := range files {
				s.call(CacheExpired, hash)
			}

			// Delete
			if err := os.RemoveAll(target); err != nil {
				Log().Errorf("remove path:%s fail, err:%v", target, err)
				continue
			}
			for _, hash := range files {
				s.call(CacheDeleted, hash)
			}
			Log().Errorf("remove path:%s success", target)
		}
	}
}
