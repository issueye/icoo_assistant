# icoo_gateway

`icoo_gateway` 是 `icoo_assistant` 后续管理控制面的独立 Go 项目骨架。

当前版本已提供：

- 独立 Go 模块
- 基础配置加载
- HTTP 服务启动
- `/healthz` 健康检查接口
- `skill`、`agent_profile`、`team` 资源的内存版 `create / list / get` API
- `agent_instance` 资源的内存版 `create / list / get / heartbeat` API
- `conversation` 单对话资源与消息追加 API
- `team conversation` 骨架：支持 `mode=team`、`internal/external/system` 消息 scope 与按 scope 查询
- 最小 `routing` 骨架：team 外部消息会自动生成入口分派和汇总占位事件
- `team member` 关系骨架：支持成员写入与查询，成员必须绑定已注册 `agent_instance`，routing 会校验 `entry_agent_id` 是否属于 team
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

## 当前 API

```text
GET  /healthz
GET  /api/v1/skills
POST /api/v1/skills
GET  /api/v1/skills/{id}
GET  /api/v1/agent-profiles
POST /api/v1/agent-profiles
GET  /api/v1/agent-profiles/{id}
GET  /api/v1/agent-instances
POST /api/v1/agent-instances
GET  /api/v1/agent-instances/{id}
POST /api/v1/agent-instances/{id}/heartbeat
GET  /api/v1/teams
POST /api/v1/teams
GET  /api/v1/teams/{id}
GET  /api/v1/teams/{id}/members
POST /api/v1/teams/{id}/members
GET  /api/v1/conversations
POST /api/v1/conversations
GET  /api/v1/conversations/{id}
GET  /api/v1/conversations/{id}/messages
POST /api/v1/conversations/{id}/messages
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

当前存储仍然是进程内内存存储，适合作为网关 API 第一阶段骨架，后续再替换成持久化层与真正的 routing/orchestration 模块。
