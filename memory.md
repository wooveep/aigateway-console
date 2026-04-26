# AIGateway Console Memory

> 本文件只保留 Console 子项目架构基线与重要 caveat。跨项目决策看根目录 `Memory.md`。

## 架构基线

- Console 不是“纯 K8s 控制台”。网关运行态领域继续以 K8s CRD 为主，用户身份领域以 Portal DB 为主。
- Consumer 主数据真源已切到 Portal DB，不再把 `key-auth` consumers 当主数据。
- Console 的 `/v1/consumers` 读写主数据走 `PortalUserJdbcService`。
- `key-auth` 仅做运行态投影缓存，由 `PortalConsumerProjectionService` 周期收敛，默认周期 30s。
- 固定投影鉴权格式：`Authorization: Bearer <raw_key>`。

## 用户状态语义

- `active`：可登录 Portal，可网关鉴权。
- `pending`：不可登录 Portal，可网关鉴权，主要用于迁移承接态。
- `disabled`：不可登录 Portal，不可网关鉴权，并禁用该用户所有 API Key。

## 迁移与初始化

- 启动期一次性回填从现存 `key-auth` consumers 导入 `portal_user`，状态为 `pending`，来源为 `migration-keyauth`。
- 历史 token 回填到 `portal_api_key`，按 `consumer_name + key_hash` 去重，重复执行必须幂等。
- 回填后必须触发一次全量投影，确保 DB 与运行态一致。
- `portaldb.autoMigrate=false` 的含义是 Console 不兜底创建 Portal 共享表，也不执行 legacy 数据迁移。
- Console 自有 PortalDB 表仍必须由 Console 启动保证存在，包括 `portal_ai_quota_balance`、`portal_ai_quota_schedule_rule`、`portal_model_binding_price_version`、`portal_ai_sensitive_*`、`job_run_record`。

## 前端与接口约束

- 用户管理围绕 DB 用户：创建、编辑、启停、软删除、密码重置、邀请码管理。
- 凭证编辑入口已下线，认证信息只读展示，来源于 `portal_api_key` 派生。
- Route / AI Route / MCP 的消费者下拉必须使用 `DB 用户 + 当前配置历史值` 并集，避免编辑存量配置时误删历史值。
- MCP 名称必须满足 RFC1123 小写子域名规则：仅 `a-z0-9.-`，首尾为字母或数字。
- K8s 422 应统一转为 `ValidationException` / HTTP 400，避免前端暴露 500。
- 邀请码、密码重置等 `LocalDateTime` 返回统一 ISO 字符串，避免前端 `Invalid Date`。
