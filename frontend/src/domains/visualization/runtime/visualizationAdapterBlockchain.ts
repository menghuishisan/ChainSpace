import type {
  BlockField,
  BlockchainStats,
  BlockTransaction,
  ChainBlock,
  CurrentBlockSummary,
  MerkleProofView,
  MerkleTreeLevelView,
  SimulatorState,
  VisualizationRecord,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import type { RawBlockchainField } from '@/types/visualizationDomain'
import {
  asArray,
  asNumber,
  asRecord,
  asString,
  getGlobalData,
  shortHash,
} from './visualizationAdapterCommon'

function buildMerkleData(state: SimulatorState): VisualizationRecord {
  const globalData = getGlobalData(state)
  const treeStructure = asArray<VisualizationRecord[]>(globalData.tree_structure)
  const latestProof = asRecord(globalData.latest_proof)
  const verification = asRecord(globalData.last_verification)

  const merkleLevels = treeStructure.map<MerkleTreeLevelView>((level, levelIndex) => ({
    level: levelIndex,
    nodes: asArray<VisualizationRecord>(level).map((node, index) => ({
      hash: shortHash(node.hash, `node-${levelIndex}-${index}`),
      index: asNumber(node.index, index),
      isLeaf: Boolean(node.is_leaf),
    })),
  }))

  const merkleProof = Object.keys(latestProof).length > 0
    ? {
      leafIndex: asNumber(latestProof.leaf_index),
      leafHash: shortHash(latestProof.leaf_hash, '--'),
      root: shortHash(latestProof.root, '--'),
      verified: Boolean(verification.valid),
      steps: asArray<string>(latestProof.siblings).map((sibling, index) => ({
        direction: asArray<string>(latestProof.directions)[index] || 'right',
        sibling: shortHash(sibling, '--'),
      })),
    } satisfies MerkleProofView
    : undefined

  return {
    merkleLevels,
    merkleProof,
    stats: {
      leafCount: asNumber(globalData.leaf_count),
      treeHeight: asNumber(globalData.tree_height),
      proofDepth: merkleProof?.steps.length ?? 0,
    } satisfies BlockchainStats,
  }
}

export function buildBlockchainVisualizationData(
  state: SimulatorState,
  runtime: VisualizationRuntimeSpec,
): VisualizationRecord {
  if (runtime.mode === 'merkle') {
    return buildMerkleData(state)
  }

  const globalData = getGlobalData(state)
  const currentBlock = asRecord(globalData.current_block)
  const chainLength = asNumber(globalData.chain_length, 1)
  const fields = asArray<RawBlockchainField>(globalData.fields)
  const fieldList = fields.length > 0
    ? fields.map<BlockField>((field, index) => ({
      key: `field-${index}`,
      name: asString(field.name, `字段 ${index + 1}`),
      value: asString(field.value, '--'),
    }))
    : [
      { key: 'hash', name: '区块哈希', value: shortHash(currentBlock.hash, '--') },
      { key: 'prev', name: '前块哈希', value: shortHash(currentBlock.prev_block_hash, '--') },
      { key: 'timestamp', name: '时间戳', value: asString(currentBlock.timestamp, '--') },
      { key: 'nonce', name: 'Nonce', value: asString(currentBlock.nonce, '--') },
    ]

  const transactions = asArray<string | VisualizationRecord>(currentBlock.transactions).map<BlockTransaction>((item, index) => {
    const tx = typeof item === 'string' ? { hash: item } : asRecord(item)
    return {
      id: asString(tx.hash, `tx-${index}`),
      hash: shortHash(tx.hash, `tx-${index}`),
      from: shortHash(tx.from, `账户 ${index + 1}`),
      to: shortHash(tx.to, `账户 ${index + 2}`),
      value: asString(tx.value, '0'),
      status: asString(tx.status, 'confirmed'),
    }
  })

  const blocks = Array.from({ length: Math.min(chainLength, 5) }, (_, index) => {
    const blockNumber = chainLength - Math.min(chainLength, 5) + index
    const isLatest = index === Math.min(chainLength, 5) - 1
    return {
      id: `block-${blockNumber}`,
      number: blockNumber,
      hash: isLatest ? shortHash(currentBlock.hash, `block-${blockNumber}`) : `block-${blockNumber}`,
      prevHash: isLatest ? shortHash(currentBlock.prev_block_hash, '--') : `block-${blockNumber - 1}`,
      txCount: isLatest ? transactions.length : Math.max(1, index + 1),
      status: (isLatest ? 'current' : 'confirmed') as ChainBlock['status'],
      explanation: isLatest ? '当前正在观察最新区块。' : '这是已经写入主链的历史区块。',
    }
  })

  return {
    blocks,
    currentBlock: {
      hash: shortHash(currentBlock.hash, '--'),
      prevHash: shortHash(currentBlock.prev_block_hash, '--'),
      merkleRoot: shortHash(currentBlock.merkle_root, '--'),
      timestamp: asString(currentBlock.timestamp, '--'),
      nonce: asString(currentBlock.nonce, '--'),
      difficulty: asString(currentBlock.difficulty, '--'),
      txCount: transactions.length,
    } satisfies CurrentBlockSummary,
    fields: fieldList,
    transactions,
    stats: {
      height: chainLength,
      txCount: transactions.length,
      fieldCount: fieldList.length,
      difficulty: asString(currentBlock.difficulty, '--'),
    } satisfies BlockchainStats,
  }
}
