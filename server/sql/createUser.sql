-- 用户表
CREATE TABLE "user"
(
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(50) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    email      VARCHAR(100),
    phone      VARCHAR(20),
    avatar     VARCHAR(255),
    role       VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    status     SMALLINT DEFAULT 1 CHECK (status IN (0, 1)),
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 约束命名（可选）
    CONSTRAINT uk_user_email UNIQUE (email),
    CONSTRAINT chk_user_username_length CHECK (length(username) >= 3)
);

-- 添加注释
COMMENT ON TABLE "user" IS '用户表';
COMMENT ON COLUMN "user".id IS '用户ID';
COMMENT ON COLUMN "user".username IS '用户名';
COMMENT ON COLUMN "user".password IS '密码（bcrypt加密）';
COMMENT ON COLUMN "user".role IS '角色：admin/user';
COMMENT ON COLUMN "user".status IS '状态：1正常/0禁用';

-- 创建索引
CREATE INDEX idx_user_username ON "user"(username);
CREATE INDEX idx_user_email ON "user"(email) WHERE email IS NOT NULL;
CREATE INDEX idx_user_status ON "user"(status);

-- 创建触发器（自动更新updated_at）
CREATE OR REPLACE FUNCTION update_user_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_updated_at
    BEFORE UPDATE ON "user"
    FOR EACH ROW
EXECUTE FUNCTION update_user_updated_at();