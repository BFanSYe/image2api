# Changelog

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
