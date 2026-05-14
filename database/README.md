# Database Design

这套表设计现在以 `MySQL 8` 为目标数据库，覆盖 AI 模拟面试的核心流程：

1. 用户维护个人档案和简历
2. 用户创建目标岗位和 JD
3. 系统生成一场面试会话
4. AI 连续出题，用户逐题作答
5. 系统对回答进行点评并汇总报告
6. 用户在历史记录里查看成长轨迹

## 核心实体

- `users`：用户
- `candidate_profiles`：候选人画像
- `resumes`：简历资产
- `job_targets`：目标岗位与 JD
- `interview_sessions`：面试会话
- `interview_questions`：AI 题目
- `interview_answers`：用户回答
- `answer_feedbacks`：单题点评
- `interview_reports`：整场面试复盘
- `report_dimension_scores`：分维度评分

## MySQL 设计约束

- 主键统一使用 `BIGINT AUTO_INCREMENT`
- 结构化扩展字段统一使用 `JSON`
- 全部业务表默认使用 `InnoDB`
- 字符集统一使用 `utf8mb4`

## 设计原则

- 优先保证一场面试从配置到复盘的链路完整
- 区分“会话级”和“题目级”数据，便于报告和追问
- 保留 JSON 扩展字段，便于后续接 AI 生成内容
