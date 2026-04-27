# icoo_proxy 项目功能分析报告

## 1. 项目定位

`icoo_proxy` 是一个基于 **Go + Wails v2 + Vue 3** 的本地桌面应用，用来提供一个运行在本机的 AI 协议转换网关，并附带可视化管理界面。

从当前实现看，它的核心职责包括：

- 启动本地 HTTP 代理服务
- 对外暴露 Anthropic、OpenAI Chat Completions、OpenAI Responses 三类入口
- 按模型与协议进行路由和转发
- 支持部分跨协议请求/响应转换
- 提供供应商、端点、鉴权 Key、路由策略、流量、项目设置等管理能力

关键依据：

- `README.md`
- `main.go:16`
- `app.go:49`
- `internal/proxy/service.go:43`

---

## 2. 技术架构概览

### 2.1 整体结构

项目可分为三层：

1. **桌面壳层**：Wails 承载桌面窗口、启动 Go 后端并向前端暴露绑定方法
2. **后端网关层**：Go HTTP 服务负责配置加载、路由、鉴权、协议转换、请求转发与状态输出
3. **前端管理台层**：Vue 3 + Pinia + Vue Router 实现管理界面

关键入口：

- `main.go:16`：Wails 应用启动入口
- `app.go:49`：应用启动时初始化各服务并自动拉起本地代理
- `internal/api/router.go:90`：注册状态接口、管理接口和代理入口
- `internal/proxy/service.go:31`、`internal/proxy/service.go:43`：代理服务创建与请求处理入口

### 2.2 后端模块划分

- `app.go`
  - 应用生命周期管理
  - 聚合供应商、端点、授权 Key、路由策略、网关状态
  - 提供给前端调用的 Wails 绑定方法
- `internal/config/config.go`
  - 负责加载 `.env` 配置与默认值
- `internal/storage/db.go`
  - 负责打开 SQLite 数据库 `.data/icoo_proxy.db`
- `internal/supplier/service.go`
  - 供应商配置 CRUD 与默认供应商种子数据
- `internal/endpoint/service.go`
  - 端点配置 CRUD 与内置端点初始化
- `internal/authkey/service.go`
  - 代理访问鉴权 Key 管理
- `internal/projectsettings/service.go`
  - 项目设置读写、校验与写回 `.env`
- `internal/api/router.go`
  - 暴露健康检查、管理接口和代理路由
- `internal/proxy/service.go`
  - 请求鉴权、模型解析、路由决策、请求转换、上游转发、响应转换、近期请求记录

### 2.3 前端模块划分

- `frontend/src/App.vue:51`
  - 应用外壳与侧边导航
- `frontend/src/router/index.js:13`
  - 页面路由定义
- `frontend/src/views/*.vue`
  - 各业务页面
- `frontend/src/stores/*.js`
  - Pinia 状态管理
- `frontend/src/components/ued/*`
  - 基础 UI 组件库

前端页面路由已明确包含：

- 概览：`frontend/src/router/index.js:16`
- 供应商管理：`frontend/src/router/index.js:21`
- 端点管理：`frontend/src/router/index.js:26`
- 授权 Key：`frontend/src/router/index.js:31`
- 流量监控：`frontend/src/router/index.js:36`
- 项目设置：`frontend/src/router/index.js:41`
- UED 规范：`frontend/src/router/index.js:46`

---

## 3. 当前已实现的核心功能

## 3.1 本地代理网关自动启动

应用启动时会初始化供应商、路由策略、端点、授权 Key 等服务，并调用 `startProxy()` 自动拉起本地 HTTP 服务。

关键依据：

- `app.go:49`
- `app.go:127`
- `app.go:270`

该能力说明本项目不是单纯的配置管理台，而是“桌面管理台 + 本地运行网关”的组合体。

## 3.2 对外 API 入口管理

系统当前支持固定的协议入口，并允许从数据库读取启用中的端点作为实际可用路由。

默认端点包括：

- `/v1/messages`
- `/anthropic/v1/messages`
- `/v1/chat/completions`
- `/openai/v1/chat/completions`
- `/v1/responses`
- `/openai/v1/responses`

关键依据：

- `app.go:480`
- `internal/endpoint/service.go:170`
- `internal/api/router.go:90`

其中：

- `internal/endpoint/service.go:98` 负责端点保存/更新
- `internal/endpoint/service.go:140` 限制内置端点不可删除
- `internal/endpoint/service.go:86` 返回启用中的端点供运行时注册

这意味着项目支持“内置标准入口 + 可配置本地入口”的模式。

## 3.3 协议路由与转发

代理服务以 `internal/proxy/service.go:43` 为核心入口，处理流程包括：

- 限制只接受 `POST`
- 校验访问鉴权
- 读取请求体
- 提取请求模型
- 解析下游协议与目标上游路由
- 按需要改写请求体
- 向上游发起请求
- 回写或转换响应结果
- 记录最近请求摘要

关键依据：

- `internal/proxy/service.go:43`

从 `README.md` 与状态说明可确认当前已实现：

- 同协议透传
- 模型别名重写
- 默认路由
- 非流式 `chat/completions <-> responses` 转换
- 非流式 `anthropic messages <-> responses` 转换
- 非流式 `anthropic messages <-> chat/completions` 转换
- 基础 tool/function/tool_use 结构映射
- 部分流式处理能力

补充说明：

- `app.go:270` 的状态说明中明确写出了已支持的转换能力
- `README.md` 也对当前功能范围与未完成项做了描述

## 3.4 状态接口与管理接口

当前后端已实现以下管理/状态接口：

- `/healthz`
- `/readyz`
- `/admin/models`
- `/admin/routes`
- `/admin/requests`

关键依据：

- `internal/api/router.go:90`

这些接口分别覆盖：

- 基础存活检查
- 代理运行就绪检查
- 默认路由/模型别名查看
- 支持路径与说明查看
- 最近请求摘要查看

这也支撑了前端“概览页”和“流量监控页”的数据来源。

## 3.5 供应商管理

供应商是上游连接配置的核心实体，字段包括：

- 名称
- 协议类型
- Base URL
- API Key
- only_stream
- User-Agent
- 启用状态
- 描述
- 模型列表
- 标签

关键依据：

- `internal/supplier/service.go:69`
- `internal/supplier/service.go:92`
- `internal/supplier/service.go:121`
- `internal/supplier/service.go:173`
- `internal/supplier/service.go:267`

可确认已具备的能力：

- 供应商列表展示
- 新建供应商
- 编辑供应商
- 删除供应商
- 默认供应商种子初始化

前端界面能力也很完整：

- 新建供应商按钮：`frontend/src/views/SuppliersView.vue:5`
- 健康检查触发：`frontend/src/views/SuppliersView.vue:96`
- 供应商提交保存：`frontend/src/views/SuppliersView.vue:467`

从默认种子数据看，当前默认预置了：

- Anthropic Default
- OpenAI Responses Default

说明系统当前重点围绕 Anthropic 与 OpenAI Responses 作为主要上游。

## 3.6 默认路由策略管理

路由策略页面并入“供应商管理”中，允许用户指定：

- 下游协议
- 供应商
- 上游协议
- 目标模型
- 启用状态

前端关键依据：

- `frontend/src/views/SuppliersView.vue:6`
- `frontend/src/views/SuppliersView.vue:118`
- `frontend/src/views/SuppliersView.vue:474`
- `frontend/src/views/SuppliersView.vue:495`

后端相关聚合出口：

- `app.go:270`

这说明系统当前已经形成“下游协议 -> 指定供应商 -> 指定模型”的默认路由策略能力。

## 3.7 授权 Key 管理

本地代理支持独立于 `.env` 的访问 Key 管理，Key 数据持久化在数据库中。

关键依据：

- `internal/authkey/service.go:52`
- `internal/authkey/service.go:71`
- `internal/authkey/service.go:83`
- `internal/authkey/service.go:97`
- `internal/authkey/service.go:145`
- `internal/authkey/service.go:160`
- `internal/authkey/service.go:218`

已实现能力包括：

- 列出鉴权 Key
- 新增/编辑 Key
- 自动生成随机 Secret
- 删除 Key
- 读取完整 Secret
- 仅合并启用状态的 Secret 到运行时配置

并且在启动代理时，会把 `.env` 中的 Key 与数据库里启用的 Key 一起合并到运行时：

- `app.go` 中 `startProxy()` 的配置合并逻辑
- `internal/config/config.go:71`

这说明系统的鉴权机制支持静态配置与管理台配置并存。

## 3.8 项目设置管理

项目设置页面当前主要维护代理监听与日志相关参数，并且保存后会触发代理重载。

前端关键依据：

- `frontend/src/views/SettingsView.vue:21`
- `frontend/src/views/SettingsView.vue:44`
- `frontend/src/views/SettingsView.vue:64`
- `frontend/src/views/SettingsView.vue:96`

后端关键依据：

- `internal/projectsettings/service.go:22`
- `internal/projectsettings/service.go:39`
- `internal/projectsettings/service.go:52`
- `internal/projectsettings/service.go:74`
- `app.go:92`
- `app.go:127`

当前可配置项包括：

- `PROXY_HOST`
- `PROXY_PORT`
- `PROXY_READ_TIMEOUT_SECONDS`
- `PROXY_WRITE_TIMEOUT_SECONDS`
- `PROXY_SHUTDOWN_TIMEOUT_SECONDS`
- `PROXY_CHAIN_LOG_PATH`
- `PROXY_CHAIN_LOG_BODIES`
- `PROXY_CHAIN_LOG_MAX_BODY_BYTES`

说明该页面目前主要负责“网关运行参数”和“链路日志参数”的维护。

## 3.9 概览页

概览页已经具备较完整的运行态可视化能力，包括：

- 代理地址
- 监听地址
- 访问模式
- 版本
- 上游就绪状态
- 支持的接口路径
- 供应商健康汇总
- 默认路由
- 默认路由策略
- 手动重载代理

关键依据：

- `frontend/src/views/OverviewView.vue:12`
- `frontend/src/views/OverviewView.vue:30`
- `frontend/src/views/OverviewView.vue:44`
- `frontend/src/views/OverviewView.vue:65`
- `frontend/src/views/OverviewView.vue:100`
- `frontend/src/views/OverviewView.vue:145`

这说明项目并不只是“配置录入”，而是已经实现了网关运行信息的集中展示。

## 3.10 流量监控

流量监控页面展示最近请求摘要，并支持：

- 手动刷新
- 每 6 秒自动刷新
- 按协议筛选
- 统计请求数、成功数、错误数、平均耗时
- 展示请求 ID、路由、模型、状态码、耗时、错误信息

关键依据：

- `frontend/src/views/TrafficView.vue:16`
- `frontend/src/views/TrafficView.vue:26`
- `frontend/src/views/TrafficView.vue:33`
- `frontend/src/views/TrafficView.vue:55`
- `frontend/src/views/TrafficView.vue:135`
- `frontend/src/views/TrafficView.vue:136`

该能力依赖后端的 recent requests 记录机制，数据来源为：

- `internal/api/router.go:120`
- `app.go:270`

## 3.11 UED 规范页面

系统包含一个 `UED 规范` 页面，用于承载内部 UI 组件规范展示或设计系统演示。

关键依据：

- `frontend/src/router/index.js:46`
- `frontend/src/App.vue:78`

虽然它不是核心业务功能，但说明项目已经开始沉淀统一组件设计体系。

---

## 4. 运行时行为分析

## 4.1 启动与关闭

程序启动时：

- 获取当前工作目录
- 初始化供应商、端点、路由策略、授权 Key 服务
- 加载配置
- 应用路由策略
- 构建 catalog
- 创建代理服务
- 打开链路日志文件
- 注册 API 路由
- 启动 HTTP Server

关键依据：

- `app.go:49`
- `internal/config/config.go:40`
- `internal/storage/db.go:11`
- `internal/api/router.go:90`

程序关闭时会关闭代理与数据库连接：

- `app.go` 中 `shutdown()`

## 4.2 配置来源

当前配置来源分为两类：

1. `.env` 文件
2. SQLite 持久化数据

其中：

- `.env` 管理运行时基础参数：`internal/config/config.go:40`
- SQLite 管理业务配置：`internal/storage/db.go:11`

持久化数据主要包括：

- suppliers
- endpoints
- auth_keys
- 路由策略相关表

## 4.3 鉴权逻辑

请求进入代理时，会先进行授权检查；系统运行态是否要求鉴权，取决于当前合并后的 Key 集合是否为空。

关键依据：

- `internal/proxy/service.go:43`
- `internal/config/config.go:71`
- `app.go:270`

系统还支持 `AllowUnauthenticatedLocal` 本地免鉴权模式状态展示，说明项目考虑了本机开发或受信环境下的访问便利性。

## 4.4 日志能力

链路日志是当前系统较明确的一项基础能力，支持：

- 配置日志路径
- 配置是否记录请求/响应体
- 配置日志体大小限制

关键依据：

- `internal/projectsettings/service.go:74`
- `internal/config/config.go:40`
- `app.go` 中 `openChainLog()`

这说明当前系统已经具备基础观测能力，但还不是完整审计平台。

---

## 5. 前端页面功能清单

根据当前路由与页面实现，功能清单如下：

| 页面 | 功能概要 | 关键位置 |
| --- | --- | --- |
| 网关概览 | 查看代理运行状态、上游就绪状态、支持路径、默认路由、路由策略，并可重载代理 | `frontend/src/views/OverviewView.vue:12` |
| 供应商管理 | 管理供应商配置、执行健康检查、配置默认路由策略 | `frontend/src/views/SuppliersView.vue:5` |
| 端点管理 | 管理本地暴露端点路径与启停状态 | `frontend/src/router/index.js:26` |
| 授权 Key | 管理访问本地代理所需的 Key | `frontend/src/router/index.js:31` |
| 流量监控 | 查看最近请求摘要与统计，支持自动刷新 | `frontend/src/views/TrafficView.vue:16` |
| 项目设置 | 维护监听与日志相关配置，保存后自动重载 | `frontend/src/views/SettingsView.vue:21` |
| UED 规范 | 查看内部组件规范/设计系统演示 | `frontend/src/router/index.js:46` |

---

## 6. 当前项目的成熟度判断

从代码实现来看，当前项目已经不是单纯原型，而是一个具备可用闭环的初版产品，主要体现在：

- 有可运行的桌面壳与本地服务
- 有基础配置持久化能力
- 有真实的代理请求处理链路
- 有前端管理台覆盖核心管理面
- 有状态、流量、日志等基本可观测能力

但从 `README.md` 与代码状态说明也能看出，它仍处于“首版可用、功能持续补全”的阶段，主要限制在：

- 跨协议流式转换仍不完整
- 工具调用映射仍是基础覆盖
- 审计与日志界面仍较轻量
- 路由可视化配置能力还可继续增强

---

## 7. 当前最突出的产品能力总结

如果从产品角度总结，当前项目最核心的功能价值是：

1. **把多种 AI 协议入口统一收口到本地网关**
2. **允许用户通过桌面管理台维护供应商、模型路由和鉴权**
3. **支持 Anthropic / OpenAI Chat / OpenAI Responses 之间的部分协议转换**
4. **提供基本运行态、健康态与近期流量可视化能力**

因此，这个项目当前最适合被理解为：

> 一个面向本地桌面场景的 AI 协议聚合与转换管理网关。

---

## 8. 建议的后续分析方向

如果后续还要继续深入，可以继续拆成以下几份文档：

- 后端模块逐文件说明
- 前端页面与 Pinia store 对照表
- 协议转换矩阵说明
- 数据库存储模型说明
- 请求生命周期时序图

---

## 9. 结论

当前 `icoo_proxy` 已经完成了从“桌面应用壳”到“本地 AI 网关管理台”的主流程建设。它的重点不在单纯 UI，而在于把：

- 上游模型服务配置
- 协议入口暴露
- 默认路由策略
- 访问鉴权
- 运行状态与流量

整合进一个本地桌面可操作系统中。

从代码现状判断，项目当前的核心定位清晰，主路径已经打通，后续工作更偏向协议兼容补全、观测增强与管理体验完善。
