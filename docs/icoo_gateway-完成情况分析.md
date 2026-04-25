# icoo_gateway 完成情况分析

## 1. 文档说明

- 文档日期：2026-04-25
- 分析对象：`icoo_gateway`
- 对照依据：[网关管理服务-开发设计文档](E:/codes/icoo_assistant/docs/网关管理服务-开发设计文档.md)

本文档基于 `docs/网关管理服务-开发设计文档.md`，结合当前仓库中的 `icoo_gateway` 实际代码实现，对其完成情况进行阶段性评估，用于后续排期、评审和开发决策。

## 2. 总体结论

按设计文档的完整目标来看，`icoo_gateway` 当前更接近“Phase 1 到 Phase 3 的内存版管理面骨架”，还不属于完整可投入使用的网关管理服务。

当前可给出以下判断：

- 按全量设计目标估算，整体完成度约为 `40%`
- 按“服务骨架 + 最小管理 API 骨架”估算，完成度约为 `70%`
- 当前已经具备基础资源管理、统一会话模型、最小 team routing 和 agent heartbeat
- 当前主要缺口集中在持久化、鉴权、审计、运行集成、真实编排闭环等平台能力

一句话判断：

> `icoo_gateway` 已经具备“证明设计方向可行”的基础骨架，但距离设计文档定义的完整控制面服务还有明显工程化差距。

## 3. 已完成的能力

### 3.1 服务骨架

以下能力已经具备：

- 独立 Go 模块
- 基础配置加载
- HTTP 服务启动
- `/healthz` 健康检查
- 服务根路由与 API 路由注册

对应实现位置：

- [cmd/icoo_gateway/main.go](E:/codes/icoo_assistant/icoo_gateway/cmd/icoo_gateway/main.go)
- [internal/config/config.go](E:/codes/icoo_assistant/icoo_gateway/internal/config/config.go)
- [internal/server/server.go](E:/codes/icoo_assistant/icoo_gateway/internal/server/server.go)
- [internal/api/router.go](E:/codes/icoo_assistant/icoo_gateway/internal/api/router.go)

### 3.2 基础资源 API

当前已经具备以下资源的内存版基础 API：

- `skill`：`create / list / get`
- `agent_profile`：`create / list / get`
- `agent_instance`：`create / list / get / heartbeat`
- `team`：`create / list / get`
- `team member`：`add / list`
- `conversation`：`create / list / get`
- `conversation message`：`append / list`

项目自述中也明确说明了当前范围，见：

- [icoo_gateway/README.md](E:/codes/icoo_assistant/icoo_gateway/README.md)

### 3.3 统一对话模型

设计文档强调单对话和 Team 对话统一建模，这一点已经开始落实：

- `conversation.mode` 支持 `single` 和 `team`
- `message.scope` 支持 `external / internal / system`
- 支持按 `scope` 查询消息

对应实现位置：

- [internal/conversation/service.go](E:/codes/icoo_assistant/icoo_gateway/internal/conversation/service.go)

### 3.4 Team 最小路由骨架

当前 Team 对话已有最小可运行 routing 行为：

- Team 对话收到外部消息后，会写入一条发给入口 Agent 的 `internal` 分派消息
- 同时会写入一条 `system` 汇总占位消息
- 如果 team 不存在、没有 `entry_agent_id`，或者入口 Agent 不属于当前 team，则写入 `system` warning

对应实现位置：

- [internal/routing/router.go](E:/codes/icoo_assistant/icoo_gateway/internal/routing/router.go)

### 3.5 Agent 心跳

`agent_instance` 已支持 heartbeat：

- 可以写入 `last_heartbeat_at`
- 当实例状态为 `offline` 或 `created` 时，heartbeat 后可恢复为 `idle`

对应实现位置：

- [internal/agentinstance/service.go](E:/codes/icoo_assistant/icoo_gateway/internal/agentinstance/service.go)

### 3.6 基础测试

当前测试情况：

- `go test ./...` 可通过
- 测试主要覆盖 `internal/api` 和 `internal/config`
- 已覆盖资源创建、查询、消息追加、Team 基础路由、heartbeat 等核心骨架行为

但当前还没有数据库集成测试、runtime 契约测试和完整闭环冒烟测试。

## 4. 设计项对照分析

| 设计项 | 文档期望 | 当前实现 | 评估 |
| --- | --- | --- | --- |
| 服务骨架 | 独立入口、配置、HTTP 服务 | 已具备 | 已完成 |
| 模块划分 | `api/auth/skill/agentprofile/agentinstance/team/conversation/routing/audit/storage/config` | 当前仅有 `api/config/skill/agentprofile/agentinstance/team/conversation/routing/server` | 部分完成 |
| Skill API | `create/list/get/patch/activate/deactivate` | 只有 `create/list/get` | 部分完成 |
| AgentProfile API | `create/list/get/patch` | 只有 `create/list/get` | 部分完成 |
| AgentInstance API | `create/list/get/heartbeat/disable` | 已有 `heartbeat`，无 `disable` | 部分完成 |
| Team API | `create/list/get/patch/member add/update/delete` | 只有 team `create/list/get` 和 member `add/list` | 部分完成 |
| Conversation API | `create/list/get/messages/runs` | 已有 `create/list/get/messages`，无 `runs` | 部分完成 |
| Audit API | `audit-events` 查询接口 | 无对应实现 | 未实现 |
| 统一对话模型 | 单对话与 Team 对话共用主模型 | 已实现 | 已完成 |
| 消息分层 | `external/internal/system` | 已实现 | 已完成 |
| Team 基础路由 | 入口 Agent 路由、内部消息流、最终对外回复 | 目前只有分派和汇总占位 | 部分完成 |
| Agent 心跳 | 在线状态更新 | 已实现 | 已完成 |
| Runtime 集成 | 上下文下发、执行回传、状态聚合 | 只有 heartbeat | 未实现 |
| 持久化层 | PostgreSQL/Redis 或持久化抽象 | 仍是进程内内存存储 | 未实现 |
| 鉴权与权限 | `auth` 模块、权限控制 | 无 | 未实现 |
| 审计能力 | 审计模型、入库与查询 | 无 | 未实现 |
| 测试体系 | 单元、集成、契约、冒烟 | 当前偏 API 单测 | 部分完成 |

## 5. 数据模型完成度评估

## 5.1 Skill

设计文档建议字段包括：

- `id`
- `name`
- `version`
- `description`
- `source_type`
- `source_uri`
- `content_digest`
- `status`
- `is_default`
- `created_at`
- `updated_at`

当前实现仅包含：

- `id`
- `name`
- `version`
- `description`
- `status`
- `created_at`
- `updated_at`

结论：

- 基础资源标识已具备
- 版本治理、来源治理、默认标记等还未落地

## 5.2 AgentProfile

设计文档建议字段包括：

- `id`
- `name`
- `model_provider`
- `model_name`
- `system_prompt`
- `tool_policy`
- `default_skill_ids`
- `status`
- `created_at`
- `updated_at`

当前实现仅包含：

- `id`
- `name`
- `model_provider`
- `model_name`
- `system_prompt`
- `status`
- `created_at`
- `updated_at`

结论：

- 模型和提示词基础能力已具备
- `tool_policy` 与默认 Skill 绑定能力尚未建立

## 5.3 AgentInstance

设计文档建议字段包括：

- `id`
- `profile_id`
- `display_name`
- `runtime_type`
- `runtime_endpoint`
- `status`
- `last_heartbeat_at`
- `team_id`
- `metadata`
- `created_at`
- `updated_at`

当前实现已包含：

- `id`
- `profile_id`
- `display_name`
- `runtime_type`
- `runtime_endpoint`
- `status`
- `last_heartbeat_at`
- `created_at`
- `updated_at`

当前缺失：

- `team_id`
- `metadata`

结论：

- 运行实例基础模型已成型
- 实例与 team 的显式归属关系仍不完整

## 5.4 Team 与 TeamMember

`Team` 当前包含：

- `id`
- `name`
- `description`
- `entry_agent_id`
- `status`
- `created_at`
- `updated_at`

`Member` 当前包含：

- `id`
- `team_id`
- `agent_id`
- `role`
- `sort_order`
- `status`
- `responsibility`
- `created_at`
- `updated_at`

缺口：

- `Team.mode` 未体现
- 还缺成员更新、删除、停用等治理动作

结论：

- Team 与成员关系的基础建模已经有了
- 但仍偏静态资源管理，治理深度不足

## 5.5 Conversation 与 Message

`Conversation` 当前已包含：

- `id`
- `mode`
- `title`
- `target_agent_id`
- `target_team_id`
- `status`
- `message_count`
- `created_by`
- `created_at`
- `updated_at`

`Message` 当前已包含：

- `id`
- `conversation_id`
- `scope`
- `role`
- `sender_type`
- `sender_id`
- `receiver_type`
- `receiver_id`
- `content`
- `sequence_no`
- `created_at`

缺口：

- `Conversation.last_run_id` 未实现
- `Message.message_type` 未实现
- 还没有独立 Run 模型

结论：

- 会话主模型已经具备基本形态
- 但运行态管理仍未纳入统一资源模型

## 5.6 Run 与 AuditEvent

当前实现状态：

- `Run`：未实现
- `AuditEvent`：未实现

这两个对象缺失，意味着：

- 会话执行过程还没有正式运行记录
- 管理行为和系统行为还没有统一审计链路

## 6. 按 Phase 评估完成情况

### 6.1 Phase 1：资源管理底座

文档目标：

- 建立服务骨架
- 落地 Skill / Agent / Team 基础 CRUD
- 统一审计模型

当前评估：

- 服务骨架：已完成
- Skill / Agent / Team 基础 CRUD：部分完成
- 审计模型：未实现
- 数据库迁移：未实现

完成度判断：

- 约 `60% - 70%`

### 6.2 Phase 2：单对话管理

文档目标：

- 建立单对话模型
- 打通会话创建、续聊、历史查询
- 建立基础 Run 记录

当前评估：

- 单对话创建：已完成
- 消息追加：已完成
- 历史查询：已完成
- Run 记录：未实现

完成度判断：

- 约 `55% - 65%`

### 6.3 Phase 3：Team 对话基础路由

文档目标：

- 建立 Team 对话模式
- 支持入口 Agent 路由与基础内部消息流
- 返回最终对外回复

当前评估：

- Team 对话模式：已完成
- 内外消息分层：已完成
- 基础入口路由：已完成
- 最终对外回复闭环：未完成

完成度判断：

- 约 `45% - 55%`

### 6.4 Phase 4：Runtime 集成

文档目标：

- 打通网关服务与本地 Runtime 交互
- 支持 Agent 实例心跳和执行结果回传
- 支持运行状态聚合视图

当前评估：

- 心跳：已完成
- 执行结果回传：未实现
- 状态聚合视图：未实现

完成度判断：

- 约 `15% - 20%`

### 6.5 Phase 5：治理增强

文档目标：

- 统计视图
- 资源启停
- 常见失败补偿
- 基础权限控制

当前评估：

- 基本未开始

完成度判断：

- 约 `0% - 10%`

## 7. 当前最关键的缺口

## 7.1 持久化层完全未落地

这是当前最大缺口。

虽然管理面 API 已经有了，但当前存储仍是进程内内存结构，意味着：

- 服务重启即丢数据
- 无法承接真实控制面职责
- 无法建立审计、历史和恢复机制

这使当前实现仍停留在“原型骨架”阶段。

## 7.2 Routing 仍是占位，不是真正编排

当前 routing 的核心行为是：

- 写一条发往入口 Agent 的内部消息
- 写一条等待汇总的系统占位消息

但缺少：

- 入口 Agent 执行结果接入
- 成员间真实消息流转
- 最终外部回复生成
- 失败与重试逻辑

所以它更像“流程事件占位器”，而不是完整 orchestration 模块。

## 7.3 平台能力几乎未开始

主要包括：

- 鉴权
- 权限控制
- 审计入库与查询
- Runtime 回传
- 运行视图聚合

这些能力决定了系统是否能从“样例服务”升级成“可管理平台”。

## 8. 风险判断

如果直接在当前基础上继续堆业务接口，而不先补齐持久化与运行模型，可能出现以下问题：

- 控制面看起来功能越来越多，但状态不可靠
- Team 对话逻辑容易继续写成临时规则
- 后续接 Runtime 时接口边界容易再次混乱
- 审计与权限会变成后补硬插，返工成本高

## 9. 建议的下一步开发优先级

建议按以下顺序推进：

### 9.1 第一优先级：补齐持久化和存储抽象

优先落地：

- `storage` 模块
- PostgreSQL 数据模型
- 迁移脚本
- 资源读写从内存迁移到持久层

目标是先让 Skill / Agent / Team / Conversation 变成真正可保存、可恢复、可查询的控制面数据。

### 9.2 第二优先级：补齐 Run 和 Audit

优先落地：

- `Run` 模型
- `AuditEvent` 模型
- 对话执行记录
- 资源操作审计
- 审计查询 API

这一步会显著提升可观测性和后续问题定位能力。

### 9.3 第三优先级：补齐缺失的治理 API

优先补齐：

- `PATCH /skills/{id}`
- `activate / deactivate`
- `PATCH /agent-profiles/{id}`
- `POST /agent-instances/{id}/disable`
- `PATCH /teams/{id}`
- `PATCH /teams/{id}/members/{memberId}`
- `DELETE /teams/{id}/members/{memberId}`
- `GET /conversations/{id}/runs`

### 9.4 第四优先级：推进真正的 Runtime 集成

需要建立：

- Runtime 执行回传接口
- 上下文下发结构
- 最小执行结果回收链路
- 运行状态聚合视图

### 9.5 第五优先级：把 Team routing 从占位升级为闭环

在上述能力具备后，再推进：

- entry agent 执行
- 成员协作消息流
- 汇总回复
- 失败回传
- 最终对外消息落库

## 10. 最终结论

`icoo_gateway` 当前已经完成了正确方向上的第一步：

- 服务骨架存在
- 基础资源 API 存在
- 统一会话模型存在
- Team 对话最小路由存在
- Agent heartbeat 存在

但从设计文档要求来看，它目前仍然是“管理面原型骨架”，而不是完整的网关管理服务。

当前最合理的判断是：

- 它已经足够支撑后续继续开发
- 但还不足以作为完整控制面交付
- 下一阶段不应继续横向加零散功能，而应优先补齐持久化、运行模型、审计和 runtime 集成

只有这样，`icoo_gateway` 才能真正从“可演示”走向“可使用”。

## 11. 建议开发计划

建议后续开发不要再按“想到什么补什么”的方式推进，而是围绕“先站稳管理面，再接执行面”的思路分阶段实施。

整体建议分为四个迭代阶段：

- 第 1 阶段：持久化底座与资源治理补齐
- 第 2 阶段：运行模型与审计闭环
- 第 3 阶段：Runtime 集成与状态回收
- 第 4 阶段：Team 对话闭环与治理增强

## 11.1 第 1 阶段：持久化底座与资源治理补齐

### 目标

把当前“进程内内存骨架”升级为“可保存、可恢复、可治理”的管理面服务。

### 核心任务

- 新增 `storage` 模块，抽象资源读写接口
- 引入 PostgreSQL，完成基础表结构和迁移脚本
- 将 `skill / agent_profile / agent_instance / team / team_member / conversation / message` 从内存存储迁移到数据库
- 为现有资源补齐缺失字段
- 补齐缺失的资源治理 API

### 建议优先补齐的 API

- `PATCH /api/v1/skills/{id}`
- `POST /api/v1/skills/{id}/activate`
- `POST /api/v1/skills/{id}/deactivate`
- `PATCH /api/v1/agent-profiles/{id}`
- `POST /api/v1/agent-instances/{id}/disable`
- `PATCH /api/v1/teams/{id}`
- `PATCH /api/v1/teams/{id}/members/{memberId}`
- `DELETE /api/v1/teams/{id}/members/{memberId}`

### 验收标准

- 服务重启后资源数据不丢失
- Skill / Agent / Team / Conversation 都能稳定完成增删改查
- team member 的新增、更新、删除可用
- 关键资源状态支持启用、停用
- 数据库迁移可以在新环境一键执行

### 预计收益

- 网关从“演示骨架”升级为“可持续开发的真实底座”
- 后续审计、运行记录和 Web 管理端都会有稳定依托

## 11.2 第 2 阶段：运行模型与审计闭环

### 目标

建立正式的运行态资源模型，让系统具备“能记录发生过什么”的能力。

### 核心任务

- 新增 `Run` 模型与持久化
- 新增 `AuditEvent` 模型与持久化
- 为对话执行、资源创建、资源修改、状态切换写入审计事件
- 补齐运行查询接口和审计查询接口
- 为 conversation 建立 `last_run_id` 等运行态关联

### 建议新增的 API

- `GET /api/v1/conversations/{id}/runs`
- `GET /api/v1/audit-events`
- `GET /api/v1/audit-events/{id}`

### 验收标准

- 每次会话执行都能落一条 Run 记录
- 每次关键资源操作都能落审计事件
- 能按 conversation 查询运行历史
- 能按资源类型、资源 ID 查询审计记录

### 预计收益

- 问题定位能力明显增强
- 后续接 Runtime 后可以形成完整的执行追踪链路

## 11.3 第 3 阶段：Runtime 集成与状态回收

### 目标

打通网关控制面与执行面的最小闭环。

### 核心任务

- 设计并实现 Runtime 执行回传接口
- 设计上下文下发结构，包括对话上下文、目标 Agent 配置、Skill 列表、Team 上下文
- 建立执行状态更新机制
- 建立运行状态聚合视图
- 让 `agent_instance` 在线状态不只靠 heartbeat，还能反映执行状态

### 建议优先落地的接口方向

- Runtime 拉取或接收执行上下文
- Runtime 回传执行结果
- Runtime 回传失败信息
- Runtime 回传内部路由请求

### 验收标准

- 一次单对话执行结果能够从 Runtime 回收到 Conversation 和 Run
- Agent 在线状态与最近执行状态可查询
- 执行失败信息能够回写并被审计记录捕获

### 预计收益

- 网关开始真正承接控制面职责
- 本地 Runtime 与网关的边界会更清晰

## 11.4 第 4 阶段：Team 对话闭环与治理增强

### 目标

让 Team 对话从“分派占位”升级为“真正可闭环的团队协作流程”。

### 核心任务

- 将 routing 从占位事件升级为真实的 Team 对话编排
- 支持 entry agent 处理外部消息
- 支持成员间内部消息流转
- 支持汇总 Agent 或入口 Agent 产出最终对外回复
- 补齐失败补偿、资源启停、基础权限控制、统计视图

### 建议能力拆分

- Team 消息路由规则
- 内部消息投递与回收
- 最终对外回复落库
- Team 运行失败回写
- 常见错误重试或补偿
- 基础权限控制与调用方识别

### 验收标准

- Team 对话能形成“外部消息 -> 内部分派 -> 内部协作 -> 最终外部回复”的闭环
- 失败路径可见、可追踪、可定位
- 关键资源可以启停
- 基础权限控制可阻止未授权调用

### 预计收益

- `icoo_gateway` 从资源管理服务升级为真正的 Team 控制面
- 后续接 Web 管理端时，不需要重做核心 API

## 11.5 建议的排期策略

如果按稳定推进的节奏，建议采用下面的排期方式：

### Sprint 1

- 建立 `storage` 模块
- 接入 PostgreSQL
- 完成资源表结构与迁移
- 把 Skill / Agent / Team / Conversation 迁入数据库

### Sprint 2

- 补齐资源治理 API
- 补齐 `Run` 与 `AuditEvent`
- 建立运行历史与审计查询

### Sprint 3

- 设计并实现 Runtime 回传协议
- 打通单对话执行结果回收闭环
- 增强 `agent_instance` 状态模型

### Sprint 4

- 打通 Team 对话真实编排闭环
- 增加失败补偿、资源启停、基础权限控制
- 为 Web 管理端准备稳定 API

## 11.6 当前最推荐的执行原则

建议整个开发过程中坚持以下原则：

- 先持久化，后复杂编排
- 先资源治理，后自治能力
- 先把单对话运行链路走通，再扩展 Team 闭环
- 先把审计和运行记录建起来，再追求高级调度
- 不再把 Team 治理逻辑回塞到本地 Runtime

## 11.7 一句话建议

`icoo_gateway` 的下一步重点，不是继续补更多零散接口，而是先把“数据库 + Run + Audit + Runtime 回传”这四根主梁立起来；主梁立稳以后，Team routing 和管理能力才会越做越顺。

## 12. 建议版本路线图

为了便于项目管理，建议把后续开发计划进一步映射到版本路线图，而不是仅停留在功能清单层面。

建议采用“每个版本只解决一类核心问题”的方式推进，避免一个版本同时混入底座改造、协议设计、路由编排和权限治理，导致目标发散。

### 12.1 建议版本划分

| 建议版本 | 核心主题 | 主要目标 |
| --- | --- | --- |
| `v0.2.0` | 持久化底座版 | 建立数据库、迁移脚本、存储抽象，完成核心资源持久化 |
| `v0.3.0` | 运行审计版 | 建立 `Run` 与 `AuditEvent`，补齐运行历史和审计查询 |
| `v0.4.0` | Runtime 接入版 | 打通单对话执行回传闭环和状态聚合 |
| `v0.5.0` | Team 闭环版 | 打通 Team 对话协作闭环，形成最小可用编排能力 |
| `v0.6.0` | 治理增强版 | 增加权限控制、资源启停、统计视图和常见失败补偿 |

### 12.2 各版本建议交付内容

#### `v0.2.0` 持久化底座版

建议交付：

- PostgreSQL 接入
- `storage` 模块
- 核心资源表结构
- 数据库迁移脚本
- 内存存储替换为数据库存储
- 补齐基础资源治理 API

完成标志：

- 服务重启后数据仍在
- 核心资源可稳定增删改查
- 本地和新环境都可完成迁移初始化

#### `v0.3.0` 运行审计版

建议交付：

- `Run` 模型与查询接口
- `AuditEvent` 模型与查询接口
- 资源操作审计
- 会话执行历史查询

完成标志：

- 一次会话处理过程可被记录、查询、追踪
- 关键资源修改行为可审计

#### `v0.4.0` Runtime 接入版

建议交付：

- Runtime 执行回传接口
- 单对话执行结果回收闭环
- 失败状态回写
- Agent 在线与执行状态聚合

完成标志：

- 单对话从触发到执行结果回收可以完整跑通
- 执行失败信息能进入 Run 和审计记录

#### `v0.5.0` Team 闭环版

建议交付：

- entry agent 处理逻辑接入
- 内部消息链路打通
- 汇总结果回写外部消息
- Team 对话最小协作闭环

完成标志：

- Team 对话可以从外部输入走到最终外部回复
- 内部协作过程可见、可追踪

#### `v0.6.0` 治理增强版

建议交付：

- 基础权限控制
- 资源启停控制
- 常见失败补偿
- 统计视图
- 为 Web 管理端准备稳定核心接口

完成标志：

- 核心资源可治理、可停用、可审计
- Web 管理端接入不需要推翻现有 API 设计

## 13. 建议角色分工

如果后续由多人并行推进，建议按“模块职责”而不是“接口列表”分工，避免边界混乱。

### 13.1 建议角色划分

| 角色方向 | 主要职责 |
| --- | --- |
| 网关后端负责人 | 把控模块边界、数据模型、版本推进和技术决策 |
| 存储与数据负责人 | 负责 PostgreSQL、迁移脚本、存储抽象、数据一致性 |
| API 与资源治理负责人 | 负责 Skill / Agent / Team / Conversation 资源接口和请求校验 |
| 运行集成负责人 | 负责 Runtime 协议、执行回传、状态聚合 |
| 编排与 Team 路由负责人 | 负责 Team 对话流转、内部消息规则、汇总闭环 |
| 测试与质量负责人 | 负责集成测试、契约测试、冒烟测试和回归基线 |

### 13.2 如果只有 1 到 2 人开发

建议按下面方式合并职责：

- 一人负责“数据层 + API 层”
- 一人负责“Runtime 集成 + Team 编排”

如果只有 1 人开发，则建议严格按阶段推进，不要同时推进数据库、Runtime 协议、权限和 Team 编排。

## 14. 建议里程碑

为了让版本推进具备检查点，建议设置以下里程碑。

### M1：资源持久化里程碑

标志：

- 核心资源全部入库
- 内存存储退出主路径
- 迁移脚本稳定可复用

完成后说明：

- `icoo_gateway` 从“原型服务”进入“可持续开发底座”阶段

### M2：运行可追踪里程碑

标志：

- `Run` 和 `AuditEvent` 正式上线
- 会话执行和资源修改可查询

完成后说明：

- 系统开始具备基本可观测性

### M3：单对话执行闭环里程碑

标志：

- Runtime 执行回传打通
- 单对话能形成完整闭环

完成后说明：

- 控制面与执行面边界初步稳定

### M4：Team 最小闭环里程碑

标志：

- Team 对话能完成“外部输入 -> 内部协作 -> 外部回复”

完成后说明：

- Team 控制面进入最小可用状态

### M5：治理可接管里程碑

标志：

- 权限、启停、统计、补偿具备基本能力

完成后说明：

- Web 管理端接入条件基本成熟

## 15. 建议时间线

如果按较稳妥节奏推进，可以参考下面的时间线模型。

### 方案 A：单人推进

| 时间段 | 建议目标 |
| --- | --- |
| 第 1 周到第 2 周 | 完成 `v0.2.0` 持久化底座版 |
| 第 3 周 | 完成 `v0.3.0` 运行审计版 |
| 第 4 周到第 5 周 | 完成 `v0.4.0` Runtime 接入版 |
| 第 6 周到第 7 周 | 完成 `v0.5.0` Team 闭环版 |
| 第 8 周 | 完成 `v0.6.0` 治理增强版基础能力 |

### 方案 B：双人并行

| 时间段 | 建议目标 |
| --- | --- |
| 第 1 周 | 并行推进数据库底座设计与 API 补齐设计 |
| 第 2 周到第 3 周 | 完成 `v0.2.0` 和 `v0.3.0` 主体 |
| 第 4 周到第 5 周 | 打通 `v0.4.0` Runtime 回传闭环 |
| 第 6 周 | 完成 `v0.5.0` Team 闭环主链路 |
| 第 7 周 | 完成 `v0.6.0` 的权限、启停、补偿基础能力 |

说明：

- 如果 Runtime 协议设计存在反复，整体节奏通常会在 `v0.4.0` 和 `v0.5.0` 阶段放慢
- 如果持久化阶段没有先把资源模型理顺，后续所有阶段都会反复返工

## 16. 主要风险与依赖

建议在排期时提前把风险和依赖写进版本计划，避免到了联调阶段才暴露问题。

### 16.1 主要风险

| 风险项 | 风险说明 | 建议应对 |
| --- | --- | --- |
| 数据模型过早定死 | 当前模型还比较骨架化，若不留扩展位，后续 Runtime 接入时可能返工 | 在 `v0.2.0` 阶段优先补齐关键字段，预留扩展字段 |
| Runtime 协议边界不清 | 控制面和执行面责任容易再次混淆 | 在 `v0.4.0` 前先固定请求/回传边界 |
| Team 编排过早复杂化 | 很容易为了“更智能”而提前引入复杂自治逻辑 | `v0.5.0` 只做最小可用闭环，不做复杂自治 |
| 缺少审计导致定位困难 | Runtime 接入后没有 Run/Audit 会很难排错 | 在 `v0.3.0` 前不要跳过运行与审计模型 |
| 单元测试通过但系统不可用 | 当前测试以 API 层为主，无法覆盖集成问题 | 补齐集成测试、契约测试和冒烟测试 |

### 16.2 关键依赖

后续推进依赖以下前置条件：

- 明确 PostgreSQL 作为主存储
- 明确 Runtime 与网关之间的最小协议边界
- 明确 AgentInstance 与 Team 的关系建模方式
- 明确审计记录保留粒度
- 明确是否需要在当前阶段就预留 Web 管理端使用场景

## 17. 推荐执行方式

如果要尽量减少返工，我最推荐的实际执行方式是：

1. 先完成 `v0.2.0`，把数据层站稳。
2. 再完成 `v0.3.0`，把运行记录和审计链路立起来。
3. 然后推进 `v0.4.0`，只打通单对话 Runtime 闭环。
4. 最后再进入 `v0.5.0`，把 Team 编排做成最小可用闭环。

这样做的好处是：

- 每一阶段都能形成可验证的成果
- 不会因为 Team 编排过早展开而拖乱底层设计
- 能最大程度保持控制面与执行面的边界稳定

## 18. 补充结论

如果把当前 `icoo_gateway` 看作一个项目阶段成果，它已经顺利完成了“从 0 到 1 搭出方向正确的原型”。

接下来的重点，不再是证明“这个服务能不能存在”，而是推进它从：

- 内存骨架
- 走向持久化底座
- 从占位 routing
- 走向真实运行闭环
- 从资源 CRUD
- 走向真正的控制面治理能力

这也是后续开发计划需要严格分阶段推进的根本原因。
