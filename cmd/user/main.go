package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	passwordFile = "config/password.json"
	denyFile     = "config/deny.json"
	logFile      = "./log/options.log"
)

var (
	optionLogFp *os.File
)

// 占位列表
type Passwords map[string]string

// 黑名单列表
type DenyList struct {
	Files    []string `json:"files"`
	Commands []string `json:"commands"`
}

func main() {
	var err error
	optionLogFp, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("打开操作日志文件失败，%s", err.Error())
		return
	}
	defer optionLogFp.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		showMenu()
		fmt.Print("\n请输入操作编号: ")
		options, _ := reader.ReadString('\n')
		options = strings.TrimSpace(options)
		fmt.Println()
		switch options {
		case "1":
			addPassword(reader)
		case "2":
			deletePassword(reader)
		case "3":
			addFileDeny(reader)
		case "4":
			deleteFileDeny(reader)
		case "5":
			addCommandDeny(reader)
		case "6":
			deleteCommandDeny(reader)
		case "7":
			PasswordsList()
		case "8":
			fileDenyList()
		case "9":
			CommandDenyList()
		case "10":
			findPasswords(reader)
		case "11":
			findDeny(reader)
		case "12":
			optionLogList()
		case "13":
			showHelp()
		case "0":
			fmt.Println("==========================================================")
			fmt.Println("已退出系统")
			fmt.Println("==========================================================")
			return
		default:
			fmt.Printf("无效的操作序号，请重新输入\n")
		}
		if options != "0" {
			fmt.Print("\n按回车键继续...")
			reader.ReadString('\n')
		}
	}
}

func showMenu() {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println("==========================================================")
	fmt.Printf(" 命令行安全管理与执行控制软件\n")
	fmt.Println("==========================================================")
	fmt.Println("\nx")
	fmt.Println(" 1. 新增占位文本")
	fmt.Println(" 2. 删除占位文本")
	fmt.Println(" 7. 查看占位文本")
	fmt.Println(" 10. 搜索占位文本")
	fmt.Println("\n黑名单管理")
	fmt.Println(" 3. 新增文件黑名单")
	fmt.Println(" 4. 删除文件黑名单")
	fmt.Println(" 5. 新增命令黑名单")
	fmt.Println(" 6. 删除命令黑名单")
	fmt.Println(" 8. 查看文件黑名单")
	fmt.Println(" 9. 查看命令黑名单")
	fmt.Println(" 11. 搜索黑名单")
	fmt.Println("\n系统功能")
	fmt.Println(" 12. 查看操作日志")
	fmt.Println(" 13. 帮助")
	fmt.Println(" 0. 退出")
	fmt.Println("==========================================================")
}

// traceRecord 记录操作日志
//
//	@param op
func traceRecord(content string) {
	fmt.Fprintf(optionLogFp, "%s %s\n", time.Now().Format("2006-01-02 15:04:05"), content)
}

// loadPasswords 加载占位文本
//
//	@return Passwords
//	@return error
func loadPasswords() (Passwords, error) {
	var data Passwords
	content, err := ioutil.ReadFile(passwordFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(Passwords), nil
		}
		return nil, err
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// savePasswords 保存占位文本
//
//	@param data
//	@return error
func savePasswords(data Passwords) error {
	os.MkdirAll("config", 0755)
	bs, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(passwordFile, bs, 0644)
}

// addPassword 新增占位文本
//
//	@param reader
func addPassword(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 新增占位文本\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入占位文本 (格式: key=value): ")
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		fmt.Printf("无效格式 key=value\n")
		return
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		fmt.Printf("key 和 value 不能为空\n")
		return
	}
	passwords, err := loadPasswords()
	if err != nil {
		fmt.Printf("读取密码文件失败: %v\n", err)
		return
	}
	if _, exists := passwords[key]; exists {
		fmt.Printf("key '%s' 已存在，将覆盖旧值\n", key)
	}

	passwords[key] = value
	err = savePasswords(passwords)
	if err != nil {
		fmt.Printf("保存密码文件失败: %v\n", err)
		return
	}
	fmt.Printf("成功添加占位文本: %s = %s\n", key, value)
	traceRecord(fmt.Sprintf("新增占位文本 %s=%s", key, value))
}

// deletePassword 删除占位文本
//
//	@param reader
func deletePassword(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 删除占位文本\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入要删除的 key: ")
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
	if key == "" {
		fmt.Printf("key 不能为空\n")
		return
	}
	passwords, err := loadPasswords()
	if err != nil {
		fmt.Printf("读取密码文件失败: %v\n", err)
		return
	}
	if _, ok := passwords[key]; !ok {
		fmt.Printf("key '%s' 不存在\n", key)
		return
	}
	delete(passwords, key)
	err = savePasswords(passwords)
	if err != nil {
		fmt.Printf("保存密码文件失败: %v\n", err)
		return
	}
	fmt.Printf("成功删除占位文本: %s\n", key)
	traceRecord(fmt.Sprintf("删除占位文本 %s", key))
}

// readDenyList 读取黑名单列表
//
//	@return DenyList
//	@return error
func readDenyList() (DenyList, error) {
	var deny DenyList
	content, err := ioutil.ReadFile(denyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return DenyList{Files: []string{}, Commands: []string{}}, nil
		}
		return deny, err
	}
	err = json.Unmarshal(content, &deny)
	if err != nil {
		return deny, err
	}
	return deny, nil
}

// saveDenyList 保存黑名单列表
//
//	@param deny
//	@return error
func saveDenyList(deny DenyList) error {
	// 确保目录存在
	os.MkdirAll("config", 0755)
	bs, err := json.MarshalIndent(deny, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(denyFile, bs, 0644)
}

// addFileDeny 增加文件黑名单
//
//	@param reader
func addFileDeny(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 新增文件黑名单\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入文件路径: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	if path == "" {
		fmt.Printf("文件路径不能为空\n")
		return
	}
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	for _, v := range deny.Files {
		if v == path {
			fmt.Printf("该文件路径已在黑名单中\n")
			return
		}
	}
	deny.Files = append(deny.Files, path)
	err = saveDenyList(deny)
	if err != nil {
		fmt.Printf("保存黑名单失败: %v\n", err)
		return
	}
	fmt.Printf("成功添加文件黑名单: %s\n", path)
	traceRecord(fmt.Sprintf("新增文件黑名单 %s", path))
}

// deleteFileDeny 删除文件黑名单
//
//	@param reader
func deleteFileDeny(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 删除文件黑名单\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入要删除的文件路径: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	index := -1
	for i, v := range deny.Files {
		if v == path {
			index = i
			break
		}
	}
	if index == -1 {
		fmt.Printf("该文件路径不在黑名单中\n")
		return
	}
	deny.Files = append(deny.Files[:index], deny.Files[index+1:]...)
	err = saveDenyList(deny)
	if err != nil {
		fmt.Printf("保存黑名单失败: %v\n", err)
		return
	}
	fmt.Printf("成功删除文件黑名单: %s\n", path)
	traceRecord(fmt.Sprintf("删除文件黑名单 %s", path))
}

// addCommandDeny 新增命令黑名单
//
//	@param reader
func addCommandDeny(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 新增命令黑名单\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入命令: ")
	cmd, _ := reader.ReadString('\n')
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		fmt.Printf("命令不能为空\n")
		return
	}
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	for _, v := range deny.Commands {
		if v == cmd {
			fmt.Printf("该命令已在黑名单中\n")
			return
		}
	}
	deny.Commands = append(deny.Commands, cmd)
	err = saveDenyList(deny)
	if err != nil {
		fmt.Printf("保存黑名单失败: %v\n", err)
		return
	}
	fmt.Printf("成功添加命令黑名单: %s\n", cmd)
	traceRecord(fmt.Sprintf("新增命令黑名单 %s", cmd))
}

// deleteCommandDeny 删除文件黑名单
//
//	@param reader
func deleteCommandDeny(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 删除命令黑名单\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入要删除的命令: ")
	cmd, _ := reader.ReadString('\n')
	cmd = strings.TrimSpace(cmd)
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	idx := -1
	for i, v := range deny.Commands {
		if v == cmd {
			idx = i
			break
		}
	}
	if idx == -1 {
		fmt.Printf("该命令不在黑名单中\n")
		return
	}
	deny.Commands = append(deny.Commands[:idx], deny.Commands[idx+1:]...)
	err = saveDenyList(deny)
	if err != nil {
		fmt.Printf("保存黑名单失败: %v\n", err)
		return
	}
	fmt.Printf("成功删除命令黑名单: %s\n", cmd)
	traceRecord(fmt.Sprintf("删除命令黑名单 %s", cmd))
}

// PasswordsList 占位文本列表
func PasswordsList() {
	fmt.Println("==========================================================")
	fmt.Printf(" 占位文本列表\n")
	fmt.Println("==========================================================")
	passwords, err := loadPasswords()
	if err != nil {
		fmt.Printf("读取密码文件失败: %v\n", err)
		return
	}
	if len(passwords) == 0 {
		fmt.Printf("占位文本为空\n")
		return
	}
	fmt.Printf("共 %d 个占位文本:\n\n", len(passwords))
	i := 1
	for k, v := range passwords {
		fmt.Printf(" %d. %s = %s\n", i, k, v)
		i++
	}
}

// fileDenyList 文件黑名单列表
func fileDenyList() {
	fmt.Println("==========================================================")
	fmt.Printf(" 文件黑名单列表\n")
	fmt.Println("==========================================================")
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	if len(deny.Files) == 0 {
		fmt.Printf("文件黑名单为空\n")
		return
	}
	fmt.Printf("共 %d 个文件黑名单:\n\n", len(deny.Files))
	for i, f := range deny.Files {
		fmt.Printf(" %d. %s\n", i+1, f)
	}
}

// CommandDenyList 命令黑名单列表
func CommandDenyList() {
	fmt.Println("==========================================================")
	fmt.Printf(" 命令黑名单列表\n")
	fmt.Println("==========================================================")
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	if len(deny.Commands) == 0 {
		fmt.Printf("命令黑名单为空\n")
		return
	}
	fmt.Printf("共 %d 个命令黑名单:\n\n", len(deny.Commands))
	for i, c := range deny.Commands {
		fmt.Printf(" %d. %s\n", i+1, c)
	}
}

// findPasswords 占位文本查询
//
//	@param reader
func findPasswords(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 搜索占位文本\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入搜索关键字: ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		fmt.Printf("关键字不能为空\n")
		return
	}
	passwords, err := loadPasswords()
	if err != nil {
		fmt.Printf("读取密码文件失败: %v\n", err)
		return
	}
	found := false
	matcheNum := 0
	for k, v := range passwords {
		if strings.Contains(k, keyword) || strings.Contains(v, keyword) {
			if !found {
				fmt.Printf("\n匹配关键字 '%s' 的占位文本:\n\n", keyword)
				found = true
			}
			matcheNum++
			fmt.Printf(" %d. %s = %s\n", matcheNum, k, v)
		}
	}
	if !found {
		fmt.Printf("未找到包含 '%s' 的占位文本\n", keyword)
	} else {
		fmt.Printf("\n共找到 %d 个匹配项\n", matcheNum)
	}
}

// findDeny 黑名单查询
//
//	@param reader
func findDeny(reader *bufio.Reader) {
	fmt.Println("==========================================================")
	fmt.Printf(" 搜索黑名单\n")
	fmt.Println("==========================================================")
	fmt.Print("请输入搜索关键字: ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		fmt.Printf("关键字不能为空\n")
		return
	}
	deny, err := readDenyList()
	if err != nil {
		fmt.Printf("读取黑名单失败: %v\n", err)
		return
	}
	matcheNum := 0
	cmdMatcheNum := 0
	for _, f := range deny.Files {
		if strings.Contains(f, keyword) {
			if matcheNum == 0 {
				fmt.Printf("\n匹配关键字 '%s' 的文件黑名单:\n\n", keyword)
			}
			matcheNum++
			fmt.Printf(" %d. %s\n", matcheNum, f)
		}
	}
	for _, c := range deny.Commands {
		if strings.Contains(c, keyword) {
			if cmdMatcheNum == 0 {
				fmt.Printf("\n匹配关键字 '%s' 的命令黑名单:\n\n", keyword)
			}
			cmdMatcheNum++
			fmt.Printf(" %d. %s\n", cmdMatcheNum, c)
		}
	}
	if matcheNum == 0 && cmdMatcheNum == 0 {
		fmt.Printf("没有找到包含 '%s' 的黑名单项\n", keyword)
	} else {
		fmt.Printf("\n共找到 %d 个匹配项 (文件: %d, 命令: %d)\n", matcheNum+cmdMatcheNum, matcheNum, cmdMatcheNum)
	}
}

// optionLogList 读取操作日志列表
func optionLogList() {
	fmt.Println("==========================================================")
	fmt.Printf(" 操作日志\n")
	fmt.Println("==========================================================")
	content, err := ioutil.ReadFile(logFile)
	if err != nil {
		fmt.Printf("无操作日志\n")
		return
	}
	lines := strings.Split(string(content), "\n")
	var validLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, line)
		}
	}
	if len(validLines) == 0 {
		fmt.Printf("无操作日志\n")
		return
	}
	fmt.Printf("最近操作日志 (只显示 20 条):\n\n")
	start := 0
	if len(validLines) > 20 {
		start = len(validLines) - 20
	}
	for i, line := range validLines[start:] {
		fmt.Printf(" %d. %s\n", i+1, line)
	}
	fmt.Printf("\n共 %d 条日志\n", len(validLines))
}

// showHelp 显示帮助内容
func showHelp() {
	fmt.Println("==========================================================")
	fmt.Printf(" 使用帮助\n")
	fmt.Println("==========================================================")
	fmt.Println("占位文本管理")
	fmt.Println(" 1. 新增占位文本: 占位文本格式为 key=value")
	fmt.Println(" 2. 删除占位文本: 输入 key 删除对应的占位文本")
	fmt.Println(" 7. 查看占位文本: 显示所有已保存的占位文本")
	fmt.Println(" 10. 搜索占位文本: 通过关键字搜索占位文本")
	fmt.Println("\n黑名单管理")
	fmt.Println(" 3. 新增文件黑名单: 添加禁止访问的文件路径")
	fmt.Println(" 4. 删除文件黑名单: 移除文件黑名单项")
	fmt.Println(" 5. 新增命令黑名单: 添加禁止执行的命令")
	fmt.Println(" 6. 删除命令黑名单: 移除命令黑名单项")
	fmt.Println(" 8. 查看文件黑名单: 显示所有文件黑名单")
	fmt.Println(" 9. 查看命令黑名单: 显示所有命令黑名单")
	fmt.Println(" 11. 搜索黑名单: 通过关键字搜索黑名单项")
	fmt.Println("\n系统功能")
	fmt.Println(" 12. 查看操作日志: 查看最近 20 条操作记录")
	fmt.Println(" 13. 帮助: 显示此帮助信息")
	fmt.Println(" 0. 退出: 退出程序")

	fmt.Println("\n提示:")
	fmt.Println(" ** 占位文本格式: key=value，例如 password=123456")
	fmt.Println(" ** 文件路径支持相对路径和绝对路径")
	fmt.Println(" ** 所有操作都会自动记录到日志文件")
}
