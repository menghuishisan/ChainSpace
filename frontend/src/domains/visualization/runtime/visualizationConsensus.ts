import type {
  ConsensusAlgorithm,
  ConsensusMechanism,
} from '@/types/visualizationDomain'

/**
 * 把具体算法归类到对应的共识机制主舞台。
 * 这样前端可以统一实验骨架，但为不同机制提供更贴切的过程表达。
 */
export function getConsensusMechanism(algorithm?: ConsensusAlgorithm): ConsensusMechanism {
  switch (algorithm) {
    case 'raft':
      return 'leader_replication'
    case 'pow':
    case 'fork_choice':
      return 'mining'
    case 'pos':
    case 'dpos':
    case 'vrf':
      return 'committee'
    case 'dag':
      return 'dag'
    case 'pbft':
    case 'hotstuff':
    case 'tendermint':
    default:
      return 'bft'
  }
}
