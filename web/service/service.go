package service

import (
	"go-search/global"
	"os"
	"runtime"
)

func Callback() map[string]interface{} {
	return map[string]interface{}{
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
		"cores":          runtime.NumCPU(),
		"version":        runtime.Version(),
		"goroutines":     runtime.NumGoroutine(),
		"dataPath":       global.CONFIG.Data,
		"dictionaryPath": global.CONFIG.Dictionary,
		"gomaxprocs":     runtime.NumCPU() * 2,
		"shard":          global.CONFIG.Shard,
		"executable":     os.Args[0],
		"dbs":            global.Container.GetDataBaseNumber(),
		// "indexCount":     global.Container.GetIndexCount(),
		// "documentCount":  global.Container.GetDocumentCount(),
		"pid":        os.Getpid(),
		"enableGzip": global.CONFIG.EnableGzip,
	}
}
