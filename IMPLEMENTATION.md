# Hydralisk 实施进度

## 已完成的包

### 1. core - 核心类型和接口 ✅
- `types/` - 7个类型文件
- `interfaces/` - 7个接口文件
- `export.go` - 导出别名
- `go.mod`

### 2. event-bus - 事件总线 ✅
- `event_bus.go`
- `go.mod`

### 3. message-bus - 消息总线 ✅
- `message_bus.go`
- `go.mod`

### 4. process-manager - 进程管理器 ✅
- `process_manager.go`
- `go.mod`

### 5. llm-client - LLM客户端 ✅
- `llm_client.go`
- `go.mod`

### 6. db - 数据库层 ✅
- `repository.go`
- `go.mod`

## 待实现的包

### 7. tools - 工具集 (16个工具)
```
tools/
├── registry.go          # 工具注册表
├── read.go              # 读取文件
├── write.go             # 写入文件
├── edit.go              # 编辑文件
├── multiedit.go         # 多文件编辑
├── apply_patch.go       # 应用补丁
├── job.go               # 进程管理
├── glob.go              # 文件匹配 (ripgrep)
├── grep.go              # 内容搜索 (ripgrep)
├── ls.go                # 目录列表
├── question.go          # 用户提问
├── todo.go              # 待办管理
├── batch.go             # 批量执行
├── task.go              # 子Agent
├── message.go           # 消息发送
├── memory_search.go     # 内存搜索
└── memory_get.go        # 内存读取
```

### 8. service - 服务层
```
service/
├── service.go           # ServiceLayer
├── session.go           # SessionService
├── process.go           # ProcessService
├── tool.go              # ToolService
└── misc.go              # MiscService
```

### 9. api - API层 (connect-go)
```
api/
├── go.mod
├── handler.go           # Connect handlers
├── generated/           # 生成的代码
│   ├── agent.pb.go
│   └── agent.connect.go
```

### 10. cmd/hydralisk - 主程序 (cobra CLI)
```
cmd/hydralisk/
├── go.mod
├── main.go              # 入口
└── config.go            # 配置加载
```

### 11. proto - Proto文件
```
proto/
├── buf.yaml
├── buf.gen.yaml
└── agent.proto
```

## 技术栈确认

| 组件 | 技术 |
|------|------|
| ORM | uptrace/bun |
| API | connect-go |
| CLI | cobra |
| 数据库 | SQLite |
| 进程 | 原生 exec.Command |
| 搜索 | ripgrep |

## 文件结构对照

| Mutalisk | Hydralisk |
|----------|-----------|
| `packages/core/src/types/*.ts` | `packages/core/types/*.go` |
| `packages/core/src/interfaces/*.ts` | `packages/core/interfaces/*.go` |
| `packages/db/src/repository.ts` | `packages/db/repository.go` |
| `packages/event-bus/src/index.ts` | `packages/event-bus/event_bus.go` |
| `packages/llm-client/src/index.ts` | `packages/llm-client/llm_client.go` |
| `packages/process-manager/src/index.ts` | `packages/process-manager/process_manager.go` |
| `packages/message-bus/src/message-bus.ts` | `packages/message-bus/message_bus.go` |
| `packages/service/src/index.ts` | `packages/service/service.go` |
| `packages/tools/src/tools/*.ts` | `packages/tools/*.go` |
| `packages/api/src/connect.ts` | `packages/api/handler.go` |
| `src/main.ts` | `cmd/hydralisk/main.go` |

## 下一步

1. 实现 tools 包 (16个工具)
2. 实现 service 包
3. 创建 proto 文件并生成代码
4. 实现 api 包
5. 实现 cmd/hydralisk 主程序
6. 创建 nix flake
7. 测试编译