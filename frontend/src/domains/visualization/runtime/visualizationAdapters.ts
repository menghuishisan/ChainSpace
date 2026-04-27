import type { SimulatorState, VisualizationRuntimeSpec } from '@/types/visualizationDomain'
import { buildAttackVisualizationData } from './visualizationAdapterAttack'
import { buildBlockchainVisualizationData } from './visualizationAdapterBlockchain'
import { buildConsensusVisualizationData } from './visualizationAdapterConsensus'
import { buildCrossChainVisualizationData } from './visualizationAdapterCrosschain'
import { buildCryptoVisualizationData } from './visualizationAdapterCrypto'
import { buildDeFiVisualizationData } from './visualizationAdapterDeFi'
import { buildEvmVisualizationData } from './visualizationAdapterEvm'
import { buildNetworkVisualizationData } from './visualizationAdapterNetwork'

/**
 * 将 simulations 原始状态适配成前端各领域画布可直接消费的结构。
 * 这里仅保留统一分发，不再承载具体领域实现细节。
 */
export function adaptVisualizationState(state: SimulatorState, runtime: VisualizationRuntimeSpec): SimulatorState {
  switch (runtime.renderer) {
    case 'consensus':
      return { ...state, data: buildConsensusVisualizationData(state, runtime) }
    case 'attack':
      return { ...state, data: buildAttackVisualizationData(state, runtime) }
    case 'blockchain':
      return { ...state, data: buildBlockchainVisualizationData(state, runtime) }
    case 'network':
      return { ...state, data: buildNetworkVisualizationData(state, runtime) }
    case 'crypto':
      return { ...state, data: buildCryptoVisualizationData(state) }
    case 'crosschain':
      return { ...state, data: buildCrossChainVisualizationData(state) }
    case 'evm':
      return { ...state, data: buildEvmVisualizationData(state) }
    case 'defi':
      return { ...state, data: buildDeFiVisualizationData(state, runtime) }
    default:
      return state
  }
}

/**
 * 将模拟器状态整理成便于排查的 JSON 文本。
 */
export function normalizeSimulationState(state: SimulatorState | null): string {
  return state ? JSON.stringify(state, null, 2) : '暂无模拟状态数据。'
}
