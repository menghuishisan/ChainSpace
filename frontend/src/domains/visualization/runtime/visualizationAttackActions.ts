import type {
  VisualizationActionDefinition,
  VisualizationActionField,
  VisualizationActionKind,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'

function kind(value: VisualizationActionKind): VisualizationActionKind {
  return value
}

function field(definition: VisualizationActionField): VisualizationActionField {
  return definition
}

function action(definition: VisualizationActionDefinition): VisualizationActionDefinition {
  return definition
}

function withReset(actions: VisualizationActionDefinition[], successMessage = '已重置当前攻击场景。') {
  return [
    ...actions,
    action({
      key: 'attack-reset',
      label: '重置场景',
      description: '恢复到攻击开始前的初始状态，便于重新观察完整过程。',
      kind: kind('module_action'),
      action: 'reset_attack',
      successMessage,
    }),
  ]
}

/**
 * 攻击类可视化动作定义。
 * 统一维护各攻击模块的实验入口，避免动作分散在多个旧文件里。
 */
export function getAttackActions(
  runtime: VisualizationRuntimeSpec,
  allowAttack: boolean,
): VisualizationActionDefinition[] {
  if (!allowAttack) {
    return []
  }

  switch (runtime.moduleKey) {
    case 'attacks/reentrancy':
      return withReset([
        action({
          key: 'reentrancy-start',
          label: '发起攻击',
          description: '执行一次存在漏洞的提取流程，观察调用栈、重入深度和余额变化。',
          kind: kind('module_action'),
          action: 'start_attack',
          successMessage: '已执行重入攻击演示。',
        }),
        action({
          key: 'reentrancy-secure',
          label: '切换修复流程',
          description: '切换到安全提取流程，对比漏洞路径与 CEI 修复方案的差异。',
          kind: kind('module_action'),
          action: 'show_secure_flow',
          successMessage: '已切换到安全提取流程。',
        }),
      ])

    case 'attacks/flashloan':
      return withReset([
        action({
          key: 'flashloan-price',
          label: '价格操纵',
          description: '发起一次典型的闪电贷价格操纵攻击。',
          kind: kind('module_action'),
          action: 'simulate_price_manipulation',
          successMessage: '已执行闪电贷价格操纵演示。',
        }),
        action({
          key: 'flashloan-governance',
          label: '治理攻击',
          description: '借入大量治理代币后发起恶意治理提案。',
          kind: kind('module_action'),
          action: 'simulate_governance_attack',
          successMessage: '已执行闪电贷治理攻击演示。',
        }),
        action({
          key: 'flashloan-liquidation',
          label: '清算攻击',
          description: '模拟通过闪电贷触发并利用清算窗口。',
          kind: kind('module_action'),
          action: 'simulate_liquidation_attack',
          successMessage: '已执行闪电贷清算攻击演示。',
        }),
      ])

    case 'attacks/access_control':
      return withReset([
        action({
          key: 'access-missing',
          label: '缺失权限检查',
          description: '演示敏感函数缺少访问控制带来的越权行为。',
          kind: kind('module_action'),
          action: 'simulate_missing_access_control',
          fields: [field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' })],
          successMessage: '已执行缺失权限检查攻击演示。',
        }),
        action({
          key: 'access-init',
          label: '初始化攻击',
          description: '演示初始化函数被抢先调用或重复调用的风险。',
          kind: kind('module_action'),
          action: 'simulate_initialize_attack',
          fields: [field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' })],
          successMessage: '已执行初始化攻击演示。',
        }),
        action({
          key: 'access-tx-origin',
          label: 'tx.origin 钓鱼',
          description: '演示通过中间合约诱导受害者调用造成的权限绕过。',
          kind: kind('module_action'),
          action: 'simulate_tx_origin_phishing',
          fields: [
            field({ key: 'victim', label: '受害者地址', type: 'text', defaultValue: '0xVictim' }),
            field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' }),
          ],
          successMessage: '已执行 tx.origin 钓鱼攻击演示。',
        }),
        action({
          key: 'access-role',
          label: '角色提升',
          description: '演示通过角色管理漏洞获得更高权限的过程。',
          kind: kind('module_action'),
          action: 'simulate_role_escalation',
          fields: [field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' })],
          successMessage: '已执行角色提升攻击演示。',
        }),
      ])

    case 'attacks/delegatecall_attack':
      return withReset([
        action({
          key: 'delegatecall-storage',
          label: '存储碰撞',
          description: '观察 delegatecall 如何把实现合约的写操作错误落到代理存储上。',
          kind: kind('module_action'),
          action: 'simulate_storage_collision',
          successMessage: '已执行存储碰撞攻击演示。',
        }),
        action({
          key: 'delegatecall-parity',
          label: 'Parity 攻击',
          description: '复现未初始化库合约被接管后触发自毁的经典案例。',
          kind: kind('module_action'),
          action: 'simulate_parity_hack',
          successMessage: '已执行 Parity 钱包攻击演示。',
        }),
      ])

    case 'attacks/integer_overflow':
      return withReset([
        action({
          key: 'overflow-add',
          label: '整数上溢',
          description: '演示接近最大值时加法回绕带来的错误结果。',
          kind: kind('module_action'),
          action: 'simulate_overflow',
          fields: [field({ key: 'bit_size', label: '位宽', type: 'number', defaultValue: 256, min: 8 })],
          successMessage: '已执行整数上溢演示。',
        }),
        action({
          key: 'overflow-sub',
          label: '整数下溢',
          description: '演示从零减一时回绕到最大值的危险行为。',
          kind: kind('module_action'),
          action: 'simulate_underflow',
          fields: [field({ key: 'bit_size', label: '位宽', type: 'number', defaultValue: 256, min: 8 })],
          successMessage: '已执行整数下溢演示。',
        }),
        action({
          key: 'overflow-mul',
          label: '乘法溢出',
          description: '观察大数乘法在旧版 Solidity 中如何静默回绕。',
          kind: kind('module_action'),
          action: 'simulate_mul_overflow',
          fields: [field({ key: 'bit_size', label: '位宽', type: 'number', defaultValue: 256, min: 8 })],
          successMessage: '已执行乘法溢出演示。',
        }),
      ])

    case 'attacks/frontrun':
      return withReset([
        action({
          key: 'frontrun-displacement',
          label: '替换攻击',
          description: '复制目标交易并支付更高 gas 抢先执行。',
          kind: kind('module_action'),
          action: 'simulate_displacement_attack',
          fields: [field({ key: 'victim_tx_hash', label: '目标交易哈希', type: 'text', defaultValue: '0xVictimTx' })],
          successMessage: '已执行替换型抢跑演示。',
        }),
        action({
          key: 'frontrun-insertion',
          label: '插入攻击',
          description: '在目标交易前后插入交易，观察三明治式获利过程。',
          kind: kind('module_action'),
          action: 'simulate_insertion_attack',
          fields: [field({ key: 'victim_swap_amount', label: '目标交易数量', type: 'text', defaultValue: '100 ETH' })],
          successMessage: '已执行插入型抢跑演示。',
        }),
        action({
          key: 'frontrun-suppression',
          label: '压制攻击',
          description: '通过高 gas 垃圾交易阻止目标交易及时上链。',
          kind: kind('module_action'),
          action: 'simulate_suppression_attack',
          fields: [
            field({ key: 'target_tx_hash', label: '目标交易哈希', type: 'text', defaultValue: '0xTargetTx' }),
            field({ key: 'reason', label: '攻击目的', type: 'text', defaultValue: '阻止清算' }),
          ],
          successMessage: '已执行压制型抢跑演示。',
        }),
      ])

    case 'attacks/governance':
      return withReset([
        action({
          key: 'governance-flashloan',
          label: '闪电贷治理',
          description: '借入大量治理代币后快速通过恶意提案。',
          kind: kind('module_action'),
          action: 'simulate_flashloan_governance',
          fields: [field({ key: 'flashloan_amount', label: '借入数量', type: 'number', defaultValue: 1000000, min: 1000 })],
          successMessage: '已执行闪电贷治理攻击演示。',
        }),
        action({
          key: 'governance-low-quorum',
          label: '低法定人数',
          description: '在低参与率窗口推动恶意提案通过。',
          kind: kind('module_action'),
          action: 'simulate_low_quorum_attack',
          successMessage: '已执行低法定人数攻击演示。',
        }),
        action({
          key: 'governance-bribery',
          label: '贿赂投票',
          description: '通过额外利益诱导持币者将票投给攻击者。',
          kind: kind('module_action'),
          action: 'simulate_bribery_attack',
          fields: [field({ key: 'bribe_amount', label: '贿赂预算', type: 'number', defaultValue: 5000, min: 1 })],
          successMessage: '已执行治理贿赂攻击演示。',
        }),
      ])

    case 'attacks/liquidation':
      return withReset([
        action({
          key: 'liquidation-price',
          label: '价格操纵清算',
          description: '操纵预言机价格后触发受害者仓位被清算。',
          kind: kind('module_action'),
          action: 'simulate_price_manipulation_liquidation',
          fields: [field({ key: 'flashloan_eth', label: '闪电贷 ETH', type: 'number', defaultValue: 5000, min: 1 })],
          successMessage: '已执行价格操纵清算演示。',
        }),
        action({
          key: 'liquidation-frontrun',
          label: '抢跑清算',
          description: '观察攻击者如何在 mempool 中抢先拿走清算奖励。',
          kind: kind('module_action'),
          action: 'simulate_frontrun_liquidation',
          successMessage: '已执行抢跑清算演示。',
        }),
        action({
          key: 'liquidation-self',
          label: '自我清算套利',
          description: '演示利用协议设计缺陷进行自我清算获利。',
          kind: kind('module_action'),
          action: 'simulate_self_liquidation',
          successMessage: '已执行自我清算套利演示。',
        }),
      ])

    case 'attacks/oracle_manipulation':
      return withReset([
        action({
          key: 'oracle-spot',
          label: '现货价格操纵',
          description: '操纵现货池价格并观察预言机引用失真。',
          kind: kind('module_action'),
          action: 'simulate_spot_price_manipulation',
          fields: [field({ key: 'flashloan_amount', label: '闪电贷数量', type: 'number', defaultValue: 1000000, min: 1 })],
          successMessage: '已执行现货价格操纵演示。',
        }),
        action({
          key: 'oracle-twap',
          label: 'TWAP 操纵',
          description: '持续多个区块影响 TWAP 结果，观察协议是否被误导。',
          kind: kind('module_action'),
          action: 'simulate_twap_manipulation',
          fields: [field({ key: 'manipulation_blocks', label: '操纵区块数', type: 'number', defaultValue: 6, min: 1 })],
          successMessage: '已执行 TWAP 操纵演示。',
        }),
        action({
          key: 'oracle-multihop',
          label: '多跳操纵',
          description: '通过多跳流动性池逐层传递错误价格。',
          kind: kind('module_action'),
          action: 'simulate_multi_hop_manipulation',
          successMessage: '已执行多跳预言机操纵演示。',
        }),
      ])

    case 'attacks/attack_51':
      return withReset([
        action({
          key: 'attack51-double-spend',
          label: '双花攻击',
          description: '模拟攻击者利用多数算力完成双花。',
          kind: kind('module_action'),
          action: 'simulate_double_spend_attack',
          fields: [
            field({ key: 'amount', label: '双花金额', type: 'number', defaultValue: 100, min: 1 }),
            field({ key: 'target_confirmations', label: '确认数', type: 'number', defaultValue: 6, min: 1 }),
          ],
          successMessage: '已执行 51% 双花攻击演示。',
        }),
      ])

    case 'attacks/selfish_mining':
      return withReset([
        action({
          key: 'selfish-mining-rounds',
          label: '模拟多轮挖矿',
          description: '观察攻击者隐藏区块后如何提高超额收益。',
          kind: kind('module_action'),
          action: 'simulate_rounds',
          fields: [field({ key: 'rounds', label: '轮次数', type: 'number', defaultValue: 12, min: 1 })],
          successMessage: '已执行自私挖矿多轮模拟。',
        }),
      ])

    case 'attacks/nothing_at_stake':
      return withReset([
        action({
          key: 'nothing-double-vote',
          label: '双签攻击',
          description: '让同一验证者同时在两条分叉上签名，观察是否会被惩罚。',
          kind: kind('module_action'),
          action: 'simulate_double_voting',
          fields: [field({ key: 'validator_id', label: '验证者 ID', type: 'text', defaultValue: 'V1' })],
          successMessage: '已执行双签攻击演示。',
        }),
        action({
          key: 'nothing-rational',
          label: '理性行为分析',
          description: '对比有无罚没机制时验证者的最优策略。',
          kind: kind('module_action'),
          action: 'simulate_rational_behavior',
          successMessage: '已执行无利害关系行为分析。',
        }),
      ])

    case 'attacks/long_range':
      return withReset([
        action({
          key: 'long-range-posterior',
          label: '后验长程攻击',
          description: '从较早的 epoch 开始重写历史链，观察新节点是否会被误导。',
          kind: kind('module_action'),
          action: 'simulate_posterior_attack',
          fields: [
            field({ key: 'start_epoch', label: '起始 Epoch', type: 'number', defaultValue: 200, min: 0 }),
            field({ key: 'keys_acquired', label: '获取私钥数', type: 'number', defaultValue: 3, min: 1 }),
          ],
          successMessage: '已执行后验长程攻击演示。',
        }),
        action({
          key: 'long-range-sync',
          label: '新节点同步',
          description: '比较有无可信检查点时，新节点能否识别伪造历史。',
          kind: kind('module_action'),
          action: 'simulate_new_node_sync',
          fields: [field({
            key: 'has_checkpoint',
            label: '是否有检查点',
            type: 'select',
            defaultValue: 'true',
            options: [
              { label: '有检查点', value: 'true' },
              { label: '无检查点', value: 'false' },
            ],
          })],
          successMessage: '已执行新节点同步演示。',
        }),
      ])

    case 'attacks/bridge_attack':
      return withReset([
        action({
          key: 'bridge-multisig',
          label: '多签妥协',
          description: '观察验证者私钥泄露后如何直接提走桥内资产。',
          kind: kind('module_action'),
          action: 'simulate_multisig_compromise',
          successMessage: '已执行跨链桥多签妥协演示。',
        }),
        action({
          key: 'bridge-signature',
          label: '签名绕过',
          description: '观察验证逻辑存在缺陷时如何伪造跨链消息。',
          kind: kind('module_action'),
          action: 'simulate_signature_bypass',
          successMessage: '已执行跨链桥签名绕过演示。',
        }),
        action({
          key: 'bridge-init',
          label: '初始化漏洞',
          description: '观察桥合约初始化错误如何导致校验机制完全失效。',
          kind: kind('module_action'),
          action: 'simulate_initialization_attack',
          successMessage: '已执行跨链桥初始化漏洞演示。',
        }),
      ])

    case 'attacks/bribery':
      return withReset([
        action({
          key: 'bribery-p-epsilon',
          label: 'P+epsilon 贿赂',
          description: '观察条件性贿赂如何改变矿工或验证者的理性选择。',
          kind: kind('module_action'),
          action: 'simulate_p_epsilon_attack',
          fields: [
            field({ key: 'epsilon', label: '额外激励比例', type: 'number', defaultValue: 0.2, min: 0 }),
            field({ key: 'miner_count', label: '目标节点数', type: 'number', defaultValue: 10, min: 1 }),
          ],
          successMessage: '已执行 P+epsilon 贿赂攻击演示。',
        }),
        action({
          key: 'bribery-governance',
          label: '治理贿赂',
          description: '模拟通过买票影响治理结果的过程。',
          kind: kind('module_action'),
          action: 'simulate_governance_bribery',
          fields: [
            field({ key: 'bribe_per_vote', label: '每票成本', type: 'number', defaultValue: 50, min: 0 }),
            field({ key: 'votes_needed', label: '目标票数', type: 'number', defaultValue: 100000, min: 1 }),
          ],
          successMessage: '已执行治理贿赂演示。',
        }),
        action({
          key: 'bribery-mev',
          label: 'MEV 贿赂',
          description: '观察搜索者如何向构建者支付额外收益以获得优先执行权。',
          kind: kind('module_action'),
          action: 'simulate_mev_bribery',
          successMessage: '已执行 MEV 贿赂演示。',
        }),
      ])

    case 'attacks/dos':
      return withReset([
        action({
          key: 'dos-loop',
          label: '无界循环',
          description: '通过不断增大遍历规模触发 gas 耗尽。',
          kind: kind('module_action'),
          action: 'simulate_unbounded_loop',
          fields: [field({ key: 'array_size', label: '数组规模', type: 'number', defaultValue: 8000, min: 1 })],
          successMessage: '已执行无界循环 DoS 演示。',
        }),
        action({
          key: 'dos-external-call',
          label: '外部调用阻塞',
          description: '模拟恶意合约拒绝收款导致主流程整体回滚。',
          kind: kind('module_action'),
          action: 'simulate_external_call_dos',
          successMessage: '已执行外部调用阻塞 DoS 演示。',
        }),
        action({
          key: 'dos-block-stuffing',
          label: '区块填充',
          description: '用大量高费用交易阻塞目标交易进入区块。',
          kind: kind('module_action'),
          action: 'simulate_block_stuffing',
          fields: [field({ key: 'target_blocks', label: '阻塞区块数', type: 'number', defaultValue: 3, min: 1 })],
          successMessage: '已执行区块填充 DoS 演示。',
        }),
      ])

    case 'attacks/fake_deposit':
      return withReset([
        action({
          key: 'fake-deposit-token',
          label: '假代币充值',
          description: '演示同名假代币如何绕过只看事件的充值检测。',
          kind: kind('module_action'),
          action: 'simulate_fake_token_attack',
          successMessage: '已执行假代币充值攻击演示。',
        }),
        action({
          key: 'fake-deposit-zero-conf',
          label: '零确认双花',
          description: '观察未等待确认就入账时的双花风险。',
          kind: kind('module_action'),
          action: 'simulate_zero_conf_attack',
          fields: [field({ key: 'confirmations', label: '确认数', type: 'number', defaultValue: 0, min: 0 })],
          successMessage: '已执行零确认双花演示。',
        }),
        action({
          key: 'fake-deposit-failed-tx',
          label: '失败交易入账',
          description: '演示只检查交易存在而不检查状态时的错误入账。',
          kind: kind('module_action'),
          action: 'simulate_failed_tx_attack',
          successMessage: '已执行失败交易入账演示。',
        }),
      ])

    case 'attacks/infinite_mint':
      return withReset([
        action({
          key: 'infinite-mint-access',
          label: '无限增发',
          description: '模拟利用铸造漏洞无限扩大供应量并稀释持有者资产。',
          kind: kind('module_action'),
          action: 'simulate_infinite_mint',
          fields: [
            field({
              key: 'attack_type',
              label: '漏洞类型',
              type: 'select',
              defaultValue: 'no_access_control',
              options: [
                { label: '无访问控制', value: 'no_access_control' },
                { label: '重入铸造', value: 'reentrancy' },
                { label: '整数溢出', value: 'overflow' },
                { label: '逻辑漏洞', value: 'logic_flaw' },
                { label: '跨链桥漏洞', value: 'bridge_exploit' },
              ],
            }),
            field({ key: 'mint_multiplier', label: '增发倍数', type: 'number', defaultValue: 10, min: 1 }),
          ],
          successMessage: '已执行无限增发攻击演示。',
        }),
        action({
          key: 'infinite-mint-dump',
          label: '增发后抛售',
          description: '观察增发完成后对价格、供应量和市值的冲击。',
          kind: kind('module_action'),
          action: 'simulate_dump_attack',
          successMessage: '已执行增发后抛售演示。',
        }),
      ])

    case 'attacks/selfdestruct':
      return withReset([
        action({
          key: 'selfdestruct-force-send',
          label: '强制转账',
          description: '演示自毁合约如何强制向目标地址注入 ETH。',
          kind: kind('module_action'),
          action: 'simulate_force_send_eth',
          fields: [field({ key: 'amount_eth', label: '注入金额', type: 'number', defaultValue: 1, min: 0 })],
          successMessage: '已执行强制转账演示。',
        }),
        action({
          key: 'selfdestruct-game',
          label: '破坏游戏流程',
          description: '观察意外收到 ETH 后，依赖余额条件的合约如何失效。',
          kind: kind('module_action'),
          action: 'simulate_game_breaking',
          successMessage: '已执行游戏逻辑破坏演示。',
        }),
        action({
          key: 'selfdestruct-proxy',
          label: '代理合约销毁',
          description: '演示错误设计下代理或实现合约被销毁后的影响。',
          kind: kind('module_action'),
          action: 'simulate_proxy_attack',
          successMessage: '已执行代理合约销毁演示。',
        }),
        action({
          key: 'selfdestruct-create2',
          label: 'CREATE2 重新部署',
          description: '观察相同地址在销毁后被重新部署并替换逻辑的风险。',
          kind: kind('module_action'),
          action: 'simulate_create2_redeploy',
          successMessage: '已执行 CREATE2 重新部署演示。',
        }),
      ])

    case 'attacks/signature_replay':
      return withReset([
        action({
          key: 'signature-replay-same-chain',
          label: '同链重放',
          description: '重放同一条链上的历史签名，观察重复执行问题。',
          kind: kind('module_action'),
          action: 'simulate_same_chain_replay',
          fields: [field({ key: 'signer', label: '签名者地址', type: 'text', defaultValue: '0xSigner' })],
          successMessage: '已执行同链签名重放演示。',
        }),
        action({
          key: 'signature-replay-cross-chain',
          label: '跨链重放',
          description: '观察相同签名在不同链上被重复利用的风险。',
          kind: kind('module_action'),
          action: 'simulate_cross_chain_replay',
          successMessage: '已执行跨链签名重放演示。',
        }),
      ])

    case 'attacks/sandwich':
      return withReset([
        action({
          key: 'sandwich-start',
          label: '三明治攻击',
          description: '在目标交易前后插入订单，观察受害者滑点和攻击者利润。',
          kind: kind('module_action'),
          action: 'simulate_sandwich_attack',
          fields: [
            field({ key: 'victim_amount', label: '目标交易量', type: 'text', defaultValue: '10000000000000000000' }),
            field({ key: 'slippage', label: '滑点比例', type: 'number', defaultValue: 0.02, min: 0 }),
          ],
          successMessage: '已执行三明治攻击演示。',
        }),
      ])

    case 'attacks/weak_randomness':
      return withReset([
        action({
          key: 'weak-randomness-predict',
          label: '预测随机数',
          description: '演示利用可预测的链上随机源在获利时才参与攻击。',
          kind: kind('module_action'),
          action: 'simulate_prediction',
          successMessage: '已执行弱随机数预测攻击演示。',
        }),
      ])

    case 'attacks/timestamp_manipulation':
      return withReset([
        action({
          key: 'timestamp-bypass',
          label: '时间锁绕过',
          description: '演示出块者微调时间戳后如何绕过时间条件。',
          kind: kind('module_action'),
          action: 'simulate_timelock_bypass',
          fields: [field({ key: 'unlock_time', label: '解锁时间', type: 'number', defaultValue: 1893456000, min: 1 })],
          successMessage: '已执行时间戳操纵绕过演示。',
        }),
      ])

    case 'attacks/tx_origin':
      return withReset([
        action({
          key: 'tx-origin-direct',
          label: '直接钓鱼',
          description: '诱导受害者调用恶意合约，观察 tx.origin 校验被绕过。',
          kind: kind('module_action'),
          action: 'simulate_direct_phishing',
          fields: [
            field({ key: 'victim', label: '受害者地址', type: 'text', defaultValue: '0xVictim' }),
            field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' }),
          ],
          successMessage: '已执行直接钓鱼演示。',
        }),
        action({
          key: 'tx-origin-nft',
          label: 'NFT 钓鱼',
          description: '通过恶意 NFT 回调触发受害者钱包中的危险逻辑。',
          kind: kind('module_action'),
          action: 'simulate_nft_phishing',
          fields: [
            field({ key: 'victim', label: '受害者地址', type: 'text', defaultValue: '0xVictim' }),
            field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' }),
          ],
          successMessage: '已执行 NFT 钓鱼演示。',
        }),
        action({
          key: 'tx-origin-approval',
          label: '授权钓鱼',
          description: '观察通过恶意前端或合约诱导无限授权后的风险。',
          kind: kind('module_action'),
          action: 'simulate_approval_phishing',
          fields: [
            field({ key: 'victim', label: '受害者地址', type: 'text', defaultValue: '0xVictim' }),
            field({ key: 'attacker', label: '攻击者地址', type: 'text', defaultValue: '0xAttacker' }),
          ],
          successMessage: '已执行授权钓鱼演示。',
        }),
      ])

    default:
      return withReset([
        action({
          key: 'attack-show-defenses',
          label: '查看防御',
          description: '展示当前攻击场景对应的防御措施和修复思路。',
          kind: kind('module_action'),
          action: 'show_defenses',
          successMessage: '已切换到防御视图。',
        }),
      ])
  }
}
