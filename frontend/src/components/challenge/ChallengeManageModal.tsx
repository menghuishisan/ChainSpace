import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Divider, Form, Input, InputNumber, Modal, Select, Space, Switch } from 'antd'
import type { DockerImageCapability } from '@/types'
import type { ChallengeManageModalProps } from '@/types/presentation'
import { CategoryMap, DifficultyMap } from '@/types'
import {
  CHALLENGE_RUNTIME_PROFILE_OPTIONS,
  CHALLENGE_SERVICE_OPTIONS,
  getDefaultRuntimeProfile,
} from '@/domains/challenge/orchestration'
import { buildChallengePresetValues } from '@/domains/challenge/management'
import { getImageCapabilities } from '@/api/admin'
import {
  capabilitySupportsAllTools,
  capabilitySupportsMode,
  describeCapability,
  getCapabilityImageRef,
} from '@/domains/runtime/imageCapabilities'

export default function ChallengeManageModal({
  open,
  loading,
  editingChallenge,
  form,
  allowDirectPublic = false,
  onCancel,
  onSubmit,
}: ChallengeManageModalProps) {
  const isPublic = Form.useWatch('is_public', form)
  const runtimeProfile = Form.useWatch('runtime_profile', form)
  const [imageCapabilities, setImageCapabilities] = useState<DockerImageCapability[]>([])

  useEffect(() => {
    let disposed = false
    const load = async () => {
      try {
        const list = await getImageCapabilities()
        if (!disposed) {
          setImageCapabilities(list || [])
        }
      } catch {
        if (!disposed) {
          setImageCapabilities([])
        }
      }
    }
    void load()
    return () => {
      disposed = true
    }
  }, [])

  const workspaceModeRequirement = useMemo(() => {
    switch (runtimeProfile) {
      case 'fork_replay':
        return ['fork_replay', 'single_chain_instance']
      case 'multi_service_lab':
        return ['multi_service_lab', 'single_user_multi_node']
      case 'static':
        return ['single_user']
      default:
        return ['single_chain_instance']
    }
  }, [runtimeProfile])

  const workspaceOptions = useMemo(() => imageCapabilities
    .filter((item) => capabilitySupportsAllTools(item, ['ide', 'terminal', 'files']))
    .filter((item) => capabilitySupportsMode(item, workspaceModeRequirement))
    .map((item) => ({
      label: getCapabilityImageRef(item),
      value: getCapabilityImageRef(item),
      title: describeCapability(item),
    })), [imageCapabilities, workspaceModeRequirement])

  return (
    <Modal
      title={editingChallenge ? '编辑题目' : '创建题目'}
      open={open}
      onCancel={onCancel}
      footer={null}
      width={760}
    >
      <Form form={form} layout="vertical" onFinish={onSubmit} className="mt-4">
        <Form.Item name="title" label="题目名称" rules={[{ required: true, message: '请输入题目名称' }]}>
          <Input placeholder="请输入题目名称" />
        </Form.Item>

        <div className="grid grid-cols-4 gap-4">
          <Form.Item name="category" label="知识分类" rules={[{ required: true, message: '请选择分类' }]}>
            <Select
              options={Object.entries(CategoryMap).map(([value, label]) => ({ value, label }))}
              onChange={(value) => {
                const runtimeProfile = getDefaultRuntimeProfile(value)
                form.setFieldsValue(buildChallengePresetValues(value, runtimeProfile))
              }}
            />
          </Form.Item>
          <Form.Item name="runtime_profile" label="环境类型" rules={[{ required: true, message: '请选择环境类型' }]}>
            <Select
              options={CHALLENGE_RUNTIME_PROFILE_OPTIONS}
              onChange={(value) => {
                const category = form.getFieldValue('category') || 'contract_vuln'
                form.setFieldsValue(buildChallengePresetValues(category, value))
              }}
            />
          </Form.Item>
          <Form.Item name="difficulty" label="难度" rules={[{ required: true, message: '请选择难度' }]}>
            <Select options={[1, 2, 3, 4, 5].map((value) => ({ value, label: DifficultyMap[value].text }))} />
          </Form.Item>
          <Form.Item name="base_points" label="基础分" rules={[{ required: true, message: '请输入基础分' }]}>
            <InputNumber min={1} className="w-full" />
          </Form.Item>
          <Form.Item name="min_points" label="最低分">
            <InputNumber min={1} className="w-full" />
          </Form.Item>
          <Form.Item name="decay_factor" label="衰减系数">
            <InputNumber min={0} max={1} step={0.05} className="w-full" />
          </Form.Item>
        </div>

        <Form.Item name="description" label="题目描述" rules={[{ required: true, message: '请输入题目描述' }]}>
          <Input.TextArea rows={4} placeholder="支持 Markdown，用于描述背景、目标和限制" />
        </Form.Item>

        <div className="grid grid-cols-2 gap-4">
          <Form.Item name="flag_type" label="Flag 类型" rules={[{ required: true, message: '请选择 Flag 类型' }]}>
            <Select options={[{ label: '静态 Flag', value: 'static' }, { label: '动态 Flag', value: 'dynamic' }]} />
          </Form.Item>
          <Form.Item name="flag_template" label="Flag 模板">
            <Input placeholder="静态题填写固定 Flag，环境题建议使用动态 Flag" />
          </Form.Item>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <Form.Item
            name="is_public"
            label="公开题目"
            valuePropName="checked"
            extra={allowDirectPublic
              ? undefined
              : isPublic
                ? '非平台管理员可以将题目改回不公开；若要重新公开，需走公开申请审核。'
                : '当前不能直接改成公开，需先保存为非公开后提交公开申请。'}
          >
            <Switch disabled={!allowDirectPublic && !isPublic} />
          </Form.Item>
          {editingChallenge ? (
            <Form.Item name="status" label="题目状态">
              <Select options={[
                { label: '草稿', value: 'draft' },
                { label: '启用', value: 'active' },
                { label: '归档', value: 'archived' },
              ]}
              />
            </Form.Item>
          ) : <div />}
        </div>

        <Divider>运行环境配置</Divider>

        <div className="grid grid-cols-2 gap-4">
          <Form.Item name={['challenge_orchestration', 'needs_environment']} label="是否需要环境" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name={['challenge_orchestration', 'workspace', 'image']} label="工作区镜像" rules={[{ required: true, message: '请选择工作区镜像' }]}>
            <Select
              options={workspaceOptions}
              placeholder="根据运行形态自动过滤可编排镜像"
              showSearch
              optionFilterProp="label"
            />
          </Form.Item>
        </div>
        {workspaceOptions.length === 0 ? (
          <Alert
            type="warning"
            showIcon
            className="mb-4"
            message="当前没有匹配镜像"
            description="请先在镜像管理中登记并启用支持对应能力的镜像，再创建该运行形态题目。"
          />
        ) : null}

        <Form.Item name="service_keys" label="附加服务">
          <Select mode="multiple" options={CHALLENGE_SERVICE_OPTIONS} placeholder="选择需要一起启动的服务组件" />
        </Form.Item>

        <div className="grid grid-cols-3 gap-4">
          <Form.Item name={['challenge_orchestration', 'fork', 'enabled']} label="启用 Fork 复现" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name={['challenge_orchestration', 'fork', 'chain']} label="目标链">
            <Input placeholder="如 ethereum / bsc / arbitrum" />
          </Form.Item>
          <Form.Item name={['challenge_orchestration', 'fork', 'block_number']} label="Fork 区块高度">
            <InputNumber min={0} className="w-full" />
          </Form.Item>
        </div>

        <Form.Item name={['challenge_orchestration', 'fork', 'rpc_url']} label="Fork RPC 地址">
          <Input placeholder="题目运行使用的 Fork RPC 地址" />
        </Form.Item>

        <Form.Item name={['challenge_orchestration', 'scenario', 'attack_goal']} label="题目目标">
          <Input.TextArea rows={3} placeholder="说明参赛者需要复现的攻击行为或达成的控制目标" />
        </Form.Item>

        <div className="grid grid-cols-2 gap-4">
          <Form.Item name={['challenge_orchestration', 'lifecycle', 'time_limit_minutes']} label="环境时长（分钟）">
            <InputNumber min={10} max={480} className="w-full" />
          </Form.Item>
          <Form.Item name={['challenge_orchestration', 'lifecycle', 'auto_destroy']} label="超时自动回收" valuePropName="checked">
            <Switch />
          </Form.Item>
        </div>

        <Form.Item name="contract_code" label="合约代码">
          <Input.TextArea rows={6} placeholder="可选，适用于真实漏洞源码或教学样例代码" />
        </Form.Item>

        <Form.Item name="setup_code" label="初始化代码">
          <Input.TextArea rows={4} placeholder="题目初始化代码或环境预置脚本" />
        </Form.Item>

        <Form.Item name="deploy_script" label="部署脚本">
          <Input.TextArea rows={4} placeholder="部署合约、初始化状态或启动服务的脚本" />
        </Form.Item>

        <Form.Item name="check_script" label="校验脚本">
          <Input.TextArea rows={4} placeholder="用于验证解题结果的脚本" />
        </Form.Item>

        <Form.List name="hints">
          {(fields, { add, remove }) => (
            <div className="space-y-3">
              <Divider>提示配置</Divider>
              {fields.map((field) => (
                <div key={field.key} className="grid grid-cols-[1fr_140px_auto] gap-3">
                  <Form.Item {...field} name={[field.name, 'content']} label="提示内容" rules={[{ required: true, message: '请输入提示内容' }]}>
                    <Input placeholder="提示内容" />
                  </Form.Item>
                  <Form.Item {...field} name={[field.name, 'cost']} label="消耗分值" rules={[{ required: true, message: '请输入消耗分值' }]}>
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                  <Form.Item label=" ">
                    <Button danger onClick={() => remove(field.name)}>删除</Button>
                  </Form.Item>
                </div>
              ))}
              <Button type="dashed" onClick={() => add({ content: '', cost: 0 })}>
                添加提示
              </Button>
            </div>
          )}
        </Form.List>

        <Form.Item name="attachments" label="附件链接">
          <Select mode="tags" tokenSeparators={[',']} placeholder="输入附件 URL，回车确认" />
        </Form.Item>

        <Form.Item name="tags" label="标签">
          <Select mode="tags" tokenSeparators={[',']} placeholder="输入标签，回车确认" />
        </Form.Item>

        <Form.Item className="mb-0 text-right">
          <Space>
            <Button onClick={onCancel}>取消</Button>
            <Button type="primary" htmlType="submit" loading={loading}>
              {editingChallenge ? '保存' : '创建'}
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  )
}
