# go-knifer

> 🍬 一组让 Go 开发保持锋利的工具。

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-00ADD8?logo=go)](https://go.dev/)

## 📚 简介

`go-knifer` 是一个面向 Go 项目的常用工具集合：把项目里反复出现的字符串处理、集合操作、编解码、加密、HTTP、JSON、缓存、定时任务、JWT、日志、配置、系统信息等能力沉淀成可复用的工具包。

项目根包 `github.com/imajinyun/go-knifer` 仅作为模块入口说明使用；实际能力按领域拆分到多个 `v*` 对外子包中，用户可以按需导入，避免把无关 API 混入业务代码。

## 🔪 `go-knifer` 名称来源

`knifer` 来自 “knife”：像一把随手可用的小刀，解决日常 Go 开发里的高频小问题。它不试图替代标准库，而是对标准库与常见工程实践做轻量封装，让代码更短、更统一、更容易维护。

## ✨ go-knifer 如何改变编码方式

以前，计算一个 MD5 往往需要在业务代码里重复写样板逻辑：

```go
sum := md5.Sum([]byte("hello"))
text := hex.EncodeToString(sum[:])
```

现在，使用 `go-knifer` 可以直接调用工具方法：

```go
text := vcrypto.MD5Hex("hello")
```

这类封装能减少重复代码、降低复制粘贴带来的隐患，也让团队内相同场景使用一致的 API。

## 🧭 按场景查找

不确定该引入哪个包？从你要做的事出发：

| 我想…… | 使用 |
| --- | --- |
| 裁剪、切分、命名转换、判空字符串 | `vstr` |
| 对切片做过滤 / 映射 / 去重 / 分页 | `vslice` |
| 创建、查询、转换、合并、差集或排序 map | `vmap` |
| 把 `any` 宽松转成 int/float/bool/string | `vconv` |
| 精确运算、舍入、表达式计算 | `vnum` |
| MD5/SHA/HMAC、AES/RSA、参数签名 | `vcrypto` |
| 非加密哈希（FNV、BKDR 等） | `vhash` |
| URL 编解码、query 构建/解析 | `vurl` |
| Base64 / Hex 编解码 | `vcodec` |
| 构建/解析 JSON、路径读写、JSON↔XML | `vjson` |
| 解析、构建、遍历 XML | `vxml` |
| 生成 UUID / Snowflake / NanoId | `vid` |
| 校验或解析身份证号 | `vident` |
| 读写文件、路径、复制、建目录 | `vfile` |
| 日期格式化/解析、偏移、天数区间 | `vdate` |
| 发起 HTTP 请求（标准库） | `vhttp` |
| 发起 HTTP 请求（基于 Resty） | `vresty` |
| 校验邮箱/手机号/IP 等 | `vvalid` |
| 敏感数据脱敏 | `vmask` |
| JWT 签发/校验 | `vjwt` |
| 定时任务调度 | `vcron` |
| FIFO/LRU/LFU/TTL 缓存 | `vcache` |

完整清单见下方模块矩阵。

## 🧩 模块

当前项目采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

| 模块 | 导入路径 | 功能说明 |
| --- | --- | --- |
| `vstr` | `github.com/imajinyun/go-knifer/vstr` | 字符串工具：空白判断、裁剪、切分、截取、格式化、provider-backed emoji、命名转换、默认值、HTML 转义，以及字符判断（空白、字母、数字、ASCII、字母或数字）。 |
| `vslice` | `github.com/imajinyun/go-knifer/vslice` | Slice 工具：包含/索引、反转、去重、拼接、过滤/映射、截取、合并、集合操作和分页。 |
| `vmap` | `github.com/imajinyun/go-knifer/vmap` | Map 工具：构造、空判断、contains/get/find、keys/values 与排序视图、map/filter/reject/partition、reduce/group/count、反转、合并/自定义冲突合并、交集/差集/对称差集、pick/omit、update/clone 和相等性判断。 |
| `vconv` | `github.com/imajinyun/go-knifer/vconv` | 宽松类型转换：string、int、int64、float64、bool、bytes 及默认值版本。 |
| `vdate` | `github.com/imajinyun/go-knifer/vdate` | 日期时间工具：常用布局、解析/格式化、日/月/年起止、偏移和比较。 |
| `vfile` | `github.com/imajinyun/go-knifer/vfile` | 文件与 IO 工具：读写复制、按行读取、mkdir/touch/delete、文件名处理、静默关闭和 provider-backed 文件系统操作。 |
| `vcodec` | `github.com/imajinyun/go-knifer/vcodec` | 编解码工具：Base64、URL-safe Base64、raw URL-safe Base64、自定义 Base64 encoding provider 和 Hex。 |
| `vurl` | `github.com/imajinyun/go-knifer/vurl` | URL 与 URI 工具：解析、标准化、相对 URL 补全、query 编解码、URL/路径/fragment 百分号编码、URL 构造、Data URI 构造、协议判断和文件 URL 转换。 |
| `vnet` | `github.com/imajinyun/go-knifer/vnet` | 网络工具：IPv4/IPv6 转换、CIDR/范围/掩码、本地端口、主机/网卡/MAC 查询、TLS 配置、address/dial/ping provider options 和 multipart 表单辅助。 |
| `vobj` | `github.com/imajinyun/go-knifer/vobj` | 对象工具：nil/空值判断、相等性、默认值、克隆/序列化、比较、类型检查和容器辅助。 |
| `vver` | `github.com/imajinyun/go-knifer/vver` | 版本工具：版本号比较、大小关系判断、表达式匹配、闭区间范围和自定义多表达式分隔符。 |
| `vref` | `github.com/imajinyun/go-knifer/vref` | 反射工具：字段查找与赋值、方法发现与调用、构造函数风格调用、类型/值工具、方法分类判断，以及显式 unsafe/unexported 字段访问选项。 |
| `vbean` | `github.com/imajinyun/go-knifer/vbean` | Bean/结构体映射工具：struct/map 互转、copy properties、tag/alias 匹配、忽略空值/零值选项和弱类型转换。 |
| `vzip` | `github.com/imajinyun/go-knifer/vzip` | ZIP、gzip、zlib 工具：压缩包创建/解压、条目读取、遍历、追加、内存条目、流式压缩和 provider-backed 归档文件操作。 |
| `vpoi` | `github.com/imajinyun/go-knifer/vpoi` | Office 文档工具：轻量 Excel XLSX 工作表列表、行读写、多工作表写入、内存工作簿创建，以及可注入的 workbook/文件系统 provider。 |
| `vmask` | `github.com/imajinyun/go-knifer/vmask` | 脱敏工具：姓名、证件号、电话、地址、邮箱、密码、车牌、银行卡、IP、护照号和信用代码遮罩。 |
| `vnum` | `github.com/imajinyun/go-knifer/vnum` | 数字工具：精确加减乘除、舍入模式、格式化、数字判断、不重复随机数、range、阶乘/组合数、最大公约数/最小公倍数、二进制转换、比较、解析、字节转换、表达式计算和奇偶判断。 |
| `vrand` | `github.com/imajinyun/go-knifer/vrand` | 随机工具：整数、浮点、布尔、字节、字符串、数字字符串、随机元素、确定性 seed，以及可重置的包级伪随机源 provider。 |
| `vid` | `github.com/imajinyun/go-knifer/vid` | ID 工具：random/simple/fast UUID、MongoDB 风格 ObjectId、Snowflake 生成器与单例 next-id、worker/datacenter id 推导、NanoId、fallback random source、isolated Snowflake 创建，以及可重置 fallback PRNG provider/seed。 |
| `vident` | `github.com/imajinyun/go-knifer/vident` | 身份标识工具：中国大陆身份证 15/18 位转换、合法性校验、校验码、可配置解析选项的生日/年龄/性别提取、省市区编码解析、遮罩，以及港澳台证件校验。 |
| `vhash` | `github.com/imajinyun/go-knifer/vhash` | 非加密 Hash 工具：Additive、FNV、可注入 32-bit hash provider，以及一组经典字符串哈希（RS、JS、PJW、ELF、BKDR、SDBM、DJB、AP、HF、HFIP、TianL、Java 默认）。 |
| `vvalid` | `github.com/imajinyun/go-knifer/vvalid` | 校验工具：邮箱、手机号、URL、IPv4/IPv6、身份证、中文和数字字符串，并支持规则敏感校验的 per-call matcher provider。 |
| `vtpl` | `github.com/imajinyun/go-knifer/vtpl` | Go html/template 渲染工具，支持单次调用配置模板名、FuncMap、分隔符、template factory 和 executor。 |
| `vregex` | `github.com/imajinyun/go-knifer/vregex` | 正则工具：匹配、分组提取、命名分组、删除、计数、索引定位、模板/函数替换、元字符转义，以及单次调用 compiler / DOTALL options。 |
| `vbool` | `github.com/imajinyun/go-knifer/vbool` | 布尔工具：取反、转 int、全真/任一为真判断。 |
| `vblf` | `github.com/imajinyun/go-knifer/vblf` | 布隆过滤器：bitmap/bitset/filter 抽象、多种字符串哈希算法、option-based 构造器，以及 provider-backed 文件初始化。 |
| `vcache` | `github.com/imajinyun/go-knifer/vcache` | 泛型缓存：FIFO、LFU、LRU、Timed、Weak、NoCache，支持 TTL、clock、淘汰监听、懒加载、ticker/runner provider 和 weak-cache finalizer provider。 |
| `vcaptcha` | `github.com/imajinyun/go-knifer/vcaptcha` | 图片验证码：线条、圆圈、扭曲、GIF 验证码，支持随机/数学表达式生成器。 |
| `vcron` | `github.com/imajinyun/go-knifer/vcron` | Cron 表达式解析与任务调度，支持默认/自定义调度器、可配置 cron options、ID random-reader/clock/sleeper/runner provider，以及单次调用隔离的默认调度器覆盖。 |
| `vcrypto` | `github.com/imajinyun/go-knifer/vcrypto` | 加密与摘要：MD5/SHA、provider-backed digest、HMAC、PBKDF2、参数签名、随机字节、支持 block-factory options 的 AES CBC/ECB/CTR/CFB/OFB/GCM、DES/3DES、RC4、Vigenere、XXTEA、RSA OAEP/PKCS#1/PSS 与可配置数据签名、PEM 与 X.509 证书工具。 |
| `vdb` | `github.com/imajinyun/go-knifer/vdb` | 基于 database/sql 的数据库工具：SQL 执行、命名参数、Entity、条件、查询构造器、事务、分页、轻量元信息查询和可注入的 `sql.Open` provider。 |
| `vdfa` | `github.com/imajinyun/go-knifer/vdfa` | DFA 词树匹配：停顿字符过滤、首个/全部匹配、密集/贪婪匹配、命中词位置元信息、包级匹配器、隔离 matcher options、`Any` 辅助函数的 JSON marshal/unmarshal provider、文本替换，以及用于包级异步初始化的可重置 async runner provider。 |
| `vhttp` | `github.com/imajinyun/go-knifer/vhttp` | 链式 HTTP 客户端、隔离/global-config 请求构建、create/get/post `WithOptions` 辅助函数、provider-backed transport/request factory/multipart writer/download save、BasicAuth、User-Agent 解析、provider-backed HTML 清理/标签过滤、可重置 transport/server starters、异步服务端 runner option 和简易服务端辅助函数。 |
| `vresty` | `github.com/imajinyun/go-knifer/vresty` | 基于 Resty v3 的 HTTP facade：链式请求、JSON/form/multipart 请求体、隔离/global-config 请求构建、create/get/post `WithOptions` 辅助函数、单次请求 client factory、可重置默认 Resty client provider、下载与轻量响应工具。 |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | 有序 JSON 对象/数组、JSON 解析与格式化、路径表达式读写、provider-backed marshal/unmarshal、可配置 Object/Array/Bean/List 转换，以及带 parser/writer options 的 XML/JSON 转换。 |
| `vxml` | `github.com/imajinyun/go-knifer/vxml` | XML 工具：解析/读取/写出/格式化、树节点访问、简单 XPath 风格查询、转义、支持 parser/codec options 的 Map/Bean 转换、transform options 和命名空间辅助。 |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT 创建、解析、签名、验签与时间字段校验，支持 HMAC、RSA、ECDSA、none 等 signer，以及 provider-backed JSON marshal/unmarshal options。 |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | 日志 facade：console/color console logger、可注入颜色工厂、日志级别、全局 logger、静态日志函数、单次调用 logger options 和 isolated logger 创建。 |
| `verr` | `github.com/imajinyun/go-knifer/verr` | 错误工具：panic recover、错误聚合、multierror 匹配、collector 构造 options、堆栈捕获/格式化、可重置 log/stack cache、可注入的 logging/stack/exit/timer/runner provider、隔离 logrus 创建，以及可选 logrus/Sentry 集成。 |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | 分组配置读取：setting/properties 风格文本和简单 YAML 子集，支持类型化读取、profile/remote/file 加载 options、环境变量展开 provider 和 watch ticker/runner provider。 |
| `vset` | `github.com/imajinyun/go-knifer/vset` | 泛型与常用类型集合工具：支持添加、删除、包含判断、集合运算，以及 JSON/YAML 编解码辅助。 |
| `vjob` | `github.com/imajinyun/go-knifer/vjob` | 可切分任务执行：职责分离任务数据与调度配置，支持泛型 Slice/Map 适配、context 取消和串行合并回调；无需开启 generic type alias 实验。 |
| `vsem` | `github.com/imajinyun/go-knifer/vsem` | 加权计数信号量：支持 context 取消、FIFO 公平等待、非阻塞获取、关闭通知与占用数查询。 |
| `vskt` | `github.com/imajinyun/go-knifer/vskt` | TCP socket 工具：普通连接、NIO/AIO server/client、协议编解码接口，以及可配置 thread-pool/listener/connection/runner provider。 |
| `vsys` | `github.com/imajinyun/go-knifer/vsys` | 系统与运行时信息：主机、OS、用户、Go runtime、进程内存、goroutine、环境变量、可重置信息缓存，以及可注入的 env/command/runtime provider。 |

## 🧭 架构与包边界

`go-knifer` 采用 `v*` 对外 facade + `internal/*` 内部实现的结构。业务代码应优先导入
`v*` 包；`internal/*` 用于沉淀具体实现，便于后续在不暴露所有内部细节的前提下持续重构。

facade 规则：

- `internal/<domain>` 负责领域实现细节和领域内测试。
- `v<domain>` 负责暴露该领域稳定的公共 API。
- 简单工具包可以手写轻量转发；较大的模块可以保留生成的 `facade.go`。无论哪种方式，
  internal 新增导出 API 时，都应先评估是否需要进入 public facade。
- `vvalid`、`vmask`、`vsem`、`vskt`、`vblf`、`vver` 等短命名继续保留，通过上方模块表说明含义，
  不再通过改名破坏已有导入路径。

可配置 API 与 Provider 注入：

- 多个子包通过 `WithXxx` helper 与 `XxxWithOptions` 变体暴露 Functional Options 模式。既有固定参数 API
  保持稳定，option 变体用于为需要高级控制的调用方提供扩展能力。
- 该模式已覆盖布隆过滤器、缓存、验证码、配置加载/监听、定时任务、加密、数据库、日期时间、DFA、错误处理、
  文件、HTTP/Resty、ID、身份解析、JSON/JWT、日志、网络、数字、POI、随机数、socket、系统信息、URL、XML、ZIP 等运行时敏感能力。
- Provider 风格的 option 允许调用方注入文件系统函数、网络/TLS dialer 或 reader、HTTP request/multipart factory、
  clock、timer/ticker、随机源、DB opener、Excel workbook factory、logger、stack capture、finalizer、环境变量查询、
  command executor、Sentry/logrus hook 等进程全局依赖，便于确定性测试与受控运行时行为。
- 包级默认值必须显式治理。例如 HTTP 全局默认值可通过 `vhttp.SnapshotGlobalConfig` 读取不可变快照；
  `vhttp.NewIsolatedRequest` 可以在不读取包级默认值的情况下构建请求；单次调用 option 不应隐式修改隐藏的全局状态。

Provider 覆盖重点：

| 领域 | 示例 |
| --- | --- |
| HTTP / Resty | `vhttp.NewIsolatedRequest`、`vhttp.NewRequestWithConfig`、`vhttp.CreateGetWithOptions`、`vhttp.CreatePostWithOptions`、`vhttp.WithTransportProvider`、`vhttp.WithRequestFactory`、`vhttp.WithMultipartWriterFactory`、`vhttp.ResetDefaultTransport`、`vhttp.WithListenAndServeFunc`、`vhttp.WithAsyncRunner`、`vhttp.CreateServerWithOptions`、`vhttp.ResetServerStarters`、`vhttp.GetWithTimeoutWithOptions`、`vhttp.GetWithParamsWithOptions`、`vhttp.PostStringWithOptions`、`vhttp.CleanHTMLWithOptions`、`vhttp.FilterHTMLTagWithOptions`、`vhttp.WithHTMLFilterCompileFunc`、`vresty.NewIsolatedRequest`、`vresty.WithGlobalConfig`、`vresty.WithRestyClientFactory`、`vresty.ConfigureDefaultRestyClientProvider`、`vresty.ResetDefaultRestyClientProvider`、`vresty.CreateRequestWithOptions`、`vresty.CreateGetWithOptions`、`vresty.CreatePostWithOptions`、`vresty.GetWithTimeoutWithOptions`、`vresty.GetWithParamsWithOptions`、`vresty.PostStringWithOptions`、`vresty.DownloadFileWithOptions` |
| 文件 / 配置 / 压缩 / POI | `vfile` provider options、`vconf.LoadWithOptions`、`vconf.WatchWithOptions`、`vconf.WatchOptions.Runner`、`vzip` provider options、`vpoi.WithOpenFileFunc`、`vpoi.WithNewFileFunc`、`vpoi.WithSaveAsFunc` |
| Cron / DFA / ID / 身份 / 随机数 | `vcron.WithDefaultSchedulerOptions`、`vcron.NewConfigWithOptions`、`vcron.WithIDRandomReader`、`vcron.WithRunner`、`vcron.CronScheduleWithOptions`、`vdfa.WithMatcherWords`、`vdfa.WithJSONMarshal`、`vdfa.WithJSONUnmarshal`、`vdfa.ContainsWithOptions`、`vdfa.ConfigureAsyncRunner`、`vdfa.ResetAsyncRunner`、`vid.NewIsolatedSnowflake`、`vid.CreateSnowflakeWithOptions`、`vid.WithSnowflakeCache`、`vid.WithFallbackRandomSource`、`vid.ConfigureDefaultFallbackRandomSourceProvider`、`vid.ResetDefaultFallbackRandomSource`、`vid.SetFallbackRandomSeed`、`vrand.ConfigureDefaultRandomSourceProvider`、`vrand.ResetDefaultRandomSource`、`vrand.SetSeed`、`vident.BirthDateWithOptions` |
| 编解码 / JSON / XML / JWT / hash | `vcodec.Base64EncodeWithEncoding`、`vcodec.Base64DecodeWithEncoding`、`vcodec.Base64RawURLEncode`、`vcodec.Base64RawURLDecode`、`vhash.Hash32`、`vjson.WithMarshalFunc`、`vjson.WithUnmarshalFunc`、`vjson.WithParseUnmarshalFunc`、`vjson.WithBeanUnmarshalFunc`、`vjson.ParseObjWithOptions`、`vjson.ParseArrayWithOptions`、`vjson.ToBeanWithOptions`、`vjson.ToListWithOptions`、`vjson.XMLToJSONWithOptions`、`vjson.ToXMLWithOptions`、`vxml.XMLToMapWithOptions`、`vxml.XMLToBeanWithOptions`、`vxml.XMLNodeToBeanWithOptions`、`vxml.TransformWithOptions`、`vxml.FormatWithOptions`、`vjwt.WithJSONMarshalFunc`、`vjwt.WithJSONUnmarshalFunc`、`vjwt.ParseTokenWithOptions`、`vjwt.WithTokenJSONOptions` |
| 加密 / 模板 / 正则 / 校验 / 字符串 | `vcrypto.Digest`、`vcrypto.DigestHex`、`vcrypto.WithAESBlockFactory`、`vcrypto.WithGCMBlockFactory`、`vcrypto.AESEncryptCBCWithOptions`、`vcrypto.AESEncryptGCMWithOptions`、`vcrypto.SignWithRSAOptions`、`vcrypto.VerifyWithRSAOptions`、`vtpl.RenderWithOptions`、`vtpl.WithFuncMap`、`vtpl.WithTemplateFactory`、`vregex.WithCompileFunc`、`vregex.WithDotAll`、`vregex.MatchWithOptions`、`vregex.ReplaceAllFuncWithOptions`、`vvalid.IsEmailWithOptions`、`vvalid.WithMobileMatcher`、`vstr.ContainsEmojiWithOptions`、`vstr.RemoveEmojiWithOptions` |
| DB / 网络 / 系统 / 反射 / socket | `vdb.WithSQLOpenFunc`、`vnet.WithConnectDialer`、`vnet.WithPingDialer`、`vnet.WithAddressNetwork`、`vnet.WithTCPAddrResolver`、`vnet.WithUploadOpenSource`、`vsys.WithGoEnvOutputFunc`、`vsys.WithGoRootEnvLookupFunc`、`vsys.WithOSEnvLookupFunc`、`vsys.WithEnvLookupFunc`、`vsys.ResetInfoCache`、`vref.WithUnsafeAccess`、`vskt.WithThreadPoolSizeFunc`、`vskt.WithRunner` |
| 错误 / 缓存 / 日志 / 运行时 | `verr.NewCollectorWithOptions`、`verr.WithCollectorLogFunc`、`verr.WithCollectorRunner`、`verr.WithCollectorContext`、`verr.WithCollectorLevel`、`verr.WithCollectorTimerFactory`、`verr.WithCollectorStackCaptureOptions`、`verr.WithLogFunc`、`verr.WithCollectorStackOptions`、`verr.WithDebugStackFunc`、`verr.WithCallersFunc`、`verr.WithFuncForPCFunc`、`verr.WithStackFrameCache`、`verr.ResetStackFrameCache`、`verr.ResetDefaultLogFunc`、`verr.NewIsolatedLogrusWithOptions`、`verr.MustExitWithOptions`、`vcache.WithClock`、`vcache.WithTickerFactory`、`vcache.WithRunner`、`vcache.WithWeakFinalizerFunc`、`vcache.WithWeakFinalizerEnabled`、`vlog.WithLogColorFactory`、`vlog.NewIsolatedLogger`、`vlog.LoggerWithOptions`、`vlog.InfoWithOptions` |

领域边界规则：

- `vhash` 面向非加密 hash 能力，例如 Additive/FNV（分桶、布隆过滤器等场景）；`vcrypto` 独占所有
  安全相关摘要（MD5/SHA 系列）、HMAC、加解密、密钥和 PEM 编解码。
- `vhttp` 是基于标准库的轻量 HTTP facade；`vresty` 是基于 Resty 的链式高级 HTTP client facade。两者都不再重复暴露 URL 工具：URL 转义、query 构建/解析、协议判断（`IsHTTP`/`IsHTTPS`、`EncodeQueryMap`、`DecodeQuery` 等）统一归 `vurl`。
- `vdb` 负责基于 `database/sql` 的 SQL 数据库辅助能力；调用方继续通过 `*sql.DB` 和单次调用 options
  控制驱动和连接池。
- `vdfa` 负责 DFA 词树匹配、停顿字符过滤、密集/贪婪匹配、命中词位置元信息和文本替换；通用字符串工具不承载词典匹配逻辑。
- `vid` 负责 UUID、Snowflake、ObjectId、NanoId 等生成型标识；`vident` 负责法定身份号码与地区证件解析，
  例如中国大陆身份证和港澳台证件号。
- `vcodec` 负责 Base64、Hex 等编码/解码算法；`vurl` 负责 URL 转义、URL/URI 解析、规范化、
  资源和协议语义。
- `vjson` 负责 JSON 对象、数组、路径和轻量 XML adapter；`vxml` 负责 XML 解析、树访问、格式化、
  namespace 和 XML 专属的 map/bean 转换。
- `vbean` 负责直接的 struct/map 属性映射、copy properties、tag/alias 匹配和弱类型转换，
  不通过 JSON 序列化绕路。
- `vobj` 是对象级便利 facade。新增具体领域逻辑应优先落到 `vstr`、`vslice`、`vmap`、`vref`
  等明确领域包，只有在对象级聚合有价值时再由 `vobj` 做轻量包装。

数据库工具归属 `internal/db`，并通过 `vdb` 对外暴露；DFA 文本匹配归属 `internal/dfa`，并通过
`vdfa` 对外暴露；Office 文档工具归属 `internal/poi`，并通过 `vpoi` 对外暴露。跨领域通用输入校验归属
`internal/validator`，并通过 `vvalid` 对外暴露；领域内解析和更丰富的操作仍留在 `vident`、`vnet`、`vurl` 等领域包。

### 错误契约

根包 `knifer` 负责跨子包的统一错误契约：错误码分类 `ErrCode`（`knifer.ErrCodeInvalidInput`、
`ErrCodeNotFound`、`ErrCodeTimeout` 等）、统一的 `knifer.Error` 类型、`CodeCarrier` 接口、
`CodeOf` 提取函数，以及 `NewError` / `WrapError` / `Errorf` 构造函数。接入的子包可以返回
`*knifer.Error`，也可以在既有 error 类型/哨兵上增加按错误码匹配能力，调用方既能按错误码匹配或提取，
又能保留错误链：

```go
if errors.Is(err, knifer.ErrCodeInvalidInput) { /* ... */ }
if code, ok := knifer.CodeOf(err); ok { /* ... */ }
```

`vcrypto` 是参考接入示范：校验错误同时匹配 `knifer.ErrCodeInvalidInput` 与既有的
`vcrypto.ErrInvalidKey` / `ErrInvalidIV` / `ErrInvalidCipherText` 哨兵。

`vjwt`、`vjson`、`vcron`、`vjob`、`vpoi`、`vcodec`、`vdate`、`vbean`、`vsem`、`verr`、
`vhttp`/`vresty` 也已接入：其错误分别匹配 `knifer.ErrCodeInvalidInput`（vjwt、vjson、
vcron、vjob、vpoi 空 sheet 名、vcodec 解码失败、vdate 解析失败、vbean 映射/转换失败、vsem 非法权重）、
`knifer.ErrCodeNotFound`（vpoi 无 sheet、vblf 初始化文件不存在）、`knifer.ErrCodeUnsupported`（vsem 已关闭）与
`knifer.ErrCodeInternal`（vhttp/vresty、vskt、vblf 读取失败、verr recover 到的 panic），
同时保留各自的 error 类型、哨兵与 cause 错误链。

## 🚀 安装

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

## 📝 快速开始

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

### 校验工具

`vvalid` 提供常用输入校验的短 public 入口，把高频布尔校验集中到一个包中，具体领域能力仍委托给对应的内部实现。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vvalid"
)

func main() {
  fmt.Println(vvalid.IsEmail("a@b.com"))
  fmt.Println(vvalid.IsMobile("13812345678"))
  fmt.Println(vvalid.IsURL("https://example.com"))
  fmt.Println(vvalid.IsIPv4("127.0.0.1"))
  fmt.Println(vvalid.IsIPv6("2001:db8::1"))
  fmt.Println(vvalid.IsIDCard("11010519491231002X"))
  fmt.Println(vvalid.IsChinese("你好"))
  fmt.Println(vvalid.IsNumberStr("-3.14"))
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
  resp := vhttp.Get("https://example.com",
    vhttp.WithTimeout(3*time.Second),
    vhttp.WithHeader("X-Client", "go-knifer"),
    vhttp.WithFollowRedirects(true),
  ).
    Query("lang", "go").
    Execute()

  if resp.Err() != nil {
    panic(resp.Err())
  }

  fmt.Println(resp.Status())
  fmt.Println(resp.ContentType())
  fmt.Println(resp.Body())
}
```

如果确实需要统一默认值，仍然可以使用全局配置；但新代码更推荐使用单次请求
options，让每个请求的超时、Header、重定向、TLS、Cookie、User-Agent 等行为在调用点
显式声明，避免全局状态影响其他请求。可用 options 包括 `WithTimeout`、`WithHeader`、
`WithHeaders`、`WithFollowRedirects`、`WithMaxRedirects`、`WithSkipTLSVerify`、
`WithTransport`、`WithClient`、`WithCookieJar` 和 `WithUserAgent`。

### Resty v3 HTTP facade

`vresty` 是基于 `resty.dev/v3` 的轻量链式 facade，适合直接发起常见 HTTP
请求。它支持 query 参数、Header、Cookie、Basic/Bearer Auth、JSON/form 请求体、
multipart 文件上传、单次请求 options、跳过 TLS 校验、重定向控制以及下载等能力；响应侧
提供状态码、Header、Cookie、Content-Type、字符串/字节正文、保存到文件等便捷方法。

```go
package main

import (
  "fmt"
  "time"

  "github.com/imajinyun/go-knifer/vresty"
)

func main() {
  resp := vresty.Post("https://api.example.com/users",
    vresty.WithTimeout(3*time.Second),
    vresty.WithHeader("X-App", "go-knifer"),
    vresty.WithUserAgent("go-knifer-demo/1.0"),
  ).
    Query("source", "demo").
    BearerAuth("token").
    BodyJSON(`{"name":"go-knifer"}`).
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

`vresty` 同样支持构造请求时传入 options，从而让每次调用独立覆盖默认行为：
`WithTimeout`、`WithHeader`、`WithHeaders`、`WithFollowRedirects`、`WithMaxRedirects`、
`WithSkipTLSVerify`、`WithRestyClient`、`WithUserAgent` 和 `WithCookieDisabled`。

简单请求和下载也可以使用快捷函数：

```go
body := vresty.GetString("https://example.com")
jsonBody := vresty.PostJSON("https://api.example.com/events", `{"event":"created"}`)
n, err := vresty.DownloadFile("https://example.com/report.csv", "./downloads")
_, _, _ = body, jsonBody, n
_ = err
```

### URL 与 URI 工具

`vurl` 集中提供 URL 解析、标准化、query 字符串处理、百分号编码、URL 构造、
协议判断、Data URI 构造和文件 URL 转换等能力。

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
  built := vurl.NewHTTPURLBuilder("example.com").AddPathSegment("a b").AddQuery("q", "go").Build()

  fmt.Println(normalized)
  fmt.Println(completed)
  fmt.Println(query)
  fmt.Println(vurl.IsWebURL(completed))
  fmt.Println(dataURI)
  fmt.Println(built)
}
```

### 网络与 IP 工具

`vnet` 提供网络辅助能力，覆盖 IPv4/IPv6 转换、CIDR 与掩码计算、IP 范围展开、
本地端口探测、主机/网卡/MAC 查询、TLS client config 创建，
以及 multipart 表单辅助。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vnet"
)

func main() {
  ipLong, _ := vnet.IPv4ToLong("127.0.0.1")
  begin, _ := vnet.BeginIP("192.168.1.9", 24)
  end, _ := vnet.EndIP("192.168.1.9", 24)

  fmt.Println(ipLong, vnet.LongToIPv4(ipLong))
  fmt.Println(begin, end, vnet.IsInRange("192.168.1.8", "192.168.1.0/24"))
  fmt.Println(vnet.HideIPPart("192.168.1.8"))
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

### Map 工具

`vmap` 提供泛型 map 常用操作。构造函数和纯函数会返回非 nil map，且不会修改输入 map；只有
`Clear`、`Update` 这类显式原地操作会修改调用方传入的 map。合并场景同时支持后者覆盖前者的
`Merge`，以及自定义冲突处理的 `MergeFunc`。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vmap"
)

func main() {
  base := vmap.Of[string, int]("a", 1, "b", 2)
  merged := vmap.Merge(base, map[string]int{"b": 20, "c": 3})
  evens := vmap.FilterValues(merged, func(v int) bool { return v%2 == 0 })
  grouped := vmap.GroupBy([]string{"go", "git", "java"}, func(s string) byte { return s[0] })

  fmt.Println(vmap.SortedKeys(merged))
  fmt.Println(evens)
  fmt.Println(grouped['g'])
}
```

### 数据库工具

`vdb` 提供基于 `database/sql` 的 SQL 辅助能力：命名参数、条件构造、
Entity 写入/更新/删除、分页、事务和轻量元信息查询。驱动选择和连接池仍由调用方控制。

```go
package main

import (
  "context"
  "database/sql"
  "fmt"

  "github.com/imajinyun/go-knifer/vdb"
)

func main() {
  var raw *sql.DB // 通常由你选择的 SQL driver 打开
  db := vdb.Use(raw, vdb.WithDialect(vdb.DialectPostgres))

  sqlText, args, _ := vdb.NewBuilder(vdb.WithDialect(vdb.DialectPostgres)).
    Select("id", "name").
    From("users").
    Where(vdb.Eq("status", "active")).
    OrderBy(vdb.Desc("id")).
    Page(vdb.NewPage(1, 20)).
    SQL()

  named, _ := vdb.ParseNamed(
    "select * from users where id = :id",
    map[string]any{"id": 1},
    vdb.DialectPostgres,
  )

  _ = db
  _ = context.Background()
  fmt.Println(sqlText, args, named.SQL, named.Params)
}
```

### Bean 与结构体映射

`vbean` 用于在 struct 与 map 之间直接复制属性，不通过 JSON 序列化绕路。
它支持 tag/alias 匹配、弱类型转换，以及忽略空值/零值等单次调用 options。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vbean"
)

type UserDTO struct {
  Name  string `bean:"name,alias=full_name|displayName"`
  Age   string `bean:"age"`
  Admin string `bean:"admin"`
}

type User struct {
  Name  string `json:"full_name"`
  Age   int    `json:"age"`
  Admin bool   `json:"admin"`
}

func main() {
  src := UserDTO{Name: "alice", Age: "42", Admin: "yes"}

  var dst User
  _ = vbean.CopyProperties(src, &dst, vbean.WithIgnoreEmpty(true))

  m, _ := vbean.ToMap(dst)
  fmt.Println(dst.Age, dst.Admin)
  fmt.Println(m["full_name"])
}
```

### 序列化工具

`vobj` 提供基于 gob 的序列化辅助能力，覆盖字节编码、泛型反序列化、
深拷贝、接口类型注册，以及可选的解码对象图类型校验。

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
  profile := Profile{Name: "go-knifer", Tags: []string{"go", "tool"}}

  data, _ := vobj.Serialize(profile)
  decoded, _ := vobj.DeserializeTo[Profile](data, Profile{})
  cloned := vobj.CloneIfPossible(profile)

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

`vmask` 提供常见敏感字段的内置遮罩规则，例如姓名、证件号、电话、地址、
邮箱、密码、车牌、银行卡、IP 地址、护照号和信用代码。

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vmask"
)

func main() {
  fmt.Println(vmask.MobilePhone("18049531999"))
  fmt.Println(vmask.Email("duandazhi-jack@gmail.com.cn"))
  fmt.Println(vmask.BankCard("11011111222233333256"))
  fmt.Println(vmask.Masked("PJ1234567", vmask.PassportType))
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
  fmt.Println(vcrypto.HMACSHA256Hex([]byte("key"), []byte("hello")))

  aesKey := []byte("1234567890123456")
  iv := []byte("abcdefghijklmnop")
  cipherText, err := vcrypto.AESEncryptCBC([]byte("secret message"), aesKey, iv)
  if err != nil {
    panic(err)
  }
  plain, err := vcrypto.AESDecryptCBC(cipherText, aesKey, iv)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(plain))

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

## 📖 文档

- 根包说明：`doc.go`
- 对外 API：各 `v*` 子包的 `doc.go` 与 facade 文件
- 测试示例：各模块下的 `*_test.go`
- 在线文档：[pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)

## 📦 下载与构建

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

## 🤝 问题反馈与建议

如果发现问题或希望补充新工具，请通过 GitHub Issues 反馈。建议提供：

- Go 版本与操作系统；
- `go-knifer` 版本或 commit；
- 最小可复现代码；
- 期望行为与实际行为；
- 相关错误日志或测试输出。

## ✅ PR（Pull Request）原则

欢迎提交 PR。为了保持工具库稳定，请尽量遵循以下原则：

1. 新增能力优先放入合适的 `internal/*` 实现包，再由对应 `v*` 包暴露对外 API；
2. 新增或修改公共 API 时补充必要注释；
3. 为核心逻辑补充单元测试，提交前执行 `go test ./...`；
4. 保持代码经过 `gofmt` 格式化；
5. 避免引入不必要的第三方依赖，优先复用标准库。

## ⭐ Star go-knifer

如果这个项目减少了你的重复代码，欢迎给它一个 Star。你的反馈和贡献会帮助它成为更趁手的 Go 工具集合。
