/**
 * 链境 ChainSpace 前端入口文件
 * 区块链教学实验平台
 */
import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'

import App from './App'
import './assets/styles/index.css'

// 设置 dayjs 中文
dayjs.locale('zh-cn')

// Ant Design 主题配置
const antdTheme = {
  token: {
    colorPrimary: '#1890FF',
    colorSuccess: '#52C41A',
    colorWarning: '#FAAD14',
    colorError: '#FF4D4F',
    colorTextBase: '#262626',
    colorBorder: '#E8E8E8',
    borderRadius: 4,
  },
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN} theme={antdTheme}>
      <App />
    </ConfigProvider>
  </React.StrictMode>,
)
