export const navItems = [
  { label: "首页", to: "/" },
  { label: "面试配置", to: "/setup" },
  { label: "模拟面试", to: "/interview" },
  { label: "复盘报告", to: "/report" },
  { label: "历史记录", to: "/history" }
];

export const interviewModeOptions = [{ label: "文字面试", value: "text" }];

export const interviewerStyleOptions = [
  { label: "均衡型", value: "balanced" },
  { label: "压力型", value: "strict" },
  { label: "鼓励型", value: "friendly" }
];

export const difficultyOptions = [
  { label: "简单", value: "easy" },
  { label: "中等", value: "medium" },
  { label: "困难", value: "hard" }
];

export const roundTypeOptions = [
  { label: "技术一面", value: "technical_1" },
  { label: "技术二面", value: "technical_2" },
  { label: "项目深挖", value: "project_deep_dive" },
  { label: "行为面试", value: "behavior" }
];

export const companyOptions = [
  "阿里巴巴",
  "腾讯",
  "字节跳动",
  "百度",
  "美团",
  "京东",
  "拼多多",
  "快手",
  "小米",
  "网易"
];

export const setupDefaults = {
  jobTitle: "Go 后端工程师",
  jobCategory: "backend",
  levelCode: "mid",
  interviewType: "technical_1",
  interviewMode: "text",
  difficultyLevel: "medium",
  questionCount: 3,
  interviewerStyle: "balanced",
  sessionName: "Go 后端模拟面试",
  roundType: "technical_1",
  companyName: "阿里巴巴",
  resumeText: "",
  resumeFileName: ""
};

export const scoreWeights = [
  { label: "技术准确性", value: "30%" },
  { label: "表达清晰度", value: "20%" },
  { label: "项目深度", value: "30%" },
  { label: "追问应对", value: "20%" }
];

const statusLabelMap = {
  draft: "待开始",
  in_progress: "进行中",
  completed: "已完成"
};

const roundTypeLabelMap = {
  technical_1: "技术一面",
  technical_2: "技术二面",
  project_deep_dive: "项目深挖",
  behavior: "行为面试"
};

const questionTypeLabelMap = {
  self_intro: "自我介绍",
  project: "项目题",
  system_design: "系统设计",
  go_runtime: "Go 基础",
  behavior: "行为面试"
};

const dimensionLabelMap = {
  communication: "表达沟通",
  depth: "项目深度",
  accuracy: "技术准确性",
  clarity: "结构清晰度"
};

export function toStatusLabel(value) {
  return statusLabelMap[value] || value || "--";
}

export function toRoundTypeLabel(value) {
  return roundTypeLabelMap[value] || value || "--";
}

export function toQuestionTypeLabel(value) {
  return questionTypeLabelMap[value] || value || "--";
}

export function toDimensionLabel(value) {
  return dimensionLabelMap[value] || value || "--";
}
