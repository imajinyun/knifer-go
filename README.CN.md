# go-knifer

> 🍬 一组让 Go 开发保持锋利的工具。

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)

## 📚 简介

`go-knifer` 是一个面向 Go 项目的常用工具集合：把项目里反复出现的字符串处理、集合操作、编解码、加密、HTTP、JSON、缓存、定时任务、JWT、日志、配置、系统信息等能力沉淀成可复用的工具包。

项目根包 `github.com/imajinyun/go-knifer` 仅作为模块入口说明使用；实际能力按领域拆分到多个 `v*` 对外子包中，用户可以按需导入，避免把无关 API 混入业务代码。

## 🔪 `go-knifer` 名称来源

`knifer` 来自 “knife”：像一把随手可用的小刀，解决日常 Go 开发里的高频小问题。它不试图替代标准库，而是对标准库与常见工程实践做轻量封装，让代码更短、更统一、更容易维护。

## ✨ go-knifer 如何改变编码方式

`go-knifer` 把项目里高频、重复、容易复制粘贴出错的工具逻辑收束到明确的 `v*` 子包中。业务代码按领域导入需要的 facade，并通过统一 API 表达相同场景。

## 🧭 按场景查找

不确定该引入哪个包？从你要做的事出发：

| 我想使用 | 实际导入 |
| --- | --- |
| FIFO/LRU/LFU/TTL 缓存 | [`vcache`](docs/doc/04-vcache.md) |
| Base64 / Hex 编解码 | [`vcodec`](docs/doc/05-vcodec.md) |
| 加载本地或远程配置，包括带 SSRF 防护的远程配置 | [`vconf`](docs/doc/06-vconf.md) |
| 把 `any` 宽松转成 int/float/bool/string | [`vconv`](docs/doc/07-vconv.md) |
| 定时任务调度 | [`vcron`](docs/doc/08-vcron.md) |
| SHA/HMAC、AES-GCM/RSA-PSS、参数签名 | [`vcrypto`](docs/doc/09-vcrypto.md) |
| 读写 CSV records、map 或 struct | [`vcsv`](docs/doc/10-vcsv.md) |
| 日期格式化/解析、偏移、天数区间 | [`vdate`](docs/doc/11-vdate.md) |
| 读写文件、路径、复制、建目录 | [`vfile`](docs/doc/15-vfile.md) |
| 校验表单/输入数据，如邮箱、手机号、IP 等 | [`vform`](docs/doc/16-vform.md) |
| 非加密哈希（FNV、BKDR 等） | [`vhash`](docs/doc/17-vhash.md) |
| 发起 HTTP 请求（标准库） | [`vhttp`](docs/doc/18-vhttp.md) |
| 生成 UUID / Snowflake / NanoId | [`vid`](docs/doc/19-vid.md) |
| 校验或解析身份证号 | [`vident`](docs/doc/20-vident.md) |
| 生成缩略图、PNG/JPEG/GIF 格式转换、读取图像元信息、生成/解码 QRCode/Barcode 或生成图形验证码 | [`vimg`](docs/doc/21-vimg.md) |
| 构建/解析 JSON、路径读写、JSON↔XML | [`vjson`](docs/doc/23-vjson.md) |
| JWT 签发/校验 | [`vjwt`](docs/doc/24-vjwt.md) |
| 构建并发送文本/HTML 邮件、内联资源、附件或账号默认配置邮件 | [`vmail`](docs/doc/26-vmail.md) |
| 创建、查询、转换、合并、差集或排序 map | [`vmap`](docs/doc/27-vmap.md) |
| 敏感数据脱敏 | [`vmask`](docs/doc/28-vmask.md) |
| 精确运算、舍入、表达式计算 | [`vnum`](docs/doc/30-vnum.md) |
| 本地评估密码强度并分级 | [`vpass`](docs/doc/32-vpass.md) |
| 发起 HTTP 请求（基于 Resty） | [`vresty`](docs/doc/37-vresty.md) |
| 对切片做过滤 / 映射 / 去重 / 分页 | [`vslice`](docs/doc/41-vslice.md) |
| 裁剪、切分、命名转换、Unicode 转义、Ant 路径匹配、文本相似度或判空字符串 | [`vstr`](docs/doc/42-vstr.md) |
| URL 编解码、query 构建/解析，或安全打开不可信 HTTP(S) 资源 | [`vurl`](docs/doc/45-vurl.md) |
| 解析、构建、遍历 XML | [`vxml`](docs/doc/47-vxml.md) |

完整清单见下方模块矩阵。

## 🧩 模块

当前项目采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

| 模块 | 导入路径 | 功能说明 |
| --- | --- | --- |
| [`vbean`](docs/doc/01-vbean.md) | `github.com/imajinyun/go-knifer/vbean` | Bean/结构体映射工具：struct/map 互转、copy properties、tag/alias 匹配、忽略空值/零值选项和弱类型转换。 |
| [`vblf`](docs/doc/02-vblf.md) | `github.com/imajinyun/go-knifer/vblf` | 布隆过滤器：bitmap/bitset/filter 抽象、多种字符串哈希算法、option-based 构造器、返回校验错误而不是 panic 的 `E` 构造器，以及 provider-backed 文件初始化。 |
| [`vbool`](docs/doc/03-vbool.md) | `github.com/imajinyun/go-knifer/vbool` | 布尔工具：取反、转 int、全真/任一为真判断。 |
| [`vcache`](docs/doc/04-vcache.md) | `github.com/imajinyun/go-knifer/vcache` | 泛型缓存：FIFO、LFU、LRU、Timed、Weak、NoCache，支持 TTL、clock、淘汰监听、懒加载、ticker/runner provider 和 weak-cache finalizer provider。 |
| [`vcodec`](docs/doc/05-vcodec.md) | `github.com/imajinyun/go-knifer/vcodec` | 编解码工具：Base64、URL-safe Base64、raw URL-safe Base64、自定义 Base64 encoding provider 和 Hex。 |
| [`vconf`](docs/doc/06-vconf.md) | `github.com/imajinyun/go-knifer/vconf` | 分组配置读取：setting/properties 风格文本、简单 YAML 子集和 TOML 解析，支持类型化读取、schema 校验、profile/remote/file 加载 options、带 SSRF 防护的 `LoadRemoteSafe`、环境变量展开 provider、watch ticker/runner provider、读取大小限制、只读快照使用方式和深拷贝 `Clone`。 |
| [`vconv`](docs/doc/07-vconv.md) | `github.com/imajinyun/go-knifer/vconv` | 宽松类型转换：string、int、int64、float64、bool、bytes 及默认值版本。 |
| [`vcron`](docs/doc/08-vcron.md) | `github.com/imajinyun/go-knifer/vcron` | Cron 表达式解析与任务调度，支持默认/自定义调度器、可配置 cron options、ID random-reader/clock/sleeper/runner provider，以及单次调用隔离的默认调度器覆盖。 |
| [`vcrypto`](docs/doc/09-vcrypto.md) | `github.com/imajinyun/go-knifer/vcrypto` | 加密与摘要：SHA-2、provider-backed digest、HMAC、PBKDF2-SHA256、参数签名、随机字节、支持 nonce/tag/block-factory options 的 AES-GCM、RSA OAEP/PSS 与可配置数据签名、PEM 与 X.509 证书工具。 |
| [`vcsv`](docs/doc/10-vcsv.md) | `github.com/imajinyun/go-knifer/vcsv` | CSV 工具：读写分隔符、注释、字段数量、宽松引号、Trim、CRLF 等 options，支持 records/map 转换、map 写出、struct tag 导出和逐行回调。 |
| [`vdate`](docs/doc/11-vdate.md) | `github.com/imajinyun/go-knifer/vdate` | 日期时间工具：常用布局、解析/格式化、日/月/年起止、偏移和比较。 |
| [`vdb`](docs/doc/12-vdb.md) | `github.com/imajinyun/go-knifer/vdb` | 基于 database/sql 的数据库工具：SQL 执行、命名参数、Entity、条件、查询构造器、事务、分页、轻量元信息查询和可注入的 `sql.Open` provider。 |
| [`vdfa`](docs/doc/13-vdfa.md) | `github.com/imajinyun/go-knifer/vdfa` | DFA 词树匹配：停顿字符过滤、首个/全部匹配、密集/贪婪匹配、命中词位置元信息、包级匹配器、隔离 matcher options、`Any` 辅助函数的 JSON marshal/unmarshal provider、文本替换，以及用于包级异步初始化的可重置 async runner provider。 |
| [`verr`](docs/doc/14-verr.md) | `github.com/imajinyun/go-knifer/verr` | 错误工具：panic recover、错误聚合、multierror 匹配、collector 构造 options、堆栈捕获/格式化、可重置 log/stack cache、可注入的 logging/stack/exit/timer/runner provider、隔离 logrus 创建，以及可选 logrus/Sentry 集成。 |
| [`vfile`](docs/doc/15-vfile.md) | `github.com/imajinyun/go-knifer/vfile` | 文件与 IO 工具：读写复制、按行读取、mkdir/touch/delete、文件名处理、静默关闭和 provider-backed 文件系统操作。 |
| [`vform`](docs/doc/16-vform.md) | `github.com/imajinyun/go-knifer/vform` | 表单与输入校验工具：邮箱、手机号、URL、IPv4/IPv6、身份证、中文和数字字符串，并支持规则敏感校验的 per-call matcher provider。 |
| [`vhash`](docs/doc/17-vhash.md) | `github.com/imajinyun/go-knifer/vhash` | 非加密 Hash 工具：Additive、FNV、可注入 32-bit hash provider，以及一组经典字符串哈希（RS、JS、PJW、ELF、BKDR、SDBM、DJB、AP、HF、HFIP、TianL、Java 默认）。 |
| [`vhttp`](docs/doc/18-vhttp.md) | `github.com/imajinyun/go-knifer/vhttp` | 链式 HTTP 客户端、隔离/global-config 请求构建、create/get/post `WithOptions` 辅助函数、显式错误 `E` 快捷函数、带错误码分类的 HTTP 错误、provider-backed transport/request factory/multipart writer/download save、安全文件下载、BasicAuth、User-Agent 解析、provider-backed HTML 清理/标签过滤、可重置 transport/server starters、异步服务端 runner option 和简易服务端辅助函数。 |
| [`vid`](docs/doc/19-vid.md) | `github.com/imajinyun/go-knifer/vid` | ID 工具：random/simple/fast UUID、MongoDB 风格 ObjectId、Snowflake 生成器与单例 next-id、worker/datacenter id 推导、NanoId、fallback random source、isolated Snowflake 创建，以及可重置 fallback PRNG provider/seed。 |
| [`vident`](docs/doc/20-vident.md) | `github.com/imajinyun/go-knifer/vident` | 身份标识工具：中国大陆身份证 15/18 位转换、合法性校验、校验码、可配置解析选项的生日/年龄/性别提取、省市区编码解析、遮罩，以及港澳台证件校验。 |
| [`vimg`](docs/doc/21-vimg.md) | `github.com/imajinyun/go-knifer/vimg` | 图像工具：按长边等比缩放缩略图、PNG/JPEG/GIF 格式互转、基础元信息（宽/高/格式）、基于 ZXing 的 QRCode/Barcode 生成与解码、PNG/SVG/ASCII/Base64 Data URI 输出、二维码 logo 嵌入、透明背景，以及线条/圆圈/扭曲/GIF 图形验证码。 |
| [`vjob`](docs/doc/22-vjob.md) | `github.com/imajinyun/go-knifer/vjob` | 可切分任务执行：职责分离任务数据与调度配置，支持泛型 Slice/Map 适配、context 取消和串行合并回调；无需开启 generic type alias 实验。 |
| [`vjson`](docs/doc/23-vjson.md) | `github.com/imajinyun/go-knifer/vjson` | 有序 JSON 对象/数组、JSON 解析与格式化、路径表达式读写、provider-backed marshal/unmarshal、可注入 scalar parse/format 函数、可配置 Object/Array/Bean/List 转换，以及带 parser/writer options 的 XML/JSON 转换。 |
| [`vjwt`](docs/doc/24-vjwt.md) | `github.com/imajinyun/go-knifer/vjwt` | JWT 创建、解析、签名、验签与时间字段校验，支持 HMAC、RSA-PSS、ECDSA，拒绝未签名的 `alg=none` token，并提供 JSON marshal/unmarshal options。 |
| [`vlog`](docs/doc/25-vlog.md) | `github.com/imajinyun/go-knifer/vlog` | 日志 facade：console/color console logger、可注入颜色工厂、日志级别、全局 logger、静态日志函数、单次调用 logger options 和 isolated logger 创建。 |
| [`vmail`](docs/doc/26-vmail.md) | `github.com/imajinyun/go-knifer/vmail` | 邮件工具：RFC 5322 地址解析、链式消息构建、MIME mixed/related/alternative 渲染、文本/HTML 正文、内联文件、字节/reader/文件附件、账号默认配置 quick send、context-aware SMTP 发送、默认强制 TLS、CRLF 注入检查、附件大小限制、envelope sender 控制，以及可注入 sender/dialer/boundary generator。 |
| [`vmap`](docs/doc/27-vmap.md) | `github.com/imajinyun/go-knifer/vmap` | Map 工具：构造、空判断、contains/get/find、keys/values 与排序视图、map/filter/reject/partition、reduce/group/count、反转、合并/自定义冲突合并、交集/差集/对称差集、pick/omit、update/clone 和相等性判断。 |
| [`vmask`](docs/doc/28-vmask.md) | `github.com/imajinyun/go-knifer/vmask` | 脱敏工具：姓名、证件号、电话、地址、邮箱、密码、车牌、银行卡、IP、护照号和信用代码遮罩。 |
| [`vnet`](docs/doc/29-vnet.md) | `github.com/imajinyun/go-knifer/vnet` | 网络工具：支持注入 IP/CIDR/int parser 的 IPv4/IPv6 转换、CIDR/范围/掩码、本地端口、主机/网卡/MAC 查询、TLS 配置、address/dial/ping provider options 和 multipart 表单辅助。 |
| [`vnum`](docs/doc/30-vnum.md) | `github.com/imajinyun/go-knifer/vnum` | 数字工具：精确加减乘除、舍入模式、provider-backed 解析/格式化、数字判断、不重复随机数、range、阶乘/组合数、最大公约数/最小公倍数、二进制转换、比较、字节转换、表达式计算和奇偶判断。 |
| [`vobj`](docs/doc/31-vobj.md) | `github.com/imajinyun/go-knifer/vobj` | 对象工具：nil/空值判断、相等性、默认值、克隆/序列化、比较、类型检查和容器辅助。 |
| [`vpass`](docs/doc/32-vpass.md) | `github.com/imajinyun/go-knifer/vpass` | 密码工具：确定性的本地评分、强度分级、强/弱谓词、字符类别信号、重复/连续字符检测，以及小型常见弱密码列表。 |
| [`vpoi`](docs/doc/33-vpoi.md) | `github.com/imajinyun/go-knifer/vpoi` | Office 文档工具：轻量 Excel XLSX 工作表列表、行读写、多工作表写入、内存工作簿创建，以及可注入的 workbook/文件系统 provider。 |
| [`vrand`](docs/doc/34-vrand.md) | `github.com/imajinyun/go-knifer/vrand` | 随机工具：整数、浮点、布尔、字节、字符串、数字字符串、随机元素、确定性 seed，以及可重置的包级伪随机源 provider。 |
| [`vref`](docs/doc/35-vref.md) | `github.com/imajinyun/go-knifer/vref` | 反射工具：字段查找与赋值、方法发现与调用、构造函数风格调用、类型/值工具、方法分类判断，以及显式 unsafe/unexported 字段访问选项。 |
| [`vregex`](docs/doc/36-vregex.md) | `github.com/imajinyun/go-knifer/vregex` | 正则工具：匹配、分组提取、命名分组、删除、计数、索引定位、模板/函数替换、元字符转义，以及单次调用 compiler / DOTALL options。 |
| [`vresty`](docs/doc/37-vresty.md) | `github.com/imajinyun/go-knifer/vresty` | 基于 Resty v3 的 HTTP facade：链式请求、JSON/form/multipart 请求体、隔离/global-config 请求构建、create/get/post `WithOptions` 辅助函数、单次请求 client factory、可重置默认 Resty client provider、下载与安全文件下载，以及轻量响应工具。 |
| [`vsem`](docs/doc/38-vsem.md) | `github.com/imajinyun/go-knifer/vsem` | 加权计数信号量：支持 context 取消、FIFO 公平等待、非阻塞获取、关闭通知与占用数查询。 |
| [`vset`](docs/doc/39-vset.md) | `github.com/imajinyun/go-knifer/vset` | 泛型与常用类型集合工具：支持添加、删除、包含判断、集合运算，以及 JSON/YAML 编解码辅助。 |
| [`vskt`](docs/doc/40-vskt.md) | `github.com/imajinyun/go-knifer/vskt` | TCP socket 工具：普通连接、NIO/AIO server/client、协议编解码接口，以及可配置 thread-pool/listener/connection/runner/IP-parser provider。 |
| [`vslice`](docs/doc/41-vslice.md) | `github.com/imajinyun/go-knifer/vslice` | Slice 工具：包含/索引、反转、去重、拼接、过滤/映射、截取、合并、集合操作和分页。 |
| [`vstr`](docs/doc/42-vstr.md) | `github.com/imajinyun/go-knifer/vstr` | 字符串与文本工具：空白判断、裁剪、切分、截取、格式化、provider-backed emoji、命名转换、默认值、Unicode 转义/反转义、Ant-style 路径匹配、rune 集 Jaccard 相似度、rune n-gram 相似度、SimHash、64-bit Hamming distance、HTML 转义，以及字符判断（空白、字母、数字、ASCII、字母或数字）。 |
| [`vsys`](docs/doc/43-vsys.md) | `github.com/imajinyun/go-knifer/vsys` | 系统与运行时信息：主机、OS、用户、Go runtime、进程内存、goroutine、环境变量、可重置信息缓存，以及可注入的 env/command/runtime provider。 |
| [`vtpl`](docs/doc/44-vtpl.md) | `github.com/imajinyun/go-knifer/vtpl` | Go html/template 渲染工具，支持单次调用配置模板名、FuncMap、分隔符、template factory 和 executor。 |
| [`vurl`](docs/doc/45-vurl.md) | `github.com/imajinyun/go-knifer/vurl` | URL 与 URI 工具：解析、标准化、相对 URL 补全、query 编解码、支持注入 query/path escape provider 的 URL/路径/fragment 百分号编码、URL 构造、Data URI 构造、协议判断、文件 URL 转换、资源打开/大小查询，以及带 SSRF 防护的 `OpenSafe` / `ContentLengthSafe` 变体。 |
| [`vver`](docs/doc/46-vver.md) | `github.com/imajinyun/go-knifer/vver` | 版本工具：版本号比较、大小关系判断、表达式匹配、闭区间范围和自定义多表达式分隔符。 |
| [`vxml`](docs/doc/47-vxml.md) | `github.com/imajinyun/go-knifer/vxml` | XML 工具：解析/读取/写出/格式化、树节点访问、简单 XPath 风格查询、转义、支持 parser/codec/scalar parser options 的 Map/Bean 转换、transform options 和命名空间辅助。 |
| [`vzip`](docs/doc/48-vzip.md) | `github.com/imajinyun/go-knifer/vzip` | ZIP、gzip、zlib 工具：压缩包创建/解压、条目读取、遍历、追加、内存条目、流式压缩、provider-backed 归档文件操作、默认有边界的解压/解压缩行为、路径穿越检查，以及解压时的符号链接逃逸检查。 |

## 🧭 架构与包边界

`go-knifer` 采用 `v*` 对外 facade + `internal/*` 内部实现的结构。业务代码应优先导入
`v*` 包；`internal/*` 用于沉淀具体实现，便于后续在不暴露所有内部细节的前提下持续重构。

facade 规则：

- `internal/<domain>` 负责领域实现细节和领域内测试。
- `v<domain>` 负责暴露该领域稳定的公共 API。
- 简单工具包可以手写轻量转发；较大的模块可以保留生成的 `facade.go`。无论哪种方式，
  internal 新增导出 API 时，都应先评估是否需要进入 public facade。
- `vform`、`vmask`、`vsem`、`vskt`、`vblf`、`vver` 等短命名继续保留，通过上方模块表说明含义，
  不再通过改名破坏已有导入路径。

可配置 API 与 Provider 注入：

- 多个子包通过 `WithXxx` helper 与 `XxxWithOptions` 变体暴露 Functional Options 模式。既有固定参数 API
  保持稳定，option 变体用于为需要高级控制的调用方提供扩展能力。
- 该模式已覆盖布隆过滤器、缓存、验证码、配置加载/监听、定时任务、加密、数据库、日期时间、DFA、错误处理、
  文件、HTTP/Resty、ID、身份解析、JSON/JWT、日志、网络、数字、POI、随机数、socket、系统信息、URL、XML、ZIP 等运行时敏感能力。
- Provider 风格的 option 允许调用方注入文件系统函数、网络/TLS dialer 或 reader、HTTP request/multipart factory、
  clock、timer/ticker、随机源、DB opener、Excel workbook factory、logger、stack capture、finalizer、环境变量查询、
  command executor、Sentry/logrus hook 等进程全局依赖，便于确定性测试与受控运行时行为。
- 对标量和解析密集型工具，也在调用点提供 provider options：`vnet` 的 IP/CIDR/int parser、`vskt` 的
  host-to-IP parser、`vjson` 的 string/int/float/bool 解析与 int/float 格式化 provider、`vxml` 的
  XML-to-map scalar parser、`vnum` 的表达式/double 解析与格式化 provider，以及 `vurl` 的 query/path
  escape provider。普通 API 继续保持标准库行为，并在内部委托给 option 变体。
- 包级默认值必须显式治理。例如 HTTP 全局默认值可通过 `vhttp.SnapshotGlobalConfig` 读取不可变快照；
  `vhttp.NewIsolatedRequest` 可以在不读取包级默认值的情况下构建请求；单次调用 option 不应隐式修改隐藏的全局状态。

Provider 覆盖重点：

| 领域 | 示例 |
| --- | --- |
| HTTP / Resty | `vhttp.NewIsolatedRequest`、`vhttp.NewRequestWithConfig`、`vhttp.Get`、`vhttp.Post`、`vhttp.GetSafe`、`vhttp.PostSafe`、`vhttp.GetStringE`、`vhttp.GetStringSafeE`、`vhttp.GetWithTimeoutE`、`vhttp.PostJSONE`、`vhttp.PostJSONSafeE`、`vhttp.DownloadBytesE`、`vhttp.DownloadBytesSafeE`、`vhttp.DownloadFile`、`vhttp.DownloadFileSafe`、`vhttp.DownloadFileSafeWithOptions`、`vhttp.NewErrorWithCode`、`vhttp.WithTransportProvider`、`vhttp.WithRequestFactory`、`vhttp.WithMultipartWriterFactory`、`vhttp.ResetDefaultTransport`、`vhttp.WithListenAndServeFunc`、`vhttp.WithAsyncRunner`、`vhttp.CreateServerWithOptions`、`vhttp.CleanHTMLWithOptions`、`vhttp.FilterHTMLTagWithOptions`、`vhttp.WithHTMLFilterCompileFunc`、`vresty.NewIsolatedRequest`、`vresty.WithGlobalConfig`、`vresty.WithRestyClientFactory`、`vresty.ConfigureDefaultRestyClientProvider`、`vresty.ResetDefaultRestyClientProvider`、`vresty.Get`、`vresty.Post`、`vresty.GetSafe`、`vresty.PostSafe`、`vresty.GetStringE`、`vresty.GetStringSafeE`、`vresty.GetWithTimeoutE`、`vresty.PostJSONE`、`vresty.PostJSONSafeE`、`vresty.DownloadBytesE`、`vresty.DownloadBytesSafeE`、`vresty.DownloadFile`、`vresty.DownloadFileSafe`、`vresty.DownloadFileSafeWithOptions` |
| 文件 / 配置 / 压缩 / POI | `vfile` provider options、`vconf.LoadWithOptions`、`vconf.LoadRemoteSafeWithOptions`、`vconf.WatchWithOptions`、`vconf.WatchOptions.Runner`、`vzip.WithMaxBytes`、`vzip` provider options、`vpoi.WithOpenFileFunc`、`vpoi.WithNewFileFunc`、`vpoi.WithSaveAsFunc` |
| Cron / DFA / ID / 身份 / 随机数 | `vcron.WithDefaultSchedulerOptions`、`vcron.NewConfigWithOptions`、`vcron.WithIDRandomReader`、`vcron.WithRunner`、`vcron.CronScheduleWithOptions`、`vdfa.WithMatcherWords`、`vdfa.WithJSONMarshal`、`vdfa.WithJSONUnmarshal`、`vdfa.ContainsWithOptions`、`vdfa.ConfigureAsyncRunner`、`vdfa.ResetAsyncRunner`、`vid.NewIsolatedSnowflake`、`vid.CreateSnowflakeWithOptions`、`vid.WithSnowflakeCache`、`vid.WithFallbackRandomSource`、`vid.ConfigureDefaultFallbackRandomSourceProvider`、`vid.ResetDefaultFallbackRandomSource`、`vid.SetFallbackRandomSeed`、`vrand.ConfigureDefaultRandomSourceProvider`、`vrand.ResetDefaultRandomSource`、`vrand.SetSeed`、`vident.BirthDateWithOptions` |
| 编解码 / 图像 / JSON / XML / JWT / hash | `vcodec.Base64EncodeWithEncoding`、`vcodec.Base64DecodeWithEncoding`、`vcodec.Base64RawURLEncode`、`vcodec.Base64RawURLDecode`、`vimg.Thumbnail`、`vimg.ConvertFormat`、`vimg.Info`、`vimg.QRCodePNG`、`vimg.QRCodeSVG`、`vimg.QRCodeASCII`、`vimg.QRCodeBytes`、`vimg.BarcodePNG`、`vimg.BarcodeSVG`、`vimg.DecodeQRCode`、`vimg.DecodeBarcode`、`vimg.SupportedEncodeBarcodeFormats`、`vimg.SupportedDecodeBarcodeFormats`、`vimg.WithQRCodeLogo`、`vimg.WithQRCodeLogoRatio`、`vimg.WithQRCodeTransparentBackground`、`vimg.NewLineCaptcha`、`vimg.NewCircleCaptcha`、`vimg.NewShearCaptcha`、`vimg.NewGifCaptcha`、`vhash.Hash32`、`vjson.WithMarshalFunc`、`vjson.WithUnmarshalFunc`、`vjson.WithParseUnmarshalFunc`、`vjson.WithBeanUnmarshalFunc`、`vjson.WithSprintFunc`、`vjson.WithParseIntFunc`、`vjson.WithParseFloatFunc`、`vjson.WithParseBoolFunc`、`vjson.WithFormatIntFunc`、`vjson.WithFormatFloatFunc`、`vjson.ParseObjWithOptions`、`vjson.ParseArrayWithOptions`、`vjson.ToBeanWithOptions`、`vjson.ToListWithOptions`、`vjson.XMLToJSONWithOptions`、`vjson.ToXMLWithOptions`、`vxml.WithScalarIntParser`、`vxml.WithScalarFloatParser`、`vxml.XMLToMapWithOptions`、`vxml.XMLNodeToMapWithOptions`、`vxml.XMLToMapIntoWithOptions`、`vxml.XMLNodeToMapIntoWithOptions`、`vxml.XMLToBeanWithOptions`、`vxml.XMLNodeToBeanWithOptions`、`vxml.TransformWithOptions`、`vxml.FormatWithOptions`、`vjwt.WithJSONMarshalFunc`、`vjwt.WithJSONUnmarshalFunc`、`vjwt.ParseTokenWithOptions`、`vjwt.WithTokenJSONOptions` |
| 加密 / 密码 / 模板 / 正则 / 校验 / 字符串 | `vcrypto.Digest`、`vcrypto.DigestHex`、`vcrypto.WithGCMBlockFactory`、`vcrypto.AESSealGCMWithOptions`、`vcrypto.AESEncryptGCMWithOptions`、`vcrypto.SignWithRSAOptions`、`vcrypto.VerifyWithRSAOptions`、`vpass.Analyze`、`vpass.Score`、`vpass.StrengthOf`、`vpass.IsStrong`、`vpass.IsWeak`、`vtpl.RenderWithOptions`、`vtpl.WithFuncMap`、`vtpl.WithTemplateFactory`、`vregex.WithCompileFunc`、`vregex.WithDotAll`、`vregex.MatchWithOptions`、`vregex.ReplaceAllFuncWithOptions`、`vform.IsEmailWithOptions`、`vform.WithMobileMatcher`、`vstr.ContainsEmojiWithOptions`、`vstr.RemoveEmojiWithOptions`、`vstr.JaccardSimilarity`、`vstr.NGramSimilarity`、`vstr.SimHash`、`vstr.HammingDistance64` |
| DB / 网络 / 数字 / URL / 系统 / 反射 / socket | `vdb.WithSQLOpenFunc`、`vnet.WithConnectDialer`、`vnet.WithPingDialer`、`vnet.WithAddressNetwork`、`vnet.WithTCPAddrResolver`、`vnet.WithUploadOpenSource`、`vnet.WithIPParser`、`vnet.WithCIDRParser`、`vnet.WithIPIntParser`、`vnet.WithWildcardIPParser`、`vnet.WithWildcardIntParser`、`vnet.IPv4ToLongWithOptions`、`vnet.IsInRangeWithOptions`、`vnum.WithParseFloatFunc`、`vnum.WithDoubleParseFloatFunc`、`vnum.WithDoubleFormatFloatFunc`、`vnum.CalculateWithOptions`、`vnum.ToDoubleWithOptions`、`vurl.WithQueryEscapeFunc`、`vurl.WithPathEscapeFunc`、`vurl.EncodeQueryWithOptions`、`vurl.EncodePathSegmentWithOptions`、`vurl.FormURLEncodeWithOptions`、`vurl.OpenSafeWithOptions`、`vurl.WithAllowedSchemes`、`vurl.WithAllowedHosts`、`vurl.WithRejectPrivateHosts`、`vurl.WithAllowLocalFiles`、`vsys.WithGoEnvOutputFunc`、`vsys.WithGoRootEnvLookupFunc`、`vsys.WithOSEnvLookupFunc`、`vsys.WithEnvLookupFunc`、`vsys.ResetInfoCache`、`vref.WithUnsafeAccess`、`vskt.WithThreadPoolSizeFunc`、`vskt.WithRunner`、`vskt.WithSocketIPParser` |
| 邮件 | `vmail.Account`、`vmail.QuickSend`、`vmail.SendAccountText`、`vmail.SendAccountHTML`、`vmail.WithQuickMessageOptions`、`vmail.WithQuickClientOptions`、`vmail.WithEnvelopeFrom`、`vmail.NewAttachmentReader`、`vmail.NewAttachmentFile`、`vmail.NewInlineReader`、`vmail.NewInlineFile`、`vmail.WithAttachmentReader`、`vmail.WithAttachmentFile`、`vmail.WithInlineReader`、`vmail.WithInlineFile`、`vmail.WithSenderProvider`、`vmail.WithDialContext`、`vmail.WithBoundaryGenerator` |
| 错误 / 缓存 / 日志 / 运行时 | `verr.NewCollectorWithOptions`、`verr.WithCollectorLogFunc`、`verr.WithCollectorRunner`、`verr.WithCollectorContext`、`verr.WithCollectorLevel`、`verr.WithCollectorTimerFactory`、`verr.WithCollectorStackCaptureOptions`、`verr.WithLogFunc`、`verr.WithCollectorStackOptions`、`verr.WithDebugStackFunc`、`verr.WithCallersFunc`、`verr.WithFuncForPCFunc`、`verr.WithStackFrameCache`、`verr.ResetStackFrameCache`、`verr.ResetDefaultLogFunc`、`verr.NewIsolatedLogrusWithOptions`、`verr.MustExitWithOptions`、`vcache.WithClock`、`vcache.WithTickerFactory`、`vcache.WithRunner`、`vcache.WithWeakFinalizerFunc`、`vcache.WithWeakFinalizerEnabled`、`vlog.WithLogColorFactory`、`vlog.NewIsolatedLogger`、`vlog.LoggerWithOptions`、`vlog.InfoWithOptions` |

领域边界规则：

- `vhash` 面向非加密 hash 能力，例如 Additive/FNV（分桶、布隆过滤器等场景）；`vcrypto` 独占
  安全相关 SHA-2 摘要、HMAC、加解密、密钥和 PEM 编解码。
- `vhttp` 是基于标准库的轻量 HTTP facade；`vresty` 是基于 Resty 的链式高级 HTTP client facade。两者都不再重复暴露 URL 工具：URL 转义、query 构建/解析、协议判断（`IsHTTP`/`IsHTTPS`、`EncodeQueryMap`、`DecodeQuery` 等）统一归 `vurl`。
- `vdb` 负责基于 `database/sql` 的 SQL 数据库辅助能力；调用方继续通过 `*sql.DB` 和单次调用 options
  控制驱动和连接池。
- `vdfa` 负责 DFA 词树匹配、停顿字符过滤、密集/贪婪匹配、命中词位置元信息和文本替换；通用字符串工具不承载词典匹配逻辑。
- `vid` 负责 UUID、Snowflake、ObjectId、NanoId 等生成型标识；`vident` 负责法定身份号码与地区证件解析，
  例如中国大陆身份证和港澳台证件号。
- `vcodec` 负责 Base64、Hex 等编码/解码算法；`vurl` 负责 URL 转义、URL/URI 解析、规范化、
  资源打开/大小查询和协议语义。
- `vjson` 负责 JSON 对象、数组、路径和轻量 XML adapter；`vxml` 负责 XML 解析、树访问、格式化、
  namespace 和 XML 专属的 map/bean 转换。
- `vbean` 负责直接的 struct/map 属性映射、copy properties、tag/alias 匹配和弱类型转换，
  不通过 JSON 序列化绕路。
- `vobj` 是对象级便利 facade。新增具体领域逻辑应优先落到 `vstr`、`vslice`、`vmap`、`vref`
  等明确领域包，只有在对象级聚合有价值时再由 `vobj` 做轻量包装。

数据库工具归属 `internal/db`，并通过 `vdb` 对外暴露；DFA 文本匹配归属 `internal/dfa`，并通过
`vdfa` 对外暴露；Office 文档工具归属 `internal/poi`，并通过 `vpoi` 对外暴露。跨领域通用输入校验归属
`internal/validator`，并通过 `vform` 对外暴露；领域内解析和更丰富的操作仍留在 `vident`、`vnet`、`vurl` 等领域包。

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
vcron、vjob、vpoi 空 sheet 名、vcodec 解码失败、vdate 解析失败、vbean 映射/转换失败、vsem 非法权重和无效 HTTP 请求输入）、
`knifer.ErrCodeTimeout`（HTTP 超时/deadline）、`knifer.ErrCodeNotFound`（vpoi 无 sheet、vblf 初始化文件不存在）、
`knifer.ErrCodeUnsupported`（vsem 已关闭、HTTP 重定向/响应体限制场景）与
`knifer.ErrCodeInternal`（其余 vhttp/vresty 传输或读取失败、vskt、vblf 读取失败、verr recover 到的 panic），
同时保留各自的 error 类型、哨兵与 cause 错误链。

### 安全与防护默认值

安全敏感工具只保留当前推荐的公共 API。`vcrypto` 保留 SHA-2 摘要、HMAC-SHA-256/384/512、
PBKDF2-SHA-256、AES-GCM、RSA-OAEP 加密和 RSA-PSS 签名。JWT RSA
签名通过 RSA-PSS 暴露（`JWTAlgPS256`、`JWTAlgPS384`、`JWTAlgPS512`、
`NewRSAPSSSigner` 以及 `PS256` / `PS384` / `PS512` 辅助函数），同时保留 HMAC 与 ECDSA signer。
未签名的 JWT `alg=none` token 会被拒绝，公共 API 不再暴露 none signer。

网络与 IO 工具默认采用有边界、显式的行为：

- TLS 工具通过 `vnet.CreateTLSConfig()` 创建 TLS 1.2+ 的配置。HTTP 客户端通过
  `WithTLSConfig` 接收显式的 `*tls.Config`；不提供跳过证书校验的便利 API。
- HTTP 与 Resty 下载在把自动识别的文件名拼接到目标目录前会先校验文件名，避免目录穿越。
  当来源 URL 不可信时，使用 `vhttp.DownloadFileSafe` / `DownloadFileSafeWithOptions` 或
  `vresty.DownloadFileSafe` / `DownloadFileSafeWithOptions`；这些安全变体会同时应用安全请求 URL 策略与保存目标校验。
- `vfile` 读取工具默认使用 `vfile.DefaultMaxBytes` 限制读取大小。需要更严格限制时使用
  `vfile.WithMaxBytes(n)`；只有调用方已经在其他层面限制输入时，才使用 `vfile.WithUnlimitedRead()`。
- `vconf` 本地与远程加载默认使用 `vconf.DefaultMaxBytes`。设置 `LoadOptions.MaxBytes` 可改变限制，
  负数表示显式关闭配置读取限制。
- 从不可信或用户可控 URL 加载远程配置时，使用 `vconf.LoadRemoteSafe` 或
  `LoadRemoteSafeWithOptions`。安全变体只接受 HTTP(S)，默认拒绝 localhost、私有网段、link-local
  和 unspecified 目标，并用同一策略校验重定向目标。需要允许私有测试服务或明确的内网主机时，设置
  `LoadOptions.RemoteAllowedHosts`。
- 从不可信输入打开远程资源时，使用 `vurl.OpenSafe`、`OpenSafeWithOptions`、
  `ContentLengthSafe` 或 `ContentLengthSafeWithOptions`。安全资源 helper 默认只允许 HTTP(S)、
  拒绝本地文件和普通文件路径、拒绝私有网络目标、检查 HTTP 状态、设置超时，并重新校验重定向。优先用
  `WithAllowedHosts` 绑定可信 host；host allowlist 只会收窄可接受的主机名，不会绕过私有地址拒绝。
  只有调用方已建立更窄的信任边界时，才放宽 `WithRejectPrivateHosts` 或 `WithAllowLocalFiles`。
- `vzip` 解压和解压缩 helper 默认有大小边界，以降低 zip bomb 风险。ZIP 条目名会在写入前清理和校验；
  解压时会通过 `filepath.EvalSymlinks` 解析目标父目录，拒绝经由符号链接逃出目标目录的条目。
  只有测试或虚拟文件系统需要替换解析器时才使用 `vzip.WithEvalSymlinks`。使用 `vzip.WithMaxBytes(n)`
  或 `UnzipToLimit` / `UnzipReaderToLimit` 设置更严格的预算；只有其他层已对可信输入设置大小限制时，
  才传入负数关闭 max-byte 限制。
- 以 `E` 结尾的布隆过滤器构造器，例如 `vblf.NewBitMapBloomFilterE`、
  `vblf.NewBitSetBloomFilterE` 和 `vblf.NewFuncFilterE`，会在 size 或 hash 配置非法时返回校验错误，
  而不是 panic。非 `E` 构造器保留用于兼容已有调用方。
- `vdb` 条件构造器会校验操作符白名单；优先使用 `Eq`、`Like`、`In`、`Between`、
  `IsNull`、`IsNotNull` 等辅助函数，而不是拼接原始 SQL 片段。
- `vskt.AioSession` 会串行化共享 session buffer 的读取，并在关闭回调期间保留 buffer，便于生命周期钩子安全检查最后收到的数据。
- JWT `alg=none` 始终拒绝；空算法或不支持的算法不会降级成无签名 token。

## 🚀 安装

项目要求 Go 1.25 或更高版本。

```bash
go get github.com/imajinyun/go-knifer
```


## ✅ 推荐 API 入口

新代码建议优先使用这些入口。可能失败的快捷请求应使用显式返回 error 的版本，避免静默吞掉失败。

| 场景 | 推荐 API |
| --- | --- |
| 构建可信的标准库 HTTP 请求 | `vhttp.Get`、`vhttp.Post`、`vhttp.NewRequest` |
| 读取可信 HTTP 响应正文并处理错误 | `vhttp.GetStringE`、`vhttp.PostJSONE`、`vhttp.DownloadBytesE` |
| 访问用户可控或其他不可信 HTTP(S) URL | `vhttp.GetStringSafeE`、`vhttp.PostJSONSafeE`、`vhttp.DownloadBytesSafeE` |
| 使用 Resty-backed HTTP facade | `vresty.Get`、`vresty.Post`、`vresty.GetStringE`、`vresty.PostJSONE` |
| 通过 Resty 访问不可信 URL | `vresty.GetStringSafeE`、`vresty.PostJSONSafeE`、`vresty.DownloadBytesSafeE` |
| 把用户可控 URL 下载到文件 | `vhttp.DownloadFileSafe` 或 `vresty.DownloadFileSafe` |
| 生成 secret、token、key、nonce 或 salt 字节 | `vrand.SecureBytes` |
| 创建 LRU 缓存 | `vcache.NewLRU` 或 `vcache.NewLRUWithTimeout` |
| 解析 cron 表达式 | `vcron.NewPattern` 或 `vcron.MustNewPattern` |
| 从信任边界加载远程配置 | `vconf.LoadRemoteSafe` 或 `vconf.LoadRemoteSafeWithOptions` |

## 📝 Quickstart 文档

README 只保留模块导航。每个 `v*` 子包的 Quickstart 示例已经拆分到上方模块矩阵中的对应链接，便于按领域查看和维护。

## 📖 文档

- 根包说明：`doc.go`
- 对外 API：各 `v*` 子包的 `doc.go` 与 facade 文件
- Quickstart 示例：上方模块矩阵中的 `docs/doc/*.md` 链接
- 在线文档：[pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)

## 📦 下载与构建

下载源码：

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

运行测试：

```bash
make test
```

本地运行 CI 测试作业同级别的门禁，包括模块校验、vet、tidy/diff 清洁度、架构检查、race/shuffle 测试、覆盖率门禁和 API 快照检查：

```bash
make ci-test
```

提交 PR 前建议运行与 CI 对齐的完整本地稳定性检查：

```bash
make check
```

`make check` 包含 `ci-test` 同级别检查，并额外运行 `golangci-lint` 和 `govulncheck`。如需单独扫描漏洞，使用：

```bash
make govulncheck
```

运行 slice、map、string、numeric 热点工具函数的 benchmark 基线：

```bash
make bench-core
```

运行对应 public facade 包的 benchmark 基线：

```bash
make bench-facade
```

只有在需要用统计方式比较性能变化时，才提高运行次数和单轮时长：

```bash
make bench-core BENCHCOUNT=10 BENCHTIME=3s
```

GitHub Actions 会运行 Go 测试矩阵、架构检查、golangci-lint、govulncheck 和 CodeQL。
Dependabot 已配置 Go modules 与 GitHub Actions 依赖更新。

格式化代码：

```bash
gofmt -w .
```

## 🛡️ 治理

- 安全报告：请参见 [SECURITY.md](./SECURITY.md)，不要在公开 Issue 中披露疑似漏洞。
- 覆盖率门禁：CI 使用 `bash bin/check_coverage.sh coverage.out` 校验仓库总覆盖率和重点包覆盖率。只有在新增测试支撑后，才提升 `COVERAGE_THRESHOLD` 或 `PACKAGE_COVERAGE_THRESHOLDS`。
- API 门禁：`make api-check` 会将根包和顶层 `v*` 包的导出符号与 `docs/api/exports.txt` 对比。只有有意修改公共 API 时才刷新并提交快照。
- 稳定性门禁：提交前优先使用 `make check`，保持 vet、架构检查、race/shuffle 测试、覆盖率、API 兼容、lint 和漏洞扫描与 CI 对齐。
- Benchmark 基线：使用 `make bench-core` 确认热点工具函数 benchmark 可运行，使用 `make bench-facade` 确认对应 public facade 包 benchmark 可运行。除非另行使用 `benchstat` 做多轮对比，否则 benchmark 输出只作为基线，不作为性能提升结论。

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
3. 为核心逻辑补充单元测试，提交前优先执行 `make check`；
4. 保持代码经过 `gofmt` 格式化；
5. 避免引入不必要的第三方依赖，优先复用标准库。

## ⭐ Star go-knifer

如果这个项目减少了你的重复代码，欢迎给它一个 Star。你的反馈和贡献会帮助它成为更趁手的 Go 工具集合。
