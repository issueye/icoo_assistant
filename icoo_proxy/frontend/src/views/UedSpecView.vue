<template>
  <section class="page-section">
    <div class="page-header">
      <h2 class="page-title">UED 组件</h2>
      <p class="page-description">
        参考 Ant Design 的中后台设计语言：清晰层级、32px 基础控件高度、6px 圆角、蓝色主操作、轻量边框和明确状态反馈。
      </p>
    </div>

    <PanelBlock title="设计 Token">
      <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <div class="sub-card">
          <p class="text-sm font-medium text-slate-900">主色</p>
          <div class="mt-3 flex items-center gap-2">
            <span class="h-6 w-10 rounded bg-[#1677ff]"></span>
            <span class="font-mono text-xs text-slate-600">#1677ff</span>
          </div>
        </div>
        <div class="sub-card">
          <p class="text-sm font-medium text-slate-900">圆角</p>
          <p class="mt-3 text-sm text-slate-600">基础 6px，面板 8px，标签 4px。</p>
        </div>
        <div class="sub-card">
          <p class="text-sm font-medium text-slate-900">控件高度</p>
          <p class="mt-3 text-sm text-slate-600">XS 24 / SM 28 / MD 32 / LG 40。</p>
        </div>
        <div class="sub-card">
          <p class="text-sm font-medium text-slate-900">状态色</p>
          <div class="mt-3 flex flex-wrap gap-2">
            <UTag variant="success">Success</UTag>
            <UTag variant="warning">Warning</UTag>
            <UTag variant="error">Error</UTag>
            <UTag variant="info">Info</UTag>
          </div>
        </div>
      </div>
    </PanelBlock>

    <div class="section-grid xl:grid-cols-2">
      <PanelBlock title="按钮">
        <div class="space-y-3">
          <div class="flex flex-wrap gap-2">
            <UButton variant="primary">Primary</UButton>
            <UButton variant="success">Success</UButton>
            <UButton variant="warning">Warning</UButton>
            <UButton variant="error">Error</UButton>
            <UButton variant="info">Info</UButton>
            <UButton variant="secondary">Secondary</UButton>
            <UButton variant="ghost">Ghost</UButton>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <UButton size="xs">XS</UButton>
            <UButton size="sm">SM</UButton>
            <UButton size="md">MD</UButton>
            <UButton size="lg">LG</UButton>
            <UButton loading>加载中</UButton>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="开关">
        <div class="space-y-3">
          <USwitch v-model="switchValue" label="启用自动健康检查" hint="配置类开关用于开关型参数，文案保持动宾结构。" />
          <USwitch :model-value="true" label="保留系统默认路由" hint="禁用态示例" :disabled="true" />
        </div>
      </PanelBlock>
    </div>

    <div class="section-grid xl:grid-cols-2">
      <PanelBlock title="标签">
        <div class="space-y-3">
          <div class="flex flex-wrap gap-2">
            <UTag variant="primary">primary</UTag>
            <UTag variant="success">success</UTag>
            <UTag variant="warning">warning</UTag>
            <UTag variant="error">error</UTag>
            <UTag variant="info">info</UTag>
            <UTag>neutral</UTag>
            <UTag code>openai-responses</UTag>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <UTag size="xs" variant="primary">xs</UTag>
            <UTag size="sm" variant="success">sm</UTag>
            <UTag size="md" variant="warning">md</UTag>
            <UTag size="lg" variant="error">lg</UTag>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="输入与下拉">
        <div class="space-y-3">
          <UInput v-model="form.name" label="名称" placeholder="请输入供应商名称" hint="表单项采用上 label、下控件布局。" />
          <USelect v-model="form.protocol" label="协议" :options="protocolOptions" />
          <UInput v-model="form.description" label="描述" placeholder="请输入用途说明" textarea />
        </div>
      </PanelBlock>

      <PanelBlock title="弹窗">
        <div class="flex flex-wrap gap-2">
          <UButton @click="showModal = true">普通弹窗</UButton>
          <UButton variant="error" @click="showConfirm = true">确认弹窗</UButton>
        </div>
      </PanelBlock>
    </div>

    <PanelBlock title="基础表格">
      <UTable :columns="columns" :rows="rows" row-key="id" empty-text="暂无组件示例数据。">
        <template #cell-status="{ value }">
          <UTag :variant="value === '启用' ? 'success' : 'error'">{{ value }}</UTag>
        </template>
        <template #actions="{ row }">
          <div class="flex gap-2">
            <UButton size="sm" variant="secondary">编辑 {{ row.id }}</UButton>
          </div>
        </template>
      </UTable>
    </PanelBlock>

    <PanelBlock title="现代化表格">
      <p class="mb-3 text-xs text-zinc-500">
        表格沿用 Ant Design 的浅表头、行 hover、高密度信息展示，并保留固定列、文字省略、Tooltip 和列对齐能力。
      </p>
      <UTable
        :columns="advancedColumns"
        :rows="advancedRows"
        row-key="id"
        fixed
        stripe
        action-width="160px"
      >
        <template #cell-status="{ value }">
          <UTag :variant="value === '启用' ? 'success' : 'error'">{{ value }}</UTag>
        </template>
        <template #actions="{ row }">
          <div class="flex gap-2">
            <UButton size="sm" variant="secondary">编辑</UButton>
            <UButton size="sm" variant="ghost">详情</UButton>
          </div>
        </template>
      </UTable>
    </PanelBlock>

    <PanelBlock title="Tooltip 组件">
      <div class="flex flex-wrap items-center gap-4">
        <UTooltip content="这是一个基础的提示文本">
          <UButton size="sm" variant="secondary">悬停查看提示</UButton>
        </UTooltip>
        <UTooltip content="提示可以包含很长的说明内容，用于补充界面中无法完整展示的信息。">
          <span class="text-sm text-zinc-600 underline decoration-dotted">长文本提示</span>
        </UTooltip>
      </div>
    </PanelBlock>

    <UModal v-model:open="showModal" title="普通弹窗">
      <p class="text-sm leading-6 text-slate-600">
        用于承载说明、预览或表单内容。
      </p>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton variant="secondary" @click="showModal = false">关闭</UButton>
          <UButton @click="showModal = false">确认</UButton>
        </div>
      </template>
    </UModal>

    <UConfirmDialog
      v-model:open="showConfirm"
      title="确认删除示例"
      message="删除后将无法恢复该示例数据。"
      description="确认弹窗适用于删除、覆盖、停用等高风险操作。"
      confirm-text="确认删除"
      cancel-text="取消"
      danger
      @confirm="showConfirm = false"
    />
  </section>
</template>

<script setup>
import { reactive, ref } from "vue";
import PanelBlock from "../components/PanelBlock.vue";
import UButton from "../components/ued/UButton.vue";
import UConfirmDialog from "../components/ued/UConfirmDialog.vue";
import UInput from "../components/ued/UInput.vue";
import UModal from "../components/ued/UModal.vue";
import USelect from "../components/ued/USelect.vue";
import USwitch from "../components/ued/USwitch.vue";
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";
import UTooltip from "../components/ued/UTooltip.vue";

const showModal = ref(false);
const showConfirm = ref(false);
const switchValue = ref(true);

const form = reactive({
  name: "",
  protocol: "openai-responses",
  description: "",
});

const protocolOptions = [
  { label: "anthropic", value: "anthropic" },
  { label: "openai-chat", value: "openai-chat" },
  { label: "openai-responses", value: "openai-responses" },
];

const columns = [
  { key: "name", title: "名称" },
  { key: "type", title: "类型" },
  { key: "status", title: "状态" },
];

const rows = [
  { id: "1", name: "供应商按钮", type: "操作组件", status: "启用" },
  { id: "2", name: "确认弹窗", type: "反馈组件", status: "启用" },
];

const advancedColumns = [
  { key: "id", title: "ID", width: "80px", fixed: "left", align: "center" },
  {
    key: "name",
    title: "名称（固定左侧 + 省略 + Tooltip）",
    width: "200px",
    fixed: "left",
    ellipsis: true,
    tooltip: true,
  },
  {
    key: "description",
    title: "说明（溢出省略 + Tooltip）",
    width: "280px",
    ellipsis: true,
    tooltip: true,
  },
  { key: "type", title: "类型", width: "120px" },
  { key: "status", title: "状态", width: "90px", align: "center" },
  { key: "count", title: "计数", width: "100px", align: "right" },
];

const advancedRows = [
  {
    id: "101",
    name: "这是一个非常长的组件名称，用于测试文字溢出省略和 Tooltip 功能",
    description:
      "这是超长说明文本，用于演示当单元格内容超出列宽时，如何通过省略号隐藏并在悬停时通过 Tooltip 展示完整内容。",
    type: "操作组件",
    status: "启用",
    count: 128,
  },
  {
    id: "102",
    name: "确认弹窗",
    description: "用于二次确认的弹窗组件，适用于删除、覆盖等高风险操作场景。",
    type: "反馈组件",
    status: "停用",
    count: 56,
  },
  {
    id: "103",
    name: "数据表格",
    description:
      "现代化表格组件，支持固定列、表头固定、斑马纹、文字溢出省略、Tooltip 提示等丰富特性。",
    type: "数据展示",
    status: "启用",
    count: 2048,
  },
  {
    id: "104",
    name: "下拉选择器",
    description: "支持单选、搜索和分组的标准下拉选择组件。",
    type: "表单组件",
    status: "启用",
    count: 12,
  },
  {
    id: "105",
    name: "开关按钮",
    description: "用于切换配置项的布尔状态，支持禁用态和提示文本。",
    type: "表单组件",
    status: "启用",
    count: 99,
  },
];
</script>
