/**
 * 前端通用常量。
 * 这里只保留当前项目真实使用的常量，移除已经被新可视化目录替代的旧组件清单。
 */

const envApiBaseUrl = (import.meta as unknown as { env: Record<string, string> }).env.VITE_API_BASE_URL

// API 基础路径
export const API_BASE_URL = '/api/v1'

// 后端服务地址
// 供 iframe / Web IDE / Web Terminal 等必须直连后端的场景使用。
export const API_SERVER_ORIGIN = (() => {
  const base = envApiBaseUrl || 'http://localhost:3000'
  return base.replace(/\/api\/v1\/?$/, '').replace(/\/$/, '')
})()

// WebSocket 地址
export const WS_BASE_URL = (import.meta as unknown as { env: Record<string, string> }).env.VITE_WS_URL || 'ws://localhost:3000/ws'

// 本地存储键名
export const STORAGE_KEYS = {
  ACCESS_TOKEN: 'chainspace_access_token',
  REFRESH_TOKEN: 'chainspace_refresh_token',
  USER_INFO: 'chainspace_user_info',
  THEME: 'chainspace_theme',
  SIDEBAR_COLLAPSED: 'chainspace_sidebar_collapsed',
} as const

// 分页默认配置
export const PAGINATION_CONFIG = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 20,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
  MAX_PAGE_SIZE: 100,
} as const

// 实验环境配置
export const EXPERIMENT_CONFIG = {
  DEFAULT_TIMEOUT: 4 * 60 * 60,
  EXTEND_DURATION: 4 * 60 * 60,
  STATUS_POLL_INTERVAL: 3000,
} as const

// 上传配置
export const UPLOAD_CONFIG = {
  MAX_FILE_SIZE: 100 * 1024 * 1024,
  IMAGE_TYPES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  DOCUMENT_TYPES: [
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  ],
  VIDEO_TYPES: ['video/mp4', 'video/webm', 'video/ogg'],
} as const

// 日期时间格式
export const DATE_FORMAT = {
  DATE: 'YYYY-MM-DD',
  TIME: 'HH:mm:ss',
  DATETIME: 'YYYY-MM-DD HH:mm:ss',
  DATETIME_SHORT: 'YYYY-MM-DD HH:mm',
} as const

// 主题色
export const COLORS = {
  PRIMARY: '#1890FF',
  SUCCESS: '#52C41A',
  WARNING: '#FAAD14',
  ERROR: '#FF4D4F',
  TEXT_PRIMARY: '#262626',
  TEXT_SECONDARY: '#8C8C8C',
  BORDER: '#E8E8E8',
  BG_LIGHT: '#E8F4FC',
} as const

// 交互工具列表
export const INTERACTION_TOOLS = [
  { key: 'ide', label: 'IDE编辑器', icon: 'code' },
  { key: 'terminal', label: '终端', icon: 'terminal' },
  { key: 'files', label: '文件管理', icon: 'folder' },
  { key: 'rpc', label: 'RPC调试', icon: 'send' },
  { key: 'explorer', label: '区块浏览器', icon: 'search' },
  { key: 'logs', label: '日志查看', icon: 'file-text' },
  { key: 'api_debug', label: 'API调试', icon: 'send' },
  { key: 'visualization', label: '可视化沙箱', icon: 'eye' },
  { key: 'network', label: '网络拓扑', icon: 'git-branch' },
] as const

// 通用筛选项
export const FILTER_OPTIONS = {
  CONTEST_TYPE: [
    { label: '全部', value: '' },
    { label: '解题赛', value: 'jeopardy' },
    { label: '对抗赛', value: 'agent_battle' },
  ],
  CONTEST_STATUS: [
    { label: '全部', value: '' },
    { label: '草稿', value: 'draft' },
    { label: '已发布', value: 'published' },
    { label: '进行中', value: 'ongoing' },
    { label: '已结束', value: 'ended' },
  ],
  COURSE_STATUS: [
    { label: '全部', value: '' },
    { label: '草稿', value: 'draft' },
    { label: '已发布', value: 'published' },
    { label: '已归档', value: 'archived' },
  ],
  USER_STATUS: [
    { label: '全部', value: '' },
    { label: '正常', value: 'active' },
    { label: '已禁用', value: 'disabled' },
  ],
  // 实验筛选配置 - 按角色区分
  // EXPERIMENT_STATUS: [
  //   { label: '全部', value: '' },
  //   { label: '草稿', value: 'draft' },
  //   { label: '已发布', value: 'published' },
  //   { label: '已归档', value: 'archived' },
  // ],
  CHALLENGE_DIFFICULTY: [
    { label: '全部', value: '' },
    { label: '入门', value: '1' },
    { label: '简单', value: '2' },
    { label: '中等', value: '3' },
    { label: '困难', value: '4' },
    { label: '专家', value: '5' },
  ],
  CHALLENGE_CATEGORY: [
    { label: '全部', value: '' },
    { label: '合约漏洞', value: 'contract_vuln' },
    { label: 'DeFi', value: 'defi' },
    { label: '共识机制', value: 'consensus' },
    { label: '密码学', value: 'crypto' },
    { label: '其他', value: 'misc' },
  ],
  VULNERABILITY_STATUS: [
    { label: '全部', value: '' },
    { label: '待转化', value: 'active' },
    { label: '已转化', value: 'converted' },
    { label: '已跳过', value: 'skipped' },
  ],
  GRADING_STATUS: [
    { label: '全部', value: '' },
    { label: '待批改', value: 'submitted' },
    { label: '已批改', value: 'graded' },
  ],
  DISCUSSION_STATUS: [
    { label: '全部', value: '' },
    { label: '正常', value: 'normal' },
    { label: '置顶', value: 'pinned' },
    { label: '已关闭', value: 'closed' },
  ],
} as const

// Docker 镜像分类
export const IMAGE_CATEGORIES = [
  { key: 'base', label: '基础开发' },
  { key: 'ethereum', label: '以太坊' },
  { key: 'bitcoin', label: '比特币' },
  { key: 'solana', label: 'Solana' },
  { key: 'substrate', label: 'Substrate' },
  { key: 'cosmos', label: 'Cosmos' },
  { key: 'move', label: 'Move' },
  { key: 'l2_privacy', label: 'L2与隐私' },
  { key: 'fabric', label: 'Hyperledger Fabric' },
  { key: 'fisco', label: 'FISCO BCOS' },
  { key: 'chainmaker', label: 'ChainMaker' },
  { key: 'infrastructure', label: '基础设施' },
  { key: 'security', label: '安全工具' },
  { key: 'simulation', label: '可视化模拟' },
] as const
