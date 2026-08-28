/** @type {import('tailwindcss').Config} */
module.exports = {
  // Every Go file that emits markup, not a hand-maintained subset: marketing.go,
  // pentests.go, domains.go, apitokens.go and payment.go all write Tailwind
  // classes in HTMX fragments, and were being purged out of the stylesheet.
  content: ["./static/index.html", "./*.go"],
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