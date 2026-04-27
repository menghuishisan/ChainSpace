/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      // 链境 ChainSpace 色彩体系
      colors: {
        primary: {
          light: '#E8F4FC',   // 浅蓝 - 背景渐变起始色
          DEFAULT: '#1890FF', // 强调色
          dark: '#096DD9',
        },
        success: '#52C41A',
        warning: '#FAAD14',
        error: '#FF4D4F',
        text: {
          primary: '#262626',   // 主文字
          secondary: '#8C8C8C', // 辅助文字
        },
        border: '#E8E8E8',
      },
      // 间距规范
      spacing: {
        'xs': '4px',
        'sm': '8px',
        'md': '16px',
        'lg': '24px',
        'xl': '32px',
      },
      // 圆角规范
      borderRadius: {
        'sm': '4px',
        'md': '8px',
        'lg': '16px',
      },
      // 阴影规范
      boxShadow: {
        'card': '0 2px 8px rgba(0, 0, 0, 0.08)',
        'hover': '0 4px 16px rgba(0, 0, 0, 0.12)',
        'modal': '0 8px 24px rgba(0, 0, 0, 0.15)',
      },
      // 字体大小规范
      fontSize: {
        'title-lg': ['24px', { lineHeight: '32px', fontWeight: '600' }],
        'title-md': ['18px', { lineHeight: '26px', fontWeight: '600' }],
        'title-sm': ['16px', { lineHeight: '24px', fontWeight: '500' }],
        'body': ['14px', { lineHeight: '22px', fontWeight: '400' }],
        'caption': ['12px', { lineHeight: '20px', fontWeight: '400' }],
      },
    },
  },
  plugins: [],
  // 避免与 Ant Design 冲突
  corePlugins: {
    preflight: false,
  },
}
