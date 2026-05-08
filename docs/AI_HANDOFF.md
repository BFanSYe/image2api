# AI Handoff - gpt2api

metadata:
  repo: `/srv/gpt2api`
  branch: `main`
  base_ref: `origin/main`
  head_commit_at_last_code_change: `e5847e1`
  commit_range: `origin/main..HEAD`
  generated_at: `2026-05-08 14:12:39 CST`
  last_runtime_update: `2026-05-08 17:00:58 CST`
  intended_reader: `AI coding agent`
  language: `zh-CN`

## 0. 接手规则

- 以当前代码为最终事实来源. 本文档只总结 `origin/main..HEAD` 范围内的修改和运行状态.
- 修改前先执行 `git status --short --branch`, 确认是否有未提交改动.
- 后端测试需要 Go 1.24. 宿主机当前 `GOTOOLCHAIN=local go version` 为 `go1.22.2`, 不能直接跑 `go test ./...`.
- 前端使用 pnpm workspace. 根目录是 `frontend/`, package manager 是 `pnpm@9.7.0`.
- 运行中的本地容器栈使用 compose project `deploy`, 主要配置文件是 `deploy/docker-compose.server.yml` + `deploy/docker-compose.no-edge.yml`.
- 如果页面显示旧内容, 优先判断为旧容器或旧静态 bundle 未更新, 不要先改源码.

## 1. 提交清单

| commit | 主题 | 主要影响 |
|---|---|---|
| `c2c0337` | Fix container runtime issues | 修复容器运行配置, 配置环境变量覆盖, 管理后台 `/admin/` 基路径, compose/no-edge 栈, 运行时存储和 Nginx 路由. |
| `3d35d5e` | Fix unsupported image resolutions | 收紧 GPT Image 2 尺寸校验, 前端模型元数据增加 `resolutions`, 文档减少不支持尺寸误导. |
| `e6d4c8c` | Support 2K and 4K image output | 增加 1K/2K/4K 生成能力, 增加 handler/service/provider 尺寸参数传递与测试. |
| `fc81189` | Use native high resolution image generation | 高清图从后处理/放大路线切到原生 Responses `image_generation` 路线, 保留 1K web 路由. |
| `240a21e` | Prefer API key native image accounts | 原生高清路由优先使用 API key 或非 ChatGPT base 的账号, Codex OAuth 作为兜底. |
| `b6e4244` | Fix 4K image rendering from native responses | 修复原生 Responses 返回 base64/事件格式时的图片提取, 避免 4K 结果空白或无法展示. |
| `d6c7374` | Remove open source links | 删除用户端和管理端前端布局里的 GitHub/开源入口和混淆链接常量. |
| `e5847e1` | Add AI handoff document | 新增本文档, 用于其他 AI 工具接手代码和运行状态. |

## 2. 运行与部署变更

### 2.1 配置加载

Files:
- `backend/pkg/config/config.go`
- `backend/pkg/config/config_test.go`

Current behavior:
- 配置优先级为: 环境变量 > `config.${KLEIN_ENV}.yaml` > `config.yaml`.
- `KLEIN_ENV` 默认值是 `dev`.
- `prod` 环境会强校验:
  - `KLEIN_DB_DSN` 必填.
  - `KLEIN_REDIS_ADDR` 必填.
  - `KLEIN_JWT_SECRET` 和 `KLEIN_JWT_REFRESH_SECRET` 长度必须 >= 32.
  - `KLEIN_AES_KEY` 长度必须 >= 32.
- 新增显式环境变量覆盖:
  - Server: `KLEIN_API_PORT`, `KLEIN_ADMIN_PORT`, `KLEIN_OPENAI_PORT`, `KLEIN_WS_PORT`, `KLEIN_PPROF_PORT`.
  - Redis: `KLEIN_REDIS_ADDR`, `KLEIN_REDIS_PASSWORD`, `KLEIN_REDIS_DB`, `KLEIN_REDIS_POOL_SIZE`.
  - JWT: `KLEIN_JWT_SECRET`, `KLEIN_JWT_REFRESH_SECRET`, `KLEIN_JWT_ACCESS_TTL`, `KLEIN_JWT_REFRESH_TTL`.
  - Logger: `KLEIN_LOG_DIR`, `KLEIN_LOG_LEVEL`, `KLEIN_LOG_MAX_SIZE_MB`, `KLEIN_LOG_MAX_AGE_DAYS`, `KLEIN_LOG_COMPRESS`, `KLEIN_LOG_CONSOLE`.
  - Provider: `KLEIN_OPENAI_BASE`, `KLEIN_GROK_BASE`, `KLEIN_PROVIDER_REQUEST_TIMEOUT`, `KLEIN_PROVIDER_RETRY`.
  - CORS: `KLEIN_CORS_ORIGINS`, comma separated.
  - Other: `KLEIN_AES_KEY`, `KLEIN_NODE_ID`, `KLEIN_CDN_BASE`, `KLEIN_BILLING_POINT_UNIT`, pool/rate-limit keys.

### 2.2 Compose 栈

Files:
- `deploy/docker-compose.server.yml`
- `deploy/docker-compose.no-edge.yml`
- `deploy/docker-compose.dev-full.yml`
- `deploy/docker-compose.yml`

Current behavior:
- `docker-compose.server.yml` 是当前本地全栈运行入口, 使用真实 provider 配置, 容器名带 `-dev`.
- `docker-compose.no-edge.yml` 禁用 `edge` 服务, 并把 `user-web` 和 `admin-web` 端口绑定到本机:
  - `127.0.0.1:17080 -> user-web:80`
  - `127.0.0.1:17088 -> admin-web:80`
- 当前运行栈已按以下命令重建过前端服务:

```bash
cd /srv/gpt2api/deploy
sudo docker compose -f docker-compose.server.yml -f docker-compose.no-edge.yml up -d --build user-web admin-web
```

Observed runtime endpoints:
- User web: `http://127.0.0.1:17080`
- Admin web: `http://127.0.0.1:17088/admin/`

Stale UI diagnosis:
- 如果用户端仍显示旧开源入口, 抓首页:

```bash
curl -s http://127.0.0.1:17080 | rg 'index-[A-Za-z0-9_-]+\.js|开源|github|sourceHref|SOURCE_HREF'
```

- 当前正确用户端 bundle 是 `/assets/index-CaRF4OrU.js`.
- 旧容器曾返回 `/assets/index-fMyEpUFe.js`, 该 bundle 内含 `开源项目`, `开源地址`, `https://github.com/432539/gpt2api/`.
- 如果 bundle hash 仍是旧值, 重建并重启 `user-web`/`admin-web`.

### 2.3 Admin SPA 基路径

Files:
- `frontend/apps/admin/vite.config.ts`
- `frontend/apps/admin/src/main.tsx`
- `frontend/apps/admin/nginx.conf`
- `deploy/nginx-dev/admin-web.conf`
- `frontend/apps/admin/Dockerfile`

Current behavior:
- Admin Vite `base` 是 `/admin/`.
- React Router 使用 `basename="/admin"`.
- Admin Docker image 把 dist 复制到 `/usr/share/nginx/html/admin`.
- Admin Nginx:
  - `/` 和 `/admin` 302 到 `/admin/`.
  - `/admin/assets/*` 支持静态资产和长期缓存.
  - `/admin/` fallback 到 `/admin/index.html`.
  - Dev Nginx 还代理 `/admin/api/` 到 `admin:17188`.

### 2.4 HK 单域名公网 HTTPS 回源入口

Active public domain:
- `https://image.zuiying.shop`
- DNS A record points to HK public IP `69.165.73.30`.

Origin domain:
- `https://image-origin.zuiying.shop`
- DNS A record points to source public IP `43.134.122.243`.

Active routing:
- `https://image.zuiying.shop/` -> user web.
- `https://image.zuiying.shop/api/` -> user API.
- `https://image.zuiying.shop/admin/` -> admin web.
- `https://image.zuiying.shop/admin/api/` -> admin API.
- `https://image.zuiying.shop/v1/` -> OpenAI-compatible API.

Active回源 path:
- HK Nginx terminates TLS for `image.zuiying.shop`.
- HK Nginx proxies all paths to `https://43.134.122.243:443`.
- HK回源 request uses `Host: image-origin.zuiying.shop`.
- HK回源 TLS SNI uses `image-origin.zuiying.shop`.
- Source Nginx terminates TLS for `image-origin.zuiying.shop`.
- Source Nginx routes paths to local gpt2api ports:
  - `/` -> `127.0.0.1:17080`.
  - `/api/` -> `127.0.0.1:17180`.
  - `/admin/` -> `127.0.0.1:17088`.
  - `/admin/api/` -> `127.0.0.1:17188`.
  - `/v1/` -> `127.0.0.1:17200`.

HK Nginx files:
- Site: `/etc/nginx/sites-available/image.zuiying.shop.conf`.
- Enabled symlink: `/etc/nginx/sites-enabled/image.zuiying.shop.conf`.
- Existing `hk-api.zuiying.shop` config was not modified.

Source Nginx files:
- Site: `/etc/nginx/sites-available/image-origin.zuiying.shop`.
- Enabled symlink: `/etc/nginx/sites-enabled/image-origin.zuiying.shop`.
- Old source symlink `/etc/nginx/sites-enabled/image.zuiying.shop` was removed after switching source origin to `image-origin.zuiying.shop`.
- Old source available file `/etc/nginx/sites-available/image.zuiying.shop` remains as an inactive historical config.

TLS:
- HK certificate path: `/etc/letsencrypt/live/image.zuiying.shop/fullchain.pem`.
- HK key path: `/etc/letsencrypt/live/image.zuiying.shop/privkey.pem`.
- Source certificate path: `/etc/letsencrypt/live/image-origin.zuiying.shop/fullchain.pem`.
- Source key path: `/etc/letsencrypt/live/image-origin.zuiying.shop/privkey.pem`.
- Source certificate observed expiry: `2026-08-06`.

Cleaned previous attempts:
- Source SSH reverse tunnel service `/etc/systemd/system/gpt2api-hk-tunnel.service` was disabled and removed.
- Source local SSH key `/home/ubuntu/.ssh/gpt2api_hk_ed25519` and `.pub` were removed.
- HK `authorized_keys` entry with comment `gpt2api-hk-origin` was removed.
- HK temporary sshd config `/etc/ssh/sshd_config.d/99-gpt2api-pubkey.conf` was removed.
- Source WireGuard files `/etc/wireguard/wg-gpt2api.conf` and `/etc/wireguard/wg-gpt2api.key` were removed.
- HK WireGuard files `/etc/wireguard/wg-gpt2api.conf` and `/etc/wireguard/wg-gpt2api.key` were removed.
- HK old loopback tunnel ports `27080`, `27088`, `27200` are no longer used.
- HK old Nginx map `/etc/nginx/conf.d/gpt2api_ws_upgrade_map.conf` was removed.

Verification commands:

```bash
curl -sS -I https://image-origin.zuiying.shop/
curl -sS https://image-origin.zuiying.shop/api/v1/ping
curl -sS https://image-origin.zuiying.shop/admin/api/v1/ping
curl -sS https://image-origin.zuiying.shop/v1/health
curl -sS -I https://image.zuiying.shop/
curl -sS -I https://image.zuiying.shop/admin
curl -sS https://image.zuiying.shop/api/v1/ping
curl -sS https://image.zuiying.shop/admin/api/v1/ping
curl -sS https://image.zuiying.shop/v1/health
```

Expected key results:
- `image-origin.zuiying.shop` direct origin returns user SPA HTML.
- `image.zuiying.shop` through HK returns user SPA HTML.
- `/admin` on HK returns redirect to `https://image.zuiying.shop/admin/`, not to `image-origin.zuiying.shop`.
- `/api/v1/ping` returns `{"pong":true}`.
- `/admin/api/v1/ping` returns `{"pong":true,"scope":"admin"}`.
- `/v1/health` returns `{"ok":true}`.
- `/v1/models` without API key returns `401` with `API Key 无效`; this confirms the route reaches the OpenAI-compatible service.

Security notes:
- Do not write the HK root password into repository files, docs, shell history, or process arguments.
- Since the password was shared in chat, rotate the HK root password after confirming there is a separate recovery path or key login.
- The origin domain is publicly reachable by design. If origin bypass must be prevented later, add source firewall rules or application-level checks that only allow HK回源.

## 3. GPT Image 2 生成链路

### 3.1 输入参数与尺寸策略

Files:
- `backend/internal/handler/generation_handler.go`
- `backend/internal/handler/openai_handler.go`
- `backend/internal/service/generation_service.go`
- `backend/internal/provider/gpt/gpt.go`
- `frontend/apps/user/src/pages/create/CreateStudioPage.tsx`
- `frontend/apps/user/src/pages/keys/DocsPage.tsx`

Supported tiers:
- `1K`
- `2K`
- `4K`

Accepted explicit sizes include:
- 1K group: `1024x1024`, `1216x832`, `832x1216`, `1152x864`, `864x1152`, `1120x896`, `896x1120`, `1344x768`, `768x1344`, `1536x640`.
- 2K group: `1248x1248`, `1536x1024`, `1024x1536`, `1440x1088`, `1088x1440`, `1392x1120`, `1120x1392`, `1664x928`, `928x1664`, `1904x816`, `2048x2048`, `2048x1152`.
- 4K group: `2480x2480`, `3056x2032`, `2032x3056`, `2880x2160`, `2160x2880`, `2784x2224`, `2224x2784`, `3312x1872`, `1872x3312`, `3808x1632`, `3840x2160`, `2160x3840`.

Validation:
- `validateGenerationCreateRequest` rejects unsupported GPT Image 2 `resolution` or explicit `size`.
- OpenAI-compatible model metadata exposes `resolutions: ["1K","2K","4K"]`.
- Frontend model metadata expects `resolutions` on image models.

Native size normalization:
- Provider native supported sizes are currently:
  - `1024x1024`
  - `1536x1024`
  - `1024x1536`
  - `2048x2048`
  - `2048x1152`
  - `3840x2160`
  - `2160x3840`
- `closestImage2NativeSize` maps unsupported-but-accepted target sizes to nearest native aspect/area.
- Example from tests: `1664x928` normalizes to `2048x1152`.

### 3.2 路由选择

Files:
- `backend/internal/service/generation_service.go`
- `backend/internal/provider/gpt/gpt.go`

Routing rules:
- `shouldUseGPTWebRoute` returns true for `1K` and for default small sizes.
- `2K`, `4K`, or explicit size with area > 1,500,000 pixels uses native Responses image route.
- Native route uses GPT provider Responses endpoint:
  - normal base: `${base}/v1/responses`
  - Codex base: `https://chatgpt.com/backend-api/codex/responses`

Account selection:
- Non GPT Image 2 tasks use existing provider account pool reserve behavior.
- GPT Image 2 web route requires OAuth account.
- GPT Image 2 native route prefers:
  - API key accounts where `BaseURL` does not look like `chatgpt.com`.
  - OAuth accounts with non-ChatGPT `BaseURL`.
- Native fallback allows Codex OAuth account.
- Ordinary ChatGPT OAuth account without Codex client id must not be used for native Responses route.
- If a Codex OAuth account is selected and has no `BaseURL`, service sets `BaseURL` to `https://chatgpt.com/backend-api/codex`.

### 3.3 Provider behavior

Files:
- `backend/internal/provider/gpt/gpt.go`
- `backend/internal/provider/gpt/web_trace_test.go`

Native Responses route:
- Sends streamed `responses` request with `image_generation` tool.
- Sets `tool_choice` to `image_generation`.
- Includes text prompt and optional `input_image` refs for edits.
- Handles `quality` mapping:
  - `draft` or `low` -> `low`
  - `standard` or `medium` -> `medium`
  - `hd` or `high` -> `high`
- Retries once without `tool_choice` when upstream reports image tool choice not found, except Codex endpoint, where it returns a specific unsupported error.

Result extraction:
- Supports direct image URL.
- Supports `b64_json`.
- Supports data URL construction from native output base64.
- Supports `response.output_item.done`.
- Supports `response.completed`.
- Supports `response.image_generation_call.partial_image`.
- Supports web route direct URLs, file IDs, sediment IDs, download fallback, and diagnostics.

Reference images:
- `readRefImage` accepts:
  - `data:` URLs.
  - `/api/v1/gen/cached/...` files from `KLEIN_STORAGE_ROOT` or `/app/storage/public`.
  - HTTP/HTTPS URLs.
- Invalid cached path traversal is rejected.

### 3.4 Upstream logging

Files:
- `backend/migrations/20260501150000_generation_upstream_log.sql`
- `backend/internal/service/generation_service.go`
- `backend/internal/provider/gpt/gpt.go`

Behavior:
- Provider stages log to `generation_upstream_logs`.
- Log includes provider, stage, method, URL, status code, duration, request excerpt, response excerpt, error, account id, task id, and JSON metadata.
- Useful stages include `codex.start`, `codex.request`, `codex.retry`, `codex.unsupported`, `codex.decode`, `codex.asset`, `codex.success`, `codex.failed`, `web.*`.

## 4. Frontend changes

### 4.1 用户端

Files:
- `frontend/apps/user/src/pages/create/CreateStudioPage.tsx`
- `frontend/apps/user/src/pages/keys/DocsPage.tsx`
- `frontend/apps/user/src/lib/types.ts`
- `frontend/apps/user/src/layouts/AppLayout.tsx`

Behavior:
- Create studio consumes image model metadata `resolutions`.
- Docs page documents 1K/2K/4K image generation support.
- Open-source/GitHub links were removed from:
  - Desktop rail bottom action area.
  - Desktop version footer.
  - Mobile header.
- Removed obfuscated `sourceHref` helper and `Github` icon import from user layout.

### 4.2 管理端

Files:
- `frontend/apps/admin/src/layouts/AdminLayout.tsx`
- `frontend/apps/admin/src/main.tsx`
- `frontend/apps/admin/vite.config.ts`

Behavior:
- Admin SPA runs under `/admin`.
- Open-source/GitHub link was removed from sidebar footer.
- Removed `SOURCE_HREF` constant and `Github` icon import from admin layout.

Known non-goal:
- `README.md` still contains `## 开源地址` and `https://github.com/432539/gpt2api`.
- That README content was not removed because the request was specifically for frontend UI.

## 5. Verification status

Fresh verification run on `2026-05-08`:

```bash
sudo docker run --rm -v /srv/gpt2api/backend:/src -w /src golang:1.24-alpine go test ./...
```

Result:
- Passed.
- Notable passing packages:
  - `github.com/kleinai/backend/internal/provider/gpt`
  - `github.com/kleinai/backend/internal/service`
  - `github.com/kleinai/backend/pkg/config`
  - `github.com/kleinai/backend/pkg/jwtpayload`

Host limitation:

```bash
GOTOOLCHAIN=local go version
```

Result:
- `go version go1.22.2 linux/amd64`

Host `go test ./...` result:
- Failed before tests because `go.mod` requires `go 1.24` and local toolchain attempted `go: download go1.24 for linux/amd64: toolchain not available`.
- Use Docker or install Go 1.24 before running backend tests on host.

Frontend verification:

```bash
cd /srv/gpt2api/frontend
pnpm -r --filter './apps/*' typecheck
```

Result:
- Passed.
- `apps/user typecheck: Done`
- `apps/admin typecheck: Done`

```bash
cd /srv/gpt2api/frontend
sudo pnpm -r --filter './apps/*' build
```

Result:
- Passed.
- User bundle includes `/assets/index-CaRF4OrU.js`.
- Admin bundle includes `/admin/assets/index-BPyDyQB-.js`.

Runtime verification after frontend container rebuild:

```bash
curl -s http://127.0.0.1:17080 | rg -n '开源|github|gpt2api|sourceHref|SOURCE_HREF'
curl -s http://127.0.0.1:17088/admin/ | rg -n '开源|github|gpt2api|sourceHref|SOURCE_HREF'
```

Result:
- Only page titles matched `gpt2api`.
- No `开源地址`, `开源项目`, `sourceHref`, `SOURCE_HREF`, or GitHub source link remained in current served HTML.

## 6. 接手 checklist

Before changing code:
- Run `git status --short --branch`.
- Read this document.
- Read the current target file, do not rely only on this summary.
- For image-generation work, inspect `backend/internal/service/generation_service.go` and `backend/internal/provider/gpt/gpt.go` first.
- For deployment issues, inspect `deploy/docker-compose.server.yml`, `deploy/docker-compose.no-edge.yml`, and current Docker labels:

```bash
sudo docker inspect klein-user-web-dev --format '{{json .Config.Labels}}'
sudo docker inspect klein-admin-web-dev --format '{{json .Config.Labels}}'
```

When changing GPT Image 2 behavior:
- Add or update tests in `backend/internal/service/generation_service_test.go`.
- Add or update tests in `backend/internal/provider/gpt/web_trace_test.go`.
- Preserve routing invariant:
  - 1K -> web route unless explicit native size requires native.
  - 2K/4K -> native route.
  - API key native account preferred.
  - Codex OAuth fallback only.
- Preserve native response parsing for URL, `b64_json`, completed output, output item done, and partial image.

When changing frontend UI:
- Run frontend typecheck and build.
- Rebuild running containers if user validates via `17080` or `17088`.
- Compare served HTML bundle hash before assuming the change failed.

When changing deploy:
- Avoid editing secrets in `deploy/env`.
- Use `docker compose config` to inspect merged config before destructive service changes.
- Use targeted `up -d --build <service>` when possible, but be aware compose may recreate dependencies.

## 7. Current risk notes

- Host Go toolchain is too old for `go.mod`. Use Docker test command or install Go 1.24.
- Frontend `sudo pnpm ... build` was used because repository/build artifact permissions were root-owned in this environment.
- `README.md` still advertises the open-source URL. Remove it only if the user asks to remove project metadata globally, not just frontend UI.
- Running containers can serve stale bundles after source commits. Always verify the bundle name returned by `curl` before doing more source edits.
