# AIGateway Console TODO

> 本文件只记录 Console 子项目自身轻量待办。跨项目任务以根目录 `TODO.md` 与 `TASK/projects/aigateway-console/` 为准。

## 已完成

- [x] Consumer 真源切到 Portal DB：`portal_user`、`portal_api_key`、`portal_invite_code`。
- [x] `/v1/consumers` 支持列表、详情、创建、更新、状态、软删除。
- [x] 独立密码重置接口：`POST /v1/consumers/{name}/password/reset`。
- [x] 邀请码管理接口：创建、列表、禁用、重新启用。
- [x] AI Quota consumer 校验改为查 DB 真源。
- [x] DB -> key-auth 周期投影与启动回填任务已幂等。
- [x] 组织架构前端新增重置密码与邀请码管理入口。
- [x] Route / AI Route / MCP 消费者下拉采用 `DB 用户 + 当前配置历史值` 并集策略。
- [x] MCP 名称前后端按 RFC1123 小写校验，K8s 422 统一转 HTTP 400。
- [x] 干净 release 环境下，Console 启动会补齐 Console 自有 PortalDB 表；`portaldb.autoMigrate=false` 不再导致 AI Quota / AI Sensitive 页面首次访问缺表。

## 待完成

- [ ] 为 `portal_user` / `portal_api_key` 增补显式索引校验脚本与启动告警。
- [ ] 增加投影同步指标：成功、失败、耗时、差异数，并接入监控。
- [ ] 增加 migration dry-run 模式，只比对不写入，用于生产演练。
- [ ] 增加邀请码使用明细表与后台查询。
- [ ] 增加 API Key 审计日志：创建、禁用、删除、使用来源。

## 发布前检查

- [ ] Console 后端编译通过。
- [ ] Console 前端 build 通过。
- [ ] 同一数据库下，Portal 注册用户可在 Console 组织架构可见。
- [ ] 邀请码时间字段显示正常，不再出现 `Invalid Date`。
- [ ] disabled 用户在一个同步周期内完成网关鉴权失效。
