/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{vue,js}"],
  theme: {
    extend: {
      screens: {
        sm: "480px",
        md: "640px",
        lg: "800px",
        xl: "1024px",
      },
      colors: {
        ant: {
          primary: "#1677ff",
          primaryHover: "#4096ff",
          primaryActive: "#0958d9",
          success: "#52c41a",
          warning: "#faad14",
          error: "#ff4d4f",
          text: "#262626",
          secondary: "#595959",
          tertiary: "#8c8c8c",
          border: "#d9d9d9",
          split: "#f0f0f0",
          layout: "#f5f5f5",
        },
      },
      boxShadow: {
        panel: "0 1px 2px rgba(0, 0, 0, 0.03), 0 4px 12px rgba(0, 0, 0, 0.04)",
        popup:
          "0 6px 16px rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 9px 28px 8px rgba(0, 0, 0, 0.05)",
      },
      fontFamily: {
        sans: ["Segoe UI", "PingFang SC", "Microsoft YaHei", "sans-serif"],
        mono: ["Cascadia Code", "Consolas", "monospace"],
      },
    },
  },
  plugins: [],
};
