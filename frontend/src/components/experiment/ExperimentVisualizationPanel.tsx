import { Suspense, lazy } from 'react'
import type { ExperimentVisualizationPanelProps } from '@/types/presentation'
import type { SimulatorState } from '@/types/visualizationDomain'
import { SimulationWrapper } from '@/components/visualization'
import { getVisualizationActions } from '@/domains/visualization/runtime/visualizationActions'
import { adaptVisualizationState } from '@/domains/visualization/runtime/visualizationAdapters'
import {
  resolveVisualizationModule,
} from '@/domains/visualization/runtime/visualizationRegistry'

const ConsensusCanvas = lazy(() => import('@/components/visualization/ConsensusCanvas'))
const AttackCanvas = lazy(() => import('@/components/visualization/AttackCanvas'))
const BlockchainCanvas = lazy(() => import('@/components/visualization/BlockchainCanvas'))
const NetworkCanvas = lazy(() => import('@/components/visualization/NetworkCanvas'))
const CryptoCanvas = lazy(() => import('@/components/visualization/CryptoCanvas'))
const CrossChainCanvas = lazy(() => import('@/components/visualization/CrossChainCanvas'))
const EVMCanvas = lazy(() => import('@/components/visualization/EVMCanvas'))
const DeFiCanvas = lazy(() => import('@/components/visualization/DeFiCanvas'))

function VisualizationFallback() {
  return (
    <div className="flex h-full items-center justify-center rounded-xl border border-slate-700 bg-slate-950/70 p-6 text-sm text-slate-300">
      正在加载当前可视化舞台...
    </div>
  )
}

export default function ExperimentVisualizationPanel({
  accessUrl,
  wsUrl,
  moduleKey,
}: ExperimentVisualizationPanelProps) {
  const runtime = resolveVisualizationModule(moduleKey)
  const actionDefinitions = getVisualizationActions(runtime)

  const renderState = (rawState: SimulatorState) => {
    const state = adaptVisualizationState(rawState, runtime)

    let content: JSX.Element | null = null

    switch (runtime.renderer) {
      case 'consensus':
        content = <ConsensusCanvas state={state} algorithm={runtime.algorithm} />
        break
      case 'attack':
        content = <AttackCanvas state={state} moduleKey={runtime.moduleKey} />
        break
      case 'blockchain':
        content = <BlockchainCanvas state={state} mode={runtime.mode} />
        break
      case 'network':
        content = <NetworkCanvas state={state} scenario={runtime.scenario} />
        break
      case 'crypto':
        content = <CryptoCanvas state={state} moduleKey={runtime.moduleKey} />
        break
      case 'crosschain':
        content = <CrossChainCanvas state={state} moduleKey={runtime.moduleKey} />
        break
      case 'evm':
        content = <EVMCanvas state={state} moduleKey={runtime.moduleKey} />
        break
      case 'defi':
        content = <DeFiCanvas state={state} protocol={runtime.protocol} />
        break
      default:
        content = null
    }

    if (!content) {
      return null
    }

    return <Suspense fallback={<VisualizationFallback />}>{content}</Suspense>
  }

  return (
    <div className="h-full min-h-0">
      <SimulationWrapper
        module={runtime.simulatorId}
        accessUrl={accessUrl}
        wsUrl={wsUrl}
        nodeCount={4}
        actionDefinitions={actionDefinitions}
        renderState={renderState}
        key={`${moduleKey || 'visualization-panel'}:${accessUrl || ''}`}
      />
    </div>
  )
}
