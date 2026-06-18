# 🔪 go-knifer

> 🔪 Go 开发的瑞士军刀，让日常编码更锋利。
>
> 🧰 提供字符串处理、集合操作、编码解码、加密、HTTP、缓存、ID 生成、日志、配置等开箱即用的工具包，按 `v*` 域包导入，只取所需。

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/imajinyun/go-knifer/actions/workflows/go.yml/badge.svg)](https://github.com/imajinyun/go-knifer/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/imajinyun/go-knifer)](./LICENSE)

## 📑 Table of Contents

- [📚 简介](#introduction)
- [✨ 为什么选择 go-knifer](#why-go-knifer)
- [🚀 安装](#install)
- [🧭 按场景查找](#find-by-scenario)
- [🧩 模块目录](#package-catalog)
- [🏗️ 架构](#architecture)
- [✅ 推荐 API](#recommended-apis)
- [📖 文档](#documentation)
- [📦 构建与测试](#build-and-test)
- [🛡️ 治理](#governance)
- [🤝 贡献](#contributing)
- [⭐ Star go-knifer](#star-go-knifer)

<a id="introduction"></a>

## 📚 简介

`go-knifer` 是一个面向 Go 项目的实用工具集合：把项目里反复出现的字符串处理、集合操作、编解码、加密、HTTP、JSON、缓存、定时任务、JWT、日志、配置、系统信息等能力沉淀成可复用的工具包。

项目根包 `github.com/imajinyun/go-knifer` 仅作为模块入口说明使用；实际 API 位于公开的 `v*` facade 子包中，应用可以只导入当前领域所需的包。

<a id="why-go-knifer"></a>

## ✨ 为什么选择 go-knifer

`knifer` 来自 “knife”：像一把随手可用的小刀，解决日常 Go 开发里的高频小问题。

- 🧰 **聚焦的 facade**：直接导入 `vstr`、`vslice`、`vhttp`、`vcrypto` 等领域包。
- 🧪 **可测试的 options**：大量 API 提供 `WithXxx` options 和 provider 注入，便于确定性测试。
- 🛡️ **安全默认值**：安全敏感工具优先采用显式错误、有边界读取、SSRF-aware URL 访问和路径穿越检查。
- 📚 **领域文档**：详细 quickstart 位于 [`docs/doc`](./docs/doc/README.CN.md)，让根 README 保持易扫读。

<a id="install"></a>

## 🚀 安装

项目要求 Go 1.25 或更高版本。

```bash
go get github.com/imajinyun/go-knifer
```

<a id="find-by-scenario"></a>

## 🧭 按场景查找

不确定该引入哪个包？从你要做的事出发：

| 我想要… | 使用 |
| --- | --- |
| 使用 FIFO/LRU/LFU/TTL 缓存 | [`vcache`](docs/doc/04-vcache.md) |
| Base64 / Hex 编解码 | [`vcodec`](docs/doc/05-vcodec.md) |
| 安全加载本地或远程配置 | [`vconf`](docs/doc/06-vconf.md) |
| SHA/HMAC、AES-GCM/RSA-PSS、参数签名 | [`vcrypto`](docs/doc/09-vcrypto.md) |
| 使用标准库辅助函数发送 HTTP 请求 | [`vhttp`](docs/doc/18-vhttp.md) |
| 使用 Resty-based 辅助函数发送 HTTP 请求 | [`vresty`](docs/doc/37-vresty.md) |
| 生成 UUID / Snowflake / NanoId | [`vid`](docs/doc/19-vid.md) |
| 敏感数据脱敏 | [`vmask`](docs/doc/28-vmask.md) |
| 创建、查询、转换、合并、差集或排序 map | [`vmap`](docs/doc/27-vmap.md) |
| 对 slice 做过滤 / 映射 / 去重 / 分页 | [`vslice`](docs/doc/41-vslice.md) |
| 裁剪、切分、命名转换、比较文本或判断空白字符串 | [`vstr`](docs/doc/42-vstr.md) |
| URL 编解码/解析，或安全打开不可信 HTTP(S) 资源 | [`vurl`](docs/doc/45-vurl.md) |

👉 每个包的完整清单见 [中文文档中心](./docs/doc/README.CN.md#package-catalog)。

<a id="package-catalog"></a>

## 🧩 模块目录

`go-knifer` 采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

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
- 🌐 在线 Go 文档：[pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)
- 🧾 API 快照：[`docs/api/exports.txt`](./docs/api/exports.txt)
- 🗺️ AI 项目地图：[`llms.txt`](./llms.txt)
- 🧯 安全策略：[`SECURITY.md`](./SECURITY.md)
- 📝 变更日志：[`CHANGELOG.md`](./CHANGELOG.md)

<a id="build-and-test"></a>

## 📦 构建与测试

下载源码：

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

运行常用本地检查：

```bash
make test        # 单元测试
make ci-test     # CI test-job 门禁
make check       # 完整本地门禁：测试、vet、lint、漏洞、覆盖率、API 检查
```

常用聚焦命令：

```bash
make quick-check
make security-check
make bench-core
make bench-facade
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

<a id="star-go-knifer"></a>

## ⭐ Star go-knifer

如果这个项目减少了你的重复代码，欢迎给它一个 Star。你的反馈和贡献会帮助它成为更趁手的 Go 工具集合。
