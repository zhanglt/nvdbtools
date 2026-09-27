# nvdbtools 详细设计说明书

## 1. 文档范围

本文在《需求规格说明书》和《概要设计说明书》的基础上，描述当前 `nvdbtools` 的包依赖、命令入口、函数职责、数据结构、文件格式、处理算法、并发模型、异常分支及测试设计。内容以现有源码为准；“改进要求”用于标识实现与可靠目标之间的差距，不表示当前已经具备该能力。

## 2. 运行时组成与依赖

### 2.1 包依赖

```text
main
└── cmd
    ├── cmd/cnnvd ──> cnnvd ──> net/http, etree, SQLite
    └── cmd/cve    ──> common ──> go-googletrans, SQLite
                                  └── github.com/neuvector/neuvector/share/utils
```

| 包/目录 | 是否在主链路 | 说明 |
| --- | --- | --- |
| `cmd/` | 是 | Cobra 命令定义、参数读取和流程编排 |
| `cnnvd/` | 是 | CNNVD HTTP 下载与 XML 入库 |
| `common/` | 是 | 描述更新、数据库格式、加解密和重建 |
| `share/` | 否 | 仓库内归档、Set、日志等通用代码；当前主链路实际引用上游 NeuVector `share/utils` |
| `cis/compliance/` | 独立入口 | 合规说明翻译，不注册到 `nvdbtools` CLI |
| `exec/` | 辅助 | Docker image 中使用的 Python token 获取脚本及依赖 |

### 2.2 进程入口

`main.go:main()` 仅调用 `cmd.Execute()`。`cmd.Execute()` 执行全局 `RootCmd`，Cobra 返回错误时调用 `os.Exit(1)`。`init()` 链负责注册 `cve`、`cnnvd` 及其子命令。

当前多数子命令使用 `Run` 而非 `RunE`；函数内部遇错后直接 `return` 时，Cobra 仍可能以退出码 0 结束。详细设计的目标约束是统一改用 `RunE`，由根命令决定退出码。

## 3. 命令详细设计

### 3.1 命令树与参数

| 命令 | 参数 | 默认值 | 输入 | 输出 |
| --- | --- | --- | --- | --- |
| `cnnvd getxml` | `--token/-t` | 空，逻辑必填 | CNNVD API | XML 文件 |
|  | `--savePath/-s` | `/tmp/nvdbtools/xml/` | 目标目录 | 同左 |
| `cnnvd importDB` | `--filePath/-f` | `/tmp/nvdbtools/xml/` | XML 目录 | `./cnnvd.db` |
| `cve unzip` | `--cvedbPath/-c` | `/tmp/nvdbtools/cvedb` | 加密 `cvedb` | 解包文件 |
|  | `--unzipPath/-u` | `/tmp/nvdbtools/cvedbsrc/` | 目标目录 | 同左 |
| `cve update` | `--unzipPath/-u` | `/tmp/nvdbtools/cvedbsrc/` | 原始 `*.tb` | 中文化 `*.tb` |
|  | `--targetPath/-t` | `/tmp/nvdbtools/cvedbtarget/` | 目标目录 | 同左 |
|  | `--proxy/-p` | `http://127.0.0.1:10809` | 翻译代理 | 无 |
| `cve rebuild` | `--srcPath/-s` | `/tmp/nvdbtools/cvedbtarget/` | 更新后文件 | 新数据库 |
|  | `--dbPath/-d` | `/tmp/nvdbtools/cvedbtemp/` | 输出目录 | compact/regular |

### 3.2 `cnnvd getxml`

#### 3.2.1 前置条件

- token 非空且仍有效。
- CNNVD API 可访问。
- 本机存在 `sed`。
- `savePath` 可删除和创建。

#### 3.2.2 调用链

```text
getxmlCmd.Run
├── common.InitPath(savePath)
├── cnnvd.GetIDlist()
│   └── POST getPageList {pageIndex:1,pageSize:100}
├── goroutine × fileID
│   └── cnnvd.GetXml(fileID, token, savePath)
└── sed -i <CDATA rule> downloaded.xml
```

#### 3.2.3 算法

1. 读取 token；为空时记录日志并返回。
2. `InitPath()` 对下载目录执行 `RemoveAll`，再以 `0755` 创建。
3. `GetIDlist()` 将 `Rplist{Index:1, Size:100}` 序列化为 JSON，通过 POST 获取列表。
4. 为每个 ID 启动一个 goroutine；每个 goroutine 下载到 `<savePath>/<id>.xml` 并把 ID 写入无缓冲 channel。
5. 主 goroutine 创建一个共享的 900 秒 timer，按列表长度接收结果。
6. 每收到一个结果，使用 `sed -i` 为 `name`、`vuln-descript` 内容添加 CDATA 边界。

#### 3.2.4 当前异常行为

- `GetIDlist()` 修改 `http.DefaultTransport` 并设置 `InsecureSkipVerify`，影响进程内其他 HTTP 请求。
- HTTP 请求未设置 context、client timeout，也未校验 status code。
- `GetXml()` 失败时返回空字符串，调用方可能处理 `<savePath>/.xml`。
- 文件以 append 模式打开；当前依赖目录预清理避免重复内容。
- timer 触发后的 `break` 只退出 `select`，不退出外层循环；timer channel 只发送一次，剩余循环可能永久等待。
- 异步完成顺序与 `list[idx]` 不一致，完成日志中的源 ID 可能错配。
- goroutine 数量等于文件数，没有并发上限。

改进实现应采用 `context.WithTimeout`、受限 worker pool、结构化结果 `{ID, Path, Err}`，并在所有 worker 退出后汇总成功/失败数。

### 3.3 `cnnvd importDB`

#### 3.3.1 数据库初始化

`getcveDB()` 使用 SQLite driver 打开当前目录的 `cnnvd.db`，执行：

```sql
DROP TABLE IF EXISTS cnnvd;
CREATE TABLE cnnvd (
    vuln_id          varchar(255),
    vuln_descript    text,
    other_id_cve_id  varchar(255)
);
CREATE INDEX cnnvd_other_id_cve_id_IDX
    ON cnnvd(other_id_cve_id);
PRAGMA synchronous = 0;
PRAGMA journal_mode = OFF;
```

`translate` 表不在此处删除，因此同一个数据库文件可保留历史机器翻译缓存。

#### 3.3.2 XML 字段映射

| XML 路径 | SQLite 字段 | 用途 |
| --- | --- | --- |
| `cnnvd/entry/vuln-id` | `vuln_id` | CNNVD 标识 |
| `cnnvd/entry/vuln-descript` | `vuln_descript` | 中文描述 |
| `cnnvd/entry/other-id/cve-id` | `other_id_cve_id` | 与 Scanner 记录关联的 CVE ID |

`BuildCVE()` 使用 `etree` 解析单个文件，对文件开启一个 transaction，逐 entry 执行参数化 insert，最后 commit。

#### 3.3.3 当前异常行为

- `ioutil.ReadDir()` 返回的所有条目都会被解析，未过滤扩展名和目录。
- XML 根节点不存在时，`root.SelectElements()` 可能因 nil 引用崩溃。
- `id`、`description`、`vulnid` 定义在 entry 循环外；缺失节点可能继承上一条记录的值。
- `Begin`、`Exec`、`Commit` 错误被忽略，日志中的“导入完成”不代表事务成功。
- `getcveDB()` 中 `DB.Ping()` 位于 `return` 之后，不可达。
- `synchronous=0` 与 `journal_mode=OFF` 优先性能而牺牲故障恢复能力。

改进实现应逐 entry 清空字段、校验 CVE ID、使用 prepared statement，并在任一 insert 失败时 rollback 当前文件。

### 3.4 `cve unzip`

#### 3.4.1 调用链

```text
unzipCmd.Run
├── common.InitPath(unzipPath)
└── common.UNzipDb(cvedbPath, unzipPath)
    ├── 读取 header length
    ├── 写出 keys
    ├── common.decrypt(cipherData, zeroKey)
    └── upstream utils.ExtractAllArchiveToFiles(...)
```

#### 3.4.2 二进制读取规则

1. 读取前 4 字节并按 big-endian 解析为 `int32 headLen`。
2. 拒绝 `headLen > 100 KiB`；当前实现未拒绝负数或零值。
3. 读取 `headLen` 字节作为 `KeyVersion` JSON，并以权限 `0400` 写为 `keys`。
4. 读取文件剩余部分作为 AES-GCM 密文。
5. 密文前 `gcm.NonceSize()` 字节为 nonce，剩余为 ciphertext/tag。
6. 解密结果为 Gzip 压缩的 Tar archive，由上游 utility 自动识别并展开。

当前 `maxExtractSize=0` 表示不限制单个成员大小；目标实现应同时限制总 payload、单文件大小和文件数量，并拒绝绝对路径、`..` 与目标目录逃逸。

### 3.5 `cve update`

#### 3.5.1 输入集合

```text
full:
  alpine_full.tb    amazon_full.tb   centos_full.tb
  debian_full.tb    mariner_full.tb  oracle_full.tb
  suse_full.tb      ubuntu_full.tb

index（原样复制）:
  alpine_index.tb   amazon_index.tb  centos_index.tb
  debian_index.tb   mariner_index.tb oracle_index.tb
  suse_index.tb     ubuntu_index.tb

other:
  apps.tb           rhel-cpe.map     keys
```

#### 3.5.2 SQLite 缓存

`getDB()` 打开 `cnnvd.db` 并尝试创建：

```sql
CREATE TABLE IF NOT EXISTS translate (
    cve_id   varchar(255),
    descript text
);
CREATE INDEX translate_cve_id_IDX ON translate(cve_id);
```

当前 index DDL 缺少 `IF NOT EXISTS`；第二次运行可能返回“index already exists”。命令仅记录该错误并继续。

#### 3.5.3 并发模型

命令为八个 full 表各启动一个 goroutine，并共享同一个 `*sql.DB`。每个 `UpdateDescription()` 创建自己的 Google translator client。`WaitGroup` 等待全部 full 表完成后，主 goroutine 串行处理 `apps.tb`，最后复制 index、CPE 和 header 文件。

#### 3.5.4 逐行转换

full 表使用 `Centos` 结构承载所有 distro 记录，以 `N` 作为查询 ID；application 表使用 `Apps`，以 `VN` 作为查询 ID。Scanner buffer 容量设置为默认 token 大小的 10 倍。

```text
function resolveDescription(cveID, sourceDescription):
    row = SELECT vuln_descript FROM cnnvd WHERE other_id_cve_id = cveID
    if row exists:
        return "cn:" + row.vuln_descript

    row = SELECT descript FROM translate WHERE cve_id = cveID
    if row exists:
        return "ts:" + row.descript

    translated = GoogleTranslate(sourceDescription, "en", "zh")
    if translated failed or empty:
        return sourceDescription

    INSERT INTO translate(cve_id, descript) VALUES(cveID, translated)
    return "ts:" + translated
```

#### 3.5.5 当前数据完整性风险

- 输出用 `O_APPEND|O_CREATE` 打开且无 `O_TRUNC`；安全性依赖 `targetPath` 先被删除。
- JSON unmarshal 失败后仍会 marshal 零值结构并写出，可能把坏输入转换成空记录。
- 固定 `Centos`/`Apps` 结构无法保留未来 schema 的未知字段。
- Scanner、marshal、writer、transaction 及 SQL 错误多处未检查。
- `CopyFile()` 打开源文件失败后继续 defer nil file，可能 panic；目标文件也未 truncate。
- `log.Fatal` 在 worker goroutine 内会直接结束整个进程，其他 writer 无清理机会。
- 全局 `tindex` 被多个 goroutine 无同步写入，存在 data race。
- 多 goroutine 首次翻译同一 CVE 时可能重复请求并竞争 SQLite 写锁。

目标实现应采用临时文件、原子 rename、原始 JSON map 的字段保留策略、单独的 translation service/cache 层，以及明确的 worker error channel。

### 3.6 `cve rebuild`

#### 3.6.1 初始化

`rebuildCmd` 首先清空 `dbPath`，调用 `MemdbOpen()` 创建 `memDB` 和临时目录，再由 `getVersion(<srcPath>/keys)` 反序列化 `KeyVer` 并读取原版本号。

#### 3.6.2 内存布局

```go
type dbBuffer struct {
    namespace string
    indexFile string
    fullFile  string
    indexBuf  bytes.Buffer
    fullBuf   bytes.Buffer
    indexSHA  [32]byte
    fullSHA   [32]byte
}

type dbSpace struct {
    buffers [8]dbBuffer
    appBuf  bytes.Buffer
    appSHA  [32]byte
    rawSHA  [][32]byte
}
```

八个 slot 的固定顺序为 Ubuntu、Debian、CentOS、Alpine、Amazon、Oracle、Mariner、SUSE。`loadDbs()` 将每个文件完整读入内存并计算 SHA256；`apps.tb` 单独存储，`rhel-cpe.map` 作为 raw file。

#### 3.6.3 输出成员矩阵

| 成员 | compact | regular |
| --- | :---: | :---: |
| Ubuntu index/full | 是 | 是 |
| Debian index/full | 是 | 是 |
| CentOS index/full | 是 | 是 |
| Alpine index/full | 是 | 是 |
| Amazon index/full | 否 | 是 |
| Oracle index/full | 否 | 是 |
| Mariner index/full | 否 | 是 |
| SUSE index/full | 否 | 是 |
| `apps.tb` | 是 | 是 |
| `rhel-cpe.map` | 否 | 是 |

#### 3.6.4 Header 构造

```go
type KeyVersion struct {
    Version    string
    UpdateTime string
    Keys       map[string]string
    Shas       map[string]string
}
```

- `Version`：沿用输入 `keys` 的版本。
- `UpdateTime`：重建时的 `time.Now().Format(time.RFC3339)`。
- `Shas`：仅包含当前输出 archive 的成员，值为小写 hex SHA256。
- `Keys`：当前来自新建 `memDB.keyVer.Keys`，实际为空 map，并未从输入 header 恢复。

#### 3.6.5 封装算法

```text
files
  → MakeTar(files)
  → GzipBytes(tar)
  → AES-256-GCM(zeroKey, randomNonce, gzip)
  → int32_be(len(headerJSON)) || headerJSON || ciphertext
  → os.Create(output)
```

`CreateDBFile()` 使用随机 nonce，因此相同输入每次产生不同 ciphertext；header 与成员内容仍可单独验证。

#### 3.6.6 当前成功判定风险

- `loadDbs()` 对缺失文件只记录日志并继续，最终始终返回 true；缺失成员会以空内容打包。
- 两次 `CreateDBFile()` 的错误被忽略，`RebuildDb()` 最终仍返回 true。
- 所有输入完整载入内存，峰值内存约为输入数据总量加 Tar/Gzip/密文副本。
- 读取输入 header 只保留版本，原 `Keys` 及其他潜在元数据不会保留。

目标成功条件应为：全部必需文件读取成功、成员 SHA 正确、两个输出均写入并可重新解包，否则整体失败。

## 4. 数据模型详细设计

### 4.1 Scanner full 记录

`Centos`/`VulFull` 映射的关键紧凑字段如下：

| JSON tag | 语义 | 中文化是否修改 |
| --- | --- | :---: |
| `N` | CVE/漏洞名称 | 否 |
| `NS` | distro namespace | 否 |
| `D` | description | 是 |
| `L` | reference link | 否 |
| `S` | severity | 否 |
| `C2`, `C3` | CVSS v2/v3 | 否 |
| `FB`, `FI` | fixed-by/fixed-in | 否 |
| `CPE`, `CVE` | CPE 与别名 | 否 |
| `Issue`, `LastMod` | 时间 | 否 |

### 4.2 Application 记录

| JSON tag | 语义 | 中文化是否修改 |
| --- | --- | :---: |
| `VN` | 漏洞名称/查询 ID | 否 |
| `AN`, `MN` | application/module | 否 |
| `D` | description | 是 |
| `AV`, `FV`, `UV` | affected/fixed/unaffected versions | 否 |
| `SC`, `SC3`, `VV2`, `VV3` | CVSS 数据 | 否 |
| `SE` | severity | 否 |

### 4.3 描述来源标记

- `cn:`：描述来自 CNNVD XML。
- `ts:`：描述来自 Google Translate 或本地翻译缓存。
- 无前缀：翻译失败后保留原描述，或输入原本无可用映射。

前缀属于当前实现约定，会改变最终展示文本；消费方如不希望显示来源标记，应在独立元数据字段中表达，而不是继续扩展字符串前缀。

## 5. 文件系统与资源生命周期

| 资源 | 创建者 | 关闭/清理方式 | 当前风险 |
| --- | --- | --- | --- |
| HTTP response body | `GetIDlist`/`GetXml` | `defer Close()` | 无 status/size/timeout 校验 |
| XML 文件 | `GetXml` | `defer Close()` | append 写、写错误忽略 |
| SQLite DB | command | 当前未显式 `Close()` | 进程退出回收，测试/长任务不理想 |
| SQLite transaction | `BuildCVE`/`getDescribe` | `Commit()` | 无 rollback、错误忽略 |
| `cvedb` 输入 | `UNzipDb` | `defer Close()` | header 边界不足 |
| 临时目录 | `MemdbOpen` | `memDB.Close()` 删除 | rebuild 正常 defer 清理 |
| 输出文件 | update/rebuild | `defer Close()` | 非原子写、部分错误忽略 |

所有显式创建的资源应在最小作用域关闭；数据库和 response body 应确保异常路径也能释放。输出目录的清理由命令层负责，库函数不应隐式删除调用方数据。

## 6. 错误码与日志设计

当前实现没有稳定的业务错误码。详细设计建议定义阶段错误并通过 `%w` 包装根因：

| 错误类别 | 建议错误标识 | CLI 退出码 |
| --- | --- | ---: |
| 参数无效 | `ErrInvalidArgument` | 2 |
| 输入不存在/格式错误 | `ErrInvalidInput` | 3 |
| CNNVD/翻译网络错误 | `ErrExternalService` | 4 |
| SQLite 错误 | `ErrDatabase` | 5 |
| 解密/解包错误 | `ErrDecodeCVEDB` | 6 |
| 写入/打包错误 | `ErrBuildCVEDB` | 7 |

日志至少包含 `stage`、`file`、`cve_id`（适用时）和根因，但不得包含账号、密码、token、完整 HTTP body。成功日志必须在持久化、flush、close 和校验完成后输出。

## 7. 安全详细设计

### 7.1 凭据

- token 通过环境变量、stdin 或权限受控文件传入；避免出现在 process list。
- Python 辅助脚本不得包含固定账号密码，也不得输出 verify token 或服务端完整响应。
- proxy URL 若包含认证信息，日志中必须脱敏。

### 7.2 网络

- 使用专用 `http.Client`，配置连接、TLS handshake、response header 和总请求超时。
- 保持系统 CA 校验；如必须使用私有 CA，通过可配置 CA bundle 注入。
- 校验 2xx status、`Content-Type` 和最大 response body。

### 7.3 文件与归档

- 对删除目标执行 `filepath.Clean`，拒绝空路径、`.`、`/`、用户目录和非允许前缀。
- Tar 成员使用 `filepath.Join` 后确认仍位于目标根目录。
- 输出文件使用 `0600` 或按部署要求明确配置权限。
- 对 XML 设置输入大小和 entry 数限制，避免资源耗尽。

### 7.4 加密边界

固定零值 AES key 用于兼容 Scanner `cvedb` 格式，不能提供秘密性；安全设计只能将其视为带完整性校验的格式封装。真正的分发安全应由制品签名、受控 registry、checksum 和访问控制承担。

## 8. 可观测性设计

每个阶段应输出以下统计：

- CNNVD：列表条目数、下载成功/失败数、XML entry 数、有效 CVE ID 数。
- update：总记录数、CNNVD 命中数、缓存命中数、新翻译数、翻译失败数、JSON 错误数。
- rebuild：每个成员的记录数、字节数、SHA256、最终文件大小及数据库版本。

统计只记录聚合数据，不记录描述正文或凭据。完整流水线应生成机器可读 manifest，关联输入 Scanner image digest、输入/输出数据库 checksum 和执行时间。

## 9. 测试用例设计

### 9.1 单元测试矩阵

| 编号 | 对象 | 场景 | 预期 |
| --- | --- | --- | --- |
| UT-01 | `BuildCVE` | 完整 entry | 三字段正确入库 |
| UT-02 | `BuildCVE` | 缺 CVE/描述 | 不继承上一条字段，按策略跳过或报错 |
| UT-03 | `getDescribe` | CNNVD 命中 | 返回 `cn:`，不访问翻译服务 |
| UT-04 | `getDescribe` | cache 命中 | 返回 `ts:`，不访问翻译服务 |
| UT-05 | `getDescribe` | 新翻译成功 | 入 cache 并返回 `ts:` |
| UT-06 | `getDescribe` | 翻译失败 | 原描述保持不变 |
| UT-07 | `UpdateDescription` | 畸形 JSONL | 返回错误且不写空记录 |
| UT-08 | `encrypt/decrypt` | 正常 round trip | 明文一致 |
| UT-09 | `decrypt` | 短密文/篡改 tag | 返回错误 |
| UT-10 | header parser | 负数/超限/截断 | 安全拒绝，无 panic |
| UT-11 | `CopyFile` | 源缺失 | 返回错误，无 panic |
| UT-12 | archive extractor | `../` 成员 | 拒绝路径穿越 |

### 9.2 组件与集成测试

1. 使用 `httptest.Server` 模拟 CNNVD 列表、下载、401、500、慢响应和超大响应。
2. 使用临时 SQLite 验证导入重跑、重复 CVE 和并发 cache 写入。
3. 使用 fixture 数据执行 `unzip → update → rebuild → unzip` 往返测试。
4. 比较往返前后记录总数及除 `D` 外的所有 JSON 字段。
5. 校验 compact/regular 成员矩阵和 header SHA。
6. 用目标 Scanner 加载两个数据库并读取版本，作为发布门禁。

### 9.3 当前测试基线

`common/updesc_test.go:TestCopyFile` 使用固定 `/tmp/source/file.txt`，没有创建父目录，也没有实际调用 `CopyFile()`；该测试当前必然可能失败且无法覆盖被测函数。应改用 `t.TempDir()`，显式调用 `CopyFile()`，并增加源不存在、覆盖已有目标等分支。

`go test ./...` 还会运行默认 `go vet`，当前 `cnnvd/fetcher.go` 将格式字符串传给 `log.Println/Fatalln`，会触发格式化告警。建立回归基线前应使用 `Printf/Fatalf` 或结构化日志修复。

## 10. 需求追踪

| 需求 | 设计实现点 | 验证重点 |
| --- | --- | --- |
| FR-01 | `main.go`、`cmd/root.go` | 命令树、退出码 |
| FR-02 | `cmd/cnnvd/getxml.go`、`cnnvd.GetIDlist/GetXml` | token、超时、下载完整性 |
| FR-03 | `cmd/cnnvd/importDB.go`、`cnnvd.BuildCVE` | XML 映射、事务、重跑 |
| FR-04 | `getdbfile.sh` | image/container 生命周期 |
| FR-05 | `cmd/cve/unzip.go`、`common.UNzipDb/decrypt` | header 边界、解密、归档安全 |
| FR-06 | `cmd/cve/update.go`、`common.UpdateDescription/getDescribe` | 描述优先级、字段保持 |
| FR-07 | `cmd/cve/rebuild.go`、`common.RebuildDb/CreateDBFile` | 成员矩阵、SHA、双输出 |
| FR-08 | `Makefile` | 单阶段重试、完整流水线 |
| FR-09 | `cis/compliance/main.go` | Description/Remediation 翻译 |

## 11. 实现优先级建议

1. 清除源码凭据并恢复 TLS 校验。
2. 修复测试基线、错误传播和虚假成功。
3. 防止解包路径穿越、无限大小及危险目录删除。
4. 保证 JSON 非描述字段完整保留，并改为原子输出。
5. 修复 SQLite transaction、重复 index、并发 cache 和 data race。
6. 增加 manifest、统计和 Scanner 兼容性集成测试。

以上改进不改变产品边界：本工具仍只负责已有 Scanner 数据库的描述中文化和兼容重封装，不负责新增漏洞检测规则。
