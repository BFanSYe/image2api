# 上游账号配置

## 账号类型

image2api 后台账号池主要包含两类上游账号:

- API key 类型账号: 适合 OpenAI 兼容接口, Responses API, 原生图片生成.
- 会话类账号: 适合需要浏览器会话或网页登录态的上游.

## 原生 gpt-image-2 高清账号

需要原生 2K/4K 图片时, 推荐添加 API key 类型账号:

```text
provider: gpt
account_type: api_key
model: gpt-image-2
base_url: 支持 Responses/Image generation 的上游地址
api_key: 上游 API key
status: enabled
```

要求:

- 上游必须原生支持 `gpt-image-2`.
- 上游必须支持 Responses API 的 `image_generation` 能力.
- 上游必须返回图片 base64, 文件 ID, URL, 或兼容事件流中的图片结果.
- 如果上游只支持 1K, 不应标记为 2K/4K 可用账号.

## 排查顺序

1. 后台确认账号状态为启用.
2. 后台确认模型映射中的上游模型为 `gpt-image-2`.
3. 提交测试任务并查看生成日志.
4. 查看上游日志中的 request/response 摘要.
5. 确认结果 URL 可通过 GET 访问, 且 `content-type` 为图片类型.

## 注意事项

- 不要把真实 API key 写进 `.env.example`, README 或 issue.
- 不要把用户任务日志和上游原始响应直接公开.
- 如果 4K 任务状态成功但图片不可见, 优先检查结果提取和缓存 URL, 不要先改上游账号.
