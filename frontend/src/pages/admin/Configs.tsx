/**
 * 平台管理员 - 系统配置页面
 */
import { useState, useEffect } from 'react'
import { Card, Form, Input, Button, message, Spin } from 'antd'
import { SaveOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import type { SystemConfig } from '@/types'
import { getConfigs, updateConfig } from '@/api/admin'

export default function Configs() {
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const [form] = Form.useForm()

  // 获取配置
  useEffect(() => {
    const fetchConfigs = async () => {
      setLoading(true)
      try {
        const data = await getConfigs()
        setConfigs(data)
        // 设置表单初始值
        const formValues: Record<string, string> = {}
        data.forEach((config) => {
          formValues[config.key] = config.value
        })
        form.setFieldsValue(formValues)
      } catch {
        // 错误由拦截器处理
      } finally {
        setLoading(false)
      }
    }
    fetchConfigs()
  }, [form])

  // 保存配置
  const handleSave = async () => {
    const values = form.getFieldsValue()
    setSaving(true)
    try {
      // 逐个更新配置
      for (const [key, value] of Object.entries(values)) {
        await updateConfig(key, value as string)
      }
      message.success('配置已保存')
    } catch {
      // 错误由拦截器处理
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <PageHeader
        title="系统配置"
        subtitle="管理平台的全局系统配置"
        extra={
          <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
            保存配置
          </Button>
        }
      />

      <Spin spinning={loading}>
        <Card>
          <Form form={form} layout="vertical">
            {configs.map((config) => (
              <Form.Item
                key={config.key}
                name={config.key}
                label={config.description || config.key}
                extra={config.key}
              >
                <Input placeholder={`请输入${config.description || config.key}`} />
              </Form.Item>
            ))}

            {configs.length === 0 && !loading && (
              <div className="text-center text-text-secondary py-8">
                暂无配置项
              </div>
            )}
          </Form>
        </Card>
      </Spin>
    </div>
  )
}
