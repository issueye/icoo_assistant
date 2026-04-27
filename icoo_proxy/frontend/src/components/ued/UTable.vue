<template>
  <div class="table-shell" :class="shellClasses">
    <div class="table-scroll" :style="scrollStyle">
      <table :class="tableClasses" :style="tableStyle">
        <colgroup v-if="tableColumns.length">
          <col
            v-for="column in tableColumns"
            :key="column.uid"
            :style="getColStyle(column)"
          />
        </colgroup>

        <thead v-if="showHeader">
          <tr>
            <th
              v-for="column in tableColumns"
              :key="column.uid"
              scope="col"
              :class="getHeaderClasses(column)"
              :style="getStickyStyle(column)"
            >
              <slot
                v-if="!column.isAction"
                :name="`header-${column.key}`"
                :column="column.raw"
              >
                {{ column.title }}
              </slot>
              <template v-else>{{ column.title }}</template>
            </th>
          </tr>
        </thead>

        <tbody v-if="hasRows">
          <tr
            v-for="(row, rowIndex) in rows"
            :key="resolveRowKey(row, rowIndex)"
            :class="getRowClasses(row, rowIndex)"
          >
            <td
              v-for="column in tableColumns"
              :key="column.uid"
              :class="getCellClasses(column, row, rowIndex)"
              :style="getStickyStyle(column)"
            >
              <slot
                v-if="column.isAction"
                name="actions"
                :row="row"
                :index="rowIndex"
              />

              <template v-else-if="column.ellipsis && !hasCellSlot(column)">
                <UTooltip
                  :content="resolveTooltipContent(column, row, rowIndex)"
                  :disabled="!column.tooltip"
                >
                  <span class="table-cell-ellipsis">
                    {{ formatCellValue(resolveCellValue(column, row, rowIndex)) }}
                  </span>
                </UTooltip>
              </template>

              <slot
                v-else
                :name="`cell-${column.key}`"
                :row="row"
                :value="resolveCellValue(column, row, rowIndex)"
                :column="column.raw"
                :index="rowIndex"
              >
                {{ formatCellValue(resolveCellValue(column, row, rowIndex)) }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="!hasRows" class="empty-state rounded-none border-0">
        <slot name="empty">{{ emptyText }}</slot>
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
    type: [String, Number],
    default: "96px",
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
  showHeader: {
    type: Boolean,
    default: true,
  },
  maxHeight: {
    type: [String, Number],
    default: "",
  },
  minWidth: {
    type: [String, Number],
    default: "",
  },
  size: {
    type: String,
    default: "middle",
  },
  rowClassName: {
    type: [String, Function],
    default: "",
  },
});

const slots = useSlots();

const normalizedColumns = computed(() =>
  props.columns
    .filter((column) => !column.hidden)
    .map((column, index) => normalizeColumn(column, index)),
);

const actionColumn = computed(() => {
  if (!slots.actions) return null;
  const actionWidth = normalizeCssSize(props.actionWidth) || "180px";
  return {
    uid: "__actions__",
    key: "actions",
    title: props.actionTitle,
    width: actionWidth,
    minWidth: actionWidth,
    align: props.actionAlign,
    fixed: "right",
    ellipsis: false,
    tooltip: false,
    isAction: true,
    raw: {
      key: "actions",
      title: props.actionTitle,
    },
  };
});

const tableColumns = computed(() => {
  const columns = [...normalizedColumns.value];
  if (actionColumn.value) {
    columns.push(actionColumn.value);
  }
  return withStickyOffsets(columns);
});

const hasRows = computed(() => props.rows.length > 0);

const hasColumnSizing = computed(() =>
  tableColumns.value.some((column) => Boolean(column.width || column.minWidth)),
);

const shellClasses = computed(() => ({
  "table-shell--empty": !hasRows.value,
  "table-shell--fixed": props.fixed,
  "table-shell--sticky-header": props.stickyHeader,
}));

const tableClasses = computed(() => [
  "admin-table",
  props.fixed ? "admin-table-fixed" : "",
  props.tableClass,
  props.stripe ? "admin-table--stripe" : "",
  props.stickyHeader ? "admin-table--sticky-header" : "",
  `admin-table--${normalizeSize(props.size)}`,
]);

const scrollStyle = computed(() => {
  const style = {};
  const maxHeight = normalizeCssSize(props.maxHeight);
  if (maxHeight) {
    style.maxHeight = maxHeight;
  }
  return style;
});

const tableStyle = computed(() => {
  const style = {};
  const minWidth = resolveTableMinWidth();
  if (minWidth) {
    style.minWidth = minWidth;
  }
  return style;
});

function normalizeColumn(column, index) {
  const key = String(column.key ?? column.dataIndex ?? `column-${index}`);
  return {
    uid: `${key}-${index}`,
    key,
    title: column.title ?? "",
    dataIndex: column.dataIndex ?? column.key,
    width: normalizeCssSize(column.width),
    minWidth: normalizeCssSize(column.minWidth),
    align: normalizeAlign(column.align),
    fixed: normalizeFixed(column.fixed),
    ellipsis: Boolean(column.ellipsis),
    tooltip: column.tooltip,
    className: column.className,
    headerClass: column.headerClass,
    cellClass: column.cellClass,
    isAction: false,
    raw: column,
  };
}

function withStickyOffsets(columns) {
  const next = columns.map((column) => ({ ...column, stickyStyle: {} }));
  let leftOffset = "0px";
  let lastLeftIndex = -1;
  let firstRightIndex = -1;

  next.forEach((column, index) => {
    if (column.fixed === "left") {
      lastLeftIndex = index;
    }
    if (column.fixed === "right" && firstRightIndex === -1) {
      firstRightIndex = index;
    }
  });

  for (const column of next) {
    if (column.fixed !== "left") continue;
    column.stickyStyle.left = leftOffset;
    leftOffset = appendCssSize(leftOffset, column.width);
  }

  let rightOffset = "0px";
  for (let index = next.length - 1; index >= 0; index -= 1) {
    const column = next[index];
    if (column.fixed !== "right") continue;
    column.stickyStyle.right = rightOffset;
    rightOffset = appendCssSize(rightOffset, column.width);
  }

  if (lastLeftIndex >= 0) {
    next[lastLeftIndex].isStickyLeftLast = true;
  }
  if (firstRightIndex >= 0) {
    next[firstRightIndex].isStickyRightFirst = true;
  }

  return next;
}

function appendCssSize(base, size) {
  if (!size) return base;
  if (base === "0px") return size;
  return `calc(${base} + ${size})`;
}

function normalizeCssSize(value) {
  if (value === 0) return "0px";
  if (!value) return "";
  return typeof value === "number" ? `${value}px` : String(value);
}

function resolveTableMinWidth() {
  const explicitMinWidth = normalizeCssSize(props.minWidth);
  if (explicitMinWidth) {
    return explicitMinWidth;
  }
  if (!props.fixed) {
    return "";
  }

  const total = tableColumns.value.reduce((sum, column) => {
    return sum + estimateColumnWidth(column);
  }, 0);
  return total > 0 ? `${Math.max(total, 960)}px` : "960px";
}

function estimateColumnWidth(column) {
  const minWidth = parsePixelSize(column.minWidth);
  if (minWidth) return minWidth;

  const width = parsePixelSize(column.width);
  if (width) return width;

  if (column.isAction) return 180;
  if (column.ellipsis) return 220;
  return 180;
}

function parsePixelSize(value) {
  if (!value) return 0;
  const match = String(value).match(/^(\d+(?:\.\d+)?)px$/);
  return match ? Number(match[1]) : 0;
}

function normalizeAlign(value) {
  if (value === "center" || value === "right") return value;
  return "left";
}

function normalizeFixed(value) {
  if (value === "left" || value === true) return "left";
  if (value === "right") return "right";
  return "";
}

function normalizeSize(value) {
  return ["small", "middle", "large"].includes(value) ? value : "middle";
}

function resolveRowKey(row, rowIndex) {
  if (typeof props.rowKey === "function") {
    return props.rowKey(row, rowIndex);
  }
  return row?.[props.rowKey] ?? rowIndex;
}

function resolveCellValue(column, row, rowIndex) {
  if (typeof column.raw.render === "function") {
    return column.raw.render(row, rowIndex);
  }

  const dataIndex = column.dataIndex;
  if (Array.isArray(dataIndex)) {
    return dataIndex.reduce((value, key) => value?.[key], row);
  }

  return row?.[dataIndex];
}

function resolveTooltipContent(column, row, rowIndex) {
  if (typeof column.tooltip === "function") {
    return column.tooltip(row, rowIndex);
  }
  return formatCellValue(resolveCellValue(column, row, rowIndex));
}

function formatCellValue(value) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return String(value);
}

function hasCellSlot(column) {
  return Boolean(slots[`cell-${column.key}`]);
}

function isStickyColumn(column) {
  return column.fixed === "left" || column.fixed === "right";
}

function getColStyle(column) {
  if (!hasColumnSizing.value) return undefined;
  const style = {};
  if (column.width) {
    style.width = column.width;
  }
  if (column.minWidth) {
    style.minWidth = column.minWidth;
  }
  return Object.keys(style).length ? style : undefined;
}

function getStickyStyle(column) {
  if (!isStickyColumn(column)) return undefined;
  return column.stickyStyle;
}

function getHeaderClasses(column) {
  return [
    column.headerClass,
    column.className,
    getAlignClass(column.align),
    {
      "is-sticky": isStickyColumn(column),
      "is-sticky-left": column.fixed === "left",
      "is-sticky-left-last": column.isStickyLeftLast,
      "is-sticky-right": column.fixed === "right",
      "is-sticky-right-first": column.isStickyRightFirst,
      "actions-header": column.isAction,
    },
  ];
}

function getCellClasses(column, row, rowIndex) {
  return [
    column.cellClass,
    column.className,
    getAlignClass(column.align),
    {
      "is-sticky": isStickyColumn(column),
      "is-sticky-left": column.fixed === "left",
      "is-sticky-left-last": column.isStickyLeftLast,
      "is-sticky-right": column.fixed === "right",
      "is-sticky-right-first": column.isStickyRightFirst,
      "is-ellipsis": column.ellipsis,
      "actions-header": column.isAction,
    },
    typeof column.raw.cellClassName === "function"
      ? column.raw.cellClassName(row, rowIndex)
      : "",
  ];
}

function getRowClasses(row, rowIndex) {
  return [
    {
      "is-striped": props.stripe,
    },
    typeof props.rowClassName === "function"
      ? props.rowClassName(row, rowIndex)
      : props.rowClassName,
  ];
}

function getAlignClass(align) {
  if (align === "center") return "is-align-center";
  if (align === "right") return "is-align-right";
  return "is-align-left";
}
</script>
