<template>
  <div class="table-shell">
    <div class="table-scroll">
      <table :class="tableClasses">
        <colgroup v-if="hasColumnSizing">
          <col
            v-for="column in columns"
            :key="column.key"
            :style="column.width ? { width: column.width } : undefined"
          />
          <col v-if="$slots.actions" :style="actionWidth ? { width: actionWidth } : undefined" />
        </colgroup>
        <thead>
          <tr>
            <th v-for="column in columns" :key="column.key" :class="column.headerClass">
              {{ column.title }}
            </th>
            <th v-if="$slots.actions">{{ actionTitle }}</th>
          </tr>
        </thead>
        <tbody v-if="rows.length">
          <tr v-for="row in rows" :key="resolveRowKey(row)">
            <td v-for="column in columns" :key="column.key" :class="column.cellClass">
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
import { computed } from "vue";

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
  actionWidth: {
    type: String,
    default: "",
  },
  emptyText: {
    type: String,
    default: "暂无数据。",
  },
  fixed: {
    type: Boolean,
    default: false,
  },
  tableClass: {
    type: [String, Array, Object],
    default: "",
  },
});

const tableClasses = computed(() => [
  "admin-table",
  props.fixed ? "admin-table-fixed" : "",
  props.tableClass,
]);

const hasColumnSizing = computed(() =>
  props.columns.some((column) => Boolean(column.width)) || Boolean(props.actionWidth),
);

function resolveRowKey(row) {
  if (typeof props.rowKey === "function") {
    return props.rowKey(row);
  }
  return row?.[props.rowKey] ?? JSON.stringify(row);
}
</script>
