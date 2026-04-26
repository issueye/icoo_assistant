import { readonly, ref } from "vue";

const items = ref([]);
let seed = 0;

function normalizeOptions(type, content, options = {}) {
  if (typeof content === "object" && content !== null) {
    return {
      type,
      content: content.content || "",
      duration: content.duration,
      key: content.key,
      closable: content.closable,
    };
  }

  return {
    type,
    content,
    duration: options.duration,
    key: options.key,
    closable: options.closable,
  };
}

export const messageItems = readonly(items);

export function closeMessage(idOrKey) {
  items.value = items.value.filter((item) => item.id !== idOrKey && item.key !== idOrKey);
}

export function openMessage(options = {}) {
  const type = normalizeType(options.type);
  const key = options.key;
  const id = key || `ued-message-${Date.now()}-${seed++}`;
  const duration = options.duration ?? (type === "loading" ? 0 : 3000);
  const item = {
    id,
    key,
    type,
    content: options.content || "",
    closable: Boolean(options.closable),
  };

  const currentIndex = key ? items.value.findIndex((message) => message.key === key) : -1;
  if (currentIndex >= 0) {
    items.value.splice(currentIndex, 1, item);
  } else {
    items.value.push(item);
  }

  if (duration > 0) {
    window.setTimeout(() => closeMessage(id), duration);
  }

  return () => closeMessage(id);
}

export const message = {
  open: openMessage,
  success(content, options) {
    return openMessage(normalizeOptions("success", content, options));
  },
  info(content, options) {
    return openMessage(normalizeOptions("info", content, options));
  },
  warning(content, options) {
    return openMessage(normalizeOptions("warning", content, options));
  },
  error(content, options) {
    return openMessage(normalizeOptions("error", content, options));
  },
  loading(content, options) {
    return openMessage(normalizeOptions("loading", content, options));
  },
  destroy(key) {
    if (key) {
      closeMessage(key);
      return;
    }
    items.value = [];
  },
};

export function useMessage() {
  return message;
}

function normalizeType(type) {
  const value = String(type || "info").toLowerCase();
  if (value === "danger") {
    return "error";
  }
  if (value === "sucess") {
    return "success";
  }
  return ["success", "info", "warning", "error", "loading"].includes(value) ? value : "info";
}
