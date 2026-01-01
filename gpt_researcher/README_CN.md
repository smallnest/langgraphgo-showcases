# GPT Researcher - Go 实现

使用 langgraphgo 框架和 langchaingo 库实现的 [gpt-researcher](https://github.com/assafelovic/gpt-researcher) Go 版本。这是一个自主研究智能体，旨在对任何给定主题进行全面研究，并生成带引用的详细事实报告。

## 概述

GPT Researcher 是一个多智能体系统，通过以下方式自动化研究过程：
1. **规划**：从查询生成聚焦的研究问题
2. **执行**：从多个网络来源收集信息
3. **发布**：将发现综合成全面的研究报告

该系统生成详细报告（2000+ 字），汇总来自 20+ 个来源的信息，完整包含引用和参考文献。

## 架构

系统由三个主要智能体组成流水线工作：

```
┌─────────────────────────────────────────────────────────────┐
│                    GPT Researcher                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────┐ │
│  │   规划器     │─────▶│   执行器     │─────▶│  发布器  │ │
│  │   智能体     │      │   智能体     │      │  智能体  │ │
│  └──────────────┘      └──────────────┘      └──────────┘ │
│        │                      │                     │      │
│        ▼                      ▼                     ▼      │
│  生成研究问题            从网络收集              生成报告  │
│                         和总结信息                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 1. 规划器智能体

**职责**：生成全面的研究问题

- 分析研究查询
- 创建 5-10 个聚焦的研究问题
- 确保问题涵盖多个视角
- 问题共同形成客观理解

**示例问题**（针对"2024年大型语言模型的最新进展是什么？"）：
1. 2024年发布的LLM有哪些主要架构创新？
2. 最近的LLM模型的推理能力如何演变？
3. LLM的效率和成本方面取得了哪些改进？
4. LLM的最新应用和用例是什么？
5. 引入了哪些伦理和安全进展？

### 2. 执行器智能体

**职责**：研究每个问题并收集信息

对于每个研究问题：
- 使用 Tavily API 执行网络搜索
- 检索最相关的来源（每个问题最多20个）
- 从网页抓取和提取内容
- 使用 LLM 总结每个来源
- 跟踪引用和相关性分数

**使用的工具**：
- `tavily_search`：带相关性排名的网络搜索
- `web_scraper`：从URL提取内容
- `summarizer`：基于LLM的总结

### 3. 发布器智能体

**职责**：将发现综合成最终报告

- 汇总所有摘要和发现
- 按主题组织信息
- 生成 2000+ 字的全面报告
- 包含引用和参考文献
- 提供客观分析和见解

## 功能特性

✅ **多智能体研究流水线**：规划器 → 执行器 → 发布器
✅ **全面报告**：2000+ 字的详细分析
✅ **来源汇总**：来自 20+ 个可信来源的信息
✅ **自动引用**：整个报告中的编号参考
✅ **网络搜索集成**：Tavily API 提供相关结果
✅ **灵活配置**：自定义模型、参数、输出
✅ **进度跟踪**：详细模式显示研究进度
✅ **多种报告类型**：研究、大纲或资源报告

## 要求

- Go 1.21 或更高版本
- OpenAI API 密钥
- Tavily API 密钥（用于网络搜索）

## 安装

```bash
# 导航到展示目录
cd showcases/gpt_researcher

# 设置环境变量
export OPENAI_API_KEY="your-openai-api-key"
export TAVILY_API_KEY="your-tavily-api-key"

# 安装依赖
go mod download
```

## 使用方法

### 基本用法

```bash
# 使用默认查询运行
go run *.go

# 使用自定义查询运行
go run *.go "您的研究问题"
```

### 示例查询

```bash
# 技术研究
go run *.go "量子计算的最新进展是什么？"

# 市场研究
go run *.go "电动汽车市场的当前状态如何？"

# 学术研究
go run *.go "CRISPR基因编辑的最新突破是什么？"

# 商业研究
go run *.go "可持续能源的新兴趋势是什么？"
```

### 配置

使用环境变量自定义行为：

```bash
# 模型配置
export GPT_MODEL="gpt-4"                    # 研究主模型
export REPORT_MODEL="gpt-4"                 # 最终报告模型
export SUMMARY_MODEL="gpt-3.5-turbo"        # 总结模型

# 搜索参数
export MAX_SEARCH_RESULTS="10"              # 每次搜索的结果数
export MAX_SOURCES_TO_USE="20"              # 使用的总来源数
export MAX_QUESTIONS="5"                    # 生成的研究问题数

# 报告配置
export REPORT_TYPE="research_report"        # 报告类型
export REPORT_FORMAT="markdown"             # 输出格式
export OUTPUT_DIR="./output"                # 保存位置
export SAVE_INTERMEDIATE="true"             # 保存报告到文件

# 详细程度
export VERBOSE="true"                       # 显示进度
```

## 工作原理

### 研究工作流

1. **初始化**：用户提供研究查询
2. **规划**：规划器智能体生成 5 个研究问题
3. **执行**：对于每个问题：
   - 使用 Tavily 搜索网络
   - 检索前 10-20 个结果
   - 抓取和总结每个来源
   - 跟踪引用
4. **发布**：发布器智能体：
   - 按问题分组发现
   - 综合全面报告
   - 添加引用和参考文献
5. **完成**：返回最终报告

### 示例输出

```
================================================================================
GPT RESEARCHER
================================================================================

📋 研究查询：2024年大型语言模型的最新进展是什么？

🎯 [规划器智能体] 正在生成研究问题...
✅ [规划器智能体] 生成了 5 个研究问题：
   1. 2024年发布的LLM有哪些主要架构创新？
   2. 最近的LLM模型的推理能力如何演变？
   3. LLM的效率和成本方面取得了哪些改进？
   4. LLM的最新应用和用例是什么？
   5. 引入了哪些伦理和安全进展？

📚 [执行器智能体] 开始研究 5 个问题...

--- 问题 1/5 ---
🔍 [执行器智能体] 正在研究：2024年发布的LLM有哪些主要架构创新...
   找到 10 个搜索结果
   ✅ 已总结：2024年Transformer架构演进
   ✅ 已总结：专家混合：新范式
   ...

📝 [发布器智能体] 正在生成最终研究报告...
✅ [发布器智能体] 报告已生成（8547 字符）

================================================================================
研究完成
================================================================================

统计信息：
- 研究问题：5
- 咨询来源：23
- 生成摘要：23
- 报告长度：8547 字符
- 持续时间：3.2 分钟

================================================================================
最终研究报告
================================================================================

# 研究报告

## 元数据

- **研究查询**：2024年大型语言模型的最新进展是什么？
- **日期**：2024年12月6日
- **总来源数**：23
- **研究持续时间**：3.2 分钟

---

## 执行摘要

2024年见证了大型语言模型（LLM）的显著进步...

[完整报告继续...]

## 参考文献

[1] 2024年Transformer架构演进 - https://...
[2] 专家混合：新范式 - https://...
...
```

## 报告类型

### 研究报告（默认）
全面的学术风格报告，包含：
- 执行摘要
- 按主题组织的详细章节
- 深入分析和证据
- 明确的结论和建议
- 2000+ 字

### 大纲报告
结构化大纲格式，包含：
- 分层组织
- 要点和标题
- 简洁的关键点
- 深入探索路线图

### 资源报告
精选资源指南，包含：
- 按类型分类的来源
- 每个资源的简要注释
- 突出显示权威来源
- 访问信息和上下文

设置方式：`export REPORT_TYPE="outline_report"`

## 项目结构

```
gpt_researcher/
├── config.go              # 配置管理
├── state.go               # 研究状态定义
├── tools.go               # 网络搜索、爬虫、总结器
├── planner_agent.go       # 问题生成智能体
├── execution_agent.go     # 信息收集智能体
├── publisher_agent.go     # 报告生成智能体
├── gpt_researcher.go      # 主工作流编排
├── main.go                # 示例应用
├── go.mod                 # Go 模块定义
├── README.md              # 英文文档
└── README_CN.md           # 本文件
```

## 与原始 Python 实现的比较

| 功能 | Python (assafelovic/gpt-researcher) | Go（本实现） |
|------|-----------------------------------|-------------|
| 规划器智能体 | ✅ | ✅ |
| 执行器智能体 | ✅ | ✅ |
| 发布器智能体 | ✅ | ✅ |
| Tavily 搜索 | ✅ | ✅ |
| 网络爬虫 | ✅ | ✅（简化版） |
| PDF 支持 | ✅ | ⚠️（计划中） |
| 来源引用 | ✅ | ✅ |
| 多种报告类型 | ✅ | ✅ |
| FastAPI 后端 | ✅ | ❌（仅CLI） |
| NextJS 前端 | ✅ | ❌（仅CLI） |
| 导出为 PDF/DOCX | ✅ | ⚠️（计划中） |
| 本地文档 | ✅ | ⚠️（计划中） |

## 最佳实践

### 1. 编写清晰的查询

✅ **好的例子**：
- "2024年量子计算的最新进展是什么？"
- "人工智能如何应用于医疗保健？"
- "远程工作的经济影响是什么？"

❌ **太模糊**：
- "告诉我关于AI的事"
- "科技有什么新东西？"

### 2. 调整配置

对于**快速研究**（更快、更便宜）：
```bash
export MAX_QUESTIONS="3"
export MAX_SOURCES_TO_USE="10"
export SUMMARY_MODEL="gpt-3.5-turbo"
export GPT_MODEL="gpt-3.5-turbo"
```

对于**深度研究**（彻底、详细）：
```bash
export MAX_QUESTIONS="10"
export MAX_SOURCES_TO_USE="30"
export GPT_MODEL="gpt-4"
export REPORT_MODEL="gpt-4"
```

### 3. 监控 API 成本

- 每次研究会话进行约 20-50 次 API 调用
- 对摘要使用更便宜的模型（gpt-3.5-turbo）
- 限制 `MAX_QUESTIONS` 和 `MAX_SOURCES_TO_USE`
- 启用 `VERBOSE="true"` 跟踪进度

### 4. 审查引用

始终验证参考文献部分的来源：
- 检查 URL 有效性
- 评估来源可信度
- 审查原始内容的准确性

## 故障排除

### API 密钥错误

```
Warning: OPENAI_API_KEY not set
Warning: TAVILY_API_KEY not set
```

**解决方案**：设置环境变量：
```bash
export OPENAI_API_KEY="sk-..."
export TAVILY_API_KEY="tvly-..."
```

### 速率限制

如果遇到速率限制：
- 减少 `MAX_QUESTIONS`（例如，减至 3）
- 减少 `MAX_SOURCES_TO_USE`（例如，减至 10）
- 在请求之间添加延迟
- 使用不同的 API 层级

### 报告空白或质量差

如果报告不充分：
- 验证 Tavily API 密钥有效
- 使查询更具体
- 增加 `MAX_QUESTIONS` 和 `MAX_SOURCES_TO_USE`
- 尝试不同的模型（gpt-4 vs gpt-3.5-turbo）

### 网络爬虫失败

某些网站可能阻止爬虫：
- 这是正常行为
- 系统将跳过失败的来源
- 增加 `MAX_SOURCES_TO_USE` 以提供冗余

## 高级用法

### 编程使用

```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    // 创建配置
    config := NewConfig()
    config.Verbose = false
    config.MaxQuestions = 3

    // 创建研究器
    researcher, err := NewGPTResearcher(config)
    if err != nil {
        log.Fatal(err)
    }

    // 进行研究
    ctx := context.Background()
    state, err := researcher.ConductResearch(ctx, "您的查询")
    if err != nil {
        log.Fatal(err)
    }

    // 访问结果
    fmt.Printf("问题：%v\n", state.Questions)
    fmt.Printf("来源：%d\n", len(state.Sources))
    fmt.Printf("报告：%s\n", state.FinalReport)
}
```

### 自定义工具集成

使用自定义工具扩展：

```go
// 添加自定义搜索工具
type CustomSearchTool struct{}

func (t *CustomSearchTool) Name() string {
    return "custom_search"
}

func (t *CustomSearchTool) Description() string {
    return "使用自定义API搜索"
}

func (t *CustomSearchTool) Call(ctx context.Context, input string) (string, error) {
    // 您的自定义实现
    return results, nil
}
```

## 性能

### 典型研究会话

- **持续时间**：3-5 分钟
- **API 调用**：25-50（取决于配置）
- **来源**：15-25 个独特来源
- **报告长度**：2000-4000 字
- **成本**：约 $0.50-2.00（使用 GPT-4）

### 优化技巧

1. **对摘要使用 gpt-3.5-turbo**（成本降低 80%）
2. **限制问题数**为大多数查询的 3-5 个
3. **批处理请求**（如果可能）
4. **缓存结果**用于重复查询

## 未来增强

计划功能：
- [ ] PDF 文档分析
- [ ] 本地文档搜索
- [ ] 导出为 PDF/DOCX
- [ ] 图像分析和包含
- [ ] 多语言支持
- [ ] 自定义报告模板
- [ ] Web UI / API 服务器
- [ ] 并行问题执行
- [ ] 来源质量评分
- [ ] 事实验证

## 许可证

本实现遵循与 langgraphgo 项目相同的许可证。

## 参考资料

- [原始 Python gpt-researcher](https://github.com/assafelovic/gpt-researcher)
- [LangGraph 文档](https://python.langchain.com/docs/langgraph)
- [Tavily 搜索 API](https://www.tavily.com/)
- [LangChain Go](https://github.com/tmc/langchaingo)

## 贡献

欢迎贡献！改进领域：
- 增强网络爬虫
- PDF/文档处理
- 额外的导出格式
- 性能优化
- 测试覆盖率

## 支持

对于问题和疑问：
- 查看上述故障排除部分
- 查看本 README 中的示例
- 在 GitHub 上开启 issue

---

**构建工具**：
- [langgraphgo](https://github.com/smallnest/langgraphgo) - 基于图的智能体编排
- [langchaingo](https://github.com/tmc/langchaingo) - LLM 集成
- [Tavily](https://www.tavily.com/) - 网络搜索 API
- [OpenAI](https://openai.com/) - 语言模型
