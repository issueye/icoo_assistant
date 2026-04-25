<template>
  <div class="table-shell" :class="{ 'is-sticky-header': stickyHeader }">
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
            <th
              v-for="(column, index) in columns"
              :key="column.key"
              :class="[
                column.headerClass,
                getAlignClass(column.align),
                { 'is-sticky': isStickyColumn(column) },
                { 'is-sticky-left': column.fixed === 'left' },
                { 'is-sticky-right': column.fixed === 'right' },
              ]"
              :style="getStickyStyle(column, index)"
            >
              {{ column.title }}
            </th>
            <th
              v-if="$slots.actions"
              :class="[
                'actions-header',
                getAlignClass(actionAlign),
                { 'is-sticky': true, 'is-sticky-right': true },
              ]"
              :style="getActionStickyStyle()"
            >
              {{ actionTitle }}
            </th>
          </tr>
        </thead>
        <tbody v-if="rows.length">
          <tr
            v-for="row in rows"
            :key="resolveRowKey(row)"
            :class="{ 'is-striped': stripe }"
          >
            <td
              v-for="(column, index) in columns"
              :key="column.key"
              :class="[
                column.cellClass,
                getAlignClass(column.align),
                { 'is-sticky': isStickyColumn(column) },
                { 'is-sticky-left': column.fixed === 'left' },
                { 'is-sticky-right': column.fixed === 'right' },
                { 'is-ellipsis': column.ellipsis },
              ]"
              :style="getStickyStyle(column, index)"
            >
              <template v-if="column.ellipsis && !slots[`cell-${column.key}`]">
                <UTooltip
                  :content="resolveTooltipContent(column, row)"
                  :disabled="!column.tooltip"
                >
                  <span class="table-cell-ellipsis">{{ row[column.key] ?? "-" }}</span>
                </UTooltip>
              </template>
              <template v-else>
                <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]">
                  {{ row[column.key] ?? "-" }}
                </slot>
              </template>
            </td>
            <td
              v-if="$slots.actions"
              :class="[
                getAlignClass(actionAlign),
                { 'is-sticky': true, 'is-sticky-right': true },
              ]"
              :style="getActionStickyStyle()"
            >
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
import { computed, useSlots } from "vue";
import UTooltip from "./UTooltip.vue";

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
  actionAlign: {
    type: String,
    default: "center",
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
  stripe: {
    type: Boolean,
    default: false,
  },
  stickyHeader: {
    type: Boolean,
    default: true,
  },
});

const slots = useSlots();

const tableClasses = computed(() => [
  "admin-table",
  props.fixed ? "admin-table-fixed" : "",
  props.tableClass,
  props.stripe ? "admin-table--stripe" : "",
  props.stickyHeader ? "admin-table--sticky-header" : "",
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

function resolveTooltipContent(column, row) {
  if (typeof column.tooltip === "function") {
    return column.tooltip(row);
  }
  return String(row[column.key] ?? "-");
}

function getAlignClass(align) {
  if (align === "center") return "is-align-center";
  if (align === "right") return "is-align-right";
  return "is-align-left";
}

function isStickyColumn(column) {
  return column.fixed === "left" || column.fixed === "right";
}

function getStickyStyle(column, index) {
  const style = {};
  if (!isStickyColumn(column)) return style;

  // Calculate left offset for left-fixed columns
  if (column.fixed === "left") {
    const parts = [];
    for (let i = 0; i < index; i++) {
      const prev = props.columns[i];
      if (prev.fixed === "left" && prev.width) {
        parts.push(prev.width);
      }
    }
    style.left = parts.length ? `calc(${parts.join(" + ")})` : "0px";
  }

  // Calculate right offset for right-fixed columns
  if (column.fixed === "right") {
    const parts = [];
    for (let i = props.columns.length - 1; i > index; i--) {
      const next = props.columns[i];
      if (next.fixed === "right" && next.width) {
        parts.push(next.width);
      }
    }
    // Also account for actions column if it exists
    if (slots.actions && props.actionWidth) {
      parts.push(props.actionWidth);
    }
    style.right = parts.length ? `calc(${parts.join(" + ")})` : "0px";
  }

  return style;
}

function getActionStickyStyle() {
  const style = {};
  if (!slots.actions) return style;
  style.right = "0px";
  return style;
}
</script>
