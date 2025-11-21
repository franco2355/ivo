/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#4CAF50',
          dark: '#388e3c',
          light: '#81c784',
        },
        secondary: {
          DEFAULT: '#2196F3',
          dark: '#1976D2',
          light: '#64B5F6',
        },
        warning: {
          DEFAULT: '#FF9800',
          dark: '#F57C00',
          light: '#FFB74D',
        },
        danger: {
          DEFAULT: '#f44336',
          dark: '#d32f2f',
          light: '#ff5252',
        },
        dark: {
          DEFAULT: '#1a1a1a',
          light: '#232323',
          lighter: '#333',
        },
        gray: {
          DEFAULT: '#666',
          light: '#f5f5f5',
          lighter: '#f9f9f9',
        }
      },
      fontFamily: {
        sans: ['system-ui', 'Segoe UI', 'Avenir', 'Helvetica', 'Arial', 'sans-serif'],
      },
      boxShadow: {
        'card': '0 4px 6px rgba(0, 0, 0, 0.1)',
        'card-hover': '0 6px 10px rgba(0, 0, 0, 0.15)',
        'lg': '0 8px 32px rgba(0,0,0,0.25)',
      },
      borderRadius: {
        'card': '12px',
      },
      transitionDuration: {
        '250': '250ms',
      }
    },
  },
  plugins: [],
}
