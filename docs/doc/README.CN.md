# 📚 go-knifer 中文文档中心

> `go-knifer` 的详细模块导航、架构说明、安全默认值和贡献工作流。

## 📑 Table of Contents

- [🧭 快速导航](#quick-navigation)
- [⭐ 从 star domain 开始](#start-with-star-domains)
- [🧩 模块目录](#package-catalog)
- [📝 Quickstart 文档](#quickstart-documents)
- [🏗️ 架构与包边界](#architecture-and-package-boundaries)
- [✅ 推荐 API 入口](#recommended-api-entry-points)
- [📦 构建、测试与发布工作流](#build-test-and-release-workflow)
- [🛡️ 治理](#governance)
- [🤝 贡献](#contributing)

<a id="quick-navigation"></a>

## 🧭 快速导航

- 🏠 中文根 README：[`../../README.CN.md`](../../README.CN.md)
- 🏠 English root README: [`../../README.md`](../../README.md)
- 🌐 在线 Go 文档：[pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)
- 🧾 公共 API 快照：[`../api/exports.txt`](../api/exports.txt)
- 🤖 机读工具目录：[`../api/tools.json`](../api/tools.json)
- 📋 可读工具目录：[`../api/tools.md`](../api/tools.md)
- 🗺️ AI 项目地图：[`../../llms.txt`](../../llms.txt)
- 🤖 机器可读 AI/CLI 元数据：[`../../ai-context.json`](../../ai-context.json)
- 🧯 安全策略：[`../../SECURITY.md`](../../SECURITY.md)
- 📝 变更日志：[`../../CHANGELOG.md`](../../CHANGELOG.md)

<a id="start-with-star-domains"></a>

## ⭐ 从 star domain 开始

下面三个领域最适合用来快速判断 `go-knifer` 是否适合你的项目。它们同时具备推荐 API 入口、可执行示例、cookbook 工作流、benchmark 命令和明确的安全边界。

| 需求 | 从这里开始 | 可信信号 |
| --- | --- | --- |
| 安全 HTTP 请求与下载 | [`vhttp`](22-vhttp.md)、[`vresty`](41-vresty.md)、[`vurl`](51-vurl.md) | helper 选择指南、安全 URL 策略清单、FAQ、benchmark 命令和标准库/Resty 边界说明。 |
| 安全加密工作流 | [`vcrypto`](11-vcrypto.md)、[`vrand`](38-vrand.md)、[`vjwt`](28-vjwt.md) | 推荐加密入口、secret 处理 FAQ、benchmark 命令和直接使用标准库的边界说明。 |
| 日常 JSON 与文件工作流 | [`vjson`](27-vjson.md)、[`vfile`](17-vfile.md) | 对象/路径/格式化/文件 IO cookbook 示例、文件系统安全建议和显式错误处理。 |

对比入口：

- 使用 [`vhttp`](22-vhttp.md) 获取标准库风格 HTTP helper；当 Resty 风格链式请求更易读时使用 [`vresty`](41-vresty.md)。
- 使用 [`vurl`](51-vurl.md) 做 URL 构造、标准化、query 编码、资源探测，或在真正发 HTTP 请求前做安全打开。
- 当推荐加密工作流可以降低误用风险时使用 [`vcrypto`](11-vcrypto.md)；当调用方需要更底层的协议控制时直接使用 Go 标准库。
- 使用 [`vjson`](27-vjson.md) 处理常见对象、路径、格式化和 XML bridge 流程；如果需要 streaming、tokenization 或完整 decoder 控制，直接使用 `encoding/json`。
- 使用 [`vfile`](17-vfile.md) 处理有边界读取、provider-backed 文件系统测试和显式文件错误；不可信路径处理应保留在调用点显式可见。

<a id="package-catalog"></a>

## 🧩 模块目录

项目采用“内部实现 + 对外 facade”的组织方式：`internal/*` 保存具体实现，`v*` 包提供稳定、可导入的公共 API。

| 模块 | 导入路径 | 功能说明 |
| --- | --- | --- |
| [`vai`](01-vai.md) | `github.com/imajinyun/go-knifer/vai` | AI adapter 工具：可注入 provider 的 chat/embedding、请求校验、防御性拷贝、确定性示例和 redaction-safe 诊断文本。 |
| [`vbean`](02-vbean.md) | `github.com/imajinyun/go-knifer/vbean` | Bean/结构体映射工具：struct/map 互转、copy properties、tag/alias 匹配、忽略空值/零值选项和弱类型转换。 |
| [`vblf`](03-vblf.md) | `github.com/imajinyun/go-knifer/vblf` | 布隆过滤器：bitmap/bitset/filter 抽象、多种字符串哈希算法、option 构造器、返回校验错误的 `E` 构造器和 provider-backed 文件初始化。 |
| [`vbool`](04-vbool.md) | `github.com/imajinyun/go-knifer/vbool` | 布尔工具：取反、转 int、全真/任一为真判断。 |
| [`vcache`](05-vcache.md) | `github.com/imajinyun/go-knifer/vcache` | 泛型缓存：FIFO、LFU、LRU、Timed、Weak、NoCache，支持 TTL、clock、淘汰监听、懒加载、ticker/runner provider 和 weak-cache finalizer provider。 |
| [`vcli`](06-vcli.md) | `github.com/imajinyun/go-knifer/vcli` | CLI 工具：context-aware 命令执行、可注入 runner、类型化 flag 解析、子命令路由、确定性 help 渲染和 ANSI color 控制。 |
| [`vcodec`](07-vcodec.md) | `github.com/imajinyun/go-knifer/vcodec` | 编解码工具：Base64、URL-safe Base64、raw URL-safe Base64、自定义 Base64 encoding provider 和 Hex。 |
| [`vconf`](08-vconf.md) | `github.com/imajinyun/go-knifer/vconf` | 分组配置读取：setting/properties 风格文本、YAML 子集和 TOML 解析，支持类型化读取、schema 校验、profile/remote/file 加载、SSRF 防护远程加载、有边界读取和 clone。 |
| [`vconv`](09-vconv.md) | `github.com/imajinyun/go-knifer/vconv` | 宽松类型转换：string、int、int64、float64、bool、bytes 及默认值版本。 |
| [`vcron`](10-vcron.md) | `github.com/imajinyun/go-knifer/vcron` | Cron 表达式解析与任务调度，支持可配置 scheduler/options、provider 注入、运行任务指标、`Wait` 和 `Shutdown(ctx)` 优雅关闭。 |
| [`vcrypto`](11-vcrypto.md) | `github.com/imajinyun/go-knifer/vcrypto` | 加密与摘要：SHA-2、HMAC、PBKDF2-SHA256、参数签名、随机字节、AES-GCM、RSA OAEP/PSS、PEM 和 X.509 工具。 |
| [`vcsv`](12-vcsv.md) | `github.com/imajinyun/go-knifer/vcsv` | CSV 工具：reader/writer options、records-to-map 转换、map 写出、struct tag 导出和逐行回调。 |
| [`vdate`](13-vdate.md) | `github.com/imajinyun/go-knifer/vdate` | 日期时间工具：常用布局、解析/格式化、日/月/年起止、偏移和比较。 |
| [`vdb`](14-vdb.md) | `github.com/imajinyun/go-knifer/vdb` | 基于 `database/sql` 的数据库工具：SQL 执行、命名参数、Entity、条件、查询构造器、事务、分页、元信息查询和可注入 `sql.Open` provider。 |
| [`vdfa`](15-vdfa.md) | `github.com/imajinyun/go-knifer/vdfa` | DFA 词树匹配：停顿字符过滤、首个/全部匹配、密集/贪婪模式、命中词元信息、matcher helper、文本替换和异步初始化 provider。 |
| [`verr`](16-verr.md) | `github.com/imajinyun/go-knifer/verr` | 错误工具：panic recover、错误聚合、multierror 匹配、堆栈捕获/格式化、logging/stack/exit/timer/runner provider 和可选 logrus/Sentry 集成。 |
| [`vfile`](17-vfile.md) | `github.com/imajinyun/go-knifer/vfile` | 文件与 IO 工具：读写复制、按行读取、mkdir/touch/delete、文件名处理、静默关闭和 provider-backed 文件系统操作。 |
| [`vform`](18-vform.md) | `github.com/imajinyun/go-knifer/vform` | 表单与输入校验工具：邮箱、手机号、URL、IPv4/IPv6、身份证、中文、数字字符串和 matcher provider。 |
| [`vftp`](19-vftp.md) | `github.com/imajinyun/go-knifer/vftp` | FTP adapter 工具：可注入 provider 的目录列表、内存下载/上传契约、请求校验、传输大小限制和防御性拷贝。 |
| [`vhan`](20-vhan.md) | `github.com/imajinyun/go-knifer/vhan` | 汉字转写 adapter 工具：可注入 provider 的中文转拼音与首字母提取、请求校验、输入长度限制和防御性拷贝。 |
| [`vhash`](21-vhash.md) | `github.com/imajinyun/go-knifer/vhash` | 非加密 Hash 工具：Additive、FNV、可注入 32-bit provider 和经典字符串哈希。 |
| [`vhttp`](22-vhttp.md) | `github.com/imajinyun/go-knifer/vhttp` | 标准库 HTTP facade：链式客户端、全局/隔离配置、显式错误快捷函数、分类 HTTP 错误、安全下载、BasicAuth、HTML helper 和 provider-backed transport/factory。 |
| [`vid`](23-vid.md) | `github.com/imajinyun/go-knifer/vid` | ID 工具：UUID、ObjectId、Snowflake、worker/datacenter 推导、NanoId、fallback random source 和隔离 Snowflake 创建。 |
| [`vident`](24-vident.md) | `github.com/imajinyun/go-knifer/vident` | 身份标识工具：中国大陆身份证转换/校验、生日/年龄/性别提取、省市区解析、遮罩和港澳台证件校验。 |
| [`vimg`](25-vimg.md) | `github.com/imajinyun/go-knifer/vimg` | 图像工具：缩略图、PNG/JPEG/GIF 转换、元信息、QR/barcode 生成与解码、二维码 logo/背景 options 和图形验证码。 |
| [`vjob`](26-vjob.md) | `github.com/imajinyun/go-knifer/vjob` | 可切分任务执行，支持 typed adapters、context 取消和串行 merge 回调。 |
| [`vjson`](27-vjson.md) | `github.com/imajinyun/go-knifer/vjson` | 有序 JSON 对象/数组、解析/格式化、路径 get/put、provider-backed marshal/unmarshal、可配置转换和 XML/JSON adapter。 |
| [`vjwt`](28-vjwt.md) | `github.com/imajinyun/go-knifer/vjwt` | JWT 创建、解析、签名、验签、时间字段校验，支持 HMAC/RSA-PSS/ECDSA 并拒绝未签名 token。 |
| [`vlog`](29-vlog.md) | `github.com/imajinyun/go-knifer/vlog` | 日志 facade：console/color logger、日志级别、全局 logger、静态函数、单次调用 options 和 isolated logger 创建。 |
| [`vmail`](30-vmail.md) | `github.com/imajinyun/go-knifer/vmail` | 邮件工具：RFC 5322 解析、MIME 消息构建、文本/HTML、内联文件、附件、quick send、context-aware SMTP、默认强制 TLS、注入检查和 provider options。 |
| [`vmap`](31-vmap.md) | `github.com/imajinyun/go-knifer/vmap` | Map 工具：构造、contains/get/find、排序 keys/values、map/filter/reject/partition、reduce/group/count、反转、合并、集合差异、pick/omit、clone 和相等性。 |
| [`vmask`](32-vmask.md) | `github.com/imajinyun/go-knifer/vmask` | 脱敏工具：姓名、证件号、电话、地址、邮箱、密码、车牌、银行卡、IP、护照号和信用代码遮罩。 |
| [`vnet`](33-vnet.md) | `github.com/imajinyun/go-knifer/vnet` | 网络工具：IPv4/IPv6 转换、CIDR/范围/掩码、本地端口、主机/网卡/MAC 查询、TLS 配置、dial/ping options 和 multipart 表单。 |
| [`vnum`](34-vnum.md) | `github.com/imajinyun/go-knifer/vnum` | 数字工具：精确运算、泛型聚合、舍入、解析/格式化 provider、不重复随机数、range、阶乘/组合数、gcd/lcm、二进制转换、字节转换和表达式计算。 |
| [`vobj`](35-vobj.md) | `github.com/imajinyun/go-knifer/vobj` | 对象工具：nil/空值判断、相等性、默认值、克隆/序列化、比较、类型检查和容器辅助。 |
| [`vpass`](36-vpass.md) | `github.com/imajinyun/go-knifer/vpass` | 密码工具：确定性本地评分、强度分级、强/弱谓词、重复/连续字符检测和常见弱密码列表。 |
| [`vpoi`](37-vpoi.md) | `github.com/imajinyun/go-knifer/vpoi` | Office 文档工具：XLSX sheet 列表、行读写、多 sheet 写入、内存 workbook 创建和可注入 workbook/文件系统 provider。 |
| [`vrand`](38-vrand.md) | `github.com/imajinyun/go-knifer/vrand` | 随机工具：整数、浮点、布尔、字节、字符串、数字字符串、随机元素、确定性 seed 和可重置伪随机源 provider。 |
| [`vref`](39-vref.md) | `github.com/imajinyun/go-knifer/vref` | 反射工具：字段查找/赋值、方法发现/调用、构造函数风格调用、nil-safe 类型/值工具、分类 helper 和显式 unsafe access options。 |
| [`vregex`](40-vregex.md) | `github.com/imajinyun/go-knifer/vregex` | 正则工具：匹配、分组提取、命名分组、删除、计数、索引定位、模板/函数替换、转义和 compiler/DOTALL options。 |
| [`vresty`](41-vresty.md) | `github.com/imajinyun/go-knifer/vresty` | Resty v3 HTTP facade：链式请求、JSON/form/multipart body、隔离/全局配置、request factory、可重置默认 client、下载、安全下载和响应 helper。 |
| [`vsem`](42-vsem.md) | `github.com/imajinyun/go-knifer/vsem` | 加权、context-aware 计数信号量，支持 FIFO 公平等待、try-acquire、关闭通知和占用数指标。 |
| [`vset`](43-vset.md) | `github.com/imajinyun/go-knifer/vset` | 泛型与常用类型集合工具，支持 add/remove/contains、集合运算和 JSON/YAML 编解码辅助。 |
| [`vskt`](44-vskt.md) | `github.com/imajinyun/go-knifer/vskt` | TCP socket 工具：普通连接、NIO/AIO server/client、协议编解码接口和可配置 thread-pool/listener/connection/runner/IP-parser provider。 |
| [`vslice`](45-vslice.md) | `github.com/imajinyun/go-knifer/vslice` | Slice 工具：contains/index、reverse、distinct、join、filter/map、sub-slice、concat、集合操作和分页。 |
| [`vssh`](46-vssh.md) | `github.com/imajinyun/go-knifer/vssh` | SSH/SFTP adapter 工具：可注入 provider 的命令执行、SFTP 风格列表、内存下载/上传契约、输出与传输大小限制和防御性拷贝。 |
| [`vstr`](47-vstr.md) | `github.com/imajinyun/go-knifer/vstr` | 字符串与文本工具：空白判断、裁剪、切分、截取、格式化、emoji helper、命名转换、Unicode 转义、Ant 匹配、文本相似度、SimHash、HTML 转义和 rune 检查。 |
| [`vsys`](48-vsys.md) | `github.com/imajinyun/go-knifer/vsys` | 系统与运行时信息：host、OS、user、Go runtime、进程内存、goroutine、环境变量、可重置信息缓存和 env/command/runtime provider。 |
| [`vtok`](49-vtok.md) | `github.com/imajinyun/go-knifer/vtok` | 分词 adapter 工具：可注入 provider 的文本分词与关键词提取、请求校验、输入/词元数量限制和防御性拷贝。 |
| [`vtpl`](50-vtpl.md) | `github.com/imajinyun/go-knifer/vtpl` | 模板渲染工具：支持 `html/template`、`text/template`、engine-neutral adapter、context-first 渲染、模板名、FuncMap、分隔符、factory 和 executor options。 |
| [`vurl`](51-vurl.md) | `github.com/imajinyun/go-knifer/vurl` | URL/URI 工具：解析、标准化、补全、query 编解码、百分号编码 provider、URL 构建、Data URI、协议判断、file URL 转换、资源打开/大小查询和 SSRF-oriented 安全变体。 |
| [`vver`](52-vver.md) | `github.com/imajinyun/go-knifer/vver` | 版本工具：版本比较、大小关系判断、表达式匹配、闭区间范围和自定义表达式分隔符。 |
| [`vxml`](53-vxml.md) | `github.com/imajinyun/go-knifer/vxml` | XML 工具：解析/读取/写出/格式化、树访问、XPath-style 查询、转义、map/bean 转换、transform options 和 namespace 辅助。 |
| [`vzip`](54-vzip.md) | `github.com/imajinyun/go-knifer/vzip` | ZIP、gzip、zlib 工具：归档创建/解压、条目读取、遍历、追加、内存条目、流式压缩、有边界解压/解压缩、路径穿越检查和符号链接逃逸检查。 |

<a id="quickstart-documents"></a>

## 📝 Quickstart 文档

每个包的 quickstart 示例位于上方链接的分包文档中，便于示例按领域聚焦并独立维护。

<a id="architecture-and-package-boundaries"></a>

## 🏗️ 架构与包边界

`go-knifer` 使用公开 `v*` 包作为 facade API，并将具体实现保留在 `internal/*` 中。应用代码应导入 `v*` 包；`internal/*` 用于实现演进，避免把每个 helper 都暴露成公共 API。

facade 规则：

- `internal/<domain>` 负责实现细节和领域内测试。
- `v<domain>` 暴露该领域稳定的公共 API。
- 简单工具包可以手写轻量 facade；较大的模块可以保留生成的 `facade.go`。
- internal 新增导出 API 时，应先评估是否需要暴露到 public facade。
- `vform`、`vmask`、`vsem`、`vskt`、`vblf`、`vver` 等短命名继续保留，通过模块目录解释含义，而不是破坏既有导入路径。

API 兼容：

- 根包和顶层 `v*` 子包是公共兼容边界。
- 其导出 API 面记录在 [`../api/exports.txt`](../api/exports.txt)，包括函数签名、导出类型定义、结构体字段、接口方法和方法集。
- 独立的 `internal/*` 包刻意排除在快照之外，以便实现包重构时不产生公共 API 噪声。
- `make api-check` 会重新生成临时快照并与仓库中的文件对比。
- 如果公共 API 变更是有意的，运行 `UPDATE_API=1 make api-check`，并将快照 diff 与实现变更一起 review。
- API 新增、删除和重命名在发布前也应同步到包 `doc.go` 注释、示例和 changelog。

可配置 API 与 provider 注入：

- 多个包通过 `WithXxx` helper 和 `XxxWithOptions` 变体提供 functional options。
- 配置密集型 API 可以使用显式 option struct，例如 `vconf.LoadOptions`。
- 既有固定参数 API 保持稳定，option-based 变体为需要高级控制的调用方提供扩展能力。
- Provider 风格 option 允许调用方注入文件系统函数、网络/TLS dialer 或 reader、HTTP request/multipart factory、clock、timer/ticker、随机源、DB opener、Excel workbook factory、logger、stack capture 函数、finalizer、环境变量查询、command executor、Sentry/logrus hook 等进程全局依赖，以便进行确定性测试和受控运行。
- 包级默认值保持显式。例如 HTTP 全局默认值可以读取为不可变快照，隔离请求构造器可以在不读取包级默认值的情况下构建请求。
- 配置对象在构建或加载阶段可变；发布后应视为只读快照。运行时变更前先 clone，再原子发布新指针，不要原地修改共享实例。

领域边界规则：

- `vhash` 负责非加密 hash helper；`vcrypto` 负责安全相关摘要、HMAC、加密和密钥/PEM 操作。
- `vhttp` 是轻量标准库 HTTP facade；`vresty` 是 Resty-based 链式 client facade。
- URL 转义、query 构建/解析和协议判断归 `vurl`，不由 HTTP facade 重复导出。
- `vdb` 负责 `database/sql` 之上的 SQL 数据库辅助能力；调用方通过 `*sql.DB` 和单次调用 options 控制驱动和连接池。
- `vdfa` 负责 DFA 词树匹配和文本替换；通用字符串 helper 不吸收词典匹配逻辑。
- `vid` 负责生成型标识；`vident` 负责法定身份号码和地区证件解析。
- `vcodec` 负责编码/解码算法；`vurl` 负责 URL/URI 解析、标准化、资源打开/大小检查和协议语义。
- `vjson` 负责 JSON 对象、数组、路径和轻量 XML adapter；`vxml` 负责 XML 解析、树访问、格式化、namespace 和 XML 专属 map/bean 转换。
- `vbean` 负责直接的 struct/map 属性映射，不通过 JSON 序列化绕路。
- `vobj` 是对象级便利 facade。新增领域逻辑仍应优先放入 `vstr`、`vslice`、`vmap`、`vref` 等清晰领域包，再在有价值时由 `vobj` 包装。

### 🚦 错误契约

根包 `knifer` 负责跨子包的统一错误契约：错误码分类 `ErrCode`、统一的 `knifer.Error` 类型、`CodeCarrier` 接口、`CodeOf` 提取函数，以及 `NewError` / `WrapError` / `Errorf` 构造函数。

接入该契约的子包会返回 `*knifer.Error`，或在既有 error 类型/哨兵上增加按错误码匹配能力，让调用方可以按 code 匹配或提取，同时保留 cause 链：

```go
if errors.Is(err, knifer.ErrCodeInvalidInput) { /* ... */ }
if code, ok := knifer.CodeOf(err); ok { /* ... */ }
```

`vcrypto` 是参考接入示范：校验错误同时匹配 `knifer.ErrCodeInvalidInput` 和既有 `vcrypto` 哨兵，例如 `ErrInvalidKey`、`ErrInvalidIV`、`ErrInvalidCipherText`。

### 🔐 安全与防护默认值

安全敏感工具只暴露当前推荐的公共 API 面：

- `vcrypto` 保留 SHA-2 摘要、HMAC-SHA-256/384/512、PBKDF2-SHA-256、AES-GCM、RSA-OAEP 加密和 RSA-PSS 签名。
- JWT RSA 签名通过 RSA-PSS 暴露，同时保留 HMAC 和 ECDSA signer；未签名的 JWT `alg=none` token 会被拒绝。
- TLS helper 创建的配置最低使用 TLS 1.2；便利 API 不绕过证书校验。
- HTTP 和 Resty 下载会在自动识别文件名拼接到目标目录前先校验文件名。
- 安全 HTTP/URL helper 默认拒绝本地、私有、link-local 和 unspecified 目标，并重新校验重定向目标。
- `vfile`、`vconf`、`vurl` 和 `vzip` 默认使用有边界读取或解压/解压缩限制。
- ZIP 解压会清理条目名、检查路径穿越，并解析目标父目录以拒绝符号链接逃逸。
- 以 `E` 结尾的布隆过滤器构造器会在 size 或 hash 配置非法时返回校验错误，而不是 panic。
- `vdb` 条件构造器会用 allowlist 校验操作符。
- `vskt.AioSession` 会串行化共享 session buffer 的读取，并在 close callback 期间保留 buffer。
- `vftp` 不打开网络连接、不读取凭据、不接触本地文件系统路径，也不记录传输数据；调用方通过 provider 注入，并在应用边界落实 FTP 安全策略。
- `vssh` 不打开网络连接、不执行 shell 命令、不读取凭据、不解析密钥、不接触本地文件系统路径，也不记录命令输出或传输数据；调用方通过 provider 注入，并在应用边界落实 SSH/SFTP 安全策略。
- `vhan` 不导入字典、不分词、不打开网络连接、不读取凭据、不接触本地文件系统路径，也不记录输入文本；调用方通过 provider 注入，并在应用边界负责字典和多音字行为。
- `vtok` 不导入字典、不执行内置分词、不排名关键词、不打开网络连接、不读取凭据、不接触本地文件系统路径，也不记录输入文本；调用方通过 provider 注入，并在应用边界负责分词与排名行为。

<a id="recommended-api-entry-points"></a>

## ✅ 推荐 API 入口

新代码建议优先使用这些 API。可能失败的请求 helper 应显式返回 error，而不是静默吞掉失败。

| 场景 | 推荐 API |
| --- | --- |
| 构建可信的标准库 HTTP 请求 | `vhttp.Get`、`vhttp.Post`、`vhttp.NewRequest` |
| 读取可信 HTTP 响应正文并处理错误 | `vhttp.GetStringE`、`vhttp.PostJSONE`、`vhttp.DownloadBytesE` |
| 访问用户可控或其他不可信 HTTP(S) URL | `vhttp.GetStringSafeE`、`vhttp.PostJSONSafeE`、`vhttp.DownloadBytesSafeE`、`vurl.OpenSafe` |
| 使用 Resty-backed HTTP facade | `vresty.Get`、`vresty.Post`、`vresty.GetStringE`、`vresty.PostJSONE` |
| 通过 Resty 访问不可信 URL | `vresty.GetStringSafeE`、`vresty.PostJSONSafeE`、`vresty.DownloadBytesSafeE` |
| 把用户可控 URL 下载到文件 | `vhttp.DownloadFileSafe` 或 `vresty.DownloadFileSafe` |
| 生成 secret、token、key、nonce 或 salt 字节 | `vrand.SecureBytes` |
| 创建 LRU 缓存 | `vcache.NewLRU` 或 `vcache.NewLRUWithTimeout` |
| 解析 cron 表达式 | `vcron.NewPattern` 或 `vcron.MustNewPattern` |
| 从信任边界加载远程配置 | `vconf.LoadRemoteSafe` 或 `vconf.LoadRemoteSafeWithOptions` |
| 在不引入网络 client 依赖的情况下使用可注入 FTP 契约 | `vftp.New`、`vftp.List`、`vftp.Download`、`vftp.Upload` |
| 在不引入网络 client 依赖的情况下使用可注入 SSH/SFTP 契约 | `vssh.New`、`vssh.Run`、`vssh.List`、`vssh.Download`、`vssh.Upload` |

<a id="build-test-and-release-workflow"></a>

## 📦 构建、测试与发布工作流

下载源码：

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

运行测试：

```bash
make test
```

诊断本地 Go/tooling/Git 环境，且不修改文件：

```bash
make doctor
```

检查是否存在可能污染测试或提交的无关未跟踪 Go 文件：

```bash
make worktree-check
```

可选安装本地 Git hooks，使提交前运行 `make quick-check`，推送前运行 `make full-check COVERAGE_FILE=/tmp/go-knifer-coverage.out`：

```bash
make install-hooks
```

本地运行 CI test-job 门禁。它会验证模块、vet、tidy/diff 清洁度、架构规则、race/shuffle 测试、覆盖率门禁和导出 API 快照：

```bash
make ci-test
```

校验供 Agent 与 CLI 自动化使用的机器可读 AI 元数据：

```bash
make ai-context-check
```

提交 PR 前运行与 CI 对齐的本地安全检查：

```bash
make check
```

`make check` 包含 `ci-test` 同级别检查，并额外运行 `golangci-lint` 和 `govulncheck`。

Benchmark 基线：

```bash
make bench-core
make bench-facade
make bench-core BENCHCOUNT=10 BENCHTIME=3s
```

有意修改导出 API 后刷新 API 快照：

```bash
UPDATE_API=1 make api-check
```

有意修改 facade、doc comment 或 Example 后刷新生成文档产物：

```bash
make docs-gen
make docs-check
```

确认生成结果符合预期后，运行仓库 `go:generate` 指令：

```bash
make generate
```

GitHub Actions 会复用 Makefile target 来执行模块校验、vet、tidy 检查、diff 清洁度、架构检查、race/shuffle 测试、覆盖率门禁和 API 兼容检查，同时运行 `golangci-lint`、`govulncheck` 和 CodeQL。Dependabot 已配置 Go modules 与 GitHub Actions 依赖更新。

格式化代码：

```bash
gofmt -w .
```

<a id="governance"></a>

## 🛡️ 治理

- 安全报告：参见 [`../../SECURITY.md`](../../SECURITY.md)。请不要在公开 Issue 中披露疑似漏洞。
- 发布说明：参见 [`../../CHANGELOG.md`](../../CHANGELOG.md)。面向用户的变更应在打发布标签前记录。
- 覆盖率门禁：CI 使用 `bash bin/check_coverage.sh coverage.out` 执行仓库基线。只有新增测试支撑后，才提升 `COVERAGE_THRESHOLD` 或 `PACKAGE_COVERAGE_THRESHOLDS`。
- API 门禁：`make api-check` 会将根包和顶层 `v*` 包的 API 签名、导出字段、接口方法和方法集与 [`../api/exports.txt`](../api/exports.txt) 对比。仅在有意修改公共 API 时提交刷新后的快照。
- 生成文档门禁：`make docs-check` 校验生成文档产物，包括 [`../api/tools.json`](../api/tools.json) 里的机读工具目录和 [`../api/tools.md`](../api/tools.md) 里的可读工具目录。仅当源码文档、facade 函数或 Example 有意变化时，才使用 `make docs-gen` 重新生成。
- AI 元数据门禁：`make ai-context-check` 校验 [`../../ai-context.json`](../../ai-context.json)，包括命令副作用、facade 清单、覆盖率门禁和安全敏感包引用。
- 工作流门禁：使用 `make doctor` 做环境诊断，使用 `make worktree-check` 阻止无关未跟踪 Go 文件污染测试或提交，使用 `make quick-check` 做快速本地验证，使用 `make security-check` 做 lint 与漏洞扫描，使用 `make full-check COVERAGE_FILE=/tmp/go-knifer-coverage.out` 做完整 pre-push 门禁，使用 `make ci-test` 对齐 GitHub Actions test-job。可选 Git hooks 可通过 `make install-hooks` 启用，并通过 `make uninstall-hooks` 关闭。
- 安全抑制：保持 `.golangci.yml`、`#nosec` 和 `//nolint:gosec` 例外范围足够窄，并在调用点说明原因；扩大排除前优先补充回归测试。
- Benchmark 基线：使用 `make bench-core` 确认热点工具函数 benchmark 可运行，或使用 `make bench-facade` 确认对应 public facade 包 benchmark 可运行。除非单独使用 `benchstat` 对比，否则输出只作为基线。

<a id="contributing"></a>

## 🤝 贡献

如果发现 bug 或希望补充新工具，请打开 GitHub Issue。建议提供：

- Go 版本与操作系统；
- `go-knifer` 版本或 commit；
- 最小可复现代码；
- 期望行为与实际行为；
- 相关错误日志或测试输出。

欢迎提交 Pull Request。为了保持工具库稳定，请尽量遵循以下原则：

1. 新增能力优先放入合适的 `internal/*` 实现包，再由对应 `v*` 包暴露公共 API；
2. 新增或修改公共 API 时补充必要注释；
3. 为核心逻辑补充单元测试，提交前运行 `go test ./...`；
4. 保持代码经过 `gofmt` 格式化；
5. 避免引入不必要的第三方依赖，优先复用标准库。
