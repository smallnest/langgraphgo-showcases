# LangGraph Go 示例集

基于 [LangGraph Go](https://github.com/smallnest/langgraphgo) 构建的 AI 智能体示例和演示项目集合。本仓库从 LangGraph Go 主项目的 showcases 目录中独立出来，提供了可直接运行的、生产级的 AI 智能体实现示例。

## 概述

本仓库包含 9 个全面的 AI 智能体实现，展示了使用 LangGraph Go 和 LangChain Go 构建智能系统的不同用例和模式。每个示例都是一个完整的、可运行的项目，配有自己的 README、示例和文档。

## 目录

- [示例项目](#示例项目)
  - [BettaFish - 公共舆情分析](#bettafish)
  - [GPT Researcher - 自主研究智能体](#gpt-researcher)
  - [Health Insights Agent - 医疗报告分析器](#health-insights-agent)
  - [PeopleHub - 人物研究智能体](#peoplehub)
  - [DeepAgents - 文件系统感知智能体](#deepagents)
  - [DeerFlow - 深度研究智能体](#deerflow)
  - [LangManus - 多智能体自动化框架](#langmanus)
  - [Profile - 数字足迹分析器](#profile)
  - [AI PDF Chatbot - 文档问答系统](#ai-pdf-chatbot)
- [快速开始](#快速开始)
- [环境要求](#环境要求)
- [许可证](#许可证)

## 示例项目

### BettaFish

**多智能体架构的深度公共舆情分析**

[BettaFish](https://github.com/666ghj/BettaFish) 项目的完整 Go 实现，包含以下特性：
- QueryEngine 带反思循环的迭代搜索优化
- MediaEngine 进行相关图像搜索
- ForumEngine 实现智能体间的 LLM 驱动讨论
- ReportEngine 生成综合性 Markdown 报告

[📂 查看 BettaFish →](./BettaFish)

**核心技术**：Tavily API、多智能体讨论、反思模式

---

### GPT Researcher

**生成综合报告的自主研究智能体**

[gpt-researcher](https://github.com/assafelovic/gpt-researcher) 的 Go 移植版，可自动化研究任务：
- 根据查询生成聚焦的研究问题
- 从 20+ 个网络源收集信息
- 生成带引用的详细 2000+ 字报告
- 支持多种报告类型（研究报告、大纲、资源报告）

[📂 查看 GPT Researcher →](./gpt_researcher)

**核心技术**：Tavily 搜索、网页抓取、多源聚合

---

### Health Insights Agent

**AI 驱动的血液报告分析**

智能健康分析系统，处理血液检测报告：
- 从文本和 PDF 报告中提取参数
- 识别潜在健康风险并评级
- 提供个性化建议（饮食、生活方式、医疗）
- 生成带置信度评分的结构化分析

[📂 查看 Health Insights Agent →](./health_insights_agent)

**核心技术**：基于 LLM 的提取、医学知识、PDF 处理

---

### PeopleHub

**自动化人物研究智能体**

[PeopleHub](https://github.com/MeirKaD/pepolehub) 的 Go 实现，用于研究个人信息：
- 通过网络搜索获取 LinkedIn 资料
- 执行全面的 Google 搜索
- 抓取并总结相关网页
- 生成详细的人物研究报告

[📂 查看 PeopleHub →](./pepolehub)

**核心技术**：并行执行、网页抓取、资料聚合

---

### DeepAgents

**具有任务管理功能的文件系统感知智能体**

强大的智能体框架，具备文件系统访问能力：
- 在工作区内读取、写入和管理文件
- 内置待办事项列表管理
- 子智能体委托处理复杂任务
- 基于模式的文件查找，支持 glob 匹配

[📂 查看 DeepAgents →](./deepagents)

**核心技术**：文件系统工具、任务委托、层次化智能体

---

### DeerFlow

**带 Web 界面的深度研究智能体**

[字节跳动 DeerFlow](https://github.com/bytedance/deer-flow) 的 Go 实现：
- 多智能体架构（规划器 → 研究员 → 报告员 → 播客）
- 通过服务器推送事件（SSE）实时进度更新
- 现代化深色主题 Web 界面
- 研究历史和结果缓存
- 可选的播客脚本生成

[📂 查看 DeerFlow →](./deerflow)

**核心技术**：SSE 流式传输、Web UI、播客生成、缓存

---

### LangManus

**多智能体 AI 自动化框架**

[LangManus](https://github.com/Darwin-lfl/langmanus) 框架的 Go 实现：
- 包含 7 个专业智能体的分层架构
- 协调器 → 规划器 → 监督者 → 工作者 → 报告员
- 带安全控制的代码执行（Python/Bash）
- 通过 Tavily 集成网络搜索
- 支持实时更新的流式传输

[📂 查看 LangManus →](./langmanus)

**核心技术**：多智能体编排、代码执行、任务规划

---

### Profile

**数字足迹分析器**

用于分析互联网数字存在的 OSINT 工具：
- 搜索 1000+ 个社交媒体平台
- AI 驱动的心理画像
- 通过 SSE 实时流式反馈
- 精美的纸质纹理 UI 设计
- 注重隐私（仅公开数据）

[📂 查看 Profile →](./profile)

**核心技术**：OSINT、社交媒体搜索、心理分析

**在线演示**：http://profile.rpcx.io

---

### AI PDF Chatbot

**智能文档问答系统**

完整的 RAG（检索增强生成）全栈应用：
- 上传并提取 PDF 文档
- 基于向量的语义搜索
- 对话式问答界面
- 后端：使用 LangChain Go 的 Go 实现
- 前端：现代化 Web 界面

[📂 查看 AI PDF Chatbot →](./ai-pdf-chatbot)

**核心技术**：RAG、向量数据库、PDF 处理、聊天界面

---

## 快速开始

每个示例都是独立的，拥有自己的依赖和设置说明。

### 通用前置要求

- **Go**：版本 1.21 或更高
- **API 密钥**：大多数示例需要：
  - OpenAI API 密钥（或兼容提供商）
  - Tavily API 密钥（用于网络搜索功能）

### 快速启动步骤

1. 从上面的列表中选择一个示例
2. 导航到其目录：
   ```bash
   cd <示例名称>
   ```
3. 按照该目录中的 README 进行特定设置
4. 设置所需的环境变量：
   ```bash
   export OPENAI_API_KEY="your-key-here"
   export TAVILY_API_KEY="your-tavily-key"  # 如果需要
   ```
5. 运行示例：
   ```bash
   go run *.go
   # 或
   go build && ./<示例名称>
   ```

## 环境要求

### 公共依赖

所有示例使用：
- [LangGraph Go](https://github.com/smallnest/langgraphgo) - 基于图的智能体编排
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM 集成库

### API 服务

大多数示例集成了：
- **OpenAI API** - 用于 LLM 功能（GPT-4、GPT-3.5 等）
- **Tavily API** - 用于网络搜索功能
- **替代提供商** - 许多支持 OpenAI 兼容 API（DeepSeek、Azure、Ollama 等）

### 环境配置

常用环境变量：
```bash
# 必需
OPENAI_API_KEY="sk-..."

# 可选
OPENAI_API_BASE="https://api.openai.com/v1"  # 自定义端点
OPENAI_MODEL="gpt-4o"                        # 模型选择
TAVILY_API_KEY="tvly-..."                    # 用于搜索功能
```

## 项目结构

```
langgraphgo-showcases/
├── BettaFish/              # 公共舆情分析
├── gpt_researcher/         # 自主研究智能体
├── health_insights_agent/  # 医疗报告分析器
├── pepolehub/             # 人物研究智能体
├── deepagents/            # 文件系统感知智能体
├── deerflow/              # 带 Web UI 的深度研究
├── langmanus/             # 多智能体框架
├── profile/               # 数字足迹分析器
├── ai-pdf-chatbot/        # PDF 问答系统
├── LICENSE                # MIT 许可证
└── README.md              # 英文说明文档
```

## 跨示例的功能特性

### 智能体模式
- ✅ 多智能体编排
- ✅ 层次化任务委托
- ✅ 反思和自我改进循环
- ✅ 并行执行
- ✅ 使用 LangGraph 进行状态管理

### 功能能力
- ✅ 网络搜索和抓取
- ✅ 文档处理（PDF、文本）
- ✅ 代码生成和执行
- ✅ 报告生成（Markdown、HTML）
- ✅ 实时流式传输
- ✅ 文件系统操作
- ✅ 任务规划和分解

### 交互界面
- ✅ 命令行界面（CLI）
- ✅ 现代化 UI 的 Web 界面
- ✅ 编程 API
- ✅ 流式传输支持（SSE）

## 使用场景

这些示例展示了以下场景的解决方案：

- 📊 **研究与分析**：自动化研究、数据收集、报告生成
- 🏥 **医疗保健**：医疗报告分析、健康洞察
- 🔍 **OSINT**：公开数据聚合、数字足迹分析
- 📝 **文档处理**：PDF 分析、问答系统
- 💼 **自动化**：任务编排、代码执行、工作流自动化
- 🌐 **Web 智能**：搜索、抓取、内容综合

## 学习路径

**初级**：从简单的示例开始
1. PeopleHub - 基础多步骤工作流
2. GPT Researcher - 经典研究模式

**中级**：探索更复杂的模式
3. Health Insights Agent - 结构化提取
4. DeepAgents - 文件系统和工具
5. AI PDF Chatbot - RAG 实现

**高级**：多智能体系统
6. BettaFish - 反思循环
7. DeerFlow - 全栈带 UI
8. LangManus - 复杂编排

## 贡献

欢迎贡献！无论是：
- Bug 修复
- 新示例
- 文档改进
- 性能优化
- 测试覆盖

请随时提出 issue 或提交 pull request。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

这些示例受到以下优秀开源项目的启发并进行了移植：
- [BettaFish](https://github.com/666ghj/BettaFish) - 原始 Python 实现
- [gpt-researcher](https://github.com/assafelovic/gpt-researcher) - 研究自动化
- [PeopleHub](https://github.com/MeirKaD/pepolehub) - 人物研究
- [ByteDance DeerFlow](https://github.com/bytedance/deer-flow) - 深度研究
- [LangManus](https://github.com/Darwin-lfl/langmanus) - 多智能体框架
- [hia](https://github.com/harshhh28/hia) - 健康洞察

特别感谢：
- [LangGraph Go](https://github.com/smallnest/langgraphgo) - 框架基础
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM 集成
- [Tavily](https://www.tavily.com/) - 网络搜索 API
- [OpenAI](https://openai.com/) - 语言模型

## 相关资源

- [LangGraph Go 文档](https://github.com/smallnest/langgraphgo)
- [LangChain Go 文档](https://github.com/tmc/langchaingo)
- [LangGraph (Python) 文档](https://python.langchain.com/docs/langgraph)

## 支持

如有问题、issue 或讨论：
- 在本仓库中提出 issue
- 查看各个示例的 README
- 访问主 [LangGraph Go 仓库](https://github.com/smallnest/langgraphgo)

---

**使用 LangGraph Go 和 LangChain Go 精心构建 ❤️**
