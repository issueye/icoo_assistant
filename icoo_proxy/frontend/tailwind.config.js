/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{vue,js}"],
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#081018",
          900: "#0d1723",
          850: "#132030",
          800: "#1d2d3f",
        },
        signal: {
          amber: "#ffb15e",
          mint: "#56c4a5",
          coral: "#ff7a6b",
          sky: "#86b6ff",
        },
      },
      boxShadow: {
        panel: "0 24px 60px rgba(0, 0, 0, 0.28)",
      },
      fontFamily: {
        sans: ["Nunito", "Avenir Next", "Segoe UI", "sans-serif"],
        mono: ["Cascadia Code", "Consolas", "monospace"],
      },
      backgroundImage: {
        ambient:
          "radial-gradient(circle at top left, rgba(255,177,94,0.16), transparent 30%), radial-gradient(circle at right center, rgba(86,196,165,0.16), transparent 30%), linear-gradient(180deg, #081018 0%, #132030 100%)",
      },
    },
  },
  plugins: [],
};
