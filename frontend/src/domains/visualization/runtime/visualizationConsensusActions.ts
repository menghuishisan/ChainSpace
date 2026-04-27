import type {
  VisualizationActionDefinition,
  VisualizationActionField,
  VisualizationActionKind,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'

function action(definition: VisualizationActionDefinition): VisualizationActionDefinition {
  return definition
}

function field(definition: VisualizationActionField): VisualizationActionField {
  return definition
}

function kind(value: VisualizationActionKind): VisualizationActionKind {
  return value
}

function buildNodeOptions(count = 8) {
  return Array.from({ length: count }, (_, index) => ({
    label: `node-${index}`,
    value: `node-${index}`,
  }))
}

function buildValidatorOptions(count = 16) {
  return Array.from({ length: count }, (_, index) => ({
    label: `validator-${index}`,
    value: `validator-${index}`,
  }))
}

function buildDelegateOptions(count = 24) {
  return Array.from({ length: count }, (_, index) => ({
    label: `delegate-${index}`,
    value: `delegate-${index}`,
  }))
}

function buildMinerOptions(count = 10) {
  return Array.from({ length: count }, (_, index) => ({
    label: `miner-${index}`,
    value: `miner-${index}`,
  }))
}

/**
 * 统一维护共识类模块的教学动作定义。
 * 这里按算法语义组织，而不是继续分散到多个零散文件里。
 */
export function getConsensusActions(
  runtime: VisualizationRuntimeSpec,
  allowFaultInjection: boolean,
): VisualizationActionDefinition[] {
  const algorithm = runtime.algorithm

  switch (algorithm) {
    case 'pbft':
      return [
        action({
          key: 'pbft-submit-request',
          label: '发起请求',
          description: '向主节点提交新的客户端请求，观察请求到提交的完整推进。',
          kind: kind('module_action'),
          action: 'submit_request',
          successMessage: '已发起新的客户端请求。',
        }),
        action({
          key: 'pbft-view-change',
          label: '触发视图切换',
          description: '模拟主节点失效后切换到新的视图与主节点。',
          kind: kind('module_action'),
          action: 'trigger_view_change',
          successMessage: '已触发 PBFT 视图切换。',
        }),
        ...(allowFaultInjection
          ? [
              action({
                key: 'pbft-byzantine',
                label: '注入拜占庭节点',
                description: '将选定节点切换为异常行为，观察 Prepare 和 Commit 是否还能达到阈值。',
                kind: kind('inject_fault'),
                successMessage: '已注入拜占庭故障。',
                preset: { type: 'byzantine', duration: 0 },
                fields: [
                  field({
                    key: 'target',
                    label: '目标节点',
                    type: 'select',
                    defaultValue: 'node-3',
                    options: buildNodeOptions(),
                  }),
                  field({
                    key: 'behavior',
                    label: '异常行为',
                    type: 'select',
                    defaultValue: 'ignore',
                    options: [
                      { label: '不响应', value: 'ignore' },
                      { label: '双重投票', value: 'double_vote' },
                      { label: '跳过阶段', value: 'skip_phase' },
                    ],
                  }),
                ],
              }),
            ]
          : []),
      ]

    case 'raft':
      return [
        action({
          key: 'raft-submit-command',
          label: '提交命令',
          description: '向当前领导者提交一条日志命令，观察复制与提交索引变化。',
          kind: kind('module_action'),
          action: 'submit_command',
          successMessage: '已提交新的日志命令。',
          fields: [
            field({
              key: 'command',
              label: '命令内容',
              type: 'text',
              defaultValue: 'set x=1',
            }),
          ],
        }),
        action({
          key: 'raft-trigger-election',
          label: '触发选举',
          description: '让指定节点立刻发起一轮新的领导者选举。',
          kind: kind('module_action'),
          action: 'trigger_election',
          successMessage: '已触发新一轮选举流程。',
          fields: [
            field({
              key: 'target',
              label: '候选节点',
              type: 'select',
              defaultValue: 'node-0',
              options: buildNodeOptions(),
            }),
          ],
        }),
        ...(allowFaultInjection
          ? [
              action({
                key: 'raft-fail-leader',
                label: '模拟领导者离线',
                description: '让当前领导者离线，观察心跳中断和重新选举。',
                kind: kind('module_action'),
                action: 'simulate_leader_failure',
                successMessage: '已让当前领导者离线。',
              }),
              action({
                key: 'raft-restore-node',
                label: '恢复节点',
                description: '恢复指定节点在线，观察它重新追赶日志。',
                kind: kind('module_action'),
                action: 'restore_node',
                successMessage: '已恢复节点在线状态。',
                fields: [
                  field({
                    key: 'target',
                    label: '恢复节点',
                    type: 'select',
                    defaultValue: 'node-0',
                    options: buildNodeOptions(),
                  }),
                ],
              }),
            ]
          : []),
      ]

    case 'pos':
      return [
        action({
          key: 'pos-force-epoch',
          label: '推进到下一 Epoch',
          description: '立刻执行一次 Epoch 切换，观察活跃验证者和最终确认高度变化。',
          kind: kind('module_action'),
          action: 'advance_epoch',
          successMessage: '已推进到下一 Epoch。',
        }),
        action({
          key: 'pos-slash-validator',
          label: '惩罚验证者',
          description: '对指定验证者执行 Slash，观察其是否退出活跃集合。',
          kind: kind('module_action'),
          action: 'slash_validator',
          successMessage: '已执行 Slash 惩罚。',
          fields: [
            field({
              key: 'target',
              label: '目标验证者',
              type: 'select',
              defaultValue: 'validator-0',
              options: buildValidatorOptions(),
            }),
            field({
              key: 'reason',
              label: '惩罚原因',
              type: 'text',
              defaultValue: 'double-sign',
            }),
          ],
        }),
      ]

    case 'dpos':
      return [
        action({
          key: 'dpos-re-elect',
          label: '重新选举委托者',
          description: '重新统计委托票数并刷新活跃委托者集合。',
          kind: kind('module_action'),
          action: 're_elect_delegates',
          successMessage: '已重新选举活跃委托者。',
        }),
        action({
          key: 'dpos-cast-vote',
          label: '追加投票',
          description: '让指定投票者把票投给目标委托者，观察排名变化。',
          kind: kind('module_action'),
          action: 'cast_vote',
          successMessage: '已提交新的委托投票。',
          fields: [
            field({
              key: 'voter',
              label: '投票者编号',
              type: 'number',
              defaultValue: 0,
              min: 0,
              max: 99,
            }),
            field({
              key: 'delegate',
              label: '目标委托者',
              type: 'select',
              defaultValue: 'delegate-0',
              options: buildDelegateOptions(),
            }),
          ],
        }),
        action({
          key: 'dpos-force-block',
          label: '立即出块',
          description: '让当前轮值委托者立刻执行一次出块，便于观察轮值推进。',
          kind: kind('module_action'),
          action: 'produce_block_now',
          successMessage: '已触发当前轮值委托者出块。',
        }),
      ]

    case 'hotstuff':
      return [
        action({
          key: 'hotstuff-submit-proposal',
          label: '发起提案',
          description: '让当前领导者立刻提出一个新区块，观察 QC 链路推进。',
          kind: kind('module_action'),
          action: 'submit_proposal',
          successMessage: '已发起新的 HotStuff 提案。',
        }),
        action({
          key: 'hotstuff-rotate-leader',
          label: '轮换领导者',
          description: '直接切换到下一位领导者，观察视图推进与 QC 状态变化。',
          kind: kind('module_action'),
          action: 'rotate_leader',
          successMessage: '已轮换到下一位领导者。',
        }),
      ]

    case 'tendermint':
      return [
        action({
          key: 'tendermint-submit-proposal',
          label: '发起提案',
          description: '让当前提议者立刻发起区块提案，观察两轮投票推进。',
          kind: kind('module_action'),
          action: 'submit_proposal',
          successMessage: '已发起 Tendermint 提案。',
        }),
        action({
          key: 'tendermint-next-round',
          label: '进入下一轮',
          description: '强制结束当前轮次，观察新一轮提议者和投票流程。',
          kind: kind('module_action'),
          action: 'force_next_round',
          successMessage: '已进入下一轮。',
        }),
      ]

    case 'pow':
      return [
        action({
          key: 'pow-mine-block',
          label: '立即出块',
          description: '让指定矿工立刻尝试产出一个新区块，便于观察链头变化。',
          kind: kind('module_action'),
          action: 'mine_block_now',
          successMessage: '已触发矿工立即出块。',
          fields: [
            field({
              key: 'miner',
              label: '目标矿工',
              type: 'select',
              defaultValue: 'miner-0',
              options: buildMinerOptions(),
            }),
          ],
        }),
        ...(allowFaultInjection
          ? [
              action({
                key: 'pow-selfish',
                label: '开启自私挖矿',
                description: '让目标矿工隐藏区块，观察分叉竞争和链头切换。',
                kind: kind('inject_fault'),
                successMessage: '已开启自私挖矿。',
                preset: { type: 'selfish_mining', duration: 0 },
                fields: [
                  field({
                    key: 'target',
                    label: '目标矿工',
                    type: 'select',
                    defaultValue: 'miner-0',
                    options: buildMinerOptions(),
                  }),
                ],
              }),
              action({
                key: 'pow-51',
                label: '触发 51% 攻击',
                description: '提高目标矿工算力，观察主链竞争和重组风险。',
                kind: kind('inject_fault'),
                successMessage: '已触发 51% 攻击。',
                preset: { type: '51_percent', duration: 0 },
                fields: [
                  field({
                    key: 'target',
                    label: '攻击矿工',
                    type: 'select',
                    defaultValue: 'miner-0',
                    options: buildMinerOptions(),
                  }),
                ],
              }),
            ]
          : []),
      ]

    case 'vrf':
      return [
        action({
          key: 'vrf-run-election',
          label: '执行一轮抽签',
          description: '立刻执行一轮 VRF 选择，观察本轮入选者和随机种子变化。',
          kind: kind('module_action'),
          action: 'run_election',
          successMessage: '已执行一轮 VRF 抽签。',
        }),
        action({
          key: 'vrf-refresh-seed',
          label: '刷新随机种子',
          description: '刷新下一轮种子，观察入选集合是否随之变化。',
          kind: kind('module_action'),
          action: 'refresh_seed',
          successMessage: '已刷新随机种子。',
        }),
      ]

    case 'fork_choice':
      return [
        action({
          key: 'fork-choice-create-fork',
          label: '制造分叉',
          description: '从指定父块创建新的竞争分支，观察规范链头如何变化。',
          kind: kind('module_action'),
          action: 'create_fork',
          successMessage: '已创建新的分叉分支。',
          fields: [
            field({
              key: 'parent',
              label: '父块哈希',
              type: 'text',
              defaultValue: 'genesis',
            }),
          ],
        }),
        action({
          key: 'fork-choice-recompute',
          label: '重新选择主链',
          description: '立刻重新执行一次 fork-choice 规则并更新规范链头。',
          kind: kind('module_action'),
          action: 'recompute_canonical',
          successMessage: '已重新计算规范链头。',
        }),
      ]

    case 'dag':
      return [
        action({
          key: 'dag-create-vertex',
          label: '创建顶点',
          description: '让指定节点立刻创建一个新顶点，观察引用关系和 tips 变化。',
          kind: kind('module_action'),
          action: 'create_vertex',
          successMessage: '已创建新的 DAG 顶点。',
          fields: [
            field({
              key: 'creator',
              label: '创建节点',
              type: 'select',
              defaultValue: 'node-0',
              options: buildNodeOptions(),
            }),
          ],
        }),
        action({
          key: 'dag-confirm-vertices',
          label: '刷新确认状态',
          description: '立刻重新计算顶点确认，观察确认计数与最新顶点变化。',
          kind: kind('module_action'),
          action: 'update_confirmations',
          successMessage: '已刷新顶点确认状态。',
        }),
      ]

    default:
      return []
  }
}
