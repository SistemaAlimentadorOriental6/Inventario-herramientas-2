/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: '#4cc253',
        'brand-dark': '#3db844',
      },
      fontFamily: {
        sans: ['Inter Variable', 'Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
