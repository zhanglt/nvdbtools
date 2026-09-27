# nvdbtools 概要设计说明书

## 1. 设计目标

`nvdbtools` 采用离线、分阶段的数据加工方式，在不改变 Scanner 漏洞匹配规则的前提下，将既有 `cvedb` 中的漏洞描述中文化。设计重点是保持数据库格式兼容、允许阶段性重试，并区分 CNNVD 官方中文描述与机器翻译描述。

## 2. 总体架构

```mermaid
flowchart LR
    A[CNNVD API] -->|XML| B[cnnvd getxml]
    B --> C[/tmp/nvdbtools/xml]
    C --> D[cnnvd importDB]
    D --> E[(cnnvd.db)]

    F[Scanner Image] -->|docker cp| G[原始 cvedb]
    G --> H[cve unzip]
    H --> I[cvedbsrc: keys / *.tb]

    E --> J[cve update]
    I --> J
    K[Google Translate] --> J
    J --> L[cvedbtarget: 中文描述 *.tb]
    L --> M[cve rebuild]
    M --> N[cvedb.compact]
    M --> O[cvedb.regular]
```

主流水线由两个输入支路组成：CNNVD 支路建立“CVE ID → 中文描述”映射；Scanner 支路取得并展开原始数据库。两者在 `cve update` 汇合，最后由 `cve rebuild` 生成可消费产物。

## 3. 模块划分

| 模块 | 主要职责 | 关键文件 |
| --- | --- | --- |
| CLI 装配 | 注册根命令和子命令 | `main.go`、`cmd/root.go` |
| CNNVD 命令 | 下载 XML、导入 SQLite | `cmd/cnnvd/*.go` |
| CVE 命令 | 解包、更新、重建数据库 | `cmd/cve/*.go` |
| CNNVD 适配器 | 调用 HTTP API、解析 XML | `cnnvd/fetcher.go` |
| 描述处理 | JSONL 读取、查询/翻译描述、缓存 | `common/utils.go` |
| 数据库封装 | header 处理、解密、压缩、打包 | `common/db.go`、`common/crypto.go`、`common/memdb.go` |
| 数据模型 | full/app/header JSON 结构 | `common/types.go` |
| 自动化 | 阶段编排、镜像数据库提取 | `Makefile`、`getdbfile.sh` |
| 辅助程序 | 合规文本翻译 | `cis/compliance/` |

`common/version.go`、`common/priority.go` 和 `share/` 包含从相关项目引入的通用能力，但不处于当前主流水线的关键路径。

## 4. CLI 与目录设计

```text
nvdbtools
├── cnnvd
│   ├── getxml   --token, --savePath
│   └── importDB --filePath
└── cve
    ├── unzip    --cvedbPath, --unzipPath
    ├── update   --unzipPath, --targetPath, --proxy
    └── rebuild  --dbPath, --srcPath
```

默认目录布局：

```text
/tmp/nvdbtools/
├── cvedb                 # 从 Scanner 镜像提取的输入
├── xml/                  # CNNVD XML
├── cvedbsrc/             # 解包后的原始表
├── cvedbtarget/          # 描述更新后的表
└── cvedbtemp/
    ├── cvedb.compact
    └── cvedb.regular
```

`cnnvd.db` 位于进程当前工作目录。`InitPath()` 会先递归删除目标路径再创建，因此 CLI 层必须把该行为视为破坏性操作并限制目标范围。

## 5. 核心处理流程

### 5.1 下载与预处理

`GetIDlist()` 向列表接口提交 `{pageIndex: 1, pageSize: 100}`，取得文件 ID。`getxml` 为每个 ID 启动 goroutine，通过 `GetXml()` 下载文件，然后调用外部 `sed -i` 为 `name` 和 `vuln-descript` 节点加入 CDATA 包装。当前超时对象在整个下载批次共享。

### 5.2 XML 导入

`importDB` 打开 `cnnvd.db`，删除并重建 `cnnvd` 表，然后逐文件调用 `BuildCVE()`。解析器遍历 `<cnnvd>/<entry>`，提取 CNNVD 编号、中文描述和 CVE 编号，在单文件事务中批量插入。SQLite 设置 `synchronous=0`、`journal_mode=OFF` 以提高导入速度，但降低异常断电时的持久性保障。

### 5.3 `cvedb` 解包

外层格式如下：

```text
+----------------------+----------------------+-----------------------------+
| header length (4 B)  | KeyVersion JSON      | AES-GCM ciphertext           |
| big-endian int32     | version/time/SHA     | nonce + Gzip(Tar(*.tb))      |
+----------------------+----------------------+-----------------------------+
```

`UNzipDb()` 校验 header 长度，写出 `keys`，使用固定兼容 key 解密剩余 payload，再通过 NeuVector archive utility 自动识别 Gzip/Tar 并展开文件。

### 5.4 描述选择与记录转换

`UpdateDescription()` 逐行解析 JSONL。系统表以字段 `N` 作为 CVE ID，应用表以 `VN` 作为漏洞 ID；仅更新字段 `D`。

```text
查询 cnnvd 表
  ├─ 命中 → "cn:" + CNNVD 描述
  └─ 未命中 → 查询 translate 表
                ├─ 命中 → "ts:" + 缓存描述
                └─ 未命中 → Google Translate
                              ├─ 成功 → 缓存并返回 "ts:" + 描述
                              └─ 失败 → 保留英文原文
```

八个 distro full 表并发处理，共享一个 `*sql.DB`；`apps.tb` 随后串行处理。index 表、`rhel-cpe.map` 和原 header 文件不参与描述转换。

### 5.5 数据库重建

`RebuildDb()` 将目标表载入内存 buffer，逐文件计算 SHA256，并构造新的 `KeyVersion`。compact 仅包含 Ubuntu、Debian、CentOS、Alpine 和 application 数据；regular 包含八个 distro、application 数据和 `rhel-cpe.map`。`CreateDBFile()` 执行 Tar → Gzip → AES-GCM → header 拼装并写盘。

## 6. 数据设计

### 6.1 SQLite

```sql
cnnvd(vuln_id, vuln_descript, other_id_cve_id)
translate(cve_id, descript)
```

`cnnvd.other_id_cve_id` 和 `translate.cve_id` 用于查询。目标设计应增加唯一约束或去重策略，避免同一 CVE 多条记录产生非确定结果。

### 6.2 Scanner 表

- `*_index.tb`：漏洞与 feature 的紧凑索引；中文化流程不修改。
- `*_full.tb`：完整系统漏洞记录；更新 `D`。
- `apps.tb`：应用漏洞记录；更新 `D`。
- `rhel-cpe.map`：RHEL CPE 映射；仅进入 regular 数据库。
- `keys`：版本、更新时间、文件 SHA256 元数据。

## 7. 并发、错误与状态管理

- 下载阶段采用每文件一个 goroutine，应增加并发上限、context cancellation 和确定性的完成统计。
- full 表更新采用八个 goroutine；SQLite 连接池负责连接复用，但写翻译缓存时仍需考虑锁竞争。
- 每个命令应将错误返回 Cobra，由根命令统一设置非零退出码；库函数不应调用 `log.Fatal`。
- 输出应先写临时文件，校验成功后原子 rename，避免失败时留下半成品。
- 重建只有在所有必需输入、SHA 计算和两个输出文件均成功后才可报告成功。

## 8. 安全设计

目标安全边界包括：

1. 凭据仅通过环境变量、secret store 或交互输入获得，不写入源码和日志。
2. CNNVD HTTPS 必须校验证书，HTTP client 应配置超时且不得修改全局 transport。
3. token 不得出现在命令回显、错误消息或构建日志。
4. 解包前校验 header 长度、payload 总大小、单文件大小和成员路径；拒绝绝对路径及 `..`。
5. 默认输出目录删除前应做路径清理与允许范围校验。
6. 固定 AES key 仅用于兼容既有格式，不应被描述为机密性保护。

当前 Python 脚本中的硬编码账号信息、`verify=False`，以及 Go 中的 `InsecureSkipVerify` 均为待整改项。

## 9. 部署与运维

- 本地开发以 `go build -o nvdbtools .` 生成单二进制。
- Python/Dockerfile 仅服务 token 获取辅助流程；主 Go CLI 不依赖该容器运行。
- 完整流水线由 Makefile 串联，但生产使用宜由 CI job 显式传递每阶段产物并保存 checksum。
- 发布前应记录输入 Scanner image digest、原数据库版本、CNNVD 数据日期、翻译缓存版本、输出 SHA256 和验证结果，以支持审计和回滚。

## 10. 测试设计

测试层次应包括：

- **单元测试：** XML 字段提取、描述优先级、前缀规则、JSON 字段保持、header 编解码和非法输入。
- **组件测试：** 使用本地 HTTP server 模拟 CNNVD/翻译接口；使用 `t.TempDir()` 验证 SQLite 与文件处理。
- **数据库往返测试：** fixture `cvedb` 执行 unzip → update → rebuild → unzip，比较记录数、非描述字段和 SHA。
- **Scanner 兼容测试：** 使用目标 Scanner 版本加载 compact/regular 数据库并报告版本。
- **失败测试：** token 无效、超时、畸形 XML/JSON、缺文件、磁盘写失败、恶意 Tar 路径和翻译不可用。

当前基线为：`go build -o nvdbtools .` 可通过；`go test ./...` 因 `cnnvd/fetcher.go` 的 `go vet` 格式化告警以及 `common/updesc_test.go` 依赖不存在的固定目录而失败。该状态应在功能改造前先修复，以建立可靠回归基线。

## 11. 扩展原则

- 新数据源应实现独立 source adapter，并输出统一的 CVE ID/description 映射，不应耦合 Scanner 匹配规则。
- 新 distro 表必须同时定义 compact/regular 归属、header SHA 和目标 Scanner 兼容验证。
- 新翻译服务应通过接口注入，支持 mock、超时、重试、限流和缓存，而不是在记录循环内直接创建网络依赖。
- 数据结构扩展应优先使用保留未知字段的转换方式，避免升级 Scanner schema 后静默丢数据。
