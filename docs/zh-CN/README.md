# OwlVigil Go SDK 中文文档

这套文档面向在 Go 应用中接入 OwlVigil 的开发者。内容以当前仓库代码为准；字段、方法签名和返回类型以 [pkg.go.dev](https://pkg.go.dev/github.com/Syrovex/owlvigil_sdk_go) 为最终参考。

## 推荐阅读路径

第一次接入时，建议依次阅读第 1～4 章；已经完成基础接入时，可以直接进入对应的任务章节。

## 第一章：入门

1. [快速开始](01-quickstart.md)：安装 SDK，完成第一次 Gateway 请求。
2. [核心概念](02-concepts.md)：理解客户端、凭证、工作区、请求生命周期和返回值。
3. [认证与客户端配置](03-authentication-configuration.md)：区分两类 Key，并配置环境、超时、重试和 HTTP Client。

## 第二章：模型调用

4. [模型调用与流式响应](04-gateway.md)：完成模型发现、普通调用以及流式响应处理。

## 第三章：Management 管理

5. [Management 与工作区](05-management.md)：了解通用调用规则、工作区和分页。
6. [成员、权限与 Gateway Key](06-management-access-and-keys.md)：管理人员权限和模型调用密钥。
7. [路由、可观测性与账单](07-management-operations.md)：管理模型路由，查看运行数据，并处理财务与账单任务。

## 第四章：Webhook 与问题处理

8. [Webhook](08-webhooks.md)：验签、接收地址管理、事件历史和重复投递。
9. [错误处理与故障排查](09-errors-troubleshooting.md)：识别错误类型并按症状定位问题。

## 第五章：参考资料

10. [API 参考与示例索引](10-reference-examples.md)：查找公开能力并选择可运行示例。

版本变化和兼容性说明见 `CHANGELOG.md`。

## 接入检查

- 为非流式调用设置 HTTP 超时，并通过 Context 控制每次请求的生命周期。
- 流式调用结束后关闭 Stream，并检查 `stream.Err()`。
- 分开保存 Management API Key 与 Gateway Key，日志中不得记录任何完整凭证。
- 只对文档明确支持幂等键的写操作进行自动重试；其他写操作超时后先查询结果。
- Webhook 接收端必须先验证签名，并根据事件 ID 防止重复处理。
- 上线前运行 `go test ./...` 和 `sh scripts/check-docs.sh`。

## SDK 的公开包

| 包 | 用途 |
| --- | --- |
| `owlvigil` | 共享配置、请求选项、错误和响应元数据 |
| `gateway` | 模型列表、Chat Completions、Responses、Embeddings、Anthropic 兼容调用和流式响应 |
| `management` | 工作区、成员、权限、Gateway Key、路由、Provider、财务、账单、日志和 Webhook 管理 |
| `webhook` | 入站 Webhook 签名验证 |

## 版本要求

- Go 1.25.12 或更高版本。
- SDK 当前版本见 `owlvigil.Version`。
- 应用应通过 Go Modules 固定 SDK 版本，并在升级前阅读 `CHANGELOG.md`。
