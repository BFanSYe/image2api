# Security

## 不应入仓的内容

以下内容不得提交到仓库:

- `.env.local` 和任何真实环境变量文件.
- API key, session token, cookie, JWT secret, AES key.
- 数据库 dump, Redis dump, 生产日志.
- 用户上传文件, 生成图片缓存, 视频缓存.
- 证书私钥, SSH key, WireGuard key.
- 真实服务器 IP, 内部跳板信息, 运维临时凭据.

## 发布前检查

公开仓库或发布镜像前执行:

```bash
git status --short
git ls-files
gitleaks detect --source . --redact --verbose
```

如果历史提交中出现过真实密钥, 必须先轮换密钥, 再清理 Git 历史, 不能只删除当前文件。

## 环境变量

生产部署必须至少设置:

```text
IMAGE2API_ENV=prod
IMAGE2API_DB_DSN=
IMAGE2API_REDIS_ADDR=
IMAGE2API_JWT_SECRET=
IMAGE2API_JWT_REFRESH_SECRET=
IMAGE2API_AES_KEY=
IMAGE2API_CORS_ORIGINS=
```

`IMAGE2API_JWT_SECRET`, `IMAGE2API_JWT_REFRESH_SECRET`, `IMAGE2API_AES_KEY` 长度必须不少于 32 字节。
