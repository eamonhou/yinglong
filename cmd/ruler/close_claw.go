package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

/* 给 openclaw 设置执行命令的规则 */

func main() {
	// 获取程序所在目录的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取程序路径失败：%v\n", err)
		return
	}
	appDir := filepath.Dir(exePath)

	mdFilepath := filepath.Join("open_claw_ruler.md")

	mdFp, err := os.OpenFile(mdFilepath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	rulerContent, err := io.ReadAll(mdFp)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintf(os.Stdout, string(rulerContent), appDir)
}
