# Yinglong (应龙)

> 一个面向 AI Agent 的受控命令执行器：让模型能执行 shell，
> 但拿不到密码明文，也碰不到黑名单里的命令和文件。

`yinglong` 是一个用 Go 写的命令行包装器。它把你交给它的命令先做一遍
**占位符注入 → 命令审计 → 文件审计**，再交给 `sh -c` 执行，并把原始命令写进历史日志。

它最初是为了限制 `openclaw` 这类 Agent 框架对宿主机的权限而写的：
模型只被允许用 `yinglong "<命令>"` 这一种方式执行 shell，
密码用 `{{name}}` 占位由程序注入，模型无从得知明文。

---

## 功能一览

| 能力 | 实现 | 说明 |
| --- | --- | --- |
| 敏感信息注入 | `internal/injector.go` | `{{key}}` → `config/password.json` 中的真实值 |
| 命令黑名单 | `internal/auditor.go` | 与注入后的整条命令做精确匹配 |
| 文件黑名单 | `internal/auditor.go` + `executor.CheckShellToken` | 命令中每个"存在的文件/目录" token 都会被比对 |
| 执行日志 | `internal/logger.go` | 追加写入 `<程序目录>/log/yl_history.log` |
| 执行超时 | `cmd/command/main.go` | `context.WithTimeout`，默认 3 分钟 |
| 防递归 | `internal/executor.go` | `YINGLONG_NESTED=1` 环境变量 + 自身路径检测 |
| 配置管理 TUI | `cmd/user/main.go` | 交互式增删查改密码占位与黑名单 |
| 规则下发 | `cmd/ruler/close_claw.go` | 把 `open_claw_ruler.md` 渲染后输出给 Agent |

---

## 工作流程

一次 `yinglong "命令"` 的完整生命周期：

```
os.Args[1:]
   │
   ├─ 1. 防递归检查       YINGLONG_NESTED=1 ? / args[0] 是自己 ?
   │
   ├─ 2. 拼成 RawCommand
   │
   ├─ 3. 注入密码         {{password}} → xyz123
   │        └→ FinalCommand
   │
   ├─ 4. 命令审计         FinalCommand ∈ deny.commands ? → 拒绝
   │
   ├─ 5. Token 级文件审计  shlex.Split(FinalCommand)
   │        对每个 token:  ~ 展开 → os.Stat
   │        ├ 不存在（如 -l、grep 关键字）→ 跳过
   │        └ 存在（文件/目录）→ 比对 deny.files → 命中则拒绝
   │
   ├─ 6. 记录日志         logger.Print("info", RawCommand)
   │        （记录的是注入前的原文，所以日志里看到的是 {{password}}）
   │
   └─ 7. 执行             sh -c FinalCommand，stdout/stderr 直连终端
```

因为审计发生在**注入之后**，黑名单比对的是真实命令；
而日志记录的是**注入之前**的原文，所以历史日志里不会残留密码明文。

---

## 目录结构

```
yinglong/
├── cmd/
│   ├── command/main.go      # yinglong 主程序：初始化 + 审计 + 执行
│   ├── user/main.go         # 交互式配置管理工具（终端菜单）
│   ├── ruler/close_claw.go  # 生成 openclaw 规则文本
│   └── watcher/main.go      # 监控 openclaw 配置变化（TODO，仅有骨架）
├── internal/
│   ├── executor.go          # OnceExecutor：命令执行与 token 文件检查
│   ├── auditor.go           # Auditer 接口 + SimpleAuditor（黑名单）
│   ├── injector.go          # Injecter 接口 + SimpleInjector（占位符替换）
│   ├── setting.go           # 读取 password.json / deny.json
│   ├── logger.go            # SimpleLog：带时间戳的追加日志
│   └── util.go              # IsNil 等工具函数
├── cst/err.go               # 公共错误常量
├── config/
│   ├── password.json        # 占位符 → 真实值
│   └── deny.json            # 命令/文件黑名单
├── log/
│   ├── yl_history.log       # yinglong 执行记录
│   └── options.log          # 配置工具的操作审计记录
└── open_claw_ruler.md       # 下发给 Agent 的规则模板（%s 会被替换为程序目录）
```

`config/` 与 `log/` 都在 `.gitignore` 中，不入库。

---

## 快速开始

### 环境要求

- Go 1.24+（`go.mod` 声明 `go 1.24.2`）
- 类 Unix 系统（执行依赖 `sh -c`；Windows 下需自行改为 `cmd /c`）

### 构建

```bash
go build -o yinglong   ./cmd/command   # 主程序
go build -o yl-manager ./cmd/user      # 配置管理 TUI
go build -o yl-ruler   ./cmd/ruler     # 规则生成器
```

代码中实际使用的外部库只有 `github.com/google/shlex`（用于把命令串切分成 token）。
`go.mod` 里的 `mark3labs/mcp-go` 及其传递依赖是已移除的 `cmd/mcp` 遗留，
当前没有代码引用，可以用 `go mod tidy` 清理。

### 配置

程序会以**可执行文件所在目录**为基准（不是当前工作目录），
启动时自动创建 `log/` 和 `config/` 两个子目录。

`config/password.json` —— 占位符与真实值：

```json
{
  "password": "xyz123456789",
  "mysql_root": "s3cr3t"
}
```

命令中写 `{{password}}`，执行时会被替换为 `xyz123456789`。

`config/deny.json` —— 黑名单：

```json
{
  "commands": [
    "rm"
  ],
  "files": [
    "/home/user/.ssh/id_rsa",
    "/home/user/project/yinglong"
  ]
}
```

- `commands`：与**注入后的完整命令**做精确字符串比较，等号才算命中。
- `files`：与命令中每个**实际存在的**文件/目录 token 做精确字符串比较。

### 运行

```bash
./yinglong "ls -l"
./yinglong "mysql -u root -p{{mysql_root}} -e 'show databases'"
./yinglong "echo {{password}} | wc -c"
```

带管道的复杂命令可以直接用，因为最终是交给 `sh -c` 执行的。

---

## 配置管理工具

`cmd/user` 提供了一个基于 `bufio` 的终端菜单，避免手改 JSON 出错：

```
 1. 新增占位文本        7. 查看占位文本      10. 搜索占位文本
 2. 删除占位文本
 3. 新增文件黑名单      8. 查看文件黑名单    11. 搜索黑名单
 4. 删除文件黑名单
 5. 新增命令黑名单      9. 查看命令黑名单
 6. 删除命令黑名单
12. 查看操作日志（最近 20 条）   13. 帮助   0. 退出
```

注意：该工具读写的是**相对于当前工作目录**的 `config/password.json`，
与主程序"相对可执行文件目录"的约定不一致。请在项目根目录下运行。

每一次增删操作都会追加到 `log/options.log`：

```
2026-07-01 09:33:02 新增占位文本 shdjasdh=281938219jdas
```

---

## 规则下发（`cmd/ruler`）

`open_claw_ruler.md` 是写给上游 Agent 的行为约束模板，其中 `%s` 会被替换为
yinglong 所在目录：

```markdown
- 执行 shell 只能使用这种格式：yinglong "命令"
- 命令中的密码用 {{密码}} 占位，你无权获取明文
- yinglong 所在文件夹 <appDir> 你无权读取、修改
```

`cmd/ruler` 读取该模板并以 `fmt.Fprintf(os.Stdout, string(rulerContent), appDir)`
渲染输出，把结果粘贴/注入到 Agent 的 system prompt 或规则文件中即可。

---

## 设计要点

**接口先行，实现可替换。**
`internal` 中四个核心组件都定义了接口 —— `Executer`、`Auditer`、`Injecter`、`Settingger`，
构造函数通过 `NewXxx(cfg)` 返回接口类型。想换成"前缀匹配审计"或"从 Vault 取密码"，
只需新写一个实现，`cmd/command` 无需改动。

**并发读配置。**
密码与黑名单在 `main` 中用两个 goroutine + `WaitGroup` 并行加载，
读完后用 `internal.IsNil` 检查接口是否被成功赋值 —— 因为 goroutine 内出错时
只打印提示而没有返回错误，靠 nil 检查兜底。

**审计与日志的取值方向是相反的。**
审计用注入后的命令（要能拦住真实命令），日志用注入前的原文（不能泄漏密码）。
这个不对称是整个安全模型的关键，改动顺序会同时破坏两边。

**递归防护是双保险。**
`YINGLONG_NESTED=1` 覆盖 `yinglong "yinglong ..."` 这种嵌套，
`filepath.Base(args[0]) == selfName` 覆盖直接写自身路径的情况。

---

## 已知限制

当前实现偏"最小可用"，以下都是**有意保留的简化**或待修项，接入生产环境前需评估：

1. **黑名单是精确匹配，不是前缀/正则匹配。**
   `deny.commands` 里的 `rm` 拦不住 `rm -rf /`，也拦不住 `echo x; rm -rf /`。
   文件黑名单同理，且 `/home/u/f` 与 `./f` 会被当作不同条目。
2. **`ExecuteString` 尚未完成。**
   `internal/executor.go` 中该方法跳过了命令审计（步骤 1 被注释），
   而且日志记录的是**注入后**的 `finalCommand`，会把密码明文写进日志。
   该方法目前没有调用方（原 `cmd/mcp` 已移除），但接口仍在。
3. **超时是全局固定的 3 分钟**，不可按命令配置，也没有取消信号透传给子进程。
4. **`cmd/ruler` 存在缺陷**：以只写方式（`O_WRONLY`）打开模板后
   又对其调用 `io.ReadAll`，同一句柄无法读回内容。
5. **`cmd/watcher` 目前只是占位实现**，只打开 `~/.openclaw/openclaw.json` 后即返回，
   尚未实现文件变更监听。
6. **`setting.go` 中注释掉了 `defer fp.Close()`**，配置文件句柄在进程生命周期内不会释放
   （单次运行场景下影响可忽略）。
7. `cmd/command/main.go` 中审计器初始化失败时，错误提示误写成了"注入器初始化失败"。

---

## 路线图

- [ ] 把黑名单升级为前缀 / glob / 正则匹配，并规范化路径后再比对
- [ ] 完成 `ExecuteString` 的审计流程，修复密码明文入日志的问题
- [ ] 实现 `cmd/watcher`，监听 openclaw 配置变更并自动刷新规则
- [ ] 支持每命令独立超时，以及 SIGTERM 透传
- [ ] 修复 `cmd/ruler` 的文件读写顺序
- [ ] 统一 `cmd/user` 与主程序的配置路径基准

---

## 许可

未声明许可证。如需对外发布，请补充 `LICENSE` 文件。
