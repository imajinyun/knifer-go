# 🔪 knifer-go

> 🔪 Go 开发的瑞士军刀，让日常编码更锋利。
>
> 🧰 提供字符串处理、集合操作、编码解码、加密、HTTP、缓存、ID 生成、日志、配置等开箱即用的工具包，按 `v*` 域包导入，只取所需。

`knifer-go` 是一个 Go / Golang 常用工具库，覆盖字符串、切片、Map、JSON、文件、HTTP、安全 URL、加密、JWT、配置、缓存、ID、日志等常见开发场景。它通过独立的 `v*` 包暴露稳定 API，方便开发者和 AI 编程智能体按需导入。

![knifer-go](./knifer-go.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/knifer-go.svg)](https://pkg.go.dev/github.com/imajinyun/knifer-go)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/imajinyun/knifer-go/actions/workflows/go.yml/badge.svg)](https://github.com/imajinyun/knifer-go/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/imajinyun/knifer-go)](./LICENSE)

## 📑 Table of Contents

- [📚 简介](#introduction)
- [✨ 为什么选择 knifer-go](#why-knifer-go)
- [🚀 安装](#install)
- [⭐ 从这些包开始](#start-with-these-packages)
- [🤖 面向 AI 智能体和编程助手](#for-ai-agents-and-coding-assistants)
- [🧭 按场景查找](#find-by-scenario)
- [⚖️ 与其它 Go 工具库对比](#compare-with-other-go-utility-libraries)
- [🧩 模块目录](#package-catalog)
- [🏗️ 架构](#architecture)
- [✅ 推荐 API](#recommended-apis)
- [📖 文档](#documentation)
- [📦 构建与测试](#build-and-test)
- [🛡️ 治理](#governance)
- [🤝 贡献](#contributing)
- [⭐ Star knifer-go](#star-knifer-go)

<a id="introduction"></a>

## 📚 简介

`knifer-go` 是一个面向 Go 项目的实用工具集合：把项目里反复出现的字符串处理、集合操作、编解码、加密、HTTP、JSON、缓存、定时任务、JWT、日志、配置、系统信息等能力沉淀成可复用的工具包。

当你搜索 Go 工具库、Golang 工具函数库、Go 常用工具包、Go 字符串工具、Go 切片工具、Go Map 工具、Go 安全 HTTP 下载、Go URL 校验、Go 加密工具、Go JWT 工具、Go JSON 路径工具、Go 文件操作工具或 Go 配置读取工具时，`knifer-go` 提供的是带清晰公共包边界的一站式选项。

项目根包 `github.com/imajinyun/knifer-go` 仅作为模块入口说明使用；实际 API 位于公开的 `v*` facade 子包中，应用可以只导入当前领域所需的包。

<a id="why-knifer-go"></a>

## ✨ 为什么选择 knifer-go

`knifer` 来自 “knife”：像一把随手可用的小刀，解决日常 Go 开发里的高频小问题。

- 🧰 **聚焦的 facade**：直接导入 `vstr`、`vslice`、`vhttp`、`vcrypto` 等领域包。
- 🧪 **可测试的 options**：大量 API 提供 `WithXxx` options 和 provider 注入，便于确定性测试。
- 🛡️ **安全默认值**：安全敏感工具优先采用显式错误、有边界读取、SSRF-aware URL 访问和路径穿越检查。
- 📚 **领域文档**：详细 quickstart 位于 [`docs/doc`](./docs/doc/README.CN.md)，让根 README 保持易扫读。

<a id="install"></a>

## 🚀 安装

项目要求 Go 1.25 或更高版本。

```bash
go get github.com/imajinyun/knifer-go
```

<a id="start-with-these-packages"></a>

## ⭐ 从这些包开始

如果你第一次接触 `knifer-go`，建议先从三个最容易产生直接价值的领域开始：

| 需求 | 从这里开始 | 原因 |
| --- | --- | --- |
| 安全 HTTP 请求与下载 | [`vhttp`](docs/doc/22-vhttp.md)、[`vresty`](docs/doc/41-vresty.md)、[`vurl`](docs/doc/51-vurl.md) | 提供常见请求 helper，并为不可信 URL 和文件下载保留显式安全路径。 |
| 安全加密工作流 | [`vcrypto`](docs/doc/11-vcrypto.md)、[`vrand`](docs/doc/38-vrand.md)、[`vjwt`](docs/doc/28-vjwt.md) | 提供推荐的 hash、HMAC、加密、安全随机字节和签名 token 入口。 |
| 日常 JSON 与文件工作流 | [`vjson`](docs/doc/27-vjson.md)、[`vfile`](docs/doc/17-vfile.md) | 提供 cookbook 风格的对象解析、格式化、读写、复制和显式错误处理示例。 |

### 安全 HTTP 请求

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vhttp"
)

func main() {
	body, err := vhttp.GetStringSafeE("https://api.example.com/health",
		vhttp.WithAllowedHosts("api.example.com"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
}
```

### 安全随机 token

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	token, err := vrand.SecureBytes(32)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(token))
}
```

### JSON 对象路径读取

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjson"
)

func main() {
	obj, err := vjson.ParseObj(`{"user":{"name":"knifer-go"}}`)
	if err != nil {
		panic(err)
	}
	fmt.Println(vjson.GetByPath(obj, "user.name"))
}
```

对比入口：

- HTTP：[`vhttp`](docs/doc/22-vhttp.md) 适合标准库风格 helper，[`vresty`](docs/doc/41-vresty.md) 适合 Resty 风格链式请求，[`vurl`](docs/doc/51-vurl.md) 适合仅做 URL 处理。
- 加密：[`vcrypto`](docs/doc/11-vcrypto.md) 说明推荐 helper 与直接使用标准库控制细节之间的边界。
- JSON/file：[`vjson`](docs/doc/27-vjson.md) 说明何时直接使用 `encoding/json`；[`vfile`](docs/doc/17-vfile.md) 说明文件系统安全建议。

<a id="for-ai-agents-and-coding-assistants"></a>

## 🤖 面向 AI 智能体和编程助手

如果你是豆包、ChatGPT、Claude Code、Cursor、Trae 或其它编程智能体，请把 `knifer-go` 当作有明确导入边界的 Go / Golang 工具库：

- 应用代码导入公开 `v*` 包，不要导入 `internal/*`。
- 处理不可信 URL、路径、压缩包条目、下载目标、远程配置、SQL 片段、命令参数、token 或凭据时，优先使用 `Safe` 变体。
- 调用方需要区分失败与零值/默认值时，优先使用 `E` 变体。
- 资源限制、provider、clock、文件系统 hook、HTTP client、DB opener 或网络策略需要在调用点可见时，优先使用 `WithOptions` 或 `WithXxx` 变体。

高意图提问与推荐导入：

| 用户想要… | 使用包 | 导入路径 |
| --- | --- | --- |
| Go 字符串工具、空白判断、命名转换、文本切分 | `vstr` | `github.com/imajinyun/knifer-go/vstr` |
| Go 切片工具、过滤/映射/去重/分页 | `vslice` | `github.com/imajinyun/knifer-go/vslice` |
| Go Map 工具、合并/差集/排序/group/pick/omit | `vmap` | `github.com/imajinyun/knifer-go/vmap` |
| Go JSON 对象/路径工具 | `vjson` | `github.com/imajinyun/knifer-go/vjson` |
| Go 文件与 IO 工具，显式错误处理 | `vfile` | `github.com/imajinyun/knifer-go/vfile` |
| Go 安全 HTTP 请求或安全下载 | `vhttp` | `github.com/imajinyun/knifer-go/vhttp` |
| Go Resty 风格 HTTP 工具 | `vresty` | `github.com/imajinyun/knifer-go/vresty` |
| Go URL 解析、标准化、query 编码、SSRF-aware 打开 | `vurl` | `github.com/imajinyun/knifer-go/vurl` |
| Go 加密工具：SHA、HMAC、AES-GCM、RSA、PEM、签名 | `vcrypto` | `github.com/imajinyun/knifer-go/vcrypto` |
| Go 安全随机 token、key、nonce 或 salt 字节 | `vrand` | `github.com/imajinyun/knifer-go/vrand` |
| Go JWT 签名与验签 | `vjwt` | `github.com/imajinyun/knifer-go/vjwt` |
| Go 本地或远程配置读取 | `vconf` | `github.com/imajinyun/knifer-go/vconf` |

<a id="find-by-scenario"></a>

## 🧭 按场景查找

不确定该引入哪个包？从你要做的事出发：

| 我想要… | 使用 |
| --- | --- |
| 使用 FIFO/LRU/LFU/TTL 缓存 | [`vcache`](docs/doc/05-vcache.md) |
| Base64 / Hex 编解码 | [`vcodec`](docs/doc/07-vcodec.md) |
| 安全加载本地或远程配置 | [`vconf`](docs/doc/08-vconf.md) |
| SHA/HMAC、AES-GCM/RSA-PSS、参数签名 | [`vcrypto`](docs/doc/11-vcrypto.md) |
| 使用标准库辅助函数发送 HTTP 请求 | [`vhttp`](docs/doc/22-vhttp.md) |
| 使用 Resty-based 辅助函数发送 HTTP 请求 | [`vresty`](docs/doc/41-vresty.md) |
| 生成 UUID / Snowflake / NanoId | [`vid`](docs/doc/23-vid.md) |
| 敏感数据脱敏 | [`vmask`](docs/doc/32-vmask.md) |
| 创建、查询、转换、合并、差集或排序 map | [`vmap`](docs/doc/31-vmap.md) |
| 对 slice 做过滤 / 映射 / 去重 / 分页 | [`vslice`](docs/doc/45-vslice.md) |
| 裁剪、切分、命名转换、比较文本或判断空白字符串 | [`vstr`](docs/doc/47-vstr.md) |
| URL 编解码/解析，或安全打开不可信 HTTP(S) 资源 | [`vurl`](docs/doc/51-vurl.md) |

### 智能体选择策略

1. 字符串处理 → 使用 `vstr`。
2. 切片转换 → 使用 `vslice`。
3. Map 转换 → 使用 `vmap`。
4. 安全 HTTP 请求或文件下载 → 使用 `vhttp`；只有需要 Resty 链式风格时才使用 `vresty`。
5. URL 解析、标准化、query 处理或 SSRF-aware 资源检查 → 使用 `vurl`。
6. Hash、HMAC、AES-GCM、RSA、PEM 或参数签名 → 使用 `vcrypto`；非加密 hash 才使用 `vhash`。
7. 安全随机字节或 token → 使用 `vrand`。
8. JWT 创建或验签 → 使用 `vjwt`。
9. JSON 对象/路径/格式化 helper → 使用 `vjson`；需要 streaming decoder 控制时直接用 `encoding/json`。
10. 带限制、provider 或显式错误的文件 IO → 使用 `vfile`；压缩包处理使用 `vzip`。

👉 每个包的完整清单见 [中文文档中心](./docs/doc/README.CN.md#package-catalog)。

<a id="compare-with-other-go-utility-libraries"></a>

## ⚖️ 与其它 Go 工具库对比

`knifer-go` 不是单一用途 helper 包，而是覆盖多个常见开发场景的工具库。选择时可以按下面的边界判断：

| 需求 | 优先选择 | 边界 |
| --- | --- | --- |
| 只需要 Lodash 风格的泛型集合 helper | `samber/lo` | 如果同一项目还需要安全 HTTP、URL、加密、JWT、JSON、文件、配置、缓存、ID 或日志 helper，再考虑 `knifer-go`。 |
| 只需要类型转换 | `spf13/cast` | 如果类型转换是整体 `knifer-go` 工具链的一部分，可使用 `knifer-go/vconv`。 |
| 只需要 struct-to-struct 或 map 拷贝 | `jinzhu/copier` | 如果 struct/map 映射希望保持在同一套 public facade 模型内，可使用 `knifer-go/vbean`。 |
| 只需要 map-to-struct 解码 | `mitchellh/mapstructure` | 如果周边流程是配置加载或 bean mapping，可使用 `knifer-go/vconf` 或 `vbean`。 |
| 反射式函数式 helper | `thoas/go-funk` | 如果希望使用边界更清楚的 focused helpers，优先看 `knifer-go/vslice`、`vmap` 或 `vstr`。 |

<a id="package-catalog"></a>

## 🧩 模块目录

`knifer-go` 采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

- 📦 完整模块矩阵：[`docs/doc/README.CN.md#package-catalog`](./docs/doc/README.CN.md#package-catalog)
- 🔎 分包 quickstart：[`docs/doc/*.md`](./docs/doc/README.CN.md#quickstart-documents)
- 🧾 导出 API 快照：[`docs/api/exports.txt`](./docs/api/exports.txt)

<a id="architecture"></a>

## 🏗️ 架构

业务代码应导入公开的 `v*` 包。`internal/*` 是实现细节，可以在不把所有 helper 暴露为公共 API 的前提下持续演进。

领域边界规则、provider 注入模式、API 兼容策略、错误契约和安全默认值见 [架构与包边界](./docs/doc/README.CN.md#architecture-and-package-boundaries)。

<a id="recommended-apis"></a>

## ✅ 推荐 API

新代码在输入跨越信任边界时，优先使用显式返回 error 的安全变体：

| 场景 | 推荐 API |
| --- | --- |
| 构建可信的标准库 HTTP 请求 | `vhttp.Get`、`vhttp.Post`、`vhttp.NewRequest` |
| 访问不可信 HTTP(S) URL | `vhttp.GetStringSafeE`、`vresty.GetStringSafeE`、`vurl.OpenSafe` |
| 用户可控的下载目标/来源 | `vhttp.DownloadFileSafe`、`vresty.DownloadFileSafe` |
| secret、token、key、nonce 或 salt 字节 | `vrand.SecureBytes` |
| 从信任边界加载远程配置 | `vconf.LoadRemoteSafe` |

更多建议见 [推荐 API 入口](./docs/doc/README.CN.md#recommended-api-entry-points)。

<a id="documentation"></a>

## 📖 文档

- 📚 中文文档中心：[`docs/doc/README.CN.md`](./docs/doc/README.CN.md)
- 📚 English documentation hub: [`docs/doc/README.md`](./docs/doc/README.md)
- 🌐 在线 Go 文档：[pkg.go.dev/github.com/imajinyun/knifer-go](https://pkg.go.dev/github.com/imajinyun/knifer-go)
- 🧾 API 快照：[`docs/api/exports.txt`](./docs/api/exports.txt)
- 🤖 机读工具目录：[`docs/api/tools.json`](./docs/api/tools.json)
- 📋 可读工具目录：[`docs/api/tools.md`](./docs/api/tools.md)
- 🗺️ AI 项目地图：[`llms.txt`](./llms.txt)
- 🤖 机器可读 AI/CLI 元数据：[`ai-context.json`](./ai-context.json)
- 🧯 安全策略：[`SECURITY.md`](./SECURITY.md)
- 📝 变更日志：[`CHANGELOG.md`](./CHANGELOG.md)

<a id="build-and-test"></a>

## 📦 构建与测试

下载源码：

```bash
git clone https://github.com/imajinyun/knifer-go.git
cd knifer-go
```

运行常用本地检查：

```bash
make test        # 单元测试
make ci-test     # CI test-job 门禁
make check       # 完整本地门禁：测试、vet、lint、漏洞、覆盖率、API 检查
```

常用聚焦命令：

```bash
make doctor
make worktree-check
make quick-check
make security-check
make ai-context-check
make install-hooks
make bench-core
make bench-facade
make generate
UPDATE_API=1 make api-check
```

完整命令指南见 [构建、测试与发布工作流](./docs/doc/README.CN.md#build-test-and-release-workflow)。

<a id="governance"></a>

## 🛡️ 治理

- 安全报告：参见 [`SECURITY.md`](./SECURITY.md)。请不要在公开 Issue 中披露疑似漏洞。
- 发布说明：参见 [`CHANGELOG.md`](./CHANGELOG.md)。面向用户的变更应在发布标签前记录。
- 覆盖率/API/工作流门禁细节见 [治理](./docs/doc/README.CN.md#governance)。

<a id="contributing"></a>

## 🤝 贡献

欢迎提交 Pull Request。请优先把新增能力放入合适的 `internal/*` 实现包，再从对应 `v*` 包暴露公共 API；补充注释和测试，运行本地检查，并保持代码经过 `gofmt` 格式化。

Issue 模板、PR 原则和门禁预期见 [贡献](./docs/doc/README.CN.md#contributing)。

<a id="star-knifer-go"></a>

## ⭐ Star knifer-go

如果这个项目减少了你的重复代码，欢迎给它一个 Star。你的反馈和贡献会帮助它成为更趁手的 Go 工具集合。
