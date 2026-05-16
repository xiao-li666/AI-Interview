# 智能模拟面试系统

一个面向程序员面试训练场景的 AI 模拟面试项目。当前仓库已经包含完整的前后端基础能力：用户端、管理员端、Go REST API、MySQL 落库、简历解析、AI 模型配置、模拟面试流程和复盘报告。

## 项目目标

这个项目不是单纯的问答页面，而是想打通一条完整的模拟面试链路：

- 用户注册、登录、个人信息维护
- 配置岗位、轮次、难度、题量、公司与简历
- 结合简历内容和面试配置生成题目
- 逐题作答并拿到单题反馈
- 生成整场复盘报告和维度评分
- 按用户隔离保存历史记录、简历和 AI 配置
- 提供独立管理员后台做全局查看与用户管理

## 当前能力

### 用户端

- 登录 / 注册 / 退出登录
- 个人设置页，可修改昵称、头像
- 首页、面试配置、模拟面试、历史记录、复盘报告、模型配置
- 左侧分组导航与顶部标签导航
- 面试配置支持岗位信息、公司选择、简历上传与解析
- 支持 PDF / DOCX 简历解析
- 支持答案历史查看、报告复盘、题目展开查看答案建议

### 管理员端

- 独立管理员登录
- 全局概览页
- 用户管理：查看详情、启用 / 禁用用户、重置密码
- 面试记录查看：会话、题目、答案、报告
- 简历查看
- AI 配置查看

### 后端能力

- Go + `net/http` 的 REST API
- MySQL Repository 落库
- 用户 JWT 鉴权
- 管理员独立 JWT 鉴权
- 用户与管理员权限隔离
- 账号禁用后，中间件会实时校验状态，不再允许旧 token 继续访问
- DeepSeek / OpenAI provider 配置与测试
- AI 出题、AI 评分、AI 报告总结的 provider 抽象

## 技术栈

### 用户前端

- Vue 3
- Vue Router
- Vite

### 管理员前端

- Vue 3
- Vue Router
- Vite

### 后端

- Go 1.22+
- `net/http`
- `database/sql`
- `github.com/go-sql-driver/mysql`

### 数据库

- MySQL 8+

### AI 提供商

- DeepSeek
- OpenAI

## 目录结构

```text
AI_Interview/
├─ frontend/                 # 用户端 Vue3 应用
├─ admin-frontend/           # 管理员端 Vue3 应用
├─ backend/                  # Go 后端
│  ├─ cmd/api/               # 启动入口
│  ├─ internal/
│  │  ├─ ai/                 # AI provider 抽象与实现
│  │  ├─ app/                # 应用装配
│  │  ├─ config/             # 环境配置
│  │  ├─ handler/            # HTTP handler / middleware
│  │  ├─ model/              # 领域模型
│  │  ├─ repository/         # 仓储接口与 MySQL 实现
│  │  └─ service/            # 业务逻辑
│  └─ pkg/response/          # 统一响应结构
├─ database/
│  └─ schema.sql             # 数据库建表脚本
└─ docs/
   └─ page-map.md
```

## 核心页面

### 用户端

- `/`：首页
- `/setup`：面试配置
- `/interview`：模拟面试
- `/history`：历史记录
- `/report?sessionId=xx`：复盘报告
- `/ai-config`：模型配置
- `/settings`：个人设置
- `/login`：登录
- `/register`：注册

### 管理员端

管理员前端默认运行在 `5174` 端口：

- `/login`：管理员登录
- `/`：后台概览
- `/users`：用户管理
- `/sessions`：面试记录
- `/resumes`：简历查看
- `/ai-configs`：AI 配置查看

## API 概览

### 健康检查

- `GET /healthz`

### 用户认证

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `PUT /api/v1/auth/me`

### 面试主链路

- `POST /api/v1/interview-plans`
- `POST /api/v1/interview-sessions`
- `GET /api/v1/interview-sessions/{id}`
- `GET /api/v1/interview-sessions/{id}/answers`
- `POST /api/v1/interview-sessions/{id}/questions/next`
- `POST /api/v1/interview-answers`
- `GET /api/v1/interview-reports/{sessionId}`
- `GET /api/v1/history`

### 简历与模型配置

- `POST /api/v1/resume/parse`
- `GET /api/v1/ai-config`
- `PUT /api/v1/ai-config`
- `POST /api/v1/ai-config/test`

### 管理员接口

- `POST /api/v1/admin/auth/login`
- `POST /api/v1/admin/auth/logout`
- `GET /api/v1/admin/auth/me`
- `GET /api/v1/admin/overview`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/{id}`
- `PUT /api/v1/admin/users/{id}/status`
- `POST /api/v1/admin/users/{id}/reset-password`
- `GET /api/v1/admin/interview-sessions`
- `GET /api/v1/admin/interview-sessions/{id}`
- `GET /api/v1/admin/interview-sessions/{id}/answers`
- `GET /api/v1/admin/interview-reports/{sessionId}`
- `GET /api/v1/admin/resumes`
- `GET /api/v1/admin/resumes/{id}`
- `GET /api/v1/admin/ai-configs`

## 数据库设计

当前核心表包括：

- `users`
- `admins`
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

完整结构见 [database/schema.sql](/E:/codexCode/AItools/AI_Interview/database/schema.sql)。

## 本地启动

### 1. 初始化数据库

先创建数据库并执行建表脚本：

```bash
mysql -u root -p < database/schema.sql
```

### 2. 启动后端

```bash
cd backend
go run ./cmd/api
```

默认地址：

- `http://127.0.0.1:8080`

### 3. 启动用户前端

```bash
cd frontend
npm install
npm run dev
```

默认地址：

- `http://127.0.0.1:5173`

### 4. 启动管理员前端

```bash
cd admin-frontend
npm install
npm run dev
```

默认地址：

- `http://127.0.0.1:5174`

## 关键环境变量

### 后端

```bash
APP_PORT=8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your-password
MYSQL_DATABASE=ai_interview
MYSQL_DSN=

JWT_SECRET=ai-interview-dev-secret
JWT_EXPIRE_HOURS=168
DEFAULT_USER_STATUS=active

ADMIN_SEED_EMAIL=admin@ai-interview.local
ADMIN_SEED_PASSWORD=Admin@123456
ADMIN_SEED_NICKNAME=系统管理员

DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash

OPENAI_API_KEY=
OPENAI_BASE_URL=
OPENAI_MODEL=gpt-5.5
```

说明：

- 如果未设置 `MYSQL_DSN`，后端会自动使用 `MYSQL_HOST / PORT / USER / PASSWORD / DATABASE` 组装连接。
- `ADMIN_SEED_*` 用于初始化管理员账号。
- 用户和管理员使用独立 token，不共享登录态。

### 用户前端

```bash
VITE_API_BASE_URL=http://127.0.0.1:8080
```

### 管理员前端

```bash
VITE_API_BASE_URL=http://127.0.0.1:8080
```

## 推荐体验流程

### 用户端

1. 注册并登录
2. 在模型配置页填入 DeepSeek / OpenAI 参数并测试连接
3. 在面试配置页填写岗位信息、选择公司、上传简历
4. 创建模拟面试
5. 在模拟面试页逐题回答
6. 到历史记录中选择面试，进入复盘报告页查看总结

### 管理员端

1. 使用管理员账号登录后台
2. 查看全局数据概览
3. 按用户检索并查看用户详情
4. 必要时禁用账号或重置密码
5. 查看面试记录、答案、简历和 AI 配置

## 当前状态

目前仓库已经不是原型，而是一版可本地跑通的 MVP：

- 用户端可登录并完成模拟面试主流程
- 数据已经落到 MySQL
- 支持 AI 配置接入真实模型
- 支持独立管理员后台

后续更值得继续投入的方向：

- 语音输入 / 语音面试链路
- 公司题库沉淀与更强的出题策略
- 更细粒度的管理员权限与审计日志
- 部署脚本、容器化和 CI/CD
- 自动化测试补齐

## 相关文档

- [database/schema.sql](/E:/codexCode/AItools/AI_Interview/database/schema.sql)
- [docs/page-map.md](/E:/codexCode/AItools/AI_Interview/docs/page-map.md)
