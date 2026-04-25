<template>
  <section class="page-section">
    <div class="page-header">
      <h2 class="page-title">UED 组件</h2>
      <p class="page-description">
        简洁、明亮、低内边距。组件以细边框、小圆角和紧凑间距为主。
      </p>
    </div>

    <div class="section-grid xl:grid-cols-2">
      <PanelBlock title="按钮">
        <div class="flex flex-wrap gap-2">
          <UButton>主要按钮</UButton>
          <UButton variant="secondary">次要按钮</UButton>
          <UButton variant="danger">危险按钮</UButton>
          <UButton variant="ghost">文本按钮</UButton>
          <UButton loading>加载中</UButton>
        </div>
      </PanelBlock>

      <PanelBlock title="开关">
        <div class="space-y-3">
          <USwitch v-model="switchValue" label="启用功能" hint="配置类开关" />
          <USwitch :model-value="true" label="已启用" hint="禁用态示例" :disabled="true" />
        </div>
      </PanelBlock>
    </div>

    <div class="section-grid xl:grid-cols-2">
      <PanelBlock title="输入与下拉">
        <div class="space-y-3">
          <UInput v-model="form.name" label="名称" placeholder="请输入名称" />
          <USelect v-model="form.protocol" label="协议" :options="protocolOptions" />
          <UInput v-model="form.description" label="描述" placeholder="请输入描述" textarea />
        </div>
      </PanelBlock>

      <PanelBlock title="弹窗">
        <div class="flex flex-wrap gap-2">
          <UButton @click="showModal = true">普通弹窗</UButton>
          <UButton variant="danger" @click="showConfirm = true">确认弹窗</UButton>
        </div>
      </PanelBlock>
    </div>

    <PanelBlock title="表格">
      <UTable :columns="columns" :rows="rows" row-key="id" empty-text="暂无组件示例数据。">
        <template #cell-status="{ value }">
          <span class="badge" :class="value === '启用' ? 'badge-success' : 'badge-danger'">{{ value }}</span>
        </template>
        <template #actions="{ row }">
          <div class="flex gap-2">
            <UButton size="sm" variant="secondary">编辑 {{ row.id }}</UButton>
          </div>
        </template>
      </UTable>
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
</script>
