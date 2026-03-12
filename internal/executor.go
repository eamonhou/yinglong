package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Executer 命令执行方式
type Executer interface {
	Execute() error
	SetLogger(logger Logger) Executer
	SetAuditor(auditor Auditer) Executer
	SetInjector(injector Injecter) Executer
}

type OnceExecutor struct {
	Cmd          *exec.Cmd
	RawCommand   string   // 原始命令
	FinalCommand string   // 最终命令
	logger       Logger   // 日志
	auditor      Auditer  // 审核者
	injector     Injecter // 注入者
}

func NewExecutor(kind string) Executer {
	var executer Executer
	switch kind {
	case "simple":
		executer = &OnceExecutor{}
	default:
		executer = &OnceExecutor{}
	}
	return executer
}

func (obj *OnceExecutor) SetLogger(logger Logger) Executer {
	obj.logger = logger
	return obj
}

func (obj *OnceExecutor) SetAuditor(auditor Auditer) Executer {
	obj.auditor = auditor
	return obj
}

func (obj *OnceExecutor) SetInjector(injector Injecter) Executer {
	obj.injector = injector
	return obj
}

func (obj *OnceExecutor) Execute() error {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stdout, "用法：yinglong <command>\n")
		return nil
	}
	args := os.Args[1:]

	// 确保不是递归调用自己
	const recursiveEnv = "YINGLONG_NESTED"
	if os.Getenv(recursiveEnv) == "1" {
		fmt.Fprintf(os.Stdout, "拦截到递归调用：禁止在 yinglong 内部再次运行 yinglong\n")
		return nil
	}

	selfPath, _ := os.Executable()
	selfName := filepath.Base(selfPath)
	if filepath.Base(args[0]) == selfName {
		fmt.Fprintf(os.Stdout, "禁止直接递归：命令中包含 %s\n", selfName)
		return nil
	}

	// 生成完整命令
	obj.RawCommand = strings.Join(args, " ")

	// 审核命令
	if pass, _ := obj.auditor.AuditCommand(obj.RawCommand); !pass {
		fmt.Fprintf(os.Stdout, "[%s] 命令禁止执行", obj.RawCommand)
		return nil
	}

	// 注入密码
	obj.FinalCommand, _ = obj.injector.Inject(obj.RawCommand)

	// 创建一个命令对象
	// 使用 sh -c 可以让你执行带管道的复杂命令
	obj.Cmd = exec.Command("sh", "-c", obj.FinalCommand)

	// 注入防止递归的环境变量
	obj.Cmd.Env = append(os.Environ(), recursiveEnv+"=1")

	obj.logger.Print("info", obj.FinalCommand)

	// 将子进程的标准输出/错误直接关联到当前进程
	obj.Cmd.Stdout = os.Stdout
	obj.Cmd.Stderr = os.Stderr

	err := obj.Cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
	return nil
}
