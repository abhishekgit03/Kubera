/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        kubera: {
          purple: '#7c3aed',
          amber: '#d97706',
          green: '#059669',
          red: '#dc2626',
        },
      },
    },
  },
  plugins: [require('@tailwindcss/typography')],
}
