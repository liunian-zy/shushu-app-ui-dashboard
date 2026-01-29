# 模块说明: server-api

## 1. 模块职责
- 提供草稿数据的 CRUD 与列表聚合接口
- 提供 OSS 预签名与签名 URL 的统一服务
- 提供媒体规则配置、校验与压缩流程
- 提供提交/二次确认/差异记录与审计
- 提供同步到线上表的校验与落库流程

## 2. 关键接口
### 2.1 认证
- `POST /api/auth/login`：账号密码登录获取 JWT
- `POST /api/auth/bootstrap`：首次初始化管理员（仅在无用户时允许）
- `GET /api/auth/me`：当前用户信息
- `GET /api/users`：用户列表（管理员）
- `POST /api/users`：创建用户（管理员）

### 2.2 任务协作
- `GET /api/tasks`：任务列表（按 `draft_version_id` 过滤）
- `POST /api/tasks`：创建任务（管理员）
- `PUT /api/tasks/:id`：更新/指派任务（管理员）
- `POST /api/tasks/:id/assist`：协助任务记录
- `GET /api/tasks/:id/actions`：任务操作历史（返回 `actor_name`/`actor_username`）

### 2.3 本地文件
- `POST /api/local-files/upload`：上传本地媒体文件，返回 `local://` 路径
- `GET /api/local-files/*path`：读取本地媒体文件内容（用于媒体预览）

### 2.4 OSS
- `POST /api/oss/pre-sign`：获取上传预签名
- `POST /api/oss/sign-url`：获取下载签名 URL

### 2.5 媒体规则与处理
- `GET /api/media/rules`：查询媒体规则
- `POST /api/media/rules`：新增媒体规则
- `PUT /api/media/rules/:id`：更新媒体规则
- `DELETE /api/media/rules/:id`：删除媒体规则
- `POST /api/media/validate`：校验媒体是否合规（支持临时规则覆盖）
- `POST /api/media/transform`：压缩/转码并写入媒体版本（支持临时规则覆盖与无损模式）

### 2.6 草稿录入
- `GET /api/draft/version-names`
- `POST /api/draft/version-names`
- `PUT /api/draft/version-names/:id`
- `DELETE /api/draft/version-names/:id`
- `GET /api/draft/banners`
- `POST /api/draft/banners`
- `PUT /api/draft/banners/:id`
- `DELETE /api/draft/banners/:id`
- `GET /api/draft/identities`
- `POST /api/draft/identities`
- `PUT /api/draft/identities/:id`
- `DELETE /api/draft/identities/:id`
- `GET /api/draft/scenes`
- `POST /api/draft/scenes`
- `PUT /api/draft/scenes/:id`
- `DELETE /api/draft/scenes/:id`
- `GET /api/draft/clothes-categories`
- `POST /api/draft/clothes-categories`
- `PUT /api/draft/clothes-categories/:id`
- `DELETE /api/draft/clothes-categories/:id`
- `GET /api/draft/photo-hobbies`
- `POST /api/draft/photo-hobbies`
- `PUT /api/draft/photo-hobbies/:id`
- `DELETE /api/draft/photo-hobbies/:id`
- `GET /api/draft/app-ui-fields`
- `POST /api/draft/app-ui-fields`
- `GET /api/draft/config-extra-steps`
- `POST /api/draft/config-extra-steps`
- `PUT /api/draft/config-extra-steps/:id`
- `DELETE /api/draft/config-extra-steps/:id`

### 2.7 提交与确认
- `POST /api/draft/submit`：提交快照并生成差异
- `POST /api/draft/confirm`：二次确认
- `GET /api/draft/submissions`：提交历史列表（返回提交人/确认人名称）

### 2.8 同步到线上
- 内网：`POST /api/sync` → 主动推送到线上 API
  - 覆盖已有版本时需要 `confirm=true`
- 线上：`POST /api/sync/push`（API Key 保护）→ 写入线上业务表
  - `modules` 支持 `version_names` 单独同步版本配置

### 2.9 TTS
- `POST /api/tts/convert`：文本转语音并落地本地文件，返回 `audio_path`/`audio_url`
- `GET /api/tts/presets`：语音预设列表（`?all=1` 管理员可查看停用项）
- `POST /api/tts/presets`：新增语音预设（管理员）
- `PUT /api/tts/presets/:id`：更新语音预设（管理员）
- `DELETE /api/tts/presets/:id`：删除语音预设（管理员，默认预设不可删除）

### 2.10 任务与模板
- `POST /api/tasks/:id/complete`：任务完成时上传草稿媒体到 OSS
- `GET /api/identity-templates`：身份模板列表
- `POST /api/identity-templates`：创建身份模板
- `PUT /api/identity-templates/:id`：更新身份模板
- `DELETE /api/identity-templates/:id`：删除身份模板
- `GET /api/identity-templates/:id/items`：获取模板明细
- `POST /api/identity-templates/:id/items`：新增模板明细
- `PUT /api/identity-template-items/:id`：更新模板明细
- `DELETE /api/identity-template-items/:id`：删除模板明细
- `POST /api/draft/identities/apply-template`：套用身份模板到草稿

### 2.11 历史与审计
- `GET /api/audit/logs`：查询审计日志
- `GET /api/field-history`：查询字段变更历史
- `GET /api/media/versions`：查询媒体版本记录

### 2.12 概览
- `GET /api/dashboard/summary`：概览统计（按 `draft_version_id` 返回任务/媒体/同步摘要）

## 3. 同步校验规则
- `app_version_name`、`location_name` 必填
- `banners.image` 必填（当同步轮播图模块）
- `identities.name` 必填（当同步身份模块）
- `scenes` 至少 1 条，且 `scenes.name` 必填（当同步场景模块）
- `clothes_categories.name`、`photo_hobbies.name` 必填（当同步对应偏好模块）
- `config_extra_steps.step_index`/`field_name`/`label` 必填（当同步额外配置模块）
- 线上表非空字段使用默认值补齐（如 `status`、`sort` 等）

## 4. 数据与存储约定
- 草稿表使用 `app_db_` 前缀
- 本地上传使用 `local://` 前缀表示内网文件路径
- OSS 仅存储 `path`，响应中返回 `*_url` 签名地址
- 同步时按 `app_version_name` 进行整表替换写入
- 版本创建时若未传 `app_version_name` 将根据 `location_name` 自动生成
- 版本创建时若未传 `feishu_field_names` 会写入默认字段列表（SD 模式包含 `SD模式`）
- 场景列表接口不返回水印/OSS 样式字段

## 5. 依赖与约束
- 依赖 `ffmpeg/ffprobe` 进行媒体处理
- 媒体规则由管理员配置并应用
- 依赖 TTS 服务 `TTS_BASE_URL` + `TTS_API_KEY`
- 认证依赖 `JWT_SECRET` 与 `JWT_EXPIRE_HOURS`
