/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#1F8457',
        accent: '#D9CAB3',
        'sidebar-bg': '#212121',
        'sidebar-hover': '#2D2D2D',
        'sidebar-active': '#3A3A3A',
        'purple-500': '#1F8457',
        'purple-400': '#1F8457',
        'purple-300': '#D9CAB3',
        'purple-250': '#D9CAB3',
        'purple-200': '#D9CAB3',
        'purple-100': '#D9CAB3',
        'purple-50': '#F6F6F6',
        'bg-light': '#F6F6F6',
        'status-error': '#FF3D89',
        'status-success': '#00BE8F',
        'status-warning': '#D4B106',
        'status-info': '#1890FF',
        'text-primary': '#000000D9',
        'text-secondary': '#00000073',
      },
      fontFamily: {
        sans: ['Noto Sans Thai Looped', 'Noto Sans Thai', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
    },
  },
  plugins: [],
  // ป้องกัน Tailwind CSS conflict กับ Ant Design
  corePlugins: {
    preflight: false,
  },
}
