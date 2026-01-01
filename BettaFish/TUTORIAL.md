# 🐠 微舆（BettaFish）：人人可用的舆情分析助手

> "打破信息茧房，还原舆情真相，预测未来走向，辅助科学决策"

## 🎯 什么是舆情分析？微舆为何而生？

### 舆情分析的三大困境

在信息爆炸的时代，我们面临着三个严重的问题：

#### 1️⃣ **信息茧房**：你看到的，只是算法想让你看到的

```
某品牌公关：
  - 刷小红书：全是好评，"完美产品！"
  - 看微博：发现大量投诉，"质量问题严重"
  - 逛知乎：技术大神吐槽，"设计缺陷"

真相是什么？你被困在哪个茧房里？
```

#### 2️⃣ **信息过载**：数据太多，真相淹没

```
调研"新能源汽车"舆情：
  - 新闻：10,000+ 篇
  - 社交媒体：100,000+ 条讨论
  - 论坛评论：500,000+ 条
  - 视频评论：1,000,000+ 条

哪些是真实声音？哪些是水军？哪些是关键趋势？
```

#### 3️⃣ **主观偏见**：人类的认知局限

```
分析某政策影响：
  - 你的立场会影响你选择看什么
  - 你的经验会影响你如何解读
  - 你的情绪会影响你的判断

如何做到客观公正？
```

### 🐠 微舆（BettaFish）：AI驱动的舆情分析解决方案

**BettaFish** = **Betta**（斗鱼：小而强大）+ **Fish**（在信息海洋中游弋）
**微舆** = 微观舆情，从细节看全局

**核心使命**：
- 🌐 **打破信息茧房** - 全网多维度信息收集
- 🎯 **还原舆情原貌** - 多视角客观分析
- 🔮 **预测未来走向** - 数据驱动趋势预判
- 💡 **辅助科学决策** - 提供可执行建议

---

## 🚀 项目背景：致敬原作，Go语言移植

### 原版BettaFish：Python多Agent舆情分析的开创者

**衷心感谢** **[666ghj/BettaFish](https://github.com/666ghj/BettaFish)** Python版本的原创者！

这是一个极具创新性的多Agent舆情分析系统，开创性地实现了：

- 🌟 **多Agent协作架构** - 5大智能引擎分工协作的设计
- 🌟 **反思循环机制** - QueryEngine的自我迭代优化思想
- 🌟 **虚拟圆桌会议** - ForumEngine的多视角整合方法
- 🌟 **从0实现** - 不依赖重型框架的工程实践

**本Go版本是在原作基础上的学习和移植，所有核心思想均来自原作者的创新设计。**

---

### Go语言移植版：基于LangGraphGo实现

使用 **[LangGraphGo](https://github.com/smallnest/langgraphgo)** 多Agent编排框架，我们将原版的优秀设计移植到了Go语言实现。

#### 🎯 为什么选择LangGraphGo？

**LangGraphGo** 是一个轻量级的Go语言多Agent编排库，提供：

```go
// 1. 图结构的Agent编排
workflow := graph.NewStateGraph()
workflow.AddNode("query_engine", "查询引擎", QueryEngineNode)
workflow.AddNode("media_engine", "媒体引擎", MediaEngineNode)
workflow.AddEdge("query_engine", "media_engine")

// 2. 统一的状态管理
type BettaFishState struct {
    Query          string
    SearchResults  []SearchResult
    Insights       []Insight
    Paragraphs     []string
}

// 3. 并行执行支持
// 多个节点自动并行处理，性能优化

// 4. 简洁的API
app, _ := workflow.Compile()
result, _ := app.Invoke(ctx, initialState)
```

**核心优势**：
- ✅ **轻量级** - 不是重型框架，只提供必要的编排能力
- ✅ **Go原生** - 充分利用Go的并发优势
- ✅ **类型安全** - 编译期类型检查
- ✅ **易于理解** - API简洁，学习成本低

---

### 🔧 Go语言实现特点

#### 1️⃣ **部署方式**

Go版本提供简洁的部署方式：

```bash
# 编译成单一可执行文件
go build -o bettafish main.go

# 直接运行
./bettafish "查询"

# 跨平台编译
GOOS=linux GOARCH=amd64 go build -o bettafish-linux
GOOS=windows GOARCH=amd64 go build -o bettafish.exe
```

#### 3️⃣ **并发实现**

Go语言的goroutine天然支持并发处理：

```go
// 并行处理多个段落
var wg sync.WaitGroup
for _, p := range paragraphs {
    wg.Add(1)
    go func(para *Paragraph) {
        defer wg.Done()
        processParagraph(para)
    }(p)
}
wg.Wait()
```

#### 4️⃣ **功能完整性**

Go版本完整实现了原版的所有核心功能：

| 功能模块          | 实现状态 | 说明                 |
| ----------------- | -------- | -------------------- |
| **QueryEngine**   | ✅        | 反思循环完整实现     |
| **MediaEngine**   | ✅        | 图片搜索和筛选       |
| **InsightEngine** | ✅        | 数据深度挖掘         |
| **ForumEngine**   | ✅        | 虚拟圆桌会议         |
| **ReportEngine**  | ✅        | Markdown报告生成     |
| **多轮反思**      | ✅        | 可配置迭代次数       |
| **并行处理**      | ✅        | 支持多段落并行       |
| **错误处理**      | ✅        | 详细的错误信息       |
| **Tavily集成**    | ✅        | 6种搜索工具          |
| **多模型支持**    | ✅        | OpenAI/DeepSeek/本地 |

---

### 🎓 技术学习价值

这个项目是学习以下技术的绝佳案例：

1. **多Agent系统架构** - 5大引擎如何协作
2. **LangGraphGo使用** - 如何编排复杂的Agent流程
3. **Go并发编程** - goroutine和channel的实战应用
4. **LLM应用开发** - 如何设计Prompt和处理响应
5. **状态管理** - 多Agent间如何共享和传递状态
6. **错误处理** - 如何优雅地处理LLM返回的异常

---

## 🔬 核心能力：微舆的舆情分析四板斧

微舆如何做到打破茧房、还原真相、预测未来、辅助决策？答案是**五大智能引擎的协同工作**。

### 第一板斧：全网信息收集 - 打破信息茧房

**主角**：QueryEngine（信息猎手）

**解决的问题**：
- ❌ 算法推荐让你只看到片面信息
- ❌ 单一渠道无法覆盖全貌
- ❌ 信息质量参差不齐

**如何打破茧房**：

```
任务：分析"某品牌手机"的舆情

传统方式：
  → 只看一个平台（微博 OR 知乎 OR 小红书）
  → 只看热门内容（算法推荐的）
  → 只看前几页（人工精力有限）

微舆 - QueryEngine：
  ✓ 多平台覆盖（新闻、社交媒体、论坛、视频平台...）
  ✓ 深度挖掘（不止看热门，也看长尾内容）
  ✓ 智能筛选（AI识别有价值信息，过滤水军噪音）
  ✓ 质量控制（反思循环机制，确保信息可靠）
```

**反思循环 - 追求完美的信息质量**：

```
第1轮搜索 → AI评估：质量72分，缺乏深度
  ↓
优化关键词，第2轮搜索 → AI评估：质量88分，基本满意
  ↓
精准查询，第3轮搜索 → AI评估：质量95分，完美！
  ↓
输出高质量信息集合
```

**实战案例**：
```
查询："华为Mate60舆情分析"

收集到的信息：
  - 官方新闻：产品发布、技术亮点
  - 科技媒体：专业评测、技术解析
  - 社交媒体：真实用户反馈、使用体验
  - 电商评论：购买动机、满意度
  - 论坛讨论：深度技术讨论、对比分析
  - 视频内容：开箱评测、使用教程

✅ 全方位覆盖，打破单一平台茧房
```

---

### 第二板斧：多视角整合 - 还原舆情原貌

**主角**：ForumEngine（真相还原者）

**解决的问题**：
- ❌ 单一视角容易产生偏见
- ❌ 极端声音掩盖真实声音
- ❌ 缺乏系统性整合

**如何还原真相**：

```
舆情分析 = 模拟一场圆桌会议

参会者：
  - Moderator：主持人，引导讨论、整合观点
  - QueryAgent：代表新闻媒体视角，提供最新动态
  - MediaAgent：代表多模态内容分析，关注视觉传播
  - InsightAgent：代表数据分析结果（通过状态传递）

讨论模式（5轮多Agent对话）：
  第1轮：Moderator 开场 - 梳理背景，提出议题
  第2轮：QueryAgent 发言 - 新闻视角分析
  第3轮：MediaAgent 发言 - 媒体视角补充
  第4轮：QueryAgent 补充 - 基于前面讨论深化
  第5轮：Moderator 总结 - 整合观点，得出结论

核心特点：
  ✓ 多轮对话（不是一次性发言，而是互相回应）
  ✓ 历史记忆（每个Agent都能看到之前的讨论）
  ✓ 思维碰撞（发现矛盾与共识，纠正错误）
  ✓ 观点演进（后续发言会基于前面的讨论优化）
```

**实战案例**（多轮讨论过程）：
```
议题："ChatGPT对就业市场的影响"

第1轮 - Moderator 开场：
  "今天讨论ChatGPT对就业的影响。请各位从不同角度分析。"

第2轮 - QueryAgent（新闻视角）：
  "主流媒体报道：AI将创造新就业机会，政府出台支持政策"

第3轮 - MediaAgent（媒体视角）：
  "社交媒体上大量焦虑情绪，'被替代'成高频词，短视频传播很快"

第4轮 - QueryAgent（基于前面讨论的补充）：
  "刚才MediaAgent提到焦虑情绪，我补充数据：
   - AI岗位需求增长180%（新机会）
   - 文案/客服岗位需求下降15%（确实存在冲击）
   这验证了社交媒体的焦虑有数据支撑，但也有新机会"

第5轮 - Moderator 总结：
  "综合各方观点，形成以下结论：
  ✓ 短期：确实存在结构性失业风险（数据+民间焦虑）
  ✓ 中长期：催生新职业机会（官方政策+市场数据）
  ✓ 关键矛盾：转型速度跟不上技术变化
  ✓ 行动建议：政府加强职业培训，企业给员工学习时间"

✅ 通过5轮多Agent对话，观点逐步深化，最终得出全面结论
✅ 后续Agent能看到前面的讨论，形成真正的"思维碰撞"
```

---

### 第三板斧：深度数据挖掘 - 预测未来走向

**主角**：InsightEngine（趋势预言家）

**解决的问题**：
- ❌ 只看表面现象，看不到深层规律
- ❌ 凭直觉判断，缺乏数据支撑
- ❌ 无法预测趋势，被动应对

**如何预测趋势**：

```
表层信息 → 深层洞察 → 趋势预测

第1层：看到什么（现象）
  "某品牌负面评论增多"

第2层：为什么（原因）
  "产品质量问题 + 售后响应慢 + 竞品压力"

第3层：意味着什么（洞察）
  "品牌信任度下降，客户流失风险上升"

第4层：未来会怎样（预测）
  "3个月内市场份额可能下降5-8%，需紧急公关"
```

**数据驱动的趋势分析**：

```
分析维度：
  📊 舆情数量变化趋势
  📈 情感倾向演变曲线
  🔥 热点话题生命周期
  👥 关键意见领袖影响力
  🌐 地域/人群分布特征
  ⚡ 突发事件响应速度

预测模型：
  - 时间序列分析
  - 情感演变追踪
  - 话题扩散模型
  - 危机预警算法
```

**实战案例**：
```
任务："预测某新能源车企的舆情走向"

数据发现：
  📊 近30天负面舆情增长35%
  📈 "自燃"相关讨论量暴涨200%
  🔥 话题扩散速度：平均12小时到达10万阅读
  👥 汽车博主开始质疑，粉丝开始动摇
  🌐 一线城市用户讨论最活跃

InsightEngine 预测：
  ⚠️ 危机等级：中高
  ⏰ 发酵时间：预计3-5天达到峰值
  💥 影响范围：可能影响Q4销量10-15%
  🎯 关键窗口：未来48小时是黄金处置期

建议行动：
  1. 立即发布官方技术说明
  2. 邀请第三方机构检测
  3. 推出用户关怀计划
  4. 启动KOL沟通

✅ 提前预警，变被动为主动
```

---

### 第四板斧：可视化呈现 - 辅助科学决策

**主角**：MediaEngine（可视化专家） + ReportEngine（决策助手）

**解决的问题**：
- ❌ 数据太多，看不懂
- ❌ 缺乏可视化，难以直观理解
- ❌ 没有行动建议，不知道怎么办

**如何辅助决策**：

```
数据 → 可视化 → 洞察 → 决策建议

MediaEngine：
  ✓ 舆情态势图（一眼看懂整体情况）
  ✓ 情感分布饼图（正/中/负比例）
  ✓ 趋势折线图（舆情演变轨迹）
  ✓ 热力图（地域/人群分布）
  ✓ 词云图（高频关键词）

ReportEngine：
  ✓ 执行摘要（3分钟看懂核心）
  ✓ 详细分析（15分钟深入理解）
  ✓ 数据支撑（有理有据）
  ✓ 行动建议（可执行方案）
```

**决策报告结构**：

```markdown
# 舆情分析报告

## 📊 核心发现（3句话说清楚）
  - 当前态势：...
  - 关键问题：...
  - 紧急程度：...

## 📈 数据可视化（一图胜千言）
  [舆情趋势图] [情感分布图] [热点词云]

## 🔍 深度分析（为什么是这样）
  - 舆情来源分析
  - 关键意见领袖
  - 讨论热点聚焦
  - 潜在风险识别

## 🔮 趋势预测（未来会怎样）
  - 短期走势（24-72小时）
  - 中期影响（1-4周）
  - 长期趋势（1-3个月）

## 💡 行动建议（该怎么办）
  - 紧急措施（24小时内）
  - 短期策略（1周内）
  - 长期规划（1个月+）

## 📚 附录（详细数据）
  - 数据来源
  - 分析方法
  - 置信度说明
```

**实战案例**：
```
报告标题：《某品牌口红舆情危机处置建议》

核心发现：
  ⚠️ 负面舆情爆发，24小时内阅读量破500万
  🔥 核心问题：产品疑似含有害成分
  ⏰ 紧急程度：高（需12小时内响应）

可视化：
  [折线图] 舆情爆发曲线 - 指数级增长
  [饼图] 情感分布 - 负面68% 中性25% 正面7%
  [词云] 高频词 - "有害" "退货" "失望" "维权"

行动建议：
  ⏰ 12小时内：
    1. 官方声明：公布检测报告
    2. 紧急下架：主动召回问题批次
    3. 赔偿方案：无条件退货+补偿

  📅 1周内：
    4. 第三方检测：邀请权威机构背书
    5. KOL沟通：争取意见领袖支持
    6. 用户沟通：建立专项客服团队

  📆 1个月：
    7. 品牌重塑：推出全新质量保障体系
    8. 公关传播：正面案例宣传

✅ 清晰决策路径，可执行性强
```

---

## 🚀 快速上手：三步开启舆情分析

### 第一步：准备 API 密钥

微舆需要两个API密钥才能工作：

```bash
# OpenAI API - AI分析大脑
export OPENAI_API_KEY="sk-你的密钥"

# Tavily API - 全网信息搜索
export TAVILY_API_KEY="tvly-你的密钥"
```

**获取方式**：
- OpenAI API: https://platform.openai.com/
- Tavily API: https://tavily.com/

**可选配置**（使用其他模型）：
```bash
# 使用 DeepSeek（国产模型，价格亲民）
export OPENAI_API_BASE="https://api.deepseek.com/v1"
export OPENAI_MODEL="deepseek-chat"

# 使用本地 Ollama（完全免费）
export OPENAI_API_BASE="http://localhost:11434/v1"
export OPENAI_MODEL="llama3.1"
```

---

### 第二步：运行舆情分析

```bash
# 基础语法
go run showcases/BettaFish/main.go "你的舆情分析问题"

# 实战示例
go run showcases/BettaFish/main.go "分析某品牌手机新品发布后的用户反馈与市场舆情"
```

---

### 第三步：查看分析报告

微舆会自动生成完整的舆情分析报告，包含：

```
📊 舆情总体态势
  ├─ 情感倾向分布
  ├─ 舆情数量趋势
  └─ 关键指标统计

🔍 深度分析
  ├─ 多维度信息收集
  ├─ 多视角观点整合
  └─ 数据驱动洞察

🔮 趋势预测
  ├─ 短期走向预判
  ├─ 潜在风险识别
  └─ 发展趋势分析

💡 决策建议
  ├─ 应对策略
  ├─ 行动优先级
  └─ 可执行方案
```

**预计耗时**：1-2分钟
**报告位置**：`final_reports/` 目录

---

## 📋 舆情分析场景手册

### 场景一：品牌危机监测

**适用情况**：
- 新产品发布后的实时监控
- 突发负面事件的快速响应
- 危机公关的效果评估

**示例问题**：
```bash
go run main.go "某品牌产品质量问题的舆情态势与危机预警"
go run main.go "某企业负面新闻的网络传播分析与应对建议"
go run main.go "某公关声明发布后的舆论反馈与效果评估"
```

**关键输出**：
- ⚠️ 危机等级评估
- 📊 负面舆情占比
- 🔥 热点话题聚焦
- ⏰ 发酵速度预测
- 💡 应急处置建议

---

### 场景二：竞品对比分析

**适用情况**：
- 新品上市前的市场调研
- 竞争策略制定
- 产品定位优化

**示例问题**：
```bash
go run main.go "对比iPhone、华为、小米旗舰机的用户口碑与优劣势"
go run main.go "分析三大外卖平台的用户满意度与痛点差异"
go run main.go "新能源汽车三强的品牌形象与市场认知对比"
```

**关键输出**：
- 📊 各品牌舆情对比
- ⭐ 用户满意度排名
- 💪 各家优势特点
- 🔻 存在问题短板
- 🎯 市场机会空白

---

### 场景三：政策影响评估

**适用情况**：
- 新政策出台后的影响分析
- 行业监管变化的应对
- 政府沟通的决策支持

**示例问题**：
```bash
go run main.go "某行业监管新规的企业反应与市场影响评估"
go run main.go "某地产新政的购房者情绪与市场预期分析"
go run main.go "某教育改革政策的家长态度与社会舆论研判"
```

**关键输出**：
- 📜 政策解读多角度
- 😊 😐 😟 情绪分布
- 🏢 行业影响评估
- 👥 不同群体反应
- 💡 企业应对策略

---

### 场景四：产品口碑追踪

**适用情况**：
- 持续监控产品口碑
- 用户反馈收集分析
- 产品迭代决策支持

**示例问题**：
```bash
go run main.go "某APP最新版本的用户评价与功能反馈分析"
go run main.go "某品牌服装新款的消费者口碑与购买意愿"
go run main.go "某餐厅品牌的顾客评价与服务质量舆情"
```

**关键输出**：
- ⭐ 综合评分趋势
- 👍 好评聚焦点
- 👎 差评主要问题
- 💬 用户真实声音
- 🔧 改进优先级

---

### 场景五：行业趋势研判

**适用情况**：
- 投资决策参考
- 战略规划制定
- 市场进入评估

**示例问题**：
```bash
go run main.go "人工智能行业的发展现状、机遇与挑战分析"
go run main.go "Web3.0技术的市场认知度与商业化前景评估"
go run main.go "新消费品牌的崛起趋势与成功要素分析"
```

**关键输出**：
- 📈 行业发展态势
- 🔥 热点技术方向
- 💰 投资热度分析
- ⚠️ 潜在风险识别
- 🔮 未来趋势预判


---

## ⚙️ 高级配置：优化你的舆情分析

### 配置1：选择合适的模型

不同模型适合不同场景：

```bash
# 🚀 深度分析（重要决策）
export OPENAI_MODEL="gpt-4o"
# 适合：品牌危机、重大政策、战略决策

# ⚡ 快速分析（日常监控）
export OPENAI_MODEL="gpt-4o-mini"
# 适合：常规监测、竞品追踪、舆情日报

# 💰 成本优化（高频使用）
export OPENAI_API_BASE="https://api.deepseek.com/v1"
export OPENAI_MODEL="deepseek-chat"
# 适合：大量分析、成本敏感场景

# 🔒 隐私保护（敏感数据）
export OPENAI_API_BASE="http://localhost:11434/v1"
export OPENAI_MODEL="llama3.1"
# 适合：内部数据、敏感信息、离线分析
```

### 配置2：调整分析深度

默认配置已经很好，如需调整可修改代码：

```go
// query_engine/agent.go

// 快速模式（速度优先）
const maxReflectionIterations = 1  // 减少反思次数
const satisfactionThreshold = 0.7  // 降低满意度要求

// 深度模式（质量优先）
const maxReflectionIterations = 3  // 增加反思次数
const satisfactionThreshold = 0.9  // 提高满意度要求
```

---

## 🔧 技术架构：基于LangGraphGo的多Agent系统

微舆的核心价值不仅在于功能，更在于**基于轻量级LangGraphGo框架，快速实现复杂多Agent系统**的技术理念。

### LangGraphGo核心实现

**main.go** - 使用LangGraphGo编排流程：

```go
package main

import (
	"context"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/query_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/media_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/insight_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/forum_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/report_engine"
)

func main() {
	// 初始化状态
	initialState := schema.NewBettaFishState(query)

	// ✨ 使用LangGraphGo创建工作流图
	workflow := graph.NewStateGraph()

	// 添加5大引擎节点
	workflow.AddNode("query_engine", "Query analysis engine", query_engine.QueryEngineNode)
	workflow.AddNode("media_engine", "Media search engine", media_engine.MediaEngineNode)
	workflow.AddNode("insight_engine", "Insight generation engine", insight_engine.InsightEngineNode)
	workflow.AddNode("forum_engine", "Forum search engine", forum_engine.ForumEngineNode)
	workflow.AddNode("report_engine", "Report generation engine", report_engine.ReportEngineNode)

	// 定义执行流程
	workflow.SetEntryPoint("query_engine")
	workflow.AddEdge("query_engine", "media_engine")
	workflow.AddEdge("media_engine", "insight_engine")
	workflow.AddEdge("insight_engine", "forum_engine")
	workflow.AddEdge("forum_engine", "report_engine")
	workflow.AddEdge("report_engine", graph.END)

	// ✨ 编译成可执行应用
	app, err := workflow.Compile()
	if err != nil {
		log.Fatalf("编译图失败: %v", err)
	}

	// ✨ 运行整个流程
	result, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("运行图失败: %v", err)
	}

	finalState := result.(*schema.BettaFishState)
	fmt.Printf("报告已生成，包含 %d 个段落。\n", len(finalState.Paragraphs))
}
```

**关键点**：
- ✅ 使用 `graph.NewStateGraph()` 创建工作流
- ✅ 用 `AddNode()` 注册5大引擎
- ✅ 用 `AddEdge()` 定义执行顺序
- ✅ 用 `Compile()` 编译成可执行应用
- ✅ 用 `Invoke()` 一键运行整个流程

**LangGraphGo的价值**：
```
使用LangGraphGo框架：
- 简洁的图结构定义（~50行）
- 框架自动处理任务调度
- 内置状态管理和错误处理
- 专注于业务逻辑实现
```

---

### 架构设计

```
用户问题
  ↓
┌─────────────────────────────────────┐
│  State Management（状态管理）         │
│  - BettaFishState 统一状态           │
│  - 五大引擎共享数据                    │
└─────────────────────────────────────┘
  ↓
┌─────────────────────────────────────┐
│  QueryEngine（信息收集层）            │
│  - 多轮反思搜索                       │
│  - 质量自我评估                       │
│  - 信息深度挖掘                       │
└─────────────────────────────────────┘
  ↓
┌──────────────┬──────────────────────┐
│ MediaEngine  │  InsightEngine       │
│ （可视化层） │  （数据分析层）          │
│ - 图片搜索   │  - 数据挖掘             │
│ - 内容筛选   │  - 趋势预测             │
└──────────────┴──────────────────────┘
  ↓
┌─────────────────────────────────────┐
│  ForumEngine（多视角整合层）           │
│  - 虚拟圆桌会议                       │
│  - 多Agent协商                       │
│  - 观点冲突解决                       │
└─────────────────────────────────────┘
  ↓
┌─────────────────────────────────────┐
│  ReportEngine（报告生成层）           │
│  - 结构化输出                         │
│  - 决策建议                          │
│  - Markdown渲染                      │
└─────────────────────────────────────┘
  ↓
舆情分析报告
```

### 核心特性

**1. 状态共享机制** (schema/state.go:54-73)
```go
type BettaFishState struct {
    Query string                // 原始问题
    ReportTitle string          // 报告标题
    Paragraphs  []*Paragraph    // 报告段落
    NewsResults []string        // QueryEngine输出
    FinalReport string          // 最终报告
    MediaResults []string       // MediaEngine输出
    InsightResults []string     // InsightEngine输出
    Discussion []string         // ForumEngine输出
}
```

**2. 反思循环机制** (query_engine/agent.go:213-266)
```go
// 为每个段落进行多轮反思优化
maxReflections := 1
for i := 0; i < maxReflections; i++ {
    // 生成反思查询
    var reflectionOutput struct {
        SearchQuery string
        SearchTool  string
        Reasoning   string
    }
    generateJSON(ctx, llm, SystemPromptReflection, input, &reflectionOutput)

    // 执行新搜索
    newResults := ExecuteSearch(ctx, reflectionOutput.SearchQuery, ...)

    // 更新段落总结
    var reflectionSummaryOutput struct {
        UpdatedParagraphLatestState string
    }
    generateJSON(ctx, llm, SystemPromptReflectionSummary, input, &reflectionSummaryOutput)
    p.Research.LatestSummary = reflectionSummaryOutput.UpdatedParagraphLatestState
}
```

**3. 多轮讨论机制** (forum_engine/agent.go:103-113)
```go
// ForumEngine的虚拟圆桌会议：5轮多Agent讨论
turns := []struct {
    Speaker string
    Prompt  string
}{
    {"Moderator", SystemPromptModerator},     // 第1轮：主持人开场
    {"QueryAgent", SystemPromptNewsAgent},    // 第2轮：新闻视角
    {"MediaAgent", SystemPromptMediaAgent},   // 第3轮：媒体视角
    {"QueryAgent", SystemPromptNewsAgent},    // 第4轮：新闻补充
    {"Moderator", SystemPromptModerator},     // 第5轮：主持人总结
}

// 每轮讨论都会记录历史，后续Agent可以看到之前的发言
for i, turn := range turns {
    historyStr := strings.Join(history, "\n\n")
    // Agent基于历史讨论进行发言...
    history = append(history, entry)
}
```

**4. 并行段落处理** (query_engine/agent.go:110-118)
```go
// 使用goroutine并行处理多个段落
var wg sync.WaitGroup
for i := range s.Paragraphs {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        processParagraph(ctx, llm, s.Paragraphs[idx])
    }(i)
}
wg.Wait()
```

### LangGraphGo的设计哲学

**为什么选择轻量级的LangGraphGo框架？**

#### LangGraphGo的核心特点：

| 特点             | 说明                           |
| ---------------- | ------------------------------ |
| ✅ **轻量级设计** | 只提供必要的编排能力，保持简洁 |
| ✅ **透明可控**   | 代码清晰，逻辑明确             |
| ✅ **灵活定制**   | 业务逻辑完全自主实现           |
| ✅ **易于理解**   | API简洁直观                    |
| ✅ **Go原生**     | 充分利用Go语言特性             |

#### 代码示例：

**使用LangGraphGo构建工作流**：
```go
// 清晰的图结构
workflow := graph.NewStateGraph()

// 简单的节点定义
workflow.AddNode("query_engine", "查询引擎", QueryEngineNode)
workflow.AddNode("analysis_engine", "分析引擎", AnalysisEngineNode)

// 直观的流程定义
workflow.AddEdge("query_engine", "analysis_engine")
workflow.AddEdge("analysis_engine", graph.END)

// 编译运行
app, _ := workflow.Compile()
result, _ := app.Invoke(ctx, state)

// 每个节点的逻辑由开发者完全控制
```

#### 微舆的技术架构：

```
LangGraphGo（编排层）
    ↓ 提供图结构和状态管理
业务引擎（实现层）
    ↓ 完全自主实现业务逻辑
    ├─ QueryEngine: 自主设计反思循环
    ├─ MediaEngine: 自主实现图片筛选
    ├─ InsightEngine: 自主开发数据分析
    ├─ ForumEngine: 自主设计圆桌会议
    └─ ReportEngine: 自主实现报告生成
```

**架构优势**：
- ✅ 使用LangGraphGo提供的编排能力（避免重复造轮子）
- ✅ 业务逻辑完全自主实现（保持灵活性和可控性）
- ✅ 清晰的分层架构（职责明确，易于维护）

这种架构设计使得微舆在保持**功能完整**的同时，实现了**清晰的代码结构**和**良好的可维护性**。

---

## ❓ 常见问题

### Q1: 为什么叫"微舆"？

**A**: "微舆"有两层含义：
1. **谐音"微鱼"**，对应英文名BettaFish（斗鱼）
2. **微观舆情**，从细微之处洞察舆情全貌

就像斗鱼虽小但战斗力强，微舆体积轻量但能力强大。

---

### Q2: 微舆与其他舆情分析工具的区别？

**A**: 核心区别在于**从0实现**的技术路线：

| 对比项   | 传统舆情工具 | 微舆                      |
| -------- | ------------ | ------------------------- |
| 技术栈   | 依赖商业框架 | 从0实现，基于LangGraphGo  |
| 分析深度 | 单次搜索     | 反思循环，多轮优化        |
| 视角覆盖 | 单一数据源   | 多Agent多视角整合         |
| 趋势预测 | 简单统计     | 数据驱动的深度洞察        |
| 定制能力 | 受限于框架   | 完全可控可定制            |
| 成本     | 昂贵         | API按需付费，可用本地模型 |

---

### Q3: 需要什么技术背景才能使用？

**A**: **零门槛！**

- **普通用户**：会打命令就行，`go run main.go "你的问题"`
- **开发者**：想定制的话，Go语言基础即可
- **研究者**：代码完全开源，可深入研究多Agent架构


---

## 🌟 总结

微舆（BettaFish）是一个基于LangGraphGo框架实现的多Agent舆情分析系统，致敬并学习了Python原版的优秀设计。

**核心价值**：
- 🌐 打破信息茧房，全网多维度信息收集
- 🎯 还原舆情原貌，多视角客观分析
- 🔮 预测未来走向，数据驱动趋势预判
- 💡 辅助科学决策，提供可执行建议

**技术特点**：
- 五大智能引擎协同工作（QueryEngine、MediaEngine、InsightEngine、ForumEngine、ReportEngine）
- 基于LangGraphGo轻量级框架实现
- Go语言天然并发优势
- 单文件部署，跨平台支持

---

## 📚 相关资源

### 官方资源
- **原始Python版BettaFish**: https://github.com/666ghj/BettaFish
- **LangGraphGo框架**: https://github.com/smallnest/langgraphgo
- **LangChainGo**: https://github.com/tmc/langchaingo

### 技术文档
- **Tavily搜索API**: https://www.tavily.com/
- **OpenAI API**: https://platform.openai.com/docs

### 社区支持
- **GitHub Issues**: https://github.com/smallnest/langgraphgo/issues
- **讨论区**: https://github.com/smallnest/langgraphgo/discussions

---

## ⭐ 支持微舆

如果微舆帮到了你：
- ⭐ 在GitHub上点个Star
- 📢 分享给你的朋友和同事
- 🐛 发现Bug？提Issue
- 💡 有想法？提PR

---

**🐠 Made with ❤️ by BettaFish Team， Rebuilt with 🐕 LangGraphGo**

