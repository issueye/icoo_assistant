# icoo_assistant

`icoo_assistant` 是一个基于 Go 的本地编码 Agent 原型，当前代码主体位于 [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)。

当前仓库已经完成 `0.1.4` 基线，能力范围包括：

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
- 项目任务历史查看入口
- 独立任务审计查询入口
- 工具职责目录与边界说明入口
- Agent Hook 审计查询入口
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

推荐先走一遍最小演示路径：

```bash
go run ./cmd/assistant --version
go run ./cmd/assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"
go run ./cmd/assistant "创建一个项目任务，用于验证后台测试"
go run ./cmd/assistant "使用 tool_catalog action=audit_paths 说明审计入口，再给出 task_audit 和 agent_hook_audit 的查询示例"
go run ./cmd/assistant "先用 agent_hook_audit action=summary 看最近运行摘要，再用 task_audit action=history status=failed 看失败任务历史"
go run ./cmd/assistant "先用 task_audit action=summary 看失败概况，再决定是否继续查看 task_audit action=history status=failed"
```

## 配置说明

Go 模块目录 [icoo_assistant/.env.example](E:\codes\icoo_assistant\icoo_assistant\.env.example) 提供了当前可用的环境变量模板，当前版本重点关注这些配置：

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

当前已经为 Agent 主循环补上基础 hook 埋点，默认会把事件写入工作区的 `.agent-hooks/events.jsonl`。`0.1.4` 继续把异常排障路径往前推了一步，除了运行时摘要外，还补上了任务侧的失败概况入口。当前埋点覆盖了：

- run started / completed / failed
- round started
- model requested / responded
- tool started / completed
- subagent started / completed
- compact auto / manual
- background notifications injected
- todo reminder injected

当前支持：

- 使用 `agent_hook_audit action=recent` 查看最近 hook 事件
- 使用 `agent_hook_audit action=summary` 查看最近 hook 事件摘要
- 使用 `name` 过滤特定事件名
- 使用 `run_id` 聚焦某一次运行
- 使用 `limit` 控制返回条数
- 使用 `tool_catalog action=audit_paths` 获取审计入口导航

## Task 持久化

`0.1.4` 继续把任务历史查询和工具边界做了收口。核心代码位于 [internal/task](E:\codes\icoo_assistant\icoo_assistant\internal\task)、[internal/tools/project_task.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\project_task.go)、[internal/tools/task_audit.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\task_audit.go)、[internal/tools/tool_catalog.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\tool_catalog.go)、[internal/tools/agent_hook_audit.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\agent_hook_audit.go) 和 [internal/background](E:\codes\icoo_assistant\icoo_assistant\internal\background)。当前支持：

- 初始化 `.tasks/` 目录
- 创建、读取、列出、更新任务
- `blockedBy` 依赖字段
- 任务完成后自动解除下游阻塞
- 通过 `project_task` 工具执行 `create / get / list / update / update_status`
- 通过 `task_id` 关联后台任务
- 查询项目任务时展示关联后台执行上下文
- 记录最近一次后台执行结果
- 保留有限条数的后台执行历史摘要
- 默认 `get` 输出保持紧凑
- 使用 `project_task action=history` 查看详细历史
- 使用 `task_audit action=summary` 查看任务执行概况与最近失败
- 使用 `task_audit action=history` 进行更独立的历史审计查询
- 使用 `task_audit action=history status=<status>` 聚焦特定执行状态
- 使用 `tool_catalog action=list|describe` 查看工具职责和推荐场景
- 使用 `tool_catalog action=audit_paths` 查看审计入口导航
- 后台启动时将关联任务推进到 `in_progress`
- 后台失败时将 `in_progress` 任务退回 `pending`

命名上，`project_task` 负责项目级持久化任务，现有 `task` 仍负责子代理委托，这样能保持会话内规划、项目任务和子任务派发的职责边界清晰。

## Tool 边界

为了让演示和上手路径更顺滑，`0.1.4` 继续把 `tool_catalog` 作为统一工具说明入口，并补上了更适合异常排障的失败概况用法。当前推荐的职责边界可以简单记成：

- `todo`：当前会话内的轻量步骤跟踪
- `project_task`：项目级持久化任务管理
- `task_audit`：项目任务执行历史审计
- `agent_hook_audit`：Agent 运行事件与排障审计
- `task`：子代理委托
- `background`：长时间运行命令
- `bash`：当前轮内应完成的快速命令

如果 Agent 对工具边界拿不准，可以先调用 `tool_catalog action=list`，再对具体工具执行 `tool_catalog action=describe`。

如果重点是“回看任务做了什么、Agent 又在运行时经历了什么”，可以优先走这条路径：

- `tool_catalog action=audit_paths`
- `task_audit action=summary`
- `agent_hook_audit action=summary`
- `project_task action=get` 或 `project_task action=history`
- `task_audit action=history`
- `task_audit action=history status=failed`
- `agent_hook_audit action=recent`

## 版本计划

- `0.0.1` 开发计划与完成度评估见 [docs/v0.0.1-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.1-开发计划.md)
- `0.0.2` 开发计划与完成度评估见 [docs/v0.0.2-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.2-开发计划.md)
- `0.0.3` 开发计划与完成度评估见 [docs/v0.0.3-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.3-开发计划.md)
- `0.0.4` 开发计划与完成度评估见 [docs/v0.0.4-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.4-开发计划.md)
- `0.0.5` 开发计划与完成度评估见 [docs/v0.0.5-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.5-开发计划.md)
- `0.0.6` 开发计划与完成度评估见 [docs/v0.0.6-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.6-开发计划.md)
- `0.0.7` 开发计划与完成度评估见 [docs/v0.0.7-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.7-开发计划.md)
- `0.0.8` 开发计划与完成度评估见 [docs/v0.0.8-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.8-开发计划.md)
- `0.0.9` 开发计划与完成度评估见 [docs/v0.0.9-开发计划.md](E:\codes\icoo_assistant\docs\v0.0.9-开发计划.md)
- `0.1.0` 开发计划与完成度评估见 [docs/v0.1.0-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.0-开发计划.md)
- `0.1.1` 开发计划与完成度评估见 [docs/v0.1.1-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.1-开发计划.md)
- `0.1.2` 开发计划与完成度评估见 [docs/v0.1.2-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.2-开发计划.md)
- `0.1.3` 开发计划与完成度评估见 [docs/v0.1.3-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.3-开发计划.md)
- `0.1.4` 开发计划与完成度评估见 [docs/v0.1.4-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.4-开发计划.md)
- 下一轮 `v0.1.5` 版本计划见 [docs/v0.1.5-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.5-开发计划.md)
