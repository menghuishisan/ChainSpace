import type {
  DeFiStats,
  SimulatorState,
  SwapRecord,
  VisualizationRecord,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import {
  asNumber,
  asRecord,
  asString,
  buildDisturbances,
  buildEventItems,
  buildMetrics,
  buildSections,
  getEvents,
  getGlobalData,
  getNodes,
} from './visualizationAdapterCommon'
import { getVisualizationEventLabel } from './visualizationEventLabels'

/**
 * 将 DeFi 类模拟状态适配成前端主画布可直接消费的教学数据。
 * 当前优先覆盖 AMM 这类以流动性池和价格曲线为核心的场景。
 */
export function buildDeFiVisualizationData(
  state: SimulatorState,
  runtime: VisualizationRuntimeSpec,
): VisualizationRecord {
  const globalData = getGlobalData(state)
  const events = getEvents(state)
  const disturbances = buildDisturbances(globalData)
  const reserveA = asNumber(globalData.reserve_a)
  const reserveB = asNumber(globalData.reserve_b)
  const price = asNumber(globalData.price, reserveA > 0 ? reserveB / reserveA : 0)
  const constantProduct = asNumber(globalData.constant_product, reserveA * reserveB)
  const swaps = events
    .filter((event) => ['swap_executed', 'liquidity_added', 'liquidity_removed'].includes(event.type))
    .map<SwapRecord>((event, index) => {
      const eventData = asRecord(event.data)
      return {
        id: event.id || `swap-${index}`,
        type: event.type,
        title: getVisualizationEventLabel(event.type),
        tokenIn: asString(eventData.token_in, 'ETH'),
        tokenOut: asString(eventData.token_out, 'USDC'),
        amountIn: asNumber(eventData.amount_in, asNumber(eventData.amount_a)),
        amountOut: asNumber(eventData.amount_out, asNumber(eventData.amount_b)),
        priceImpact: asNumber(eventData.price_impact_percent, asNumber(eventData.price_impact)),
      }
    })

  return {
    pool: {
      pair: asString(globalData.pair, 'ETH / USDC'),
      reserveA,
      reserveB,
      price,
      constantProduct,
      feeRate: asNumber(globalData.fee_percent, 0.3) / 100,
      tvl: reserveA * price + reserveB,
    },
    curve: {
      points: Array.from({ length: 16 }, (_, index) => {
        const minX = Math.max((reserveA || 1) * 0.35, 1)
        const maxX = (reserveA || 1) * 2.4
        const x = minX + (maxX - minX) * (index / 15)
        return { x, y: constantProduct > 0 ? constantProduct / x : 0 }
      }),
      current: { x: reserveA, y: reserveB },
    },
    swaps,
    stats: {
      tvl: reserveA * price + reserveB,
      spotPrice: price,
      slippage: swaps[0]?.priceImpact || 0,
      eventCount: swaps.length,
    } satisfies DeFiStats,
    metrics: buildMetrics(state, runtime, globalData, events, getNodes(state)),
    sections: buildSections(state, globalData),
    events: buildEventItems(events),
    disturbances,
    observationTips: [
      '先看池子曲线和当前点位，理解价格为什么会随着储备变化而移动。',
      '再看最近一次交换或加减流动性的结果，确认输入、输出和价格影响是否一致。',
      '如果叠加了攻击或故障覆盖层，要重点观察价格、储备和滑点是否因此出现异常变化。',
    ],
  }
}
