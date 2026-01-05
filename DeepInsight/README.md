# DeepInsight - 深度调研与洞察分析系统

基于 LangGraphGo 的多智能体深度调研和洞察分析系统，专注于深度研究而非舆情分析。

## 与 BettaFish 的区别

| 特性 | BettaFish | DeepInsight |
|------|-----------|-------------|
| **核心关注** | 舆情分析、情感倾向 | 深度调研、专家洞察 |
| **数据来源** | 社交媒体、网民评论 | 学术研究、专家观点、行业报告 |
| **分析重点** | 情感分析、平台差异 | 因果分析、机制研究、理论框架 |
| **输出内容** | 舆情报告 | 深度洞察报告 |
| **关键指标** | 情感倾向、传播数据 | 关键洞察、专家观点、证据链条 |

## 核心组件

### 1. Schema (schema/state.go)
- **DeepInsightState**: 全局状态管理
- **Insight**: 关键洞察和发现结构
- **SearchResult**: 包含研究分析字段（专家观点、理论支撑、数据质量）
- **GraphRAG**: 知识图谱支持（实体、概念、事件、理论、方法论）

### 2. QueryEngine (query_engine/)
- **功能**: 深度调研和资料收集
- **特点**:
  - 学术导向的搜索策略
  - 关注研究方法和数据来源
  - 多维度、跨学科的搜索
- **工具**: 6种专业搜索工具（基础搜索、深度搜索、最新资料、图片搜索、按日期搜索）

### 3. InsightEngine (insight_engine/)
- **功能**: 专家洞察分析和深度解读
- **特点**:
  - 专家观点和理论支撑
  - 因果分析和机制研究
  - 证据链构建和案例验证
  - 洞察提炼和趋势预测

### 4. MediaEngine (media_engine/)
- **功能**: 视觉和上下文信息收集
- **特点**:
  - 图表和数据可视化
  - 多模态内容分析
  - 上下文信息整合

### 5. ForumEngine (forum_engine/)
- **功能**: 专家讨论会模拟
- **角色**:
  - **Moderator**: 讨论主持人，整合观点
  - **QueryExpert**: 研究查询专家，提供实证数据
  - **MediaExpert**: 多媒体分析专家，关注视觉信息
  - **InsightExpert**: 洞察分析专家，提供理论支撑

### 6. ReportEngine (report_engine/)
- **功能**: 深度洞察报告生成
- **输出**: 格式化的 Markdown 报告

## 主要变化（相比 BettaFish）

### State Schema 变化
```go
// 旧版 (BettaFish) - 舆情分析字段
Sentiment     string
SentimentScore float64
Platform      string

// 新版 (DeepInsight) - 深度洞察字段
KeyInsights      []string    // 关键洞察
SourceType       string      // 来源类型（学术/行业/政府/专家）
Reliability      float64     // 来源可靠性
Methodology      string      // 研究方法论
DataQuality      string      // 数据质量
```

### Prompt 策略变化

#### QueryEngine
- **BettaFish**: 关注新闻收集、事实核查、破除谣言
- **DeepInsight**: 关注研究方法论、理论框架、学术观点、实证研究

#### InsightEngine
- **BettaFish**: 关注社交媒体情绪、网民观点、平台差异
- **DeepInsight**: 关注专家观点、因果分析、机制研究、洞察提炼

#### ForumEngine
- **BettaFish**: 舆情分析师讨论情感倾向和传播效果
- **DeepInsight**: 专家讨论会分析理论观点和研究结论

## 使用方法

### 环境变量
```bash
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_API_BASE="https://api.openai.com/v1"  # 可选
export OPENAI_MODEL="gpt-4"  # 可选
export TAVILY_API_KEY="your-tavily-api-key"
```

### 运行
```bash
cd showcases/DeepInsight

# 基本用法
go run main.go "人工智能的发展趋势"

# 指定输出文件
go run main.go -o report.md "人工智能的发展趋势"

# 简单模式（跳过深度洞察和专家讨论，更快速）
go run main.go -simple "人工智能的发展趋势"

# 组合使用
go run main.go -o output.md -simple "Claude Code使用经验总结"
```

### 参数说明
- `-o <文件>`: 指定输出文件路径（可选）
- `-simple`: 简单模式，跳过 `insight_engine` 和 `forum_engine`，仅进行基础调研和媒体搜索（可选）

### 输出
生成一个包含以下内容的深度洞察报告：
1. **研究摘要**: 核心研究发现和洞察
2. **详细分析**: 多个深度调研段落
3. **视觉信息**: 图表和数据可视化
4. **专家讨论**: 多轮专家讨论记录

## 项目结构
```
DeepInsight/
├── main.go                    # 入口文件
├── schema/
│   └── state.go              # 状态管理（包含洞察结构）
├── query_engine/
│   ├── agent.go              # 深度调研节点
│   ├── prompts.go            # 研究方法论提示词
│   └── tools.go              # Tavily 搜索工具
├── insight_engine/
│   ├── agent.go              # 洞察分析节点
│   └── prompts.go            # 专家洞察提示词
├── media_engine/
│   ├── agent.go              # 多媒体搜索节点
│   └── prompts.go            # 多模态内容提示词
├── forum_engine/
│   └── agent.go              # 专家讨论会节点
├── report_engine/
│   ├── agent.go              # 报告生成节点
│   └── template.go           # 报告模板
└── README.md                 # 本文件
```

## 核心特点

1. **研究导向**: 从舆情分析转向深度研究，关注理论、方法和实证
2. **专家视角**: 强调专家观点、理论框架和学术研究
3. **因果分析**: 不仅描述现象，更关注原因和机制
4. **洞察提炼**: 从大量信息中提炼有价值的深层洞察
5. **多模态整合**: 结合文字、图像、数据进行综合分析

## 适用场景

- 学术研究辅助
- 行业深度分析
- 技术趋势研究
- 市场洞察分析
- 政策研究评估

## 未来改进方向

1. 添加学术数据库搜索（Google Scholar、arXiv）
2. 集成研究论文 PDF 解析
3. 支持引用管理和参考文献生成
4. 添加研究方法评估工具
5. 支持 GraphRAG 知识图谱增强
