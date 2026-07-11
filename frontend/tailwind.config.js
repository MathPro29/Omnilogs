/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#6647F0',
        'sidebar-bg': '#100446',
        'sidebar-hover': '#1C0B68',
        'sidebar-active': '#2A158A',
        'purple-500': '#3C22AC',
        'purple-400': '#4F33CE',
        'purple-300': '#7B68EE',
        'purple-250': '#8D74FF',
        'purple-200': '#AE9CFF',
        'purple-100': '#CFC4FF',
        'purple-50': '#F0EDFF',
        'status-error': '#FF3D89',
        'status-success': '#00BE8F',
        'status-warning': '#D4B106',
        'status-info': '#1890FF',
        'text-primary': '#000000D9',
        'text-secondary': '#00000073',
      },
      fontFamily: {
        sans: ['Noto Sans Thai', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
    },
  },
  plugins: [],
  // ป้องกัน Tailwind CSS conflict กับ Ant Design
  corePlugins: {
    preflight: false,
  },
}
