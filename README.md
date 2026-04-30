# icoo_assistant

> CLI 编码助手智能体 — 基于 Anthropic Claude API 的本地编程助手

[![Go Version](https://img.shields.io/badge/Go-1.23.0-00ADD8?logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-0.1.36-blue)](#)

---

## 简介

`icoo_assistant` 是一个用 Go 编写的 **CLI 编码助手**，在工作区目录中运行，通过自然语言对话执行编程任务。它使用工具驱动的智能体循环，能够读写文件、执行 Shell 命令、管理后台任务、跟踪待办事项，并支持子智能体任务委派。

### 核心能力

- **代码操作**: 读取、写入、精确编辑文件
- **命令执行**: 执行 Shell 命令（含超时与危险命令拦截）
- **后台任务**: 启动异步后台作业并轮询完成通知
- **任务管理**: 持久化项目任务（支持依赖链、状态流转、后台关联）
- **审计分析**: 任务执行历史审计、智能体运行时事件审计
- **对话压缩**: 自动/手动压缩长对话，避免上下文溢出
- **子智能体**: 将复杂任务委派给子智能体并行执行
- **技能系统**: 加载 SKILL.md 领域知识文件

---

## 快速开始

### 环境要求

- Go 1.23.0+
- Anthropic API Key（可选，用于真实模式）

### 安装

```bash
git clone <repository-url>
cd icoo_assistant
```

### 配置

```bash
# 1. 创建配置文件
cp .env.example .env

# 2. 编辑 .env，填入你的 API Key
# ANTHROPIC_API_KEY=sk-ant-xxx
```

### 运行

```bash
# 构建二进制
go build -o assistant ./cmd/assistant

# 自检
./assistant check

# 单次查询
./assistant "列出所有可用工具"

# 交互模式
./assistant
```

---

## 配置项

配置文件为工作目录下的 `.env` 文件。完整配置项：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `AGENT_SYSTEM_PROMPT` | 自动生成 | 系统提示词 |
| `AGENT_SKILLS_DIR` | `skills` | 技能定义目录 |
| `AGENT_MAX_ROUNDS` | `20` | 最大迭代轮数 |
| `AGENT_COMMAND_TIMEOUT_SECONDS` | `120` | Shell 命令超时（秒） |
| `AGENT_COMPACT_THRESHOLD` | `50000` | 自动压缩 Token 阈值 |
| `AGENT_TRANSCRIPT_DIR` | `.transcripts` | 对话转录存储目录 |
| `ANTHROPIC_API_KEY` | (空) | API Key，为空则运行在 Fake 模式 |
| `ANTHROPIC_BASE_URL` | (空) | 自定义 API 基础 URL |
| `ANTHROPIC_MODEL` | `claude-opus-4-7` | 模型名称 |
| `ANTHROPIC_MAX_TOKENS` | `16000` | 最大输出 Token |
| `ANTHROPIC_ENABLE_PROMPT_CACHE` | `false` | 启用提示缓存 |
| `ANTHROPIC_ENABLE_THINKING` | `true` | 启用扩展思考 |
| `ANTHROPIC_ENABLE_STREAMING` | `true` | 启用流式响应 |

---

## 运行模式

| 模式 | 条件 | 行为 |
|------|------|------|
| **Real** | 设置了 `ANTHROPIC_API_KEY` | 调用真实 Anthropic API |
| **Fake** | 未设置 `ANTHROPIC_API_KEY` | 返回空响应，用于本地干运行和配置验证 |

Fake 模式下运行 `check` 命令可查看最小可行路径指引。

---

## 命令说明

```
icoo_assistant 0.1.36

Usage:
  assistant [query]           单次任务模式
  assistant check             自检诊断
  assistant doctor            (同 check)
  assistant --version         版本信息
  assistant --help            帮助信息

Examples:
  assistant                                   启动交互 REPL
  assistant check                             诊断当前环境
  assistant "read README and summarize"       单次查询
```

---

## 可用工具

| 工具 | 说明 |
|------|------|
| `bash` | 执行 Shell 命令 |
| `read_file` | 读取文件内容 |
| `write_file` | 写入/覆盖文件 |
| `edit_file` | 精确字符串替换编辑 |
| `background` | 启动/查询后台异步任务 |
| `project_task` | 项目任务 CRUD |
| `task_audit` | 项目任务执行历史审计 |
| `agent_hook_audit` | 智能体运行时事件审计 |
| `todo` | 会话内待办事项管理 |
| `compact` | 手动触发对话压缩 |
| `task` | 委托子智能体执行任务 |
| `tool_catalog` | 列出/描述所有可用工具 |
| `load_skill` | 加载领域知识技能 |

---

## 项目结构

```
icoo_assistant/
├── .env.example              # 配置模板
├── go.mod / go.sum           # Go 模块定义
├── cmd/
│   └── assistant/            # CLI 入口
│       ├── main.go           # 参数解析与路由
│       ├── app.go            # 组件装配与运行
│       ├── check.go          # 自检诊断
│       ├── version.go        # 版本号
│       └── *_test.go         # 测试
└── internal/
    ├── agent/                # 核心智能体循环
    ├── background/           # 后台任务管理
    ├── commandutil/          # Shell 命令执行与验证
    ├── compact/              # 对话压缩
    ├── config/               # 配置加载
    ├── hookaudit/            # 事件审计读取
    ├── llm/                  # LLM 客户端（Anthropic + Fake）
    ├── skill/                # SKILL.md 加载器
    ├── subagent/             # 子智能体运行器
    ├── task/                 # 持久化任务管理
    ├── todo/                 # 待办事项管理
    ├── tools/                # 工具注册与实现
    └── workspace/            # 工作区文件操作
```

---

## 数据存储

程序在工作区创建以下目录：

| 目录 | 内容 | 格式 |
|------|------|------|
| `.tasks/` | 项目任务 | `task_{id}.json` |
| `.background/` | 后台作业 | `job_{id}.json` |
| `.transcripts/` | 对话转录 | `conversation_{runID}.json` |
| `.agent-hooks/` | 运行时事件 | `events.jsonl` |

---

## 开发

```bash
# 运行测试
go test ./...

# 构建
go build -o assistant ./cmd/assistant
```

### 添加技能

在 `skills/<name>/` 目录下创建 `SKILL.md` 文件，支持 YAML 前置元数据：

```markdown
---
name: my-skill
description: A custom domain skill.
---

# Skill Content

...
```

---

## License

Internal project.
