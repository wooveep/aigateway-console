# AIGateway Console Memory（当前架构基线）

## 1. 架构不是“纯 K8s”了，而是双领域双存储
- 网关运行态领域：继续以 K8s CRD 为主（Route/Domain/Plugin/MCP 等）。
- 用户身份领域：以 Portal DB 为主（`portal_user`、`portal_api_key`、`portal_invite_code`）。
- 结论：Consumer 业务真源已经切到 DB，不再把 `key-auth` consumers 当主数据。

## 2. Consumer 真源与投影机制
- Console 的 `/v1/consumers` 读写主数据走 `PortalUserJdbcService`。
- `key-auth` 仅做运行态投影缓存，由 `PortalConsumerProjectionService` 周期收敛（默认 30s）。
- 固定投影鉴权格式：`Authorization: Bearer <raw_key>`。
- 用户状态语义：
  - `active`：可登录 Portal，可网关鉴权。
  - `pending`：不可登录 Portal，可网关鉴权（迁移承接态）。
  - `disabled`：不可登录 Portal，不可网关鉴权（并禁用该用户所有 API Key）。

## 3. 迁移策略（幂等）
- 启动期一次性回填：从现存 `key-auth` consumers 导入 `portal_user`（`pending`，`source=migration-keyauth`）。
- 历史 token 回填 `portal_api_key`，按 `consumer_name + key_hash` 去重，保证重复执行幂等。
- 回填后立刻触发一次全量投影，确保 DB 与运行态一致。

## 4. 前端组织架构基线
- 用户管理围绕 DB 用户：创建、编辑、启用/禁用、独立密码重置、邀请码管理。
- 凭证编辑入口已下线；认证信息只读展示（来自 `portal_api_key` 派生）。
- Route/AI Route/MCP 的消费者下拉应使用并集策略：`DB 用户 + 当前配置历史值`，避免编辑时误丢。

## 5. MCP 命名约束
- MCP 名称必须满足 RFC1123 小写子域名规则（仅 `a-z0-9.-`，首尾字母数字）。
- 前后端都要校验，K8s 422 统一转为 `ValidationException`（HTTP 400），避免前端看到 500。

## 6. 时间字段一致性
- 邀请码、密码重置等 `LocalDateTime` 返回统一 ISO 字符串，避免前端 `Invalid Date`。
- 前端解析仅按标准日期解析，减少后端时间结构分支兼容逻辑。

## 7. PortalDB 自有表初始化口径
- `portaldb.autoMigrate=false` 的真实语义是：Console 不负责兜底创建 Portal 共享表，也不执行 legacy 数据迁移。
- Console 自有 PortalDB 表仍必须由 Console 启动时保证存在，包括 `portal_ai_quota_balance`、`portal_ai_quota_schedule_rule`、`portal_model_binding_price_version`、`portal_ai_sensitive_*`、`job_run_record`。
- 若共享表已经由 Portal migration 建好，Console 即使在 release 环境关闭 `autoMigrate`，也应能安全创建上述自有表并正常打开 `AI Quota`、`AI Sensitive` 等页面。
