# cos-sync

把本地文件传到腾讯云 COS 桶，直接拿到一个公网可访问的分享链接。单二进制、跨平台、零运行时依赖。

## 特性

- **单二进制**：Go 编译，Windows / macOS / Linux 全覆盖，下载即用
- **多桶配置**：`~/.cos-sync/config.yaml` 里管理多对凭证，按别名切换
- **时间戳前缀**：自动给文件名加 `20060102_150405_` 前缀，避免重名覆盖
- **直接拼公有 URL**：适合"公有读私有写"的桶，链接永久有效
- **静默模式**：`-q` 只吐 URL，方便管道串接
- **Markdown 输出**：`--md` 直接吐 `![](url)`，贴博客 / wiki 即用
- **自动复制剪贴板**：上传成功后链接自动写入系统剪贴板，stderr 一行提示；没装剪贴板工具时静默跳过，不刷屏

## 安装

从 `dist/` 目录里挑对应平台的二进制：

| 平台 | 文件 |
|---|---|
| Linux x86_64 | `cos-sync-linux-amd64` |
| Linux ARM64 | `cos-sync-linux-arm64` |
| macOS Intel | `cos-sync-darwin-amd64` |
| macOS Apple Silicon | `cos-sync-darwin-arm64` |
| Windows x86_64 | `cos-sync-windows-amd64.exe` |

放到 `PATH` 里（如 `/usr/local/bin/cos-sync` 或 `~/bin/cos-sync`）即可全局调用。

## 配置

复制模板：

```bash
mkdir -p ~/.cos-sync
cp config.example.yaml ~/.cos-sync/config.yaml
chmod 600 ~/.cos-sync/config.yaml   # 含密钥，务必限权
```

填入真实凭证：

```yaml
default: work                       # 不传 --bucket 时用这个

buckets:
  work:
    secret_id:  AKIDxxxxxxxxxxxx
    secret_key: xxxxxxxxxxxxxxxx
    region:     ap-guangzhou        # 不带 cos. 前缀
    bucket:     my-work-bucket-1234567890   # 完整名（含 appid）
  personal:
    secret_id:  AKIDyyyyyyyyyyyy
    secret_key: yyyyyyyyyyyyyyyy
    region:     ap-beijing
    bucket:     my-personal-bucket-1234567890
```

字段说明：

- `default`：缺省别名。不传 `--bucket` 时走这里
- `secret_id` / `secret_key`：腾讯云 API 密钥，需要在 CAM 里有 `cos:PutObject` 权限
- `region`：地域代码，如 `ap-beijing` / `ap-shanghai` / `ap-guangzhou`
- `bucket`：桶完整名，形如 `<name>-<appid>`

## 使用

```bash
# 上传到 default 桶，打印人类可读的一行
cos-sync photo.jpg
# Uploaded → https://my-work-bucket-1234567890.cos.ap-guangzhou.myqcloud.com/20260619_153022_photo.jpg

# 指定别名
cos-sync photo.jpg --bucket personal

# 静默模式，只输出 URL，便于管道
cos-sync screenshot.png -q | xclip    # Linux
cos-sync screenshot.png -q | pbcopy   # macOS

# Markdown 图片语法输出，方便贴进博客 / wiki
cos-sync screenshot.png --md
# ![](https://my-work-bucket-1234567890.cos.ap-guangzhou.myqcloud.com/20260619_153022_screenshot.png)

# flag 位置随意
cos-sync --bucket personal photo.jpg -q
```

实际写入 COS 的 object key 形如 `20260619_153022_photo.jpg`（本地时区），文件名重名也互不覆盖。

## 退出码

| 码 | 含义 |
|---|---|
| 0 | 上传成功 |
| 1 | 运行时错误（文件读不到、网络/COS 上传失败） |
| 2 | 配置或参数错误（无配置、别名不存在、flag 不对） |

## 从源码构建

需要 Go 1.21+：

```bash
make build           # 五个平台全编一遍，产物在 dist/
make clean           # 清掉 dist/
VERSION=v0.1.0 make build    # 把版本号嵌进 --version 输出
```

## 设计边界

- 仅单文件上传，不支持批量 / 目录递归
- 时间戳用本机时区，不做时区参数化
- 不做预签名 URL（公有读桶用不到；如未来要切私有桶，再扩展）
- 不显示上传进度（单文件，COS SDK 完成就返回）

## 剪贴板依赖

| 平台 | 依赖 |
|---|---|
| macOS / Windows | 系统原生，无需额外工具 |
| Linux X11 | `xclip` 或 `xsel`，缺了静默跳过 |
| Linux Wayland | `wl-copy`，缺了静默跳过 |

无剪贴板环境（如 SSH 到无头服务器）时上传照常成功，stderr 不输出任何剪贴板相关内容。
