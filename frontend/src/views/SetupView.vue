<template>
  <div class="page-grid two-column-layout">
    <section class="panel">
      <div class="panel-header">
        <div>
          <h3>面试配置</h3>
          <p>先配置岗位、目标公司与简历，系统会结合这些信息生成更贴近真实场景的题目。</p>
        </div>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <form class="config-grid" @submit.prevent="handleSubmit">
        <article class="config-card">
          <h4>岗位信息</h4>
          <div class="field-list">
            <label class="field">
              <span>岗位名称</span>
              <input v-model="form.jobTitle" type="text" placeholder="例如：Go 后端工程师" />
            </label>
            <label class="field">
              <span>岗位类别</span>
              <input v-model="form.jobCategory" type="text" placeholder="例如：backend" />
            </label>
            <label class="field">
              <span>岗位级别</span>
              <input v-model="form.levelCode" type="text" placeholder="例如：mid / senior" />
            </label>
            <div class="field">
              <span>目标公司</span>
              <div class="inline-field">
                <select v-model="form.companyName">
                  <option v-for="company in state.companyOptions" :key="company" :value="company">
                    {{ company }}
                  </option>
                </select>
                <button type="button" class="ghost-button mini-button" @click="showAddCompany = !showAddCompany">
                  添加
                </button>
              </div>
            </div>
            <div v-if="showAddCompany" class="field field-inline-card">
              <span>新增公司</span>
              <div class="inline-field">
                <input
                  v-model="newCompanyName"
                  type="text"
                  placeholder="输入公司名称"
                  @keydown.enter.prevent="handleAddCompany"
                />
                <button type="button" class="primary-button mini-button" @click="handleAddCompany">
                  保存
                </button>
              </div>
              <p class="helper-text">公司名称不能重复，保存后会加入下拉列表。</p>
            </div>
          </div>
        </article>

        <article class="config-card">
          <h4>面试设置</h4>
          <div class="field-list">
            <label class="field">
              <span>会话名称</span>
              <input v-model="form.sessionName" type="text" />
            </label>
            <label class="field">
              <span>面试轮次</span>
              <select v-model="form.roundType">
                <option v-for="item in roundTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>题目类型</span>
              <select v-model="form.interviewType">
                <option v-for="item in roundTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>面试模式</span>
              <select v-model="form.interviewMode">
                <option v-for="item in interviewModeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
          </div>
        </article>

        <article class="config-card">
          <h4>难度与风格</h4>
          <div class="field-list">
            <label class="field">
              <span>难度等级</span>
              <select v-model="form.difficultyLevel">
                <option v-for="item in difficultyOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>面试官风格</span>
              <select v-model="form.interviewerStyle">
                <option
                  v-for="item in interviewerStyleOptions"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>题目数量</span>
              <input v-model.number="form.questionCount" type="number" min="1" max="8" />
            </label>
          </div>
        </article>

        <article class="config-card config-card-full">
          <h4>简历内容</h4>
          <div class="field-list">
            <div class="field">
              <span>上传简历</span>
              <label class="upload-box">
                <input
                  class="upload-input"
                  type="file"
                  accept=".pdf,.docx,.txt,.md,.json"
                  @change="handleResumeUpload"
                />
                <strong>{{ uploadLabel }}</strong>
                <span>支持 PDF、DOCX、TXT、MD、JSON，解析后的文本会自动用于 AI 出题。</span>
              </label>
            </div>
            <label class="field">
              <span>简历文本</span>
              <textarea
                v-model="form.resumeText"
                rows="10"
                placeholder="也可以直接粘贴简历内容，建议保留项目经历、技术栈、业绩结果等关键信息。"
              />
            </label>
          </div>
        </article>
      </form>
    </section>

    <section class="panel slim-panel">
      <div class="panel-header">
        <h3>配置预览</h3>
      </div>

      <div class="summary-list">
        <div class="summary-row">
          <span>岗位名称</span>
          <strong>{{ form.jobTitle || "--" }}</strong>
        </div>
        <div class="summary-row">
          <span>目标公司</span>
          <strong>{{ form.companyName || "--" }}</strong>
        </div>
        <div class="summary-row">
          <span>面试轮次</span>
          <strong>{{ toRoundTypeLabel(form.roundType) }}</strong>
        </div>
        <div class="summary-row">
          <span>难度等级</span>
          <strong>{{ difficultyText }}</strong>
        </div>
        <div class="summary-row">
          <span>题目数量</span>
          <strong>{{ form.questionCount }}</strong>
        </div>
        <div class="summary-row">
          <span>简历状态</span>
          <strong>{{ resumeStatusText }}</strong>
        </div>
      </div>

      <div class="sidebar-card">
        <p class="sidebar-card-title">AI 出题依据</p>
        <p class="sidebar-card-text">岗位要求、面试轮次、简历经历、目标公司近年题型偏好。</p>
      </div>

      <button class="primary-button button-block" :disabled="state.loading.setup" @click="handleSubmit">
        {{ state.loading.setup ? "创建中..." : "创建并开始面试" }}
      </button>
    </section>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from "vue";
import { useRouter } from "vue-router";
import {
  companyOptions,
  difficultyOptions,
  interviewModeOptions,
  interviewerStyleOptions,
  roundTypeOptions,
  toRoundTypeLabel
} from "../data/mock";
import { api } from "../lib/api";
import { useInterviewState } from "../state/interview";

const router = useRouter();
const { state, updateConfig, createInterviewFlow, setError, addCompanyOption } = useInterviewState();

const form = reactive({
  ...state.config
});

const showAddCompany = ref(false);
const newCompanyName = ref("");
const parsingResume = ref(false);

const difficultyText = computed(
  () => difficultyOptions.find((item) => item.value === form.difficultyLevel)?.label || "--"
);

const resumeStatusText = computed(() => {
  if (form.resumeText?.trim()) {
    return form.resumeFileName ? `已上传 ${form.resumeFileName}` : "已填写简历文本";
  }

  return "未上传";
});

const uploadLabel = computed(() => {
  if (parsingResume.value) {
    return "简历解析中...";
  }

  return form.resumeFileName || "点击上传 PDF / DOCX / TXT / MD / JSON 简历";
});

watch(
  form,
  (value) => {
    updateConfig(value);
  },
  { deep: true }
);

if (!state.companyOptions?.length) {
  state.companyOptions = [...companyOptions];
}

async function handleSubmit() {
  setError("");

  if (!form.companyName?.trim()) {
    setError("请先选择目标公司");
    return;
  }

  if (!form.resumeText?.trim()) {
    setError("请先上传或填写简历内容");
    return;
  }

  try {
    const detail = await createInterviewFlow();
    router.push({
      path: "/interview",
      query: {
        sessionId: String(detail.session.id)
      }
    });
  } catch {
    // state.error already set in store
  }
}

async function handleResumeUpload(event) {
  const file = event.target.files?.[0];
  if (!file) {
    return;
  }

  try {
    parsingResume.value = true;
    setError("");
    const result = await api.parseResume(file);
    form.resumeFileName = result.fileName || file.name;
    form.resumeText = (result.text || "").trim();
  } catch (error) {
    setError(error.message || "简历读取失败，请检查文件格式");
  } finally {
    parsingResume.value = false;
    event.target.value = "";
  }
}

function handleAddCompany() {
  try {
    addCompanyOption(newCompanyName.value);
    form.companyName = newCompanyName.value.trim();
    newCompanyName.value = "";
    showAddCompany.value = false;
    setError("");
  } catch (error) {
    setError(error.message);
  }
}
</script>
