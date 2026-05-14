# AI Interview

`AI Interview` 是一个面向后端岗位训练场景的 AI 模拟面试项目，当前采用 `Vue 3 + Vite` 作为前端、`Go + net/http + MySQL` 作为后端。

项目目标不是只做一个问答页，而是打通一条完整链路：

- 配置目标岗位与面试参数
- 生成整场面试题目
- 逐题回答与即时反馈
- 自动汇总整场复盘报告
- 保存历史记录，支持后续复练
- 配置自己的 AI 提供商（当前支持 `DeepSeek` 和 `OpenAI`）

## 当前状态

目前仓库已经不是纯页面原型，前后端主链路已经具备基础联调能力：

- 前端已实现首页、面试配置、模拟面试、复盘报告、历史记录、模型配置页面
- 前端已经接入真实后端接口，不再只是静态页面展示
- 后端已接入 MySQL 仓储，不再只使用内存数据
- 后端支持面试计划创建、会话创建、取题、答题、报告生成、历史查询、AI 配置保存与测试

当前也有几个需要明确的限制：

- 还没有登录鉴权，接口默认按 `userId` 维度读写
- 简历上传、JD 解析、岗位管理还没有完整落地
- 语音链路还没有接入，当前以文本回答为主
- AI 评分和报告链路已接入提供商抽象，但整体仍处于 MVP 阶段

## 技术栈

- 前端：`Vue 3`、`Vue Router`、`Vite`
- 后端：`Go 1.22+`、标准库 `net/http`、`database/sql`
- 数据库：`MySQL 8+`
- AI 提供商：`DeepSeek`、`OpenAI`

## 目录结构

```text
AI_Interview/
├─ frontend/                # Vue 3 前端
│  ├─ src/
│  │  ├─ assets/            # 样式
│  │  ├─ components/        # 通用组件
│  │  ├─ data/              # 默认配置与占位数据
│  │  ├─ lib/               # API 请求封装
│  │  ├─ router/            # 路由
│  │  ├─ state/             # 前端状态管理
│  │  └─ views/             # 页面视图
│  ├─ package.json
│  └─ vite.config.js
├─ backend/                 # Go API 服务
│  ├─ cmd/api/              # 服务启动入口
│  ├─ internal/
│  │  ├─ ai/                # AI provider 抽象及实现
│  │  ├─ app/               # 应用装配
│  │  ├─ config/            # 环境配置读取
│  │  ├─ handler/           # HTTP 路由与 handler
│  │  ├─ model/             # 领域模型
│  │  ├─ repository/        # 仓储接口与 MySQL 实现
│  │  └─ service/           # 业务逻辑
│  ├─ pkg/response/         # 统一响应
│  └─ go.mod
├─ database/
│  ├─ schema.sql            # 数据库建表脚本
│  └─ README.md
├─ docs/
│  └─ page-map.md           # 页面与数据表映射说明
└─ README.md
```

## 核心页面

- `/`：首页，展示项目定位、健康状态和历史概览
- `/setup`：面试配置页，创建计划与会话
- `/interview`：模拟面试页，逐题作答并获取反馈
- `/report`：整场复盘报告页
- `/history`：历史练习记录页
- `/ai-config`：AI 提供商配置页

## 核心接口

### 健康检查

- `GET /healthz`

### 面试主链路

- `POST /api/v1/interview-plans`
- `POST /api/v1/interview-sessions`
- `GET /api/v1/interview-sessions/{id}`
- `GET /api/v1/interview-sessions/{id}/answers`
- `POST /api/v1/interview-sessions/{id}/questions/next`
- `POST /api/v1/interview-answers`
- `GET /api/v1/interview-reports/{sessionId}`
- `GET /api/v1/history?userId=1`

### AI 提供商配置

- `GET /api/v1/ai-config?userId=1`
- `PUT /api/v1/ai-config`
- `POST /api/v1/ai-config/test`

## 环境要求

本地运行建议准备下面这些环境：

- `Node.js 18+`
- `npm 9+`
- `Go 1.22+`
- `MySQL 8+`

## 数据库初始化

先创建数据库并导入表结构：

```bash
mysql -u root -p < database/schema.sql
```

如果你是手动执行，也可以直接打开 [database/schema.sql](E:\codexCode\AItools\AI_Interview\database\schema.sql) 在 MySQL 客户端中运行。

## 后端启动

进入后端目录：

```bash
cd backend
go run ./cmd/api
```

默认启动后：

- API 地址：`http://127.0.0.1:8080`

### 后端环境变量

后端支持以下环境变量：

```bash
APP_PORT=8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your-password
MYSQL_DATABASE=ai_interview

# 可选：如果你希望直接传完整 DSN
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/ai_interview?parseTime=true&charset=utf8mb4

# DeepSeek
DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash

# OpenAI
OPENAI_API_KEY=
OPENAI_BASE_URL=
OPENAI_MODEL=gpt-5.5
```

说明：

- 若未设置 `MYSQL_DSN`，后端会使用 `MYSQL_HOST / PORT / USER / PASSWORD / DATABASE` 组装连接串
- 若未配置 AI Key，默认 AI 能力可能无法正常调用真实模型
- 当前代码中存在本地默认 MySQL 密码回退值，实际使用时建议显式通过环境变量覆盖，不要依赖默认值

## 前端启动

进入前端目录：

```bash
cd frontend
npm install
npm run dev
```

默认启动后：

- 前端地址：`http://127.0.0.1:5173`

### 前端环境变量

前端当前主要使用一个环境变量：

```bash
VITE_API_BASE_URL=http://127.0.0.1:8080
```

如果不传，前端默认也会请求 `http://127.0.0.1:8080`。

## 本地联调流程

推荐按这个顺序启动：

1. 启动 MySQL，并执行 `database/schema.sql`
2. 启动后端：`cd backend && go run ./cmd/api`
3. 启动前端：`cd frontend && npm install && npm run dev`
4. 打开 `http://127.0.0.1:5173`
5. 先进入 `/ai-config` 配置自己的模型接口
6. 再进入 `/setup` 创建一场面试，继续到 `/interview` 完成答题与复盘

## 典型业务流程

当前版本的大致交互顺序如下：

1. 在前端配置岗位、轮次、难度、题量等参数
2. 前端调用 `POST /api/v1/interview-plans` 创建面试计划
3. 前端调用 `POST /api/v1/interview-sessions` 创建会话与初始题集
4. 面试过程中调用 `POST /api/v1/interview-sessions/{id}/questions/next` 逐题取题
5. 用户提交回答后调用 `POST /api/v1/interview-answers`
6. 会话完成后调用 `GET /api/v1/interview-reports/{sessionId}` 拉取整场复盘
7. 首页和历史页通过 `GET /api/v1/history` 展示练习记录

## 数据表概览

核心业务表包括：

- `users`
- `candidate_profiles`
- `resumes`
- `job_targets`
- `interview_plans`
- `interview_sessions`
- `interview_questions`
- `interview_answers`
- `answer_feedbacks`
- `interview_reports`
- `report_dimension_scores`
- `user_ai_provider_configs`

完整结构见 [database/schema.sql](E:\codexCode\AItools\AI_Interview\database\schema.sql)。

## 当前完成度判断

如果从“能不能本地跑通 MVP 主流程”这个标准看，目前状态可以概括为：

- 前端页面骨架：已完成
- 前端与后端接口接入：已完成基础联调
- 后端与 MySQL 落库：已接通
- AI 配置能力：已支持基础配置与测试
- 产品完整度：还没到正式可用版本，仍然是持续迭代中的 MVP

## 后续建议

比较值得优先推进的几件事：

1. 增加用户登录和鉴权，去掉前端手填 `userId`
2. 落地简历上传、JD 管理、岗位管理
3. 为前后端补齐 `.env.example`
4. 为关键 API 增加集成测试
5. 增加部署说明与生产环境配置
6. 优化 AI 调用失败时的降级和提示策略

## 相关文档

- [backend/README.md](E:\codexCode\AItools\AI_Interview\backend\README.md)
- [database/README.md](E:\codexCode\AItools\AI_Interview\database\README.md)
- [docs/page-map.md](E:\codexCode\AItools\AI_Interview\docs\page-map.md)
