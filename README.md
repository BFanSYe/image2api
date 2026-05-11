# image2api

image2api 是一个面向图片生成和 OpenAI 兼容接口的 AIGC 聚合平台。项目包含用户创作端, 管理后台, OpenAI 兼容 API, 后台任务 Worker, 账号池和计费运营能力。

本仓库是独立维护的二开版本, 项目显示名统一为 `image2api`, Docker 镜像名统一以 `image2api` 开头。

## 核心能力

- 图片生成: 支持文生图, 图生图, 异步任务, 历史记录, 预览和下载.
- 原生高清: 支持上游原生 `gpt-image-2` Responses/Image generation 路线, 保留 1K/2K/4K 原生结果, 不做虚假放大.
- OpenAI 兼容 API:
  - `GET /v1/models`
  - `POST /v1/chat/completions`
  - `POST /v1/images/generations`
  - `POST /v1/images/edits`
  - `GET /v1/images/generations/:task_id`
  - `POST /v1/video/generations`
  - `GET /v1/video/generations/:task_id`
- 后台运营: 用户, 账单, CDK, 优惠码, 模型价格, Token/API Key 账号池, 代理, 请求日志, 上游日志.
- 账号池: 支持 API key 类型账号和会话类账号, 支持检测, 熔断, 轮换和失败记录.
- 部署: 支持本地开发, 单机 Docker Compose, Nginx/Caddy 反向代理.

## 技术栈

- 后端: Go 1.24, Gin, GORM, MySQL, Redis.
- 前端: React 18, Vite, TypeScript, Tailwind CSS.
- 部署: Docker, Docker Compose, Nginx, Caddy.

## 仓库结构

```text
.
├── backend/     # Go 后端: API / Admin / OpenAI 兼容 / Worker
├── frontend/    # 用户前台 + 管理后台
├── deploy/      # Docker Compose, Nginx, Caddy, 环境变量模板
├── docs/        # API, 部署, 上游账号, 图片生成说明
├── scripts/     # 本地开发辅助脚本
└── tools/       # 账号与数据转换工具
```

## 端口说明

### 对外端口

- `17080`: 用户前台.
- `17088`: 管理后台.
- `17200`: OpenAI 兼容 API.

### 内部端口

- `17180`: 用户后端 API.
- `17188`: 管理后台 API.
- `17200`: OpenAI 兼容 API.
- `23306` 或 `13306`: MySQL, 取决于 Compose 文件.
- `16379`: Redis.
- `18191`: FlareSolverr.

## Docker 镜像命名

本仓库内 Compose 默认使用以下镜像名:

```text
image2api/backend:dev
image2api/backend:latest
image2api/user-web:dev
image2api/user-web:latest
image2api/admin-web:dev
image2api/admin-web:latest
```

如发布到 GitHub Container Registry, 推荐使用:

```text
ghcr.io/BFanSYe/image2api-backend
ghcr.io/BFanSYe/image2api-user-web
ghcr.io/BFanSYe/image2api-admin-web
```

## 快速启动

### 1. 拉取代码

```bash
git clone https://github.com/BFanSYe/image2api.git
cd image2api
```

### 2. 准备环境变量

```bash
cp deploy/env/.env.example deploy/env/.env.local
```

必须修改以下生产密钥:

- `IMAGE2API_MYSQL_ROOT_PASSWORD`
- `IMAGE2API_MYSQL_PASSWORD`
- `IMAGE2API_DB_DSN`
- `IMAGE2API_JWT_SECRET`
- `IMAGE2API_JWT_REFRESH_SECRET`
- `IMAGE2API_AES_KEY`
- `IMAGE2API_CORS_ORIGINS`

密钥生成示例:

```bash
openssl rand -hex 32
```

### 3. 本地完整容器栈

```bash
cd deploy
docker compose -f docker-compose.dev-full.yml up -d --build
```

检查状态:

```bash
docker compose -f docker-compose.dev-full.yml ps
docker logs -f image2api-api-dev
docker logs -f image2api-admin-dev
docker logs -f image2api-openai-dev
docker logs -f image2api-worker-dev
```

### 4. 生产 Compose

```bash
cd deploy
docker compose --env-file ./env/.env.local -f docker-compose.yml up -d --build
```

访问入口:

- 用户前台: `http://你的域名:17080`
- 管理后台: `http://你的域名:17088/admin/`
- OpenAI 兼容 API: `http://你的域名:17200/v1`

## 本地开发

只启动 MySQL 和 Redis:

```bash
cd deploy
docker compose -f docker-compose.dev.yml up -d
```

后端:

```bash
cd backend
go test ./...
go run ./cmd/api
```

前端:

```bash
cd frontend
corepack enable
pnpm install
pnpm --filter @image2api/user dev
pnpm --filter @image2api/admin dev
```

## 上游账号和 gpt-image-2

- 后台添加 API key 类型上游账号时, `base_url` 应指向支持 Responses/Image generation 的上游.
- 需要原生 2K/4K 时, 上游必须真实支持 `gpt-image-2` 对应分辨率.
- 不支持原生高清的账号不会被当作高清账号使用.
- 上游失败原因可在后台生成日志和上游日志中查看.

更多说明见:

- [上游账号配置](docs/UPSTREAM_ACCOUNTS.md)
- [图片生成说明](docs/IMAGE_GENERATION.md)

## 安全边界

- 不要提交 `.env.local`, 数据库 dump, Redis dump, 上传文件, 生成图片缓存, 用户数据和真实上游 token.
- 公开仓库前必须运行敏感信息扫描.
- 本仓库当前默认许可证为保留权利, 公开开源前需要重新确认 LICENSE.

## 文档

- [开发规范](docs/01-开发规范-总览.md)
- [后端规范](docs/02-后端规范.md)
- [数据库设计](docs/03-数据库设计.md)
- [API 规范](docs/04-API规范.md)
- [前端规范](docs/05-前端规范.md)
- [部署与运维规范](docs/06-部署与运维规范.md)
- [安全说明](SECURITY.md)
- [变更记录](CHANGELOG.md)

## License

当前为保留权利版本, 见 [LICENSE](LICENSE). 如后续需要公开开源, 请先确认上游授权和所有第三方素材/代码许可证。
