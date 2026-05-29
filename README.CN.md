# go-knifer

> 🍬 A set of Go tools that keep development sharp.

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-00ADD8?logo=go)](https://go.dev/)

## 📚 Introduction

`go-knifer` 是一个面向 Go 项目的常用工具集合，定位类似 Java 生态中的 Hutool：把项目里反复出现的字符串处理、集合操作、编解码、加密、HTTP、JSON、缓存、定时任务、JWT、日志、配置、系统信息等能力沉淀成可复用的工具包。

项目根包 `github.com/imajinyun/go-knifer` 仅作为模块入口说明使用；实际能力按领域拆分到多个 `v*` 对外子包中，用户可以按需导入，避免把无关 API 混入业务代码。

## 🔪 Origin of the `go-knifer` name

`knifer` 来自 “knife”：像一把随手可用的小刀，解决日常 Go 开发里的高频小问题。它不试图替代标准库，而是对标准库与常见工程实践做轻量封装，让代码更短、更统一、更容易维护。

## ✨ How go-knifer changes the way we code

以前，计算一个 MD5 往往需要在业务代码里重复写样板逻辑：

```go
sum := md5.Sum([]byte("hello"))
text := hex.EncodeToString(sum[:])
```

现在，使用 `go-knifer` 可以直接调用工具方法：

```go
text := vbase.MD5Hex("hello")
```

这类封装能减少重复代码、降低复制粘贴带来的隐患，也让团队内相同场景使用一致的 API。

## 🧩 Module

当前项目采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

| 模块 | 导入路径 | 功能说明 |
| --- | --- | --- |
| `vbase` | `github.com/imajinyun/go-knifer/vbase` | 基础工具：字符串、随机 ID、编解码、日期时间、数字、集合、Map、类型转换、文件 IO、正则、命名转换等。 |
| `vbf` | `github.com/imajinyun/go-knifer/vbf` | 布隆过滤器：bitmap/bitset/filter 抽象，以及多种字符串哈希算法。 |
| `vcache` | `github.com/imajinyun/go-knifer/vcache` | 泛型缓存：FIFO、LFU、LRU、Timed、Weak、NoCache，支持 TTL、淘汰监听与懒加载。 |
| `vcaptcha` | `github.com/imajinyun/go-knifer/vcaptcha` | 图片验证码：线条、圆圈、扭曲、GIF 验证码，支持随机/数学表达式生成器。 |
| `vcron` | `github.com/imajinyun/go-knifer/vcron` | Cron 表达式解析与任务调度，支持默认调度器和自定义调度器。 |
| `vcrypto` | `github.com/imajinyun/go-knifer/vcrypto` | 加密与摘要：MD5/SHA、HMAC、随机字节、AES-CBC/AES-GCM、RSA-OAEP、RSA PEM 编解码。 |
| `vextra` | `github.com/imajinyun/go-knifer/vextra` | 额外工具：gzip/zlib、zip/unzip、emoji、Go template 渲染、常用校验。 |
| `vhttp` | `github.com/imajinyun/go-knifer/vhttp` | 链式 HTTP 客户端、下载、全局 Header/Timeout、BasicAuth、User-Agent 解析、简易服务端。 |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | 有序 JSON 对象/数组、JSON 解析与格式化、路径表达式读写、Bean/List 转换、XML/JSON 转换。 |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT 创建、解析、签名、验签与时间字段校验，支持 HMAC、RSA、ECDSA、none 等 signer。 |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | 日志 facade：console/color console logger、日志级别、全局 logger 与静态日志函数。 |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | 分组配置读取：setting/properties 风格文本和简单 YAML 子集，支持类型化读取。 |
| `vskt` | `github.com/imajinyun/go-knifer/vskt` | TCP socket 工具：普通连接、NIO/AIO server/client、协议编解码接口。 |
| `vsys` | `github.com/imajinyun/go-knifer/vsys` | 系统与运行时信息：主机、OS、用户、Go runtime、进程内存、goroutine、环境变量等。 |

## 🚀 Install

项目要求 Go 1.20 或更高版本。

```bash
go get github.com/imajinyun/go-knifer
```

Go 会按实际导入的子包拉取模块，例如：

```go
import (
 "github.com/imajinyun/go-knifer/vbase"
 "github.com/imajinyun/go-knifer/vhttp"
)
```

## 📝 Quick start

### 基础工具与 JSON

```go
package main

import (
 "fmt"

 "github.com/imajinyun/go-knifer/vbase"
 "github.com/imajinyun/go-knifer/vjson"
)

func main() {
 name := vbase.DefaultIfBlank("", "go-knifer")

 obj := vjson.NewObject().
  Set("id", vbase.FastUUID()).
  Set("name", name).
  Set("tags", []string{"go", "tool"})

 fmt.Println(obj.GetString("name"))
 fmt.Println(obj.ToStringPretty())
}
```

### LRU 缓存与懒加载

```go
package main

import (
 "fmt"
 "time"

 "github.com/imajinyun/go-knifer/vcache"
)

func main() {
 c := vcache.NewLRUWithTimeout[string, int](3, 5*time.Minute)
 c.Put("answer", 42)

 value, ok := c.Get("answer")
 fmt.Println(value, ok)

 loaded, err := c.GetOrLoad("miss", func() (int, error) {
  return 100, nil
 })
 fmt.Println(loaded, err)
}
```

### 链式 HTTP 请求

```go
package main

import (
 "fmt"
 "time"

 "github.com/imajinyun/go-knifer/vhttp"
)

func main() {
 vhttp.SetGlobalTimeout(3 * time.Second)

 resp := vhttp.Get("https://example.com").
  Query("lang", "go").
  Header("X-Client", "go-knifer").
  FollowRedirects(true).
  Execute()

 if resp.Err() != nil {
  panic(resp.Err())
 }

 fmt.Println(resp.Status())
 fmt.Println(resp.ContentType())
 fmt.Println(resp.Body())
}
```

### 摘要与 JWT

```go
package main

import (
 "fmt"
 "time"

 "github.com/imajinyun/go-knifer/vcrypto"
 "github.com/imajinyun/go-knifer/vjwt"
)

func main() {
 fmt.Println(vcrypto.SHA256Hex("hello"))

 key := []byte("secret")
 token, err := vjwt.NewJWT().
  SetSubject("user-1").
  SetPayload("role", "admin").
  SetExpiresAt(time.Now().Add(time.Hour)).
  SetKey(key).
  Sign()
 if err != nil {
  panic(err)
 }

 jwt, err := vjwt.ParseJWT(token)
 if err != nil {
  panic(err)
 }

 fmt.Println(jwt.SetKey(key).Verify())
}
```

## 📖 Doc

- 根包说明：`doc.go`
- 对外 API：各 `v*` 子包的 `doc.go` 与 facade 文件
- 测试示例：各模块下的 `*_test.go`
- 在线文档：[pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)

## 📦 Download & Build

下载源码：

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

运行测试：

```bash
go test ./...
```

格式化代码：

```bash
gofmt -w .
```

## 🤝 Provide feedback or suggestions on bugs

如果发现问题或希望补充新工具，请通过 GitHub Issues 反馈。建议提供：

- Go 版本与操作系统；
- `go-knifer` 版本或 commit；
- 最小可复现代码；
- 期望行为与实际行为；
- 相关错误日志或测试输出。

## ✅ Principles of PR (pull request)

欢迎提交 PR。为了保持工具库稳定，请尽量遵循以下原则：

1. 新增能力优先放入合适的 `internal/*` 实现包，再由对应 `v*` 包暴露对外 API；
2. 新增或修改公共 API 时补充必要注释；
3. 为核心逻辑补充单元测试，提交前执行 `go test ./...`；
4. 保持代码经过 `gofmt` 格式化；
5. 避免引入不必要的第三方依赖，优先复用标准库。

## ⭐ Star go-knifer

如果这个项目减少了你的重复代码，欢迎给它一个 Star。你的反馈和贡献会帮助它成为更趁手的 Go 工具集合。
