# P3 网关资源域

- 状态：`doing`
- 依赖：`P1`, `P2`

## 目标

- 迁移 Route/Domain/Service/Proxy/TLS/Wasm/AI Route/Provider/MCP 及其 K8s 适配层。

## 任务

- [x] 建立 `utility/clients/k8s` 接口与 fake client。
- [x] 为 `utility/clients/k8s` 补 `kubectl` 驱动的 real client，并接入 `ConfigMap/Secret` 持久化。
- [x] 迁移 Route 资源模型与控制器。
- [x] 迁移 Domain / TLS。
- [x] 迁移 Service / ServiceSource / ProxyServer。
- [x] 迁移 WasmPlugin / WasmPluginInstance。
- [x] 迁移 AI Route / Llm Provider。
- [x] 迁移 MCP Server 资源。
- [x] `P3-CP-01`：`mcp-servers` 真实读模型迁移到 `Ingress + higress-config + McpBridge/registry + route auth/plugin instance`。
- [x] `P3-CP-02`：`mcp-servers` 保存/删除副作用对齐 Java save strategies，覆盖 `match_list / servers[] / route auth / default-mcp plugin`。
- [ ] `P3-CP-03`：`ai-routes` public/internal/fallback/plugin/EnvoyFilter 写路径对齐，并补删除清理。
- [ ] `P3-CP-04`：`ai-providers` `ai-proxy / ServiceSource / service-scope instance / ACTIVE_PROVIDER_ID / ai-route resync` 对齐。
- [x] `P3-CP-05`：Portal/Jobs 对真实控制面聚合字段兼容回归。
- [x] `P3-CP-06`：real cluster smoke、删除清理与回归验证。

## 验收点

- [x] K8s 资源 CRUD 不再局限于内存 fake client。
- [ ] K8s 资源 CRUD 与旧实现行为一致。
- [ ] 关键注解、ConfigMap、CRD 映射无缺项。
- [x] `MCP / AI Route / Provider` 读取真相源正确，不再依赖 `console.aigateway.io/type=resource` 承载核心资源。
- [ ] `MCP / AI Route / Provider` 写入后的运行时副作用完整，包含 plugin instance / fallback / EnvoyFilter / ServiceSource 联动。
- [ ] `MCP / AI Route / Provider` 删除清理完整，不残留历史兼容路由、matchRules、EnvoyFilter 与 service-scope instance。
- [x] `/v1/routes`、`/v1/domains`、`/v1/mcp*`、`/v1/ai/*` 可 smoke。
- [x] `/v1/wasm-plugins/:name/readme` 与 `global/domain/route/service` plugin instance 首批缺口已补齐。
- [x] MCP route 显式 metadata 与 `ingressClass` 首批传递链路已补齐。

## 测试

- [x] Fake client 单元测试。
- [x] gateway service 保护资源 / internal 资源 / plugin instance 行为测试。
- [ ] 资源转换测试。
- [x] 冲突/校验失败测试。
- [x] Route / MCP / AI Route 契约样本测试。
- [x] Route / Service / Wasm / MCP 首批深校验与 builtin metadata 回退测试。

## 本轮说明

- 当前 P3 已从“HTTP 契约 + 内存版资源存储”推进到“HTTP 契约 + real/fake 双 client”，其中 real client 通过 `kubectl` 落实际 `ConfigMap/Secret` 持久化。
- 资源域仍未完成到旧 Java 的 CRD / 注解 / 字段级语义完全对照；当前重点已经转到“控制面真相源回归 + 写侧副作用对齐”。
- 本轮进一步补上 Wasm `readme` 查询，以及 service scope plugin instances + delete 操作，优先满足 Java parity 第二批中最影响前端和 review 的接口缺口。
- 本轮继续补上 `ingressClass` 配置传递、MCP 显式 `routeMetadata`、以及 Route/Service/Wasm 首批字段级校验；builtin Wasm plugin 已支持从 legacy spec/readme 回退加载。
- 最近新增 aftercare 第二批：
  - `routes / ai-routes / ai-providers` 读链已切到真实 `Ingress / ConfigMap / WasmPlugin`
  - `mcp-servers` 已进入 `Ingress + higress-config + McpBridge/registry + route auth/plugin instance` 联合模型
  - `ai-routes / ai-providers` 已补首批 runtime 写路径，但仍需继续对齐 Java 的完整 save strategy 与 delete cleanup
  - `P3-CP-01` 已完成：`mcp-servers` 真实读链已切到 `Ingress + higress-config + McpBridge/registry + route auth/plugin instance`
  - `P3-CP-05` 已完成：Portal `GetAgentCatalogOptions / inspectMcpServer / ListAIQuotaRoutes` 与 Jobs `reconcileMCPAuth` 已消费真实 `mcp-servers / ai-routes / ai-providers`
  - `P3-CP-02` 已完成：`MCP save strategy` 的 `consumer level` 注解回写、`DIRECT/REDIRECT_ROUTE` 的 `transportType / rewrite host / sse_path_suffix` 恢复、`pathRewritePrefix=/` 语义、`OPEN_API / DATABASE` 异常分支、`default-mcp / route auth` 历史兼容清理、以及 `match_list / servers[]` 保留未知字段的 Java 更新语义都已补齐
  - `P3-CP-04` 本轮继续补齐 `provider` 字段级 parity：`openaiExtraCustomUrls` 多 IP static registry、`vertex-auth.internal` extra service source、`ai-route` 运行时 service 引用改为跟随 provider 实际 `dns/static/custom service`，并补 `openai/azure/qwen/zhipuai/claude` 的 endpoint 规则与 URL/域名严格校验，以及 `claude/qwen/zhipuai/vertex/bedrock/ollama` 的保存前归一化
  - `P3-CP-06` 已在 `minikube / aigateway-system` 完成 smoke：临时 `provider / ai-route / mcp-server` 的 create/update/delete/cleanup 已通过，同时修复了 live cluster 下 `WasmPlugin update resourceVersion` 与 `MCP higress-config` 真相源两处问题
