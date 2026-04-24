# icoo_assistant

`icoo_assistant` 是一个基于 Go 的本地编码 Agent 原型，当前代码主体位于 [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)。

当前仓库已经在 `0.1.30` 初代可用版基线上完成 `0.1.31` 与 `0.1.32` 两轮 post-release hardening，并在 `0.1.33-v0.1.36` 正式切入 Team / Protocol 主线，能力范围包括：

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
- `.team/` 基础结构与 teammate registry 骨架
- `.team/inbox/` message store 与基础查询入口
- lead -> teammate 最小 request / response 线程
- `.team/requests/` request lifecycle 持久化与协议查询入口
- 工具职责目录与边界说明入口
- Agent Hook 审计查询入口
- 后台命令执行与结果回流

## 仓库结构

- [docs](E:\codes\icoo_assistant\docs)：需求、开发计划、版本计划文档
- [icoo_assistant](E:\codes\icoo_assistant\icoo_assistant)：Go 模块源码
- [icoo_assistant/internal/task](E:\codes\icoo_assistant\icoo_assistant\internal\task)：项目级任务持久化模块
- [icoo_assistant/internal/background](E:\codes\icoo_assistant\icoo_assistant\internal\background)：后台任务模块
- [icoo_assistant/internal/team](E:\codes\icoo_assistant\icoo_assistant\internal\team)：Team config 与 teammate registry 模块

## 快速开始

```bash
cd icoo_assistant
go test ./...
cp .env.example .env
go run ./cmd/assistant --version
go run ./cmd/assistant check
go run ./cmd/assistant --help
go run ./cmd/assistant
```

如果配置了 `ANTHROPIC_API_KEY`，程序会使用真实 Anthropic 客户端；否则会回退到 fake client，方便先验证本地框架是否跑通。

推荐先走一遍最小 happy path：

```bash
go run ./cmd/assistant check
go run ./cmd/assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"
go run ./cmd/assistant "创建一个项目任务，用于验证后台测试"
go run ./cmd/assistant "使用 tool_catalog action=audit_paths 说明审计入口，再给出 task_audit 和 agent_hook_audit 的查询示例"
```

如果 `assistant check` 输出里显示当前是 `mode=fake`，上面这条最小路径仍然可以作为本地 dry run 参考；如果你想获得完整的真实 Agent 行为，建议先配置 `ANTHROPIC_API_KEY` 后再重新执行 `check`。

`0.1.31` 开始，`assistant --help` 和 `assistant check` 会显示同一套首次上手顺序；执行完 `assistant check` 后，自检输出也会直接提示“从第 2 步继续”，减少 README、帮助信息和 CLI 输出之间的理解落差。

`0.1.32` 开始，CLI 默认会按源码仓库场景展示 `go run ./cmd/assistant ...` 形式的命令，并额外提示“如果已经安装二进制，可把前缀替换成 assistant”，避免用户从 README 进入 CLI 后又被命令形式切换绊住。

`0.1.33` 开始，运行时会初始化 `.team/` 基础目录，并提供 `team_registry` 工具来管理 `lead` 配置和 teammate 名册，为后续 inbox / message bus 做准备。

`0.1.34` 开始，运行时会初始化 `.team/inbox/`，并提供 `team_message` 工具来持久化 lead / teammate 消息和查看收件箱，为下一版的最小通信链路做准备。

`0.1.35` 开始，`team_message` 已支持 `request / reply / thread`，因此现在已经可以把 lead -> teammate -> lead 的最小通信样本完整落到磁盘并回看整条线程。

`0.1.36` 开始，运行时会初始化 `.team/requests/`，并提供 `team_protocol` 工具来查看 request 的 `pending / responded` 生命周期，让 Team request 不再只是 inbox 里的原始消息。

要真正完成“由模型驱动的最小闭环”，仍然建议在 `anthropic mode` 下执行这条路径；`fake mode` 更适合确认 CLI、自检、文档路径和降级提示是否符合预期。

`0.1.29` 开始，如果你在 fake 模式下直接运行单轮命令或进入 REPL，CLI 也会直接提示当前处于降级模式，并明确说明“空输出是 fake client 的预期行为”，避免把它误判成卡死或无响应。

## 初代可用版定位

`0.1.36` 的定位仍然不是“功能很多”，而是“在保持最小闭环清晰的前提下，把 Team request 变成带生命周期的协议对象”。当前版本更适合：

- 本地验证 Agent 骨架是否能启动和自检
- 通过固定 happy path 理解工具边界
- 创建项目任务并理解审计入口
- 初始化 team config 并登记最小 teammate roster
- 向 teammate inbox 写入和查看最小消息样本
- 走通一次 lead request -> teammate reply -> thread review
- 查询 request 当前是 `pending` 还是 `responded`
- 在 fake 模式下做 dry run，或在真实 API Key 下做实际模型调用

当前版本还不追求：

- 安装器或桌面封装
- 多模型适配
- 自动联网诊断或自动修复配置
- 面向复杂场景的完整工作流模板

更多演示命令：

```bash
go run ./cmd/assistant --version
go run ./cmd/assistant --help
go run ./cmd/assistant "先用 agent_hook_audit action=summary 看最近运行摘要，再用 task_audit action=history status=failed 看失败任务历史"
go run ./cmd/assistant "先用 task_audit action=summary 看失败概况，再决定是否继续查看 task_audit action=history status=failed"
go run ./cmd/assistant "先用 task_audit action=summary status=failed 看失败原因分类，再决定是否继续查看失败历史"
go run ./cmd/assistant "先用 task_audit action=summary status=failed 对比各失败原因的最近样本，再决定下一步复盘哪一类失败"
go run ./cmd/assistant "先用 task_audit action=summary status=failed 看最近失败趋势，再决定是否继续查看详细失败历史"
go run ./cmd/assistant "先用 task_audit action=summary 看 priority_failure_reason、priority_failure_basis、priority_failure_context、priority_failure_pattern_hint、priority_failure_sample_target、priority_failure_sample_compare、priority_failure_compare_target、priority_failure_change_hint、priority_failure_trend_hint 和 priority_failure_hint，再决定先排查哪类失败"
go run ./cmd/assistant "先用 task_audit action=history id=task-a reason=timeout limit=2 查看最近两条超时失败，并关注 latest_sample、latest_failure_command、latest_failure_error、latest_failure_signature、latest_failure_updated_at、latest_failure_entry、reason=timeout、pair_summary、role=previous 与 role=latest"
go run ./cmd/assistant "先用 task_audit action=summary reason=timeout 聚焦超时失败，再决定是否继续查看 task_audit action=history reason=timeout"
go run ./cmd/assistant "先用 tool_catalog action=describe name=team_registry 说明 .team 的边界，再创建 teammate alice，role reviewer，model claude-opus-4-7"
go run ./cmd/assistant "使用 team_message action=send to=alice kind=request body='请先阅读最新计划'，然后查看 team_message action=inbox recipient=alice"
go run ./cmd/assistant "先用 team_message action=request to=alice body='请确认最新计划' request_id=req-demo 创建请求，再用 team_message action=reply from=alice request_id=req-demo body='已确认' 返回响应，最后查看 team_message action=thread request_id=req-demo"
go run ./cmd/assistant "先用 team_protocol action=summary 看当前请求概况，再用 team_protocol action=list status=pending 查看未响应请求"
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

目前支持这些基础入口：

- `go run ./cmd/assistant`：启动 REPL
- `go run ./cmd/assistant "your task"`：执行单轮任务
- `go run ./cmd/assistant check`：执行环境自检并确认最小运行前提
- `go run ./cmd/assistant doctor`：`check` 的等价别名
- `go run ./cmd/assistant --version` / `--help`：查看版本或帮助

如果已经安装了 `assistant` 二进制，也可以把上面这些命令里的 `go run ./cmd/assistant` 替换成 `assistant`。

## 模式边界

当前版本有两种最重要的运行模式：

- `fake mode`：
  当没有配置 `ANTHROPIC_API_KEY` 时启用。适合做本地 dry run、验证 CLI、自检、任务入口和审计路径，但不会提供真实模型回答。
- `anthropic mode`：
  当配置了 `ANTHROPIC_API_KEY` 时启用。适合执行真实的模型调用与实际问答流程。

可以通过 `go run ./cmd/assistant check` 直接确认当前处于哪种模式。

## 环境自检

`0.1.27` 开始，CLI 增加了 `check` 自检入口，用来在真正运行 Agent 之前快速确认当前仓库是否具备最小可用前提。建议先执行：

```bash
go run ./cmd/assistant check
```

当前自检会直接给出这些信息：

- 当前工作区与 `.env` 状态
- 当前会使用 `fake` 还是 `anthropic` client
- `skills` 目录是否已配置
- `.transcripts`、`.tasks`、`.team`、`.team/inbox`、`.team/requests`、`.background`、`.agent-hooks` 这些运行目录是否已就绪
- 固定的 `minimal_happy_path`
- 下一步建议直接运行什么命令

## 最小 Happy Path

如果你只是想确认这个仓库已经到了“可以开始用”的状态，建议固定按这条顺序走：

1. `go run ./cmd/assistant check`
2. `go run ./cmd/assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"`
3. `go run ./cmd/assistant "创建一个项目任务，用于验证后台测试"`
4. `go run ./cmd/assistant "使用 tool_catalog action=audit_paths 说明审计入口，再给出 task_audit 和 agent_hook_audit 的查询示例"`

这条路径对应的目标分别是：

- 第 1 步先确认环境是否具备最小运行前提
- 第 2 步先理解工具边界
- 第 3 步确认任务入口可用
- 第 4 步确认任务侧和运行时侧的审计入口都能被正确指引

补充说明：

- 在 `anthropic mode` 下，这条路径可以作为真实最小闭环使用
- 在 `fake mode` 下，这条路径更适合作为 dry run 和上手路径验证

## 常见卡点

如果最小 happy path 没有按预期跑通，优先看这几类情况：

- `assistant check` 显示 `mode=fake`：
  这表示当前没有启用真实模型调用。你仍然可以做本地 dry run，但如果你期待真实回答，需要先配置 `ANTHROPIC_API_KEY`，然后重新执行 `go run ./cmd/assistant check`。
- 单轮命令看起来“没有输出”：
  `0.1.29` 起，CLI 会直接提示 fake client 在降级模式下默认不会生成真实回答。这通常不是卡死，而是当前就在 fake 模式。
- REPL 能启动，但回答看起来像没反应：
  如果启动时已经看到 fake mode warning，就优先把它当作降级模式提示，而不是运行异常。先退出 REPL，执行 `go run ./cmd/assistant check` 确认当前模式和最小路径。
- 最小 happy path 在第 2 到 4 步中断：
  先重新执行 `go run ./cmd/assistant check`，确认 `.tasks`、`.background`、`.agent-hooks` 这些目录仍是 ready，再重试对应步骤。

## 当前限制

为了让 `0.1.30` 保持一个可交付但边界清楚的初代版本，当前有这些明确限制：

- fake 模式不会生成真实回答，只用于 dry run 与路径验证
- 完整的模型驱动最小闭环仍然依赖 `anthropic mode`
- 当前主要围绕 Anthropic client 构建，还没有多模型切换
- 当前没有安装器、桌面 GUI 或一键初始化流程
- 当前的排障以本地提示和 README 为主，不包含联网探测
- 当前只到了 Team config、teammate registry、inbox message store、最小 request/response thread 和基础 request lifecycle，还没有真实 teammate loop、协议审批和自治认领
- 当前更适合最小闭环验证与能力理解，不等价于成熟生产 Agent

## v0.1.36 验证清单

下面这组检查项就是当前版本的交付口径：

- [x] `go run ./cmd/assistant --version` 正常
- [x] `go run ./cmd/assistant --help` 正常
- [x] `go run ./cmd/assistant check` 能给出明确结果
- [x] `assistant check`、`--help` 与 README 的首次上手路径编号一致
- [x] `assistant check` 会明确提示从第 2 步继续
- [x] CLI 默认输出的命令形式与 README 的 `go run ./cmd/assistant` 用法一致
- [x] CLI 会明确提示已安装二进制时可替换成 `assistant`
- [x] `.team/` 基础目录与默认 team config 会自动初始化
- [x] 已提供 `team_registry` 工具用于管理 team config 与 teammate roster
- [x] `.team/inbox/` message store 会自动初始化
- [x] `.team/requests/` request store 会自动初始化
- [x] 已提供 `team_message` 工具用于写入和查看 teammate inbox
- [x] 已提供 `team_message request / reply / thread` 形成最小 lead -> teammate 通信链路
- [x] 已提供 `team_protocol` 工具用于查看 request lifecycle 与 pending/responded 状态
- [x] README 已提供固定的最小 happy path
- [x] 用户可以通过 README 找到项目任务、后台执行和审计入口边界
- [x] fake 模式与真实模式的边界已清楚写明
- [x] `go test ./...` 通过

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

当前已经为 Agent 主循环补上基础 hook 埋点，默认会把事件写入工作区的 `.agent-hooks/events.jsonl`。`0.1.30` 作为初代可用版基线，继续保留了最近几版补齐的 `assistant check`、固定 `minimal_happy_path`、fake 模式明确反馈，以及任务失败复盘能力，包括 `latest_sample`、`latest_failure_command`、`latest_failure_error`、`latest_failure_signature`、`latest_failure_updated_at`、`latest_failure_entry`、`role=previous` / `role=latest` 和 `pair_summary`。当前埋点覆盖了：

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

`0.1.30` 继续把重点放在“可用版交付”的主线上。核心代码位于 [internal/task](E:\codes\icoo_assistant\icoo_assistant\internal\task)、[internal/tools/project_task.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\project_task.go)、[internal/tools/task_audit.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\task_audit.go)、[internal/tools/tool_catalog.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\tool_catalog.go)、[internal/tools/agent_hook_audit.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\agent_hook_audit.go)、[internal/background](E:\codes\icoo_assistant\icoo_assistant\internal\background) 和 [cmd/assistant](E:\codes\icoo_assistant\icoo_assistant\cmd\assistant)。当前支持：

- 初始化 `.tasks/` 目录
- 使用 `assistant check` 进行环境自检
- 使用 `assistant check` 查看固定的最小 happy path
- 在 fake 模式或空输出场景下获得更明确的 CLI 提示
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
- 在 `task_audit action=summary` 中查看基础失败原因分类
- 在 `task_audit action=summary` 中对比各失败原因的最近样本
- 在 `task_audit action=summary` 中查看最近失败趋势
- 在 `task_audit action=summary` 中查看 `priority_failure_reason` 与 `priority_failure_hint`
- 在 `task_audit action=summary` 中查看 `priority_failure_basis`
- 在 `task_audit action=summary` 中查看 `priority_failure_context`
- 在 `task_audit action=summary` 中查看 `priority_failure_pattern_hint`
- 在 `task_audit action=summary` 中查看 `priority_failure_sample_target`
- 在 `task_audit action=summary` 中查看 `priority_failure_sample_compare`
- 在 `task_audit action=summary` 中查看 `priority_failure_compare_target`
- 在 `task_audit action=summary` 中查看 `priority_failure_change_hint`
- 在 `task_audit action=summary` 中查看 `priority_failure_trend_hint`
- 在 `task_audit action=history` 中查看 `latest_sample`
- 在 `task_audit action=history` 中查看 `latest_failure_command`
- 在 `task_audit action=history` 中查看 `latest_failure_error`
- 在 `task_audit action=history` 中查看 `latest_failure_signature`
- 在 `task_audit action=history` 中查看 `latest_failure_updated_at`
- 在 `task_audit action=history` 中查看 `latest_failure_entry`
- 在 `task_audit action=history` 的失败条目中直接查看 `reason=<reason>`
- 在 `task_audit action=history` 的聚焦双样本场景中查看 `role=previous` / `role=latest`
- 在 `task_audit action=history` 的聚焦双样本场景中查看 `pair_summary`
- 使用 `task_audit action=summary reason=<reason>` 聚焦某一类失败原因
- 使用 `task_audit action=history` 进行更独立的历史审计查询
- 使用 `task_audit action=history status=<status>` 聚焦特定执行状态
- 使用 `task_audit action=history reason=<reason>` 聚焦特定失败原因
- 使用 `tool_catalog action=list|describe` 查看工具职责和推荐场景
- 使用 `tool_catalog action=audit_paths` 查看审计入口导航
- 后台启动时将关联任务推进到 `in_progress`
- 后台失败时将 `in_progress` 任务退回 `pending`

命名上，`project_task` 负责项目级持久化任务，现有 `task` 仍负责子代理委托，这样能保持会话内规划、项目任务和子任务派发的职责边界清晰。

## Team 基础设施

`0.1.33` 是从单 Agent 主线进入 Team 主线的第一版，当前核心代码位于 [internal/team](E:\codes\icoo_assistant\icoo_assistant\internal\team)、[internal/tools/team_registry.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\team_registry.go) 和 [cmd/assistant/check.go](E:\codes\icoo_assistant\icoo_assistant\cmd\assistant\check.go)。当前支持：

- 初始化 `.team/` 目录
- 初始化 `.team/config.json`
- 初始化 `.team/teammates/` registry 目录
- 默认生成 `lead_id=lead` 的 team config
- 通过 `team_registry action=get_config|update_config` 查询或更新 team config
- 通过 `team_registry action=create|get|list|update` 管理 teammate roster
- 在 `assistant check` 中查看 `.team/` 是否 ready，以及当前 `lead_id` 和 `teammate_count`

当前还不支持：

- 真正运行中的 message bus / teammate loop
- 自动消费 inbox 并执行消息
- 收件确认与自动消费
- 协议审批、auto claim、identity reinjection

## Team Message Store

`0.1.35` 在 Team 主线里把消息承载层推进成了最小通信闭环，当前核心代码位于 [internal/tools/team_message.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\team_message.go) 和 [internal/team](E:\codes\icoo_assistant\icoo_assistant\internal\team)。当前支持：

- 初始化 `.team/inbox/` 目录
- 按 recipient 在 `.team/inbox/<recipient>.jsonl` 下持久化消息
- 通过 `team_message action=send` 写入 lead -> teammate 或 teammate -> lead 消息
- 通过 `team_message action=request` 创建带 `request_id` 的请求
- 通过 `team_message action=reply` 基于既有请求返回响应
- 通过 `team_message action=inbox` 查看某个 recipient 的最近消息
- 通过 `team_message action=thread` 查看某个 `request_id` 的完整线程
- 在 `assistant check` 中查看 inbox 目录是否 ready

## Team Protocol

`0.1.36` 开始，Team 主线进入最小协议层。当前核心代码位于 [internal/team/request.go](E:\codes\icoo_assistant\icoo_assistant\internal\team\request.go) 和 [internal/tools/team_protocol.go](E:\codes\icoo_assistant\icoo_assistant\internal\tools\team_protocol.go)。当前支持：

- 初始化 `.team/requests/` 目录
- 在 `team_message action=request` 时自动创建持久化 request record
- 在 `team_message action=reply` 时自动把 request 状态从 `pending` 更新为 `responded`
- 通过 `team_protocol action=get` 查看单个 request 的协议状态
- 通过 `team_protocol action=list` 按 `status`、`from`、`to` 过滤 request
- 通过 `team_protocol action=summary` 查看当前 request 概况
- 在 `assistant check` 中查看 request 目录是否 ready

当前还不支持：

- 真正运行中的 teammate 轮询 inbox
- 收件确认、已读状态或更复杂的协议状态机
- lead -> teammate 的自动执行与结果回流

## Tool 边界

为了让演示和上手路径更顺滑，`0.1.30` 继续沿着 `assistant check -> minimal_happy_path -> 常见卡点反馈 -> 任务与审计入口` 这条主线收口，`tool_catalog` 仍然负责工具边界说明。`docs` 目录现在也增加了 [版本计划写法说明.md](E:\codes\icoo_assistant\docs\版本计划写法说明.md)、[v0.1.30-初代可用版路线图.md](E:\codes\icoo_assistant\docs\v0.1.30-初代可用版路线图.md) 和 [v0.1.30-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.30-开发计划.md)。当前推荐的职责边界可以简单记成：

- `todo`：当前会话内的轻量步骤跟踪
- `project_task`：项目级持久化任务管理
- `team_registry`：`.team/` 下的 team config 与 teammate registry 管理
- `team_message`：`.team/inbox/` 下的持久化消息写入、request/reply 与 thread 查看
- `team_protocol`：`.team/requests/` 下的 request lifecycle 查询与协议状态查看
- `task_audit`：项目任务执行历史审计
- `agent_hook_audit`：Agent 运行事件与排障审计
- `task`：子代理委托
- `background`：长时间运行命令
- `bash`：当前轮内应完成的快速命令

如果 Agent 对工具边界拿不准，可以先调用 `tool_catalog action=list`，再对具体工具执行 `tool_catalog action=describe`。

如果重点是“回看任务做了什么、Agent 又在运行时经历了什么”，可以优先走这条路径：

- `assistant check`
- `tool_catalog action=audit_paths`
- `task_audit action=summary`
- `task_audit action=summary status=failed`
- `task_audit action=summary status=failed` 并重点看 `recent_failure_trend`
- `task_audit action=summary` 并重点看 `priority_failure_reason`
- `task_audit action=summary` 并结合 `priority_failure_basis`
- `task_audit action=summary` 并结合 `priority_failure_context`
- `task_audit action=summary` 并结合 `priority_failure_pattern_hint`
- `task_audit action=summary` 并结合 `priority_failure_sample_target`
- `task_audit action=summary` 并结合 `priority_failure_sample_compare`
- `task_audit action=summary` 并结合 `priority_failure_compare_target`
- `task_audit action=summary` 并结合 `priority_failure_change_hint`
- `task_audit action=summary` 并结合 `priority_failure_trend_hint`
- `task_audit action=history reason=timeout limit=2` 并结合 `latest_sample`、`latest_failure_command`、`latest_failure_error`、`latest_failure_signature`、`latest_failure_updated_at`、`latest_failure_entry`、`reason=timeout`、`pair_summary`、`role=previous` / `role=latest`
- `task_audit action=summary reason=timeout`
- `agent_hook_audit action=summary`
- `project_task action=get` 或 `project_task action=history`
- `task_audit action=history`
- `task_audit action=history status=failed`
- `task_audit action=history reason=timeout`
- `agent_hook_audit action=recent`

## 版本计划

`v0.1.30` 的目标已经调整为“交付一个可以使用的初代版本”，对应倒排路线见 [docs/v0.1.30-初代可用版路线图.md](E:\codes\icoo_assistant\docs\v0.1.30-初代可用版路线图.md)。

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
- `0.1.5` 开发计划与完成度评估见 [docs/v0.1.5-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.5-开发计划.md)
- `0.1.6` 开发计划与完成度评估见 [docs/v0.1.6-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.6-开发计划.md)
- `0.1.7` 开发计划与完成度评估见 [docs/v0.1.7-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.7-开发计划.md)
- `0.1.8` 开发计划与完成度评估见 [docs/v0.1.8-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.8-开发计划.md)
- `0.1.9` 开发计划与完成度评估见 [docs/v0.1.9-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.9-开发计划.md)
- `0.1.10` 开发计划与完成度评估见 [docs/v0.1.10-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.10-开发计划.md)
- `0.1.11` 开发计划与完成度评估见 [docs/v0.1.11-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.11-开发计划.md)
- `0.1.12` 开发计划与完成度评估见 [docs/v0.1.12-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.12-开发计划.md)
- `0.1.13` 开发计划与完成度评估见 [docs/v0.1.13-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.13-开发计划.md)
- `0.1.14` 开发计划与完成度评估见 [docs/v0.1.14-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.14-开发计划.md)
- `0.1.15` 开发计划与完成度评估见 [docs/v0.1.15-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.15-开发计划.md)
- `0.1.16` 开发计划与完成度评估见 [docs/v0.1.16-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.16-开发计划.md)
- `0.1.17` 开发计划与完成度评估见 [docs/v0.1.17-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.17-开发计划.md)
- `0.1.18` 开发计划与完成度评估见 [docs/v0.1.18-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.18-开发计划.md)
- `0.1.19` 开发计划与完成度评估见 [docs/v0.1.19-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.19-开发计划.md)
- `0.1.20` 开发计划与完成度评估见 [docs/v0.1.20-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.20-开发计划.md)
- `0.1.21` 开发计划与完成度评估见 [docs/v0.1.21-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.21-开发计划.md)
- `0.1.22` 开发计划与完成度评估见 [docs/v0.1.22-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.22-开发计划.md)
- `0.1.23` 开发计划与完成度评估见 [docs/v0.1.23-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.23-开发计划.md)
- `0.1.24` 开发计划与完成度评估见 [docs/v0.1.24-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.24-开发计划.md)
- `0.1.25` 开发计划与完成度评估见 [docs/v0.1.25-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.25-开发计划.md)
- `0.1.26` 开发计划与完成度评估见 [docs/v0.1.26-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.26-开发计划.md)
- `0.1.27` 开发计划与完成度评估见 [docs/v0.1.27-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.27-开发计划.md)
- `0.1.28` 开发计划与完成度评估见 [docs/v0.1.28-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.28-开发计划.md)
- `0.1.29` 开发计划与完成度评估见 [docs/v0.1.29-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.29-开发计划.md)
- `0.1.30` 开发计划与完成度评估见 [docs/v0.1.30-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.30-开发计划.md)
- `0.1.31` 开发计划与完成度评估见 [docs/v0.1.31-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.31-开发计划.md)
- `0.1.32` 开发计划与完成度评估见 [docs/v0.1.32-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.32-开发计划.md)
- `0.1.33` 开发计划与完成度评估见 [docs/v0.1.33-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.33-开发计划.md)
- `0.1.34` 开发计划与完成度评估见 [docs/v0.1.34-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.34-开发计划.md)
- `0.1.35` 开发计划与完成度评估见 [docs/v0.1.35-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.35-开发计划.md)
- `0.1.36` 开发计划与完成度评估见 [docs/v0.1.36-开发计划.md](E:\codes\icoo_assistant\docs\v0.1.36-开发计划.md)
- `v0.1.30` 初代可用版交付路线见 [docs/v0.1.30-初代可用版路线图.md](E:\codes\icoo_assistant\docs\v0.1.30-初代可用版路线图.md)
- `v0.1.31+` 后续迭代路线见 [docs/v0.1.31+-推荐版本路线图.md](E:\codes\icoo_assistant\docs\v0.1.31+-推荐版本路线图.md)
