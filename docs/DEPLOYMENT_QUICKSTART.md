# 部署快速开始

本页用于新服务器首次部署 image2api。推荐先使用 no-edge 栈完成本机验证, 再接入公网域名和 TLS。

## 1. 前置条件

确认服务器已安装:

- Docker Engine.
- Docker Compose plugin.
- Git.
- 可访问 GitHub 仓库.

检查命令:

```bash
docker version
docker compose version
git --version
```

## 2. 克隆仓库

```bash
git clone https://github.com/BFanSYe/image2api.git
cd image2api
```

## 3. 推荐: no-edge 本机验证

使用 no-edge 栈时, 前台和后台只监听 `127.0.0.1`, 适合先验证容器构建, 数据库初始化和后台登录。

```bash
cd deploy
docker compose -f docker-compose.server.yml -f docker-compose.no-edge.yml config
docker compose -f docker-compose.server.yml -f docker-compose.no-edge.yml up -d --build
docker compose -f docker-compose.server.yml -f docker-compose.no-edge.yml ps
```

验证入口:

```bash
curl -I http://127.0.0.1:17080
curl -I http://127.0.0.1:17088/admin/
```

默认后台账号:

```text
username: admin
password: admin123
```

首次登录后立即修改密码。

## 4. 生产 TLS 栈

仅当已经准备好域名和证书时使用 `docker-compose.yml`。没有证书时不要直接启动 TLS 栈。

复制配置:

```bash
cd deploy
cp env/.env.example env/.env.local
```

生成密钥:

```bash
openssl rand -hex 32
```

必须修改:

```text
IMAGE2API_MYSQL_ROOT_PASSWORD
IMAGE2API_MYSQL_PASSWORD
IMAGE2API_DB_DSN
IMAGE2API_JWT_SECRET
IMAGE2API_JWT_REFRESH_SECRET
IMAGE2API_AES_KEY
IMAGE2API_CORS_ORIGINS
```

准备证书文件:

```text
deploy/certs/image2api.example.crt
deploy/certs/image2api.example.key
deploy/certs/admin.image2api.example.crt
deploy/certs/admin.image2api.example.key
deploy/certs/api.image2api.example.crt
deploy/certs/api.image2api.example.key
```

按真实域名修改:

```text
deploy/nginx/user.conf
deploy/nginx/admin.conf
deploy/nginx/openai.conf
```

启动:

```bash
docker compose --env-file ./env/.env.local -f docker-compose.yml config
docker compose --env-file ./env/.env.local -f docker-compose.yml up -d --build
docker compose --env-file ./env/.env.local -f docker-compose.yml ps
```

## 5. 配置上游账号

部署完成后, 登录后台添加上游账号。真实图片生成必须配置至少一个可用上游账号。

原生 2K/4K 图片要求:

- 使用 API key 类型账号.
- 上游支持 `gpt-image-2`.
- 上游支持 Responses/Image generation.
- 上游真实支持请求的 2K/4K 分辨率.

## 6. 常用排查

查看服务状态:

```bash
docker compose ps
```

查看后端日志:

```bash
docker logs -f image2api-api
docker logs -f image2api-admin
docker logs -f image2api-openai
docker logs -f image2api-worker
```

查看 no-edge 日志:

```bash
docker logs -f image2api-api-dev
docker logs -f image2api-admin-dev
docker logs -f image2api-openai-dev
docker logs -f image2api-worker-dev
```

清理并重建:

```bash
docker compose down
docker compose up -d --build
```

不要删除 MySQL volume, 除非确认要清空数据库。
