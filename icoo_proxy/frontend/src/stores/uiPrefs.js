import { defineStore } from "pinia";
import { GetUiPrefs, SaveUiPrefs } from "../lib/wailsApp";

const defaultPrefs = () => ({
  theme: "blue",
  buttonSize: "md",
});

const themes = {
  blue: {
    "--ued-color-primary": "#1677ff",
    "--ued-color-primary-hover": "#4096ff",
    "--ued-color-primary-active": "#0958d9",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#e6f4ff",
  },
  green: {
    "--ued-color-primary": "#52c41a",
    "--ued-color-primary-hover": "#73d13d",
    "--ued-color-primary-active": "#389e0d",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#f6ffed",
  },
  purple: {
    "--ued-color-primary": "#722ed1",
    "--ued-color-primary-hover": "#9254de",
    "--ued-color-primary-active": "#531dab",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#f9f0ff",
  },
  orange: {
    "--ued-color-primary": "#fa8c16",
    "--ued-color-primary-hover": "#ffa940",
    "--ued-color-primary-active": "#d46b08",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#fff7e6",
  },
  red: {
    "--ued-color-primary": "#f5222d",
    "--ued-color-primary-hover": "#ff4d4f",
    "--ued-color-primary-active": "#cf1322",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#fff1f0",
  },
  cyan: {
    "--ued-color-primary": "#13c2c2",
    "--ued-color-primary-hover": "#36cfc9",
    "--ued-color-primary-active": "#08979c",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#e6fffb",
  },
  dark: {
    "--ued-color-primary": "#141414",
    "--ued-color-primary-hover": "#262626",
    "--ued-color-primary-active": "#000000",
    "--ued-color-primary-foreground": "#ffffff",
    "--ued-color-primary-soft": "#f5f5f5",
  },
};

const buttonSizes = {
  xs: {
    "--ued-size-xs": "20px",
    "--ued-size-sm": "24px",
    "--ued-size-md": "28px",
    "--ued-size-lg": "36px",
  },
  sm: {
    "--ued-size-xs": "22px",
    "--ued-size-sm": "26px",
    "--ued-size-md": "30px",
    "--ued-size-lg": "38px",
  },
  md: {
    "--ued-size-xs": "24px",
    "--ued-size-sm": "28px",
    "--ued-size-md": "32px",
    "--ued-size-lg": "40px",
  },
  lg: {
    "--ued-size-xs": "28px",
    "--ued-size-sm": "32px",
    "--ued-size-md": "36px",
    "--ued-size-lg": "44px",
  },
};

function applyToDocument(prefs) {
  const root = document.documentElement;
  const themeVars = themes[prefs.theme] || themes.blue;
  for (const [key, value] of Object.entries(themeVars)) {
    root.style.setProperty(key, value);
  }
  const sizeVars = buttonSizes[prefs.buttonSize] || buttonSizes.md;
  for (const [key, value] of Object.entries(sizeVars)) {
    root.style.setProperty(key, value);
  }
  root.setAttribute("data-theme", prefs.theme);
  root.setAttribute("data-button-size", prefs.buttonSize);
}

export const useUiPrefsStore = defineStore("uiPrefs", {
  state: () => ({
    prefs: defaultPrefs(),
    ready: false,
  }),
  getters: {
    theme: (state) => state.prefs.theme,
    buttonSize: (state) => state.prefs.buttonSize,
    themeOptions: () => [
      { label: "蓝色", value: "blue", color: "#1677ff" },
      { label: "绿色", value: "green", color: "#52c41a" },
      { label: "紫色", value: "purple", color: "#722ed1" },
      { label: "橙色", value: "orange", color: "#fa8c16" },
      { label: "红色", value: "red", color: "#f5222d" },
      { label: "青色", value: "cyan", color: "#13c2c2" },
      { label: "深色", value: "dark", color: "#141414" },
    ],
    buttonSizeOptions: () => [
      { label: "紧凑 (XS)", value: "xs" },
      { label: "偏小 (SM)", value: "sm" },
      { label: "默认 (MD)", value: "md" },
      { label: "偏大 (LG)", value: "lg" },
    ],
  },
  actions: {
    init() {
      applyToDocument(this.prefs);
      this.loadFromBackend();
    },
    async loadFromBackend() {
      try {
        const prefs = await GetUiPrefs();
        if (prefs && prefs.theme) {
          this.prefs.theme = prefs.theme;
          this.prefs.buttonSize = prefs.buttonSize || "md";
          applyToDocument(this.prefs);
        }
      } catch (e) {
        // use defaults already applied
      } finally {
        this.ready = true;
      }
    },
    async setTheme(theme) {
      this.prefs.theme = theme;
      applyToDocument(this.prefs);
      try {
        await SaveUiPrefs({ theme: this.prefs.theme, buttonSize: this.prefs.buttonSize });
      } catch (e) {
        // saved locally even if backend fails
      }
    },
    async setButtonSize(size) {
      this.prefs.buttonSize = size;
      applyToDocument(this.prefs);
      try {
        await SaveUiPrefs({ theme: this.prefs.theme, buttonSize: this.prefs.buttonSize });
      } catch (e) {
        // saved locally even if backend fails
      }
    },
  },
});
