package internal

type Auditer interface {
	AuditCommand(target string) (bool, error)
	AuditFile(target string) (bool, error)
	SetDenyCommandList(denyCommandList map[string]struct{}) Auditer
	SetDenyFileList(denyFileList map[string]struct{}) Auditer
}

type SimpleAuditor struct {
	DenyCommandList map[string]struct{} //禁止执行的命令列表
	DenyFileList    map[string]struct{} //禁止访问的文件列表
}

func NewAuditor(kind string) Auditer {
	var auditer Auditer
	switch kind {
	case "simple":
		auditer = &SimpleAuditor{}
	default:
		auditer = &SimpleAuditor{}
	}
	return auditer
}

func (obj *SimpleAuditor) SetDenyCommandList(denyCommandList map[string]struct{}) Auditer {
	obj.DenyCommandList = denyCommandList
	return obj
}

func (obj *SimpleAuditor) SetDenyFileList(denyFileList map[string]struct{}) Auditer {
	obj.DenyFileList = denyFileList
	return obj
}

// AuditCommand 检查命令是否被禁止
//
//	@param target 要执行的命令
//	@return bool true 可执行，fasle 不可执行
//	@return error
func (obj *SimpleAuditor) AuditCommand(command string) (bool, error) {
	if _, ok := obj.DenyCommandList[command]; ok {
		return false, nil
	}
	return true, nil
}

// AuditFile 检查文件是否被禁止
//
//	@param filename 文件名
//	@return bool true 可执行，fasle 不可执行
//	@return error
func (obj *SimpleAuditor) AuditFile(filename string) (bool, error) {
	if _, ok := obj.DenyFileList[filename]; ok {
		return false, nil
	}
	return true, nil
}
