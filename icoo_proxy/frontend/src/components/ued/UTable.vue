<template>
  <div class="table-shell">
    <div class="table-scroll">
      <table class="admin-table">
        <thead>
          <tr>
            <th v-for="column in columns" :key="column.key">
              {{ column.title }}
            </th>
            <th v-if="$slots.actions">{{ actionTitle }}</th>
          </tr>
        </thead>
        <tbody v-if="rows.length">
          <tr v-for="row in rows" :key="resolveRowKey(row)">
            <td v-for="column in columns" :key="column.key">
              <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]">
                {{ row[column.key] ?? "-" }}
              </slot>
            </td>
            <td v-if="$slots.actions">
              <slot name="actions" :row="row" />
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!rows.length" class="empty-state rounded-none border-0">
        {{ emptyText }}
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  columns: {
    type: Array,
    default: () => [],
  },
  rows: {
    type: Array,
    default: () => [],
  },
  rowKey: {
    type: [String, Function],
    default: "id",
  },
  actionTitle: {
    type: String,
    default: "操作",
  },
  emptyText: {
    type: String,
    default: "暂无数据。",
  },
});

function resolveRowKey(row) {
  if (typeof props.rowKey === "function") {
    return props.rowKey(row);
  }
  return row?.[props.rowKey] ?? JSON.stringify(row);
}
</script>
