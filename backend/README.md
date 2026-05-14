# Backend

后端当前采用标准库 `net/http + database/sql`，并已接入 `MySQL` 仓储实现。

## 目录

```text
backend/
├─ cmd/
│  └─ api/
├─ internal/
│  ├─ app/
│  ├─ config/
│  ├─ handler/
│  ├─ model/
│  ├─ repository/
│  │  └─ memory/
│  └─ service/
├─ pkg/
│  └─ response/
└─ go.mod
```

## 当前实现

- `handler`：HTTP 路由、请求解析、响应输出
- `service`：业务规则、题目生成、回答评分占位逻辑
- `repository`：仓储接口与公共错误
- `repository/mysql`：MySQL 落库实现
- `repository/memory`：保留的内存实现，便于本地快速替换

## API

- `GET /healthz`
- `POST /api/v1/interview-plans`
- `POST /api/v1/interview-sessions`
- `GET /api/v1/interview-sessions/{id}`
- `POST /api/v1/interview-sessions/{id}/questions/next`
- `POST /api/v1/interview-answers`
- `GET /api/v1/interview-reports/{sessionId}`
- `GET /api/v1/history?userId=1`

## 运行

```bash
cd backend
go run ./cmd/api
```

默认端口：

- `APP_PORT=8080`

默认的本地 MySQL 配置：

- `MYSQL_HOST=127.0.0.1`
- `MYSQL_PORT=3306`
- `MYSQL_USER=root`
- `MYSQL_PASSWORD=6204222lL@`
- `MYSQL_DATABASE=ai_interview`

如果你想直接传完整 DSN，也支持：

- `MYSQL_DSN=...`

## 建议下一步

1. 增加 `resume`、`job_target` 相关 API
2. 引入鉴权和用户体系
3. 把当前规则式评分替换成真实 AI 评分链路
