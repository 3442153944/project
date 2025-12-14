-- -- ============================================
-- -- 1. 用户相关表（补充约束和索引）
-- -- ============================================
--
-- -- 为用户表添加唯一约束（如果不存在）
-- do $$
--     begin
--         if not exists (select 1 from pg_constraint where conname = 'user_username_unique') then
--             alter table "user" add constraint user_username_unique unique (username);
--         end if;
--
--         if not exists (select 1 from pg_constraint where conname = 'user_email_unique') then
--             alter table "user" add constraint user_email_unique unique (email);
--         end if;
--     end $$;
--
-- -- 创建索引
-- create index if not exists idx_user_status on "user"(status);
-- create index if not exists idx_user_role on "user"(role);
-- create index if not exists idx_user_created_at on "user"(created_at);


-- ============================================
-- 2. 设备管理表
-- ============================================

create table if not exists device (
                                      id serial primary key,
                                      user_id integer not null,
                                      device_name varchar(100) not null,
                                      device_type varchar(20) not null,
                                      device_id varchar(100) not null,
                                      os_version varchar(50),
                                      app_version varchar(50),
                                      ip_address varchar(50),
                                      last_active timestamp,
                                      status smallint default 1,
                                      created_at timestamp default CURRENT_TIMESTAMP,
                                      updated_at timestamp default CURRENT_TIMESTAMP,
                                      constraint fk_device_user foreign key (user_id) references "user"(id) on delete cascade
);

comment on table device is '用户设备表';
comment on column device.id is '设备ID';
comment on column device.user_id is '用户ID';
comment on column device.device_name is '设备名称';
comment on column device.device_type is '设备类型：mobile/web/windows/mac/linux';
comment on column device.device_id is '设备唯一标识';
comment on column device.os_version is '操作系统版本';
comment on column device.app_version is '应用版本';
comment on column device.ip_address is 'IP地址';
comment on column device.last_active is '最后活跃时间';
comment on column device.status is '状态：1正常/0禁用';

create unique index if not exists idx_device_unique on device(user_id, device_id);
create index if not exists idx_device_user on device(user_id);
create index if not exists idx_device_type on device(device_type);
create index if not exists idx_device_status on device(status);


-- ============================================
-- 3. 文件管理表
-- ============================================

create table if not exists file (
                                    id bigserial primary key,
                                    user_id integer not null,
                                    parent_id bigint,
                                    file_name varchar(255) not null,
                                    file_path varchar(1000) not null,
                                    file_type varchar(20),
                                    file_size bigint,
                                    file_hash varchar(64),
                                    mime_type varchar(100),
                                    is_directory boolean default false,
                                    is_deleted boolean default false,
                                    version integer default 1,
                                    share_code varchar(32),
                                    share_expire timestamp,
                                    deleted_at timestamp,
                                    created_at timestamp default CURRENT_TIMESTAMP,
                                    updated_at timestamp default CURRENT_TIMESTAMP,
                                    constraint fk_file_user foreign key (user_id) references "user"(id) on delete cascade,
                                    constraint fk_file_parent foreign key (parent_id) references file(id) on delete cascade
);

comment on table file is '文件表';
comment on column file.id is '文件ID';
comment on column file.user_id is '用户ID';
comment on column file.parent_id is '父目录ID';
comment on column file.file_name is '文件名';
comment on column file.file_path is '文件完整路径';
comment on column file.file_type is '文件类型：doc/image/video/audio/other';
comment on column file.file_size is '文件大小（字节）';
comment on column file.file_hash is '文件哈希值（SHA256）';
comment on column file.mime_type is 'MIME类型';
comment on column file.is_directory is '是否为目录';
comment on column file.is_deleted is '是否已删除';
comment on column file.version is '文件版本号';
comment on column file.share_code is '分享码';
comment on column file.share_expire is '分享过期时间';
comment on column file.deleted_at is '删除时间';

create index if not exists idx_file_user on file(user_id);
create index if not exists idx_file_parent on file(parent_id);
create index if not exists idx_file_hash on file(file_hash);
create index if not exists idx_file_deleted on file(is_deleted);
create index if not exists idx_file_share on file(share_code) where share_code is not null;
create index if not exists idx_file_path on file using hash(file_path);
create index if not exists idx_file_user_path on file(user_id, file_path);


-- 文件版本历史表
create table if not exists file_version (
                                            id bigserial primary key,
                                            file_id bigint not null,
                                            version integer not null,
                                            file_size bigint,
                                            file_hash varchar(64),
                                            storage_path varchar(1000),
                                            created_by integer,
                                            created_at timestamp default CURRENT_TIMESTAMP,
                                            constraint fk_version_file foreign key (file_id) references file(id) on delete cascade,
                                            constraint fk_version_user foreign key (created_by) references "user"(id) on delete set null
);

comment on table file_version is '文件版本历史表';
comment on column file_version.id is '版本ID';
comment on column file_version.file_id is '文件ID';
comment on column file_version.version is '版本号';
comment on column file_version.file_hash is '文件哈希值';
comment on column file_version.storage_path is '存储路径';
comment on column file_version.created_by is '创建人';

create index if not exists idx_version_file on file_version(file_id);
create unique index if not exists idx_version_unique on file_version(file_id, version);


-- ============================================
-- 4. 同步管理表
-- ============================================

create table if not exists sync_task (
                                         id bigserial primary key,
                                         user_id integer not null,
                                         device_id integer not null,
                                         file_id bigint not null,
                                         task_type varchar(20) not null,
                                         sync_status varchar(20) default 'pending',
                                         progress integer default 0,
                                         error_message text,
                                         started_at timestamp,
                                         completed_at timestamp,
                                         created_at timestamp default CURRENT_TIMESTAMP,
                                         constraint fk_sync_user foreign key (user_id) references "user"(id) on delete cascade,
                                         constraint fk_sync_device foreign key (device_id) references device(id) on delete cascade,
                                         constraint fk_sync_file foreign key (file_id) references file(id) on delete cascade
);

comment on table sync_task is '同步任务表';
comment on column sync_task.id is '任务ID';
comment on column sync_task.user_id is '用户ID';
comment on column sync_task.device_id is '设备ID';
comment on column sync_task.file_id is '文件ID';
comment on column sync_task.task_type is '任务类型：upload/download/delete';
comment on column sync_task.sync_status is '同步状态：pending/syncing/completed/failed';
comment on column sync_task.progress is '进度（0-100）';
comment on column sync_task.error_message is '错误信息';

create index if not exists idx_sync_user on sync_task(user_id);
create index if not exists idx_sync_device on sync_task(device_id);
create index if not exists idx_sync_status on sync_task(sync_status);
create index if not exists idx_sync_created on sync_task(created_at);
create index if not exists idx_sync_user_status on sync_task(user_id, sync_status);


-- ============================================
-- 5. 下载历史表
-- ============================================

create table if not exists download_history (
                                                id bigserial primary key,
                                                user_id integer not null,
                                                device_id integer not null,
                                                file_id bigint,
                                                file_name varchar(255),
                                                file_size bigint,
                                                download_status varchar(20) default 'pending',
                                                download_speed bigint,
                                                ip_address varchar(50),
                                                started_at timestamp,
                                                completed_at timestamp,
                                                created_at timestamp default CURRENT_TIMESTAMP,
                                                constraint fk_download_user foreign key (user_id) references "user"(id) on delete cascade,
                                                constraint fk_download_device foreign key (device_id) references device(id) on delete cascade,
                                                constraint fk_download_file foreign key (file_id) references file(id) on delete set null
);

comment on table download_history is '下载历史记录表';
comment on column download_history.id is '记录ID';
comment on column download_history.user_id is '用户ID';
comment on column download_history.device_id is '设备ID';
comment on column download_history.file_id is '文件ID';
comment on column download_history.file_name is '文件名（快照）';
comment on column download_history.file_size is '文件大小';
comment on column download_history.download_status is '下载状态：pending/downloading/completed/failed/cancelled';
comment on column download_history.download_speed is '下载速度（字节/秒）';

create index if not exists idx_download_user on download_history(user_id);
create index if not exists idx_download_device on download_history(device_id);
create index if not exists idx_download_created on download_history(created_at);
create index if not exists idx_download_user_created on download_history(user_id, created_at desc);


-- ============================================
-- 6. 权限管理表
-- ============================================

create table if not exists permission (
                                          id serial primary key,
                                          permission_code varchar(50) not null unique,
                                          permission_name varchar(100) not null,
                                          parent_id integer,
                                          permission_type varchar(20),
                                          description text,
                                          sort_order integer default 0,
                                          status smallint default 1,
                                          created_at timestamp default CURRENT_TIMESTAMP,
                                          constraint fk_permission_parent foreign key (parent_id) references permission(id) on delete cascade
);

comment on table permission is '权限表';
comment on column permission.id is '权限ID';
comment on column permission.permission_code is '权限代码';
comment on column permission.permission_name is '权限名称';
comment on column permission.parent_id is '父权限ID';
comment on column permission.permission_type is '权限类型：menu/button/api';
comment on column permission.description is '权限描述';
comment on column permission.sort_order is '排序';
comment on column permission.status is '状态：1启用/0禁用';

create index if not exists idx_permission_parent on permission(parent_id);
create index if not exists idx_permission_code on permission(permission_code);


create table if not exists role (
                                    id serial primary key,
                                    role_code varchar(50) not null unique,
                                    role_name varchar(100) not null,
                                    description text,
                                    status smallint default 1,
                                    created_at timestamp default CURRENT_TIMESTAMP,
                                    updated_at timestamp default CURRENT_TIMESTAMP
);

comment on table role is '角色表';
comment on column role.id is '角色ID';
comment on column role.role_code is '角色代码';
comment on column role.role_name is '角色名称';
comment on column role.description is '角色描述';
comment on column role.status is '状态：1启用/0禁用';

create index if not exists idx_role_code on role(role_code);


create table if not exists role_permission (
                                               id serial primary key,
                                               role_id integer not null,
                                               permission_id integer not null,
                                               created_at timestamp default CURRENT_TIMESTAMP,
                                               constraint fk_role_perm_role foreign key (role_id) references role(id) on delete cascade,
                                               constraint fk_role_perm_permission foreign key (permission_id) references permission(id) on delete cascade,
                                               constraint uk_role_permission unique (role_id, permission_id)
);

comment on table role_permission is '角色权限关联表';

create index if not exists idx_role_perm_role on role_permission(role_id);
create index if not exists idx_role_perm_permission on role_permission(permission_id);


create table if not exists user_role (
                                         id serial primary key,
                                         user_id integer not null,
                                         role_id integer not null,
                                         created_at timestamp default CURRENT_TIMESTAMP,
                                         constraint fk_user_role_user foreign key (user_id) references "user"(id) on delete cascade,
                                         constraint fk_user_role_role foreign key (role_id) references role(id) on delete cascade,
                                         constraint uk_user_role unique (user_id, role_id)
);

comment on table user_role is '用户角色关联表';

create index if not exists idx_user_role_user on user_role(user_id);
create index if not exists idx_user_role_role on user_role(role_id);


-- ============================================
-- 7. 字典管理表
-- ============================================

create table if not exists dict_type (
                                         id serial primary key,
                                         dict_code varchar(50) not null unique,
                                         dict_name varchar(100) not null,
                                         description text,
                                         status smallint default 1,
                                         created_at timestamp default CURRENT_TIMESTAMP,
                                         updated_at timestamp default CURRENT_TIMESTAMP
);

comment on table dict_type is '字典类型表';
comment on column dict_type.id is '字典类型ID';
comment on column dict_type.dict_code is '字典代码';
comment on column dict_type.dict_name is '字典名称';
comment on column dict_type.description is '描述';
comment on column dict_type.status is '状态：1启用/0禁用';

create index if not exists idx_dict_type_code on dict_type(dict_code);


create table if not exists dict_data (
                                         id serial primary key,
                                         dict_type_id integer not null,
                                         dict_label varchar(100) not null,
                                         dict_value varchar(100) not null,
                                         dict_sort integer default 0,
                                         css_class varchar(50),
                                         tag_type varchar(20),
                                         remark text,
                                         status smallint default 1,
                                         created_at timestamp default CURRENT_TIMESTAMP,
                                         updated_at timestamp default CURRENT_TIMESTAMP,
                                         constraint fk_dict_data_type foreign key (dict_type_id) references dict_type(id) on delete cascade
);

comment on table dict_data is '字典数据表';
comment on column dict_data.id is '字典数据ID';
comment on column dict_data.dict_type_id is '字典类型ID';
comment on column dict_data.dict_label is '字典标签';
comment on column dict_data.dict_value is '字典值';
comment on column dict_data.dict_sort is '排序';
comment on column dict_data.css_class is '样式类';
comment on column dict_data.tag_type is '标签类型';
comment on column dict_data.remark is '备注';
comment on column dict_data.status is '状态：1启用/0禁用';

create index if not exists idx_dict_data_type on dict_data(dict_type_id);
create index if not exists idx_dict_data_value on dict_data(dict_value);
create index if not exists idx_dict_data_sort on dict_data(dict_sort);


-- ============================================
-- 8. 日志管理表
-- ============================================

create table if not exists operation_log (
                                             id bigserial primary key,
                                             user_id integer,
                                             device_id integer,
                                             operation_type varchar(50),
                                             operation_module varchar(50),
                                             operation_desc text,
                                             request_method varchar(10),
                                             request_url varchar(500),
                                             request_params text,
                                             response_result text,
                                             ip_address varchar(50),
                                             user_agent varchar(500),
                                             status smallint,
                                             error_message text,
                                             execution_time integer,
                                             created_at timestamp default CURRENT_TIMESTAMP,
                                             constraint fk_log_user foreign key (user_id) references "user"(id) on delete set null,
                                             constraint fk_log_device foreign key (device_id) references device(id) on delete set null
);

comment on table operation_log is '操作日志表';
comment on column operation_log.id is '日志ID';
comment on column operation_log.user_id is '用户ID';
comment on column operation_log.device_id is '设备ID';
comment on column operation_log.operation_type is '操作类型：upload/download/delete/share';
comment on column operation_log.operation_module is '操作模块';
comment on column operation_log.operation_desc is '操作描述';
comment on column operation_log.request_method is '请求方法';
comment on column operation_log.request_url is '请求URL';
comment on column operation_log.status is '状态：1成功/0失败';
comment on column operation_log.execution_time is '执行时长（毫秒）';

create index if not exists idx_log_user on operation_log(user_id);
create index if not exists idx_log_created on operation_log(created_at);
create index if not exists idx_log_type on operation_log(operation_type);


-- ============================================
-- 9. 存储配置表
-- ============================================

create table if not exists storage_config (
                                              id serial primary key,
                                              user_id integer not null,
                                              total_quota bigint default 10737418240,
                                              used_quota bigint default 0,
                                              file_count integer default 0,
                                              last_sync timestamp,
                                              created_at timestamp default CURRENT_TIMESTAMP,
                                              updated_at timestamp default CURRENT_TIMESTAMP,
                                              constraint fk_storage_user foreign key (user_id) references "user"(id) on delete cascade,
                                              constraint uk_storage_user unique (user_id)
);

comment on table storage_config is '存储配置表';
comment on column storage_config.id is '配置ID';
comment on column storage_config.user_id is '用户ID';
comment on column storage_config.total_quota is '总配额（字节）默认10GB';
comment on column storage_config.used_quota is '已用配额（字节）';
comment on column storage_config.file_count is '文件数量';
comment on column storage_config.last_sync is '最后同步时间';

create index if not exists idx_storage_user on storage_config(user_id);


-- ============================================
-- 10. 分享管理表
-- ============================================

create table if not exists share_record (
                                            id bigserial primary key,
                                            user_id integer not null,
                                            file_id bigint not null,
                                            share_code varchar(32) not null unique,
                                            share_password varchar(20),
                                            expire_time timestamp,
                                            download_limit integer,
                                            download_count integer default 0,
                                            visit_count integer default 0,
                                            status smallint default 1,
                                            created_at timestamp default CURRENT_TIMESTAMP,
                                            constraint fk_share_user foreign key (user_id) references "user"(id) on delete cascade,
                                            constraint fk_share_file foreign key (file_id) references file(id) on delete cascade
);

comment on table share_record is '分享记录表';
comment on column share_record.id is '分享ID';
comment on column share_record.user_id is '分享人ID';
comment on column share_record.file_id is '文件ID';
comment on column share_record.share_code is '分享码';
comment on column share_record.share_password is '提取密码';
comment on column share_record.expire_time is '过期时间';
comment on column share_record.download_limit is '下载次数限制';
comment on column share_record.download_count is '已下载次数';
comment on column share_record.visit_count is '访问次数';
comment on column share_record.status is '状态：1有效/0失效';

create index if not exists idx_share_user on share_record(user_id);
create index if not exists idx_share_code on share_record(share_code);
create index if not exists idx_share_created on share_record(created_at);


-- ============================================
-- 11. 初始化基础数据
-- ============================================

-- 初始化字典类型
insert into dict_type (dict_code, dict_name, description)
select * from (values
                   ('device_type', '设备类型', '用户设备类型枚举'),
                   ('file_type', '文件类型', '文件类型枚举'),
                   ('sync_status', '同步状态', '同步任务状态枚举'),
                   ('download_status', '下载状态', '下载状态枚举'),
                   ('user_status', '用户状态', '用户状态枚举'),
                   ('operation_type', '操作类型', '操作日志类型枚举')
              ) as v(dict_code, dict_name, description)
where not exists (select 1 from dict_type where dict_type.dict_code = v.dict_code);

-- 初始化字典数据
insert into dict_data (dict_type_id, dict_label, dict_value, dict_sort)
select dt.id, v.dict_label, v.dict_value, v.dict_sort
from (values
          -- 设备类型
          ('device_type', '手机端', 'mobile', 1),
          ('device_type', 'Web端', 'web', 2),
          ('device_type', 'Windows端', 'windows', 3),
          ('device_type', 'Mac端', 'mac', 4),
          ('device_type', 'Linux端', 'linux', 5),

          -- 文件类型
          ('file_type', '文档', 'doc', 1),
          ('file_type', '图片', 'image', 2),
          ('file_type', '视频', 'video', 3),
          ('file_type', '音频', 'audio', 4),
          ('file_type', '其他', 'other', 5),

          -- 同步状态
          ('sync_status', '等待中', 'pending', 1),
          ('sync_status', '同步中', 'syncing', 2),
          ('sync_status', '已完成', 'completed', 3),
          ('sync_status', '失败', 'failed', 4),

          -- 下载状态
          ('download_status', '等待中', 'pending', 1),
          ('download_status', '下载中', 'downloading', 2),
          ('download_status', '已完成', 'completed', 3),
          ('download_status', '失败', 'failed', 4),
          ('download_status', '已取消', 'cancelled', 5),

          -- 用户状态
          ('user_status', '正常', '1', 1),
          ('user_status', '禁用', '0', 2),

          -- 操作类型
          ('operation_type', '上传', 'upload', 1),
          ('operation_type', '下载', 'download', 2),
          ('operation_type', '删除', 'delete', 3),
          ('operation_type', '分享', 'share', 4),
          ('operation_type', '移动', 'move', 5),
          ('operation_type', '重命名', 'rename', 6)
     ) as v(type_code, dict_label, dict_value, dict_sort)
         join dict_type dt on dt.dict_code = v.type_code
where not exists (
    select 1 from dict_data dd
    where dd.dict_type_id = dt.id
      and dd.dict_value = v.dict_value
);

-- 初始化角色
insert into role (role_code, role_name, description)
select * from (values
                   ('admin', '系统管理员', '拥有所有权限'),
                   ('user', '普通用户', '基础文件同步权限')
              ) as v(role_code, role_name, description)
where not exists (select 1 from role where role.role_code = v.role_code);

-- 初始化权限
insert into permission (permission_code, permission_name, permission_type, description, sort_order)
select * from (values
                   ('file:upload', '文件上传', 'api', '允许上传文件', 1),
                   ('file:download', '文件下载', 'api', '允许下载文件', 2),
                   ('file:delete', '文件删除', 'api', '允许删除文件', 3),
                   ('file:share', '文件分享', 'api', '允许分享文件', 4),
                   ('file:manage', '文件管理', 'menu', '文件管理菜单', 5),
                   ('user:manage', '用户管理', 'menu', '用户管理菜单（仅管理员）', 6),
                   ('system:config', '系统配置', 'menu', '系统配置菜单（仅管理员）', 7)
              ) as v(permission_code, permission_name, permission_type, description, sort_order)
where not exists (select 1 from permission where permission.permission_code = v.permission_code);