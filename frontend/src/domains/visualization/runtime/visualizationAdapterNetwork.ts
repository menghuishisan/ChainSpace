import type {
  NetworkStats,
  SimulatorState,
  VisualizationRecord,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import {
  asArray,
  asNumber,
  asRecord,
  asString,
  getEvents,
  getGlobalData,
  getNodes,
} from './visualizationAdapterCommon'
import { getVisualizationEventLabel } from './visualizationEventLabels'

export function buildNetworkVisualizationData(
  state: SimulatorState,
  runtime: VisualizationRuntimeSpec,
): VisualizationRecord {
  const globalData = getGlobalData(state)
  const nodes = getNodes(state).map((node, index) => {
    const nodeData = asRecord(node.data)
    return {
      id: node.id,
      label: asString(nodeData.name, runtime.scenario === 'discovery' ? `节点 ${index + 1}` : `拓扑节点 ${index + 1}`),
      type: asString(nodeData.type, index === 0 ? 'bootstrap' : 'full'),
      status: node.status || 'online',
      peers: asArray<string>(nodeData.peers),
      latency: asNumber(nodeData.latency, 40 + index * 5),
      partition: asString(nodeData.partition, ''),
    }
  })
  const events = getEvents(state)

  return {
    scenario: runtime.scenario,
    nodes,
    messages: events.map((event, index) => ({
      id: event.id || `message-${index}`,
      from: asString(event.source, nodes[0]?.id || 'node-1'),
      to: asString(event.target, nodes[1]?.id || 'node-2'),
      type: event.type,
      progress: 100,
      description: getVisualizationEventLabel(event.type),
    })),
    stats: {
      totalNodes: nodes.length,
      onlineNodes: nodes.filter((node) => node.status !== 'offline').length,
      avgLatency: nodes.length > 0 ? nodes.reduce((sum, node) => sum + node.latency, 0) / nodes.length : 0,
      sessionCount: asNumber(globalData.session_count),
      edgeCount: asNumber(globalData.edge_count),
      connectivity: runtime.scenario === 'discovery'
        ? (asNumber(globalData.session_count) > 0 ? '发现进行中' : '等待新节点加入')
        : (Boolean(globalData.is_connected) ? '连通' : '可能断开'),
    } satisfies NetworkStats,
  }
}
