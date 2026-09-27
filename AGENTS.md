# Repository Guidelines

## 项目结构与模块组织

`main.go` 启动 Cobra CLI，`cmd/root.go` 注册命令组。`cmd/cnnvd/` 实现 CNNVD XML 下载与导入，`cmd/cve/` 负责解包、更新和重建 NeuVector `cvedb`。数据库格式、加解密、翻译及打包逻辑位于 `common/`；CNNVD HTTP/XML 处理位于 `cnnvd/`；`share/` 提供归档辅助函数。`cis/compliance/` 是独立的合规文本翻译程序，不属于主 CLI。根目录与 `exec/` 都包含辅助脚本和 Python 依赖，修改前应确认目标打包场景。

## 构建、测试与开发命令

- `go build -o nvdbtools .`：仅构建 CLI，不运行数据流水线。
- `go test ./...`：执行全部 Go 测试。
- `gofmt -w <files>`：格式化修改过的 Go 文件；`go vet ./...` 执行补充静态检查。
- `make xml|import|getcve|unzip|update|rebuild`：分别运行一个数据处理阶段。
- `make build`：运行依赖网络、Docker 和翻译服务的完整流水线。该流程会重建 `/tmp/nvdbtools` 下的目录，执行前必须检查输入。

使用 `./nvdbtools <command> --help` 查看参数；多数阶段默认读写 `/tmp/nvdbtools/`。

## 编码风格与命名约定

遵循 Go 标准约定：使用 `gofmt` 生成的 tab 缩进；包名简短、小写；导出标识符使用 `PascalCase`，内部标识符使用 `camelCase`。Cobra 命令装配放在 `cmd/`，可复用处理逻辑放在其外。不得随意修改 `N`、`D`、`VN` 等紧凑 JSON tag，它们关系到 Scanner 数据库兼容性。错误信息应包含操作和文件上下文，避免静默降级。

## 测试规范

使用标准 `testing` 包，测试文件命名为 `*_test.go`，测试函数命名为 `TestXxx`。优先使用 `t.TempDir()`，不要依赖固定 `/tmp` 路径；解析器和记录转换应使用 table-driven tests。测试不得依赖在线 CNNVD、Google Translate、Docker、凭据或本地代理。数据库变更应验证解包记录、checksum，以及 `cvedb.compact` 和 `cvedb.regular` 两种输出。

## Commit 与 Pull Request

历史提交使用简短、祈使式中文主题，例如 `优化代码`、`增加并发执行`。每个 commit 应聚焦单一改动并说明具体行为。Pull Request 需说明受影响的流水线阶段、列出验证命令、记录数据格式或兼容性影响，并关联相关 issue。CLI 行为变化应附代表性日志；通常不需要截图。

## 安全与配置

不得提交 CNNVD 凭据、登录 token、代理 secret、生成的 `cnnvd.db` 或重建后的 CVE 数据库。通过 CLI 参数或本地环境工具传入 token，并在日志和评审内容中脱敏。下载的 XML 与已有 `cvedb` 均应按不可信输入处理。
