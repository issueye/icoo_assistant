# icoo_assistant

`icoo_assistant` 是一个基于 Go 的本地编码 Agent 原型，当前代码主体位于 [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)。

当前 `0.0.1` 版本目标是交付一个可运行的单 Agent MVP，包含这些基础能力：

- LLM 对话循环
- 工具注册与调用分发
- 工作区内文件读取、写入、编辑
- 本地命令执行
- Todo 管理
- 对话压缩与 transcript 落盘
- Skill 加载
- Subagent 摘要委托

## 仓库结构

- [docs](E:\codes\icoo_assistant\docs)：需求、开发计划、版本计划文档
- [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)：Go 模块源码

## 快速开始

```bash
cd icoo_assistant
go test ./...
cp .env.example .env
go run ./cmd/assistant --version
go run ./cmd/assistant --help
go run ./cmd/assistant
```

如果配置了 `ANTHROPIC_API_KEY`，程序会使用真实 Anthropic 客户端；否则会回退到 fake client，方便先验证本地框架是否跑通。

## 配置说明

Go 模块目录 [icoo_assistant/.env.example](E:\codes\icoo_assistant\icoo_assistant\.env.example) 提供了首版可用的环境变量模板，`0.0.1` 重点关注这些配置：

- `ANTHROPIC_API_KEY`：配置后启用真实 Anthropic 客户端
- `ANTHROPIC_MODEL`：默认 `claude-opus-4-7`
- `AGENT_MAX_ROUNDS`：单次任务最大循环轮数
- `AGENT_COMMAND_TIMEOUT_SECONDS`：命令工具超时秒数
- `AGENT_COMPACT_THRESHOLD`：上下文压缩触发阈值
- `AGENT_TRANSCRIPT_DIR`：transcript 输出目录

## CLI

目前支持三种基础入口：

- `go run ./cmd/assistant`：启动 REPL
- `go run ./cmd/assistant "your task"`：执行单轮任务
- `go run ./cmd/assistant --version` / `--help`：查看版本或帮助

## 版本计划

- `0.0.1` 开发计划与完成度评估见 [docs/v0.0.1-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.1-开发计划.md)
- 下一轮 `0.0.2` 版本计划见 [docs/v0.0.2-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.2-开发计划.md)
