# 发布流程

本页记录 image2api 独立维护版的发布检查和 tag 规则。所有命令默认在仓库根目录执行, 特别注明的除外。

## 版本规则

使用语义化版本:

```text
vMAJOR.MINOR.PATCH
```

当前首个独立维护版:

```text
v0.1.0
```

## 发布前检查

检查工作区:

```bash
git status --short --branch
git diff --check
```

扫描敏感信息:

```bash
sudo docker run --rm -v "$PWD:/repo" zricethezav/gitleaks:latest detect --source /repo --redact --verbose

tmp_dir=$(mktemp -d)
git archive HEAD | tar -x -C "$tmp_dir"
sudo docker run --rm -v "$tmp_dir:/repo:ro" zricethezav/gitleaks:latest dir /repo --redact --verbose
rm -rf "$tmp_dir"
```

执行后端测试:

```bash
sudo docker run --rm -v "$PWD/backend:/src" -w /src golang:1.24-alpine go test ./...
```

执行前端检查:

```bash
cd frontend
corepack prepare pnpm@9.7.0 --activate
pnpm install --frozen-lockfile
pnpm -r typecheck
pnpm -r --filter './apps/*' build
```

校验 Compose:

```bash
cd deploy
docker compose -f docker-compose.dev-full.yml config
docker compose -f docker-compose.server.yml -f docker-compose.no-edge.yml config
docker compose --env-file ./env/.env.example -f docker-compose.yml config
```

## 发布步骤

提交并推送 `main` 后, 等待 GitHub Actions 通过, 再创建发布 tag。

创建 tag:

```bash
git tag -a v0.1.0 -m "image2api v0.1.0"
git push origin v0.1.0
```

只推送当前发布 tag, 不使用 `git push --tags`, 避免把历史本地 tag 一并发布。

创建 GitHub Release 时, release notes 从 `CHANGELOG.md` 同步。

## 发布后检查

确认:

```bash
git ls-remote --tags origin
```

并检查 GitHub Releases 页面是否出现对应版本。
