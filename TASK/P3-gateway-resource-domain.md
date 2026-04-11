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

## 验收点

- [x] K8s 资源 CRUD 不再局限于内存 fake client。
- [ ] K8s 资源 CRUD 与旧实现行为一致。
- [ ] 关键注解、ConfigMap、CRD 映射无缺项。
- [x] `/v1/routes`、`/v1/domains`、`/v1/mcp*`、`/v1/ai/*` 可 smoke。

## 测试

- [x] Fake client 单元测试。
- [x] gateway service 保护资源 / internal 资源 / plugin instance 行为测试。
- [ ] 资源转换测试。
- [x] 冲突/校验失败测试。
- [x] Route / MCP / AI Route 契约样本测试。

## 本轮说明

- 当前 P3 已从“HTTP 契约 + 内存版资源存储”推进到“HTTP 契约 + real/fake 双 client”，其中 real client 通过 `kubectl` 落实际 `ConfigMap/Secret` 持久化。
- 资源域仍未完成到旧 Java 的 CRD / 注解 / 字段级语义完全对照；当前重点是先把持久化入口、保护资源规则和插件实例链路收口。
