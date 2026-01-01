# BettaFish（Go 实现）

这是 [BettaFish](https://github.com/666ghj/BettaFish) 项目的**完整复刻** Go 版本，使用 [langgraphgo](https://github.com/smallnest/langgraphgo) 和 [langchaingo](https://github.com/tmc/langchaingo) 构建。

本项目实现了深度舆情分析的完整多智能体架构。

## 功能特性

- **QueryEngine（查询引擎）**：
  - 生成结构化的研究计划（大纲）
  - 使用 Tavily API 执行深度网络搜索
  - 实现**反思循环**以迭代优化搜索结果和摘要
  - 使用专门的提示词进行搜索、摘要和反思

- **MediaEngine（媒体引擎）**：
  - 使用 Tavily 的图像搜索功能搜索相关图片

- **InsightEngine（洞察引擎）**：
  - （模拟）挖掘内部数据以获取洞察

- **ForumEngine（论坛引擎）**：
  - 促进由 LLM 驱动的"NewsAgent"、"MediaAgent"和"Moderator"之间的讨论，以综合发现

- **ReportEngine（报告引擎）**：
  - 将所有发现编译成综合 Markdown 报告

## 前置要求

您需要以下 API 密钥：
- `OPENAI_API_KEY`：用于 LLM 推理（推荐 GPT-4o，或任何 OpenAI 兼容的 API）
- `TAVILY_API_KEY`：用于网络搜索和图像搜索

**可选**：对于使用替代 LLM 提供商（例如 DeepSeek、Azure OpenAI 或任何 OpenAI 兼容 API）：
- `OPENAI_API_BASE`：设置为您的自定义 API 端点。这允许您使用任何 OpenAI 兼容的服务。例如：
  - DeepSeek：`https://api.deepseek.com/v1`
  - Azure OpenAI：`https://YOUR_RESOURCE_NAME.openai.azure.com/openai/deployments/YOUR_DEPLOYMENT_NAME`
  - 本地模型（Ollama、vLLM 等）：`http://localhost:11434/v1`
- `OPENAI_MODEL`：如果需要，覆盖默认模型名称。这在以下情况下特别有用：
  - 在不同的 OpenAI 模型之间切换（例如 `gpt-4o`、`gpt-4o-mini`、`gpt-4-turbo`）
  - 使用具有特定模型名称的替代提供商（例如 `deepseek-chat`、`claude-3-haiku` 等）

## 使用方法

### 基本用法（OpenAI）

```bash
export OPENAI_API_KEY="sk-..."
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "您的研究主题"
```

### 使用替代提供商（例如 DeepSeek）

```bash
export OPENAI_API_KEY="your-deepseek-api-key"
export OPENAI_API_BASE="https://api.deepseek.com/v1"
export OPENAI_MODEL="deepseek-chat"  # 指定模型名称
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "您的研究主题"
```

### 使用不同的 OpenAI 模型

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_MODEL="gpt-4o-mini"  # 或 gpt-4o、gpt-4-turbo 等
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "您的研究主题"
```

### 使用本地模型（Ollama 示例）

```bash
export OPENAI_API_KEY="ollama"  # 对于本地模型可以是任何值
export OPENAI_API_BASE="http://localhost:11434/v1"
export OPENAI_MODEL="llama3.1"  # 指定模型名称
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "您的研究主题"
```

## 架构

BettaFish 采用多引擎协作架构：

```
用户查询
    ↓
QueryEngine（查询引擎）
    ├── 生成研究大纲
    ├── 执行深度搜索（Tavily）
    └── 反思循环优化
    ↓
MediaEngine（媒体引擎）
    └── 搜索相关图片
    ↓
InsightEngine（洞察引擎）
    └── 挖掘内部数据
    ↓
ForumEngine（论坛引擎）
    ├── NewsAgent（新闻智能体）
    ├── MediaAgent（媒体智能体）
    └── Moderator（主持人）
    ↓
ReportEngine（报告引擎）
    └── 生成综合报告
    ↓
最终 Markdown 报告
```

## 工作流程

### 1. QueryEngine - 深度搜索与反思

**核心功能**：
- **大纲生成**：创建结构化研究计划
- **深度搜索**：使用 Tavily API 进行网络搜索
- **反思循环**：迭代优化搜索结果和摘要
  - 搜索 → 摘要 → 反思 → 改进 → 重复

**反思循环机制**：
```
初始搜索
    ↓
生成摘要
    ↓
反思评估 ← ─ ─ ─ ┐
    ↓             │
质量判断         │
    ├─ 满意 → 完成
    └─ 不满意 ───┘（重新搜索/优化）
```

### 2. MediaEngine - 图像搜索

- 使用 Tavily 的图像搜索 API
- 查找与主题相关的视觉内容
- 增强报告的视觉吸引力

### 3. InsightEngine - 数据洞察

- 模拟内部数据挖掘
- 提供额外的背景信息
- 补充网络搜索结果

### 4. ForumEngine - 多智能体讨论

**参与者**：
- **NewsAgent**：关注新闻和事实
- **MediaAgent**：关注媒体和图像
- **Moderator**：主持讨论，综合观点

**讨论流程**：
1. Moderator 提出讨论议题
2. NewsAgent 提供新闻视角
3. MediaAgent 提供媒体视角
4. Moderator 总结关键见解
5. 达成共识

### 5. ReportEngine - 报告生成

- 整合所有引擎的发现
- 生成结构化 Markdown 报告
- 包含摘要、详细分析、图片和结论

## 项目结构

```
BettaFish/
├── main.go                    # 主入口点
├── schema/                    # 数据结构定义
│   └── state.go              # 全局状态定义
├── query_engine/             # 查询引擎
│   ├── query_engine.go       # 搜索与反思逻辑
│   └── prompts.go            # 提示词模板
├── media_engine/             # 媒体引擎
│   └── media_engine.go       # 图像搜索
├── insight_engine/           # 洞察引擎
│   └── insight_engine.go     # 数据挖掘
├── forum_engine/             # 论坛引擎
│   └── forum_engine.go       # 多智能体讨论
├── report_engine/            # 报告引擎
│   └── report_engine.go      # 报告生成
├── sentiment_model/          # 情感分析
│   └── sentiment_model.go    # 情感分析逻辑
├── mind_spider/              # 网络爬虫
│   └── mind_spider.go        # 网页抓取
├── DIFF.md                   # 与原项目的差异
└── README.md                 # 本文档
```

## 核心特性

### 反思循环（Reflection Loop）

BettaFish 的关键创新是**反思循环**机制：

1. **搜索阶段**：使用 Tavily API 获取初始结果
2. **摘要阶段**：LLM 生成搜索结果摘要
3. **反思阶段**：LLM 评估摘要质量
4. **判断阶段**：
   - 如果满意 → 继续下一步
   - 如果不满意 → 返回搜索阶段，使用改进的查询

这确保了高质量的研究结果。

### 多智能体协作

**ForumEngine** 模拟真实的讨论环境：

- **NewsAgent**：提供基于事实的新闻视角
- **MediaAgent**：提供视觉和媒体视角
- **Moderator**：引导讨论，确保全面性

通过多个视角的协作，生成更全面、更平衡的分析。

### 综合报告

最终报告包含：

- **执行摘要**：关键发现概述
- **详细分析**：按主题组织的深入分析
- **视觉内容**：相关图片和图表
- **讨论摘要**：论坛讨论的关键点
- **结论**：综合见解和建议

## 示例用法

### 技术趋势分析

```bash
go run showcases/BettaFish/main.go "2024年人工智能的最新发展趋势"
```

### 舆情分析

```bash
go run showcases/BettaFish/main.go "电动汽车市场的公众舆情分析"
```

### 事件研究

```bash
go run showcases/BettaFish/main.go "量子计算突破对科技行业的影响"
```

### 产品调研

```bash
go run showcases/BettaFish/main.go "消费者对新能源产品的态度和看法"
```

## 配置选项

### LLM 模型选择

通过环境变量配置模型：

```bash
# GPT-4o（推荐）
export OPENAI_MODEL="gpt-4o"

# GPT-4o-mini（更快、更便宜）
export OPENAI_MODEL="gpt-4o-mini"

# GPT-4 Turbo
export OPENAI_MODEL="gpt-4-turbo"

# DeepSeek
export OPENAI_MODEL="deepseek-chat"
```

### API 端点配置

```bash
# OpenAI（默认）
# 无需设置 OPENAI_API_BASE

# DeepSeek
export OPENAI_API_BASE="https://api.deepseek.com/v1"

# Azure OpenAI
export OPENAI_API_BASE="https://YOUR_RESOURCE.openai.azure.com/openai/deployments/YOUR_DEPLOYMENT"

# 本地 Ollama
export OPENAI_API_BASE="http://localhost:11434/v1"
```

## 与原项目的对比

| 功能 | 原 BettaFish (Python) | 本实现 (Go) |
|------|---------------------|------------|
| QueryEngine | ✅ | ✅ |
| 反思循环 | ✅ | ✅ |
| MediaEngine | ✅ | ✅ |
| InsightEngine | ✅ | ✅ |
| ForumEngine | ✅ | ✅ |
| ReportEngine | ✅ | ✅ |
| Tavily 集成 | ✅ | ✅ |
| 多智能体讨论 | ✅ | ✅ |
| 图像搜索 | ✅ | ✅ |
| 语言 | Python | Go |
| 框架 | LangGraph | LangGraphGo |
| 性能 | - | 更快 |
| 部署 | - | 单二进制文件 |

## 性能考虑

### 执行时间

典型的研究会话：

- **QueryEngine**（反思循环）：30-60 秒
- **MediaEngine**：5-10 秒
- **InsightEngine**：5-10 秒
- **ForumEngine**（多轮讨论）：20-40 秒
- **ReportEngine**：10-15 秒
- **总计**：约 70-135 秒（1-2 分钟）

### API 成本估算

每次完整分析的 API 调用：

- **搜索请求**：5-10 次（Tavily）
- **LLM 调用**：15-30 次（取决于反思循环和讨论轮次）
- **图像搜索**：1-3 次

**成本优化建议**：
- 使用 `gpt-4o-mini` 替代 `gpt-4o`（成本降低约 80%）
- 使用 DeepSeek 或本地模型进一步降低成本
- 限制反思循环的迭代次数
- 减少论坛讨论的轮次

## 故障排除

### API 密钥错误

```
Error: OPENAI_API_KEY not set
```

**解决方案**：
```bash
export OPENAI_API_KEY="your-api-key"
export TAVILY_API_KEY="your-tavily-key"
```

### 连接超时

如果遇到超时错误：
- 检查网络连接
- 验证 API 端点配置正确
- 增加超时时间限制
- 检查 API 服务状态

### 搜索结果质量差

如果搜索结果不理想：
- 使用更具体的查询
- 增加反思循环的迭代次数
- 使用更强大的模型（GPT-4o 而非 GPT-4o-mini）
- 检查 Tavily API 配额

### JSON 解析错误

系统包含容错机制：
- 自动重试解析
- 降级到文本模式
- 记录错误以供调试

## 高级用法

### 自定义反思循环

修改 `query_engine/query_engine.go` 中的反思逻辑：

```go
// 调整满意度阈值
const satisfactionThreshold = 0.8

// 调整最大迭代次数
const maxReflectionIterations = 3
```

### 自定义论坛智能体

在 `forum_engine/forum_engine.go` 中添加新的智能体：

```go
type CustomAgent struct {
    Name string
    Role string
}

func (a *CustomAgent) Respond(context string) (string, error) {
    // 自定义逻辑
}
```

### 自定义报告格式

修改 `report_engine/report_engine.go` 中的报告模板：

```go
const reportTemplate = `
# 自定义报告标题

## 您的自定义部分
...
`
```

## 最佳实践

### 1. 编写清晰的查询

✅ **好的查询**：
- "分析2024年电动汽车市场的消费者情绪和主要趋势"
- "评估人工智能在医疗保健领域的应用及公众反应"
- "研究可再生能源政策的社会舆论和影响"

❌ **模糊的查询**：
- "AI 是什么"
- "告诉我关于汽车的事"

### 2. 选择合适的模型

- **深度分析**：使用 GPT-4o
- **快速研究**：使用 GPT-4o-mini
- **成本优先**：使用 DeepSeek 或本地模型

### 3. 监控 API 使用

- 启用详细日志记录
- 跟踪 API 调用次数
- 设置成本预警
- 使用缓存减少重复调用

### 4. 优化报告质量

- 提供具体的研究主题
- 使用高质量的 LLM 模型
- 允许足够的反思迭代
- 审查和验证生成的内容

## 开发路线图

计划中的功能：

- [ ] 支持更多搜索引擎（Google、Bing）
- [ ] 实时数据流分析
- [ ] 情感分析可视化
- [ ] 多语言支持
- [ ] 自定义报告模板
- [ ] PDF 导出
- [ ] Web 界面
- [ ] API 服务器模式
- [ ] 批量分析
- [ ] 结果缓存和重用

## 许可证

MIT License - 与父项目 langgraphgo 相同

## 参考资料

- [原始 BettaFish 项目](https://github.com/666ghj/BettaFish) - Python 实现
- [LangGraphGo](https://github.com/smallnest/langgraphgo) - 基于图的智能体框架
- [LangChainGo](https://github.com/tmc/langchaingo) - LLM 集成库
- [Tavily API](https://www.tavily.com/) - 搜索 API

## 贡献

欢迎贡献！改进领域：

- 增强搜索算法
- 改进反思循环逻辑
- 添加更多智能体类型
- 优化性能
- 增加测试覆盖率
- 改进文档

## 支持

如有问题和疑问：
- 查看故障排除部分
- 查看本 README 中的示例
- 在 langgraphgo GitHub 仓库中提交 issue

---

**构建工具**：
- [langgraphgo](https://github.com/smallnest/langgraphgo) - 基于图的智能体编排
- [langchaingo](https://github.com/tmc/langchaingo) - LLM 集成
- [Tavily](https://www.tavily.com/) - 搜索和图像 API
- Go 1.21+ - 高性能编程语言
