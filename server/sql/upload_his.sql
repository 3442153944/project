-- 文件上传记录表
CREATE TABLE IF NOT EXISTS upload_history (
                                              id BIGSERIAL PRIMARY KEY,
                                              user_id INTEGER NOT NULL,
                                              file_name VARCHAR(255) NOT NULL,
                                              original_name VARCHAR(255),
                                              file_size BIGINT,
                                              file_type VARCHAR(100),
                                              storage_path VARCHAR(500),
                                              upload_status VARCHAR(20) DEFAULT 'pending' CHECK (upload_status IN ('pending', 'uploading', 'completed', 'failed', 'cancelled')),
                                              upload_speed BIGINT,
                                              progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
                                              ip_address VARCHAR(50),
                                              user_agent VARCHAR(500),
                                              error_message TEXT,
                                              started_at TIMESTAMP,
                                              completed_at TIMESTAMP,
                                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                              updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加注释
COMMENT ON TABLE upload_history IS '文件上传记录表';
COMMENT ON COLUMN upload_history.id IS '上传记录ID';
COMMENT ON COLUMN upload_history.user_id IS '用户ID';
COMMENT ON COLUMN upload_history.file_name IS '存储文件名';
COMMENT ON COLUMN upload_history.original_name IS '原始文件名';
COMMENT ON COLUMN upload_history.file_size IS '文件大小（字节）';
COMMENT ON COLUMN upload_history.file_type IS '文件类型/MIME类型';
COMMENT ON COLUMN upload_history.storage_path IS '存储路径';
COMMENT ON COLUMN upload_history.upload_status IS '上传状态：pending/uploading/completed/failed/cancelled';
COMMENT ON COLUMN upload_history.upload_speed IS '上传速度（字节/秒）';
COMMENT ON COLUMN upload_history.progress IS '上传进度（0-100）';
COMMENT ON COLUMN upload_history.ip_address IS 'IP地址';
COMMENT ON COLUMN upload_history.user_agent IS '用户代理';
COMMENT ON COLUMN upload_history.error_message IS '错误信息';
COMMENT ON COLUMN upload_history.started_at IS '开始上传时间';
COMMENT ON COLUMN upload_history.completed_at IS '完成时间';

-- 创建索引
CREATE INDEX idx_upload_user_id ON upload_history(user_id);
CREATE INDEX idx_upload_status ON upload_history(upload_status);
CREATE INDEX idx_upload_created_at ON upload_history(created_at DESC);
CREATE INDEX idx_upload_file_name ON upload_history(file_name);

-- 创建触发器（自动更新updated_at）
CREATE OR REPLACE FUNCTION update_upload_history_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_upload_history_updated_at
    BEFORE UPDATE ON upload_history
    FOR EACH ROW
EXECUTE FUNCTION update_upload_history_updated_at();