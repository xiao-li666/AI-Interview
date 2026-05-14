# Page Map

## 1. 首页 `/`

用途：

- 展示产品定位
- 展示训练概览数据
- 提供快速进入配置页和历史页的入口

关联数据：

- `interview_sessions`
- `interview_reports`

## 2. 面试配置页 `/setup`

用途：

- 录入目标岗位、级别、轮次
- 上传简历或粘贴 JD
- 选择面试模式、难度和题量

关联数据：

- `candidate_profiles`
- `resumes`
- `job_targets`
- `interview_sessions`

## 3. 模拟面试页 `/interview`

用途：

- 展示 AI 当前题目
- 提供文本或语音回答入口
- 展示当前评分维度和追问逻辑

关联数据：

- `interview_sessions`
- `interview_questions`
- `interview_answers`
- `answer_feedbacks`

## 4. 复盘报告页 `/report`

用途：

- 展示整场面试综合得分
- 展示维度分和总结点评
- 给出下一轮训练建议

关联数据：

- `interview_reports`
- `report_dimension_scores`
- `answer_feedbacks`

## 5. 历史记录页 `/history`

用途：

- 按时间查看历史练习
- 支持后续扩展筛选、趋势图、专项对比

关联数据：

- `interview_sessions`
- `interview_reports`
