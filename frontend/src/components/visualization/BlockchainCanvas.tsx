import { Tag } from 'antd'
import {
  BlockOutlined,
  FieldBinaryOutlined,
  NodeIndexOutlined,
  SafetyCertificateOutlined,
  TransactionOutlined,
} from '@ant-design/icons'
import type {
  BlockchainCanvasProps,
  BlockchainStats,
  BlockField,
  BlockTransaction,
  ChainBlock,
  CurrentBlockSummary,
  MerkleProofView,
  MerkleTreeLevelView,
  VisualizationDisturbanceItem,
} from '@/types/visualizationDomain'

function buildDisturbances(globalData: Record<string, unknown>): VisualizationDisturbanceItem[] {
  const activeFaults = Array.isArray(globalData.active_faults)
    ? (globalData.active_faults as Record<string, unknown>[])
    : []
  const activeAttacks = Array.isArray(globalData.active_attacks)
    ? (globalData.active_attacks as Record<string, unknown>[])
    : []

  return [
    ...activeFaults.map((item, index) => ({
      id: String(item.id ?? `fault-${index}`),
      type: String(item.type ?? 'fault'),
      target: String(item.target ?? '--'),
      label: '故障联动',
      summary: `当前存在故障注入：${String(item.type ?? '--')}，作用目标为 ${String(item.target ?? '--')}。`,
    })),
    ...activeAttacks.map((item, index) => ({
      id: String(item.id ?? `attack-${index}`),
      type: String(item.type ?? 'attack'),
      target: String(item.target ?? '--'),
      label: '攻击联动',
      summary: `当前存在联动攻击：${String(item.type ?? '--')}，作用目标为 ${String(item.target ?? '--')}。`,
    })),
  ]
}

function fieldTone(index: number): string {
  const tones = [
    'border-sky-200 bg-sky-50',
    'border-emerald-200 bg-emerald-50',
    'border-amber-200 bg-amber-50',
    'border-fuchsia-200 bg-fuchsia-50',
  ]
  return tones[index % tones.length]
}

/**
 * 区块结构模式渲染
 */
function renderChainMode(
  blocks: ChainBlock[],
  currentBlock: CurrentBlockSummary | undefined,
  fields: BlockField[],
  transactions: BlockTransaction[],
) {
  return (
    <>
      {/* 核心指标卡片 */}
      <section className="grid gap-3 lg:grid-cols-3">
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs text-slate-500">当前区块哈希</div>
          <div className="mt-1 break-all font-mono text-sky-700 text-sm">{currentBlock?.hash || '--'}</div>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs text-slate-500">前块哈希</div>
          <div className="mt-1 break-all font-mono text-emerald-700 text-sm">{currentBlock?.prevHash || '--'}</div>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs text-slate-500">Merkle Root</div>
          <div className="mt-1 break-all font-mono text-amber-700 text-sm">{currentBlock?.merkleRoot || '--'}</div>
        </div>
      </section>

      {/* 链式结构 */}
      <section className="rounded-lg border border-slate-200 bg-white p-3">
        <div className="mb-3 text-xs font-medium text-slate-900">链式结构与区块关系</div>
        <div className="grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
          {/* 区块列表 */}
          <div className="space-y-2">
            {blocks.length > 0 ? (
              <div className="flex gap-2 overflow-x-auto pb-2">
                {blocks.map((block) => (
                  <div
                    key={block.id}
                    className={`min-w-[120px] rounded-lg border p-3 text-xs shrink-0 ${
                      block.status === 'current'
                        ? 'border-sky-300 bg-sky-50'
                        : 'border-slate-300 bg-slate-50'
                    }`}
                  >
                    <div className="font-semibold text-slate-900"># {block.number}</div>
                    <div className="mt-1 text-slate-500 truncate" title={block.hash}>哈希: {block.hash}</div>
                    <div className="mt-1 text-slate-500">前块: {block.prevHash}</div>
                    <div className="mt-2 text-sm font-medium text-sky-700">{block.txCount} 笔交易</div>
                    <div className="mt-1 text-slate-500 leading-relaxed">{block.explanation}</div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="rounded-lg border border-dashed border-slate-300 p-6 text-sm text-slate-500 text-center">
                等待区块数据...
              </div>
            )}
          </div>

          {/* 区块详情 */}
          <div className="space-y-3">
            <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
              <div className="text-xs font-medium text-slate-900 mb-2">字段变化会带来什么</div>
              <div className="space-y-2 text-xs text-slate-600">
                <div>1. 修改任意字段后，当前区块哈希都会重新计算</div>
                <div>2. 如果前块哈希不匹配，区块就无法正确接到原来的链上</div>
                <div>3. 新交易进入区块后，Merkle Root 和交易数量也会一起变化</div>
              </div>
            </div>
            {currentBlock && (
              <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">当前区块摘要</div>
                <div className="grid grid-cols-2 gap-1 text-xs">
                  <span className="text-slate-500">时间戳:</span><span>{currentBlock.timestamp}</span>
                  <span className="text-slate-500">Nonce:</span><span>{currentBlock.nonce}</span>
                  <span className="text-slate-500">难度:</span><span>{currentBlock.difficulty}</span>
                  <span className="text-slate-500">交易数:</span><span>{currentBlock.txCount}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </section>

      {/* 区块结构拆解 */}
      <section className="grid gap-3 lg:grid-cols-2">
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs font-medium text-slate-900 mb-2">区块结构拆解</div>
          <div className="grid gap-2 md:grid-cols-2">
            {fields.length > 0 ? (
              fields.map((entry, index) => (
                  <div key={entry.key} className={`rounded-lg border p-2 text-xs ${fieldTone(index)}`}>
                    <div className="text-slate-600">{entry.name}</div>
                    <div className="mt-1 break-all font-mono text-slate-800">{entry.value}</div>
                  </div>
              ))
            ) : (
              <div className="col-span-2 rounded-lg border border-dashed border-slate-300 p-4 text-sm text-slate-500 text-center">
                暂无区块字段数据
              </div>
            )}
          </div>
        </div>

        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs font-medium text-slate-900 mb-2">区块内交易</div>
          <div className="space-y-2 max-h-[200px] overflow-auto">
            {transactions.length > 0 ? (
              transactions.slice(0, 6).map((tx) => (
                <div key={tx.id} className="rounded bg-slate-50 p-2 text-xs">
                  <div className="font-mono text-sky-700 truncate">{tx.hash}</div>
                  <div className="mt-1 text-slate-500">
                    {tx.from} → {tx.to}
                  </div>
                  <div className="mt-1 flex justify-between text-slate-600">
                    <span>金额: {tx.value}</span>
                    <span className="text-emerald-700">{tx.status}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-sm text-slate-500 text-center py-4">暂无交易</div>
            )}
          </div>
        </div>
      </section>
    </>
  )
}

/**
 * Merkle树模式渲染
 */
function renderMerkleMode(
  levels: MerkleTreeLevelView[],
  proof: MerkleProofView | undefined,
  stats: BlockchainStats,
) {
  return (
    <>
      {/* 核心指标 */}
      <section className="grid gap-3 lg:grid-cols-3">
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs text-slate-500">根哈希</div>
          <div className="mt-1 break-all font-mono text-sky-700 text-sm">{proof?.root || '--'}</div>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3 text-center">
          <div className="text-xs text-slate-500">叶子数量</div>
          <div className="mt-1 text-2xl font-semibold text-emerald-700">{stats.leafCount ?? 0}</div>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3 text-center">
          <div className="text-xs text-slate-500">树高</div>
          <div className="mt-1 text-2xl font-semibold text-amber-700">{stats.treeHeight ?? 0}</div>
        </div>
      </section>

      {/* Merkle树层级 */}
      <section className="rounded-lg border border-slate-200 bg-white p-3">
        <div className="mb-3 flex items-center gap-2 text-xs font-medium text-slate-900">
          <NodeIndexOutlined />
          Merkle 树层级
        </div>
        <div className="space-y-3 max-h-[250px] overflow-auto">
          {levels.length > 0 ? (
            levels.map((level) => (
              <div key={level.level} className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                <div className="text-xs font-medium text-sky-700 mb-2">Level {level.level}</div>
                <div className="flex flex-wrap gap-2">
                  {level.nodes.map((node, index) => (
                    <div
                      key={`${level.level}-${index}`}
                      className={`min-w-[80px] rounded-lg p-2 text-xs ${
                        proof?.leafIndex === node.index || index === 0
                          ? 'border border-emerald-200 bg-emerald-50 text-emerald-800'
                          : 'border border-slate-200 bg-white text-slate-700'
                      }`}
                    >
                      <div className="text-slate-500">
                        {node.isLeaf ? `叶子 ${node.index}` : `节点 ${index + 1}`}
                      </div>
                      <div className="mt-1 break-all font-mono">{node.hash}</div>
                    </div>
                  ))}
                </div>
              </div>
            ))
          ) : (
            <div className="text-sm text-slate-500 text-center py-4">等待树层级数据...</div>
          )}
        </div>
      </section>

      {/* 证明路径 */}
      <section className="grid gap-3 lg:grid-cols-2">
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
            <SafetyCertificateOutlined />
            证明路径
          </div>
          <div className="space-y-2 max-h-[200px] overflow-auto">
            {proof?.steps && proof.steps.length > 0 ? (
              proof.steps.map((step, index) => (
                <div key={index} className="rounded bg-slate-50 p-2 text-xs">
                  <div className="flex justify-between text-slate-500 mb-1">
                    <span>步骤 {index + 1}</span>
                    <span className="text-sky-700">{step.direction}</span>
                  </div>
                  <div className="break-all font-mono text-slate-600">{step.sibling}</div>
                </div>
              ))
            ) : (
              <div className="text-sm text-slate-500 text-center py-4">暂无证明路径</div>
            )}
          </div>
        </div>

        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <div className="text-xs font-medium text-slate-900 mb-2">观察重点</div>
          <div className="space-y-2 text-xs text-slate-600">
            <div>1. 先看叶子如何汇总成父节点，理解根哈希为什么能代表整棵树</div>
            <div>2. 再看证明路径，确认验证时到底重建了哪条链路</div>
            <div>3. 如果引入联动攻击或故障，重点看哪一步先破坏了证明一致性</div>
          </div>
        </div>
      </section>
    </>
  )
}

/**
 * 区块链可视化组件
 * 展示区块结构、链式结构、Merkle树等区块链基础知识
 * 
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - 动画和状态指示清晰
 */
export default function BlockchainCanvas({ state, mode = 'chain' }: BlockchainCanvasProps) {
  const data = (state.data || {}) as {
    stats?: BlockchainStats
    blocks?: ChainBlock[]
    currentBlock?: CurrentBlockSummary
    fields?: BlockField[]
    transactions?: BlockTransaction[]
    levels?: MerkleTreeLevelView[]
    proof?: MerkleProofView
  }
  const globalData = (state.global_data || {}) as Record<string, unknown>
  const stats = data.stats || {}
  const disturbances = buildDisturbances(globalData)

  // 计算过程进度
  const progress = mode === 'merkle'
    ? Math.min(100, Math.max(22, (data.levels?.length || 0) * 18 + ((data.proof?.steps.length || 0) * 12)))
    : Math.min(100, Math.max(24, (data.blocks?.length || 0) * 18 + (data.transactions?.length || 0) * 6))

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#eef5ff_0%,#e7eef9_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" icon={<FieldBinaryOutlined />} className="m-0">
              模式: {mode === 'merkle' ? 'Merkle证明' : '区块结构'}
            </Tag>
            <Tag color="cyan" icon={<BlockOutlined />} className="m-0">
              步骤: {state.tick}
            </Tag>
            <Tag color="green" icon={<TransactionOutlined />} className="m-0">
              交易数: {stats.txCount ?? 0}
            </Tag>
            {data.blocks && data.blocks.length > 0 && (
              <Tag color="purple" className="m-0">
                区块: {data.blocks.length}
              </Tag>
            )}
          </div>
          <div className="text-xs text-slate-500">
            {mode === 'merkle' ? 'Merkle树与证明验证' : '区块结构与链式生命周期'}
          </div>
        </div>

        {/* 过程进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>可视化进度</span>
            <span>{progress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-emerald-500 transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="flex-1 overflow-auto bg-[linear-gradient(180deg,#f7fbff_0%,#f3f6fa_100%)] p-3">
        <div className="space-y-3">
            {/* 联动影响提示 */}
            {disturbances.length > 0 && (
              <section className="rounded-lg border border-amber-200 bg-amber-50 p-3">
                <div className="mb-2 text-xs font-medium text-amber-800">
                  当前存在联动影响
                </div>
                <div className="grid gap-2 md:grid-cols-2">
                  {disturbances.map((item) => (
                    <div key={item.id} className="rounded border border-amber-100 bg-white p-2 text-xs">
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-slate-900">{item.label}</span>
                        <span className="text-amber-700">{item.type}</span>
                      </div>
                      <div className="mt-1 text-slate-600">{item.summary}</div>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {/* 过程轨道 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">
                {mode === 'merkle' ? 'Merkle证明轨道' : '区块生命周期轨道'}
              </div>
              <div className="flex items-center gap-1 overflow-x-auto pb-1">
                {(mode === 'merkle'
                  ? ['准备叶子', '向上汇总', '生成证明', '验证根哈希']
                  : ['收集交易', '构建区块头', '计算哈希', '接入链上']
                ).map((item, index) => {
                  const active = progress >= (index + 1) * 25
                  return (
                    <div key={item} className="flex items-center">
                      <div
                        className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-xs whitespace-nowrap transition-all ${
                          active
                            ? 'border-sky-300 bg-sky-100 text-sky-800 shadow-[0_10px_25px_rgba(56,189,248,0.15)]'
                            : 'border-slate-200 bg-white text-slate-600'
                        }`}
                      >
                        {active && (
                          <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-sky-500" />
                        )}
                        <span>{item}</span>
                      </div>
                      {index < 3 && (
                        <div className={`mx-1 h-0.5 w-6 ${
                          active ? 'bg-sky-300' : 'bg-slate-300'
                        }`} />
                      )}
                    </div>
                  )
                })}
              </div>
            </section>

            {/* 主舞台可视化 */}
            {mode === 'merkle'
              ? renderMerkleMode(data.levels || [], data.proof, stats)
              : renderChainMode(data.blocks || [], data.currentBlock, data.fields || [], data.transactions || [])
            }
            <section className="grid gap-3 xl:grid-cols-[1.1fr_0.9fr]">
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="mb-2 text-xs font-medium text-slate-900">
                  {mode === 'merkle' ? '结果摘要' : '当前区块摘要'}
                </div>
                {mode === 'merkle' ? (
                  <div className="space-y-2">
                    <div className="rounded-lg border border-sky-100 bg-sky-50 p-3 text-xs">
                      <div className="text-slate-500">根哈希</div>
                      <div className="mt-1 break-all font-mono text-sky-700">{data.proof?.root || '--'}</div>
                    </div>
                    <div className="grid gap-2 sm:grid-cols-2">
                      <div className="rounded-lg border border-emerald-100 bg-emerald-50 p-3 text-center">
                        <div className="text-lg font-semibold text-emerald-700">{stats.leafCount ?? 0}</div>
                        <div className="text-xs text-slate-500">叶子数</div>
                      </div>
                      <div className="rounded-lg border border-amber-100 bg-amber-50 p-3 text-center">
                        <div className="text-lg font-semibold text-amber-700">{stats.treeHeight ?? 0}</div>
                        <div className="text-xs text-slate-500">树高</div>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="grid gap-2 md:grid-cols-2">
                    <div className="rounded-lg border border-sky-100 bg-sky-50 p-3 text-xs">
                      <div className="text-slate-500">区块哈希</div>
                      <div className="mt-1 break-all font-mono text-sky-700">{data.currentBlock?.hash || '--'}</div>
                    </div>
                    <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs">
                      <div className="grid grid-cols-2 gap-1">
                        <span className="text-slate-500">时间戳:</span>
                        <span>{data.currentBlock?.timestamp || '--'}</span>
                        <span className="text-slate-500">Nonce:</span>
                        <span>{data.currentBlock?.nonce || '--'}</span>
                        <span className="text-slate-500">难度:</span>
                        <span>{data.currentBlock?.difficulty || '--'}</span>
                        <span className="text-slate-500">交易数:</span>
                        <span>{data.currentBlock?.txCount || 0}</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="mb-2 text-xs font-medium text-slate-900">
                  {mode === 'merkle' ? '证明路径' : '链式摘要'}
                </div>
                <div className="space-y-2 max-h-[220px] overflow-auto">
                  {mode === 'merkle' ? (
                    data.proof?.steps && data.proof.steps.length > 0 ? (
                      data.proof.steps.map((step, index) => (
                        <div key={index} className="rounded-lg border border-slate-200 bg-slate-50 p-2 text-xs">
                          <div className="flex items-center justify-between">
                            <span className="font-medium text-slate-900">步骤 {index + 1}</span>
                            <span className="text-sky-700">{step.direction}</span>
                          </div>
                          <div className="mt-1 break-all font-mono text-slate-600">{step.sibling}</div>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs text-slate-500">
                        暂无证明路径
                      </div>
                    )
                  ) : (
                    data.blocks && data.blocks.length > 0 ? (
                      data.blocks.slice(-6).reverse().map((block) => (
                        <div
                          key={block.id}
                          className={`rounded-lg border p-2 text-xs ${
                            block.status === 'current'
                              ? 'border-sky-200 bg-sky-50'
                              : 'border-slate-200 bg-slate-50'
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <span className="font-medium text-slate-900"># {block.number}</span>
                            <span className="text-sky-700">{block.txCount} 笔</span>
                          </div>
                          <div className="mt-1 truncate text-slate-500" title={block.hash}>{block.hash}</div>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs text-slate-500">
                        暂无区块数据
                      </div>
                    )
                  )}
                </div>
              </div>
            </section>

            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-900">观察提示</div>
              <div className="grid gap-2 md:grid-cols-3">
                {(mode === 'merkle'
                  ? ['先看叶子如何汇总成父节点。', '理解根哈希为什么能代表整棵树。', '再看证明路径如何重建验证。']
                  : ['先看区块如何链接成链。', '再看区块头字段如何影响哈希。', '最后观察交易进入区块后怎样改变整体状态。']
                ).map((item) => (
                  <div key={item} className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs leading-6 text-slate-600">
                    {item}
                  </div>
                ))}
              </div>
            </section>
          </div>
        </div>
    </div>
  )
}

