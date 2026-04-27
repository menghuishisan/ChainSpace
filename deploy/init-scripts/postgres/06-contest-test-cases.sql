-- ============================================
-- ChainSpace 比赛基线数据（解题赛 + 对抗赛）
-- 附件对象路径对应本地 MinIO 初始化目录：
-- deploy/init-scripts/minio/chainspace/challenges/0/attachments/*
-- ============================================

BEGIN;

-- 清理历史基线数据
DELETE FROM agent_battle_scores
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM agent_battle_events
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM agent_contracts
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM agent_battle_rounds
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM contest_scores
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM contest_submissions
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM challenge_envs
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM contest_registrations
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM team_members
WHERE team_id IN (
    SELECT id FROM teams WHERE contest_id IN (
        SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛')
    )
);
DELETE FROM teams
WHERE contest_id IN (SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛'));
DELETE FROM contest_challenges
WHERE contest_id IN (SELECT id FROM contests WHERE title = '链境公开解题赛（全形态覆盖）');
DELETE FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛');
DELETE FROM challenges WHERE source_ref LIKE 'seed:baseline:contest:%';

-- 1) 解题赛题目：9分类 + 4运行形态
WITH base AS (
    SELECT
        COALESCE((SELECT id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1)) AS creator_id,
        COALESCE((SELECT school_id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT school_id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1), 1) AS school_id
)
INSERT INTO challenges (
    creator_id, school_id, title, description, category, runtime_profile, difficulty,
    base_points, min_points, decay_factor, contract_code, deploy_script, check_script,
    flag_template, flag_type, flag_regex, flag_secret, validation_config, hints, attachments,
    challenge_orchestration, tags, source_type, source_ref, status, is_public
)
SELECT
    creator_id, school_id,
    '合约权限绕过：Vault Keeper',
    '修复 owner 校验缺失导致的敏感迁移漏洞。',
    'contract_vuln', 'single_chain_instance', 2,
    400, 220, 0.12,
    'pragma solidity ^0.8.20; contract VaultKeeper { address public owner; constructor(){owner=msg.sender;} function migrate(address) external {} }',
    '#!/usr/bin/env bash
forge test -vv',
    '#!/usr/bin/env bash
forge test --match-test test_Patch',
    'flag{vault-keeper-fixed}', 'dynamic', '^flag\\{[a-z0-9\\-]{8,80}\\}$', 'vault-keeper-2026',
    '{"mode":"script","entry":"scripts/validate.sh"}'::jsonb,
    '["定位敏感函数权限边界。"]'::jsonb,
    '["s3://chainspace/challenges/0/attachments/contract_vuln_vault_keeper.md"]'::jsonb,
    '{"mode":"single_chain_instance","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files"]},"services":[{"key":"chain","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc"]},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":120}}'::jsonb,
    '["access-control","solidity"]'::jsonb,
    'preset', 'seed:baseline:contest:contract_vuln', 'active', true
FROM base;

WITH base AS (
    SELECT
        COALESCE((SELECT id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1)) AS creator_id,
        COALESCE((SELECT school_id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT school_id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1), 1) AS school_id
)
INSERT INTO challenges (
    creator_id, school_id, title, description, category, runtime_profile, difficulty,
    base_points, min_points, decay_factor, contract_code, deploy_script, check_script,
    flag_template, flag_type, flag_regex, flag_secret, validation_config, hints, attachments,
    challenge_orchestration, tags, source_type, source_ref, status, is_public
)
SELECT
    creator_id, school_id,
    'Visor Finance 主网 Fork 复现',
    '基于真实攻击交易复现重入并完成修复验证。',
    'defi', 'fork_replay', 4,
    700, 360, 0.08,
    'pragma solidity ^0.8.20; interface IVisorLike { function withdraw(uint256,address,address,address[] calldata) external; }',
    '#!/usr/bin/env bash
forge test --match-test test_VisorForkReplay -vv',
    '#!/usr/bin/env bash
forge test --match-test test_NoReentrancy -vv',
    'flag{visor-mainnet-replay}', 'dynamic', '^flag\\{[a-z0-9\\-]{8,80}\\}$', 'visor-fork-2026',
    '{"mode":"fork_replay","target_tx":"0x69272d8c84d67d1da2f6425b339192fa472898dce936f24818fda415c1c1ff3f"}'::jsonb,
    '["fork 到 13654410 块并复现攻击。"]'::jsonb,
    '["s3://chainspace/challenges/0/attachments/visor_finance_incident_brief.md","s3://chainspace/challenges/0/attachments/visor_finance_attack_trace.txt"]'::jsonb,
    '{"mode":"fork_replay","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files","logs"]},"services":[{"key":"fork-node","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc"]},"fork":{"enabled":true,"chain":"ethereum","chain_id":1,"rpc_url":"https://ethereum.publicnode.com","block_number":13654410,"target_tx_hash":"0x69272d8c84d67d1da2f6425b339192fa472898dce936f24818fda415c1c1ff3f"},"scenario":{"contract_address":"0xC9f27A50f82571C1C8423A42970613b8dBDA14ef"},"lifecycle":{"time_limit_minutes":150}}'::jsonb,
    '["defi","fork","mainnet"]'::jsonb,
    'preset', 'seed:baseline:contest:defi_fork', 'active', true
FROM base;

WITH base AS (
    SELECT
        COALESCE((SELECT id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1)) AS creator_id,
        COALESCE((SELECT school_id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT school_id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1), 1) AS school_id
)
INSERT INTO challenges (
    creator_id, school_id, title, description, category, runtime_profile, difficulty,
    base_points, min_points, decay_factor, contract_code, deploy_script, check_script,
    flag_template, flag_type, flag_regex, flag_secret, validation_config, hints, attachments,
    challenge_orchestration, tags, source_type, source_ref, status, is_public
)
SELECT creator_id, school_id, title, description, category, runtime_profile, difficulty,
       base_points, min_points, decay_factor, contract_code, deploy_script, check_script,
       flag_template, 'dynamic', '^flag\\{[a-z0-9\\-]{8,80}\\}$', flag_secret, validation_config::jsonb, hints::jsonb, attachments::jsonb,
       challenge_orchestration::jsonb, tags::jsonb, 'preset', source_ref, 'active', true
FROM base,
LATERAL (
    VALUES
    (
        'PBFT 共识容错参数调优', '修复 view-change 参数漂移。', 'consensus', 'multi_service_lab', 3, 480, 260, 0.11,
        'pragma solidity ^0.8.20; contract ConsensusMarker { uint256 public v; }',
        '#!/usr/bin/env bash
echo "consensus lab"',
        '#!/usr/bin/env bash
grep -q "stable" /workspace/output/consensus_report.txt',
        'flag{pbft-tuned}', 'pbft-lab-2026',
        '{"mode":"checkpoint"}',
        '["核对三节点超时参数一致性。"]',
        '["s3://chainspace/challenges/0/attachments/consensus_pbft_lab.md"]',
        '{"mode":"multi_service_lab","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files","logs"]},"services":[{"key":"geth","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]},{"key":"blockscout","image":"chainspace/blockscout:latest","ports":[{"name":"web","port":4000,"protocol":"http","expose_as":"explorer"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc","explorer"],"shared_network":true},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":120}}',
        '["consensus","pbft"]', 'seed:baseline:contest:consensus_multi'
    ),
    (
        '预言机与索引联调演练', '在同一题目环境中联调 Chainlink、The Graph、Blockscout 与链节点。', 'toolchain', 'multi_service_lab', 4, 560, 320, 0.10,
        'pragma solidity ^0.8.20; contract OracleRegistry { mapping(bytes32=>address) public feeds; }',
        '#!/usr/bin/env bash
echo "toolchain lab"',
        '#!/usr/bin/env bash
grep -q "toolchain-ready" /workspace/output/toolchain_report.txt',
        'flag{oracle-indexer-toolchain-ready}', 'toolchain-lab-2026',
        '{"mode":"service_integration"}',
        '["分别验证 explorer、api_debug 与 rpc 入口是否都可工作。"]',
        '["s3://chainspace/challenges/0/attachments/toolchain_oracle_graph_lab.md"]',
        '{"mode":"multi_service_lab","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files","logs"]},"services":[{"key":"geth","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]},{"key":"blockscout","image":"chainspace/blockscout:latest","ports":[{"name":"web","port":4000,"protocol":"http","expose_as":"explorer"}]},{"key":"chainlink","image":"chainspace/chainlink:latest","ports":[{"name":"api","port":6688,"protocol":"http","expose_as":"api_debug"}]},{"key":"thegraph","image":"chainspace/thegraph:latest","ports":[{"name":"graphql","port":8000,"protocol":"http","expose_as":"api_debug"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc","explorer","api_debug"],"shared_network":true},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":140}}',
        '["oracle","graph","toolchain"]', 'seed:baseline:contest:toolchain_multi'
    ),
    (
        'ECDSA Nonce 重用取证', '离线恢复签名私钥片段。', 'crypto', 'static', 3, 420, 230, 0.10,
        '',
        '#!/usr/bin/env bash
echo "offline crypto"',
        '#!/usr/bin/env bash
python3 /workspace/scripts/verify_nonce_reuse.py',
        'flag{nonce-reuse-detected}', 'crypto-static-2026',
        '{"mode":"offline"}',
        '["观察重复 r 值。"]',
        '["s3://chainspace/challenges/0/attachments/crypto_ecdsa_nonce_reuse.md","s3://chainspace/challenges/0/attachments/ecdsa_nonce_reuse_samples.txt"]',
        '{"mode":"static","needs_environment":false,"workspace":{"image":"chainspace/eth-dev:latest"},"services":[],"topology":{"mode":"workspace_only","exposed_entries":["workspace"]},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":90}}',
        '["crypto","forensics"]', 'seed:baseline:contest:crypto_static'
    ),
    (
        '跨链消息重放防护', '修复桥接消息重放窗口缺陷。', 'cross_chain', 'multi_service_lab', 4, 520, 300, 0.10,
        'pragma solidity ^0.8.20; contract BridgeInbox { mapping(bytes32=>bool) public ok; }',
        '#!/usr/bin/env bash
echo "bridge lab"',
        '#!/usr/bin/env bash
grep -q "replay-protected" /workspace/output/bridge_check.log',
        'flag{cross-chain-replay-protected}', 'crosschain-2026',
        '{"mode":"integration"}',
        '["检查 sourceChain + nonce 联合校验。"]',
        '["s3://chainspace/challenges/0/attachments/cross_chain_replay_lab.md"]',
        '{"mode":"multi_service_lab","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files","logs"]},"services":[{"key":"chain-a","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]},{"key":"chain-b","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":9545,"protocol":"http","expose_as":"rpc"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc"],"shared_network":true},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":130}}',
        '["cross-chain","bridge"]', 'seed:baseline:contest:cross_chain_multi'
    ),
    (
        'NFT 铸造白名单绕过', '修复 Merkle 白名单编码冲突。', 'nft', 'single_chain_instance', 2, 360, 200, 0.13,
        'pragma solidity ^0.8.20; contract MintPass { mapping(address=>bool) public minted; }',
        '#!/usr/bin/env bash
forge test --match-test test_NFTWhitelist',
        '#!/usr/bin/env bash
forge test --match-test test_NoBypass',
        'flag{nft-whitelist-fixed}', 'nft-2026',
        '{"mode":"unit_test"}',
        '["验证 proof 不能跨账户复用。"]',
        '["s3://chainspace/challenges/0/attachments/nft_whitelist_bypass.md"]',
        '{"mode":"single_chain_instance","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files"]},"services":[{"key":"chain","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc"]},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":100}}',
        '["nft","merkle"]', 'seed:baseline:contest:nft_single'
    ),
    (
        '字节码级代理合约逆向', '还原代理实现槽位并给出修复方案。', 'reverse', 'static', 4, 500, 280, 0.09,
        '',
        '#!/usr/bin/env bash
echo "reverse static"',
        '#!/usr/bin/env bash
grep -q "implementation-slot" /workspace/output/reverse_report.md',
        'flag{proxy-slot-recovered}', 'reverse-2026',
        '{"mode":"manual"}',
        '["先识别 delegatecall 入口。"]',
        '["s3://chainspace/challenges/0/attachments/reverse_proxy_bytecode.md","s3://chainspace/challenges/0/attachments/reverse_proxy_runtime.txt"]',
        '{"mode":"static","needs_environment":false,"workspace":{"image":"chainspace/security:latest"},"services":[],"topology":{"mode":"workspace_only","exposed_entries":["workspace"]},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":100}}',
        '["reverse","bytecode"]', 'seed:baseline:contest:reverse_static'
    ),
    (
        '冷热钱包签名流程加固', '修复链 ID 与域分离缺失造成的重放风险。', 'key_management', 'single_chain_instance', 3, 430, 240, 0.11,
        'pragma solidity ^0.8.20; contract SafeSigner { bytes32 public domain; }',
        '#!/usr/bin/env bash
forge test --match-test test_KeyMgmt',
        '#!/usr/bin/env bash
forge test --match-test test_NoReplay',
        'flag{key-flow-hardened}', 'keymgmt-2026',
        '{"mode":"unit_test"}',
        '["检查 domain separator 与 chainId。"]',
        '["s3://chainspace/challenges/0/attachments/key_management_signing_flow.md"]',
        '{"mode":"single_chain_instance","needs_environment":true,"workspace":{"image":"chainspace/eth-dev:latest","interaction_tools":["ide","terminal","files"]},"services":[{"key":"chain","image":"chainspace/geth:latest","ports":[{"name":"rpc","port":8545,"protocol":"http","expose_as":"rpc"}]}],"topology":{"mode":"workspace_with_services","exposed_entries":["workspace","rpc"]},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":110}}',
        '["key-management","signature"]', 'seed:baseline:contest:key_management_single'
    ),
    (
        '链上监控告警误报修复', '修复阈值抖动导致的重复告警。', 'misc', 'static', 1, 260, 160, 0.15,
        '',
        '#!/usr/bin/env bash
echo "misc static"',
        '#!/usr/bin/env bash
grep -q "deduplicated-alert" /workspace/output/alert_rules.md',
        'flag{alert-noise-reduced}', 'misc-2026',
        '{"mode":"document"}',
        '["引入冷却窗口与去重规则。"]',
        '["s3://chainspace/challenges/0/attachments/misc_alert_tuning.md"]',
        '{"mode":"static","needs_environment":false,"workspace":{"image":"chainspace/eth-dev:latest"},"services":[],"topology":{"mode":"workspace_only","exposed_entries":["workspace"]},"fork":{"enabled":false},"lifecycle":{"time_limit_minutes":60}}',
        '["ops","misc"]', 'seed:baseline:contest:misc_static'
    )
) AS v(
    title, description, category, runtime_profile, difficulty, base_points, min_points, decay_factor,
    contract_code, deploy_script, check_script, flag_template, flag_secret, validation_config, hints, attachments,
    challenge_orchestration, tags, source_ref
);

-- 2) 比赛主体（解题赛 + 对抗赛）
WITH base AS (
    SELECT
        COALESCE((SELECT id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1)) AS creator_id,
        COALESCE((SELECT school_id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT school_id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1), 1) AS school_id
)
INSERT INTO contests (
    school_id, creator_id, title, description, type, level, rules, battle_orchestration,
    start_time, end_time, registration_start, registration_end, status, is_public,
    max_participants, team_min_size, team_max_size, dynamic_score, first_blood_bonus
)
SELECT school_id, creator_id,
       '链境公开解题赛（全形态覆盖）',
       '覆盖 9 分类 + 4 运行形态，含真实主网 Fork 题。',
       'jeopardy', 'school',
       '禁止破坏平台环境；按动态 Flag 判分。',
       '{}'::jsonb,
       NOW() - INTERVAL '2 days',
       NOW() + INTERVAL '5 days',
       NOW() - INTERVAL '5 days',
       NOW() + INTERVAL '1 day',
       'published', true, 300, 1, 3, true, 25
FROM base;

WITH base AS (
    SELECT
        COALESCE((SELECT id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1)) AS creator_id,
        COALESCE((SELECT school_id FROM users WHERE role = 'teacher' AND deleted_at IS NULL ORDER BY id LIMIT 1),
                 (SELECT school_id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1), 1) AS school_id
)
INSERT INTO contests (
    school_id, creator_id, title, description, type, level, rules, battle_orchestration,
    start_time, end_time, registration_start, registration_end, status, is_public,
    max_participants, team_min_size, team_max_size, dynamic_score, first_blood_bonus
)
SELECT school_id, creator_id,
       '链境智能体攻防对抗赛',
       '完整轮次推进、裁判事件、积分与回放样例。',
       'agent_battle', 'school',
       '升级窗口后锁定策略；裁判按收益+防守+效率评分。',
       '{
         "shared_chain":{"image":"chainspace/geth:latest","chain_type":"ethereum","network_id":31338,"block_time":2,"rpc_url":"http://battle-chain:8545"},
         "judge":{"image":"chainspace/judge:latest","judge_contract":"0x9F7B5a6E3d2C0A8f0f8F68B2dD8e1E2A83B4d4c1","token_contract":"0x8C7C8F7aB5d8f5C92A4B7A2a0f2f8a2A4567b321","strategy_interface":"IAgentStrategy","resource_model":"gas+capital+latency","scoring_model":"profit_weighted","score_weights":{"profit":45,"defense":25,"efficiency":20,"stability":10},"allowed_actions":["swap","arbitrage","liquidate","hedge","defend"],"forbidden_calls":["selfdestruct","delegatecall_untrusted"]},
         "team_workspace":{"image":"chainspace/eth-dev:latest","display_name":"Agent Strategy Workspace","interaction_tools":["ide","terminal","files","logs","api_debug"],"resources":{"cpu":"1200m","memory":"2Gi","storage":"12Gi"}},
         "spectate":{"enable_monitor":true,"enable_replay":true},
         "lifecycle":{"round_duration_seconds":900,"upgrade_window_seconds":300,"total_rounds":3,"auto_cleanup":true}
       }'::jsonb,
       NOW() - INTERVAL '1 day',
       NOW() + INTERVAL '2 days',
       NOW() - INTERVAL '3 days',
       NOW() + INTERVAL '12 hours',
       'published', true, 60, 2, 4, true, 0
FROM base;

INSERT INTO contest_challenges (contest_id, challenge_id, points, current_points, sort_order, is_visible)
SELECT c.id, ch.id, ch.base_points, ch.base_points, row_number() OVER (ORDER BY ch.id), true
FROM contests c
JOIN challenges ch ON ch.source_ref LIKE 'seed:baseline:contest:%'
WHERE c.title = '链境公开解题赛（全形态覆盖）';

-- 3) 队伍/报名/积分/提交/环境
WITH ids AS (
    SELECT
        (SELECT id FROM users WHERE phone = '13900000001' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS u1,
        (SELECT id FROM users WHERE phone = '13900000002' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS u2,
        (SELECT id FROM users WHERE phone = '13900000003' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS u3,
        (SELECT COALESCE(school_id, 1) FROM users WHERE phone = '13900000001' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS school_id
)
INSERT INTO teams (contest_id, name, token, leader_id, school_id, description, status, workspace_status)
SELECT c.id, v.name, v.token, v.leader_id, ids.school_id, v.description, 'active', 'running'
FROM contests c
CROSS JOIN ids
JOIN LATERAL (
    VALUES
    ('Aegis-Lab', 'TEAM-AEGIS-2026', ids.u1, '解题赛安全工程队'),
    ('Delta-Trace', 'TEAM-DELTA-2026', ids.u2, '解题赛链上取证队')
) AS v(name, token, leader_id, description) ON TRUE
WHERE c.title = '链境公开解题赛（全形态覆盖）';

WITH ids AS (
    SELECT
        (SELECT id FROM users WHERE phone = '13900000001' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS u1,
        (SELECT id FROM users WHERE phone = '13900000002' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS u2,
        (SELECT COALESCE(school_id, 1) FROM users WHERE phone = '13900000001' AND deleted_at IS NULL ORDER BY id LIMIT 1) AS school_id
)
INSERT INTO teams (contest_id, name, token, leader_id, school_id, description, status, workspace_status)
SELECT c.id, v.name, v.token, v.leader_id, ids.school_id, v.description, 'active', 'running'
FROM contests c
CROSS JOIN ids
JOIN LATERAL (
    VALUES
    ('Alpha-Quant', 'BATTLE-ALPHA-2026', ids.u1, '攻防收益导向策略'),
    ('Beta-Guard', 'BATTLE-BETA-2026', ids.u2, '防守稳态策略')
) AS v(name, token, leader_id, description) ON TRUE
WHERE c.title = '链境智能体攻防对抗赛';

INSERT INTO team_members (team_id, user_id, role, joined_at)
SELECT t.id, u.id, CASE WHEN u.id = t.leader_id THEN 'leader' ELSE 'member' END, NOW() - INTERVAL '1 day'
FROM teams t
JOIN LATERAL (
    VALUES
    ('Aegis-Lab', '13900000001'),
    ('Aegis-Lab', '13900000003'),
    ('Delta-Trace', '13900000002'),
    ('Alpha-Quant', '13900000001'),
    ('Beta-Guard', '13900000002'),
    ('Beta-Guard', '13900000003')
) AS m(team_name, user_phone) ON m.team_name = t.name
JOIN users u ON u.phone = m.user_phone
WHERE t.contest_id IN (
    SELECT id FROM contests WHERE title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛')
)
ON CONFLICT (team_id, user_id) DO NOTHING;

INSERT INTO contest_registrations (contest_id, user_id, team_id, status, registered_at)
SELECT c.id, tm.user_id, tm.team_id, 'approved', NOW() - INTERVAL '1 day'
FROM contests c
JOIN teams t ON t.contest_id = c.id
JOIN team_members tm ON tm.team_id = t.id
WHERE c.title IN ('链境公开解题赛（全形态覆盖）', '链境智能体攻防对抗赛')
ON CONFLICT (contest_id, user_id) DO NOTHING;

INSERT INTO contest_scores (contest_id, team_id, user_id, total_score, solve_count, last_solve_at, rank, first_blood_count)
SELECT c.id, t.id, t.leader_id,
       CASE WHEN t.name = 'Aegis-Lab' THEN 860 ELSE 790 END,
       CASE WHEN t.name = 'Aegis-Lab' THEN 5 ELSE 4 END,
       NOW() - INTERVAL '2 hours',
       CASE WHEN t.name = 'Aegis-Lab' THEN 1 ELSE 2 END,
       CASE WHEN t.name = 'Aegis-Lab' THEN 2 ELSE 1 END
FROM contests c
JOIN teams t ON t.contest_id = c.id
WHERE c.title = '链境公开解题赛（全形态覆盖）';

INSERT INTO contest_submissions (contest_id, challenge_id, user_id, team_id, flag, is_correct, points, submitted_at, ip)
SELECT c.id, ch.id, t.leader_id, t.id, 'flag{submitted-' || ch.id || '}', true, ch.base_points - 20, NOW() - INTERVAL '3 hours', '10.10.20.15'
FROM contests c
JOIN teams t ON t.contest_id = c.id AND t.name = 'Aegis-Lab'
JOIN challenges ch ON ch.source_ref IN (
    'seed:baseline:contest:contract_vuln',
    'seed:baseline:contest:defi_fork',
    'seed:baseline:contest:consensus_multi'
)
WHERE c.title = '链境公开解题赛（全形态覆盖）';

INSERT INTO challenge_envs (
    env_id, contest_id, challenge_id, user_id, team_id, status, pod_name, fork_pod_name,
    access_url, fork_rpc_url, flag, started_at, expires_at
)
SELECT
    'seed-env-' || c.id || '-' || ch.id || '-' || t.id,
    c.id, ch.id, t.leader_id, t.id, 'running',
    'challenge-' || ch.id || '-pod',
    CASE WHEN ch.runtime_profile = 'fork_replay' THEN 'fork-' || ch.id || '-pod' ELSE NULL END,
    '/workspace/contest/' || c.id || '/challenge/' || ch.id,
    CASE WHEN ch.runtime_profile = 'fork_replay' THEN 'http://fork-' || ch.id || ':8545' ELSE NULL END,
    'flag{runtime-ready}',
    NOW() - INTERVAL '40 minutes',
    NOW() + INTERVAL '1 hours'
FROM contests c
JOIN teams t ON t.contest_id = c.id AND t.name = 'Aegis-Lab'
JOIN challenges ch ON ch.source_ref IN (
    'seed:baseline:contest:defi_fork',
    'seed:baseline:contest:cross_chain_multi'
)
WHERE c.title = '链境公开解题赛（全形态覆盖）';

-- 4) 对抗赛轮次、合约、事件、分数
INSERT INTO agent_battle_rounds (
    contest_id, round_number, start_time, end_time, status, chain_rpc_url,
    block_height, tx_count, gas_used, winner_team_id, summary
)
SELECT c.id, r.round_number, r.start_at, r.end_at, r.status, 'http://battle-chain:8545',
       r.block_height, r.tx_count, r.gas_used, tw.id, r.summary
FROM contests c
JOIN LATERAL (
    VALUES
    (1, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '11 hours 45 minutes', 'finished', 18234501::bigint, 128, 3200456::bigint, 'Alpha-Quant', 'Alpha 套利路径领先'),
    (2, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '5 hours 45 minutes', 'finished', 18234890::bigint, 141, 3519902::bigint, 'Beta-Guard', 'Beta 防守反超'),
    (3, NOW() - INTERVAL '10 minutes', NOW() + INTERVAL '5 minutes', 'running', 18235220::bigint, 96, 2498870::bigint, 'Alpha-Quant', '第三轮进行中')
) AS r(round_number, start_at, end_at, status, block_height, tx_count, gas_used, winner_name, summary) ON TRUE
JOIN teams tw ON tw.contest_id = c.id AND tw.name = r.winner_name
WHERE c.title = '链境智能体攻防对抗赛';

INSERT INTO agent_contracts (
    contest_id, round_id, team_id, submitter_id, contract_address, source_code, byte_code, abi,
    version, status, is_active, deployed_at, deploy_tx_hash, gas_used, compile_log
)
SELECT c.id, r.id, t.id, t.leader_id,
       CASE WHEN t.name = 'Alpha-Quant' THEN '0xaA0000000000000000000000000000000000A101' ELSE '0xbB0000000000000000000000000000000000B202' END,
       CASE WHEN t.name = 'Alpha-Quant' THEN 'contract AlphaStrategy { function act() external {} }' ELSE 'contract BetaStrategy { function act() external {} }' END,
       '0x6080604052',
       '[{"inputs":[],"name":"act","outputs":[],"stateMutability":"nonpayable","type":"function"}]',
       CASE WHEN r.round_number = 3 THEN 3 ELSE r.round_number END,
       'active', true, r.start_time + INTERVAL '3 minutes',
       CASE WHEN t.name = 'Alpha-Quant' THEN '0x1f7a3dd2a12a0452b56b285f4b53f0cad9d84cf2d091e661eb742bd3f7fbca01' ELSE '0x2e4cc56fdd83a8cb5fc2b2f547a0f2f071ef612c17253f9f6ae7a14f24c1b502' END,
       CASE WHEN t.name = 'Alpha-Quant' THEN 521004 ELSE 498772 END,
       'compile success'
FROM contests c
JOIN agent_battle_rounds r ON r.contest_id = c.id
JOIN teams t ON t.contest_id = c.id
WHERE c.title = '链境智能体攻防对抗赛';

INSERT INTO agent_battle_events (
    contest_id, round_id, team_id, event_type, tx_hash, block_number, from_address, to_address,
    value, gas_used, event_data, timestamp
)
SELECT c.id, r.id, t.id, e.event_type, e.tx_hash, e.block_number, e.from_addr, e.to_addr, e.value, e.gas, e.data::jsonb, e.ts
FROM contests c
JOIN agent_battle_rounds r ON r.contest_id = c.id
JOIN teams t ON t.contest_id = c.id
JOIN LATERAL (
    VALUES
    ('strategy_action', '0x31f9fe527f8f2ff1294d7394ff946f657822640d7728aa8f534b90d3d9c6d001', r.block_height + 1, '0x1111111111111111111111111111111111111111', '0x2222222222222222222222222222222222222222', '1200', 105430::bigint, '{"action":"arbitrage","profit":"34.22"}', r.start_time + INTERVAL '4 minutes'),
    ('defense_action', '0x4f93bc81c133ccf311a4a7d11302a3d6304d63ffb5f1c4fcb3a0b61783346002', r.block_height + 2, '0x3333333333333333333333333333333333333333', '0x4444444444444444444444444444444444444444', '0', 89440::bigint, '{"action":"hedge","result":"mitigated"}', r.start_time + INTERVAL '7 minutes'),
    ('settlement', '0x59a8af90ed9ce0f7d27e5541793eb9d96f5f49862e95db77c3f0615989f7f003', r.block_height + 3, '0x5555555555555555555555555555555555555555', '0x6666666666666666666666666666666666666666', '0', 63221::bigint, '{"score_delta":18}', r.start_time + INTERVAL '10 minutes')
) AS e(event_type, tx_hash, block_number, from_addr, to_addr, value, gas, data, ts) ON TRUE
WHERE c.title = '链境智能体攻防对抗赛';

INSERT INTO agent_battle_scores (
    contest_id, team_id, round_id, score, rank, token_balance, tx_count,
    success_count, fail_count, gas_spent, profit, details
)
SELECT
    c.id, t.id, r.id,
    CASE
        WHEN t.name = 'Alpha-Quant' AND r.round_number = 1 THEN 118
        WHEN t.name = 'Beta-Guard' AND r.round_number = 1 THEN 102
        WHEN t.name = 'Alpha-Quant' AND r.round_number = 2 THEN 95
        WHEN t.name = 'Beta-Guard' AND r.round_number = 2 THEN 121
        WHEN t.name = 'Alpha-Quant' AND r.round_number = 3 THEN 110
        ELSE 99
    END,
    CASE
        WHEN r.round_number = 2 AND t.name = 'Beta-Guard' THEN 1
        WHEN r.round_number IN (1, 3) AND t.name = 'Alpha-Quant' THEN 1
        ELSE 2
    END,
    CASE WHEN t.name = 'Alpha-Quant' THEN '1006420' ELSE '1005110' END,
    CASE WHEN t.name = 'Alpha-Quant' THEN 36 ELSE 34 END,
    CASE WHEN t.name = 'Alpha-Quant' THEN 30 ELSE 28 END,
    6,
    CASE WHEN t.name = 'Alpha-Quant' THEN 890211 ELSE 845009 END,
    CASE WHEN t.name = 'Alpha-Quant' THEN '421.37' ELSE '396.12' END,
    jsonb_build_object(
        'profit', CASE WHEN t.name = 'Alpha-Quant' THEN 45 ELSE 40 END,
        'defense', CASE WHEN t.name = 'Beta-Guard' THEN 29 ELSE 24 END,
        'efficiency', CASE WHEN t.name = 'Alpha-Quant' THEN 20 ELSE 19 END,
        'stability', CASE WHEN t.name = 'Beta-Guard' THEN 10 ELSE 9 END
    )
FROM contests c
JOIN teams t ON t.contest_id = c.id
JOIN agent_battle_rounds r ON r.contest_id = c.id
WHERE c.title = '链境智能体攻防对抗赛'
ON CONFLICT (contest_id, team_id, round_id) DO NOTHING;

COMMIT;
