-- ChainSpace 数据库表结构
-- 自动生成，基于 GORM 模型定义
-- 执行顺序: 00-schema.sql -> 01-init.sql -> 02-seed-data.sql

-- ============================================
-- 基础表
-- ============================================

-- 学校表
CREATE TABLE IF NOT EXISTS schools (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    logo VARCHAR(500),
    address VARCHAR(500),
    contact VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(100),
    website VARCHAR(200),
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    expire_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_schools_deleted_at ON schools(deleted_at);

-- 班级表
CREATE TABLE IF NOT EXISTS classes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    name VARCHAR(100) NOT NULL,
    grade VARCHAR(20),
    major VARCHAR(100),
    description TEXT,
    status VARCHAR(20) DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS idx_classes_deleted_at ON classes(deleted_at);
CREATE INDEX IF NOT EXISTS idx_classes_school_id ON classes(school_id);

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    school_id INTEGER REFERENCES schools(id),
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    phone VARCHAR(20),
    real_name VARCHAR(50),
    avatar VARCHAR(500),
    role VARCHAR(20) NOT NULL,
    student_no VARCHAR(50),
    class_id INTEGER REFERENCES classes(id),
    status VARCHAR(20) DEFAULT 'active',
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip VARCHAR(50),
    must_change_pwd BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_school_id ON users(school_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_student_no ON users(student_no);
CREATE INDEX IF NOT EXISTS idx_users_class_id ON users(class_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE deleted_at IS NULL AND phone IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL AND email IS NOT NULL;

-- ============================================
-- 课程相关表
-- ============================================

-- 课程表
CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    cover VARCHAR(500),
    code VARCHAR(20),
    invite_code VARCHAR(20) UNIQUE,
    category VARCHAR(50),
    tags JSONB,
    status VARCHAR(20) DEFAULT 'draft',
    is_public BOOLEAN DEFAULT FALSE,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    max_students INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_courses_deleted_at ON courses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_courses_school_id ON courses(school_id);
CREATE INDEX IF NOT EXISTS idx_courses_teacher_id ON courses(teacher_id);
CREATE INDEX IF NOT EXISTS idx_courses_code ON courses(code);

-- 课程学生关联表
CREATE TABLE IF NOT EXISTS course_students (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    course_id INTEGER NOT NULL REFERENCES courses(id),
    student_id INTEGER NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    progress FLOAT DEFAULT 0,
    last_access TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active',
    UNIQUE(course_id, student_id)
);
CREATE INDEX IF NOT EXISTS idx_course_students_deleted_at ON course_students(deleted_at);

-- 章节表
CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    course_id INTEGER NOT NULL REFERENCES courses(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft'
);
CREATE INDEX IF NOT EXISTS idx_chapters_deleted_at ON chapters(deleted_at);
CREATE INDEX IF NOT EXISTS idx_chapters_course_id ON chapters(course_id);

-- 学习资料表
CREATE TABLE IF NOT EXISTS materials (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    chapter_id INTEGER NOT NULL REFERENCES chapters(id),
    title VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    content TEXT,
    url VARCHAR(500),
    file_size BIGINT DEFAULT 0,
    duration INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS idx_materials_deleted_at ON materials(deleted_at);
CREATE INDEX IF NOT EXISTS idx_materials_chapter_id ON materials(chapter_id);

-- 学习进度表
CREATE TABLE IF NOT EXISTS material_progress (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    material_id INTEGER NOT NULL REFERENCES materials(id),
    student_id INTEGER NOT NULL REFERENCES users(id),
    progress FLOAT DEFAULT 0,
    duration INTEGER DEFAULT 0,
    last_position INTEGER DEFAULT 0,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(material_id, student_id)
);
CREATE INDEX IF NOT EXISTS idx_material_progress_deleted_at ON material_progress(deleted_at);

-- ============================================
-- 实验相关表
-- ============================================

-- 实验表
CREATE TABLE IF NOT EXISTS experiments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    chapter_id INTEGER NOT NULL REFERENCES chapters(id),
    creator_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(30) NOT NULL,
    mode VARCHAR(20) NOT NULL DEFAULT 'single',
    difficulty INTEGER DEFAULT 1,
    estimated_time INTEGER DEFAULT 60,
    max_score INTEGER DEFAULT 100,
    pass_score INTEGER DEFAULT 60,
    auto_grade BOOLEAN DEFAULT FALSE,
    grading_strategy VARCHAR(30) DEFAULT 'checkpoint',
    sort_order INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft',
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    allow_late BOOLEAN DEFAULT FALSE,
    late_deduction INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiments_deleted_at ON experiments(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiments_school_id ON experiments(school_id);
CREATE INDEX IF NOT EXISTS idx_experiments_chapter_id ON experiments(chapter_id);
CREATE INDEX IF NOT EXISTS idx_experiments_creator_id ON experiments(creator_id);
CREATE INDEX IF NOT EXISTS idx_experiments_mode ON experiments(mode);

CREATE TABLE IF NOT EXISTS experiment_workspaces (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL UNIQUE REFERENCES experiments(id),
    image VARCHAR(200) NOT NULL,
    display_name VARCHAR(100),
    cpu VARCHAR(20) NOT NULL,
    memory VARCHAR(20) NOT NULL,
    storage VARCHAR(20) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_experiment_workspaces_deleted_at ON experiment_workspaces(deleted_at);

CREATE TABLE IF NOT EXISTS experiment_workspace_tools (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    workspace_id INTEGER NOT NULL REFERENCES experiment_workspaces(id),
    tool_key VARCHAR(50) NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_workspace_tools_deleted_at ON experiment_workspace_tools(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_workspace_tools_workspace_id ON experiment_workspace_tools(workspace_id);

CREATE TABLE IF NOT EXISTS experiment_topologies (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL UNIQUE REFERENCES experiments(id),
    template VARCHAR(50) NOT NULL,
    shared_network BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_experiment_topologies_deleted_at ON experiment_topologies(deleted_at);

CREATE TABLE IF NOT EXISTS experiment_topology_exposed_entries (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    topology_id INTEGER NOT NULL REFERENCES experiment_topologies(id),
    entry_key VARCHAR(50) NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_topology_exposed_entries_deleted_at ON experiment_topology_exposed_entries(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_topology_exposed_entries_topology_id ON experiment_topology_exposed_entries(topology_id);

CREATE TABLE IF NOT EXISTS experiment_tools (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    tool_key VARCHAR(50) NOT NULL,
    label VARCHAR(100),
    kind VARCHAR(100),
    target VARCHAR(50),
    student_facing BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_tools_deleted_at ON experiment_tools(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_tools_experiment_id ON experiment_tools(experiment_id);
ALTER TABLE experiment_tools ALTER COLUMN kind TYPE VARCHAR(100);

CREATE TABLE IF NOT EXISTS experiment_init_scripts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    scope_type VARCHAR(20) NOT NULL,
    scope_key VARCHAR(50),
    script TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_init_scripts_deleted_at ON experiment_init_scripts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_init_scripts_experiment_id ON experiment_init_scripts(experiment_id);

CREATE TABLE IF NOT EXISTS experiment_collaborations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL UNIQUE REFERENCES experiments(id),
    max_members INTEGER DEFAULT 4
);
CREATE INDEX IF NOT EXISTS idx_experiment_collaborations_deleted_at ON experiment_collaborations(deleted_at);

CREATE TABLE IF NOT EXISTS experiment_role_bindings (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    collaboration_id INTEGER NOT NULL REFERENCES experiment_collaborations(id),
    role_key VARCHAR(50) NOT NULL,
    label VARCHAR(100),
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_role_bindings_deleted_at ON experiment_role_bindings(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_role_bindings_collaboration_id ON experiment_role_bindings(collaboration_id);

CREATE TABLE IF NOT EXISTS experiment_role_binding_nodes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    role_binding_id INTEGER NOT NULL REFERENCES experiment_role_bindings(id),
    node_key VARCHAR(50) NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_role_binding_nodes_deleted_at ON experiment_role_binding_nodes(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_role_binding_nodes_role_binding_id ON experiment_role_binding_nodes(role_binding_id);

CREATE TABLE IF NOT EXISTS experiment_role_binding_tools (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    role_binding_id INTEGER NOT NULL REFERENCES experiment_role_bindings(id),
    tool_key VARCHAR(50) NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_role_binding_tools_deleted_at ON experiment_role_binding_tools(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_role_binding_tools_role_binding_id ON experiment_role_binding_tools(role_binding_id);

CREATE TABLE IF NOT EXISTS experiment_nodes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    node_key VARCHAR(50) NOT NULL,
    name VARCHAR(100),
    image VARCHAR(200) NOT NULL,
    role VARCHAR(50),
    cpu VARCHAR(20) NOT NULL,
    memory VARCHAR(20) NOT NULL,
    storage VARCHAR(20) NOT NULL,
    student_facing BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_nodes_deleted_at ON experiment_nodes(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_nodes_experiment_id ON experiment_nodes(experiment_id);

CREATE TABLE IF NOT EXISTS experiment_node_ports (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    node_id INTEGER NOT NULL REFERENCES experiment_nodes(id),
    port INTEGER NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_node_ports_deleted_at ON experiment_node_ports(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_node_ports_node_id ON experiment_node_ports(node_id);

CREATE TABLE IF NOT EXISTS experiment_node_tools (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    node_id INTEGER NOT NULL REFERENCES experiment_nodes(id),
    tool_key VARCHAR(50) NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_node_tools_deleted_at ON experiment_node_tools(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_node_tools_node_id ON experiment_node_tools(node_id);

CREATE TABLE IF NOT EXISTS experiment_services (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    service_key VARCHAR(50) NOT NULL,
    name VARCHAR(100),
    image VARCHAR(200) NOT NULL,
    role VARCHAR(50),
    purpose VARCHAR(100),
    student_facing BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_services_deleted_at ON experiment_services(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_services_experiment_id ON experiment_services(experiment_id);

CREATE TABLE IF NOT EXISTS experiment_service_ports (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    service_id INTEGER NOT NULL REFERENCES experiment_services(id),
    port INTEGER NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_service_ports_deleted_at ON experiment_service_ports(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_service_ports_service_id ON experiment_service_ports(service_id);

CREATE TABLE IF NOT EXISTS experiment_service_env_vars (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    service_id INTEGER NOT NULL REFERENCES experiment_services(id),
    env_key VARCHAR(100) NOT NULL,
    env_value TEXT,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_service_env_vars_deleted_at ON experiment_service_env_vars(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_service_env_vars_service_id ON experiment_service_env_vars(service_id);

CREATE TABLE IF NOT EXISTS experiment_assets (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    asset_key VARCHAR(50) NOT NULL,
    name VARCHAR(200),
    source_type VARCHAR(30) NOT NULL,
    bucket VARCHAR(100),
    object_path VARCHAR(500),
    mount_path VARCHAR(500) NOT NULL,
    required BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_assets_deleted_at ON experiment_assets(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_assets_experiment_id ON experiment_assets(experiment_id);

CREATE TABLE IF NOT EXISTS experiment_checkpoints (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    checkpoint_key VARCHAR(50) NOT NULL,
    type VARCHAR(30) NOT NULL,
    target VARCHAR(50),
    path VARCHAR(500),
    command TEXT,
    expected TEXT,
    script TEXT,
    score INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_checkpoints_deleted_at ON experiment_checkpoints(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_checkpoints_experiment_id ON experiment_checkpoints(experiment_id);

-- 实验环境表
CREATE TABLE IF NOT EXISTS experiment_envs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    env_id VARCHAR(50) NOT NULL UNIQUE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    session_id INTEGER,
    user_id INTEGER NOT NULL REFERENCES users(id),
    school_id INTEGER NOT NULL REFERENCES schools(id),
    status VARCHAR(20) DEFAULT 'pending',
    session_mode VARCHAR(20) NOT NULL DEFAULT 'single',
    primary_instance_key VARCHAR(50),
    started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    extend_count INTEGER DEFAULT 0,
    snapshot_at TIMESTAMP WITH TIME ZONE,
    snapshot_url VARCHAR(500),
    error_message TEXT
);
CREATE INDEX IF NOT EXISTS idx_experiment_envs_deleted_at ON experiment_envs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_envs_experiment_id ON experiment_envs(experiment_id);
CREATE INDEX IF NOT EXISTS idx_experiment_envs_session_id ON experiment_envs(session_id);
CREATE INDEX IF NOT EXISTS idx_experiment_envs_user_id ON experiment_envs(user_id);
CREATE INDEX IF NOT EXISTS idx_experiment_envs_school_id ON experiment_envs(school_id);

CREATE TABLE IF NOT EXISTS experiment_sessions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    session_key VARCHAR(64) NOT NULL UNIQUE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    school_id INTEGER NOT NULL REFERENCES schools(id),
    mode VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    primary_env_id VARCHAR(50),
    max_members INTEGER DEFAULT 1,
    current_member_count INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_experiment_sessions_deleted_at ON experiment_sessions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_sessions_experiment_id ON experiment_sessions(experiment_id);
CREATE INDEX IF NOT EXISTS idx_experiment_sessions_school_id ON experiment_sessions(school_id);

ALTER TABLE experiment_envs
    DROP CONSTRAINT IF EXISTS fk_experiment_envs_session_id;

ALTER TABLE experiment_envs
    ADD CONSTRAINT fk_experiment_envs_session_id
    FOREIGN KEY (session_id) REFERENCES experiment_sessions(id);

CREATE TABLE IF NOT EXISTS experiment_session_members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    session_id INTEGER NOT NULL REFERENCES experiment_sessions(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    role_key VARCHAR(50),
    assigned_node_key VARCHAR(50),
    join_status VARCHAR(20) DEFAULT 'joined',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_experiment_session_members_deleted_at ON experiment_session_members(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_session_members_session_id ON experiment_session_members(session_id);
CREATE INDEX IF NOT EXISTS idx_experiment_session_members_user_id ON experiment_session_members(user_id);

CREATE TABLE IF NOT EXISTS experiment_session_messages (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    session_id INTEGER NOT NULL REFERENCES experiment_sessions(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text'
);
CREATE INDEX IF NOT EXISTS idx_experiment_session_messages_deleted_at ON experiment_session_messages(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_session_messages_session_id ON experiment_session_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_experiment_session_messages_user_id ON experiment_session_messages(user_id);

CREATE TABLE IF NOT EXISTS experiment_runtime_instances (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_env_id INTEGER NOT NULL REFERENCES experiment_envs(id),
    instance_key VARCHAR(50) NOT NULL,
    kind VARCHAR(20) NOT NULL,
    pod_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    student_facing BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_experiment_runtime_instances_deleted_at ON experiment_runtime_instances(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_runtime_instances_env_id ON experiment_runtime_instances(experiment_env_id);

CREATE TABLE IF NOT EXISTS experiment_runtime_tools (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    runtime_instance_id INTEGER NOT NULL REFERENCES experiment_runtime_instances(id),
    tool_key VARCHAR(50) NOT NULL,
    port INTEGER NOT NULL,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_experiment_runtime_tools_deleted_at ON experiment_runtime_tools(deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiment_runtime_tools_instance_id ON experiment_runtime_tools(runtime_instance_id);

-- 实验提交表
CREATE TABLE IF NOT EXISTS submissions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    experiment_id INTEGER NOT NULL REFERENCES experiments(id),
    student_id INTEGER NOT NULL REFERENCES users(id),
    school_id INTEGER NOT NULL REFERENCES schools(id),
    env_id VARCHAR(50),
    content TEXT,
    file_url VARCHAR(500),
    snapshot_url VARCHAR(500),
    score INTEGER,
    auto_score INTEGER,
    manual_score INTEGER,
    feedback TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    graded_at TIMESTAMP WITH TIME ZONE,
    grader_id INTEGER REFERENCES users(id),
    is_late BOOLEAN DEFAULT FALSE,
    attempt_number INTEGER DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_submissions_deleted_at ON submissions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_submissions_experiment_id ON submissions(experiment_id);
CREATE INDEX IF NOT EXISTS idx_submissions_student_id ON submissions(student_id);
CREATE INDEX IF NOT EXISTS idx_submissions_school_id ON submissions(school_id);

CREATE TABLE IF NOT EXISTS submission_check_results (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    submission_id INTEGER NOT NULL REFERENCES submissions(id),
    checkpoint_key VARCHAR(50) NOT NULL,
    checkpoint_type VARCHAR(30) NOT NULL,
    target VARCHAR(50),
    passed BOOLEAN DEFAULT FALSE,
    score INTEGER DEFAULT 0,
    details TEXT,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_submission_check_results_deleted_at ON submission_check_results(deleted_at);
CREATE INDEX IF NOT EXISTS idx_submission_check_results_submission_id ON submission_check_results(submission_id);

-- Docker镜像表
CREATE TABLE IF NOT EXISTS docker_images (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(100) NOT NULL UNIQUE,
    tag VARCHAR(50) NOT NULL,
    registry VARCHAR(200),
    description TEXT,
    category VARCHAR(50),
    default_resources JSONB,
    features JSONB,
    env_vars JSONB,
    ports JSONB,
    base_image VARCHAR(100),
    size BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    is_built_in BOOLEAN DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_docker_images_deleted_at ON docker_images(deleted_at);

-- ============================================
-- 竞赛相关表
-- ============================================

-- 竞赛表
CREATE TABLE IF NOT EXISTS contests (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    school_id INTEGER REFERENCES schools(id),
    creator_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL,
    level VARCHAR(20) DEFAULT 'practice',
    cover VARCHAR(500),
    rules TEXT,
    battle_orchestration JSONB,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    registration_start TIMESTAMP WITH TIME ZONE,
    registration_end TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'draft',
    is_public BOOLEAN DEFAULT FALSE,
    max_participants INTEGER DEFAULT 0,
    team_min_size INTEGER DEFAULT 1,
    team_max_size INTEGER DEFAULT 1,
    dynamic_score BOOLEAN DEFAULT TRUE,
    first_blood_bonus INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_contests_deleted_at ON contests(deleted_at);
CREATE INDEX IF NOT EXISTS idx_contests_school_id ON contests(school_id);
CREATE INDEX IF NOT EXISTS idx_contests_creator_id ON contests(creator_id);

-- 挑战题目表
CREATE TABLE IF NOT EXISTS challenges (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    school_id INTEGER REFERENCES schools(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    runtime_profile VARCHAR(40) DEFAULT 'single_chain_instance',
    difficulty INTEGER DEFAULT 1,
    base_points INTEGER DEFAULT 100,
    min_points INTEGER DEFAULT 50,
    decay_factor FLOAT DEFAULT 0.1,
    contract_code TEXT,
    setup_code TEXT,
    deploy_script TEXT,
    check_script TEXT,
    flag_template VARCHAR(200),
    flag_type VARCHAR(20) DEFAULT 'static',
    flag_regex VARCHAR(500),
    flag_secret VARCHAR(200),
    validation_config JSONB,
    hints JSONB,
    attachments JSONB,
    challenge_orchestration JSONB,
    tags JSONB,
    source_type VARCHAR(20) DEFAULT 'user_created',
    source_ref VARCHAR(255),
    status VARCHAR(20) DEFAULT 'draft',
    is_public BOOLEAN DEFAULT FALSE,
    solve_count INTEGER DEFAULT 0,
    attempt_count INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_challenges_deleted_at ON challenges(deleted_at);
CREATE INDEX IF NOT EXISTS idx_challenges_creator_id ON challenges(creator_id);
CREATE INDEX IF NOT EXISTS idx_challenges_school_id ON challenges(school_id);

-- 竞赛题目关联表
CREATE TABLE IF NOT EXISTS contest_challenges (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    challenge_id INTEGER NOT NULL REFERENCES challenges(id),
    points INTEGER DEFAULT 100,
    current_points INTEGER DEFAULT 100,
    sort_order INTEGER DEFAULT 0,
    is_visible BOOLEAN DEFAULT TRUE,
    UNIQUE(contest_id, challenge_id)
);
CREATE INDEX IF NOT EXISTS idx_contest_challenges_deleted_at ON contest_challenges(deleted_at);

-- 队伍表
CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    name VARCHAR(100) NOT NULL,
    token VARCHAR(50) UNIQUE,
    leader_id INTEGER NOT NULL REFERENCES users(id),
    school_id INTEGER REFERENCES schools(id),
    avatar VARCHAR(500),
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    workspace_pod_name VARCHAR(100),
    workspace_status VARCHAR(20) DEFAULT 'inactive'
);
CREATE INDEX IF NOT EXISTS idx_teams_deleted_at ON teams(deleted_at);
CREATE INDEX IF NOT EXISTS idx_teams_contest_id ON teams(contest_id);
CREATE INDEX IF NOT EXISTS idx_teams_leader_id ON teams(leader_id);
CREATE INDEX IF NOT EXISTS idx_teams_school_id ON teams(school_id);

-- 队伍成员表
CREATE TABLE IF NOT EXISTS team_members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    team_id INTEGER NOT NULL REFERENCES teams(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_team_members_deleted_at ON team_members(deleted_at);

-- 竞赛报名表
CREATE TABLE IF NOT EXISTS contest_registrations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    team_id INTEGER REFERENCES teams(id),
    status VARCHAR(20) DEFAULT 'pending',
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(contest_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_contest_registrations_deleted_at ON contest_registrations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_contest_registrations_team_id ON contest_registrations(team_id);

-- 题目环境表
CREATE TABLE IF NOT EXISTS challenge_envs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    env_id VARCHAR(50) NOT NULL UNIQUE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    challenge_id INTEGER NOT NULL REFERENCES challenges(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    team_id INTEGER REFERENCES teams(id),
    status VARCHAR(20) DEFAULT 'pending',
    pod_name VARCHAR(100),
    fork_pod_name VARCHAR(100),
    access_url VARCHAR(500),
    fork_rpc_url VARCHAR(500),
    flag VARCHAR(200),
    started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);
CREATE INDEX IF NOT EXISTS idx_challenge_envs_deleted_at ON challenge_envs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_challenge_envs_contest_id ON challenge_envs(contest_id);
CREATE INDEX IF NOT EXISTS idx_challenge_envs_challenge_id ON challenge_envs(challenge_id);
CREATE INDEX IF NOT EXISTS idx_challenge_envs_user_id ON challenge_envs(user_id);
CREATE INDEX IF NOT EXISTS idx_challenge_envs_team_id ON challenge_envs(team_id);

-- 竞赛提交表
CREATE TABLE IF NOT EXISTS contest_submissions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    challenge_id INTEGER NOT NULL REFERENCES challenges(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    team_id INTEGER REFERENCES teams(id),
    flag VARCHAR(200),
    is_correct BOOLEAN DEFAULT FALSE,
    points INTEGER DEFAULT 0,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip VARCHAR(50)
);
CREATE INDEX IF NOT EXISTS idx_contest_submissions_deleted_at ON contest_submissions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_contest_submissions_contest_id ON contest_submissions(contest_id);
CREATE INDEX IF NOT EXISTS idx_contest_submissions_challenge_id ON contest_submissions(challenge_id);
CREATE INDEX IF NOT EXISTS idx_contest_submissions_user_id ON contest_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_submissions_team_id ON contest_submissions(team_id);

-- 竞赛分数表
CREATE TABLE IF NOT EXISTS contest_scores (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    team_id INTEGER REFERENCES teams(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    total_score INTEGER DEFAULT 0,
    solve_count INTEGER DEFAULT 0,
    last_solve_at TIMESTAMP WITH TIME ZONE,
    rank INTEGER DEFAULT 0,
    first_blood_count INTEGER DEFAULT 0,
    UNIQUE(contest_id, team_id)
);
CREATE INDEX IF NOT EXISTS idx_contest_scores_deleted_at ON contest_scores(deleted_at);
CREATE INDEX IF NOT EXISTS idx_contest_scores_user_id ON contest_scores(user_id);

-- ============================================
-- 智能体对抗相关表
-- ============================================

-- 对抗赛轮次表
CREATE TABLE IF NOT EXISTS agent_battle_rounds (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    round_number INTEGER NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'pending',
    chain_rpc_url VARCHAR(200),
    block_height BIGINT DEFAULT 0,
    tx_count INTEGER DEFAULT 0,
    gas_used BIGINT DEFAULT 0,
    winner_team_id INTEGER REFERENCES teams(id),
    summary TEXT
);
CREATE INDEX IF NOT EXISTS idx_agent_battle_rounds_deleted_at ON agent_battle_rounds(deleted_at);
CREATE INDEX IF NOT EXISTS idx_agent_battle_rounds_contest_id ON agent_battle_rounds(contest_id);

-- 智能体合约表
CREATE TABLE IF NOT EXISTS agent_contracts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    round_id INTEGER NOT NULL REFERENCES agent_battle_rounds(id),
    team_id INTEGER NOT NULL REFERENCES teams(id),
    submitter_id INTEGER NOT NULL REFERENCES users(id),
    contract_address VARCHAR(50),
    source_code TEXT,
    byte_code TEXT,
    abi TEXT,
    version INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'pending',
    is_active BOOLEAN DEFAULT FALSE,
    deployed_at TIMESTAMP WITH TIME ZONE,
    deploy_tx_hash VARCHAR(100),
    gas_used BIGINT DEFAULT 0,
    error_message TEXT,
    compile_log TEXT
);
CREATE INDEX IF NOT EXISTS idx_agent_contracts_deleted_at ON agent_contracts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_agent_contracts_contest_id ON agent_contracts(contest_id);
CREATE INDEX IF NOT EXISTS idx_agent_contracts_team_id ON agent_contracts(team_id);
CREATE INDEX IF NOT EXISTS idx_agent_contracts_submitter_id ON agent_contracts(submitter_id);

-- 对抗赛事件表
CREATE TABLE IF NOT EXISTS agent_battle_events (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    round_id INTEGER NOT NULL REFERENCES agent_battle_rounds(id),
    team_id INTEGER REFERENCES teams(id),
    event_type VARCHAR(50) NOT NULL,
    tx_hash VARCHAR(100),
    block_number BIGINT DEFAULT 0,
    from_address VARCHAR(50),
    to_address VARCHAR(50),
    value VARCHAR(100),
    gas_used BIGINT DEFAULT 0,
    event_data JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_agent_battle_events_deleted_at ON agent_battle_events(deleted_at);
CREATE INDEX IF NOT EXISTS idx_agent_battle_events_contest_id ON agent_battle_events(contest_id);
CREATE INDEX IF NOT EXISTS idx_agent_battle_events_round_id ON agent_battle_events(round_id);
CREATE INDEX IF NOT EXISTS idx_agent_battle_events_team_id ON agent_battle_events(team_id);

-- 对抗赛分数表
CREATE TABLE IF NOT EXISTS agent_battle_scores (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    contest_id INTEGER NOT NULL REFERENCES contests(id),
    team_id INTEGER NOT NULL REFERENCES teams(id),
    round_id INTEGER NOT NULL REFERENCES agent_battle_rounds(id),
    score INTEGER DEFAULT 0,
    rank INTEGER DEFAULT 0,
    token_balance VARCHAR(100) DEFAULT '0',
    tx_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    fail_count INTEGER DEFAULT 0,
    gas_spent BIGINT DEFAULT 0,
    profit VARCHAR(100) DEFAULT '0',
    details JSONB,
    UNIQUE(contest_id, team_id, round_id)
);
CREATE INDEX IF NOT EXISTS idx_agent_battle_scores_deleted_at ON agent_battle_scores(deleted_at);

-- 对抗赛配置表

-- ============================================
-- 通知相关表
-- ============================================

-- 通知表
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    school_id INTEGER REFERENCES schools(id),
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    link VARCHAR(500),
    related_id INTEGER,
    related_type VARCHAR(50),
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    sender_id INTEGER REFERENCES users(id),
    extra JSONB
);
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications(deleted_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_school_id ON notifications(school_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);

-- 通知模板表
CREATE TABLE IF NOT EXISTS notification_templates (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(50) NOT NULL UNIQUE,
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    description TEXT,
    variables JSONB,
    status VARCHAR(20) DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS idx_notification_templates_deleted_at ON notification_templates(deleted_at);

-- ============================================
-- 讨论相关表
-- ============================================

-- 帖子表
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    author_id INTEGER NOT NULL REFERENCES users(id),
    course_id INTEGER REFERENCES courses(id),
    experiment_id INTEGER REFERENCES experiments(id),
    contest_id INTEGER REFERENCES contests(id),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    tags JSONB,
    view_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active',
    last_reply_at TIMESTAMP WITH TIME ZONE,
    last_reply_by INTEGER REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_posts_school_id ON posts(school_id);
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_course_id ON posts(course_id);
CREATE INDEX IF NOT EXISTS idx_posts_experiment_id ON posts(experiment_id);
CREATE INDEX IF NOT EXISTS idx_posts_contest_id ON posts(contest_id);

-- 回复表
CREATE TABLE IF NOT EXISTS replies (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    post_id INTEGER NOT NULL REFERENCES posts(id),
    author_id INTEGER NOT NULL REFERENCES users(id),
    parent_id INTEGER REFERENCES replies(id),
    content TEXT NOT NULL,
    like_count INTEGER DEFAULT 0,
    is_accepted BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS idx_replies_deleted_at ON replies(deleted_at);
CREATE INDEX IF NOT EXISTS idx_replies_post_id ON replies(post_id);
CREATE INDEX IF NOT EXISTS idx_replies_author_id ON replies(author_id);
CREATE INDEX IF NOT EXISTS idx_replies_parent_id ON replies(parent_id);

-- 帖子点赞表
CREATE TABLE IF NOT EXISTS post_likes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    post_id INTEGER NOT NULL REFERENCES posts(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    UNIQUE(post_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_post_likes_deleted_at ON post_likes(deleted_at);

-- 回复点赞表
CREATE TABLE IF NOT EXISTS reply_likes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    reply_id INTEGER NOT NULL REFERENCES replies(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    UNIQUE(reply_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_reply_likes_deleted_at ON reply_likes(deleted_at);

-- ============================================
-- 系统相关表
-- ============================================

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT,
    type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    "group" VARCHAR(50) DEFAULT 'general'
);
CREATE INDEX IF NOT EXISTS idx_system_configs_deleted_at ON system_configs(deleted_at);

-- 漏洞数据源表
CREATE TABLE IF NOT EXISTS vulnerability_sources (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSONB,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_sync_error TEXT
);
CREATE INDEX IF NOT EXISTS idx_vulnerability_sources_deleted_at ON vulnerability_sources(deleted_at);

-- 漏洞数据表
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    source_id INTEGER NOT NULL REFERENCES vulnerability_sources(id),
    external_id VARCHAR(100),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    severity VARCHAR(20),
    category VARCHAR(50),
    chain VARCHAR(50),
    amount FLOAT,
    attack_date TIMESTAMP WITH TIME ZONE,
    block_number BIGINT DEFAULT 0,
    contract_address VARCHAR(100) DEFAULT '',
    attack_tx_hash VARCHAR(100) DEFAULT '',
    fork_block_number BIGINT DEFAULT 0,
    related_contracts JSONB,
    related_tokens JSONB,
    attacker_addresses JSONB,
    victim_addresses JSONB,
    evidence_links JSONB,
    technique VARCHAR(200) DEFAULT '',
    reference VARCHAR(500),
    vuln_code TEXT,
    attack_code TEXT,
    analysis TEXT,
    tags JSONB,
    source_snapshot JSONB,
    metadata JSONB,
    enrich_status VARCHAR(20) DEFAULT 'pending',
    conversion_score INTEGER DEFAULT 0,
    conversion_notes TEXT,
    runtime_profile_suggestion VARCHAR(40) DEFAULT 'fork_replay',
    status VARCHAR(20) DEFAULT 'discovered',
    converted_id INTEGER REFERENCES challenges(id)
);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_deleted_at ON vulnerabilities(deleted_at);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_source_id ON vulnerabilities(source_id);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_external_id ON vulnerabilities(external_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vulnerabilities_source_external_id
ON vulnerabilities(source_id, external_id)
WHERE deleted_at IS NULL AND external_id IS NOT NULL;

-- 跨校申请表
CREATE TABLE IF NOT EXISTS cross_school_applications (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    from_school_id INTEGER NOT NULL REFERENCES schools(id),
    to_school_id INTEGER NOT NULL REFERENCES schools(id),
    applicant_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(30) NOT NULL,
    target_id INTEGER NOT NULL,
    target_type VARCHAR(30) NOT NULL,
    reason TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    reviewer_id INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reject_reason TEXT
);
CREATE INDEX IF NOT EXISTS idx_cross_school_applications_deleted_at ON cross_school_applications(deleted_at);
CREATE INDEX IF NOT EXISTS idx_cross_school_applications_from_school_id ON cross_school_applications(from_school_id);
CREATE INDEX IF NOT EXISTS idx_cross_school_applications_to_school_id ON cross_school_applications(to_school_id);
CREATE INDEX IF NOT EXISTS idx_cross_school_applications_applicant_id ON cross_school_applications(applicant_id);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    school_id INTEGER REFERENCES schools(id),
    module VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    target_type VARCHAR(50),
    target_id INTEGER,
    description TEXT,
    request_ip VARCHAR(50),
    user_agent VARCHAR(500),
    request_data JSONB,
    response_code INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_operation_logs_deleted_at ON operation_logs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_school_id ON operation_logs(school_id);

-- 题目公开申请表
CREATE TABLE IF NOT EXISTS challenge_publish_requests (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    challenge_id INTEGER NOT NULL REFERENCES challenges(id),
    applicant_id INTEGER NOT NULL REFERENCES users(id),
    reason TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    reviewer_id INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_comment TEXT
);
CREATE INDEX IF NOT EXISTS idx_challenge_publish_requests_deleted_at ON challenge_publish_requests(deleted_at);
CREATE INDEX IF NOT EXISTS idx_challenge_publish_requests_challenge_id ON challenge_publish_requests(challenge_id);
CREATE INDEX IF NOT EXISTS idx_challenge_publish_requests_applicant_id ON challenge_publish_requests(applicant_id);

-- ============================================
-- 异步任务表
-- ============================================
CREATE TABLE IF NOT EXISTS async_tasks (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    payload JSONB,
    status VARCHAR(20) DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    last_error TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_async_tasks_type ON async_tasks(type);
CREATE INDEX IF NOT EXISTS idx_async_tasks_status ON async_tasks(status);
CREATE INDEX IF NOT EXISTS idx_async_tasks_next_retry_at ON async_tasks(next_retry_at);

-- ============================================
-- 完成提示
-- ============================================
SELECT 'ChainSpace database schema created successfully! Total: 32 tables.' AS message;
