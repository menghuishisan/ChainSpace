-- ============================================
-- ChainSpace 实验基线数据（全量重写版）
-- 说明：
-- 1. 直接清理旧实验基线数据，不兼容旧种子结构。
-- 2. 覆盖 9 种实验类型。
-- 3. 覆盖 visualizationRegistry.ts 中登记的全部可视化模块。
-- 4. 统一生成工作区、工具、拓扑、服务、节点、资源、检查点与协作配置。
-- ============================================

BEGIN;

DO $$
DECLARE
    v_teacher_id bigint;
    v_school_id bigint;
    v_course_id bigint;
    v_chapter_id bigint;
BEGIN
    SELECT id, COALESCE(school_id, 1)
    INTO v_teacher_id, v_school_id
    FROM users
    WHERE deleted_at IS NULL
    ORDER BY CASE WHEN role = 'teacher' THEN 0 ELSE 1 END, id
    LIMIT 1;

    IF v_teacher_id IS NULL THEN
        RAISE EXCEPTION '缺少可用用户，无法初始化实验基线数据';
    END IF;

    SELECT id
    INTO v_chapter_id
    FROM chapters
    WHERE deleted_at IS NULL
    ORDER BY id
    LIMIT 1;

    IF v_chapter_id IS NULL THEN
        SELECT id
        INTO v_course_id
        FROM courses
        WHERE teacher_id = v_teacher_id AND deleted_at IS NULL
        ORDER BY id
        LIMIT 1;

        IF v_course_id IS NULL THEN
            INSERT INTO courses (
                school_id,
                teacher_id,
                title,
                description,
                status,
                is_public,
                max_students,
                category
            ) VALUES (
                v_school_id,
                v_teacher_id,
                '区块链系统实验课程（实验基线）',
                '平台自动创建，用于挂载实验基线与可视化全量测试数据。',
                'active',
                false,
                200,
                'blockchain'
            )
            RETURNING id INTO v_course_id;
        END IF;

        INSERT INTO chapters (
            course_id,
            title,
            description,
            sort_order,
            status
        ) VALUES (
            v_course_id,
            '实验章节：实验基线全量覆盖',
            '覆盖全部实验类型、运行模式和可视化模块的基线章节。',
            1,
            'active'
        )
        RETURNING id INTO v_chapter_id;
    END IF;
END $$;

CREATE TEMP TABLE seed_context ON COMMIT DROP AS
WITH creator AS (
    SELECT id, COALESCE(school_id, 1) AS school_id
    FROM users
    WHERE deleted_at IS NULL
    ORDER BY CASE WHEN role = 'teacher' THEN 0 ELSE 1 END, id
    LIMIT 1
),
chapter AS (
    SELECT id AS chapter_id
    FROM chapters
    WHERE deleted_at IS NULL
    ORDER BY id
    LIMIT 1
)
SELECT creator.id AS creator_id, creator.school_id, chapter.chapter_id
FROM creator
CROSS JOIN chapter;

CREATE TEMP TABLE seed_visualization_modules (
    module_key VARCHAR(100) PRIMARY KEY
) ON COMMIT DROP;

INSERT INTO seed_visualization_modules (module_key) VALUES
('attacks/access_control'),
('attacks/attack_51'),
('attacks/bribery'),
('attacks/bridge_attack'),
('attacks/delegatecall_attack'),
('attacks/dos'),
('attacks/fake_deposit'),
('attacks/flashloan'),
('attacks/frontrun'),
('attacks/governance'),
('attacks/infinite_mint'),
('attacks/integer_overflow'),
('attacks/liquidation'),
('attacks/long_range'),
('attacks/nothing_at_stake'),
('attacks/oracle_manipulation'),
('attacks/reentrancy'),
('attacks/sandwich'),
('attacks/selfdestruct'),
('attacks/selfish_mining'),
('attacks/signature_replay'),
('attacks/timestamp_manipulation'),
('attacks/tx_origin'),
('attacks/weak_randomness'),
('blockchain/block_structure'),
('blockchain/chain_sync'),
('blockchain/difficulty'),
('blockchain/gas'),
('blockchain/light_client'),
('blockchain/mempool'),
('blockchain/merkle_tree'),
('blockchain/mining'),
('blockchain/state_trie'),
('blockchain/transaction'),
('blockchain/utxo'),
('blockchain/wallet'),
('consensus/dag'),
('consensus/dpos'),
('consensus/fork_choice'),
('consensus/hotstuff'),
('consensus/pbft'),
('consensus/pos'),
('consensus/pow'),
('consensus/raft'),
('consensus/tendermint'),
('consensus/vrf'),
('crosschain/atomic_swap'),
('crosschain/bridge'),
('crosschain/data_availability'),
('crosschain/finality'),
('crosschain/ibc'),
('crosschain/light_client'),
('crosschain/optimistic_rollup'),
('crosschain/oracle_bridge'),
('crosschain/plasma'),
('crosschain/relay'),
('crosschain/state_channel'),
('crosschain/zk_rollup'),
('crypto/asymmetric'),
('crypto/bls'),
('crypto/commitment'),
('crypto/elliptic_curve'),
('crypto/encoding'),
('crypto/hash'),
('crypto/kdf'),
('crypto/merkle_proof'),
('crypto/mpc'),
('crypto/randomness'),
('crypto/ring_sig'),
('crypto/signature'),
('crypto/symmetric'),
('crypto/threshold_sig'),
('crypto/zkp'),
('defi/amm'),
('defi/concentrated_liquidity'),
('defi/governance'),
('defi/insurance'),
('defi/interest_model'),
('defi/lending'),
('defi/liquidation'),
('defi/liquidity_pool'),
('defi/options'),
('defi/perpetual'),
('defi/stablecoin'),
('defi/ve_token'),
('defi/yield_aggregator'),
('evm/abi_codec'),
('evm/call_trace'),
('evm/create_create2'),
('evm/delegatecall'),
('evm/disassembler'),
('evm/event_log'),
('evm/evm_executor'),
('evm/gas_profiler'),
('evm/proxy_pattern'),
('evm/state_diff'),
('evm/storage_layout'),
('network/bgp_hijack'),
('network/block_propagation'),
('network/discovery'),
('network/eclipse_attack'),
('network/gossip'),
('network/kademlia'),
('network/nat_traversal'),
('network/partition'),
('network/sybil_attack'),
('network/topology');

CREATE TEMP TABLE seed_experiments (
    title VARCHAR(200) PRIMARY KEY,
    description TEXT,
    type VARCHAR(30),
    mode VARCHAR(20),
    difficulty INTEGER,
    max_score INTEGER,
    pass_score INTEGER,
    auto_grade BOOLEAN,
    grading_strategy VARCHAR(30),
    estimated_time INTEGER,
    sort_order INTEGER,
    status VARCHAR(20),
    allow_late BOOLEAN,
    late_deduction INTEGER,
    visualization_module_key VARCHAR(100)
) ON COMMIT DROP;

INSERT INTO seed_experiments (
    title,
    description,
    type,
    mode,
    difficulty,
    max_score,
    pass_score,
    auto_grade,
    grading_strategy,
    estimated_time,
    sort_order,
    status,
    allow_late,
    late_deduction,
    visualization_module_key
) VALUES
('【实验基线】代码开发：Solidity 安全工程与单元测试', '围绕 Solidity 合约开发、调试、测试与基础安全修复的完整代码开发实验。', 'code_dev', 'single', 2, 100, 60, true, 'checkpoint', 90, 10, 'published', true, 5, NULL),
('【实验基线】命令操作：以太坊命令行部署与运维', '通过命令行完成合约编译、部署、调用、日志检查与环境运维。', 'command_op', 'single', 2, 100, 60, true, 'checkpoint', 80, 20, 'published', true, 5, NULL),
('【实验基线】数据分析：链上交易日志与 Gas 行为分析', '对交易样本、日志与 Gas 行为进行提取、比对和分析。', 'data_analysis', 'single', 3, 110, 66, true, 'checkpoint', 90, 30, 'published', true, 5, NULL),
('【实验基线】工具使用：链上工具链协同排障', '综合使用 IDE、终端、区块浏览器、链上接口与 API 调试台完成链路排障。', 'tool_usage', 'single', 2, 100, 60, true, 'checkpoint', 75, 40, 'published', true, 5, NULL),
('【实验基线】配置调试：多服务配置调优与依赖治理', '在多节点环境下调整运行参数、依赖与服务暴露配置，完成稳定性修复。', 'config_debug', 'multi_node', 3, 120, 72, true, 'checkpoint', 100, 50, 'published', true, 5, NULL),
('【实验基线】逆向分析：EVM 字节码逆向与漏洞定位', '对 EVM 字节码与调用轨迹进行逆向分析，定位关键风险点。', 'reverse', 'single', 4, 130, 78, true, 'checkpoint', 110, 60, 'published', true, 5, NULL),
('【实验基线】故障排查：节点故障注入与恢复演练', '模拟节点中断、网络分区与链路抖动，完成故障恢复与连通性验证。', 'troubleshoot', 'multi_node', 3, 120, 72, true, 'checkpoint', 100, 70, 'published', true, 5, NULL),
('【实验基线】多人协作：攻防推演实验', '攻击手、防守手与观察员围绕同一实验环境协作完成复现、隔离与复盘。', 'collaboration', 'collaboration', 4, 140, 84, true, 'checkpoint', 120, 80, 'published', true, 5, NULL);

INSERT INTO seed_experiments (
    title,
    description,
    type,
    mode,
    difficulty,
    max_score,
    pass_score,
    auto_grade,
    grading_strategy,
    estimated_time,
    sort_order,
    status,
    allow_late,
    late_deduction,
    visualization_module_key
)
SELECT
    '【实验基线】可视化：' || module_key,
    CASE split_part(module_key, '/', 1)
        WHEN 'attacks' THEN '围绕 ' || module_key || ' 模块观察攻击路径、状态变化与防御要点。'
        WHEN 'blockchain' THEN '围绕 ' || module_key || ' 模块观察区块链结构、状态与数据流转。'
        WHEN 'consensus' THEN '围绕 ' || module_key || ' 模块观察共识流程、阶段推进与节点状态。'
        WHEN 'crosschain' THEN '围绕 ' || module_key || ' 模块观察跨链消息、验证与资产流转。'
        WHEN 'crypto' THEN '围绕 ' || module_key || ' 模块观察密码学计算过程、输入输出与验证效果。'
        WHEN 'defi' THEN '围绕 ' || module_key || ' 模块观察协议状态、价格变化与资金路径。'
        WHEN 'evm' THEN '围绕 ' || module_key || ' 模块观察 EVM 执行、状态差异与调用轨迹。'
        WHEN 'network' THEN '围绕 ' || module_key || ' 模块观察网络拓扑、传播过程与攻击影响。'
        ELSE '围绕 ' || module_key || ' 模块完成实验基线验证。'
    END,
    'visualization',
    'single',
    CASE split_part(module_key, '/', 1)
        WHEN 'attacks' THEN 3
        WHEN 'consensus' THEN 3
        WHEN 'defi' THEN 3
        WHEN 'evm' THEN 3
        ELSE 2
    END,
    100,
    60,
    true,
    'checkpoint',
    CASE split_part(module_key, '/', 1)
        WHEN 'consensus' THEN 100
        WHEN 'attacks' THEN 95
        WHEN 'defi' THEN 95
        WHEN 'crosschain' THEN 95
        ELSE 80
    END,
    100 + ROW_NUMBER() OVER (ORDER BY module_key),
    'published',
    true,
    5,
    module_key
FROM seed_visualization_modules;

CREATE TEMP TABLE cleanup_titles ON COMMIT DROP AS
SELECT title FROM seed_experiments
UNION
SELECT legacy_title
FROM (
    VALUES
    ('区块结构与交易生命周期可视化实验'),
    ('Solidity 安全工程与单元测试实验'),
    ('以太坊命令行部署与运维实验'),
    ('链上交易日志与 Gas 行为分析实验'),
    ('链上工具链协同排障实验'),
    ('多服务配置调优与依赖治理实验'),
    ('EVM 字节码逆向与漏洞定位实验'),
    ('节点故障注入与恢复演练实验'),
    ('多人协作攻防推演实验'),
    ('DeFi 攻击路径可视化复现实验')
) AS legacy(legacy_title);

CREATE TEMP TABLE target_experiments ON COMMIT DROP AS
SELECT id
FROM experiments
WHERE title IN (SELECT title FROM cleanup_titles);

CREATE TEMP TABLE target_submissions ON COMMIT DROP AS
SELECT id
FROM submissions
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_envs ON COMMIT DROP AS
SELECT id
FROM experiment_envs
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_runtime_instances ON COMMIT DROP AS
SELECT id
FROM experiment_runtime_instances
WHERE experiment_env_id IN (SELECT id FROM target_envs);

CREATE TEMP TABLE target_sessions ON COMMIT DROP AS
SELECT id
FROM experiment_sessions
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_collaborations ON COMMIT DROP AS
SELECT id
FROM experiment_collaborations
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_role_bindings ON COMMIT DROP AS
SELECT id
FROM experiment_role_bindings
WHERE collaboration_id IN (SELECT id FROM target_collaborations);

CREATE TEMP TABLE target_workspaces ON COMMIT DROP AS
SELECT id
FROM experiment_workspaces
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_topologies ON COMMIT DROP AS
SELECT id
FROM experiment_topologies
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_nodes ON COMMIT DROP AS
SELECT id
FROM experiment_nodes
WHERE experiment_id IN (SELECT id FROM target_experiments);

CREATE TEMP TABLE target_services ON COMMIT DROP AS
SELECT id
FROM experiment_services
WHERE experiment_id IN (SELECT id FROM target_experiments);

DELETE FROM submission_check_results
WHERE submission_id IN (SELECT id FROM target_submissions);

DELETE FROM submissions
WHERE id IN (SELECT id FROM target_submissions);

DELETE FROM experiment_runtime_tools
WHERE runtime_instance_id IN (SELECT id FROM target_runtime_instances);

DELETE FROM experiment_runtime_instances
WHERE id IN (SELECT id FROM target_runtime_instances);

DELETE FROM experiment_session_messages
WHERE session_id IN (SELECT id FROM target_sessions);

DELETE FROM experiment_session_members
WHERE session_id IN (SELECT id FROM target_sessions);

DELETE FROM experiment_envs
WHERE id IN (SELECT id FROM target_envs);

DELETE FROM experiment_sessions
WHERE id IN (SELECT id FROM target_sessions);

DELETE FROM experiment_role_binding_tools
WHERE role_binding_id IN (SELECT id FROM target_role_bindings);

DELETE FROM experiment_role_binding_nodes
WHERE role_binding_id IN (SELECT id FROM target_role_bindings);

DELETE FROM experiment_role_bindings
WHERE id IN (SELECT id FROM target_role_bindings);

DELETE FROM experiment_collaborations
WHERE id IN (SELECT id FROM target_collaborations);

DELETE FROM experiment_workspace_tools
WHERE workspace_id IN (SELECT id FROM target_workspaces);

DELETE FROM experiment_workspaces
WHERE id IN (SELECT id FROM target_workspaces);

DELETE FROM experiment_topology_exposed_entries
WHERE topology_id IN (SELECT id FROM target_topologies);

DELETE FROM experiment_topologies
WHERE id IN (SELECT id FROM target_topologies);

DELETE FROM experiment_tools
WHERE experiment_id IN (SELECT id FROM target_experiments);

DELETE FROM experiment_service_ports
WHERE service_id IN (SELECT id FROM target_services);

DELETE FROM experiment_service_env_vars
WHERE service_id IN (SELECT id FROM target_services);

DELETE FROM experiment_services
WHERE id IN (SELECT id FROM target_services);

DELETE FROM experiment_node_ports
WHERE node_id IN (SELECT id FROM target_nodes);

DELETE FROM experiment_node_tools
WHERE node_id IN (SELECT id FROM target_nodes);

DELETE FROM experiment_nodes
WHERE id IN (SELECT id FROM target_nodes);

DELETE FROM experiment_init_scripts
WHERE experiment_id IN (SELECT id FROM target_experiments);

DELETE FROM experiment_assets
WHERE experiment_id IN (SELECT id FROM target_experiments);

DELETE FROM experiment_checkpoints
WHERE experiment_id IN (SELECT id FROM target_experiments);

DELETE FROM experiments
WHERE id IN (SELECT id FROM target_experiments);

INSERT INTO experiments (
    school_id,
    chapter_id,
    creator_id,
    title,
    description,
    type,
    mode,
    difficulty,
    estimated_time,
    max_score,
    pass_score,
    auto_grade,
    grading_strategy,
    sort_order,
    status,
    allow_late,
    late_deduction
)
SELECT
    ctx.school_id,
    ctx.chapter_id,
    ctx.creator_id,
    se.title,
    se.description,
    se.type,
    se.mode,
    se.difficulty,
    se.estimated_time,
    se.max_score,
    se.pass_score,
    se.auto_grade,
    se.grading_strategy,
    se.sort_order,
    se.status,
    se.allow_late,
    se.late_deduction
FROM seed_experiments se
CROSS JOIN seed_context ctx;

CREATE TEMP TABLE seeded_experiments ON COMMIT DROP AS
SELECT
    e.id,
    s.title,
    s.description,
    s.type,
    s.mode,
    s.difficulty,
    s.max_score,
    s.pass_score,
    s.auto_grade,
    s.grading_strategy,
    s.estimated_time,
    s.sort_order,
    s.status,
    s.allow_late,
    s.late_deduction,
    s.visualization_module_key
FROM experiments e
JOIN seed_experiments s ON s.title = e.title;

INSERT INTO experiment_workspaces (
    experiment_id,
    image,
    display_name,
    cpu,
    memory,
    storage
)
SELECT
    se.id,
    CASE
        WHEN se.type IN ('tool_usage', 'reverse', 'troubleshoot') THEN 'chainspace/security:latest'
        ELSE 'chainspace/eth-dev:latest'
    END,
    CASE
        WHEN se.type = 'visualization' THEN '可视化实验工作区'
        WHEN se.mode = 'collaboration' THEN '协作实验工作区'
        ELSE '实验工作区'
    END,
    CASE
        WHEN se.mode = 'collaboration' THEN '1000m'
        WHEN se.mode = 'multi_node' THEN '900m'
        ELSE '700m'
    END,
    CASE
        WHEN se.mode = 'collaboration' THEN '2Gi'
        WHEN se.mode = 'multi_node' THEN '1536Mi'
        ELSE '1Gi'
    END,
    CASE
        WHEN se.type = 'visualization' THEN '6Gi'
        ELSE '10Gi'
    END
FROM seeded_experiments se;

INSERT INTO experiment_workspace_tools (workspace_id, tool_key, sort_order)
SELECT
    ew.id,
    tool.tool_key,
    tool.sort_order
FROM experiment_workspaces ew
JOIN seeded_experiments se ON se.id = ew.experiment_id
JOIN LATERAL (
    VALUES
    ('ide', 0),
    ('terminal', 1),
    ('files', 2),
    ('logs', 3),
    ('network', 4)
) AS tool(tool_key, sort_order) ON TRUE
WHERE (
    (se.type = 'visualization' AND tool.tool_key IN ('terminal', 'logs'))
    OR (se.type = 'code_dev' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs'))
    OR (se.type = 'command_op' AND tool.tool_key IN ('terminal', 'files', 'logs'))
    OR (se.type = 'data_analysis' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs'))
    OR (se.type = 'tool_usage' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs'))
    OR (se.type = 'config_debug' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs'))
    OR (se.type = 'reverse' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs'))
    OR (se.type = 'troubleshoot' AND tool.tool_key IN ('terminal', 'logs', 'network'))
    OR (se.type = 'collaboration' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs', 'network'))
);

INSERT INTO experiment_topologies (
    experiment_id,
    template,
    shared_network
)
SELECT
    se.id,
    CASE
        WHEN se.mode IN ('multi_node', 'collaboration') THEN 'multi_role_lab'
        WHEN se.type IN ('visualization', 'command_op', 'data_analysis', 'tool_usage', 'config_debug', 'troubleshoot', 'collaboration') THEN 'workspace_with_services'
        ELSE 'workspace_only'
    END,
    CASE
        WHEN se.mode IN ('multi_node', 'collaboration') THEN true
        ELSE false
    END
FROM seeded_experiments se;

INSERT INTO experiment_topology_exposed_entries (topology_id, entry_key, sort_order)
SELECT
    et.id,
    exposed.entry_key,
    exposed.sort_order
FROM experiment_topologies et
JOIN seeded_experiments se ON se.id = et.experiment_id
JOIN LATERAL (
    VALUES
    ('workspace', 0),
    ('rpc', 1),
    ('explorer', 2),
    ('api_debug', 3),
    ('visualization', 4),
    ('network', 5)
) AS exposed(entry_key, sort_order) ON TRUE
WHERE
    exposed.entry_key = 'workspace'
    OR (exposed.entry_key = 'visualization' AND se.type = 'visualization')
    OR (exposed.entry_key = 'rpc' AND se.type IN ('code_dev', 'command_op', 'data_analysis', 'tool_usage', 'config_debug', 'troubleshoot', 'collaboration'))
    OR (exposed.entry_key = 'explorer' AND se.type IN ('data_analysis', 'tool_usage', 'collaboration'))
    OR (exposed.entry_key = 'api_debug' AND se.type IN ('command_op', 'tool_usage', 'config_debug', 'troubleshoot', 'collaboration'))
    OR (exposed.entry_key = 'network' AND se.type IN ('troubleshoot', 'collaboration'));

INSERT INTO experiment_tools (
    experiment_id,
    tool_key,
    label,
    kind,
    target,
    student_facing,
    sort_order
)
SELECT
    se.id,
    tool.tool_key,
    tool.label,
    CASE
        WHEN tool.tool_key = 'visualization' THEN se.visualization_module_key
        ELSE tool.kind
    END,
    CASE
        WHEN tool.tool_key = 'visualization' THEN 'simulation'
        WHEN tool.tool_key = 'rpc' THEN 'geth'
        WHEN tool.tool_key = 'explorer' THEN
            CASE
                WHEN se.type IN ('data_analysis', 'tool_usage', 'collaboration') THEN 'blockscout'
                ELSE 'geth'
            END
        WHEN tool.tool_key = 'api_debug' THEN
            CASE
                WHEN se.type = 'tool_usage' THEN 'chainlink'
                WHEN se.type = 'config_debug' THEN 'thegraph'
                WHEN se.type = 'troubleshoot' THEN 'chainlink'
                ELSE 'geth'
            END
        ELSE tool.target
    END,
    true,
    tool.sort_order
FROM seeded_experiments se
JOIN LATERAL (
    VALUES
    ('ide', 'IDE', 'workspace', 'workspace', 0),
    ('terminal', '终端', 'workspace', 'workspace', 1),
    ('files', '文件管理', 'workspace', 'workspace', 2),
    ('logs', '日志面板', 'workspace', 'workspace', 3),
    ('rpc', '链上接口', 'rpc', 'geth', 4),
    ('explorer', '区块浏览器', 'explorer', 'geth', 5),
    ('api_debug', '接口调试台', 'api_debug', 'geth', 6),
    ('visualization', '可视化沙箱', 'visualization', 'simulation', 7),
    ('network', '组网面板', 'network', 'workspace', 8)
) AS tool(tool_key, label, kind, target, sort_order) ON TRUE
WHERE (
    (se.type = 'visualization' AND tool.tool_key IN ('terminal', 'logs', 'visualization'))
    OR (se.type = 'code_dev' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs', 'rpc'))
    OR (se.type = 'command_op' AND tool.tool_key IN ('terminal', 'files', 'logs', 'rpc', 'api_debug'))
    OR (se.type = 'data_analysis' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs', 'rpc', 'explorer'))
    OR (se.type = 'tool_usage' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs', 'rpc', 'explorer', 'api_debug'))
    OR (se.type = 'config_debug' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs', 'rpc', 'api_debug'))
    OR (se.type = 'reverse' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs'))
    OR (se.type = 'troubleshoot' AND tool.tool_key IN ('terminal', 'logs', 'network', 'rpc', 'api_debug'))
    OR (se.type = 'collaboration' AND tool.tool_key IN ('ide', 'terminal', 'files', 'logs', 'network', 'rpc', 'explorer', 'api_debug'))
);

INSERT INTO experiment_services (
    experiment_id,
    service_key,
    name,
    image,
    role,
    purpose,
    student_facing,
    sort_order
)
SELECT
    se.id,
    service.service_key,
    service.name,
    service.image,
    service.role,
    service.purpose,
    true,
    service.sort_order
FROM seeded_experiments se
JOIN LATERAL (
    VALUES
    ('simulation', '可视化模拟器', 'chainspace/simulation:latest', 'visualization', 'interactive_visualization', 0),
    ('geth', '以太坊节点', 'chainspace/geth:latest', 'rpc', 'chain_runtime', 1),
    ('blockscout', '区块浏览器服务', 'chainspace/blockscout:latest', 'explorer', 'block_explorer', 2),
    ('chainlink', '预言机节点服务', 'chainspace/chainlink:latest', 'oracle', 'oracle_runtime', 3),
    ('thegraph', '索引服务', 'chainspace/thegraph:latest', 'indexer', 'graph_query', 4)
) AS service(service_key, name, image, role, purpose, sort_order) ON TRUE
WHERE
    (se.type = 'visualization' AND service.service_key = 'simulation')
    OR (se.type IN ('code_dev', 'command_op', 'data_analysis', 'tool_usage', 'config_debug', 'troubleshoot', 'collaboration') AND service.service_key = 'geth')
    OR (se.type IN ('data_analysis', 'tool_usage', 'collaboration') AND service.service_key = 'blockscout')
    OR (se.type IN ('tool_usage', 'troubleshoot') AND service.service_key = 'chainlink')
    OR (se.type IN ('data_analysis', 'config_debug') AND service.service_key = 'thegraph');

INSERT INTO experiment_service_ports (service_id, port, sort_order)
SELECT
    es.id,
    port.port,
    port.sort_order
FROM experiment_services es
JOIN LATERAL (
    VALUES
    (8080, 0),
    (8545, 0),
    (4000, 0),
    (6688, 0),
    (8000, 0)
) AS port(port, sort_order) ON TRUE
WHERE
    (es.service_key = 'simulation' AND port.port = 8080)
    OR (es.service_key = 'geth' AND port.port = 8545)
    OR (es.service_key = 'blockscout' AND port.port = 4000)
    OR (es.service_key = 'chainlink' AND port.port = 6688)
    OR (es.service_key = 'thegraph' AND port.port = 8000);

INSERT INTO experiment_service_env_vars (service_id, env_key, env_value, sort_order)
SELECT
    es.id,
    env.env_key,
    env.env_value,
    env.sort_order
FROM experiment_services es
JOIN LATERAL (
    VALUES
    ('CHAIN_ID', '11155111', 0),
    ('WS_MODE', 'true', 1),
    ('LOG_LEVEL', 'info', 2)
) AS env(env_key, env_value, sort_order) ON TRUE
WHERE es.service_key = 'geth';

INSERT INTO experiment_service_env_vars (service_id, env_key, env_value, sort_order)
SELECT
    es.id,
    env.env_key,
    env.env_value,
    env.sort_order
FROM experiment_services es
JOIN LATERAL (
    VALUES
    ('NETWORK', 'ChainSpace', 0),
    ('SUBNETWORK', 'Experiment', 1),
    ('COIN', 'ETH', 2)
) AS env(env_key, env_value, sort_order) ON TRUE
WHERE es.service_key = 'blockscout';

INSERT INTO experiment_service_env_vars (service_id, env_key, env_value, sort_order)
SELECT
    es.id,
    env.env_key,
    env.env_value,
    env.sort_order
FROM experiment_services es
JOIN LATERAL (
    VALUES
    ('CHAINLINK_EVM_CHAIN_ID', '11155111', 0),
    ('CHAINLINK_API_EMAIL', 'admin@chainspace.local', 1)
) AS env(env_key, env_value, sort_order) ON TRUE
WHERE es.service_key = 'chainlink';

INSERT INTO experiment_service_env_vars (service_id, env_key, env_value, sort_order)
SELECT
    es.id,
    env.env_key,
    env.env_value,
    env.sort_order
FROM experiment_services es
JOIN LATERAL (
    VALUES
    ('GRAPH_LOG', 'info', 0)
) AS env(env_key, env_value, sort_order) ON TRUE
WHERE es.service_key = 'thegraph';

INSERT INTO experiment_service_env_vars (service_id, env_key, env_value, sort_order)
SELECT
    es.id,
    env.env_key,
    env.env_value,
    env.sort_order
FROM experiment_services es
JOIN seeded_experiments se ON se.id = es.experiment_id
JOIN LATERAL (
    VALUES
    ('SIMULATION_MODULE', COALESCE(se.visualization_module_key, 'blockchain/block_structure'), 0),
    ('SIMULATION_PROFILE', 'teaching', 1)
) AS env(env_key, env_value, sort_order) ON TRUE
WHERE es.service_key = 'simulation';

INSERT INTO experiment_nodes (
    experiment_id,
    node_key,
    name,
    image,
    role,
    cpu,
    memory,
    storage,
    student_facing,
    sort_order
)
SELECT
    se.id,
    node.node_key,
    node.name,
    'chainspace/geth:latest',
    node.role,
    node.cpu,
    node.memory,
    node.storage,
    true,
    node.sort_order
FROM seeded_experiments se
JOIN LATERAL (
    VALUES
    ('validator-a', '验证节点 A', 'validator', '700m', '1Gi', '8Gi', 0),
    ('validator-b', '验证节点 B', 'validator', '700m', '1Gi', '8Gi', 1),
    ('observer', '观察节点', 'observer', '600m', '1Gi', '8Gi', 2)
) AS node(node_key, name, role, cpu, memory, storage, sort_order) ON TRUE
WHERE se.mode IN ('multi_node', 'collaboration');

INSERT INTO experiment_node_ports (node_id, port, sort_order)
SELECT
    en.id,
    port.port,
    port.sort_order
FROM experiment_nodes en
JOIN LATERAL (
    VALUES
    (8545, 0),
    (8546, 1),
    (30303, 2)
) AS port(port, sort_order) ON TRUE;

INSERT INTO experiment_node_tools (node_id, tool_key, sort_order)
SELECT
    en.id,
    tool.tool_key,
    tool.sort_order
FROM experiment_nodes en
JOIN LATERAL (
    VALUES
    ('terminal', 0),
    ('logs', 1),
    ('rpc', 2),
    ('network', 3),
    ('explorer', 4)
) AS tool(tool_key, sort_order) ON TRUE
WHERE
    (en.role = 'validator' AND tool.tool_key IN ('terminal', 'logs', 'rpc', 'network'))
    OR (en.role = 'observer' AND tool.tool_key IN ('logs', 'rpc', 'explorer'));

INSERT INTO experiment_init_scripts (
    experiment_id,
    scope_type,
    scope_key,
    script,
    sort_order
)
SELECT
    se.id,
    'workspace',
    'workspace',
    script.script,
    script.sort_order
FROM seeded_experiments se
JOIN LATERAL (
    VALUES
    ('mkdir -p /workspace/output /workspace/fixtures', 0),
    ('echo "workspace initialized $(date -u +%FT%TZ)" > /workspace/output/init.log', 1)
) AS script(script, sort_order) ON TRUE;

INSERT INTO experiment_init_scripts (
    experiment_id,
    scope_type,
    scope_key,
    script,
    sort_order
)
SELECT
    se.id,
    'workspace',
    'workspace',
    'printf ''%s\n'' ''' || se.type || ''' > /workspace/output/experiment_type.txt',
    2
FROM seeded_experiments se;

INSERT INTO experiment_init_scripts (
    experiment_id,
    scope_type,
    scope_key,
    script,
    sort_order
)
SELECT
    se.id,
    'workspace',
    'workspace',
    'printf ''%s\n'' ''' || se.visualization_module_key || ''' > /workspace/output/visualization_module.txt',
    3
FROM seeded_experiments se
WHERE se.visualization_module_key IS NOT NULL;

INSERT INTO experiment_init_scripts (
    experiment_id,
    scope_type,
    scope_key,
    script,
    sort_order
)
SELECT
    se.id,
    'content',
    'content',
    script.script,
    script.sort_order
FROM seeded_experiments se
JOIN LATERAL (
    VALUES
    ('cp -r /workspace/fixtures/* /workspace/ 2>/dev/null || true', 10),
    ('echo "content bootstrap done" >> /workspace/output/init.log', 11)
) AS script(script, sort_order) ON TRUE;

INSERT INTO experiment_init_scripts (
    experiment_id,
    scope_type,
    scope_key,
    script,
    sort_order
)
SELECT
    se.id,
    'node',
    en.node_key,
    'echo "node=' || en.node_key || ' ready $(date -u +%FT%TZ)" > /workspace/node-ready.log',
    20
FROM seeded_experiments se
JOIN experiment_nodes en ON en.experiment_id = se.id;

INSERT INTO experiment_assets (
    experiment_id,
    asset_key,
    name,
    source_type,
    bucket,
    object_path,
    mount_path,
    required,
    sort_order
)
SELECT
    se.id,
    'runbook',
    '实验说明手册',
    'object_storage',
    'chainspace',
    'experiments/assets/lab_runbook.md',
    '/workspace/fixtures/lab_runbook.md',
    true,
    0
FROM seeded_experiments se;

INSERT INTO experiment_assets (
    experiment_id,
    asset_key,
    name,
    source_type,
    bucket,
    object_path,
    mount_path,
    required,
    sort_order
)
SELECT
    se.id,
    'context',
    CASE
        WHEN se.type = 'visualization' THEN '可视化上下文数据'
        ELSE '实验上下文数据'
    END,
    'object_storage',
    'chainspace',
    CASE
        WHEN se.visualization_module_key IS NOT NULL THEN
            CASE split_part(se.visualization_module_key, '/', 1)
                WHEN 'attacks' THEN 'experiments/assets/reentrancy_attack_trace.txt'
                WHEN 'blockchain' THEN 'experiments/assets/block_and_tx_dataset.txt'
                WHEN 'consensus' THEN 'experiments/assets/consensus_fault_injection.txt'
                WHEN 'crosschain' THEN 'experiments/assets/network_topology_events.txt'
                WHEN 'crypto' THEN 'experiments/assets/block_and_tx_dataset.txt'
                WHEN 'defi' THEN 'experiments/assets/reentrancy_attack_trace.txt'
                WHEN 'evm' THEN 'experiments/assets/evm_bytecode_samples.txt'
                WHEN 'network' THEN 'experiments/assets/network_topology_events.txt'
                ELSE 'experiments/assets/block_and_tx_dataset.txt'
            END
        WHEN se.type = 'reverse' THEN 'experiments/assets/evm_bytecode_samples.txt'
        WHEN se.type IN ('config_debug', 'troubleshoot') THEN 'experiments/assets/consensus_fault_injection.txt'
        WHEN se.type IN ('tool_usage', 'collaboration') THEN 'experiments/assets/network_topology_events.txt'
        WHEN se.type = 'command_op' THEN 'experiments/assets/ethereum_log_samples.txt'
        WHEN se.type = 'data_analysis' THEN 'experiments/assets/block_and_tx_dataset.txt'
        ELSE 'experiments/assets/reentrancy_attack_trace.txt'
    END,
    '/workspace/fixtures/context_asset.txt',
    true,
    1
FROM seeded_experiments se;

INSERT INTO experiment_assets (
    experiment_id,
    asset_key,
    name,
    source_type,
    bucket,
    object_path,
    mount_path,
    required,
    sort_order
)
SELECT
    se.id,
    'logs',
    '链上日志样本',
    'object_storage',
    'chainspace',
    'experiments/assets/ethereum_log_samples.txt',
    '/workspace/fixtures/ethereum_log_samples.txt',
    CASE
        WHEN se.type IN ('command_op', 'data_analysis', 'tool_usage', 'config_debug', 'troubleshoot', 'collaboration') THEN true
        ELSE false
    END,
    2
FROM seeded_experiments se
WHERE se.type IN ('command_op', 'data_analysis', 'tool_usage', 'config_debug', 'troubleshoot', 'collaboration');

INSERT INTO experiment_checkpoints (
    experiment_id,
    checkpoint_key,
    type,
    target,
    path,
    command,
    expected,
    script,
    score,
    sort_order
)
SELECT
    se.id,
    checkpoint.checkpoint_key,
    checkpoint.type,
    checkpoint.target,
    checkpoint.path,
    checkpoint.command,
    checkpoint.expected,
    checkpoint.script,
    checkpoint.score,
    checkpoint.sort_order
FROM seeded_experiments se
JOIN LATERAL (
    VALUES
    ('init-log', 'file_exists', 'workspace', '/workspace/output/init.log', NULL, NULL, NULL, 20, 0),
    ('fixtures-list', 'command_exec', 'workspace', NULL, 'ls /workspace/fixtures', NULL, NULL, 20, 1),
    ('type-marker', 'file_content', 'workspace', '/workspace/output/experiment_type.txt', NULL, se.type, NULL, 25, 2)
) AS checkpoint(checkpoint_key, type, target, path, command, expected, script, score, sort_order) ON TRUE;

INSERT INTO experiment_checkpoints (
    experiment_id,
    checkpoint_key,
    type,
    target,
    path,
    command,
    expected,
    script,
    score,
    sort_order
)
SELECT
    se.id,
    'visualization-module',
    'file_content',
    'workspace',
    '/workspace/output/visualization_module.txt',
    NULL,
    se.visualization_module_key,
    NULL,
    35,
    3
FROM seeded_experiments se
WHERE se.visualization_module_key IS NOT NULL;

INSERT INTO experiment_checkpoints (
    experiment_id,
    checkpoint_key,
    type,
    target,
    path,
    command,
    expected,
    script,
    score,
    sort_order
)
SELECT
    se.id,
    'content-bootstrap',
    'custom_script',
    'workspace',
    NULL,
    NULL,
    NULL,
    'test -f /workspace/fixtures/lab_runbook.md && test -f /workspace/fixtures/context_asset.txt && grep -q "content bootstrap done" /workspace/output/init.log',
    35,
    3
FROM seeded_experiments se
WHERE se.visualization_module_key IS NULL;

INSERT INTO experiment_collaborations (experiment_id, max_members)
SELECT id, 4
FROM seeded_experiments
WHERE type = 'collaboration';

INSERT INTO experiment_role_bindings (
    collaboration_id,
    role_key,
    label,
    sort_order
)
SELECT
    ec.id,
    role.role_key,
    role.label,
    role.sort_order
FROM experiment_collaborations ec
JOIN seeded_experiments se ON se.id = ec.experiment_id
JOIN LATERAL (
    VALUES
    ('attacker', '攻击手', 0),
    ('defender', '防守手', 1),
    ('observer', '观察员', 2)
) AS role(role_key, label, sort_order) ON TRUE
WHERE se.type = 'collaboration';

INSERT INTO experiment_role_binding_nodes (
    role_binding_id,
    node_key,
    sort_order
)
SELECT
    rb.id,
    CASE rb.role_key
        WHEN 'attacker' THEN 'validator-a'
        WHEN 'defender' THEN 'validator-b'
        ELSE 'observer'
    END,
    0
FROM experiment_role_bindings rb
JOIN experiment_collaborations ec ON ec.id = rb.collaboration_id
JOIN seeded_experiments se ON se.id = ec.experiment_id
WHERE se.type = 'collaboration';

INSERT INTO experiment_role_binding_tools (
    role_binding_id,
    tool_key,
    sort_order
)
SELECT
    rb.id,
    tool.tool_key,
    tool.sort_order
FROM experiment_role_bindings rb
JOIN experiment_collaborations ec ON ec.id = rb.collaboration_id
JOIN seeded_experiments se ON se.id = ec.experiment_id
JOIN LATERAL (
    VALUES
    ('terminal', 0),
    ('logs', 1),
    ('network', 2),
    ('rpc', 3),
    ('api_debug', 4),
    ('explorer', 5)
) AS tool(tool_key, sort_order) ON TRUE
WHERE se.type = 'collaboration'
  AND (
      (rb.role_key IN ('attacker', 'defender') AND tool.tool_key IN ('terminal', 'logs', 'network', 'rpc', 'api_debug'))
      OR (rb.role_key = 'observer' AND tool.tool_key IN ('logs', 'explorer'))
  );

COMMIT;
