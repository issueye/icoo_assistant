# icoo_proxy 优化开发计划

## 1. 计划目标

本计划基于当前项目分析结果与代码审查结论制定，目标是把 `icoo_proxy` 从“功能可用的初版桌面网关”推进到“更安全、更稳定、更易维护的可持续迭代版本”。

本轮优化重点聚焦四个方面：

1. **安全性修复**：先补齐当前已经确认的鉴权边界问题
2. **配置与数据一致性修复**：避免保存配置、删除数据后出现系统状态不一致
3. **路由与供应商模型重构**：解决 OpenAI Chat / Responses 共享同一上游配置的结构缺陷
4. **可维护性与验证体系增强**：补齐测试、错误反馈、运行态校验能力

---

## 2. 优化原则

在本计划执行过程中，建议遵循以下原则：

### 2.1 先修确定性问题，再做增强

优先处理已经确认的 BUG 和设计缺陷，不先做大规模重构或 UI 美化。

### 2.2 先保证运行正确，再提升配置体验

优先保证：

- 代理启动/重载逻辑正确
- 删除/保存行为不会破坏运行态
- 路由与鉴权结果可预测

### 2.3 后端能力先收敛，再调整前端交互

当前部分问题本质在后端模型设计，不适合只通过前端限制来规避。前端应服务于后端真实能力，而不是掩盖后端缺陷。

### 2.4 每个阶段都要有可验证产出

每一阶段都应配套：

- 单元测试或集成测试
- 手工验证场景
- 明确的回归检查点

### 2.5 优先消除重复参数与重叠模块

当前项目除了已确认的 BUG 外，还存在一类会持续放大维护成本的问题：**同一能力被多个参数表达、同一职责被多个模块重复承担**。这类问题短期内不一定直接报错，但会导致：

- 配置来源不清晰
- 修改一处需要同步多处
- 前后端语义逐渐漂移
- 后续重构成本快速升高

因此，本轮优化中应坚持：

- 一个领域只保留一个主配置入口
- 一个运行时能力只保留一个主实现来源
- 兼容旧参数时要有明确退役计划
- 重复定义先收敛，再扩展新功能

---

## 3. 优化阶段划分

建议将本轮优化拆分为 **四个阶段**。

---

## 4. 第一阶段：高优先级缺陷修复与参数去重

### 4.1 阶段目标

解决已经确认的高优先级问题，防止：

- 安全边界失效
- 保存配置导致数据丢失
- 删除数据导致系统重载失败
- 多供应商配置失真
- 重复参数继续并存造成配置歧义

### 4.2 重复参数与模块重叠分析

在当前实现中，可以确认以下“参数重复”和“模块能力重复”问题已经值得纳入本轮优化。

#### A. 鉴权参数重复：`PROXY_API_KEY` 与 `PROXY_API_KEYS` 并存

**位置**

- `internal/config/config.go:19`
- `internal/config/config.go:20`
- `internal/config/config.go:50`
- `internal/config/config.go:51`
- `internal/config/config.go:72`
- `internal/config/config.go:73`
- `internal/authkey/service.go:83`
- `internal/authkey/service.go:226`

**现状问题**

当前系统同时存在：

- `.env` 单值参数：`PROXY_API_KEY`
- `.env` 多值参数：`PROXY_API_KEYS`
- 数据库存储的 auth keys

运行时又通过 `AuthKeys()` 和 `MergeSecrets()` 做二次合并。这导致“同一种能力有三种输入来源、两层聚合逻辑”。

**优化建议**

- 将 `.env` 层统一收敛为一个多值入口，例如仅保留 `PROXY_API_KEYS`
- `PROXY_API_KEY` 作为兼容读取项，进入废弃流程
- 运行时只保留一处聚合逻辑，避免 `AuthKeys()` 与 `MergeSecrets()` 语义重叠
- 管理台新增说明：静态 key 与数据库 key 的优先关系和合并规则

**移除建议**

- 中长期移除 `PROXY_API_KEY`
- 将 `PROXY_API_KEY` 迁移并合并到 `PROXY_API_KEYS`

---

#### B. 默认路由参数与数据库路由策略能力重叠

**位置**

- `internal/config/config.go:31`
- `internal/config/config.go:32`
- `internal/config/config.go:33`
- `internal/config/config.go:54`
- `internal/config/config.go:55`
- `internal/config/config.go:56`
- `internal/bootstrap/policy_config.go:16`

**现状问题**

当前默认路由既可以来自：

- `.env` 中的 `PROXY_DEFAULT_ANTHROPIC_ROUTE`
- `.env` 中的 `PROXY_DEFAULT_CHAT_ROUTE`
- `.env` 中的 `PROXY_DEFAULT_RESPONSES_ROUTE`

也可以来自数据库里的 route policy，并在启动时覆盖到运行时配置中。

这导致“默认路由配置”同时有两套来源：

- 静态 `.env`
- 动态数据库

**优化建议**

- 明确 route policy 为主来源
- `.env` 默认路由参数仅作为兼容模式或初始化导入来源
- 启动时如果数据库已有 route policy，则忽略 `.env` 默认路由
- 在后续版本中移除这三项 `.env` 路由参数的常规使用

**移除建议**

- 中长期移除：
  - `PROXY_DEFAULT_ANTHROPIC_ROUTE`
  - `PROXY_DEFAULT_CHAT_ROUTE`
  - `PROXY_DEFAULT_RESPONSES_ROUTE`

---

#### C. 默认端点定义重复：`app.go` 与 `endpoint/service.go` 各维护一份

**位置**

- `app.go:284`
- `app.go:447`
- `app.go:480`
- `internal/endpoint/service.go:170`
- `internal/endpoint/service.go:173`
- `internal/endpoint/service.go:175`
- `internal/endpoint/service.go:177`

**现状问题**

默认端点列表目前在两个地方维护：

- `app.go` 的 `defaultEndpointRoutes()`
- `internal/endpoint/service.go` 的 `defaultEndpoints()`

这意味着一旦新增、删除或修改默认端点，需要同时改两份定义，存在明显同步风险。

**优化建议**

- 抽取单一默认端点定义源
- 由 `endpoint` 模块统一导出默认端点元数据
- `app.go` 只消费该定义，不再自带一份默认路由表

**移除建议**

- 移除 `app.go` 中本地维护的默认端点常量定义
- 保留 `endpoint` 模块作为唯一来源

---

#### D. 概览与流量模块存在数据加载重叠

**位置**

- `frontend/src/stores/overview.js:2`
- `frontend/src/stores/overview.js:53`
- `frontend/src/stores/overview.js:67`
- `frontend/src/stores/traffic.js:2`
- `frontend/src/stores/traffic.js:52`
- `frontend/src/stores/traffic.js:65`

**现状问题**

`OverviewStore` 和 `TrafficStore` 都依赖 `GetOverview()`，并且都从中读取 `recent_requests`。这属于“同一接口、同一字段、多个 store 分散消费且重复拉取”的情况。

**优化建议**

- 拆分独立流量查询接口，避免流量页长期依赖总览接口
- 或者建立共享的 overview/traffic 数据层，减少重复请求
- 明确：概览页读摘要，流量页读专用请求列表接口

**移除建议**

- 中长期移除 TrafficStore 对 `GetOverview()` 的直接依赖
- 改为专用 `GetRecentRequests()` 或 `ListRecentRequests()`

---

#### E. 端点与授权 Key 页面对“重载生效”的交互提示重复

**位置**

- `frontend/src/views/EndpointsView.vue:13`
- `frontend/src/views/EndpointsView.vue:26`
- `frontend/src/views/AuthKeysView.vue:13`

**现状问题**

端点页和授权 Key 页都显式维护“保存后重载代理生效”的交互提示和按钮逻辑，但后端对不同资源的生效方式并不完全一致。这会让 UI 文案和真实行为逐渐产生偏差。

**优化建议**

- 将“资源变更后是否需要重载”抽象为统一能力描述
- 后端返回资源级别的 `requires_reload` 或 `applied_immediately` 信息
- 前端复用统一提示组件，而不是在多个页面各写一套文案

**移除建议**

- 移除页面级硬编码的重载语义
- 改为统一状态反馈模型

### 4.3 任务清单

#### 任务 1：修复 `AllowUnauthenticatedLocal` 的本地来源校验

**问题来源**

- `internal/proxy/service.go:591`
- `internal/proxy/service.go:593`

**目标**

只有本地来源请求才允许在无 Key 情况下访问。

**建议实施点**

- 在 `authorize()` 中解析 `r.RemoteAddr`
- 判断是否为 loopback 地址（`127.0.0.1` / `::1`）
- 非本地来源必须要求 Key
- 保留现有 Bearer / `x-api-key` 支持

**验收标准**

- 本机请求在允许配置下可无 Key 访问
- 非本机请求在无 Key 情况下被拒绝
- 单元测试覆盖 local / non-local 两种情况

---

#### 任务 2：修复项目设置保存会覆盖 `.env` 其他配置的问题

**问题来源**

- `internal/projectsettings/service.go:39`
- `internal/projectsettings/service.go:74`

**目标**

保存设置时只修改受管理字段，不丢失未管理字段、注释和其他配置项。

**建议实施点**

- 重构 `.env` 读写逻辑
- 支持“读取现有键值 -> 定向更新 -> 回写”
- 至少保留未知配置项
- 优先保留 `PROXY_API_KEY`、`PROXY_API_KEYS`、`PROXY_ALLOW_UNAUTHENTICATED_LOCAL`、`PROXY_MODEL_ROUTES` 等现有字段

**验收标准**

- 保存项目设置后，未被设置页管理的 `.env` 字段仍然存在
- 相关测试覆盖“保留未知 key”场景

---

#### 任务 3：修复删除被引用供应商后导致代理重载失败的问题

**问题来源**

- `app.go:159`
- `internal/bootstrap/policy_config.go:17`

**目标**

删除供应商前完成引用校验，避免删库成功但重载失败。

**建议实施点**

- 在删除 supplier 前查询 route policy 是否引用它
- 如果有启用中的策略引用，则禁止删除
- 返回明确业务错误，例如“该供应商正被默认路由策略使用”
- 前端弹窗中可补充更清晰提示

**验收标准**

- 被策略引用的供应商不能被直接删除
- 删除未被引用的供应商后可正常重载
- 增加相应测试

---

#### 任务 4：修复 OpenAI Chat / Responses 共享单一上游配置的问题，并同步清理重复参数

**问题来源**

- `internal/bootstrap/policy_config.go:36`
- `internal/bootstrap/policy_config.go:37`
- `internal/config/config.go:27`

**目标**

让 `openai-chat` 和 `openai-responses` 能真正独立绑定不同上游供应商，或者明确收敛为一种受控模型，同时避免旧参数和新参数长期重复并存。

**建议实施方案（推荐）**

后端配置结构升级为：

- Chat 独立上游配置
- Responses 独立上游配置

例如概念上拆成：

- `OpenAIChatBaseURL / APIKey / OnlyStream / UserAgent`
- `OpenAIResponsesBaseURL / APIKey / OnlyStream / UserAgent`

并同步修改：

- `bootstrap.ApplyRoutePolicies()`
- `proxy.upstreamURL()`
- `proxy.applyRequestHeaders()`
- 相关状态展示与前端文案

**参数清理要求**

- 新结构上线后，不再继续扩大 `OpenAIBaseURL / OpenAIApiKey` 的职责
- 为旧 OpenAI 通用参数设计迁移路径
- 避免新老参数长期双写

**备选方案（不推荐长期使用）**

如果短期内不重构，可先在前端限制 Chat / Responses 只能共用同一 OpenAI 供应商。

**验收标准**

- 两类协议在不同 supplier 配置下能正确请求对应上游
- 不再出现后写策略覆盖前写配置的问题
- 参数迁移逻辑明确且有兼容说明
- 增加覆盖此场景的测试

---

#### 任务 5：推进重复参数收敛与移除计划

**目标**

建立配置参数退役机制，减少“同一能力多个参数共存”的情况。

**建议实施点**

- 将 `PROXY_API_KEY` 标记为兼容参数，主入口收敛到 `PROXY_API_KEYS`
- 将 `.env` 默认路由参数标记为兼容参数，主入口收敛到数据库 route policy
- 为每个待退役参数补充：
  - 读取兼容策略
  - 写入策略
  - UI 展示策略
  - 文档说明
- 在后续版本中逐步移除旧参数写入能力，仅保留只读兼容期

**验收标准**

- 形成参数清单、主参数、兼容参数、退役时序
- 新代码不再继续写入重复参数

---

### 4.4 第一阶段交付物

- 后端安全修复
- `.env` 保留式更新机制
- supplier 删除前引用校验
- OpenAI 上游配置模型修正
- 重复参数收敛方案与迁移策略
- 对应单元测试与回归测试

---

## 5. 第二阶段：中优先级稳定性增强与模块收敛

### 5.1 阶段目标

修复当前“功能表面可用，但语义不稳定”的问题，提升配置可信度与行为一致性。

### 5.2 任务清单

#### 任务 5：禁止禁用供应商继续被默认路由使用

**问题来源**

- `internal/routepolicy/service.go:124`
- `internal/bootstrap/policy_config.go:17`

**建议实施点**

- 保存 route policy 时，若 supplier 为 disabled，则拒绝保存
- 应用 route policy 时，若 supplier 被禁用，则返回可解释错误
- 前端路由策略下拉中优先只展示启用供应商，或对禁用项做显著标记

**验收标准**

- 禁用供应商不能作为有效默认路由目标继续生效

---

#### 任务 6：补齐跨协议错误响应转换

**问题来源**

- `internal/proxy/service.go:511`

**建议实施点**

- 当上游返回错误时，不直接透传原错误结构
- 按下游协议统一包装错误响应
- Anthropic 下游返回 Anthropic 风格错误
- OpenAI 下游返回 OpenAI 风格错误

**验收标准**

- 跨协议错误场景下客户端拿到符合下游协议的错误结构

---

#### 任务 7：将授权 Key 生成统一收口到后端

**问题来源**

- `frontend/src/stores/authKeys.js:12`
- `frontend/src/stores/authKeys.js:14`

**建议实施点**

- 新增后端方法，例如 `GenerateAuthKeySecret()`
- 前端点击“生成”时调用后端
- 删除前端本地随机生成逻辑

**验收标准**

- 所有生成出来的 Key 都由后端统一产生
- 不依赖浏览器/WebView 的 crypto 能力差异

---

#### 任务 8：提升健康检查真实性

**问题来源**

- `internal/supplier/health.go:53`
- `internal/supplier/health.go:95`

**建议实施点**

- 不再把 `200-499` 全部视为 reachable
- 区分：
  - reachable
  - auth_failed
  - warning
  - unreachable
- 优先探测更接近真实 API 能力的接口
- 前端健康状态颜色与说明同步优化

**验收标准**

- 健康检查结果能更准确反映“配置是否可用”

---

#### 任务 9：收敛模块重叠与重复数据加载

**目标**

减少同一职责被多个模块分散承担的问题，降低同步修改成本。

**建议实施点**

- 抽取统一默认端点定义源，移除 `app.go` 与 `endpoint/service.go` 的双份维护
- 为流量页提供专用请求列表接口，减少 `OverviewStore` / `TrafficStore` 对 `GetOverview()` 的重复依赖
- 统一“是否需要重载代理”的资源反馈模型，移除页面级硬编码提示
- 梳理 `GetOverview()` 的职责边界，避免它持续膨胀为万能聚合接口

**验收标准**

- 默认端点只保留一处定义来源
- 流量页不再依赖概览接口读取请求明细
- 资源更新后的重载提示改为统一机制

---

### 5.3 第二阶段交付物

- 更一致的 supplier / route policy 关系约束
- 更合理的错误协议输出
- 后端统一 Key 生成机制
- 更可靠的健康检查结果
- 默认端点定义单一化
- 概览/流量模块职责收敛
- 重载提示机制统一

---

## 6. 第三阶段：运行态一致性与可观测性增强

### 6.1 阶段目标

解决“已保存但未成功生效”的可见性问题，提升问题排查能力。

### 6.2 任务清单

#### 任务 9：区分“持久化成功”和“运行态应用成功”

**问题来源**

- `app.go:114`
- `app.go:146`
- `app.go:159`
- `app.go:196`

**建议实施点**

- 把“保存成功”和“重载成功”拆成两个阶段性结果
- 给前端返回结构化结果，例如：
  - `saved: true/false`
  - `reloaded: true/false`
  - `message`
  - `details`
- 前端 UI 明确提示用户：
  - 配置已保存，但代理未成功应用
  - 或配置与运行态均已成功更新

**验收标准**

- 用户能清楚知道当前是“配置有问题”还是“重载失败”

---

#### 任务 10：增加预校验 / dry-run 能力

**建议实施点**

在真正落盘和重载前，增加一层预校验：

- route policy 引用是否合法
- supplier 是否可用
- 配置组合是否冲突
- 端点路径是否合法、是否重复

**验收标准**

- 常见错误在保存前就能被发现，而不是写入后才失败

---

#### 任务 11：增强错误日志与诊断输出

**建议实施点**

- 重载失败时输出更具体的错误来源
- 状态页增加最近一次重载结果说明
- 记录配置应用失败的阶段信息

**验收标准**

- 出现异常时可以更快定位问题属于：配置、供应商、策略、上游连通性还是协议转换

---

### 6.3 第三阶段交付物

- 结构化保存/重载反馈
- 预校验逻辑
- 更清晰的错误诊断信息

---

## 7. 第四阶段：测试体系与工程质量提升

### 7.1 阶段目标

为后续协议扩展和功能迭代建立可靠回归基础。

### 7.2 测试补强任务

建议新增以下测试：

#### 后端测试

1. `AllowUnauthenticatedLocal` 本地 / 非本地访问分支测试
2. 删除被 route policy 引用 supplier 的保护测试
3. OpenAI Chat / Responses 双 supplier 场景测试
4. `.env` 保留未知 key 的保存测试
5. route policy 使用 disabled supplier 的校验测试
6. 跨协议错误响应结构测试
7. health check 状态分类测试
8. 默认端点单一来源定义测试
9. 重复参数迁移兼容测试（`PROXY_API_KEY` -> `PROXY_API_KEYS`）
10. `.env` 默认路由参数与数据库 route policy 优先级测试

#### 前端测试（如后续引入测试框架）

1. 设置页保存失败提示
2. 供应商删除失败提示
3. 路由策略与 supplier 启用状态联动展示
4. Auth Key 生成与复制交互
5. 状态页对 reload 成功/失败的反馈
6. 流量页改用专用请求接口后的刷新与自动刷新行为
7. 统一重载提示组件在端点页/授权页的行为一致性

### 7.3 工程增强建议

- 补充更明确的错误码或错误类型
- 增加关键服务的集成测试
- 为重要变更建立最小回归清单
- 为后续协议转换能力补充转换矩阵测试

---

## 8. 建议执行顺序

推荐执行顺序如下：

### 第一批（必须先做）

1. 本地免鉴权校验修复
2. `.env` 保留式更新
3. supplier 删除前引用校验
4. OpenAI Chat / Responses 上游配置拆分
5. 重复参数主入口收敛（鉴权参数、默认路由参数）

### 第二批（紧接着做）

6. 禁用 supplier 与 route policy 约束
7. 跨协议错误结构统一
8. 后端统一生成 auth key secret
9. 健康检查增强
10. 默认端点定义单一化
11. 流量页与概览页数据职责拆分
12. 统一页面重载反馈机制

### 第三批（收尾增强）

13. 结构化保存/重载反馈
14. 预校验机制
15. 日志与诊断增强
16. 测试体系补强

---

## 9. 影响范围预估

本次优化主要影响以下模块：

### 后端重点文件

- `internal/proxy/service.go`
- `internal/projectsettings/service.go`
- `internal/config/config.go`
- `internal/bootstrap/policy_config.go`
- `internal/routepolicy/service.go`
- `internal/supplier/service.go`
- `internal/supplier/health.go`
- `internal/endpoint/service.go`
- `internal/authkey/service.go`
- `app.go`

### 前端重点文件

- `frontend/src/stores/authKeys.js`
- `frontend/src/stores/suppliers.js`
- `frontend/src/stores/settings.js`
- `frontend/src/stores/overview.js`
- `frontend/src/stores/traffic.js`
- `frontend/src/views/SuppliersView.vue`
- `frontend/src/views/AuthKeysView.vue`
- `frontend/src/views/EndpointsView.vue`
- `frontend/src/views/SettingsView.vue`
- `frontend/src/views/OverviewView.vue`

---

## 10. 计划完成后的预期收益

如果按本计划完成优化，项目将获得以下提升：

1. **安全性提升**：避免“本地免鉴权”被错误暴露为全局免鉴权
2. **数据可靠性提升**：保存设置不再丢失现有配置
3. **运行稳定性提升**：删除、保存、重载流程更可控
4. **路由正确性提升**：多 supplier 场景与协议路由行为更真实一致
5. **维护效率提升**：更清晰的错误反馈和更扎实的测试回归体系
6. **产品可信度提升**：前端展示与后端真实能力更一致

---

## 11. 最终建议

建议本轮不要先做大范围 UI 重构，也不要先扩展更多协议特性，而是优先完成“**安全 + 配置一致性 + 路由正确性 + 可验证性**”四项基础建设。

原因很简单：

- 当前主要风险不在“功能不够多”
- 而在“已有功能在边界条件下是否可靠”

因此，最合理的开发策略是：

> 先把网关内核和配置行为做稳，再继续扩展协议能力和管理体验。

---

## 12. 附：建议的任务拆分格式

后续如果要正式进入实现阶段，建议把每项任务继续拆成如下格式：

- 任务名称
- 涉及文件
- 修改点
- 风险点
- 测试点
- 完成标准

这样方便后续逐项落地与验收。
