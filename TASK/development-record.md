# GoFrame 重写开发记录

## 2026-04-11

### 阶段 / 任务

- `P0` 契约样本补齐与记录基线建立
- `P1` GoFrame 基座收口
- `P2` 平台核心迁移
- `P3` 网关资源域首版迁移

### 改动摘要

- 新增统一开发记录文件，作为阶段文档之外的 review 汇总真相源。
- `backend/` 新增管理员鉴权中间件、会话管理、用户接口、Dashboard 接口、系统初始化与 `aigateway-config` 读写。
- `utility/clients/{k8s,grafana,portaldb}` 从占位接口补为可配置 client 抽象，默认使用内存/fake 形态支撑迁移与测试。
- `gateway` 域新增内存版资源 CRUD，覆盖 `Route / Domain / TLS / Service / ServiceSource / ProxyServer / Wasm / AI Route / Provider / MCP` 及插件实例接口。
- 补充 `resource/dashboard/*.json` 与 `resource/public/mcp-templates/*`，让 Dashboard 配置下载和 MCP 模板静态访问能在新后端工作。

### 关键文件 / 接口

- `backend/internal/cmd/cmd.go`
- `backend/internal/middleware/auth.go`
- `backend/internal/service/platform/service.go`
- `backend/internal/service/gateway/service.go`
- `backend/internal/controller/{session,user,system,dashboard,gateway}/*`
- `backend/utility/clients/{k8s,grafana,portaldb}/*`
- `/session/*`
- `/user/*`
- `/system/*`
- `/dashboard/*`
- `/v1/routes`
- `/v1/domains`
- `/v1/tls-certificates`
- `/v1/services`
- `/v1/service-sources`
- `/v1/proxy-servers`
- `/v1/wasm-plugins`
- `/v1/ai/*`
- `/v1/mcpServer*`

### 验证结果

- `go test ./...` 通过。
- 手工 smoke 已验证：
  - `GET /system/info`
  - `POST /system/init`
  - `POST /session/login`
  - `GET /user/info`
  - `GET /dashboard/info`
  - `GET /healthz/ready`
  - `GET /system/aigateway-config`
  - `GET /v1/routes`
  - `GET /v1/mcpServer`

### 与 backend-java-legacy 的行为差异

- `P2` 当前管理员配置真相源仍通过内存版 `k8s` fake client 承载，接口行为已对齐到可初始化、可登录、可改密，但未接真实 K8s Secret。
- `P2` Dashboard 目前保留接口、静态配置读取和 native 数据结构；Grafana 代理与 datasource 自动初始化暂未接真实后端。
- `P3` 当前资源域是控制面契约优先的首版实现，CRUD 和 smoke 可用，但仍未完成真实 CRD / ConfigMap / 注解映射的一致性迁移。

### Review 关注点 / 风险

- `gateway` 域当前以 in-memory store 稳定 HTTP 契约，后续接真实 K8s client 时要重点 review 资源字段映射是否与 Java 旧实现一致。
- `platform` 域已形成会话和系统初始化主链，但管理员 Secret、ConfigMap 冲突控制、Dashboard 代理仍需在真实环境 client 接入后复核。
- 当前 `success/data/message` 信封和 HTTP 状态已统一，但“错误码枚举化”仍可在后续迭代继续收口。

## 2026-04-11（第二批）

### 阶段 / 任务

- `P3` 网关资源域收口
- `P4` Portal 业务域第一批启动

### 改动摘要

- `utility/clients/k8s` 新增 `kubectl` 驱动的 real client，实现 `ConfigMap / Secret / generic resource ConfigMap` 持久化，保留 fake client 供测试和契约样本复用。
- 为 `gateway` service 补保护资源、internal 资源、plugin instance 的行为测试，收紧 P3 的基础约束。
- `utility/clients/portaldb` 从占位健康检查升级为真实 SQL client，支持 MySQL driver、schema bootstrap、测试态 `NewFromDB` 注入。
- 新增 `portal` 域 service 与 controller，打通 `Consumer / Org Accounts / Org Departments / Invite Codes` 首批接口。
- `cmd` 装配补齐 `portal` 路由绑定，并扩展 `hack/config.yaml` 的 `k8s / portaldb / grafana` 配置样例。

### 关键文件 / 接口

- `backend/utility/clients/k8s/client.go`
- `backend/utility/clients/portaldb/client.go`
- `backend/internal/service/gateway/service_test.go`
- `backend/internal/service/portal/service.go`
- `backend/internal/service/portal/service_test.go`
- `backend/internal/controller/portal/http.go`
- `/v1/consumers*`
- `/v1/org/departments*`
- `/v1/org/accounts*`
- `/v1/portal/invite-codes*`

### 验证结果

- `go test ./...` 通过。
- 新增验证覆盖：
  - P3：保护资源删除拦截、internal 资源写入拦截、plugin instance round-trip
  - P4：invite code 创建、org account 状态更新的 sqlmock 测试

### 与 backend-java-legacy 的行为差异

- P3 real client 当前把通用资源落为带标签的 `ConfigMap`，还没有直接映射到旧 Java 使用的 CRD/注解组合。
- P4 第一批采用手写 SQL service + schema bootstrap 方式先打通真相源，没有生成 GoFrame `DAO / DO / Entity` 产物。
- P4 目前只覆盖 `Consumer / Org / Invite Code`，`Asset Grant / Model Assets / Pricing / Stats / AI Quota / AI Sensitive` 仍未迁移。
- `org/template`、`org/export`、`org/import` 仍未实现，本轮明确延后。

### Review 关注点 / 风险

- `kubectl` real client 依赖运行环境中的 `kubectl` 与 kubeconfig/namespace 配置，部署环境需要额外验证二进制和权限。
- P3 资源目前是“真实 K8s 持久化 + 泛化资源模型”，不是“真实 CRD 字段级全量对齐”，review 时要特别关注这个边界。
- P4 首批 DB 接入优先保证接口闭环和可测性，后续补 `DAO / DO / Entity` 时要避免与当前 SQL schema bootstrap 产生重复定义。

## 2026-04-11（第三批）

### 阶段 / 任务

- `P4` Model Assets / Agent Catalog / Asset Grants
- `P5` Jobs / Reconcile

### 改动摘要

- `portaldb` schema bootstrap 继续扩展，补上 `portal_asset_grant / portal_model_asset / portal_model_binding / portal_model_binding_price_version / portal_agent_catalog / ai_sensitive_* / job_run_record`。
- `portal` 域新增 `Asset Grants`、`Model Assets`、`Model Bindings`、`Binding Price Versions`、`Agent Catalog` 服务与 HTTP 路由，补齐前端依赖的 `/v1/assets/*/grants`、`/v1/ai/model-assets*`、`/v1/ai/agent-catalog*`。
- `portal` 写路径统一接入 hook，给后续任务体系提供单一入口。
- 新增 `internal/service/jobs` 与 `internal/controller/jobs`，落地统一 job registry、cron 注册、手工触发、最近执行结果查询、幂等跳过与 `job_run_record` 落库。
- 首批迁移四类任务：
  - `portal-consumer-projection`
  - `portal-consumer-level-auth-reconcile`
  - `ai-sensitive-projection`
  - `ai-plugin-execution-order-reconcile`

### 关键文件 / 接口

- `backend/internal/service/portal/assets.go`
- `backend/internal/controller/portal/http.go`
- `backend/internal/service/jobs/service.go`
- `backend/internal/controller/jobs/http.go`
- `backend/internal/job/jobs.go`
- `backend/internal/cmd/cmd.go`
- `/v1/assets/{assetType}/{assetId}/grants`
- `/v1/ai/model-assets*`
- `/v1/ai/agent-catalog*`
- `/internal/jobs`
- `/internal/jobs/{name}`
- `/internal/jobs/{name}/trigger`

### 验证结果

- `go test ./...` 通过。
- 新增验证覆盖：
  - P4：`ReplaceAssetGrants` sqlmock 测试、`GetAgentCatalogOptions` K8s options 测试
  - P5：`portal-consumer-projection` 手工触发测试、同快照幂等跳过测试

### 与 backend-java-legacy 的行为差异

- P4 当前仍使用手写 SQL service，没有生成 GoFrame `DAO / DO / Entity` 产物。
- `portal-consumer-projection` 当前把 Portal 用户投影到 generic `consumers` 资源，而不是旧 Java 中的 key-auth consumer/credential 真实控制面对象。
- `ai-sensitive-projection` 当前先把 Portal DB 规则聚合为 `ai-sensitive-projections/default` 资源，尚未接入旧 Java 的 `ai-data-masking` plugin instance 配置投影细节。
- `ai-plugin-execution-order-reconcile` 当前只对 generic `wasm-plugins` 资源的 `phase/priority` 做对齐，未迁移旧 Java 的 built-in WasmPlugin CR 版本迁移与 legacy CR 合并逻辑。
- `P6` 未启动，本轮明确保持旧命名与现有前端调用路径不变。

### Review 关注点 / 风险

- `P4` 的 HTTP 契约已补齐，但 Portal 领域模型仍处于“service + schema bootstrap”阶段，后续补 `DAO / DO / Entity` 时要避免出现第二套字段命名。
- `P5` 的运行记录与幂等已经落库，但“基础指标”目前主要体现为最近执行状态和运行耗时，还没有外接 Prometheus/Grafana 指标面。
- `consumer projection` 与 `AI sensitive projection` 都是“迁移路径优先”的首版实现，真实运行时目标资源若切换为更贴近旧 Java 的对象模型，需要再次 review 任务输出契约。

## 2026-04-11（第四批）

### 阶段 / 任务

- `P6` 切换与命名收口

### 改动摘要

- 后端把 `/system/higress-config` 硬切为 `/system/aigateway-config`，并把 ConfigMap 业务键从 `higress` 改为 `aigateway`。
- `platform` 服务新增一次性迁移逻辑，读取到旧 `higress` 键时会自动改写为 `aigateway` 并回写到 `ConfigMap`。
- 前端同步切换 `system` service、系统页标题、语言存储键、默认域名、AI Route 示例文案和页面标题到 `aigateway` 命名。
- `backend/go.mod` 与全部 Go import path 已切换到 `github.com/alibaba/aigateway-group/aigateway-console/backend`。
- 当前项目目录已从 `higress-console` 改名为 `aigateway-console`，并同步更新父仓库活跃脚本、Helm 依赖路径和根级协作文档引用。
- Helm chart、安装脚本、镜像默认名、MCP 模板值和公开 README/CONTRIBUTING 文档已收口到 `aigateway-console`。

### 关键文件 / 接口

- `backend/internal/controller/system/http.go`
- `backend/internal/service/platform/service.go`
- `backend/internal/service/platform/service_test.go`
- `backend/utility/clients/k8s/client.go`
- `frontend/src/services/system.ts`
- `frontend/src/i18n.ts`
- `frontend/src/views/system/SystemPage.vue`
- `frontend/src/views/gateway/DomainPage.vue`
- `frontend/src/views/ai/AiRoutePage.vue`
- `backend/go.mod`
- `helm/Chart.yaml`
- `helm/install-console.sh`
- `../scripts/aigateway-dev.py`
- `../higress/helm/higress/{Chart.yaml,Chart.lock}`
- `GET /system/aigateway-config`
- `PUT /system/aigateway-config`

### 验证结果

- `go test ./...` 通过。
- `npm run build` 通过。
- `helm template aigateway-console ./helm` 通过。
- `helm dependency update && helm template aigateway ./`（`higress/helm/higress`）通过。
- 手工 smoke 已验证：
  - `GET /system/info`
  - `GET /healthz/ready`
  - `GET /landing`
  - `POST /system/init`
  - `POST /session/login`
  - `GET /system/aigateway-config`（登录后）

### 与 backend-java-legacy 的行为差异

- 本轮按硬切策略执行，`/system/higress-config` 旧接口不再保留兼容别名。
- 运行时兼容仅保留一次性迁移：旧 `ConfigMap.data.higress` 会自动翻转为 `ConfigMap.data.aigateway`，不做长期双读双写。
- 历史归档名称未整体重写；`backend-java-legacy/`、根目录历史记忆文件与 `higress/` 子项目中的历史语义仍保持原样。

### Review 关注点 / 风险

- Helm 默认镜像仓库已切到 `.../aigateway-console`，实际发布环境需要确认对应镜像标签存在，避免部署时拉取失败。
- 父仓库目前只收口了活跃脚本和协作文档；其他历史材料中仍可能保留旧名称，这是刻意保留而非遗漏。
- `P6` 当前已完成工程内硬切，但若后续还要同步外部仓库名、发布流水线或文档站链接，需要在仓库外继续做配套迁移。
