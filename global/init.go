package global

import (
	"os"
	"path/filepath"
	"sync"
)

func init() {
	Init()
}

var RootDir string

var once = new(sync.Once)

// once.Do能保證方法只被執行一次(即使同時有多個goroutine調用)
func Init() {
	once.Do(func() {
		//這兩個func只會執行一次
		inferRootDir()
		initConfig()
	})
}

// inferRootDir 推斷出項目根目錄
func inferRootDir() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
		// 這裡要確保項目根目錄下存在template目錄
		if exists(d + "/template") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	RootDir = infer(cwd)
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
