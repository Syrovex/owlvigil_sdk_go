# 成员、权限与 Gateway Key

## Management：成员与访问控制

访问控制由团队、成员、邀请、角色、权限以及成员级权限覆盖共同组成。先设计角色，再批量添加成员，能减少后续逐人修补权限。

### 资源关系

<!-- evidence: management/teams.go, management/members.go, management/rbac.go -->
```text
Workspace
├── Teams
├── Members
│   └── Member permission overrides
├── Invitations
└── Roles
    └── Permissions
```

### 推荐流程

1. 使用 `ListPermissions` 查看可授予权限。
2. 使用 `ListRoles` 查看已有角色；需要时调用 `CreateRole`。
3. 使用 `CreateTeam` 建立组织分组。
4. 使用 `CreateInvitation` 邀请尚未加入的用户，或使用 `CreateMember` 添加符合契约的成员。
5. 使用 `GetMemberPermissions` 验证最终权限。
6. 只有例外情况才使用 `UpdateMemberPermissions` 覆盖成员权限。

### 团队和成员

团队支持 list/create/get/update/delete，成员也提供对应生命周期方法。筛选成员时优先使用 `ListMembersWithFilters`，不要用未发布的 query 参数猜测服务端行为。

| 资源 | 方法 |
| --- | --- |
| 团队 | `ListTeams`, `CreateTeam`, `GetTeam`, `UpdateTeam`, `DeleteTeam`, `DeleteTeamWithResult` |
| 成员 | `ListMembers`, `ListMembersWithFilters`, `ListRoleOptions`, `CreateMember`, `GetMember`, `UpdateMember`, `DeleteMember`, `DeleteMemberWithResult` |

删除方法的 `WithResult` 版本会返回确认对象。删除前记录资源 ID 和 Request ID，并确认成员是否仍拥有关键资源或唯一管理员职责。

### 邀请

邀请支持 list/create/resend/revoke。重复发送和撤销都是写操作，且当前不支持幂等键自动重试；UI 应避免双击触发，不确定结果应通过列表或详情核对。邮箱属于个人信息，不应出现在普通诊断日志中。

对应方法为 `ListInvitations`、`CreateInvitation`、`ResendInvitation`、`ResendInvitationWithResult`、`RevokeInvitation` 和 `RevokeInvitationWithResult`。

### 角色与权限覆盖

角色适合表达大多数成员的稳定权限。成员覆盖适合临时或例外授权。使用 `ResetMemberPermissions` 可清除覆盖并恢复角色基线。

修改权限后，通过读取接口验证最终状态，不要只依据写请求没有报错就假设权限已符合预期。

完整方法包括 `ListRoles`、`CreateRole`、`GetRole`、`UpdateRole`、`DeleteRole`、`DeleteRoleWithResult`、`ListPermissions`、`GetMemberPermissions`、`UpdateMemberPermissions` 和 `ResetMemberPermissions`。

## Management：Gateway Key 管理

Gateway Key 用于 `gateway.Client` 的模型调用。Management API Key 负责管理 Gateway Key，但两者不能互相替代。

### 创建 Key

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
key, meta, err := client.CreateGatewayKey(
	ctx,
	&management.CreateGatewayKeyRequest{
		WorkspaceID: workspaceID,
		Name:        "production-app",
	},
	owlvigil.WithIdempotencyKey("gateway-key-production-app-v1"),
)
if err != nil {
	return err
}

```

创建响应中的 `Key` 或 `Secret` 字段可能只显示一次。它表示供 `gateway.Client` 调用模型时使用的 Gateway Key，不是 Management API Key。收到后应立即保存到密钥管理服务，不要打印完整的 `key` 对象。

### 生命周期

- `CreateGatewayKey`：创建新的模型调用密钥；创建时可以使用幂等键。
- `ListGatewayKeys` / `GetGatewayKey`：查询 Key 和脱敏信息。
- `UpdateGatewayKey`：修改名称、归属、限流、预算等可变字段。
- `RotateGatewayKey`：生成新 secret。
- `EnableGatewayKey` / `DisableGatewayKey`：控制使用状态。
- `DeleteGatewayKey`：删除资源。

需要确认对象时使用 `EnableGatewayKeyWithResult`、`DisableGatewayKeyWithResult` 或 `DeleteGatewayKeyWithResult`。

### 安全轮换

1. 调用 `RotateGatewayKey`，把返回的新 Gateway Key 直接保存到密钥管理服务。
2. 更新消费者并等待部署完成。
3. 用只读模型请求验证新值。
4. 确认旧值不再被使用后禁用或删除旧凭证。

轮换期间不要在日志、持续集成（CI）输出或命令行历史中暴露新旧 Gateway Key。只有创建操作支持幂等键；轮换请求超时且结果不确定时不要自动重试，应通过 `GetGatewayKey` 或 Request ID 核对结果。

### 最小权限

按应用、环境和责任边界分别创建 Key。不要让开发、预发布和生产共享同一个 secret。使用 Team、Assignee、限流和预算字段时，以当前 Go 类型及服务端权限为准。
