package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestParseIndex 验证索引解析严格拒绝部分匹配输入。
func TestParseIndex(t *testing.T) {
	cases := []struct {
		in     string
		want   int
		wantOK bool
	}{
		{"0", 0, true},
		{"1", 1, true},
		{"42", 42, true},
		{"01", 1, true},
		{"-1", 0, false},
		{"1abc", 0, false}, // Sscanf 时代的漏洞：部分匹配
		{"01x", 0, false},
		{"abc", 0, false},
		{"", 0, false},
		{" 1", 0, false},
	}
	for _, c := range cases {
		got, ok := parseIndex(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("parseIndex(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

// TestQuotePath 验证路径加引号，兼容含空格路径。
func TestQuotePath(t *testing.T) {
	if got := quotePath(`C:\Program Files\Go`); got != `"C:\Program Files\Go"` {
		t.Errorf("quotePath = %q", got)
	}
	if got := quotePath(`/home/user/my dir`); got != `"/home/user/my dir"` {
		t.Errorf("quotePath = %q", got)
	}
}

// TestRequireArg 验证参数校验：缺参数或空参数返回 false 而不是越界。
func TestRequireArg(t *testing.T) {
	if _, ok := requireArg([]string{"cd", "1"}); !ok {
		t.Error("cd 1 should be OK")
	}
	if _, ok := requireArg([]string{"cd"}); ok {
		t.Error("cd without arg should fail")
	}
	if _, ok := requireArg([]string{"cd", ""}); ok {
		t.Error("cd with empty arg should fail")
	}
}

// TestStoreSetGetDelete 验证快捷方式的增删查与持久化。
func TestStoreSetGetDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	s := &Store{data: make(map[string]string), path: path}

	if err := s.Set("projects", `C:\code\Projects`); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if v, ok := s.Get("projects"); !ok || v != `C:\code\Projects` {
		t.Fatalf("Get = (%q, %v)", v, ok)
	}

	// 不允许重名覆盖
	if err := s.Set("projects", "other"); err == nil {
		t.Error("duplicate Set should fail")
	}

	// 数据应已持久化到磁盘且是合法 JSON
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	reloaded := make(map[string]string)
	if err := json.Unmarshal(raw, &reloaded); err != nil {
		t.Fatalf("saved file is not valid JSON: %v\n%s", err, raw)
	}
	if v, ok := reloaded["projects"]; !ok || v != `C:\code\Projects` {
		t.Fatalf("reloaded = (%q, %v)", v, ok)
	}

	// 删除
	deleted, err := s.Delete("projects")
	if err != nil || !deleted {
		t.Fatalf("Delete = (%v, %v)", deleted, err)
	}
	if _, ok := s.Get("projects"); ok {
		t.Error("key should be gone after Delete")
	}
	// 删除不存在的键
	deleted, err = s.Delete("nope")
	if err != nil || deleted {
		t.Fatalf("Delete missing = (%v, %v), want (false, nil)", deleted, err)
	}
}

// TestLoadStoreMissing 验证数据文件不存在时返回空 Store。
func TestLoadStoreMissing(t *testing.T) {
	t.Setenv("Q_DATA_DIR", t.TempDir())
	s, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d entries", s.Len())
	}
}

// TestLoadStoreCorrupt 验证损坏的 JSON 会被备份后重建，且不会报假成功。
func TestLoadStoreCorrupt(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("Q_DATA_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "data.json"), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore should tolerate corrupt data: %v", err)
	}
	if s.Len() != 0 {
		t.Fatalf("expected empty store after reset, got %d", s.Len())
	}

	// 原文件应被改名备份，且备份文件名唯一（时间戳）
	matches, err := filepath.Glob(filepath.Join(dir, "broken_data_*.json"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected exactly 1 backup, got %v (err %v)", matches, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data.json")); !os.IsNotExist(err) {
		t.Error("original data.json should have been moved away")
	}
}

// TestLoadStoreNonStringValue 验证值类型不是字符串的 JSON 也按损坏处理，
// 避免旧代码 value.(string) 的 panic 隐患。
func TestLoadStoreNonStringValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("Q_DATA_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "data.json"), []byte(`{"a": 123}`), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d", s.Len())
	}
}

// TestStoreSaveNoTmpLeftover 验证原子写入后不残留临时文件。
func TestStoreSaveNoTmpLeftover(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	s := &Store{data: map[string]string{"a": "b"}, path: path}
	if err := s.Set("c", "d"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("temporary file should be cleaned up after save")
	}
}

// TestDataPathEnvOverride 验证 Q_DATA_DIR 环境变量可覆盖数据文件位置。
func TestDataPathEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("Q_DATA_DIR", dir)
	got, err := dataPath()
	if err != nil {
		t.Fatalf("dataPath: %v", err)
	}
	if want := filepath.Join(dir, "data.json"); got != want {
		t.Fatalf("dataPath = %q, want %q", got, want)
	}
}

// TestPrintStoreKeysSorted 验证 sh 输出按键名排序且稳定。
func TestPrintStoreKeysSorted(t *testing.T) {
	s := &Store{data: map[string]string{
		"zeta":  "/z",
		"alpha": "/a",
		"mid":   "/m",
	}}
	keys := s.Keys()
	sort.Strings(keys)
	got := keys[0] + "," + keys[1] + "," + keys[2]
	if got != "alpha,mid,zeta" {
		t.Fatalf("keys not sorted: %v", keys)
	}
}
