# icoo_assistant

`icoo_assistant` 是一个基于 Go 的本地编码 Agent 原型，当前代码主体位于 [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)。

当前仓库已经完成 `0.0.7` 基线，能力范围包括：

- LLM 对话循环
- 工具注册与调用分发
- 工作区内文件读取、写入、编辑
- 本地命令执行
- Todo 管理
- 对话压缩与 transcript 落盘
- Skill 加载
- Subagent 摘要委托
- 项目级任务持久化骨架
- 项目任务工具入口
- 项目任务与后台执行关联展示
- 项目任务与后台执行基础状态联动
- 项目任务有限执行历史
- 后台命令执行与结果回流

## 仓库结构

- [docs](E:\codes\icoo_assistant\docs)：需求、开发计划、版本计划文档
- [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)：Go 模块源码
- [icoo_assistant/internal/task](E:\codes\icoo_assistant\icoo_assistant\internal\task)：项目级任务持久化模块
- [icoo_assistant/internal/background](E:\codes\icoo_assistant\icoo_assistant\internal\background)：后台任务模块

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

## Background 执行

后台命令执行能力现在已经和项目任务建立了基础关联，包含这些部分：

- [internal/background](E:\codes\icoo_assistant\icoo_assistant\internal\background)：后台任务管理器与结果通知
- [internal/tools/background.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\background.go)：后台任务工具入口
- [internal/agent/loop.go](E:\codes\icoo_assistant\icoo_assistant\internal\agent\loop.go)：主循环中的后台完成结果注入点

当前支持：

- 启动后台命令
- 查询单个后台任务状态
- 列出后台任务
- 使用 `task_id` 过滤后台任务
- 启动关联后台任务时推进任务状态
- 主循环自动轮询已完成后台任务并注入摘要结果

## Agent Hook

当前已经为 Agent 主循环补上基础 hook 埋点，默认会把事件写入工作区的 `.agent-hooks/events.jsonl`。这批埋点覆盖了：

- run started / completed / failed
- round started
- model requested / responded
- tool started / completed
- subagent started / completed
- compact auto / manual
- background notifications injected
- todo reminder injected

## Task 持久化

`0.0.7` 已经把项目任务和后台执行推进到了“基础状态联动 + 有限历史记录”。核心代码位于 [internal/task](E:\codes\icoo_assistant\icoo_assistant\internal\task)、[internal/tools/project_task.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\project_task.go) 和 [internal/background](E:\codes\icoo_assistant\icoo_assistant\internal\background)。当前支持：

- 初始化 `.tasks/` 目录
- 创建、读取、列出、更新任务
- `blockedBy` 依赖字段
- 任务完成后自动解除下游阻塞
- 通过 `project_task` 工具执行 `create / get / list / update / update_status`
- 通过 `task_id` 关联后台任务
- 查询项目任务时展示关联后台执行上下文
- 记录最近一次后台执行结果
- 保留有限条数的后台执行历史摘要
- 后台启动时将关联任务推进到 `in_progress`
- 后台失败时将 `in_progress` 任务退回 `pending`

命名上，`project_task` 负责项目级持久化任务，现有 `task` 仍负责子代理委托，这样能保持会话内规划、项目任务和子任务派发的职责边界清晰。

## 版本计划

- `0.0.1` 开发计划与完成度评估见 [docs/v0.0.1-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.1-开发计划.md)
- `0.0.2` 开发计划与完成度评估见 [docs/v0.0.2-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.2-开发计划.md)
- `0.0.3` 开发计划与完成度评估见 [docs/v0.0.3-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.3-开发计划.md)
- `0.0.4` 开发计划与完成度评估见 [docs/v0.0.4-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.4-开发计划.md)
- `0.0.5` 开发计划与完成度评估见 [docs/v0.0.5-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.5-开发计划.md)
- `0.0.6` 开发计划与完成度评估见 [docs/v0.0.6-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.6-开发计划.md)
- `0.0.7` 开发计划与完成度评估见 [docs/v0.0.7-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.7-开发计划.md)
- 下一轮 `0.0.8` 版本计划见 [docs/v0.0.8-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.8-开发计划.md)
