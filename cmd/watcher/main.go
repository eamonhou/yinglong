package main

import (
	"fmt"
	"os"
	"path/filepath"
)

/**
 * TODO
 * 监控者
 * 启动一个脚本监控openclaw的配置文件变化
 *
 */

func main() {
	// 获取当前OS的用户目录
	userDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}

	openClawConfigFilepath := filepath.Join(userDir, ".openclaw", "openclaw.json")

	_, err = os.OpenFile(openClawConfigFilepath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

}
