/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#00ADD8',
        'primary-dark': '#007D9C',
        'primary-light': '#5DC9E2',
        surface: '#FFFFFF',
        canvas: '#F8FAFC',
        ink: '#1E293B',
        muted: '#64748B',
        success: '#10B981',
        danger: '#EF4444',
      },
      borderRadius: {
        card: '12px',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
