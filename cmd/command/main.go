package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// 获取程序所在目录的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取程序路径失败：%v\n", err)
		return
	}
	appDir := filepath.Dir(exePath)

	// 初始化日志
	logFilepath := filepath.Join(appDir, "log", "yl_history.log")
	logFp, err := os.OpenFile(logFilepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("打开日志文件失败：%v\n", err)
		return
	}

	// 加载密码配置
	var secrets = map[string]string{}
	passFilepath := filepath.Join(appDir, "config", "password.json")
	passFp, err := os.OpenFile(passFilepath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("打开密码配置文件失败：%v\n", err)
		return
	}
	passConfig, err := io.ReadAll(passFp)
	if err != nil {
		fmt.Printf("读取密码配置失败：%v\n", err)
		return
	}
	if err := json.Unmarshal(passConfig, &secrets); err != nil {
		fmt.Printf("密码配置格式错误：%v\n", err)
		return
	}

	//
	// 加载黑名单配置
	//
	var denies map[string]interface{}
	denyFilepath := filepath.Join(appDir, "config", "deny.json")
	denyFp, err := os.OpenFile(denyFilepath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("打开黑名单配置文件失败：%v\n", err)
		return
	}
	denyConfig, err := io.ReadAll(denyFp)
	if err != nil {
		fmt.Printf("读取黑名单配置失败：%v\n", err)
		return
	}
	if err := json.Unmarshal(denyConfig, &denies); err != nil {
		fmt.Printf("黑名单配置格式错误：%v\n", err)
		return
	}
	// 加载命令黑名单
	deniesCommandList := make(map[string]struct{})
	if denyCommandsItf, ok := denies["commands"]; ok {
		denyCommandsItfs, ok := denyCommandsItf.([]interface{})
		if !ok {
			fmt.Printf("黑名单配置错误\n")
			return
		}
		for _, denyCommandItf := range denyCommandsItfs {
			denyCommand, ok := denyCommandItf.(string)
			if !ok {
				fmt.Printf("黑名单配置项格式错误：%v\n", denyCommand)
				return
			}
			deniesCommandList[denyCommand] = struct{}{}
		}
	}
	// 加载文件黑名单
	deniesFileList := make(map[string]struct{})
	if denyFilesItf, ok := denies["files"]; ok {
		denyFilesItfs, ok := denyFilesItf.([]interface{})
		if !ok {
			fmt.Printf("黑名单配置错误\n")
			return
		}
		for _, denyFileItf := range denyFilesItfs {
			denyFile, ok := denyFileItf.(string)
			if !ok {
				fmt.Printf("黑名单配置项格式错误：%v\n", denyFile)
				return
			}
			deniesFileList[denyFile] = struct{}{}
		}
	}

	// 获取待执行命令
	if len(os.Args) < 2 {
		fmt.Printf("用法：yinglong <command>\n")
		return
	}
	args := os.Args[1:]

	// 检查要执行的命令是否在黑名单中
	if _, ok := deniesCommandList[args[0]]; ok {
		fmt.Printf("该命令已禁止执行：[%s]\n", args[0])
		return
	}

	rawCommand := strings.Join(args, " ")
	logFp.WriteString(fmt.Sprintf("执行命令：%s\n", rawCommand))

	// 注入密码
	finalCommand := rawCommand
	for k, v := range secrets {
		finalCommand = strings.ReplaceAll(finalCommand, "{{"+k+"}}", v)
	}

	// 创建一个命令对象
	// 使用 sh -c 可以让你执行带管道的复杂命令
	cmd := exec.Command("sh", "-c", finalCommand)

	// 将子进程的标准输出/错误直接关联到当前进程
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
	// fmt.Fprintf(os.Stdout, "Hello Yinglong\n")
}
