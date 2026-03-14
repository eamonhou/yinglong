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
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	// 获取程序所在目录的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取程序路径失败：%v\n", err)
		return 1
	}
	appDir := filepath.Dir(exePath)

	// 【新增】确保必须的目录存在，防止运行崩溃
	os.MkdirAll(filepath.Join(appDir, "log"), 0755)
	os.MkdirAll(filepath.Join(appDir, "config"), 0755)

	// DEV
	// appDir = ""

	// 初始化日志类
	logFilepath := filepath.Join(appDir, "log", "yl_history.log")
	ylLogger, err := internal.NewSimpleLog(internal.LoggerConfig{
		Filepath: logFilepath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败\n")
	}
	defer ylLogger.Close()

	// 初始化设置类
	passFilepath := filepath.Join(appDir, "config", "password.json")
	ylSettingger := internal.NewSimpleSetting(internal.SettingConfig{})

	startAppWg := &sync.WaitGroup{}

	// 密码注入器
	var commandInjector internal.Injecter
	startAppWg.Add(1)
	go func() {
		defer startAppWg.Done()
		secrets, err := ylSettingger.ReadPasswordSetting(passFilepath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取密码设置失败，%s", err.Error())
			return
		}
		commandInjector = internal.NewInjector("simple").SetRelationMap(secrets)
	}()

	// 命令审核者
	var commandAuditor internal.Auditer
	startAppWg.Add(1)
	go func() {
		defer startAppWg.Done()
		denyFilepath := filepath.Join(appDir, "config", "deny.json")
		denyCommandList, denyFileList, err := ylSettingger.ReadDenySetting(denyFilepath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取黑名单设置失败，%s", err.Error())
			return
		}
		commandAuditor = internal.NewAuditor("simple").
			SetDenyCommandList(denyCommandList).
			SetDenyFileList(denyFileList)
	}()

	startAppWg.Wait()

	// 判断初始化是否成功
	if internal.IsNil(commandInjector) {
		fmt.Fprintf(os.Stderr, "注入器初始化失败\n")
		return 1
	}
	if internal.IsNil(commandAuditor) {
		fmt.Fprintf(os.Stderr, "注入器初始化失败\n")
		return 1
	}

	// 设置命令执行超时时间
	ctx, cancel := context.WithTimeout(ctx, time.Second*60*3)
	defer cancel()

	// 命令执行器
	commandExecuter, _ := internal.NewOnceExecutor(internal.ExecuterConfig{
		Logger:   ylLogger,
		Auditor:  commandAuditor,
		Injector: commandInjector,
	})
	if err := commandExecuter.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "执行命令失败: %v", err)
		return 1
	}
	return 0
}
