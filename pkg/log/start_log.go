package log

import (
	"fmt"
	"reflect"
	"strings"

	"mcp/config"
)

// PrintConfigFlags 打印配置中所有包含 "ENABLE" 和 "IS" 的变量值
func PrintConfigFlags(cfg *config.MCPConfig) {
	fmt.Println("=== 配置标志位信息 ===")

	// 使用反射获取配置结构体的所有字段
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 如果字段是结构体，递归检查其字段
		if field.Kind() == reflect.Struct {
			printStructFlags(field, fieldType.Name+".")
		}
	}

	fmt.Println("===================")
}

// printStructFlags 递归打印结构体中包含 "ENABLE" 和 "IS" 的字段
func printStructFlags(s reflect.Value, prefix string) {
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		fieldType := t.Field(i)

		// 获取字段的mapstructure标签
		tag := fieldType.Tag.Get("mapstructure")
		if tag == "" {
			continue
		}

		// 检查字段名是否包含 "ENABLE" 或 "IS"
		if strings.Contains(strings.ToUpper(tag), "ENABLE") || strings.Contains(strings.ToUpper(tag), "IS") {
			// 打印字段名和值
			fmt.Printf("%s%s: %v\n", prefix, tag, field.Interface())
		}

		// 如果字段是结构体，递归检查
		if field.Kind() == reflect.Struct {
			printStructFlags(field, prefix+tag+".")
		}
	}
}
