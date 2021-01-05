package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// IsExist ;判断文件/路径是否存在
func IsExist(arg string) bool {
	_, err := os.Stat(arg) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// GetCurrDir ;获得当前目录
func GetCurrDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return dir
}

// GetFileLength ;获得文件长度
func GetFileLength(arg string) int64 {
	stat, err := os.Stat(arg)
	if err != nil {
		return 0
	}
	return stat.Size()
}

// GetFileNameByExt 获得目录下指定类型的所有文件名(不含路径)，不包括文件夹
func GetFileNameByExt(rootPath string, exts ...string) []string {
	var extMap map[string]struct{}
	doAllFile := len(exts) == 0
	if !doAllFile {
		extMap := make(map[string]struct{})
		for _, ext := range exts {
			extMap[ext] = Void
		}
	}

	back := []string{}
	results, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return back
	}
	for _, info := range results {
		if info.IsDir() {
			continue
		}
		fname := info.Name()
		ext := filepath.Ext(fname)
		if !doAllFile {
			if _, ok := extMap[ext]; !ok {
				continue
			}
		}
		back = append(back, fname)
	}
	return back
}

// GetFileAbsPathByExt ;获得目录下指定类型的所有文件的绝对路径，不包括文件夹
func GetFileAbsPathByExt(rootPath string, exts ...string) []string {
	var extMap map[string]struct{}
	doAllFile := len(exts) == 0
	if !doAllFile {
		extMap := make(map[string]struct{})
		for _, ext := range exts {
			extMap[ext] = Void
		}
	}

	back := []string{}
	results, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return back
	}
	for _, info := range results {
		if info.IsDir() {
			continue
		}
		fname := filepath.Join(rootPath, info.Name())
		ext := filepath.Ext(fname)
		if !doAllFile {
			if _, ok := extMap[ext]; !ok {
				continue
			}
		}
		back = append(back, fname)
	}
	return back
}

// GetSubDirs ;获得目录下面的所有子目录(相对路径)
func GetSubDirs(rootPath string) []string {
	back := []string{}
	results, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return back
	}
	for _, p := range results {
		if !p.IsDir() {
			continue
		}
		back = append(back, p.Name())
	}
	return back
}

// MustCreateFile ;强行创建文件
func MustCreateFile(filePathName string) error {
	path, name := filepath.Split(filePathName)
	if path == "" || name == "" {
		return errors.New("path or file not exist")
	}
	if !IsExist(path) { //先创建路径
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	if !IsExist(filePathName) { //再创建文件
		fd, err := os.Create(filePathName)
		if err != nil {
			return err
		}
		if err := fd.Close(); err != nil {
			return err
		}
	}
	return nil
}

// MustCreateFile ;强行创建路径
func MustCreatePath(path string) error {
	if !IsExist(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
