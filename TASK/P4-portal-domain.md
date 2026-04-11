# P4 Portal 业务域

- 状态：`doing`
- 依赖：`P1`, `P2`

## 目标

- 迁移 Portal MySQL 真相源相关能力，并统一成 GoFrame DAO/DO/Entity 模式。

## 任务

- [x] 建立 Portal DB client、Schema bootstrap 与 MySQL 接入骨架。
- [ ] 建立 Portal DB DAO、DO、Entity。
- [x] 迁移 Consumer 首批接口。
- [x] 迁移组织树与账号管理首批 CRUD。
- [x] 迁移 Invite Code。
- [x] 迁移 Asset Grant。
- [x] 迁移 Model Assets / Agent Catalog / Pricing / Stats 首批主链。
- [ ] 迁移 AI Quota / AI Sensitive Word DB 能力。

## 验收点

- [x] Portal DB 访问入口已不再依赖原生 JDBC。
- [x] Consumer/Org/Model Assets/Agent Catalog/Grant 首批接口可工作。
- [x] 与旧前端契约保持可迁移。

## 测试

- [x] Portal service sqlmock 测试。
- [x] Model Assets / Grants / Agent Catalog 服务测试。
- [ ] MySQL testcontainers 集成测试。
- [ ] Consumer / Org / Invite Code / Model Assets / Agent Catalog 契约测试。

## 本轮说明

- 本轮先按第一批范围交付 `Consumer / Org / Invite Code`，并通过 `portaldb` real client + schema bootstrap 打通最小真相源链路。
- 本轮继续补上 `Asset Grant / Model Assets / Model Binding / Binding Price Version / Agent Catalog`，并补齐前端依赖的 `/v1/assets/*/grants`、`/v1/ai/model-assets*`、`/v1/ai/agent-catalog*` 接口。
- `portaldb` schema bootstrap 现已覆盖 `portal_asset_grant / portal_model_asset / portal_model_binding / portal_model_binding_price_version / portal_agent_catalog / ai_sensitive_* / job_run_record`，为 P5 和后续 AI 域迁移预留真相源表。
- 当前仍未补齐 GoFrame `DAO / DO / Entity` 生成产物，`AI Quota / AI Sensitive Word` 正式管理 API 也还没有拉进来。
- `org/template`、`org/export`、`org/import` 继续延后，避免把第一批 Portal DB 接入做成超大提交。
