# icoo_proxy

`icoo_proxy` 是一个基于 `Go + Wails v2` 的本地 AI 协议转换网关桌面应用。

当前首版已经开始落地的方向是：

- 使用 Wails v2 提供桌面管理壳
- 在应用启动时自动拉起本地 HTTP 代理
- 对外暴露 Anthropic、OpenAI `chat/completions`、OpenAI `responses` 三套入口
- 提供模型别名、默认路由、基础鉴权和状态查看能力

当前版本已经具备：

- Wails v2 桌面骨架
- 代理配置加载
- 本地网关启动与关闭
- `GET /healthz`
- `GET /readyz`
- `GET /admin/models`
- `GET /admin/routes`
- `GET /admin/requests`
- `POST /v1/messages`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- 同协议模型别名转发
- 非流式 `chat/completions <-> responses` 跨协议转换
- 非流式 `anthropic messages <-> responses` 跨协议转换
- 非流式 `anthropic messages <-> chat/completions` 跨协议转换
- 基础 function tools 定义映射
- 非流式 tool call / tool result 基础映射
- 桌面页状态概览与代理重载

当前版本尚未完成：

- 更完整的流式事件兼容
- 更细的请求日志和审计界面
- 路由可视化编辑

## 快速开始

```bash
cd icoo_proxy
copy .env.example .env
go test ./...
wails dev
```

如果你只想验证后端编译，也可以执行：

```bash
go test ./...
```

## 构建

Windows 下可以直接执行根目录构建脚本：

```powershell
.\build.ps1
```

默认流程会：

- 在缺少 `frontend/node_modules` 时安装前端依赖
- 执行 `go test ./...`
- 执行 `npm run build`
- 生成 `build/bin/icoo_proxy.exe`

如果还想同时执行桌面端正式构建，可以加上：

```powershell
.\build.ps1 -WailsBuild
```

## 配置

通过项目目录下的 `.env` 进行配置，当前重点变量如下：

- `PROXY_HOST` / `PROXY_PORT`

上游供应商、默认路由和模型路由通过桌面端“供应商管理”维护；下游访问授权通过“授权 Key”维护，不再写入项目 `.env`。

## 当前路由说明

当前同时支持官方路径和带命名前缀的路径：

```text
POST /v1/messages
POST /anthropic/v1/messages
POST /v1/chat/completions
POST /openai/v1/chat/completions
POST /v1/responses
POST /openai/v1/responses
```

当前已实现：

- 同协议透传
- 默认模型路由
- 模型别名重写
- 非流式 `chat/completions -> responses`
- 非流式 `responses -> chat/completions`
- 非流式 `anthropic messages -> responses`
- 非流式 `responses -> anthropic messages`
- 非流式 `anthropic messages -> chat/completions`
- 非流式 `chat/completions -> anthropic messages`
- `chat tools <-> responses tools`
- `anthropic tools <-> responses tools`
- `chat tools <-> anthropic tools`
- 非流式 tool call / function_call / tool_use 基础互转
- OpenAI 与 Anthropic 风格错误结构
- 最近请求摘要

当前未实现：

- 跨协议流式转换
- 更完整的工具调用覆盖，例如并行工具、严格 schema 和复杂 content part
