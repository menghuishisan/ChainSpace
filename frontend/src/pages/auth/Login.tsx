/**
 * 登录页面
 * 左右分栏布局：左侧平台介绍，右侧登录表单
 */
import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Form, Input, Button, Checkbox, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { Link, Shield, Container, Trophy } from 'lucide-react'
import { useUserStore } from '@/store'
import { RoleDefaultRoute } from '@/types'
import type { LoginFormValues } from '@/types/presentation'

/**
 * 平台特性列表
 */
const features = [
  { icon: <Shield className="w-5 h-5" />, text: '多租户架构，数据安全隔离' },
  { icon: <Container className="w-5 h-5" />, text: '容器化实验环境，一键启动' },
  { icon: <Link className="w-5 h-5" />, text: '覆盖主流区块链技术栈' },
  { icon: <Trophy className="w-5 h-5" />, text: 'CTF竞赛系统，解题+对抗' },
]

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login, isLoggedIn, user, loading } = useUserStore()
  const [form] = Form.useForm<LoginFormValues>()

  // 如果已登录，跳转到对应页面
  useEffect(() => {
    if (isLoggedIn && user) {
      const from = (location.state as { from?: { pathname: string } })?.from?.pathname
      const defaultRoute = RoleDefaultRoute[user.role]
      navigate(from || defaultRoute, { replace: true })
    }
  }, [isLoggedIn, user, navigate, location])

  // 处理登录
  const handleLogin = async (values: LoginFormValues) => {
    try {
      await login(values.phone, values.password)
      message.success('登录成功')
    } catch {
      // 错误由拦截器处理
    }
  }

  return (
    <div className="bg-gradient-login min-h-screen flex relative overflow-hidden">
      {/* 装饰性渐变光斑，增强科技感 */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="gradient-animated w-72 h-72 rounded-full blur-3xl opacity-50 absolute -left-10 -top-12 float-soft" />
        <div className="gradient-animated w-64 h-64 rounded-full blur-3xl opacity-40 absolute right-4 top-10 float-soft" style={{ animationDelay: '1.2s' }} />
      </div>

      {/* 左侧：平台介绍 */}
      <div className="hidden lg:flex lg:w-3/5 flex-col justify-center items-center p-12 fade-in">
        <div className="max-w-lg">
          {/* Logo 和标题 */}
          <div className="flex items-center mb-6">
            <div className="w-12 h-12 rounded-xl bg-primary flex items-center justify-center mr-4">
              <Link className="w-7 h-7 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-primary">链境 ChainSpace</h1>
              <p className="text-text-secondary">区块链技术教学与实验平台</p>
            </div>
          </div>

          {/* 平台特性 */}
          <div className="mt-12 space-y-6">
            {features.map((feature, index) => (
              <div key={index} className="flex items-center text-text-primary">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center mr-4 text-primary">
                  {feature.icon}
                </div>
                <span className="text-base">{feature.text}</span>
              </div>
            ))}
          </div>

          {/* 装饰性链路卡片 - 替换SVG，提升质感 */}
          <div className="mt-16 grid grid-cols-2 sm:grid-cols-4 gap-4">
            {['创世区块', '验证区块', '交易区块', '最新区块'].map((title, idx) => (
              <div
                key={title}
                className="glass-card relative p-4 overflow-hidden fade-in"
                style={{ animationDelay: `${idx * 0.05}s` }}
              >
                <div className="absolute inset-0 gradient-animated opacity-60" />
                <div className="relative flex flex-col">
                  <span className="text-xs text-text-secondary"># {idx}</span>
                  <span className="text-lg font-semibold text-text-primary mt-1">{title}</span>
                  <span className="text-xs text-text-secondary mt-2">Txs: {8 + idx * 3}</span>
                  <span className="text-xs text-text-secondary">哈希：0x8b...{idx}c</span>
                  <div className="mt-3 h-1.5 w-full bg-white/30 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary float-soft"
                      style={{ width: `${40 + idx * 15}%` }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* 右侧：登录表单 */}
      <div className="w-full lg:w-2/5 flex items-center justify-center p-8 fade-in" style={{ animationDelay: '0.15s' }}>
        <div className="w-full max-w-md">
          <div className="card glass-card relative overflow-hidden">
            <div className="absolute inset-x-6 top-0 h-1 rounded-full gradient-animated" />

            {/* 移动端 Logo */}
            <div className="lg:hidden flex items-center justify-center mb-8">
              <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center mr-3">
                <Link className="w-6 h-6 text-white" />
              </div>
              <span className="text-xl font-bold text-text-primary">链境 ChainSpace</span>
            </div>

            <h2 className="text-2xl font-semibold text-center mb-8 text-text-primary">
              欢迎登录
            </h2>

            <Form
              form={form}
              name="login"
              onFinish={handleLogin}
              initialValues={{ remember: true }}
              size="large"
              className="space-y-2"
            >
              <Form.Item
                name="phone"
                rules={[
                  { required: true, message: '请输入手机号' },
                  { pattern: /^1\d{10}$/, message: '请输入正确的手机号' },
                ]}
              >
                <Input
                  prefix={<UserOutlined className="text-text-secondary" />}
                  placeholder="请输入手机号"
                  maxLength={11}
                  className="rounded-lg"
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
              >
                <Input.Password
                  prefix={<LockOutlined className="text-text-secondary" />}
                  placeholder="请输入密码"
                  className="rounded-lg"
                />
              </Form.Item>

              <Form.Item>
                <div className="flex justify-between items-center">
                  <Form.Item name="remember" valuePropName="checked" noStyle>
                    <Checkbox>记住我</Checkbox>
                  </Form.Item>
                  <a className="text-primary hover:text-primary-dark">
                    忘记密码？
                  </a>
                </div>
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  className="w-full"
                  shape="round"
                  loading={loading}
                >
                  登 录
                </Button>
              </Form.Item>
            </Form>

            {/* 底部说明 */}
            <div className="text-center text-text-secondary text-sm mt-6">
              <p>请使用学校分配的账号登录</p>
              <p className="mt-1">如有问题请联系管理员</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
