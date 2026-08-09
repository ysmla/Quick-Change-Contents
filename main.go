package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	"golang.design/x/clipboard"
)

func main() {
	//初始化数据对象
	InitDataFile()
	type_list := []string{"ls", "cd", "mi", "cg", "sh", "de", "fi"}
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
	} else if input_type == "cd" {
		key := command[2]
		if KeyExists(quick_dict, key) {
			var n int
			fmt.Sscanf(key, "%d", &n)
			command := "cd " + filepath.Join(path, quick_dict[n])
			clipboard.Write(clipboard.FmtText, []byte(command))
			fmt.Printf("已将 %s 复制到剪贴板\n", command)
		} else {
			fmt.Println("输入的键不存在")
		}
	} else if input_type == "mi" {
		name := command[2]
		if name == "" {
			fmt.Println("请输入快捷方式的名字")
			return
		}
		err := WriteData(name, path)
		if err != nil {
			fmt.Println("写入数据失败:", err)
			return
		}
	} else if input_type == "cg" {
		value := ReadData(command[2])
		if value != nil {
			new_command := "cd " + value.(string)
			clipboard.Write(clipboard.FmtText, []byte(new_command))
			fmt.Printf("已将 %s 复制到剪贴板\n", new_command)
		}
	} else if input_type == "sh" {
		PrintData()
	} else if input_type == "de" {
		DeleteData(command[2])
	} else if input_type == "fi" {
		FindData(command[2])
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

// KeyExists 判断字典内是否存在以 key 为键名的键。
// key 为用户传入的字符串，先检查是否为数字（int），再检查字典中是否存在该键。
func KeyExists[V any](m map[int]V, key string) bool {
	var n int
	_, err := fmt.Sscanf(key, "%d", &n)
	if err != nil {
		return false
	}
	_, ok := m[n]
	return ok
}

// getDataPath 返回 exe 同目录下 data.json 的绝对路径。
func getDataPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "data.json"
	}
	return filepath.Join(filepath.Dir(exe), "data.json")
}

// InitDataFile 初始化 data.json 数据文件。
// 若文件不存在则创建；若文件存在但内容不是合法 JSON 键值对，
// 则将原文件重命名为 broken_data.json，再创建新的 data.json。
// 返回加载到的键值对数据。
func InitDataFile() map[string]interface{} {
	filename := getDataPath()

	// 读取文件
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建空 JSON 对象
			os.WriteFile(filename, []byte("{}"), 0644)
			return make(map[string]interface{})
		}
		fmt.Printf("读取 %s 失败: %v\n", filename, err)
		return nil
	}

	// 尝试解析为 JSON 键值对
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		// 不是合法 JSON 键值对，重命名原文件
		brokenFile := filepath.Join(filepath.Dir(filename), "broken_data.json")
		fmt.Printf("⚠ %s 内容不是合法 JSON 键值对，已重命名为 broken_data.json，并创建新的 data.json\n", filename)
		os.Rename(filename, brokenFile)
		os.WriteFile(filename, []byte("{}"), 0644)
		return make(map[string]interface{})
	}

	return result
}

// WriteData 将 key-value 追加写入 data.json。
func WriteData(key string, value interface{}) error {
	filename := getDataPath()

	// 读取现有数据
	data := make(map[string]interface{})
	if file, err := os.ReadFile(filename); err == nil {
		json.Unmarshal(file, &data)
	}

	// 不允许重名键
	if _, exists := data[key]; exists {
		return fmt.Errorf("键 \"%s\" 已存在，不允许覆盖", key)
	}

	// 追加键值对
	data[key] = value

	// 写回文件
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	return os.WriteFile(filename, out, 0644)
}

// ReadData 在 data.json 中查找 key 对应的值。
// 若找不到则打印提示并返回 nil。
func ReadData(key string) interface{} {
	filename := getDataPath()

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("读取 data.json 失败: %v\n", err)
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Printf("解析 data.json 失败: %v\n", err)
		return nil
	}

	value, ok := data[key]
	if !ok {
		fmt.Printf("⚠ data.json 中未找到键 \"%s\"\n", key)
		return nil
	}

	return value
}

// PrintData 读取 data.json 并用 printColoredMap 打印所有键值对。
func PrintData() {
	filename := getDataPath()

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("读取 data.json 失败: %v\n", err)
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(file, &raw); err != nil {
		fmt.Printf("解析 data.json 失败: %v\n", err)
		return
	}

	// 转为 map[int]string 以适配 printColoredMap
	data := make(map[int]string, len(raw))
	i := 0
	for k, v := range raw {
		data[i] = fmt.Sprintf("%s: %v", k, v)
		i++
	}

	printColoredMap(data)
}

// DeleteData 在 data.json 中删除 key 对应的数据。
// 若找不到则打印提示。
func DeleteData(key string) {
	filename := getDataPath()

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("读取 data.json 失败: %v\n", err)
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Printf("解析 data.json 失败: %v\n", err)
		return
	}

	if _, ok := data[key]; !ok {
		fmt.Printf("⚠ data.json 中未找到键 \"%s\"，删除失败\n", key)
		return
	}

	delete(data, key)

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}
	if err := os.WriteFile(filename, out, 0644); err != nil {
		fmt.Printf("写入 data.json 失败: %v\n", err)
		return
	}
	fmt.Printf("已删除键 \"%s\"\n", key)
}

// FindData 在 data.json 中查找 key 对应的值并打印。
func FindData(key string) {
	value := ReadData(key)
	if value != nil {
		fmt.Printf("%s: %v\n", key, value)
	}
}
