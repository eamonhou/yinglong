package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

// Executer 命令执行方式
type Executer interface {
	Execute(ctx context.Context) error
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

// Execute 执行命令
//
//	@param ctx
//	@return error
func (obj *OnceExecutor) Execute(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("用法：yinglong \"<command>\"\n")
	}
	args := os.Args[1:]

	// 确保不是递归调用自己
	const recursiveEnv = "YINGLONG_NESTED"
	if os.Getenv(recursiveEnv) == "1" {
		return fmt.Errorf("拦截到递归调用：禁止在 yinglong 内部再次运行 yinglong\n")
	}

	selfPath, _ := os.Executable()
	selfName := filepath.Base(selfPath)
	if filepath.Base(args[0]) == selfName {
		return fmt.Errorf("禁止直接递归：命令中包含 %s\n", selfName)
	}

	// 生成完整命令
	obj.RawCommand = strings.Join(args, " ")

	// 注入密码
	var err error
	obj.FinalCommand, err = obj.injector.Inject(obj.RawCommand)
	if err != nil {
		return fmt.Errorf("密码注入失败\n")
	}

	// 审核命令
	if pass, _ := obj.auditor.AuditCommand(obj.FinalCommand); !pass {
		return fmt.Errorf("%s 命令禁止执行\n", obj.RawCommand)
	}

	// 审核文件
	commandTokens, err := shlex.Split(obj.FinalCommand)
	for _, token := range commandTokens {
		isFile, isDir, err := obj.CheckShellToken(ctx, token)
		if err != nil {
			return fmt.Errorf("检查命令token发生错误: %v\n", err)
		}
		if !isFile && !isDir { //不是文件
			continue
		}
		pass, err := obj.auditor.AuditFile(token)
		if err != nil {
			return fmt.Errorf("审核命令发生错误: %w\n", err)
		}
		if !pass {
			return fmt.Errorf("%s 不允许访问\n", token)
		}
	}

	// 创建一个命令对象
	// 使用 sh -c 可以让你执行带管道的复杂命令
	obj.Cmd = exec.CommandContext(ctx, "sh", "-c", obj.FinalCommand)

	// 注入防止递归的环境变量
	obj.Cmd.Env = append(os.Environ(), recursiveEnv+"=1")

	obj.logger.Print("info", obj.RawCommand)

	// 将子进程的标准输出/错误直接关联到当前进程
	obj.Cmd.Stdout = os.Stdout
	obj.Cmd.Stderr = os.Stderr

	err = obj.Cmd.Run()

	// 检查是否是超时导致退出
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("命令执行超时\n")
	}
	if err != nil {
		return err
	}
	return nil
}

// CheckShellToken 检查 token 是否为现存的文件或目录
// 返回值:
// isFile: true 表示是普通文件
// isDir:  true 表示是目录
// err:    如果 token 既不是现存文件也不是目录（比如 "-la", 或不存在的路径），则返回 err
func (obj *OnceExecutor) CheckShellToken(ctx context.Context, token string) (isFile bool, isDir bool, err error) {
	// 1. 处理 Shell 的波浪号 (~)
	if strings.HasPrefix(token, "~/") {
		if homeDir, e := os.UserHomeDir(); e == nil {
			token = filepath.Join(homeDir, token[2:])
		}
	} else if token == "~" {
		if homeDir, e := os.UserHomeDir(); e == nil {
			token = homeDir
		}
	}

	// 2. 检查文件/目录属性
	info, err := os.Stat(token)
	if err != nil {
		// 如果不存在或是个普通的命令参数(如 "-l")，os.Stat 会报错
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}

	// 3. 判断类型
	if info.IsDir() {
		return false, true, nil
	}
	return true, false, nil
}
