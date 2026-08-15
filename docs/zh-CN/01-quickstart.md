# 快速开始：完成第一次模型调用

本教程会创建一个最小 Go 程序，列出当前 Gateway Key 可访问的模型。完成后，你会得到一个可以继续扩展为 Chat 或流式调用的客户端。

## 准备工作

- Go 1.25.12 或更高版本。
- 一个 OwlVigil Gateway Key。
- 一个启用了 Go Modules 的项目。

不要把真实密钥写进源码、提交到 Git，或粘贴到 Issue。

## 第一步：安装 SDK

在你的 Go 项目中运行：

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go get github.com/Syrovex/owlvigil_sdk_go
```

把 Gateway Key 放到环境变量中：

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
export OWLVIGIL_GATEWAY_KEY='your-gateway-key'
```

`your-gateway-key` 是占位符，请替换为真实值，但不要把真实命令保存到 shell 脚本或文档。

## 第二步：创建可运行程序

创建 `main.go`：

<!-- evidence: examples/quickstart/main.go, scripts/check-docs-api/main_test.go -->
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/gateway"
)

func main() {
	if err := run(); err != nil {
		slog.Error("OwlVigil quickstart failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	key := os.Getenv("OWLVIGIL_GATEWAY_KEY")
	if key == "" {
		return errors.New("OWLVIGIL_GATEWAY_KEY is required")
	}

	client := gateway.NewClient(
		owlvigil.WithAPIKey(key),
		owlvigil.WithTimeout(30*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	models, meta, err := client.ListModels(ctx)
	if err != nil {
		var apiErr *owlvigil.APIError
		if errors.As(err, &apiErr) {
			return fmt.Errorf("list Gateway models: status=%d code=%s request_id=%s message=%s",
				apiErr.StatusCode, apiErr.Code, apiErr.RequestID, apiErr.Message)
		}
		return fmt.Errorf("list Gateway models: %w", err)
	}

	fmt.Printf("request_id=%s models=%d\n", meta.RequestID, len(models.Data))
	for _, model := range models.Data {
		fmt.Println(model.ID)
	}
	return nil
}
```

仓库中的 [`examples/quickstart`](../../examples/quickstart/) 使用同一套调用方式，并由 `go test ./examples/...` 持续编译验证。

## 第三步：运行并验证

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go run .
```

成功时，程序会输出请求 ID、模型数量和模型 ID，例如：

<!-- evidence: examples/quickstart/main.go -->
```text
request_id=req_... models=2
model-a
model-b
```

模型列表由你的工作区和 Gateway Key 决定，不要依赖示例中的名称。保存 `request_id` 有助于排查服务端请求。

## 下一步：发起 Chat 请求

从上一步输出中选择一个模型 ID：

<!-- evidence: examples/quickstart/main.go, scripts/check-docs-api/main_test.go -->
```go
resp, meta, err := client.CreateChatCompletion(ctx, &gateway.ChatCompletionRequest{
	Model: models.Data[0].ID,
	Messages: []gateway.Message{
		{Role: "user", Content: "用一句话介绍 OwlVigil。"},
	},
})
if err != nil {
	return err
}
if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
	return fmt.Errorf("empty chat response: request_id=%s", meta.RequestID)
}
fmt.Println(resp.Choices[0].Message.Content)

```

生产代码不要直接假设 `Choices[0]` 一定存在；上面的检查用于避免空响应导致 panic。

## 使用 Management 客户端

Management 使用独立的 API Key，不能拿 Gateway Key 代替：

<!-- evidence: examples/quickstart/main.go, scripts/check-docs-api/main_test.go -->
```go
client := management.NewClient(
	owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)

```

继续阅读 [Management](05-management.md#management) 选择具体工作流。

## 常见失败

- `401`：凭证缺失、失效，或把 Management API Key 与 Gateway Key 混用了。
- `403`：凭证有效，但没有所需工作区或权限范围。
- `context deadline exceeded`：调用超过应用上下文或 HTTP Client 的超时。
- `429`：服务限流。SDK 不会自动重试；调用方应遵循服务端等待提示。不要重放模型生成请求。

继续阅读[核心概念](02-concepts.md)，理解客户端、Context、返回值和幂等性。更多失败处理方式见[错误处理](09-errors-troubleshooting.md#错误处理)和[故障排查](09-errors-troubleshooting.md#故障排查)。
