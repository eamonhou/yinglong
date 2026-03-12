package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"yinglong/internal"
)

func main() {
	ctx := context.Background()

	// 获取程序所在目录的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stdout, "获取程序路径失败：%v\n", err)
		return
	}
	appDir := filepath.Dir(exePath)

	// DEV
	// appDir = ""

	// 初始化日志类
	logFilepath := filepath.Join(appDir, "log", "yl_history.log")
	ylLogger := internal.NewLogger("simple").SetFile(logFilepath)
	defer ylLogger.Close()

	// 初始化设置类
	passFilepath := filepath.Join(appDir, "config", "password.json")
	ylSettingger := internal.NewSettingger("simple")

	startAppWg := &sync.WaitGroup{}

	var passInjector internal.Injecter
	startAppWg.Add(1)
	go func() {
		defer startAppWg.Done()
		// 密码注入器
		secrets, err := ylSettingger.ReadPasswordSetting(passFilepath)
		if err != nil {
			fmt.Fprintf(os.Stdout, "读取密码设置失败，%s", err.Error())
			return
		}
		passInjector = internal.NewInjector("simple").SetRelationMap(secrets)
	}()

	var commandAuditor internal.Auditer
	startAppWg.Add(1)
	go func() {
		defer startAppWg.Done()
		// 命令审核者
		denyFilepath := filepath.Join(appDir, "config", "deny.json")
		denyCommandList, denyFileList, err := ylSettingger.ReadDenySetting(denyFilepath)
		if err != nil {
			fmt.Fprintf(os.Stdout, "读取黑名单设置失败，%s", err.Error())
			return
		}
		commandAuditor = internal.NewAuditor("simple").
			SetDenyCommandList(denyCommandList).
			SetDenyFileList(denyFileList)
	}()

	startAppWg.Wait()

	// 设置命令执行超时时间
	ctx, cancel := context.WithTimeout(ctx, time.Second*60*3)
	defer cancel()

	// 命令执行器
	commandExecuter := internal.NewExecutor("simple").
		SetLogger(ylLogger).
		SetAuditor(commandAuditor).
		SetInjector(passInjector)

	commandExecuter.Execute(ctx)
}
