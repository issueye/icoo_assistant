# icoo_gateway

`icoo_gateway` 是 `icoo_assistant` 后续管理控制面的独立 Go 项目骨架。

当前版本已提供：

- 独立 Go 模块
- 基础配置加载
- HTTP 服务启动
- `/healthz` 健康检查接口
- `skill` 资源的内存版 `create / list / get / patch / activate / deactivate` API
- `agent_profile` 资源的内存版 `create / list / get / patch` API
- `team` 资源的内存版 `create / list / get / patch` API
- `agent_instance` 资源的内存版 `create / list / get / heartbeat / disable` API
- `conversation` 单对话资源与消息追加 API；在 `sqlite` 模式下已支持 conversation/message 持久化
- `Run` 最小运行历史骨架，支持按 conversation 查询运行记录；在 `sqlite` 模式下已支持持久化
- `team conversation` 骨架：支持 `mode=team`、`internal/external/system` 消息 scope 与按 scope 查询
- 最小 `routing` 骨架：team 外部消息会自动生成入口分派和汇总占位事件
- `team` 与 `team member` 关系骨架：支持成员写入、更新、删除与查询；在 `sqlite` 模式下已支持持久化
- `audit-events` 最小查询 API，可查看关键资源变更的审计记录；在 `sqlite` 模式下已支持持久化
- `storage` 泛型仓储接口骨架，作为后续 PostgreSQL 持久化替换的边界准备
- `bootstrap` 内存依赖提供器，已把 `App` 初始化与具体内存实现解耦
- `storage/sqlite` 纯 Go + GORM provider，已打通 SQLite 启动与审计表自动迁移
- `storage/postgres` 占位提供器与数据库配置项，已打通未来接 PostgreSQL 的启动入口
- 基础单元测试

## 快速开始

```bash
cd icoo_gateway
cp .env.example .env
go test ./...
go run ./cmd/icoo_gateway
```

默认监听地址：

```text
127.0.0.1:18080
```

如果要启用 SQLite：

```text
GATEWAY_STORAGE_DRIVER=sqlite
GATEWAY_SQLITE_PATH=./data/icoo_gateway.db
```

当前 SQLite 模式下，`audit-events` 已使用纯 Go SQLite + GORM 持久化，其余资源仍为内存实现。

## 当前 API

```text
GET  /healthz
GET  /api/v1/skills
POST /api/v1/skills
GET  /api/v1/skills/{id}
PATCH /api/v1/skills/{id}
POST /api/v1/skills/{id}/activate
POST /api/v1/skills/{id}/deactivate
GET  /api/v1/agent-profiles
POST /api/v1/agent-profiles
GET  /api/v1/agent-profiles/{id}
PATCH /api/v1/agent-profiles/{id}
GET  /api/v1/agent-instances
POST /api/v1/agent-instances
GET  /api/v1/agent-instances/{id}
POST /api/v1/agent-instances/{id}/heartbeat
POST /api/v1/agent-instances/{id}/disable
GET  /api/v1/teams
POST /api/v1/teams
GET  /api/v1/teams/{id}
PATCH /api/v1/teams/{id}
GET  /api/v1/teams/{id}/members
POST /api/v1/teams/{id}/members
PATCH /api/v1/teams/{id}/members/{memberId}
DELETE /api/v1/teams/{id}/members/{memberId}
GET  /api/v1/conversations
POST /api/v1/conversations
GET  /api/v1/conversations/{id}
GET  /api/v1/conversations/{id}/messages
POST /api/v1/conversations/{id}/messages
GET  /api/v1/conversations/{id}/runs
GET  /api/v1/audit-events
GET  /api/v1/audit-events/{id}
```

消息模型当前支持：

- `scope=external`：用户与系统主消息流
- `scope=internal`：team 内部消息流
- `scope=system`：系统说明或路由事件占位

可以通过 `GET /api/v1/conversations/{id}/messages?scope=internal` 这类方式按 scope 查询。

当前最小 routing 行为：

- team 对话收到 `external` 消息后，会自动为 `entry_agent_id` 生成一条 `internal` 分派消息
- 同时追加一条 `system` 汇总占位消息
- 如果目标 team 没有配置 `entry_agent_id`，则只写入一条 `system` warning
- 如果 `entry_agent_id` 不在当前 team members 中，也只写入一条 `system` warning

`agent_instance` 当前已支持 heartbeat：

- 调用 `POST /api/v1/agent-instances/{id}/heartbeat` 会写入 `last_heartbeat_at`
- 如果实例原来是 `offline` 或 `created`，heartbeat 后会回到 `idle`

当前最小审计能力：

- 关键资源创建、更新、启停和部分会话写入动作会记录 `audit-events`
- 可以通过 `GET /api/v1/audit-events` 查询审计列表
- 可以通过 `GET /api/v1/audit-events/{id}` 查询单条审计事件
- 当 `GATEWAY_STORAGE_DRIVER=sqlite` 时，审计事件会持久化到 SQLite

当前最小运行历史能力：

- 对话收到 `external` 消息后，会生成一条最小 `run` 记录
- `conversation.last_run_id` 会更新到最近一次运行
- 可以通过 `GET /api/v1/conversations/{id}/runs` 查询运行历史
- 当 `GATEWAY_STORAGE_DRIVER=sqlite` 时，运行记录会持久化到 SQLite

当前会话持久化能力：

- 当 `GATEWAY_STORAGE_DRIVER=sqlite` 时，`conversations` 与 `conversation_messages` 会持久化到 SQLite
- 这使 `conversation -> message -> run` 最小链路在 SQLite 模式下已具备落库能力

当前 Team 持久化能力：

- 当 `GATEWAY_STORAGE_DRIVER=sqlite` 时，`teams` 与 `team_members` 会持久化到 SQLite
- Team 成员活跃状态仍会影响 routing 对 `entry_agent_id` 的校验结果

当前存储抽象状态：

- 已新增 `internal/storage` 泛型仓储接口
- 当前各领域服务仍然使用内存实现，但已开始对齐统一仓储边界
- 已新增 `internal/bootstrap` 依赖提供器，`api.NewApp()` 默认走内存依赖装配
- 已新增 `GATEWAY_STORAGE_DRIVER`、`GATEWAY_SQLITE_PATH` 与 `GATEWAY_DATABASE_URL` 配置项
- `sqlite` 模式当前已通过 GORM 接入纯 Go SQLite，并持久化 `audit-events`、`runs`、`conversations`、`conversation_messages`、`teams`、`team_members`
- 当 `GATEWAY_STORAGE_DRIVER=postgres` 时，当前会进入 PostgreSQL provider 骨架并明确提示“尚未实现”
- 后续接 PostgreSQL 时，可以按仓储接口逐步替换，而不必先重写 API 层

当前存储仍然是进程内内存存储，适合作为网关 API 第一阶段骨架，后续再替换成持久化层与真正的 routing/orchestration 模块。
