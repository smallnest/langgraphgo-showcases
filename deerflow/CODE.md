# DeerFlow 代码实现解读

本文档详细解读了 DeerFlow 深度研究智能体的代码实现。该项目使用 Go 语言编写，基于 `langgraphgo` 框架构建多智能体工作流。

## 1. 核心架构与状态管理

DeerFlow 的核心是一个基于状态图（State Graph）的工作流，定义在 `graph.go` 中。

### 1.1 状态定义 (`State`)

整个工作流共享一个单一的状态对象，所有节点都从这里读取数据并将结果写回这里。

```go
// graph.go
type State struct {
    Request         Request  `json:"request"`          // 初始用户请求
    Plan            []string `json:"plan"`             // 规划的研究步骤
    ResearchResults []string `json:"research_results"` // 每个步骤的研究成果
    Images          []string `json:"images"`           // 收集到的图片链接
    FinalReport     string   `json:"final_report"`     // 最终生成的 HTML 报告
    PodcastScript   string   `json:"podcast_script"`   // 播客脚本内容
    GeneratePodcast bool     `json:"generate_podcast"` // 是否生成播客的标志
    Step            int      `json:"step"`             // 当前执行步数（部分逻辑使用）
}
```

### 1.2 图结构 (`NewGraph`)

工作流被定义为一个有向图：

1.  **EntryPoint**: `planner` 节点。
2.  **流程**: `planner` -> `researcher` -> `reporter`。
3.  **条件分支**: `reporter` 完成后，逻辑判断：
    *   如果 `GeneratePodcast` 为 `true` -> 进入 `podcast` 节点。
    *   否则 -> 直接结束 (`END`)。

```go
// graph.go 核心逻辑
workflow.AddNode("planner", ..., PlannerNode)
workflow.AddNode("researcher", ..., ResearcherNode)
workflow.AddNode("reporter", ..., ReporterNode)
workflow.AddNode("podcast", ..., PodcastNode)

// 关键的条件边
workflow.AddConditionalEdge("reporter", func(ctx context.Context, state any) string {
    s := state.(*State)
    if s.GeneratePodcast {
        return "podcast"
    }
    return graph.END
})
```

## 2. 智能体节点实现

所有节点的具体逻辑都在 `nodes.go` 中实现。

### 2.1 规划器 (PlannerNode)

*   **功能**: 接收用户的自然语言查询，分解为结构化的研究步骤。
*   **输入**: `state.Request.Query`
*   **输出**: 更新 `state.Plan` 和 `state.GeneratePodcast`
*   **实现细节**:
    *   使用 LLM 提示词要求返回 JSON 格式的计划。
    *   **鲁棒性处理**: 实现了 JSON 解析失败的 `fallback` 机制。如果 LLM 返回的不是标准 JSON，代码会尝试按行解析文本作为计划步骤，并通过关键词（如 "播客"、"podcast"）手动检测是否需要生成播客。

### 2.2 研究员 (ResearcherNode)

*   **功能**: 执行规划好的每一个步骤。
*   **输入**: `state.Plan`
*   **输出**: 更新 `state.ResearchResults`
*   **实现细节**:
    *   遍历 `Plan` 切片。
    *   对每个步骤单独调用 LLM，提示词为“你是一名研究员...请查找详细信息”。
    *   目前实现使用的是 LLM 的内部知识库进行“模拟”研究，而非真实的实时网络搜索（虽然架构上预留了图片字段 `Images`，但当前节点逻辑中置为了 `nil`）。
    *   结果被追加到 `ResearchResults` 数组中。

### 2.3 报告者 (ReporterNode)

*   **功能**: 汇总研究发现，生成最终的 HTML 报告。
*   **输入**: `state.ResearchResults`
*   **输出**: 更新 `state.FinalReport`
*   **实现细节**:
    *   将所有研究结果拼接成一个长文本。
    *   提示词要求 LLM 生成 Markdown 格式的报告。
    *   **图片处理**: 包含一段特殊的逻辑来处理图片占位符 `[IMAGE_X:Title]`。代码使用正则表达式提取这些占位符，并将其替换为标准的 HTML `<img>` 标签（如果有收集到图片的话）。
    *   **渲染**: 使用 `github.com/gomarkdown/markdown` 库将 LLM 生成的 Markdown 转换为最终的 HTML。

### 2.4 播客制作人 (PodcastNode)

*   **功能**: 将严肃的研究报告转化为轻松的双人对话脚本。
*   **输入**: `state.ResearchResults`
*   **输出**: 更新 `state.PodcastScript`
*   **实现细节**:
    *   提示 LLM 生成两个主持人（Host 1 和 Host 2）的对话，要求返回 JSON 格式。
    *   **UI 渲染**: 代码并没有直接保存 JSON，而是构建了一段包含 HTML/CSS 的字符串。这段 HTML 使用了类似消息气泡的样式（Host 1 为蓝色系，Host 2 为在粉色系），可以直接嵌入在网页中展示。
    *   包含一个“导出 JSON”的按钮功能嵌入在 HTML 中。

## 3. 应用程序入口与交互

`main.go` 提供了两种运行模式。

### 3.1 命令行模式 (CLI)

如果运行 `./deerflow "查询内容"`，程序会：
1.  创建一个初始状态。
2.  直接调用 `graph.Invoke`。
3.  在终端打印最终报告。

### 3.2 Web 服务器模式

如果直接运行 `./deerflow`，程序启动 HTTP 服务器（默认端口 8085）。

*   **静态资源**: 通过 `embed` 包将 `web/` 目录下的 HTML/JS/CSS 文件打包在二进制中，便于分发。
*   **API `/api/run`**: 核心接口，使用 **SSE (Server-Sent Events)** 技术。
    *   **实时反馈**: 利用 Go 的 Channel (`logChan`) 捕获节点执行过程中的 `logf` 输出，并实时推送到前端。
    *   **缓存/回放系统**:
        *   根据查询内容生成唯一的文件名。
        *   检查 `data/[query]/` 目录下是否存在该查询的历史记录。
        *   如果存在，进入**回放模式 (`replayRun`)**：读取历史日志和结果，以 200ms 的间隔模拟推送给前端，实现“秒开”体验且节省 Token。
        *   如果不存在，启动新的 Graph 执行，并在结束后保存元数据、日志和结果到磁盘。
*   **API `/api/history`**: 扫描 `data/` 目录，返回所有历史查询记录，按时间倒序排列。

## 4. 关键工具函数

*   **`logf`**: 一个双向日志函数。它既会打印到标准输出（Stdout），也会尝试非阻塞地发送到 Context 中的 channel。这使得同一个函数可以在 CLI 模式下仅打印日志，而在 Web 模式下同时推送 SSE 更新。
*   **`getLLM`**: 封装了 LLM 的初始化逻辑，目前使用 `langchaingo/llms/openai`，支持通过环境变量配置 API Key 和 Base URL（如使用 DeepSeek 等兼容接口）。

---
*本文档基于当前代码库版本生成，旨在帮助开发者快速理解 DeerFlow 的实现细节。*
