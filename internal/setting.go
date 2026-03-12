package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"yinglong/cst"
)

type Settingger interface {
	ReadPasswordSetting(filepath string) (map[string]string, error)
	ReadDenySetting(filepath string) (map[string]struct{}, map[string]struct{}, error)
}

func NewSettingger(kind string) Settingger {
	var setting Settingger

	switch kind {
	case "simple":
		setting = NewSimpleSetting()
	default:
		setting = NewSimpleSetting()
	}
	return setting
}

type SimpleSetting struct {
	//
}

func NewSimpleSetting() *SimpleSetting {
	return &SimpleSetting{}
}

// ReadPasswordSetting 读取密码设置
//
//	@param ctx
//	@param filepath 配置文件路径
//	@return map[string]string
//	@return error
func (obj *SimpleSetting) ReadPasswordSetting(filepath string) (map[string]string, error) {
	passFp, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer passFp.Close()

	result := make(map[string]string)
	err = json.NewDecoder(passFp).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ReadDenySetting 读取黑名单设置
//
//	@param ctx
//	@param filepath 黑名单文件路径
//	@return map[string]struct{}
//	@return map[string]struct{}
//	@return error
func (obj *SimpleSetting) ReadDenySetting(filepath string) (map[string]struct{}, map[string]struct{}, error) {
	var denies map[string]interface{}

	// 读取配置文件
	denyFp, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, nil, fmt.Errorf("打开黑名单配置文件失败：%w\n", err)
	}
	defer denyFp.Close()

	err = json.NewDecoder(denyFp).Decode(&denies)
	if err != nil {
		return nil, nil, fmt.Errorf("黑名单配置格式错误：%w\n", err)
	}

	// 加载命令黑名单
	denyCommandList := make(map[string]struct{})
	if denyCommandsItf, ok := denies["commands"]; ok {
		denyCommandsItfs, ok := denyCommandsItf.([]interface{})
		if !ok {
			return nil, nil, cst.DenyConfigError
		}
		for _, denyCommandItf := range denyCommandsItfs {
			denyCommand, ok := denyCommandItf.(string)
			if !ok {
				return nil, nil, cst.DenyConfigError
			}
			denyCommandList[denyCommand] = struct{}{}
		}
	}

	// 加载文件黑名单
	denyFileList := make(map[string]struct{})
	if denyFilesItf, ok := denies["files"]; ok {
		denyFilesItfs, ok := denyFilesItf.([]interface{})
		if !ok {
			return nil, nil, cst.DenyConfigError
		}
		for _, denyFileItf := range denyFilesItfs {
			denyFile, ok := denyFileItf.(string)
			if !ok {
				return nil, nil, cst.DenyConfigError
			}
			denyFileList[denyFile] = struct{}{}
		}
	}
	return denyCommandList, denyFileList, nil
}
