# GoFrame 重写任务台账

## 状态约定

- `todo`：尚未开始
- `doing`：正在进行
- `done`：已完成
- `blocked`：被外部条件阻塞

## 阶段顺序

1. `P0-baseline-analysis.md`
2. `P1-goframe-foundation.md`
3. `P2-platform-core.md`
4. `P3-gateway-resource-domain.md`
5. `P4-portal-domain.md`
6. `P5-jobs-and-reconcile.md`
7. `P6-cutover-and-rename.md`

## 统一要求

- 每阶段都要补充“验收点”和“单元测试/集成测试”。
- 每阶段完成后都要执行：
  - `go test ./...`
  - 阶段内 smoke 验证
  - 与 `backend-java-legacy/` 的行为差异记录
- 变更真相源、表结构、K8s 资源映射或前端接口时，需要同步更新阶段文档。

## 当前进度

- `P0`：done
- `P1`：done
- `P2`：done
- `P3`：doing
- `P4`：todo
- `P5`：todo
- `P6`：todo
