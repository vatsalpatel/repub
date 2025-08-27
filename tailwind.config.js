/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.templ",
    "./cmd/server/**/*.go"
  ],
  theme: {
    extend: {
      colors: {
        'pub-blue': '#007acc',
        'pub-blue-dark': '#005999',
        'pub-light': '#f0f8ff',
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}