-- ChainSpace 测试数据初始化脚本
-- 执行顺序：00-schema.sql -> 01-init.sql -> 02-seed-data.sql

-- 统一密码：ChainSpace2025
-- 登录方式：手机号 + 密码
-- "ChainSpace2025" 的 bcrypt 哈希（cost=10）
-- $2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu

-- ============================================
-- 1. 创建测试学校
-- ============================================
INSERT INTO schools (name, code, logo, address, contact, phone, email, website, description, status, created_at, updated_at)
VALUES (
    '华东理工大学',
    'ECUST',
    '/uploads/logos/ecust.png',
    '上海市徐汇区梅陇路130号',
    '教务处',
    '021-64252763',
    'jwc@ecust.edu.cn',
    'https://www.ecust.edu.cn',
    '华东理工大学信息学院开设区块链技术相关课程与实验教学。',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO schools (name, code, logo, address, contact, phone, email, website, description, status, created_at, updated_at)
VALUES (
    '浙江大学',
    'ZJU',
    '/uploads/logos/zju.png',
    '浙江省杭州市西湖区余杭塘路866号',
    '计算机学院',
    '0571-87951111',
    'cs@zju.edu.cn',
    'https://www.zju.edu.cn',
    '浙江大学计算机相关院系在区块链与系统安全方向具备较强研究和教学基础。',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- 2. 创建班级
-- ============================================
INSERT INTO classes (school_id, name, grade, major, description, status, created_at, updated_at)
SELECT s.id, '计算机 2101 班', '2021级', '计算机科学与技术', '2021级计算机科学与技术专业 1 班', 'active', NOW(), NOW()
FROM schools s
WHERE s.code = 'ECUST'
ON CONFLICT DO NOTHING;

INSERT INTO classes (school_id, name, grade, major, description, status, created_at, updated_at)
SELECT s.id, '软件工程 2101 班', '2021级', '软件工程', '2021级软件工程专业 1 班', 'active', NOW(), NOW()
FROM schools s
WHERE s.code = 'ECUST'
ON CONFLICT DO NOTHING;

-- ============================================
-- 3. 创建测试账号
-- ============================================
INSERT INTO users (password, email, phone, real_name, role, status, created_at, updated_at)
SELECT *
FROM (
    VALUES (
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    'admin@chainspace.edu.cn',
    '13800000001',
    '陈志远',
    'platform_admin',
    'active',
    NOW(),
    NOW()
)
) AS v(password, email, phone, real_name, role, status, created_at, updated_at)
WHERE NOT EXISTS (
    SELECT 1 FROM users u WHERE u.phone = '13800000001' AND u.deleted_at IS NULL
);

INSERT INTO users (school_id, password, email, phone, real_name, role, status, created_at, updated_at)
SELECT
    s.id,
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    'jwc_admin@ecust.edu.cn',
    '13800000002',
    '刘明华',
    'school_admin',
    'active',
    NOW(),
    NOW()
FROM schools s
WHERE s.code = 'ECUST'
  AND NOT EXISTS (
      SELECT 1 FROM users u WHERE u.phone = '13800000002' AND u.deleted_at IS NULL
  );

INSERT INTO users (school_id, password, email, phone, real_name, role, status, created_at, updated_at)
SELECT
    s.id,
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    'wangzq@ecust.edu.cn',
    '13812345678',
    '王志强',
    'teacher',
    'active',
    NOW(),
    NOW()
FROM schools s
WHERE s.code = 'ECUST'
  AND NOT EXISTS (
      SELECT 1 FROM users u WHERE u.phone = '13812345678' AND u.deleted_at IS NULL
  );

INSERT INTO users (school_id, password, email, phone, real_name, role, status, created_at, updated_at)
SELECT
    s.id,
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    'zhangxy@ecust.edu.cn',
    '13987654321',
    '张晓燕',
    'teacher',
    'active',
    NOW(),
    NOW()
FROM schools s
WHERE s.code = 'ECUST'
  AND NOT EXISTS (
      SELECT 1 FROM users u WHERE u.phone = '13987654321' AND u.deleted_at IS NULL
  );

INSERT INTO users (school_id, password, email, phone, real_name, role, student_no, status, created_at, updated_at)
SELECT
    s.id,
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    '20210101@mail.ecust.edu.cn',
    '13900000001',
    '赵宇轩',
    'student',
    '20210101',
    'active',
    NOW(),
    NOW()
FROM schools s
WHERE s.code = 'ECUST'
  AND NOT EXISTS (
      SELECT 1 FROM users u WHERE u.phone = '13900000001' AND u.deleted_at IS NULL
  );

INSERT INTO users (school_id, password, email, phone, real_name, role, student_no, status, created_at, updated_at)
SELECT
    s.id,
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    '20210102@mail.ecust.edu.cn',
    '13900000002',
    '林思彤',
    'student',
    '20210102',
    'active',
    NOW(),
    NOW()
FROM schools s
WHERE s.code = 'ECUST'
  AND NOT EXISTS (
      SELECT 1 FROM users u WHERE u.phone = '13900000002' AND u.deleted_at IS NULL
  );

INSERT INTO users (school_id, password, email, phone, real_name, role, student_no, status, created_at, updated_at)
SELECT
    s.id,
    '$2a$10$2KsnifRBvgyni2YU.pg2se2vKde/RA0SkI1H/KwvMVC7UIVBUhvMu',
    '20210103@mail.ecust.edu.cn',
    '13900000003',
    '孙浩然',
    'student',
    '20210103',
    'active',
    NOW(),
    NOW()
FROM schools s
WHERE s.code = 'ECUST'
  AND NOT EXISTS (
      SELECT 1 FROM users u WHERE u.phone = '13900000003' AND u.deleted_at IS NULL
  );

-- ============================================
-- 4. 初始化系统配置
-- ============================================
INSERT INTO system_configs (key, value, type, description, is_public, "group", created_at, updated_at)
VALUES
    ('platform_name', 'ChainSpace 链境', 'string', '平台名称', true, 'general', NOW(), NOW()),
    ('platform_logo', '/logo.svg', 'string', '平台 Logo', true, 'general', NOW(), NOW()),
    ('allow_registration', 'false', 'bool', '是否允许自主注册', false, 'security', NOW(), NOW()),
    ('default_experiment_timeout', '7200', 'int', '实验环境默认超时时间（秒）', false, 'experiment', NOW(), NOW()),
    ('max_experiment_extend_times', '3', 'int', '实验环境最大延期次数', false, 'experiment', NOW(), NOW())
ON CONFLICT (key) DO NOTHING;

DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '================================================';
    RAISE NOTICE '       ChainSpace 测试数据初始化完成';
    RAISE NOTICE '================================================';
    RAISE NOTICE '统一密码: ChainSpace2025';
    RAISE NOTICE '登录方式: 手机号 + 密码';
    RAISE NOTICE '平台管理员: 13800000001';
    RAISE NOTICE '学校管理员: 13800000002';
    RAISE NOTICE '教师: 13812345678, 13987654321';
    RAISE NOTICE '学生: 13900000001, 13900000002, 13900000003';
    RAISE NOTICE '================================================';
END $$;
