import type {
  SimulatorState,
  VisualizationEntityCard,
  VisualizationMetricCard,
  VisualizationRecord,
  VisualizationStateSection,
} from '@/types/visualizationDomain'
import {
  asArray,
  asNumber,
  asString,
  buildEventItems,
  formatNumber,
  getEvents,
  getGlobalData,
} from './visualizationAdapterCommon'

export function buildCrossChainVisualizationData(state: SimulatorState): VisualizationRecord {
  const globalData = getGlobalData(state)
  const events = getEvents(state)
  const transactions = asArray<VisualizationRecord>(globalData.recent_transactions)
  const pools = asArray<VisualizationRecord>(globalData.bridge_pools)
  const validators = asArray<VisualizationRecord>(globalData.validator_overview)
  const chains = asArray<VisualizationRecord>(globalData.chain_overview)
  const latestTx = transactions[0]
  const signedCount = asNumber(latestTx?.signatures)
  const requiredSigs = asNumber(latestTx?.required_sigs)
  const confirmations = asNumber(latestTx?.confirmations)
  const requiredConfirm = asNumber(latestTx?.required_confirm)
  const matchedPool = latestTx
    ? pools.find((pool) => (
      asString(pool.source_chain) === asString(latestTx.source_chain)
      && asString(pool.dest_chain) === asString(latestTx.dest_chain)
      && asString(pool.token) === asString(latestTx.token)
    ))
    : pools[0]
  const challengeEnd = asString(latestTx?.challenge_end)
  const challengeLabel = challengeEnd && !challengeEnd.startsWith('0001-01-01') ? challengeEnd : '不适用'

  const metrics: VisualizationMetricCard[] = [
    { key: 'bridge_type', label: '桥模式', value: asString(globalData.bridge_type, '--'), hint: '决定跨链采用锁定-铸造还是流动性桥等模式。' },
    { key: 'security_model', label: '安全模型', value: asString(globalData.security_model, '--'), hint: '决定是否依赖确认数、多签或挑战期。' },
    { key: 'transactions', label: '跨链请求数', value: String(asNumber(globalData.transaction_count)), hint: '本轮已经发起的跨链请求数量。' },
    { key: 'total_bridged', label: '累计跨链数量', value: asString(globalData.total_bridged, '0'), hint: '用于观察资产是否在桥上持续流动。' },
  ]

  const entities: VisualizationEntityCard[] = []
  if (latestTx) {
    entities.push({
      id: asString(latestTx.id, 'latest-tx'),
      title: '最近一笔跨链请求',
      subtitle: `${asString(latestTx.source_chain, '--')} -> ${asString(latestTx.dest_chain, '--')}`,
      status: asString(latestTx.status, 'unknown'),
      details: [
        { label: '用户', value: asString(latestTx.user, '--') },
        { label: '资产', value: `${asString(latestTx.amount, '--')} ${asString(latestTx.token, '')}`.trim() },
        { label: '源链交易', value: asString(latestTx.source_tx_hash, '--') },
        { label: '目标链交易', value: asString(latestTx.dest_tx_hash, '尚未执行') },
      ],
    })
  }
  entities.push(
    ...validators.slice(0, 3).map((validator, index) => ({
      id: `validator-${index}`,
      title: asString(validator.name, `验证者 ${index + 1}`),
      subtitle: '桥验证者',
      status: validator.is_active ? 'active' : 'inactive',
      details: [
        { label: '投票权重', value: `${formatNumber(asNumber(validator.voting_power) * 100, 2)}%` },
        { label: '已签名次数', value: String(asNumber(validator.signed_count)) },
        { label: '漏签次数', value: String(asNumber(validator.missed_count)) },
      ],
    })),
  )

  const sections: VisualizationStateSection[] = []
  if (latestTx) {
    sections.push({
      key: 'bridge_lifecycle',
      title: '本次跨链流程校验',
      items: [
        { label: '当前状态', value: asString(latestTx.status, '--') },
        { label: '确认进度', value: `${confirmations}/${requiredConfirm}` },
        { label: '签名进度', value: `${signedCount}/${requiredSigs}` },
        { label: '挑战期结束', value: challengeLabel },
      ],
    })
  }
  if (matchedPool) {
    sections.push({
      key: 'bridge_pool',
      title: '当前桥池快照',
      items: [
        { label: '池子', value: `${asString(matchedPool.source_chain, '--')} -> ${asString(matchedPool.dest_chain, '--')}` },
        { label: '源链锁定', value: asString(matchedPool.source_locked, '--') },
        { label: '目标链铸造', value: asString(matchedPool.dest_minted, '--') },
        { label: '累计交易数', value: String(asNumber(matchedPool.tx_count)) },
      ],
    })
  }
  if (chains.length >= 2) {
    sections.push({
      key: 'chains',
      title: '两端链参数',
      items: chains.slice(0, 2).flatMap((chain) => ([
        { label: `${asString(chain.chain_name, '--')} 最终确认`, value: `${asNumber(chain.finality_blocks)} 个区块` },
        { label: `${asString(chain.chain_name, '--')} 出块时间`, value: asString(chain.block_time, '--') },
      ])),
    })
  }

  return {
    title: '跨链桥演示',
    summary: '学生应当能直接看出一笔跨链请求是否真正完成了“源链锁定/确认 -> 验证者签名 -> 目标链执行”这条闭环，而不是只看到请求已发起。',
    metrics,
    entities,
    sections,
    events: buildEventItems(events),
    observationTips: [
      '先确认最近一笔请求现在处于哪个阶段，再看确认数和签名数是否真的满足门槛。',
      '如果状态已经完成，应该能同时看到目标链交易哈希或到账结果，而不是只停留在源链事件。',
      '不同桥模式和安全模型下，判断成功的标准不同，不能只看事件是否触发。',
    ],
  }
}
