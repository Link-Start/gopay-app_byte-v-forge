# GoPay protocol notes from DanOps-1 PR #28

Source: `DanOps-1/Gpt-Agreement-Payment` PR #28, file `output/gopay_2.8.0_extract/REVERSE_KNOWLEDGE_BASE.md`.

App-service notes retained here. Payment/checkout/tokenization execution is owned by `gpt-private` and is intentionally not implemented in this repo.

- GoPay app default headers now target app version `2.8.0` / build `2080`.
- GoPay app device headers now keep the PR #28 captured app/profile shape, but generate per-device persisted device identity and lower hardware/network IDs (`x-uniqueid`, `D1`, MediaDrm/Widevine, WiFi BSSID/SSID, `m1_connection_id`, `m1_signature`, `m1_signature_time`, `m1_device_uuid`, AppsFlyer, Firebase app instance ID, FCM-like `X-DeviceToken`, AD ID/App Set ID metadata, `X-IMEI`, and `X-IpAddress`). The generated device also binds a realistic Android make/model/screen profile and uses APK-observed no-space forms such as `Android,16` and `Redmi,23117RK66C`. Use `GOPAY_STATIC_DEVICE_IDENTITY=1` only for deterministic protocol debugging; do not use it for signup/payment flows because a globally reused device identity is easy to correlate.
- `X-M1` follows the 2.8.0 string labels observed in the APK: `3:appsflyerId`, `6:wifiMac`, `7:wifiSSID`, `8:screen`, `9:locationMethod`, `11:widevineId`, `13:signature`, `14:signatureTime`, `15:firebaseAppInstanceId`, `16:deviceUUID`.
- A connected Android device profile can be exported with `scripts/gopay-adb-device-profile.sh` and sourced into the app-service environment. The script only reads non-secret device/profile fields; `X-DeviceToken` is app-private and is generated when `GOPAY_DEVICE_TOKEN` is not provided.
- Default location is aligned with `Gojek-Timezone=Asia/Jakarta` / `Gojek-Country-Code=ID`; override `GOPAY_LOCATION` and `GOPAY_LOCATION_ACCURACY` when replaying a fresh capture.
- Dynamic egress is acquired through `PROXY_RUNTIME_HTTP_ADDR`; `GenerateDeviceProxy` creates a fresh sticky ID session before phone/signup probing instead of relying on static GoPay egress envs.
- Signup now intentionally avoids a fixed machine-speed sequence: callers should add human-like spacing before signup, and the GoPay app service waits another configurable random interval before `/cvs/v1/initiate` (`GOPAY_SIGNUP_INITIATE_JITTER_MIN_SECONDS`, `GOPAY_SIGNUP_INITIATE_JITTER_MAX_SECONDS`, defaults `8..25`). A signup rate-limit response marks the current device/egress/phone state with `_signup_cooldown_until` (`GOPAY_SIGNUP_RATE_LIMIT_COOLDOWN_SECONDS`, default `900`) so the same generated state is not immediately reused.
- GoPay app signing defaults to endpoint-aware `auto`: legacy v1 for auth/CVS/login endpoints, v2 for confirmed v2 endpoints such as `customer.gopayapi.com/api/v1/users/pin/tokens/nb`. Override with `GOPAY_SIGN_VERSION=v1|v2` only for focused debugging.
- PR #28 notes do not mention TLS/JA3/ClientHello details. They only mention server-side device fingerprinting and `x-m1`/device fields. The app now binds one Android TLS profile when a device/proxy state is generated and reuses it for that state. Useful envs:
  - `GOPAY_TLS_PROFILE=okhttp4_android_12` pins a profile.
  - `GOPAY_TLS_PROFILES=okhttp4_android_12,okhttp4_android_13` limits the random pool.
  - `GOPAY_TLS_RANDOM_EXTENSION_ORDER=1` enables experimental per-connection TLS extension order randomization; default is off to keep device-level TLS behavior stable.
  - `GOPAY_TLS_FORCE_HTTP1=1` restores forced HTTP/1.1 if needed.
Useful local probes:

```bash
python3 scripts/probe-gopay-v2-signature.py --self-test
GOPAY_SSO_TOKEN=... GOPAY_SIGNED_MSG_TEMPLATE=/tmp/big_msg_1867.bin \
  python3 scripts/probe-gopay-v2-signature.py --send
```

`--self-test` verifies the deterministic v2 cipher vector from PR #28. Full server acceptance still requires a fresh SSO token and a captured `signed_msg` template matching the device fields.
