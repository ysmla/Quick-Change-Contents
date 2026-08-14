package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/fatih/color"
	"golang.design/x/clipboard"
)

// 支持的子命令列表。
var commands = []string{"ls", "cd", "mi", "cg", "sh", "de", "fi", "help"}

func main() {
	os.Exit(run(os.Args[1:]))
}

// run 执行命令并返回退出码。拆出来便于单元测试。
func run(args []string) int {
	// 初始化剪贴板。无图形环境（如无 X 服务器的 Linux）下会失败，
	// 此时仅 cd/cg 无法复制命令，其余功能不受影响。
	if err := clipboard.Init(); err != nil {
		fmt.Printf("⚠ 剪贴板不可用: %v（cd/cg 命令将无法复制命令）\n", err)
	}

	store, err := loadStore()
	if err != nil {
		fmt.Printf("初始化数据文件失败: %v\n", err)
		return 1
	}

	if len(args) == 0 {
		usage()
		return 0
	}
	cmd := args[0]
	if !slices.Contains(commands, cmd) {
		fmt.Printf("参数错误: 未知命令 %q\n", cmd)
		usage()
		return 1
	}

	switch cmd {
	case "help":
		usage()
		return 0
	case "ls":
		return printDir()
	case "cd":
		key, ok := requireArg(args)
		if !ok {
			return 1
		}
		return changeDir(key)
	case "mi":
		name, ok := requireArg(args)
		if !ok {
			return 1
		}
		return makeShortcut(store, name)
	case "cg":
		name, ok := requireArg(args)
		if !ok {
			return 1
		}
		return copyShortcut(store, name)
	case "sh":
		printStore(store)
		return 0
	case "de":
		name, ok := requireArg(args)
		if !ok {
			return 1
		}
		return deleteShortcut(store, name)
	case "fi":
		name, ok := requireArg(args)
		if !ok {
			return 1
		}
		return findShortcut(store, name)
	}
	return 0
}

// usage 打印使用说明。
func usage() {
	fmt.Println(`q - 快速目录跳转工具

用法: q <命令> [参数]

命令:
  ls        列出当前目录的内容（目录带 / 后缀）
  cd <序号> 将 "cd <路径>" 复制到剪贴板
  mi <名称> 将当前目录保存为快捷方式（不允许重名）
  cg <名称> 将快捷方式的 "cd" 命令复制到剪贴板
  sh        显示所有已保存的快捷方式（按名称排序）
  de <名称> 删除一个快捷方式
  fi <名称> 查找并显示一个快捷方式
  help      显示本帮助`)
}

// requireArg 检查子命令参数是否齐全。
// 缺参数时打印提示并返回 false，避免越界访问。
func requireArg(args []string) (string, bool) {
	if len(args) < 2 || args[1] == "" {
		fmt.Printf("缺少参数: 用法 q %s <参数>\n", args[0])
		return "", false
	}
	return args[1], true
}

// parseIndex 将字符串严格解析为非负索引。
// 使用 strconv.Atoi 拒绝 "1abc"、"01x" 这类部分匹配输入。
func parseIndex(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// quotePath 给路径加双引号，兼容含空格的路径（cmd/pwsh/bash 均支持）。
func quotePath(p string) string {
	return "\"" + p + "\""
}

// copyToClipboard 将文本写入剪贴板，失败时报告真实错误而不是假装成功。
func copyToClipboard(text string) bool {
	if clipboard.Write(clipboard.FmtText, []byte(text)) == nil {
		fmt.Printf("⚠ 剪贴板写入失败，未复制: %s\n", text)
		return false
	}
	fmt.Printf("已将 %s 复制到剪贴板\n", text)
	return true
}

// printDir 彩色打印当前目录条目，目录带 / 后缀，按文件名排序。
func printDir() int {
	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return 1
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return 1
	}

	keyColor := color.New(color.FgCyan, color.Bold) // 加粗青色
	valueColor := color.New(color.FgGreen)
	maxKeyLen := len(strconv.Itoa(len(entries) - 1))

	for i, e := range entries {
		keyColor.Printf("%*d", maxKeyLen, i)
		fmt.Print(" : ")
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		valueColor.Print(name)
		fmt.Println()
	}
	return 0
}

// changeDir 将 "cd <当前目录下第 n 个条目>" 复制到剪贴板。
func changeDir(key string) int {
	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return 1
	}
	n, ok := parseIndex(key)
	if !ok {
		fmt.Printf("输入的键不存在: %q\n", key)
		return 1
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return 1
	}
	if n >= len(entries) {
		fmt.Printf("输入的键不存在: %q（当前目录共 %d 项）\n", key, len(entries))
		return 1
	}

	target := filepath.Join(path, entries[n].Name())
	if !copyToClipboard("cd " + quotePath(target)) {
		return 1
	}
	return 0
}

// makeShortcut 将当前目录保存为快捷方式。
func makeShortcut(store *Store, name string) int {
	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return 1
	}
	if err := store.Set(name, path); err != nil {
		fmt.Printf("写入数据失败: %v\n", err)
		return 1
	}
	fmt.Printf("已保存快捷方式 %q -> %s\n", name, path)
	return 0
}

// copyShortcut 将快捷方式对应的 cd 命令复制到剪贴板。
func copyShortcut(store *Store, name string) int {
	value, ok := store.Get(name)
	if !ok {
		fmt.Printf("⚠ 未找到键 %q\n", name)
		return 1
	}
	if !copyToClipboard("cd " + quotePath(value)) {
		return 1
	}
	return 0
}

// printStore 按名称排序打印所有快捷方式。
func printStore(store *Store) {
	if store.Len() == 0 {
		fmt.Println("（暂无快捷方式，使用 q mi <名称> 保存当前目录）")
		return
	}
	keys := store.Keys()
	sort.Strings(keys)

	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	keyColor := color.New(color.FgCyan, color.Bold)
	valueColor := color.New(color.FgGreen)

	for _, k := range keys {
		keyColor.Printf("%-*s", maxKeyLen, k)
		fmt.Print(" : ")
		v, _ := store.Get(k)
		valueColor.Print(v)
		fmt.Println()
	}
}

// deleteShortcut 删除一个快捷方式。
func deleteShortcut(store *Store, name string) int {
	deleted, err := store.Delete(name)
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return 1
	}
	if !deleted {
		fmt.Printf("⚠ 未找到键 %q，删除失败\n", name)
		return 1
	}
	fmt.Printf("已删除键 %q\n", name)
	return 0
}

// findShortcut 查找并显示一个快捷方式。
func findShortcut(store *Store, name string) int {
	value, ok := store.Get(name)
	if !ok {
		fmt.Printf("⚠ 未找到键 %q\n", name)
		return 1
	}
	fmt.Printf("%s: %s\n", name, value)
	return 0
}

// Store 保存快捷方式键值对（键值均为字符串）。
type Store struct {
	data map[string]string
	path string
}

// loadStore 从数据文件加载快捷方式。
// 文件不存在时返回空 Store；文件损坏时先备份再重建。
func loadStore() (*Store, error) {
	path, err := dataPath()
	if err != nil {
		return nil, err
	}
	s := &Store{data: make(map[string]string), path: path}

	// 确保数据目录存在（首次运行时用户配置目录可能尚未创建）。
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil // 首次运行，返回空数据
		}
		return nil, fmt.Errorf("读取 %s 失败: %w", path, err)
	}

	if err := json.Unmarshal(raw, &s.data); err != nil {
		// 内容不是合法 JSON（或值不是字符串），备份原文件后重建。
		// 注意：Unmarshal 失败时可能已把部分键残留进 map，必须重置。
		s.data = make(map[string]string)
		backup := filepath.Join(filepath.Dir(path),
			fmt.Sprintf("broken_data_%s.json", time.Now().Format("20060102-150405")))
		if err := os.Rename(path, backup); err != nil {
			return nil, fmt.Errorf("备份损坏的 %s 到 %s 失败: %w", path, backup, err)
		}
		fmt.Printf("⚠ %s 内容不是合法 JSON 键值对，已备份为 %s，数据文件已重置\n", path, backup)
	}
	return s, nil
}

// Get 返回键对应的值。
func (s *Store) Get(key string) (string, bool) {
	v, ok := s.data[key]
	return v, ok
}

// Len 返回快捷方式数量。
func (s *Store) Len() int { return len(s.data) }

// Keys 返回所有键。
func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Set 保存键值对，不允许覆盖已有键。
func (s *Store) Set(key, value string) error {
	if _, ok := s.data[key]; ok {
		return fmt.Errorf("键 %q 已存在，不允许覆盖", key)
	}
	s.data[key] = value
	if err := s.save(); err != nil {
		delete(s.data, key) // 写盘失败时回滚内存中的修改
		return err
	}
	return nil
}

// Delete 删除键，返回是否删除成功。
func (s *Store) Delete(key string) (bool, error) {
	if _, ok := s.data[key]; !ok {
		return false, nil
	}
	delete(s.data, key)
	if err := s.save(); err != nil {
		s.data[key] = key // 写盘失败时回滚
		return false, err
	}
	return true, nil
}

// save 原子写回数据文件：先写临时文件再重命名，避免中途崩溃损坏数据。
func (s *Store) save() error {
	out, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", tmp, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("写入 %s 失败: %w", s.path, err)
	}
	return nil
}

// dataPath 返回数据文件路径。
// 优先使用用户配置目录（即使安装到只读位置也能写入）；
// 可用环境变量 Q_DATA_DIR 覆盖；并自动迁移旧版位于 exe 同目录的 data.json。
func dataPath() (string, error) {
	if dir := os.Getenv("Q_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "data.json"), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取用户配置目录失败: %w", err)
	}
	dir = filepath.Join(dir, "q")
	path := filepath.Join(dir, "data.json")

	// 迁移旧数据：用户配置目录尚无数据文件，而 exe 同目录存在旧 data.json 时复制过去。
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if old, err := os.ReadFile(legacyDataPath()); err == nil {
			if err := os.MkdirAll(dir, 0755); err == nil {
				if err := os.WriteFile(path, old, 0644); err == nil {
					fmt.Printf("已迁移旧数据文件 %s -> %s\n", legacyDataPath(), path)
				}
			}
		}
	}
	return path, nil
}

// legacyDataPath 返回旧版本的数据文件路径（exe 同目录）。
func legacyDataPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "data.json"
	}
	return filepath.Join(filepath.Dir(exe), "data.json")
}
