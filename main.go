package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
)

func main() {
	type_list := []string{"ls", "cd"}
	command := os.Args
	if len(command) == 1 {
		fmt.Println("请输入参数")
		return
	}
	input_type := command[1]
	if InMap(ConvertStrSlice2Map(type_list), input_type) == false {
		fmt.Println("参数错误")
		return
	}

	//获取当前工作目录
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	dict, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	quick_dict := make(map[int]string)
	for i, _ := range dict {
		quick_dict[i] = dict[i].Name()
	}

	if input_type == "ls" {
		printColoredMap(quick_dict)
	}
}

// ConvertStrSlice2Map 将字符串 slice 转为 map[string]struct{}。
func ConvertStrSlice2Map(sl []string) map[string]struct{} {
	set := make(map[string]struct{}, len(sl))
	for _, v := range sl {
		set[v] = struct{}{}
	}
	return set
}

// InMap 判断字符串是否在 map 中。
func InMap(m map[string]struct{}, s string) bool {
	_, ok := m[s]
	return ok
}

// 整理打印字典
func printColoredMap(m map[int]string) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// 计算最大键宽度
	maxKeyLen := 0
	for _, k := range keys {
		if len(fmt.Sprintf("%d", k)) > maxKeyLen {
			maxKeyLen = len(fmt.Sprintf("%d", k))
		}
	}

	// 创建颜色对象
	keyColor := color.New(color.FgCyan, color.Bold) // 加粗青色
	valueColor := color.New(color.FgGreen)

	for _, k := range keys {
		keyColor.Printf("%*d", maxKeyLen, k)
		fmt.Print(" : ")
		valueColor.Print(m[k])
		fmt.Println()
	}
}
