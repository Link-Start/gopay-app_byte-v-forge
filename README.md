# gopay-app

`gopay-app` 是 GoPay App 账号与支付运行时服务，负责多用户状态、设备指纹、账号动作、OTP 接入和 GoPay payment runtime。

## 核心能力

- 管理 GoPay App 账号、设备指纹、代理会话、token 生命周期和多用户运行态。
- 支持登录、注册、改绑手机号、PIN、注销、余额检查和账号状态查询。
- 接收 WhatsApp/SMS OTP webhook，并把验证码投递到等待中的业务流程。
- 承载 Midtrans + GoPay linking/payment runtime；GPT 侧只传入已准备好的 checkout 参数。
- 提供 gRPC、Dashboard HTTP API 和 GoPay 管理前端模块。

## 使用方式

业务服务通过 proto/gRPC、HTTP webhook 或部署配置集成 GoPay 能力，不直接读写本服务状态存储。短期运行态进入 Redis TTL，长期事实由服务自有存储维护。

## 入口

- 服务入口：`cmd/gopay-app-server`
- 契约真源：`proto/gopay_app.proto`
- Dashboard API：`/api/gopay/*`
- 前端模块：`webui/`
- 工作流素材：`workflows/`

## 常用检查

```sh
./scripts/generate-proto.sh
git diff --check
```
