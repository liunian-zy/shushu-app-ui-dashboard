# 模块说明: devops

## 1. 模块职责
- 统一 Docker Compose 编排开发环境
- 保证 Web/API/TTS/MySQL/Redis 在同一网络下可联动
- 提供可复用的容器化启动方式

## 2. 关键文件
- `docker-compose.yml`：服务编排入口
- `docker-compose.online.yml`：线上同步 API 编排入口（仅 server）
- `server/Dockerfile`：Go API 镜像构建，包含 ffmpeg 与 Go 代理配置
- `web/Dockerfile`：前端生产构建（Nginx 静态服务）
- `scripts/app_db_transfer.sh`：导出/导入 `app_db_` 表数据脚本

## 3. 服务清单
- `mysql`：MySQL 8（内网环境使用容器）
- `redis`：Redis 7
- `tts`：TTS 服务（仓库内 `./tts` 目录）
- `server`：Go API 服务
- `web`：前端 Vite 服务

## 4. 运行约定
- 内网 MySQL 使用 `mysql` 容器（`mysql:3306`），端口由 `MYSQL_HOST_PORT` 控制
- MySQL 库名由 `MYSQL_DATABASE` 控制，密码由 `MYSQL_ROOT_PASSWORD` 控制
- API 服务端口由根目录 `.env` 的 `APP_PORT` 控制（默认 `18080`）
- Web 对外端口由根目录 `.env` 的 `WEB_PORT` 控制（默认 `5173`）
- `APP_MODE=internal` 部署内网全功能；`APP_MODE=online` 仅保留同步 API
- 线上使用 `SYNC_API_KEY` 保护 `/api/sync/push`
- 本地测试线上模式使用根目录 `.env.online`
- Redis 通过容器名互联
- Redis 宿主机映射端口由 `REDIS_HOST_PORT` 控制（默认 `16379`）
- 需要设置 `TTS_API_KEY`（TTS 服务与 API 服务一致）
- 本地媒体目录通过 `LOCAL_STORAGE_HOST_PATH` 挂载到 `LOCAL_STORAGE_ROOT`
- `VITE_API_PROXY` 用于前端代理到 API 容器
- `JWT_SECRET` 必须配置，用于签发登录令牌
- `APP_MODE=internal` 启动时会自动执行 `server/migrations/*.sql` 初始化草稿表
- Web 生产容器通过 `web/nginx.conf.template` 反向代理 `/api`，上游由 `API_UPSTREAM` 控制
