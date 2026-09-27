# nvdbtools 需求规格说明书

## 1. 文档目的

本文根据当前源码反向整理 `nvdbtools` 的业务目标、功能边界、接口、数据约束和验收标准。本文描述的是 Scanner CVE 数据库中文化工具，不代表 NeuVector Scanner 的漏洞检测需求。

## 2. 产品范围

### 2.1 目标

系统应从 CNNVD 获取 CVE 中文描述，将其导入本地 SQLite 数据库，并用这些描述更新已有 NeuVector Scanner `cvedb` 中的漏洞说明；CNNVD 无对应记录时，可调用 Google Translate 生成中文说明。处理完成后，系统应输出与 Scanner 数据格式兼容的 `cvedb.compact` 和 `cvedb.regular`。

### 2.2 范围外能力

- 不执行镜像、容器或主机漏洞扫描。
- 不生成新的 CVE、受影响版本、修复版本、CPE 或软件包匹配规则。
- 不保证导入 CNNVD 数据后增加 Scanner 漏洞覆盖率。
- 不负责将重建数据库发布到镜像仓库或生产环境。

## 3. 用户与运行环境

主要用户为维护 Scanner 漏洞数据库的工程人员。运行环境应具备 Go 1.19 或兼容版本；完整流水线还依赖 Docker、可访问 CNNVD 和 Google Translate 的网络、CNNVD 登录 token，以及可选 HTTP proxy。默认工作目录位于 `/tmp/nvdbtools/`。

## 4. 功能需求

### FR-01 命令行入口

系统应提供 `nvdbtools` Cobra CLI，并包含 `cnnvd`、`cve` 两组命令。无子命令或参数不完整时，应显示帮助或明确错误，不应报告虚假成功。

### FR-02 CNNVD XML 下载

`nvdbtools cnnvd getxml` 应：

1. 接收必填 `--token/-t` 和可选 `--savePath/-s`，默认输出到 `/tmp/nvdbtools/xml/`。
2. 调用 CNNVD 文件列表接口取得下载 ID，再并发下载 XML。
3. 以请求头 `token` 访问下载接口，并在 900 秒总体超时内完成处理。
4. 对 `<vuln-descript>` 和 `<name>` 内容进行 XML 预处理。
5. 执行前初始化输出目录；调用方必须知晓该操作会删除目录原内容。

### FR-03 CNNVD 数据导入

`nvdbtools cnnvd importDB` 应顺序解析指定目录内的 XML，将 `vuln-id`、`vuln-descript`、`other-id/cve-id` 写入当前目录下的 `cnnvd.db`。每次导入应重建 `cnnvd` 表，并为 CVE 编号创建索引；`translate` 缓存表应保留。

### FR-04 Scanner 数据库获取

`getdbfile.sh` 应从本地 `neuvector/scanner:latest` 创建临时容器，将 `/etc/neuvector/db/cvedb` 复制到 `/tmp/nvdbtools/cvedb`，随后清理临时容器。本地镜像不存在时，可尝试拉取镜像并明确报告结果。

### FR-05 数据库解包

`nvdbtools cve unzip` 应读取输入文件的 4 字节大端 header 长度和 JSON header，使用兼容密钥执行 AES-GCM 解密，并将 Gzip/Tar 内容展开到目标目录。header 最大允许 100 KiB；应输出 `keys` 文件及内部数据文件。

### FR-06 描述更新

`nvdbtools cve update` 应处理八个 distro 的 `*_full.tb` 与 `apps.tb`：

1. 按 CVE ID 优先查询 `cnnvd` 表，命中结果添加 `cn:` 前缀。
2. 未命中时查询 `translate` 缓存，命中结果添加 `ts:` 前缀。
3. 缓存未命中时翻译原描述，将成功结果写入缓存并添加 `ts:` 前缀。
4. 翻译失败时保留原描述。
5. 原样复制八个 `*_index.tb`、`rhel-cpe.map` 和 `keys`。
6. 初始化目标目录后才写入，避免历史内容与新内容混合。

描述更新不得有意改变漏洞编号、namespace、严重级别、评分、版本条件或 CPE 信息。

### FR-07 数据库重建

`nvdbtools cve rebuild` 应从 `keys` 读取原数据库版本，重新计算内容 SHA256，并更新时间戳。输出要求如下：

- `cvedb.compact`：Ubuntu、Debian、CentOS、Alpine 的 index/full 表和 `apps.tb`。
- `cvedb.regular`：八个 distro 的 index/full 表、`apps.tb` 和 `rhel-cpe.map`。
- 文件封装格式：4 字节大端 JSON header 长度、JSON header、AES-GCM 加密的 Gzip/Tar payload。

### FR-08 流水线编排

Makefile 应支持单阶段目标 `xml`、`import`、`getcve`、`unzip`、`update`、`rebuild`，以及完整的 `make build` 流程。各阶段应允许使用中间产物单独重试。

### FR-09 合规文本翻译

`cis/compliance` 可作为独立辅助程序，逐行翻译 `Description:` 和 `Remediation:` 内容。该能力不属于主 CLI 数据库流水线。

## 5. 外部接口

| 接口 | 方法/形式 | 用途 |
| --- | --- | --- |
| CNNVD 文件列表 | `POST /web/vulDataDownload/getPageList` | 获取 XML 文件 ID |
| CNNVD 文件下载 | `POST /web/vulDataDownload/download` | 使用 token 下载 XML |
| Google Translate | `go-googletrans` | CNNVD 缺失时翻译英文描述 |
| Docker Engine | CLI | 创建 Scanner 容器并复制 `cvedb` |
| SQLite | `cnnvd.db` | 保存 CNNVD 描述及翻译缓存 |

## 6. 数据与兼容性要求

- `*.tb` 为每行一个 JSON 对象的文本文件；紧凑字段名如 `N`、`D`、`VN`、`NS` 不得改变。
- 更新前后记录数及非描述字段应一致；未知字段不得因反序列化/序列化而丢失。
- index 与 full 表、header 中的 SHA256 必须一致。
- 输出应能被目标版本 Scanner 解密、展开和加载。

## 7. 非功能需求

- **可靠性：** 任一必需输入缺失、XML/JSON 解析失败、翻译失败或打包失败时，应返回非零状态并指出阶段、文件和原因。
- **可恢复性：** 失败后应保留已完成阶段的有效中间产物；重试不得追加重复记录。
- **性能：** XML 下载和 distro full 表更新可并发；SQLite 写入应避免锁冲突。
- **安全性：** 不得在源码或日志中保存账号、密码、token；生产路径应启用 TLS 校验。归档展开必须防止路径穿越并设置合理大小限制。
- **可测试性：** 单元测试不得依赖公网、Docker、固定 `/tmp` 目录或真实凭据。

## 8. 当前实现约束与风险

以下为源码现状，不应视为目标设计：

- Python token 脚本包含硬编码登录信息并关闭 TLS 校验；Go HTTP 客户端也关闭了 TLS 校验。
- CVE 数据库使用源码内固定零值 AES key，属于格式兼容机制而非安全密钥管理。
- 解包大小限制为 `0`，归档文件名缺少充分的路径穿越校验。
- 多处错误仅记录日志或被忽略；重建函数可能在单个输出失败时仍返回成功。
- `UpdateDescription` 以 append 模式打开文件，依赖调用前清空目标目录。
- JSON 更新使用固定结构，未来新增字段可能被丢弃。
- 当前 `go test ./...` 会因 `go vet` 日志格式告警及固定 `/tmp/source` 测试路径失败。

## 9. 验收标准

1. `go build -o nvdbtools .` 成功，CLI 帮助列出所有预期命令。
2. 使用离线 fixture 可完成 XML 导入，并按 CVE ID 得到正确中文描述。
3. 对固定测试 `cvedb` 执行 unzip/update/rebuild 后，记录数和非描述字段保持一致。
4. CNNVD 命中、缓存命中、在线翻译成功、翻译失败四条路径均有测试。
5. 两种输出的 header、SHA256、Tar 成员和 Scanner 加载均通过校验。
6. 所有失败路径返回非零状态，日志不包含凭据。
