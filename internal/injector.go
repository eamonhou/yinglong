package internal

import (
	"strings"
)

type Injecter interface {
	Inject(raw string) (string, error)
	SetRelationMap(relationMap map[string]string) Injecter
}

// 简单注入器
type SimpleInjector struct {
	strBuilder  strings.Builder // 拼接大量字符串时使用缓冲区拼接节省内存
	relationMap map[string]string
}

func NewInjector(kind string) Injecter {
	var injecter Injecter

	strBuilder := strings.Builder{}

	switch kind {
	case "simple":
		injecter = &SimpleInjector{
			strBuilder: strBuilder,
		}
	default:
		injecter = &SimpleInjector{
			strBuilder: strBuilder,
		}
	}
	return injecter
}

func (obj *SimpleInjector) SetRelationMap(relationMap map[string]string) Injecter {
	obj.relationMap = relationMap
	return obj
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
		result = strings.ReplaceAll(result, obj.strBuilder.String(), v)
	}
	return result, nil
}
