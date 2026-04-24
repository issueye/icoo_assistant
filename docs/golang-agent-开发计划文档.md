# 基于 learn-claude-code 教程的 Go 版 Agent 系统开发计划文档

## 1. 文档目标

本文档用于指导基于教程内容实现一个 Go 版 Agent Harness 系统，明确开发阶段、模块拆分、里程碑、交付物与验证方式，降低实现过程中的返工与偏航风险。

## 2. 总体开发策略

采用 **分阶段迭代** 开发策略：

- 先实现最小可用闭环，再逐步增加复杂能力。
- 先做单 Agent，再做多 Agent。
- 先做内聚的本地持久化，再做协作协议与隔离环境。
- 每一阶段都要求“可运行、可验证、可演示”。

## 3. 总体架构计划

### 3.1 架构原则
- 核心循环稳定，新增能力尽量通过模块和工具扩展完成。
- 所有状态优先外置到磁盘，避免只存在内存。
- 并发逻辑与业务逻辑分离。
- 模型提供决策，Harness 负责执行、安全和持久化。

### 3.2 目录规划建议

```text
icoo_assistant/
  cmd/
    assistant/
      main.go
  internal/
    agent/
    llm/
    tools/
    workspace/
    todo/
    compact/
    task/
    background/
    skill/
    team/
    protocol/
    worktree/
    storage/
  configs/
  skills/
  docs/
  testdata/
```

## 4. 开发阶段规划

## Phase 1：MVP 基础闭环

### 4.1 目标
建立可运行的单 Agent 基础框架，完成消息循环、基础工具执行与安全文件访问。

### 4.2 范围
- Agent Loop
- Tool Registry / Dispatch
- Bash/Command Tool
- Read/Write/Edit Tool
- Workspace Path Sandbox
- 基础日志

### 4.3 任务拆解
1. 定义 LLM Client 接口。
2. 实现消息结构与主循环。
3. 实现工具定义与注册机制。
4. 实现文件读写编辑工具。
5. 实现命令执行工具与超时控制。
6. 实现工作区路径校验。
7. 编写 CLI 入口和基础配置加载。

### 4.4 交付物
- 可运行 CLI 程序
- 最小工具集
- 单元测试与冒烟用例

### 4.5 验收标准
- 可通过自然语言驱动执行文件读取、写入、编辑和命令执行。
- 工具执行结果可正确回注到模型消息中。
- 禁止访问工作区外路径。

## Phase 2：会话内任务管理与长会话支持

### 4.6 目标
增强单 Agent 的任务规划能力和长会话持续工作能力。

### 4.7 范围
- TodoWrite
- Reminder 注入
- Context Compact
- Transcript 持久化
- 手动 compact 能力

### 4.8 任务拆解
1. 设计 Todo 数据结构。
2. 实现 Todo 更新校验规则。
3. 增加 nag reminder 机制。
4. 实现工具结果微压缩。
5. 实现超阈值自动摘要压缩。
6. 实现 transcript 落盘与恢复入口。

### 4.9 交付物
- 会话级 Todo 管理模块
- 压缩模块
- transcript 文件输出

### 4.10 验收标准
- 多步骤任务中 Todo 状态变化正确。
- 压缩触发后系统仍能继续执行任务。
- transcript 文件可被读取用于问题复盘。

## Phase 3：复杂任务能力扩展

### 4.11 目标
支持复杂任务拆分、按需知识加载和异步任务执行。

### 4.12 范围
- Subagent
- Skill Loader
- Background Tasks

### 4.13 任务拆解
1. 定义 Subagent 启动与摘要返回机制。
2. 实现 Skill 目录扫描与 frontmatter 解析。
3. 实现按需注入技能内容的工具。
4. 实现后台任务管理器。
5. 实现后台通知队列回流主循环。
6. 补充并发与错误路径测试。

### 4.14 交付物
- 子 Agent 模块
- 技能管理模块
- 后台执行模块

### 4.15 验收标准
- 主 Agent 可把任务委托给子 Agent 并接收摘要。
- 系统可识别技能列表并按需加载全文。
- 长耗时任务后台执行时，主循环不被阻塞。

## Phase 4：任务持久化系统

### 4.16 目标
将任务管理从会话级提升为项目级、持久化、可依赖的任务系统。

### 4.17 范围
- Task CRUD
- blockedBy 依赖
- owner/worktree 字段
- 任务状态流转

### 4.18 任务拆解
1. 设计 task JSON 结构。
2. 实现任务创建、查询、更新、列表功能。
3. 实现依赖解除逻辑。
4. 实现 owner 和 worktree 绑定字段。
5. 增加任务目录初始化与恢复。

### 4.19 交付物
- `.tasks/` 任务持久化模块
- 任务状态查询接口

### 4.20 验收标准
- 可创建依赖任务图。
- 上游任务完成后，下游任务自动解锁。
- 重启程序后任务状态不丢失。

## Phase 5：多 Agent 协作

### 4.21 目标
支持长期存活的队友 Agent、消息通信和团队状态管理。

### 4.22 范围
- Agent Teams
- Teammate Lifecycle
- Message Bus / Inbox
- Team Config

### 4.23 任务拆解
1. 设计 team config 结构。
2. 实现 teammate spawn/shutdown 基础能力。
3. 实现 inbox JSONL 消息总线。
4. 实现队友循环与状态更新。
5. 增加 lead 与 teammate 的协作命令。

### 4.24 交付物
- `.team/` 目录结构
- 队友管理模块
- 消息总线模块

### 4.25 验收标准
- 可同时运行多个队友 Agent。
- 队友间可稳定收发消息。
- 队友状态可被查询和展示。

## Phase 6：团队协议与自治

### 4.26 目标
让多 Agent 协作具备审批机制、自主任务认领能力和更稳定的长时间运行特性。

### 4.27 范围
- Request-Response Protocol
- Shutdown Approval
- Plan Approval
- Idle Polling
- Auto Claim Task
- Identity Reinjection

### 4.28 任务拆解
1. 设计请求/响应协议消息格式。
2. 实现 request_id 跟踪表。
3. 实现关机审批与计划审批工具。
4. 实现 IDLE 阶段轮询逻辑。
5. 实现自动认领任务逻辑。
6. 实现压缩后的身份重注入机制。

### 4.29 交付物
- 协议模块
- 自治执行模块
- 任务认领模块

### 4.30 验收标准
- 可发起审批、收到响应并更新状态。
- 空闲 Agent 可自动认领未阻塞任务。
- 会话压缩后 Agent 身份不会丢失。

## Phase 7：Worktree 隔离执行

### 4.31 目标
实现任务与代码工作目录的一一绑定，支撑隔离式并行开发。

### 4.32 范围
- Worktree Registry
- Task-Worktree Binding
- Worktree Keep/Remove
- Lifecycle Events
- Isolated Command/File Execution

### 4.33 任务拆解
1. 设计 worktree 索引结构。
2. 封装 git worktree 创建、查询、删除命令。
3. 实现任务与 worktree 自动绑定。
4. 实现 worktree 事件日志。
5. 实现指定 worktree 下的命令和文件操作。
6. 增加异常补偿逻辑。

### 4.34 交付物
- `.worktrees/index.json`
- `.worktrees/events.jsonl`
- worktree 管理模块

### 4.35 验收标准
- 可为任务创建独立 worktree。
- 可在指定 worktree 中执行命令。
- 删除 worktree 时可同步完成任务或解绑。

## 5. 模块开发顺序建议

推荐按以下顺序实现，减少返工：

1. `llm`
2. `agent`
3. `workspace`
4. `tools`
5. `todo`
6. `compact`
7. `skill`
8. `background`
9. `task`
10. `team`
11. `protocol`
12. `worktree`

原因：该顺序从主干闭环开始，逐步扩展为复杂能力，能保证每阶段都基于稳定底座演进。

## 6. 测试计划

### 6.1 单元测试
- 路径沙箱校验
- 文件工具逻辑
- 工具注册与分发
- Todo 状态校验
- Task 依赖解除
- Skill frontmatter 解析
- 请求响应状态机

### 6.2 集成测试
- 单 Agent 任务执行流
- Subagent 委托流
- 后台任务通知流
- Team 收件箱通信流
- Task + Worktree 绑定流

### 6.3 端到端测试
场景建议：
1. 用户要求修改项目文件并验证。
2. 用户要求拆分任务并后台运行测试。
3. 用户要求多个 Agent 协作完成两个独立任务。
4. 用户要求在两个 worktree 中并行处理需求。

## 7. 风险控制计划

### 7.1 模型输出不可控
措施：
- 增加最大循环轮数。
- 对工具参数做强校验。
- 对危险动作添加确认层。

### 7.2 并发状态冲突
措施：
- 对任务、消息总线、事件日志采用串行写入或锁控制。
- 避免多个 goroutine 直接写同一文件。

### 7.3 Worktree 清理不完整
措施：
- 引入 before/after/failed 事件。
- 失败时保留现场，允许人工恢复。

### 7.4 长会话性能下降
措施：
- 分层压缩。
- 限制每轮注入的工具结果大小。
- 摘要与转储并行处理。

## 8. 建议里程碑

### M1：基础闭环可运行
完成 Phase 1，支持基本工具操作。

### M2：单 Agent 稳定工作
完成 Phase 2 和 Phase 3，支持 Todo、压缩、子任务、技能和后台任务。

### M3：项目级任务管理
完成 Phase 4，支持任务图和持久化。

### M4：多 Agent 协作
完成 Phase 5 和 Phase 6，支持队友、协议和自治。

### M5：隔离式并行开发
完成 Phase 7，支持 worktree 隔离执行。

## 9. 人员角色建议

- **架构/核心开发**：负责 agent loop、工具系统、持久化与并发框架。
- **平台开发**：负责 task/team/protocol/worktree 模块。
- **测试工程**：负责端到端场景、恢复性测试、并发稳定性测试。
- **产品/技术负责人**：定义各阶段验收标准与优先级取舍。

## 10. 预计交付节奏建议

如果按单团队推进，可参考以下节奏：

- 第 1 周：Phase 1
- 第 2 周：Phase 2
- 第 3 周：Phase 3 + Phase 4
- 第 4 周：Phase 5
- 第 5 周：Phase 6
- 第 6 周：Phase 7 + 系统联调

## 11. 最终交付清单

1. Go 项目源码
2. CLI 可执行程序
3. 示例配置文件
4. skills 示例目录
5. 测试用例
6. 使用说明文档
7. 演示脚本或 demo 场景
8. 架构图和状态流说明

## 12. 总结

该项目的正确开发方式不是一次性堆满全部能力，而是沿着教程的能力演进链条逐层实现。建议优先把单 Agent 闭环做稳，再逐步补齐任务持久化、多 Agent 协作、自治机制和 worktree 隔离，这样最符合教程本身的设计逻辑，也最适合 Go 在并发与工程化上的优势发挥。
