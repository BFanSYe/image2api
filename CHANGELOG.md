# Changelog

## v0.2.0 - 2026-05-17

### Added

- 新增灵感库页 `/inspire`: 全站作品按时间汇聚为公开瀑布流, 支持图片 / 视频筛选, 点开右抽屉可一键复用提示词或复制.
- 新增用户端接口 `GET /api/v1/inspire/feed?kind=&cursor=&page_size=`, 游标分页, 无需登录.
- 用户前台整体视觉重做为 Aurora 暗色风格: 粉紫液态 mesh 背景动效, Geist Sans / Mono 字体, 1px 描边玻璃卡, 顶部水平导航 + 头像下拉.
- 新增数据库迁移 `20260517190000_inspire_default_public.sql`.

### Changed

- **BREAKING**: `generation_result.is_public` 默认值从 `0` 改为 `1`, 升级时该迁移会将所有现存的成功生成结果回灌为公开. 部署前请确认运营策略允许全员公开历史作品; 如不希望公开, 升级前手工 `UPDATE generation_result SET is_public = 0` 或跳过该迁移.
- 用户前台默认主题切换为 dark.
- 灵感库的提示词复制走 `navigator.clipboard` + `execCommand('copy')` 兜底, 兼容更多移动浏览器.

### Fixed

- 修复用户端创作页 Studio 比例下拉被下方 "创建图片" 区遮挡的问题.
- 修复灵感页移动端抽屉无法滚动, 导致 "用这个提示词" / 复制按钮够不到的问题.
- 修复点击灵感页 "用这个提示词" 跳回 Studio 后提示词未自动填充的问题.

## v0.1.0 - 2026-05-12

### Added

- 新增公开发布部署快速开始文档.
- 新增发布流程文档和 GitHub Release 首版说明.
- 统一项目显示名为 `image2api`.
- 统一 Docker 镜像命名为 `image2api/backend`, `image2api/user-web`, `image2api/admin-web`.
- 统一后端 Go module 为 `github.com/BFanSYe/image2api/backend`.
- 统一前端 workspace 包名为 `@image2api/*`.
- 支持后台 API key 类型上游账号用于原生图片生成.
- 支持上游原生 `gpt-image-2` Responses/Image generation 路线.
- 支持原生 1K/2K/4K 图片结果展示.

### Changed

- 环境变量前缀从历史命名迁移为 `IMAGE2API_`.
- 本地容器, 网络和 volume 命名统一为 `image2api-*`.
- OpenAI 兼容 API key 前缀统一为 `sk-image2api-`.
- 删除包含运行态细节的内部交接文档.

### Fixed

- 修复原生 Responses 返回 base64/事件格式时图片提取失败导致前台看不到 4K 图片的问题.
- 修复高清图片路径只记录成功状态但前端无法渲染的问题.
