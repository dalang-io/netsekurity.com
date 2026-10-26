/** @type {import('tailwindcss').Config} */
module.exports = {
  // Every Go file that emits markup, not a hand-maintained subset: marketing.go,
  // pentests.go, domains.go, apitokens.go and payment.go all write Tailwind
  // classes in HTMX fragments, and were being purged out of the stylesheet.
  content: ["./static/index.html", "./*.go"],
  // alertHTML() builds "nsk-alert-" + kind at runtime, which the content scan
  // cannot see; without this the ok variant was purged and the toast shipped
  // with no background.
  safelist: [{ pattern: /^nsk-(alert|badge)-(ok|warn|bad|blue|muted|info)$/ }],
  theme: {
    extend: {
      colors: {
        ink: "#05070f",
        panel: "#0b0f1c",
        // Customer app (dashboard, checkout): the light theme shared with
        // dalang.io/dashboard — yellow primary, near-black ink, hard shadows.
        primary: { DEFAULT: "#facc15", dark: "#eab308" },
        brand: "#1c1a17",
      },
      fontFamily: {
        jakarta: ['"Plus Jakarta Sans"', "ui-sans-serif", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};