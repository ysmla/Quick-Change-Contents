# Quick-Change-Dict

跨平台命令行快速目录跳转工具。将常用目录保存为快捷方式，通过剪贴板一键跳转。

[English](README.md)

## 安装

```bash
# 克隆并编译
git clone https://github.com/ysmla/Quick-Change-Dict.git
cd Quick-Change-Dict
go build -o q .

# 添加到 PATH（可选，方便全局使用）
# 将项目目录添加到系统环境变量 PATH 中
```

## 命令

| 命令 | 用法 | 说明 |
|------|------|------|
| `ls` | `q ls` | 列出当前目录文件（彩色输出） |
| `cd` | `q cd <序号>` | 按序号将 `cd <路径>` 命令复制到剪贴板 |
| `mi` | `q mi <名称>` | 将当前目录保存为快捷方式（不允许重名） |
| `cg` | `q cg <名称>` | 将已保存快捷方式的 `cd` 命令复制到剪贴板 |
| `sh` | `q sh` | 显示所有已保存的快捷方式 |
| `de` | `q de <名称>` | 删除一个快捷方式 |
| `fi` | `q fi <名称>` | 查找并显示一个快捷方式 |

## 示例

```bash
# 列出当前目录
> q ls
  0 : Documents
  1 : Projects
  2 : Downloads

# 保存快捷方式
> q mi projects        # 将当前目录保存为 "projects"

# 使用快捷方式（复制 cd 命令到剪贴板）
> q cg projects
已将 cd /home/user/code/Projects 复制到剪贴板

# 查看所有快捷方式
> q sh
  0 : projects: /home/user/code/Projects
  1 : downloads: /home/user/Downloads

# 删除快捷方式
> q de downloads
已删除键 "downloads"

# 查找快捷方式
> q fi projects
projects: /home/user/code/Projects
```

## 数据存储

所有快捷方式存储在可执行文件同目录下的 `data.json` 文件中，首次运行自动创建。
