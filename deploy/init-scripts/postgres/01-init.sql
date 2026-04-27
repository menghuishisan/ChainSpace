-- ChainSpace PostgreSQL 初始化脚本
-- 执行时机：容器首次启动后的初始化阶段

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- 初始化默认漏洞源
-- 真实漏洞自动转题链路至少需要一个事件发现源，这里默认启用 SlowMist Hacked
INSERT INTO vulnerability_sources (
    name,
    type,
    config,
    description,
    is_active
)
VALUES (
    'SlowMist Hacked',
    'slowmist',
    '{"page_url":"https://hacked.slowmist.io/"}'::jsonb,
    '默认真实漏洞事件发现源，用于同步 SlowMist Hacked 公布的链上安全事件并进入自动增强流水线。',
    TRUE
)
ON CONFLICT DO NOTHING;

SELECT 'ChainSpace 数据库初始化完成。' AS message;
