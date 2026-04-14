# P5 任务与调和

- 状态：`doing`（主链已完成，aftercare 持续中）
- 依赖：`P3`, `P4`

## 目标

- 迁移全部 `@Scheduled` 任务到 GoFrame `cron/job`，并支持幂等与手动触发。

## 任务

- [x] 迁移 Consumer Projection。
- [x] 迁移 Consumer Level Auth Reconcile。
- [x] 迁移 AI Sensitive Projection。
- [x] 迁移 AI Plugin Execution Order Reconcile。
- [x] 建立 job 触发来源、运行记录和幂等控制。

## 验收点

- [x] 所有任务既可定时触发，也可手工触发。
- [x] 重复执行不会破坏目标态。
- [x] 失败路径有明确错误记录和最近执行信息。

## 测试

- [x] cron/job 单元测试。
- [x] 幂等测试。
- [x] 手工触发入口测试。
- [ ] 失败重试测试。

## Aftercare

- [ ] `P5-AF-01`：补失败重试与失败恢复测试。
- [ ] `P5-AF-02`：继续完善 jobs 的运行态观测和失败排查体验。

## 本轮说明

- 新增 `internal/service/jobs` 统一任务服务，负责 job registry、cron 注册、手工触发、最近执行结果查询和 `job_run_record` 持久化。
- 新增受保护内部接口 `/internal/jobs`、`/internal/jobs/{name}`、`/internal/jobs/{name}/trigger`，便于后续运维和 review。
- `portal-consumer-projection` 目前把 Portal 用户投影为 K8s generic resource `consumers`；后续若切换到真实 key-auth / consumer CR，再在该任务内部替换投影目标。
- `portal-consumer-level-auth-reconcile` 当前会按 `allowedConsumerLevels` 展开 `routes / ai-routes / mcp-servers` 的 `allowedConsumers`。
- `ai-sensitive-projection` 当前把 Portal DB 中的检测/替换/系统配置聚合到 `ai-sensitive-projections/default` 资源，先保证任务链路和状态落库。
- `ai-plugin-execution-order-reconcile` 当前对齐 `ai-statistics`、`ai-data-masking` 的 phase/priority 到 generic `wasm-plugins` 资源，未进入 CRD 字段级迁移。
- 当前判断：`P5` 主链迁移已完成，剩余工作集中在失败重试验证、运行态观测和排障体验。
