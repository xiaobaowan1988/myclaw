# 从零开始，用 Go 构建一个 AI 编程助手 —— OpenClaw 实战解析

> 1911 行代码，5 个阶段，19 个测试用例。这篇文章完整记录我如何从一个空目录出发，一步步构建出一个可以自主阅读、编辑、搜索代码的 AI Agent。

---

## 一、为什么要做这个项目？

市面上有 Claude Code、Cursor、GitHub Copilot 这些 AI 编程工具，但它们要么闭源，要么价格不菲。我想回答一个问题：

**一个 AI 编程助手的核心到底是什么？能不能用最少的代码把它造出来？**

于是有了 OpenClaw —— 一个用 Go 语言 + Google Gemini Flash 模型构建的开源 AI 编程 Agent。整个项目不到 2000 行代码，却实现了完整的 Agent 循环：思考 → 调用工具 → 观察结果 → 继续思考。

---

## 二、整体架构：一张图看懂

```
用户输入
   │
   ▼
┌──────────┐     ┌─────────────┐     ┌──────────────┐
│  CLI层   │────▶│  Agent 循环  │────▶│  Gemini API  │
│ (Cobra)  │     │ think→act→  │     │  (流式输出)   │
└──────────┘     │ observe     │     └──────────────┘
                 └──────┬──────┘
                        │
                 ┌──────▼──────┐
                 │  7 个工具    │
                 │ read/write/ │
                 │ edit/search │
                 │ glob/ls/cmd │
                 └─────────────┘
```

项目目录结构：

```
openclaw/
├── main.go                    # 入口，15 行
├── cmd/root.go                # CLI 命令定义
├── config/config.go           # 配置加载
└── internal/
    ├── agent/agent.go         # 核心：Agent 循环
    ├── llm/gemini.go          # Gemini API 客户端
    ├── tools/                 # 7 个工具实现
    ├── context/manager.go     # 上下文管理
    └── ui/                    # 终端渲染
```

---

## 三、Phase 1：地基 —— 项目骨架与 Gemini 客户端

### 3.1 配置系统

第一步不是写 AI 逻辑，而是解决一个朴素的问题：API Key 从哪来？

```go
// config/config.go
type Config struct {
    APIKey      string  `yaml:"api_key"`
    Model       string  `yaml:"model"`
    MaxTokens   int32   `yaml:"max_tokens"`
    Temperature float32 `yaml:"temperature"`
}
```

我设计了两层配置覆盖机制：
1. **YAML 文件**（`~/.openclaw.yaml`）—— 持久化配置
2. **环境变量**（`GEMINI_API_KEY`）—— 优先级更高，适合 CI/CD

**为什么这样设计？** 因为 API Key 属于敏感信息，不应该硬编码在代码里。环境变量是 12-Factor App 的标准做法，而 YAML 文件给开发者一个舒适的本地配置方式。

### 3.2 Gemini 客户端：流式输出 + 函数调用

这是整个项目最关键的一层抽象。Gemini API 的两大核心能力：

1. **流式输出（Streaming）**—— 让用户看到 AI "正在思考"
2. **函数调用（Function Calling）**—— 让 AI 能操作外部工具

```go
// internal/llm/gemini.go
func (c *Client) GenerateStream(
    ctx context.Context,
    systemPrompt string,
    messages []Message,
    tools []ToolDef,
    onText func(chunk string),  // 每收到一小段文字就回调
) ([]ToolCall, error) {
```

这个函数的签名揭示了核心设计思想：

- `onText` 回调实现流式输出，文字一个 chunk 一个 chunk 地吐出来
- 返回值 `[]ToolCall` 代表模型想调用的工具
- 如果返回空切片，说明模型已经完成思考，可以直接把文字展示给用户

**消息格式转换** 是对接 API 时最繁琐的部分。我定义了内部消息类型，然后写了转换函数：

```go
type Message struct {
    Role       string      // "user", "model", "tool"
    Text       string
    ToolCalls  []ToolCall  // 模型发起的工具调用
    ToolResult *ToolResult // 工具执行结果
}
```

三种角色，三种不同的转换逻辑。其中 `tool` 角色的消息需要包装成 `FunctionResponse`，这是 Gemini API 的特殊要求。

---

## 四、Phase 2：工具系统 —— 给 AI 装上手脚

AI 模型再聪明，不能读文件、不能执行命令，就只是一个聊天机器人。工具系统把 AI 变成 Agent。

### 4.1 注册表模式

```go
// internal/tools/registry.go
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]llm.ParamDef
    Execute     ExecuteFunc  // func(args map[string]any) (string, error)
}

type Registry struct {
    tools map[string]*Tool
}
```

为什么用注册表（Registry）模式？

- **解耦**：每个工具独立一个文件，互不干扰
- **可扩展**：加新工具只需要 `r.Register(NewTool())`
- **统一接口**：Agent 不需要知道每个工具的细节，只需要 `registry.Execute(name, args)`

### 4.2 七个核心工具

| 工具 | 用途 | 文件 |
|------|------|------|
| `read_file` | 读取文件内容（带行号） | readfile.go |
| `write_file` | 创建/覆盖文件 | writefile.go |
| `edit_file` | 精确替换文件中的字符串 | editfile.go |
| `search` | 正则搜索文件内容 | search.go |
| `glob` | 按模式查找文件 | glob.go |
| `list_dir` | 列出目录内容 | listdir.go |
| `run_cmd` | 执行 Shell 命令 | runcmd.go |

每个工具都遵循相同的模式，举 `edit_file` 为例：

```go
func EditFileTool() *Tool {
    return &Tool{
        Name:        "edit_file",
        Description: "Edit a file by replacing an exact string match...",
        Parameters: map[string]llm.ParamDef{
            "path":       {Type: "string", Required: true},
            "old_string": {Type: "string", Required: true},
            "new_string": {Type: "string", Required: true},
        },
        Execute: func(args map[string]any) (string, error) {
            // 1. 读取文件
            // 2. 检查 old_string 是否存在且唯一
            // 3. 替换并写回
        },
    }
}
```

**关键设计决策**：`edit_file` 要求 `old_string` 必须在文件中恰好出现一次。如果出现 0 次或多次，直接报错。这避免了 AI 误改不该改的地方。

---

## 五、Phase 3：Agent 循环 —— 项目的灵魂

这是最核心的 50 行代码，也是整个 AI Agent 的灵魂：

```go
func (a *Agent) Run(ctx context.Context, userMessage string) (string, error) {
    a.history = append(a.history, llm.Message{Role: "user", Text: userMessage})
    a.history = a.ctxMgr.TrimHistory(a.history)

    for i := 0; i < maxToolLoops; i++ {  // 最多 25 轮
        // 1. 调用 Gemini，流式获取响应
        toolCalls, err := a.client.GenerateStream(...)

        // 2. 没有工具调用 → AI 已经完成思考，返回文字
        if len(toolCalls) == 0 {
            return text, nil
        }

        // 3. 有工具调用 → 执行每个工具
        for _, tc := range toolCalls {
            output, err := a.registry.Execute(tc.Name, tc.Args)
            // 4. 把结果追加到对话历史
            a.history = append(a.history, toolResultMessage)
        }
        // 5. 带着工具结果，回到步骤 1
    }
}
```

**这就是 AI Agent 的本质**：一个循环。

```
用户消息 → LLM 思考 → 要调工具吗？
                         │
                    ┌────┤
                    │    │
                   是    否 → 返回文字给用户
                    │
                    ▼
              执行工具，获得结果
                    │
                    ▼
              把结果追加到历史
                    │
                    ▼
              回到 "LLM 思考"
```

**25 轮上限** 是一个安全阀。如果 AI 陷入了无限循环（比如反复读取同一个文件），这个限制会强制中断。

---

## 六、Phase 4：CLI 与终端 UI

### 6.1 两种运行模式

```bash
# 交互模式：像聊天一样
./openclaw

# 单次模式：执行一个任务就退出
./openclaw "解释 main.go 的作用"
```

用 Cobra 框架实现，核心就几行：

```go
var rootCmd = &cobra.Command{
    Use:   "openclaw [message]",
    RunE:  runAgent,
}
```

如果 `args` 非空，走单次模式；否则进入交互式 REPL 循环。

### 6.2 工具调用可视化

当 AI 调用工具时，用户会看到黄色的指示：

```
❯ 帮我看看 main.go 的内容

  [read_file: main.go]

这个文件是项目的入口点，它做了以下几件事...
```

这个小细节让用户知道 AI 在做什么，建立信任感。

### 6.3 Spinner 动画

```go
// 用 Unicode Braille 字符做旋转动画
frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
```

用 goroutine + channel 实现的非阻塞动画，`Start()` 启动，`Stop()` 清除。

---

## 七、Phase 5：安全与上下文管理

### 7.1 危险命令拦截

AI 可以执行 Shell 命令，这很强大但也很危险。我加了一道安全网：

```go
var destructivePatterns = []string{
    "rm -rf /",
    "rm -rf ~",
    "mkfs.",
    "dd if=",
    ":(){:|:&};:",        // Fork 炸弹
    "chmod -R 777 /",
    "git push --force origin main",
    "git reset --hard",
    "> /dev/sda",
}
```

所有命令执行前都要过一遍黑名单。匹配到就直接拒绝，不给 AI 任何执行机会。

### 7.2 上下文窗口管理

对话越长，发送给 API 的 token 越多，成本越高，速度越慢。上下文管理器用一个简单但有效的策略：

```go
func (m *Manager) TrimHistory(messages []llm.Message) []llm.Message {
    if len(messages) <= m.maxMessages {
        return messages
    }
    // 保留第一条消息（用户的初始意图）
    // + 最近的 N-1 条消息
    result = append(result, messages[0])
    result = append(result, messages[len(messages)-keep:]...)
    return result
}
```

**为什么保留第一条消息？** 因为它通常包含用户的核心意图。如果丢掉了，AI 会"忘记"自己在做什么。

### 7.3 仓库感知

Agent 启动时会自动检测 Git 仓库，把目录结构注入系统提示词：

```
## Repository
Git repository: /home/user/myproject
Top-level structure:
  cmd/
  internal/
  go.mod
  main.go
```

这样 AI 在回答第一个问题时就已经了解项目的全貌。

---

## 八、测试策略

整个项目有 19 个测试用例，覆盖每一层：

```
=== RUN   TestConvertTools          ✅ LLM 工具格式转换
=== RUN   TestConvertMessages       ✅ LLM 消息格式转换
=== RUN   TestRegistry              ✅ 7 个工具都注册成功
=== RUN   TestReadFileTool          ✅ 读文件带行号
=== RUN   TestWriteFileTool         ✅ 写文件自动建目录
=== RUN   TestEditFileTool          ✅ 精确替换 + 错误处理
=== RUN   TestSearchTool            ✅ 正则搜索 + glob 过滤
=== RUN   TestGlobTool              ✅ 文件查找
=== RUN   TestListDirTool           ✅ 目录列表
=== RUN   TestRunCmdTool            ✅ 命令执行
=== RUN   TestRunCmdSafety          ✅ 危险命令拦截
=== RUN   TestBuildSystemPrompt     ✅ 系统提示词生成
=== RUN   TestAgentNew              ✅ Agent 初始化
=== RUN   TestAgentReset            ✅ 对话重置
=== RUN   TestTrimHistory           ✅ 上下文裁剪
=== RUN   TestRepoRoot              ✅ Git 仓库检测
=== RUN   TestRepoSummary           ✅ 仓库摘要
=== RUN   TestPrinterColor          ✅ 颜色输出
=== RUN   TestSpinnerStartStop      ✅ 动画启停
```

**测试哲学**：不依赖外部 API。所有测试都是纯本地的，用临时文件和目录验证工具行为。Agent 的集成测试验证结构初始化和历史管理，不实际调用 Gemini。

---

## 九、技术选型背后的思考

| 选择 | 替代方案 | 选择理由 |
|------|---------|---------|
| Go | Python, Rust | 编译为单一二进制，部署简单；并发天生强 |
| Gemini Flash | GPT-4, Claude | 速度快、便宜、Function Calling 好用 |
| Cobra | flag 标准库 | 子命令扩展性好，自带 help 生成 |
| YAML 配置 | TOML, JSON | 人类可读性最好，Go 社区广泛使用 |
| 注册表模式 | Switch-case | 可扩展，工具之间完全解耦 |

---

## 十、项目数据

- **总代码量**：1,911 行（含测试）
- **生产代码**：约 1,400 行
- **测试代码**：约 500 行
- **依赖**：3 个直接依赖（Cobra、genai、yaml）
- **测试用例**：19 个，全部通过
- **工具数量**：7 个

---

## 十一、如何运行

```bash
# 克隆项目
git clone https://github.com/xiaobaowan1988/myclaw.git
cd myclaw

# 设置 API Key
export GEMINI_API_KEY="your-key-here"

# 编译
go build -o openclaw .

# 开始使用
./openclaw
```

---

## 十二、总结

回到最初的问题：**AI 编程助手的核心是什么？**

答案出奇地简单 —— 就是一个循环：

1. 把用户的话和工具描述发给 LLM
2. LLM 说想调哪个工具
3. 执行工具，把结果塞回对话
4. 回到第 1 步

剩下的一切 —— 流式输出、彩色终端、安全防护、上下文管理 —— 都是围绕这个循环的锦上添花。

**2000 行代码就能写出一个 AI Agent。** 这不是在炫技，而是在说明：AI 应用的门槛正在快速降低。真正的挑战不是代码量，而是对工具设计、安全边界和用户体验的思考。

项目是开源的，欢迎 star 和 PR。下一篇文章我们聊聊如何给它加上更强的能力。

---

*作者注：本项目使用 Claude Code 辅助开发，全程代码已推送到 GitHub。*
