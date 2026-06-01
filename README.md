# gopay-app

独立 GoPay App 多用户管理服务，提供 gRPC API、Dashboard HTTP API、WebUI 远程模块与 OTP webhook。

## 职责

- 管理 GoPay App 账号状态、设备指纹、代理会话和 token 生命周期。
- 支持登录、注册、改绑手机号、PIN、注销、余额检查和状态查询。
- 接收 WhatsApp/SMS OTP webhook，并以 `channel + target + otp` 投递到等待中的 n8n flow；账号详情提供一次性手动 OTP 兜底，不写入最新 OTP 缓存。
- 承载 Midtrans + GoPay linking/payment runtime；GPT 只负责 ChatGPT checkout/Stripe/snap_token 准备。
- 提供 `/api/gopay/*` 与 `/mf/gopay/*`，GoPayAccount 页面归属本仓。

GoPay 侧 payment action 使用中性 `/api/gopay/actions/gopay-payment/*`；`activate`、Plus 等 GPT 业务语义不进入本仓。

## 入口

- gRPC: `cmd/gopay-app-server`
- Dashboard HTTP API: `/api/gopay`
- WebUI remote: `/mf/gopay/remoteEntry.js`
- Proto: `proto/gopay_app.proto`

## 关键环境变量

- `GOPAY_APP_PORT`：gRPC 端口，默认 `50051`。
- `GOPAY_HTTP_LISTEN_ADDR`：Dashboard HTTP/API 监听地址，默认 `:8080`。
- `GOPAY_DASHBOARD_STATIC_DIR`：GoPay WebUI 静态文件目录，默认 `/app/dashboard/gopay`。
- `GOPAY_N8N_WEBHOOK_BASE_URL`：GoPay account workflow 的 n8n webhook base。
- `GOPAY_STATE_REDIS_URL`：Redis URL，必填。
- `GOPAY_STATE_KEY_PREFIX`：状态 key 前缀，默认 `byte-v-forge:gopay-app:state`。
- `GOPAY_STATE_TTL_SECONDS`：状态 TTL，默认 7 天。
- `PROXY_RUNTIME_HTTP_ADDR`：proxy-runtime HTTP 地址。
- `GOPAY_OTP_WEBHOOK_LISTEN_ADDR`：OTP webhook HTTP 监听地址，默认 `:8081`。
- `GOPAY_OTP_SUBMIT_URL`：OTP webhook 提交地址，默认本服务 `/api/gopay/otp/submit`。

## 生成 proto

```bash
./scripts/generate-proto.sh
```
