/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./static/index.html", "./dashboard.go", "./admin.go", "./stack.go", "./header.go"],
  theme: {
    extend: {
      colors: {
        ink: "#05070f",
        panel: "#0b0f1c",
      },
    },
  },
  plugins: [],
};