# cleaner

定时清理指定文件与目录内容的常驻进程工具。

## 功能

- 按配置间隔定时执行清理任务
- 删除 `files` 列表中的指定文件
- 清空 `dirs` 目录下的子目录和文件（保留 `ignore` 中的项）
- 日志同时输出到控制台和进程工作目录下的 `cleaner.log`

## 目录结构

```
cleaner/
├── main.go                 # 程序入口、定时调度、启动流程
├── config.json             # 业务配置文件
├── internal/
│   ├── config/
│   │   └── config.go      # 配置结构体、加载、校验、默认值
│   ├── cleaner/
│   │   └── cleaner.go     # 核心清理逻辑：删文件、清目录、忽略规则
│   └── util/
│       ├── fs.go          # 文件/目录工具、路径校验
│       └── logger.go      # 日志封装
├── go.mod
└── go.sum
```

## 配置说明

配置文件为 JSON 格式，示例：

```json
{
	"interval": 10,
	"dirs": [
		{
			"path": "D:\\apps\\JianyingPro\\",
			"ignore": [
				"filename"
			]
		}
	],
	"files": [
		"D:\\apps\\JianyingPro\\filename"
	]
}
```

### 参数

| 字段 | 说明 |
|------|------|
| `interval` | 定时任务间隔，单位秒；未配置或 ≤ 0 时默认 **60** 秒 |
| `dirs` | 待清理的目录列表 |
| `dirs[].path` | 目录路径，不能为空；Windows 路径长度需 > 3，Linux 需 > 1 |
| `dirs[].ignore` | 忽略列表，路径相对于 `path`；如 `"filename.exe"`、`"dir/test.exe"` |
| `files` | 待删除的文件列表，**必须为绝对路径** |

### 清理规则

**目录 (`dirs`)**

- 清理配置目录下的子目录和文件，**不删除**配置目录本身
- `ignore` 中的文件或目录及其子树不会被删除
- 路径不存在、或路径实际是文件：跳过并输出 ERROR 日志

**文件 (`files`)**

- 必须是绝对路径，相对路径在加载配置时报错
- 路径不存在：跳过并输出 ERROR 日志
- 路径是目录：跳过并输出 ERROR 日志
- 删除成功时输出文件的**绝对路径**

## 日志格式

每条日志包含时间、源码文件名、行号和消息：

```
2026-06-15 18:56:30 cleaner.go:58 deleted file: D:\apps\JianyingPro\cache.tmp
2026-06-15 18:56:30 main.go:44 cleaner started, interval=10s, config=D:\cleaner\config.json
```

日志同时写入：

- 标准输出（控制台）
- 进程工作目录下的 `cleaner.log`

## 构建与运行

```powershell
# 构建
go build -o cleaner.exe .

# 使用默认 config.json（进程工作目录下）
.\cleaner.exe

# 指定配置文件
.\cleaner.exe -config D:\path\to\config.json
```

程序启动后立即执行一次清理，之后按 `interval` 定时重复执行。收到 `SIGINT` 或 `SIGTERM` 时优雅退出。

## 测试

```powershell
go test ./... -v
```
