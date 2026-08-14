# Quick-Change-Contents

跨平台命令行快速目录跳转工具。将常用目录保存为快捷方式，通过剪贴板一键跳转。

[English](README.md)

## 安装

```bash
# 克隆并编译
git clone https://github.com/ysmla/Quick-Change-Contents.git
cd Quick-Change-Contents
go build -o q .

# 添加到 PATH（可选，方便全局使用）
# 将项目目录添加到系统环境变量 PATH 中
```

## 命令

| 命令 | 用法 | 说明 |
|------|------|------|
| `ls` | `q ls` | 列出当前目录内容（目录带 `/` 后缀标记） |
| `cd` | `q cd <序号>` | 按序号将 `cd <路径>` 命令复制到剪贴板（含空格路径自动加引号） |
| `mi` | `q mi <名称>` | 将当前目录保存为快捷方式（不允许重名） |
| `cg` | `q cg <名称>` | 将已保存快捷方式的 `cd` 命令复制到剪贴板 |
| `sh` | `q sh` | 显示所有已保存的快捷方式（按名称排序） |
| `de` | `q de <名称>` | 删除一个快捷方式 |
| `fi` | `q fi <名称>` | 查找并显示一个快捷方式 |
| `help` | `q help` | 显示使用帮助 |

## 示例

```bash
# 列出当前目录（目录带 / 后缀）
> q ls
  0 : Documents/
  1 : Projects/
  2 : Downloads/

# 保存快捷方式
> q mi projects        # 将当前目录保存为 "projects"

# 使用快捷方式（复制 cd 命令到剪贴板）
> q cg projects
已将 cd "/home/user/code/Projects" 复制到剪贴板

# 查看所有快捷方式（按名称排序）
> q sh
downloads   : /home/user/Downloads
projects    : /home/user/code/Projects

# 删除快捷方式
> q de downloads
已删除键 "downloads"

# 查找快捷方式
> q fi projects
projects: /home/user/code/Projects
```

## 数据存储

快捷方式保存在**用户配置目录**下的 `data.json` 中：

- Windows：`%AppData%\q\data.json`
- Linux/macOS：`~/.config/q/data.json`（或 `$XDG_CONFIG_HOME/q/data.json`）

这样即使把程序安装到只读目录也能正常写入。首次运行自动创建。如需覆盖位置，可设置环境变量 `Q_DATA_DIR` 指向任意可写目录。

> **迁移：** 如果你使用过旧版本（数据文件在 exe 同目录），首次运行时旧数据会自动复制到新位置。

如果 `data.json` 损坏（不是合法 JSON），会自动备份为 `broken_data_<时间戳>.json` 并重建新文件，不会静默丢失数据。
