# go-knifer

> 🍬 A set of Go tools that keep development sharp.

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-00ADD8?logo=go)](https://go.dev/)

## 📚 Introduction

`go-knifer` 是一个面向 Go 项目的常用工具集合：把项目里反复出现的字符串处理、集合操作、编解码、加密、HTTP、JSON、缓存、定时任务、JWT、日志、配置、系统信息等能力沉淀成可复用的工具包。

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
text := vhash.MD5Hex("hello")
```

这类封装能减少重复代码、降低复制粘贴带来的隐患，也让团队内相同场景使用一致的 API。

## 🧩 Module

当前项目采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

| 模块 | 导入路径 | 功能说明 |
| --- | --- | --- |
| `vstr` | `github.com/imajinyun/go-knifer/vstr` | 字符串工具：空白判断、裁剪、切分、截取、格式化、emoji、命名转换、默认值和 HTML 转义。 |
| `vslice` | `github.com/imajinyun/go-knifer/vslice` | Slice 工具：包含/索引、反转、去重、拼接、过滤/映射、截取、合并、集合操作和分页。 |
| `vmap` | `github.com/imajinyun/go-knifer/vmap` | Map 工具：空判断、keys、values、反转和合并。 |
| `vconv` | `github.com/imajinyun/go-knifer/vconv` | 宽松类型转换：string、int、int64、float64、bool、bytes 及默认值版本。 |
| `vdate` | `github.com/imajinyun/go-knifer/vdate` | 日期时间工具：常用布局、解析/格式化、日/月/年起止、偏移和比较。 |
| `vfile` | `github.com/imajinyun/go-knifer/vfile` | 文件与 IO 工具：读写复制、按行读取、mkdir/touch/delete、文件名处理和静默关闭。 |
| `vcodec` | `github.com/imajinyun/go-knifer/vcodec` | 编解码工具：Base64、URL-safe Base64、Hex 和 URL query 转义。 |
| `vurl` | `github.com/imajinyun/go-knifer/vurl` | URL 与 URI 工具：解析、标准化、相对 URL 补全、query 编解码、Data URI 构造、协议判断和文件 URL 转换。 |
| `vobj` | `github.com/imajinyun/go-knifer/vobj` | 对象工具：nil/空值判断、相等性、默认值、克隆/序列化、比较、类型检查和容器辅助。 |
| `vser` | `github.com/imajinyun/go-knifer/vser` | 序列化工具：gob 编码/解码、泛型反序列化、深拷贝、类型注册和可选的解码类型校验。 |
| `vver` | `github.com/imajinyun/go-knifer/vver` | 版本工具：版本号比较、大小关系判断、表达式匹配、闭区间范围和自定义多表达式分隔符。 |
| `vref` | `github.com/imajinyun/go-knifer/vref` | 反射工具：字段查找与赋值、方法发现与调用、构造函数风格调用、类型/值工具和方法分类判断。 |
| `vzip` | `github.com/imajinyun/go-knifer/vzip` | ZIP、gzip、zlib 工具：压缩包创建/解压、条目读取、遍历、追加、内存条目和流式压缩。 |
| `vdes` | `github.com/imajinyun/go-knifer/vdes` | 脱敏工具：姓名、证件号、电话、地址、邮箱、密码、车牌、银行卡、IP、护照号和信用代码遮罩。 |
| `vnum` | `github.com/imajinyun/go-knifer/vnum` | 数字工具：精确加减乘除、舍入模式、格式化、数字判断、不重复随机数、range、阶乘/组合数、最大公约数/最小公倍数、二进制转换、比较、解析、字节转换、表达式计算和奇偶判断。 |
| `vrand` | `github.com/imajinyun/go-knifer/vrand` | 随机工具：整数、浮点、布尔、字节、字符串、数字字符串和随机元素。 |
| `vid` | `github.com/imajinyun/go-knifer/vid` | ID 工具：random/simple/fast UUID、MongoDB 风格 ObjectId、Snowflake 生成器与单例 next-id、worker/datacenter id 推导和 NanoId。 |
| `vhash` | `github.com/imajinyun/go-knifer/vhash` | Hash 工具：Additive、FNV、MD5、SHA-1、SHA-256 Hex。 |
| `vvalidator` | `github.com/imajinyun/go-knifer/vvalidator` | 校验工具：邮箱、手机号、URL、IPv4、中文和数字字符串。 |
| `vtemplate` | `github.com/imajinyun/go-knifer/vtemplate` | Go html/template 渲染工具。 |
| `vregex` | `github.com/imajinyun/go-knifer/vregex` | 正则工具：匹配、分组提取、命名分组、删除、计数、索引定位、模板/函数替换和元字符转义。 |
| `vchar` | `github.com/imajinyun/go-knifer/vchar` | 字符工具：空白、字母、数字、ASCII、字母或数字判断。 |
| `vbool` | `github.com/imajinyun/go-knifer/vbool` | 布尔工具：取反、转 int、全真/任一为真判断。 |
| `vblf` | `github.com/imajinyun/go-knifer/vblf` | 布隆过滤器：bitmap/bitset/filter 抽象，以及多种字符串哈希算法。 |
| `vcache` | `github.com/imajinyun/go-knifer/vcache` | 泛型缓存：FIFO、LFU、LRU、Timed、Weak、NoCache，支持 TTL、淘汰监听与懒加载。 |
| `vcaptcha` | `github.com/imajinyun/go-knifer/vcaptcha` | 图片验证码：线条、圆圈、扭曲、GIF 验证码，支持随机/数学表达式生成器。 |
| `vcron` | `github.com/imajinyun/go-knifer/vcron` | Cron 表达式解析与任务调度，支持默认调度器和自定义调度器。 |
| `vcrypto` | `github.com/imajinyun/go-knifer/vcrypto` | 加密与摘要：MD5/SHA、HMAC、随机字节、AES-CBC/AES-GCM、RSA-OAEP、RSA PEM 编解码。 |
| `vhttp` | `github.com/imajinyun/go-knifer/vhttp` | 链式 HTTP 客户端、下载、全局 Header/Timeout、BasicAuth、User-Agent 解析、简易服务端。 |
| `vresty` | `github.com/imajinyun/go-knifer/vresty` | 基于 Resty v3 的 HTTP facade：链式请求、JSON/form/multipart 请求体、全局 Header/Timeout、下载与轻量响应工具。 |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | 有序 JSON 对象/数组、JSON 解析与格式化、路径表达式读写、Bean/List 转换、XML/JSON 转换。 |
| `vxml` | `github.com/imajinyun/go-knifer/vxml` | XML 工具：解析/读取/写出/格式化、树节点访问、简单 XPath 风格查询、转义、Map/Bean 转换和命名空间辅助。 |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT 创建、解析、签名、验签与时间字段校验，支持 HMAC、RSA、ECDSA、none 等 signer。 |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | 日志 facade：console/color console logger、日志级别、全局 logger 与静态日志函数。 |
| `verr` | `github.com/imajinyun/go-knifer/verr` | 错误工具：panic recover、错误聚合、multierror 匹配、堆栈捕获/格式化，以及可选 logrus/Sentry 集成。 |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | 分组配置读取：setting/properties 风格文本和简单 YAML 子集，支持类型化读取。 |
| `vset` | `github.com/imajinyun/go-knifer/vset` | 泛型与常用类型集合工具：支持添加、删除、包含判断、集合运算，以及 JSON/YAML 编解码辅助。 |
| `vjob` | `github.com/imajinyun/go-knifer/vjob` | 可切分任务执行：职责分离任务数据与调度配置，支持泛型 Slice/Map 适配、context 取消和串行合并回调；无需开启 generic type alias 实验。 |
| `vsem` | `github.com/imajinyun/go-knifer/vsem` | 加权计数信号量：支持 context 取消、FIFO 公平等待、非阻塞获取、关闭通知与占用数查询。 |
| `vskt` | `github.com/imajinyun/go-knifer/vskt` | TCP socket 工具：普通连接、NIO/AIO server/client、协议编解码接口。 |
| `vsys` | `github.com/imajinyun/go-knifer/vsys` | 系统与运行时信息：主机、OS、用户、Go runtime、进程内存、goroutine、环境变量等。 |

## 🧭 架构与包边界

`go-knifer` 采用 `v*` 对外 facade + `internal/*` 内部实现的结构。业务代码应优先导入
`v*` 包；`internal/*` 用于沉淀具体实现，便于后续在不暴露所有内部细节的前提下持续重构。

facade 规则：

- `internal/<domain>` 负责领域实现细节和领域内测试。
- `v<domain>` 负责暴露该领域稳定的公共 API。
- 简单工具包可以手写轻量转发；较大的模块可以保留生成的 `facade.go`。无论哪种方式，
  internal 新增导出 API 时，都应先评估是否需要进入 public facade。
- `vdes`、`vser`、`vsem`、`vskt`、`vblf`、`vver` 等短命名继续保留，通过上方模块表说明含义，
  不再通过改名破坏已有导入路径。

领域边界规则：

- `vhash` 面向通用 hash 能力，例如 Additive/FNV 和简单摘要快捷方法；`vcrypto` 面向安全相关摘要、
  HMAC、加解密、密钥和 PEM 编解码。
- `vhttp` 是基于标准库的轻量 HTTP facade；`vresty` 是基于 Resty 的链式高级 HTTP client facade。
- `vcodec` 负责 Base64、Hex、URL query escaping 等编码/解码算法；`vurl` 负责 URL/URI 解析、规范化、
  资源和协议语义。
- `vjson` 负责 JSON 对象、数组、路径和轻量 XML adapter；`vxml` 负责 XML 解析、树访问、格式化、
  namespace 和 XML 专属的 map/bean 转换。
- `vobj` 是对象级便利 facade。新增具体领域逻辑应优先落到 `vstr`、`vslice`、`vmap`、`vser`、`vref`
  等明确领域包，只有在对象级聚合有价值时再由 `vobj` 做轻量包装。

部分 `internal` 包，例如 `db`、`dfa`、`poi`，是有意保留的领域占位。它们用于说明未来能力归属，
当前不提供运行时 API。

## 🚀 Install

项目要求 Go 1.20 或更高版本。

```bash
go get github.com/imajinyun/go-knifer
```

Go 会按实际导入的子包拉取模块，例如：

```go
import (
  "github.com/imajinyun/go-knifer/vstr"
  "github.com/imajinyun/go-knifer/vhttp"
)
```

## 📝 Quick start

### 基础工具与 JSON

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vid"
  "github.com/imajinyun/go-knifer/vjson"
  "github.com/imajinyun/go-knifer/vstr"
)

func main() {
  name := vstr.DefaultIfBlank("", "go-knifer")

  obj := vjson.NewObject().
    Set("id", vid.FastUUID()).
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

### Resty v3 HTTP facade

`vresty` 是基于 `resty.dev/v3` 的轻量链式 facade，适合直接发起常见 HTTP
请求。它支持 query 参数、Header、Cookie、Basic/Bearer Auth、JSON/form 请求体、
multipart 文件上传、单请求超时、跳过 TLS 校验、重定向控制以及下载等能力；响应侧
提供状态码、Header、Cookie、Content-Type、字符串/字节正文、保存到文件等便捷方法。

```go
package main

import (
  "fmt"
  "time"

  "github.com/imajinyun/go-knifer/vresty"
)

func main() {
  vresty.SetGlobalTimeout(5 * time.Second)
  vresty.SetGlobalHeader("X-App", "go-knifer")

  resp := vresty.Post("https://api.example.com/users").
    Query("source", "demo").
    BearerAuth("token").
    BodyJSON(`{"name":"go-knifer"}`).
    Timeout(3 * time.Second).
    Execute()

  if resp.Err() != nil {
    panic(resp.Err())
  }
  if !resp.IsOK() {
    panic(fmt.Sprintf("unexpected status: %d", resp.Status()))
  }

  fmt.Println(resp.ContentType())
  fmt.Println(resp.Body())
}
```

简单请求和下载也可以使用快捷函数：

```go
body := vresty.GetString("https://example.com")
jsonBody := vresty.PostJSON("https://api.example.com/events", `{"event":"created"}`)
n, err := vresty.DownloadFile("https://example.com/report.csv", "./downloads")
_, _, _ = body, jsonBody, n
_ = err
```

### URL 与 URI 工具

`vurl` 集中提供 URL 解析、标准化、query 字符串处理、协议判断、Data URI 构造
和文件 URL 转换等能力。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vurl"
)

func main() {
  normalized := vurl.Normalize(`example.com\docs/a b`, true, true)
  completed, _ := vurl.Complete("https://example.com/base/", "next?id=1")
  query := vurl.BuildQuery(map[string]any{"lang": "go", "page": 1})
  dataURI := vurl.DataURIBase64("text/plain", "aGVsbG8=")

  fmt.Println(normalized)
  fmt.Println(completed)
  fmt.Println(query)
  fmt.Println(vurl.IsWebURL(completed))
  fmt.Println(dataURI)
}
```

### 对象工具

`vobj` 提供 nil 安全的对象辅助能力，覆盖相等性判断、空值判断、默认值、
克隆/序列化、比较和类型检查等常见数据处理场景。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vobj"
)

type Profile struct {
  Name string
  Tags []string
}

func main() {
  name := "go-knifer"
  profile := Profile{Name: name, Tags: []string{"go", "tool"}}

  cloned := vobj.CloneIfPossible(profile)
  fmt.Println(vobj.Equal(1, int64(1)))
  fmt.Println(vobj.IsEmpty([]string{}))
  fmt.Println(vobj.DefaultIfNil(&name, "default"))
  fmt.Println(vobj.Contains(cloned.Tags, "go"))
  fmt.Println(vobj.TypeName(profile))
}
```

### 序列化工具

`vser` 提供基于 gob 的序列化辅助能力，覆盖字节编码、泛型反序列化、
深拷贝、接口类型注册，以及可选的解码对象图类型校验。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vser"
)

type Profile struct {
  Name string
  Tags []string
}

func main() {
  profile := Profile{Name: "go-knifer", Tags: []string{"go", "tool"}}

  data, _ := vser.Serialize(profile)
  decoded, _ := vser.DeserializeTo[Profile](data, Profile{})
  cloned := vser.CloneIfPossible(profile)

  fmt.Println(decoded.Name)
  fmt.Println(cloned.Tags)
}
```

### 版本工具

`vver` 提供版本号比较与表达式匹配能力。表达式支持比较符（`>`、`>=`、
`<`、`<=`、`≥`、`≤`）、`1.0.0-1.5.0` 这样的闭区间、`1.0.0-` 这样的
开放区间，以及使用自定义分隔符的多表达式匹配。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vver"
)

func main() {
  fmt.Println(vver.CompareVersion("1.0.0", "1.0.2"))
  fmt.Println(vver.IsGreaterThan("1.13.0", "1.12.1c"))
  fmt.Println(vver.MatchEl("1.0.2", ">=1.0.0;1.2.0"))
  fmt.Println(vver.MatchElWithDelimiter("1.0.2", "<1.0.1,1.0.2-1.1.1", ","))
}
```

### ZIP、gzip 与 zlib 工具

`vzip` 提供压缩包创建/解压、条目读取、遍历、追加、内存条目写入，
以及 byte/string 级别的 gzip 和 zlib 压缩解压能力。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vzip"
)

func main() {
  _ = vzip.ZipEntries("demo.zip", vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
  data, _ := vzip.GetBytes("demo.zip", "hello.txt")
  gz, _ := vzip.GzipString(string(data))
  text, _ := vzip.UnGzipString(gz)

  fmt.Println(text)
}
```

### 脱敏工具

`vdes` 提供常见敏感字段的内置遮罩规则，例如姓名、证件号、电话、地址、
邮箱、密码、车牌、银行卡、IP 地址、护照号和信用代码。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vdes"
)

func main() {
  fmt.Println(vdes.MobilePhone("18049531999"))
  fmt.Println(vdes.Email("duandazhi-jack@gmail.com.cn"))
  fmt.Println(vdes.BankCard("11011111222233333256"))
  fmt.Println(vdes.Desensitized("PJ1234567", vdes.PassportType))
}
```

### 正则工具

`vregex` 提供安全的正则辅助能力，覆盖全量匹配、子串查找、捕获分组、
命名分组、删除、计数、索引定位、模板/函数替换，以及正则元字符转义。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vregex"
)

func main() {
  text := "date=2026-05-31; score=100"

  fmt.Println(vregex.GetByName(`(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})`, text, "year"))
  fmt.Println(vregex.ExtractMulti(`score=(\d+)`, text, "score:$1"))
  fmt.Println(vregex.DelFirst(`\d+`, text))
  fmt.Println(vregex.ReplaceAllFunc(text, `\d+`, func(m vregex.MatchResult) string {
    return "[" + m.Text + "]"
  }))
  fmt.Println(vregex.Escape("a+b(c)"))
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

### 泛型集合

`vset` 提供泛型 `Set[T]` 和常用类型构造函数。对外的泛型 facade 使用普通泛型
类型实现，而不是 generic type alias，因此默认 Go 工具链和 `go vet` 下无需开启
`GOEXPERIMENT=aliastypeparams`。

```go
package main

import (
  "encoding/json"
  "fmt"

  "github.com/imajinyun/go-knifer/vset"
)

func main() {
  tags := vset.New("go", "tool")
  tags.Add("sdk")

  other := vset.New("tool", "cli")
  fmt.Println(tags.Contains("go"))
  fmt.Println(tags.Union(other).Members())
  fmt.Println(tags.Intersect(other).Members())

  data, _ := json.Marshal(tags)
  var decoded vset.Set[string]
  _ = json.Unmarshal(data, &decoded)
  fmt.Println(decoded.Equal(tags))

  ids := vset.NewInt(1, 2, 3)
  ids.Remove(2)
  fmt.Println(ids.Members())
}
```

### 可切分任务执行

`vjob` 将任务接口和调度配置拆开：任务只需要实现 `Len` 和按区间执行的
`Run`，`Options` 负责控制分片大小和最大并发数。`Options` 零值合法：
`Run` 默认把整个任务作为一个分片串行执行；需要指定批大小或并发度时使用
`RunWith`。每个分片返回的 `Merge` 会在分片执行成功后按顺序串行回放，适合
worker 并发构造局部结果，再安全地合并到共享结果中。`Batch[T]` 是对内部实现的
facade 包装类型，不是 generic type alias，因此 `go vet` 不需要额外实验开关。

```go
package main

import (
  "context"
  "fmt"
  "sync"

  "github.com/imajinyun/go-knifer/vjob"
)

func main() {
  values := []int{1, 2, 3, 4}
  var (
    mu  sync.Mutex
    sum int
  )

  job := vjob.NewBatch(func(ctx context.Context, batch []int) (vjob.Merge, error) {
    local := 0
    for _, v := range batch {
      local += v
    }
    return func() error {
      mu.Lock()
      defer mu.Unlock()
      sum += local
      return nil
    }, nil
  }, values)

  if err := vjob.RunWith(context.Background(), job, vjob.Options{BatchSize: 2, MaxConcurrency: 2}); err != nil {
    panic(err)
  }
  fmt.Println(sum)
}
```

长期复用的业务任务也可以直接内嵌 `vjob.Options`，由任务自身携带默认调度配置：

```go
type UserImportJob struct {
  vjob.Options
  users []User
}

func (j *UserImportJob) Len() int { return len(j.users) }

func (j *UserImportJob) Run(ctx context.Context, start, end int) (vjob.Merge, error) {
  batch := j.users[start:end]
  return func() error {
    return saveUsers(batch)
  }, nil
}

err := vjob.RunWith(ctx, job, job.Options)
```

### 错误恢复与堆栈工具

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/verr"
)

func main() {
  err := verr.Recover(func() error {
    panic("boom")
  }, "run risky job")
  if err != nil {
    fmt.Println(err)
    fmt.Println(verr.GetStack(err))
  }

  collector := verr.NewCollector()
  collector.GoRun(func() error { return fmt.Errorf("task failed") }, "async task")
  if err := collector.Error(); err != nil {
    fmt.Println(err)
  }
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
