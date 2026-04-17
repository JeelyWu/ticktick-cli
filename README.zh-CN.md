# tick

**中文** | [English](README.md)

[![Unit Tests](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml/badge.svg?branch=master)](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml)

`tick` 是一个基于 Go 编写的 CLI，用来通过官方 Open API 操作 TickTick 国际版和滴答清单。

它覆盖了一条比较实用的日常工作流：

- 完成一次认证后长期使用
- 查看项目和任务
- 创建、更新、移动、完成任务
- 用 quick add 语法快速录入任务
- 通过本地配置保存默认值，减少重复输入

## 功能概览

- 同时支持 TickTick 国际版和滴答清单中国大陆版
- OAuth 登录，优先自动接收 localhost 回调
- 支持远程机器和 SSH 场景下手动粘贴回调 URL
- 项目命令：`add`、`get`、`ls`、`rm`、`update`
- 任务命令：`add`、`get`、`ls`、`update`、`move`、`done`、`rm`
- 便捷命令：`today`、`inbox`、`quick add`
- 支持本地默认配置：输出格式、默认项目、收件箱项目、服务区域

## 安装

### GitHub Releases

每个打了 tag 的版本都会在 GitHub Releases 发布对应平台的压缩包：

- macOS: `darwin/arm64`、`darwin/amd64`
- Linux: `linux/arm64`、`linux/amd64`
- Windows: `windows/amd64`

发布页面：

```text
https://github.com/JeelyWu/ticktick-cli/releases
```

当前每个压缩包里包含：

- 可执行文件，Windows 下是 `tick.exe`
- `README.md`

你可以手动下载对应平台的压缩包，解压后把可执行文件放到 `PATH` 里的某个目录。

示例：

```bash
tar -xzf tick_0.0.1_linux_amd64.tar.gz
install -m 0755 tick /usr/local/bin/tick
```

```bash
tar -xzf tick_0.0.1_darwin_arm64.tar.gz
install -m 0755 tick "$HOME/.local/bin/tick"
```

Windows 下解压 `tick_0.0.1_windows_amd64.zip`，把 `tick.exe` 放到你的 `PATH` 中即可。

### macOS 和 Linux 安装脚本

仓库自带一个 Unix 安装脚本：

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | bash
```

安装指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | \
  VERSION=v0.0.1 bash
```

安装到自定义目录：

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | \
  VERSION=v0.0.1 INSTALL_DIR="$HOME/.local/bin" bash
```

如果你不想用管道直接交给 `bash`，也可以先下载 [scripts/install.sh](scripts/install.sh)，再本地执行。

### 从源码构建

要求：

- 已安装 Go，版本以 [go.mod](go.mod) 为准

构建本地二进制：

```bash
make build
```

生成的文件在 `bin/tick`。

## 前置准备

登录前，需要先在对应服务上创建一个开发者应用：

- TickTick 国际版：`https://developer.ticktick.com/manage`
- 滴答清单中国大陆版：`https://developer.dida365.com/manage`

你需要拿到这些配置：

- `client_id`
- `client_secret`
- redirect URL

本地开发推荐使用这个 redirect URL：

```text
http://localhost:14573/callback
```

`tick` 通过环境变量读取 client secret：

```bash
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
```

## 首次配置

### 1. 选择服务区域

`tick` 默认使用 `ticktick`。如果你使用滴答清单中国大陆版，需要切换到 `dida365`。

```bash
tick config set service.region ticktick
```

```bash
tick config set service.region dida365
```

查看当前配置：

```bash
tick config get service.region
```

如果你已经登录过，又要切换区域，先清掉旧 token，再重新认证：

```bash
tick auth logout
tick auth login --client-id YOUR_CLIENT_ID --redirect-url http://localhost:14573/callback
```

### 2. 执行登录

TickTick 国际版：

```bash
tick config set service.region ticktick
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

滴答清单：

```bash
tick config set service.region dida365
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

### 3. 检查认证状态

```bash
tick auth status
```

也可以同时查看版本和当前区域：

```bash
tick version --verbose
```

## 本地与远程机器上的登录行为

本地开发时，`tick auth login` 会优先尝试自动接收 localhost 回调。

如果浏览器无法访问当前机器的回调地址，比如你在远程机器或 SSH 会话中执行命令，`tick` 会退回到手动粘贴回调 URL 的模式，并输出：

```text
Paste the full callback URL:
```

这时把浏览器地址栏中的完整回调地址复制出来，例如：

```text
http://localhost:14573/callback?code=abc123&state=xyz456
```

再把整段 URL 粘贴回终端即可。

## 常用命令

顶层帮助：

```bash
tick auth --help
tick project --help
tick task --help
tick quick --help
tick config --help
tick today --help
tick inbox --help
tick version --help
```

### 项目命令

列出所有项目：

```bash
tick project ls
```

按精确名称或 ID 查看一个项目：

```bash
tick project get Work
```

创建一个项目：

```bash
tick project add Work
```

创建一个笔记类型项目，并指定颜色：

```bash
tick project add Notes --kind NOTE --color '#F18181'
```

### 任务查询

列出所有未完成任务：

```bash
tick task ls
```

查看某个项目下的任务：

```bash
tick task ls --project Work
```

只看逾期任务：

```bash
tick task ls --project Work --overdue
```

查看今天到期或已逾期的任务：

```bash
tick task ls --today
```

按状态过滤：

```bash
tick task ls --status completed
```

按优先级过滤：

```bash
tick task ls --priority 5
```

按日期范围过滤：

```bash
tick task ls --from 2026-04-01 --to 2026-04-30
```

输出 JSON：

```bash
tick task ls --json
```

等价的显式写法：

```bash
tick task ls --output json
```

### 创建与更新任务

在某个项目中创建任务：

```bash
tick task add "Write spec" --project Work --due 2026-04-20
```

创建全天任务：

```bash
tick task add "Review roadmap" --project Work --due 2026-04-20 --all-day
```

创建高优先级任务，并附带描述和内容：

```bash
tick task add "Ship v0.0.2" \
  --project Work \
  --priority 5 \
  --desc "Publish binaries and verify release assets" \
  --content "Double-check the GitHub Release page"
```

按精确标题或 ID 查看任务：

```bash
tick task get "Write spec"
```

更新标题和截止日期：

```bash
tick task update "Write spec" --title "Write detailed spec" --due 2026-04-21
```

把任务移动到另一个项目：

```bash
tick task move "Write spec" --to Personal
```

如果多个项目里有同名任务，指定来源项目：

```bash
tick task move "Write spec" --project Work --to Personal
```

把任务标记为完成：

```bash
tick task done "Write spec"
```

删除任务：

```bash
tick task rm "Write spec"
```

### 便捷命令

查看今天到期或已逾期的任务：

```bash
tick today
```

查看当前配置的 inbox 项目：

```bash
tick inbox
```

以 JSON 输出：

```bash
tick today --json
tick inbox --json
```

### Quick add

`tick quick add` 支持紧凑的快速录入语法：

- 普通文本会被当成任务标题
- `#ProjectName` 指定项目
- `!1`、`!3`、`!5` 指定优先级
- `^YYYY-MM-DD` 指定截止日期

示例：

```bash
tick quick add "Write spec #Work !5 ^2026-04-10"
```

```bash
tick quick add "Buy milk #Personal ^2026-04-18"
```

如果已经配置了 `task.default_project`，可以省略 `#ProjectName`：

```bash
tick config set task.default_project Work
tick quick add "Prepare launch notes !3 ^2026-04-22"
```

### 配置命令

查看完整本地配置：

```bash
tick config list
```

读取单个配置值：

```bash
tick config get service.region
```

设置默认输出格式：

```bash
tick config set output.default json
```

为 `quick add` 设置默认项目：

```bash
tick config set task.default_project Work
```

设置 `tick inbox` 使用的 inbox 项目 ID：

```bash
tick config set task.inbox_project_id YOUR_INBOX_PROJECT_ID
```

## 输出格式与优先级

支持的输出格式：

- `table`
- `json`

设置默认输出格式：

```bash
tick config set output.default table
```

优先级取值：

- `0` 表示无优先级
- `1` 表示低优先级
- `3` 表示中优先级
- `5` 表示高优先级

## 发布流程

本地校验 GoReleaser 配置：

```bash
make release-check
```

本地构建 snapshot 发布包：

```bash
make release
```

发布正式版本：

```bash
git tag v0.0.2
git push origin v0.0.2
```

推送 `v*` tag 后会触发 [release.yml](.github/workflows/release.yml)，自动运行测试、构建归档、生成 checksum，并上传到 GitHub Releases。
