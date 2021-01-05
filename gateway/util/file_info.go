package util

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	rds "github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
)

var (
	TBFile string = "file_info"
	TBTag  string = "tag_map"
)

const (
	dot string = ","
)

// 文件状态
const (
	Untracked uint32 = 0 // 未被添加状态
	Tracked   uint32 = 1 // 已添加任务
	Pulling   uint32 = 2 // 正在拉取
	Pulled    uint32 = 3 // 已拉取
	PullFail  uint32 = 4 // 拉取失败
	Expired   uint32 = 5 // 已过期
	Deleting  uint32 = 6 // 正在删除
	Deleted   uint32 = 7 // 已删除
)

var fstatusMap map[uint32]struct{}
var fstatusOnce = new(sync.Once)

func GetFStats() map[uint32]struct{} {
	fstatusOnce.Do(func() {
		fstatusMap = make(map[uint32]struct{})
		fstatusMap[Untracked] = Void
		fstatusMap[Tracked] = Void
		fstatusMap[Pulling] = Void
		fstatusMap[Pulled] = Void
		fstatusMap[PullFail] = Void
		fstatusMap[Expired] = Void
		fstatusMap[Deleting] = Void
		fstatusMap[Deleted] = Void
	})
	return fstatusMap
}

func FStatToStr(stat uint32) string {
	switch stat {
	case Untracked:
		return "UNTRACKED"
	case Tracked:
		return "TRACKED"
	case Pulling:
		return "PULLING"
	case Pulled:
		return "PULLED"
	case PullFail:
		return "PULL_FAIL"
	case Expired:
		return "EXPIRED"
	case Deleting:
		return "DELETING"
	case Deleted:
		return "DELETED"
	default:
		return "UNKNOWN"
	}
}

// FileInfo ;文件信息
type FileInfo struct {
	Hash         string    `json:"hash"`        // 文件hash
	RecordTime   time.Time `json:"record_time"` // 记录时间
	status       uint32    // 状态
	filePathName string    // 所属文件路径
}

func NewFileInfo() *FileInfo {
	return &FileInfo{status: Untracked, RecordTime: time.Now()}
}

func (f *FileInfo) SetStatus(stat uint32) {
	atomic.SwapUint32(&f.status, stat)
}

func (f *FileInfo) GetStatus() uint32 {
	return atomic.LoadUint32(&f.status)
}

func (f *FileInfo) InStatus(stat uint32) bool {
	return atomic.CompareAndSwapUint32(&f.status, stat, f.status)
}

func (f *FileInfo) IsExist() bool {
	return IsExist(f.filePathName)
}

func (f *FileInfo) SetFilePath(arg string) {
	f.filePathName = arg
}

func (f *FileInfo) maxIdx() int {
	return 4
}

func (f *FileInfo) Get(ctx context.Context) error {
	if err := f.GetRds(ctx, RdsAddr); err != nil {
		Log().Errorf("get redis fail, err:%v", err)
		if err = f.GetSql(ctx, PgAddr); err != nil {
			Log().Errorf("get sql fail, err:%v", err)
			return err
		}
		return f.SetRds(ctx, RdsAddr)
	}
	return nil
}

func (f *FileInfo) Set(ctx context.Context) error {
	if err := f.SetSql(ctx, PgAddr); err != nil {
		Log().Errorf("set sql fail, err:%v", err)
		return err
	}
	// 删除rds
	if err := f.DelRds(ctx, RdsAddr); err != nil {
		Log().Errorf("del redis fail, err:%v", err)
		return err
	}
	return nil
}

func (f *FileInfo) SetSql(ctx context.Context, dbAddr string) error {
	conn, err := pgx.Connect(ctx, dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	/*
		INSERT INTO gate_file(hash, record_time, status, abs_path)
			VALUES('hash1', '1.mp4', '2020-10-05 08:55:11', 1) ON conflict(hash) DO UPDATE
			SET status = 2
	*/
	sql := fmt.Sprintf(`INSERT INTO %s(hash, record_time, status, abs_path) 
							VALUES($1, $2, $3, $4) ON conflict(hash) DO UPDATE 
							SET status = %d, abs_path = '%s'`, TBFile, f.status, f.filePathName)
	_, err = conn.Exec(ctx, sql, f.Hash, f.RecordTime, f.status, f.filePathName)
	return err
}

func (f *FileInfo) GetSql(ctx context.Context, dbAddr string) error {
	conn, err := pgx.Connect(ctx, dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	sql := fmt.Sprintf(`SELECT hash, record_time, status, abs_path FROM %s WHERE 
							hash = '%s' LIMIT 1`, TBFile, f.Hash)
	return conn.QueryRow(ctx, sql).Scan(&f.Hash, &f.RecordTime, &f.status, &f.filePathName)
}

func (f *FileInfo) SetRds(ctx context.Context, dbAddr string) error {
	conn, err := rds.Dial("tcp", dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return err
	}

	value := f.Hash + dot + f.RecordTime.Format("2006-01-02 15:04:05") + dot +
		fmt.Sprintf("%d", f.status) + dot + f.filePathName
	if _, err = conn.Do("SET", f.Hash, value); err != nil {
		return err
	}
	return err
}

func (f *FileInfo) SetRdsExpire(ctx context.Context, dbAddr string, expireSec int) error {
	conn, err := rds.Dial("tcp", dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", f.Hash, expireSec)
	return err
}

func (f *FileInfo) GetRds(ctx context.Context, dbAddr string) error {
	conn, err := rds.Dial("tcp", dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return err
	}

	if f.Hash == "" {
		return errors.New("hash nil")
	}
	str, err := rds.String(conn.Do("GET", f.Hash))
	if err != nil {
		return err
	}
	results := strings.Split(str, dot)
	if len(results) < f.maxIdx() {
		return errors.New("data not complete")
	}
	f.Hash = results[0]
	if f.RecordTime, err = time.Parse("2006-01-02 15:04:05", results[1]); err != nil {
		return err
	}
	stat, err := strconv.ParseUint(results[2], 10, 64)
	if err != nil {
		return err
	}
	f.SetStatus(uint32(stat))
	f.filePathName = results[3]
	return nil
}

func (f *FileInfo) DelRds(ctx context.Context, dbAddr string) error {
	conn, err := rds.Dial("tcp", dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return err
	}

	_, err = conn.Do("DEL", f.Hash)
	return err
}
