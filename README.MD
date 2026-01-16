# OKTalk 后端服务

基于 Go 和 Gin 框架开发的 AI 语音交互英语学习系统后端。

## 技术栈

- **语言**: Go 1.24+
- **Web 框架**: Gin
- **数据库**: MySQL (GORM)
- **AI 服务**:
  - 语音识别: 阿里云 Fun-ASR
  - 大语言模型: 通义千问 Qwen3-Max
  - 语音合成: 阿里云 CosyVoice

## 项目结构

```
OKtalk/
├── cmd/
│   └── server/          # 主程序入口
│       └── main.go
├── internal/            # 内部代码（不对外暴露）
│   ├── handlers/        # HTTP 处理器
│   ├── services/        # 业务逻辑层
│   └── models/          # 数据模型
├── api/                 # API 路由定义
│   └── routes.go
├── pkg/                 # 公共包（可被外部引用）
├── config/              # 配置管理
│   └── config.go
├── go.mod
├── go.sum
└── README.MD
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 运行服务

```bash
go run cmd/server/main.go
```

服务默认运行在 `:8080` 端口。

### 3. 测试接口

```bash
# 测试健康检查接口
curl http://localhost:8080/ping
```

预期响应：
```json
{
  "message": "pong",
  "status": "ok"
}
```

## 环境变量

- `PORT`: 服务端口（默认: 8080）

## 开发规范

- 遵循 Go 标准项目布局
- 代码组织清晰，职责分离
- API 接口遵循 RESTful 规范
- 错误处理完善，返回明确的错误信息

## 核心功能（规划中）

1. **AI 语音对话** - 支持儿童英语对话交互
2. **AI 发音纠正** - 提供发音评分和反馈
3. **智能学习报告** - 生成阶段性学习报告

## License

本项目为毕业设计项目。
