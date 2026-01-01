# DeerFlow - 深度研究智能体

[ByteDance DeerFlow](https://github.com/bytedance/deer-flow) 深度研究智能体的 Go 实现，使用 [langgraphgo](https://github.com/smallnest/langgraphgo) 和 [langchaingo](https://github.com/tmc/langchaingo) 构建。

DeerFlow 是一个智能的多智能体研究系统，可以自主对任何主题进行深度研究，生成全面的报告，并可选择性地创建播客脚本以实现引人入胜的内容传递。

## 概述

DeerFlow 编排多个 AI 智能体执行结构化研究：

```
用户查询 → 规划器 → 研究员 → 报告者 → （可选）播客 → 最终输出
```

系统分解复杂的研究任务，系统地收集信息，并将发现综合成专业、格式良好的报告。

## 功能特性

### 🎯 多智能体架构
- **规划器智能体**：将查询分解为结构化的研究计划
- **研究员智能体**：使用 LLM 执行每个研究步骤
- **报告者智能体**：将发现综合成全面的 HTML 报告
- **播客智能体**：生成引人入胜的播客脚本（可选）

### 🌐 现代 Web 界面
- **实时进度**：使用服务器发送事件（SSE）的实时更新
- **深色主题 UI**：专业、护眼的界面
- **研究历史**：查看和重播过去的研究会话
- **结果缓存**：即时重播以前的查询

### 💻 双操作模式
- **Web 服务器**：交互式基于浏览器的界面
- **CLI 模式**：快速命令行执行

### 📊 丰富的输出格式
- **HTML 报告**：结构良好、带样式的研究报告
- **播客脚本**：用于音频制作的对话内容
- **持久存储**：自动保存研究结果

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                       DeerFlow                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────┐│
│  │  规划器  │───▶│  研究员  │───▶│  报告者  │───▶│ 播客 ││
│  │  智能体  │    │  智能体  │    │  智能体  │    │智能体││
│  └──────────┘    └──────────┘    └──────────┘    └──────┘│
│       │               │                │              │    │
│       ▼               ▼                ▼              ▼    │
│  从查询生成       使用 LLM        以 HTML 格式     生成   │
│  计划            执行步骤         创建报告         脚本   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 工作流程

1. **规划阶段**
   - 用户提交研究查询
   - 规划器智能体分析并创建逐步研究计划
   - 检测是否请求生成播客

2. **研究阶段**
   - 研究员智能体执行每个步骤
   - 使用 LLM 收集信息
   - 为每个研究步骤收集发现

3. **报告阶段**
   - 报告者智能体综合所有研究结果
   - 生成格式良好的 HTML 报告
   - 包含适当的结构和样式

4. **播客阶段**（可选）
   - 播客智能体创建对话脚本
   - 格式化内容以供音频传递
   - 保持参与度和流畅性

## 前置要求

- **Go**：版本 1.21 或更高
- **API 密钥**：OpenAI 兼容的 API（OpenAI、DeepSeek 等）
- **浏览器**：用于 UI 的现代网络浏览器（Chrome、Firefox、Safari、Edge）

## 安装

```bash
# 导航到 deerflow 目录
cd showcases/deerflow

# 设置环境变量
export OPENAI_API_KEY="your-api-key-here"

# 可选：如果使用 DeepSeek 或其他提供商
export OPENAI_API_BASE="https://api.deepseek.com/v1"

# 构建应用程序
go build -o deerflow .
```

## 使用方法

### Web 界面（推荐）

启动 Web 服务器：

```bash
./deerflow
```

然后打开浏览器并导航到：
```
http://localhost:8085
```

**Web 界面功能**：
- 在输入框中输入您的研究查询
- 观看实时进度更新
- 查看格式化的 HTML 报告
- 访问研究历史
- 即时重播以前的搜索

### 命令行界面

对于快速的一次性查询：

```bash
# 基本用法
./deerflow "您的研究问题"

# 示例查询
./deerflow "量子计算的最新进展是什么？"
./deerflow "解释 AI 对医疗保健的影响"
./deerflow "可再生能源的当前状态如何？"
```

### 示例查询

**技术研究**：
```bash
./deerflow "2024 年 AI 的突破性发展是什么？"
```

**科学研究**：
```bash
./deerflow "火星探索的最新发现是什么？"
```

**商业研究**：
```bash
./deerflow "电子商务的新兴趋势是什么？"
```

**生成播客**：
```bash
./deerflow "创建关于区块链技术的播客"
./deerflow "生成关于人工智能的播客脚本"
```

## 配置

### 环境变量

| 变量 | 描述 | 默认值 | 必需 |
|------|------|--------|------|
| `OPENAI_API_KEY` | OpenAI API 密钥 | 无 | ✅ 是 |
| `OPENAI_API_BASE` | API 基础 URL | OpenAI 默认值 | ❌ 否 |

### 服务器配置

Web 服务器默认在端口 **8085** 上运行。要更改此设置，请修改 `main.go`：

```go
server := &http.Server{
    Addr: ":8085",  // 在此处更改端口
    ReadHeaderTimeout: 3 * time.Second,
}
```

## 项目结构

```
deerflow/
├── main.go              # 入口点、HTTP 服务器、CLI 处理器
├── graph.go             # 图结构和状态定义
├── nodes.go             # 智能体实现（规划器、研究员、报告者、播客）
├── nginx.conf           # Nginx 配置（用于生产部署）
├── web/                 # 前端资源
│   ├── index.html       # 主 Web 界面
│   ├── styles.css       # UI 样式
│   └── script.js        # 客户端 JavaScript
├── data/                # 研究结果存储（自动创建）
│   └── [query]/         # 每个唯一查询一个文件夹
│       ├── metadata.json    # 查询元数据
│       ├── logs.json        # 研究过程日志
│       ├── report.html      # 生成的 HTML 报告
│       └── podcast.txt      # 播客脚本（如果生成）
└── README.md            # 本文档
```

## 工作原理

### 1. 规划器智能体

**输入**：用户查询

**处理**：
- 分析查询以了解研究范围
- 创建结构化的逐步研究计划
- 从关键词检测播客生成意图

**输出**：
```json
{
  "plan": ["步骤 1: ...", "步骤 2: ...", "步骤 3: ..."],
  "generate_podcast": true/false
}
```

**示例计划**：
对于查询"量子计算的最新进展是什么？"：
1. 搜索量子计算的最新研究进展
2. 调查主要的量子计算公司和项目
3. 分析量子计算的实际应用案例
4. 总结未来发展趋势和挑战

### 2. 研究员智能体

**输入**：研究计划

**处理**：
- 按顺序执行每个步骤
- 使用 LLM 收集详细信息
- 收集全面的发现

**输出**：每个步骤的研究结果数组

### 3. 报告者智能体

**输入**：所有研究结果

**处理**：
- 将发现综合成连贯的报告
- 以 HTML 格式化内容并具有适当的结构
- 添加样式以实现专业外观
- 可选择包含图像占位符

**输出**：完整的 HTML 报告

### 4. 播客智能体（可选）

**输入**：研究结果和最终报告

**处理**：
- 将技术内容转换为对话格式
- 创建引人入胜的对话或独白
- 保持信息准确性

**输出**：对话风格的播客脚本

## Web 界面功能

### 实时进度更新

Web 界面在研究期间提供实时更新：
- 初始规划阶段
- 每个研究步骤执行
- 报告生成
- 播客脚本创建

### 研究历史

- 自动保存所有研究会话
- 按时间戳浏览以前的查询
- 即时重播缓存的结果
- 对重复查询无冗余 API 调用

### 缓存系统

DeerFlow 智能地缓存研究结果：
- 每个唯一查询保存在 `data/[sanitized-query]/` 中
- 对相同查询的后续请求使用缓存的数据
- 通过模拟进度快速重播以获得更好的用户体验

## API 端点

### POST /api/run

执行研究查询。

**查询参数**：
- `query`（必需）：研究问题

**响应**：服务器发送事件流

**事件类型**：
- `update`：进度更新
- `log`：研究过程日志
- `result`：最终报告和播客脚本
- `error`：错误消息

**示例**：
```javascript
const eventSource = new EventSource('/api/run?query=Your+question');
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // 处理不同的事件类型
};
```

### GET /api/history

检索研究历史。

**响应**：
```json
[
  {
    "query": "研究问题",
    "timestamp": "2024-12-06T10:30:00Z",
    "dir_name": "Research_question"
  }
]
```

## 高级用法

### 自定义 LLM 模型

修改 `nodes.go` 以使用不同的模型：

```go
func getLLM() (llms.Model, error) {
    return openai.New(
        openai.WithModel("gpt-4"),  // 在此处更改模型
    )
}
```

### 扩展智能体

通过以下方式添加新的智能体节点：

1. 在 `nodes.go` 中定义节点函数：
```go
func MyCustomNode(ctx context.Context, state any) (any, error) {
    s := state.(*State)
    // 您的逻辑在此处
    return s, nil
}
```

2. 在 `graph.go` 中注册节点：
```go
workflow.AddNode("custom", "Custom node description", MyCustomNode)
workflow.AddEdge("previous_node", "custom")
```

### 生产部署

对于生产环境，使用包含的 nginx.conf：

```bash
# 复制 nginx 配置
sudo cp nginx.conf /etc/nginx/sites-available/deerflow
sudo ln -s /etc/nginx/sites-available/deerflow /etc/nginx/sites-enabled/

# 启动 DeerFlow
./deerflow &

# 重启 nginx
sudo systemctl restart nginx
```

## 故障排除

### API 密钥未设置

```
Please set OPENAI_API_KEY environment variable
```

**解决方案**：
```bash
export OPENAI_API_KEY="sk-..."
```

### 连接被拒绝

如果 Web 界面无法加载：
- 检查端口 8085 是否可用
- 验证应用程序正在运行
- 检查防火墙设置

### 报告为空或不完整

如果报告不充分：
- 验证 API 密钥有效且有额度
- 如果使用非 OpenAI 提供商，请检查 API 基础 URL
- 尝试使用更具体的查询
- 检查网络连接

### JSON 解析错误

系统包括后备解析：
- 如果 LLM 返回格式错误的 JSON，它使用简单的文本解析
- 检查日志以了解解析问题
- 考虑使用更强大的模型（GPT-4 vs GPT-3.5）

## 性能考虑

### 响应时间

- **规划**：2-5 秒
- **研究**：5-15 秒（取决于计划步骤）
- **报告**：5-10 秒
- **播客**：5-10 秒（如果启用）
- **总计**：通常 15-40 秒

### 成本优化

- 对研究步骤使用更便宜的模型（gpt-3.5-turbo）
- 对最终报告使用高级模型（gpt-4）
- 缓存结果以避免重复的 API 调用
- 对于更简单的查询，限制研究计划步骤

### 缓存优势

- 对重复查询**零成本**
- **即时结果**（每个日志重播 200 毫秒）
- 对相同问题**一致的输出**
- 为用户**节省带宽**

## 未来增强

计划功能：
- [ ] 真实的网络搜索集成（Tavily、Google、Bing）
- [ ] 多语言支持
- [ ] PDF 导出
- [ ] 从播客脚本生成音频
- [ ] 协作研究会话
- [ ] 自定义报告模板
- [ ] 图像搜索和包含
- [ ] 带链接的来源引用
- [ ] 导出为各种格式（Markdown、Word 等）

## 与 ByteDance DeerFlow 的比较

| 功能 | ByteDance DeerFlow (Python) | 本实现 (Go) |
|------|----------------------------|-------------|
| 多智能体架构 | ✅ | ✅ |
| 研究规划 | ✅ | ✅ |
| 网络搜索 | ✅ | ⚠️ 基于 LLM（计划中） |
| 报告生成 | ✅ | ✅ |
| Web 界面 | ✅ | ✅ |
| CLI 支持 | ✅ | ✅ |
| 播客生成 | ❌ | ✅ |
| 结果缓存 | ⚠️ | ✅ |
| SSE 实时更新 | ⚠️ | ✅ |
| 历史浏览 | ⚠️ | ✅ |
| 语言 | Python | Go |

## 许可证

MIT License - 与父项目 langgraphgo 相同

## 参考资料

- [ByteDance DeerFlow](https://github.com/bytedance/deer-flow) - 原始 Python 实现
- [LangGraph Go](https://github.com/smallnest/langgraphgo) - 基于图的智能体框架
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM 集成库

## 贡献

欢迎贡献！改进领域：
- 真实的网络搜索集成
- 增强的 UI/UX
- 额外的导出格式
- 性能优化
- 测试覆盖率
- 文档改进

## 支持

对于问题和疑问：
- 查看故障排除部分
- 查看本 README 中的示例
- 在 langgraphgo GitHub 仓库上开启 issue

---

**构建工具**：
- [langgraphgo](https://github.com/smallnest/langgraphgo) - 基于图的智能体编排
- [langchaingo](https://github.com/tmc/langchaingo) - LLM 集成
- 服务器发送事件用于实时更新
- 嵌入式 Go Web 服务器以实现简单性
