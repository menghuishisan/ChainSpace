import { useEffect, useMemo, useState } from 'react'
import { DatePicker, Divider, Form, InputNumber, Select, Switch } from 'antd'
import type { DockerImageCapability } from '@/types'

import {
  BATTLE_RESOURCE_MODEL_OPTIONS,
  BATTLE_SCORING_MODEL_OPTIONS,
  BATTLE_STRATEGY_OPTIONS,
} from '@/domains/contest/management'
import { getImageCapabilities } from '@/api/admin'
import {
  capabilitySupportsAllTools,
  capabilitySupportsMode,
  describeCapability,
  getCapabilityImageRef,
} from '@/domains/runtime/imageCapabilities'

const { RangePicker } = DatePicker

export default function ContestBattleConfigFields() {
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

  const sharedChainOptions = useMemo(() => imageCapabilities
    .filter((item) => capabilitySupportsMode(item, ['single_chain_instance', 'multi_service_lab']))
    .filter((item) => capabilitySupportsAnySharedChainTool(item))
    .map((item) => ({ label: getCapabilityImageRef(item), value: getCapabilityImageRef(item), title: describeCapability(item) })), [imageCapabilities])

  const judgeImageOptions = useMemo(() => imageCapabilities
    .filter((item) => capabilitySupportsMode(item, ['single_user']))
    .filter((item) => capabilitySupportsAllTools(item, ['logs']))
    .map((item) => ({ label: getCapabilityImageRef(item), value: getCapabilityImageRef(item), title: describeCapability(item) })), [imageCapabilities])

  const workspaceOptions = useMemo(() => imageCapabilities
    .filter((item) => capabilitySupportsMode(item, ['single_user', 'single_chain_instance']))
    .filter((item) => capabilitySupportsAllTools(item, ['ide', 'terminal', 'files']))
    .map((item) => ({ label: getCapabilityImageRef(item), value: getCapabilityImageRef(item), title: describeCapability(item) })), [imageCapabilities])

  return (
    <>
      <Form.Item
        name="time_range"
        label="比赛时间"
        rules={[{ required: true, message: '请选择比赛时间' }]}
      >
        <RangePicker showTime format="YYYY-MM-DD HH:mm" className="w-full" />
      </Form.Item>

      <div className="grid grid-cols-2 gap-4">
        <Form.Item
          name="registration_start"
          label="报名开始时间"
          dependencies={['time_range']}
          rules={[
            ({ getFieldValue }) => ({
              validator(_, value) {
                const timeRange = getFieldValue('time_range')
                const contestStart = timeRange?.[0]
                if (!value || !contestStart || !value.isAfter(contestStart)) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error('报名开始时间不能晚于比赛开始时间'))
              },
            }),
          ]}
        >
          <DatePicker showTime format="YYYY-MM-DD HH:mm" className="w-full" />
        </Form.Item>

        <Form.Item
          name="registration_end"
          label="报名截止时间"
          dependencies={['time_range', 'registration_start']}
          rules={[
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value) {
                  return Promise.resolve()
                }

                const timeRange = getFieldValue('time_range')
                const contestStart = timeRange?.[0]
                const registrationStart = getFieldValue('registration_start')

                if (contestStart && value.isAfter(contestStart)) {
                  return Promise.reject(new Error('报名截止时间不能晚于比赛开始时间'))
                }

                if (registrationStart && value.isBefore(registrationStart)) {
                  return Promise.reject(new Error('报名结束时间不能早于报名开始时间'))
                }

                return Promise.resolve()
              },
            }),
          ]}
        >
          <DatePicker showTime format="YYYY-MM-DD HH:mm" className="w-full" />
        </Form.Item>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Form.Item name="max_participants" label="最大参赛人数">
          <InputNumber min={0} className="w-full" placeholder="留空或 0 表示不限制" />
        </Form.Item>
        <Form.Item name="team_size_min" label="最小队伍人数">
          <InputNumber min={1} className="w-full" />
        </Form.Item>
        <Form.Item name="team_size_max" label="最大队伍人数">
          <InputNumber min={1} className="w-full" />
        </Form.Item>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <Form.Item name="dynamic_score" label="启用动态计分" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="first_blood_bonus" label="一血奖励">
          <InputNumber min={0} className="w-full" />
        </Form.Item>
      </div>

      <Form.Item noStyle shouldUpdate={(prev, next) => prev.type !== next.type}>
        {({ getFieldValue }) => getFieldValue('type') === 'agent_battle' ? (
          <>
            <Divider>对抗赛规则配置</Divider>

            <div className="grid grid-cols-3 gap-4">
              <Form.Item
                name={['battle_orchestration', 'shared_chain', 'image']}
                label="共享链镜像"
                rules={[{ required: true, message: '请选择共享链镜像' }]}
              >
                <Select options={sharedChainOptions} showSearch optionFilterProp="label" />
              </Form.Item>
              <Form.Item
                name={['battle_orchestration', 'judge', 'image']}
                label="裁判镜像"
              >
                <Select options={judgeImageOptions} showSearch optionFilterProp="label" allowClear />
              </Form.Item>
              <Form.Item
                name={['battle_orchestration', 'team_workspace', 'image']}
                label="队伍工作区镜像"
                rules={[{ required: true, message: '请选择队伍工作区镜像' }]}
              >
                <Select options={workspaceOptions} showSearch optionFilterProp="label" />
              </Form.Item>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <Form.Item
                name={['battle_orchestration', 'judge', 'strategy_interface']}
                label="策略接口"
                rules={[{ required: true, message: '请选择策略接口' }]}
              >
                <Select options={BATTLE_STRATEGY_OPTIONS} />
              </Form.Item>
              <Form.Item
                name={['battle_orchestration', 'judge', 'resource_model']}
                label="资源模型"
                rules={[{ required: true, message: '请选择资源模型' }]}
              >
                <Select options={BATTLE_RESOURCE_MODEL_OPTIONS} />
              </Form.Item>
              <Form.Item
                name={['battle_orchestration', 'judge', 'scoring_model']}
                label="评分模型"
                rules={[{ required: true, message: '请选择评分模型' }]}
              >
                <Select options={BATTLE_SCORING_MODEL_OPTIONS} />
              </Form.Item>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Form.Item
                name={['battle_orchestration', 'team_workspace', 'interaction_tools']}
                label="队伍工作区工具"
                rules={[{ required: true, message: '请选择工作区工具' }]}
              >
                <Select
                  mode="multiple"
                  options={[
                    { label: 'IDE', value: 'ide' },
                    { label: '终端', value: 'terminal' },
                    { label: '文件', value: 'files' },
                    { label: '日志', value: 'logs' },
                    { label: '区块浏览器', value: 'explorer' },
                    { label: 'API 调试', value: 'api_debug' },
                    { label: '可视化', value: 'visualization' },
                    { label: '组网面板', value: 'network' },
                    { label: 'RPC', value: 'rpc' },
                  ]}
                />
              </Form.Item>
              <Form.Item
                name={['battle_orchestration', 'team_workspace', 'resources', 'cpu']}
                label="工作区 CPU"
              >
                <Select
                  options={[
                    { label: '1', value: '1' },
                    { label: '2', value: '2' },
                    { label: '4', value: '4' },
                  ]}
                />
              </Form.Item>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <Form.Item name={['battle_orchestration', 'lifecycle', 'round_duration_seconds']} label="单轮时长（秒）">
                <InputNumber min={60} className="w-full" />
              </Form.Item>
              <Form.Item name={['battle_orchestration', 'lifecycle', 'upgrade_window_seconds']} label="升级窗口（秒）">
                <InputNumber min={30} className="w-full" />
              </Form.Item>
              <Form.Item name={['battle_orchestration', 'lifecycle', 'total_rounds']} label="总轮数">
                <InputNumber min={1} className="w-full" />
              </Form.Item>
            </div>

            <div className="grid grid-cols-4 gap-4">
              <Form.Item name={['battle_orchestration', 'judge', 'score_weights', 'resource']} label="资源分权重">
                <InputNumber min={0} className="w-full" />
              </Form.Item>
              <Form.Item name={['battle_orchestration', 'judge', 'score_weights', 'attack']} label="攻击分权重">
                <InputNumber min={0} className="w-full" />
              </Form.Item>
              <Form.Item name={['battle_orchestration', 'judge', 'score_weights', 'defense']} label="防御分权重">
                <InputNumber min={0} className="w-full" />
              </Form.Item>
              <Form.Item name={['battle_orchestration', 'judge', 'score_weights', 'survival']} label="生存分权重">
                <InputNumber min={0} className="w-full" />
              </Form.Item>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Form.Item name={['battle_orchestration', 'spectate', 'enable_monitor']} label="启用实时观战" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item name={['battle_orchestration', 'spectate', 'enable_replay']} label="启用赛后回放" valuePropName="checked">
                <Switch />
              </Form.Item>
            </div>
          </>
        ) : null}
      </Form.Item>
    </>
  )
}

function capabilitySupportsAnySharedChainTool(capability: DockerImageCapability): boolean {
  return capabilitySupportsAllTools(capability, ['rpc']) || capabilitySupportsAllTools(capability, ['network'])
}
