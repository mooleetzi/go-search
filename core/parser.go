package core

import (
	"flag"
	"fmt"
	"go-search/global"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"runtime"
)

// Parser 解析器
func Parser() *global.Config {

	var addr = flag.String("addr", "127.0.0.1:5678", "设置监听地址和端口")
	// 兼容windows
	dir := fmt.Sprintf(".%sdata", string(os.PathSeparator))

	var debug = flag.Bool("debug", true, "设置是否开启调试模式")

	var dataDir = flag.String("data", dir, "设置数据存储目录")

	var dictionaryPath = flag.String("dictionary", "searcher/words/data/dict.txt", "设置词典路径")

	var gomaxprocs = flag.Int("gomaxprocs", runtime.NumCPU()*2, "设置GOMAXPROCS")

	var enableGzip = flag.Bool("enableGzip", true, "是否开启gzip压缩")
	var timeout = flag.Int64("timeout", 10*60, "数据库超时关闭时间(秒)")

	var configPath = flag.String("config", "", "配置文件路径，配置此项其他参数忽略")

	flag.Parse()

	config := &global.Config{}

	if *configPath != "" {
		// 解析配置文件
		file, err := ioutil.ReadFile(*configPath)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(file, config)
		if err != nil {
			panic(err)
		}
		return config
	}
	config = &global.Config{
		Addr:       *addr,
		Data:       *dataDir,
		Debug:      *debug,
		Dictionary: *dictionaryPath,
		Gomaxprocs: *gomaxprocs,
		EnableGzip: *enableGzip,
		Timeout:    *timeout,
	}

	return config
}
