package internal

import (
	"strings"
)

type Injecter interface {
	Inject(raw string) (string, error)
}

// 简单注入器
type SimpleInjector struct {
	strBuilder  strings.Builder // 拼接大量字符串时使用缓冲区拼接节省内存
	relationMap map[string]string
}

type InjectorConfig struct {
	RelationMap map[string]string
}

func NewInjector(cfg InjectorConfig) Injecter {
	return &SimpleInjector{
		relationMap: cfg.RelationMap,
	}
}

// Inject 注入
//
//	@param raw
//	@return string
//	@return error
func (obj *SimpleInjector) Inject(raw string) (string, error) {
	result := raw
	for k, v := range obj.relationMap {
		obj.strBuilder.WriteString("{{")
		obj.strBuilder.WriteString(k)
		obj.strBuilder.WriteString("}}")
		placeholder := obj.strBuilder.String()
		obj.strBuilder.Reset()
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result, nil
}
