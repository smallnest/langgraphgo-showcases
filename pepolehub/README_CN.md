# PeopleHub (Go Port)

这是 [PeopleHub](https://github.com/MeirKaD/pepolehub) 调研代理的 Go 语言实现版本，基于 [langgraphgo](https://github.com/smallnest/langgraphgo) 构建。

它通过以下步骤自动化人物调研过程：
1.  获取 LinkedIn 个人资料（通过网络搜索）。
2.  在网络上搜索相关信息（通过 Tavily）。
3.  抓取相关网页内容。
4.  总结网页内容（通过 OpenAI）。
5.  生成一份全面的调研报告。

## 前置要求

你需要以下 API Key：
*   `OPENAI_API_KEY`: 用于生成摘要和报告。
*   `TAVILY_API_KEY`: 用于搜索网络和 LinkedIn 资料。

## 使用方法

设置环境变量并运行：

```bash
export OPENAI_API_KEY="sk-..."
export TAVILY_API_KEY="tvly-..."
go run showcases/pepolehub/*.go -name "John Doe" -linkedin "https://linkedin.com/in/johndoe"
```

## 功能特性

*   **图工作流**: 使用 `langgraphgo` 编排调研步骤。
*   **真实实现**: 使用 Tavily 进行搜索，OpenAI 进行智能处理（无 Mock）。
*   **并行执行**: 并行执行 LinkedIn 数据获取和网络搜索。
*   **条件路由**: 根据搜索结果动态决定是否抓取网页。
*   **状态管理**: 使用 `FieldMerger` 进行健壮的状态管理。

## 架构

代理遵循以下图工作流：

1.  **Start**: 初始化调研。
2.  **并行步骤**:
    *   `FetchLinkedIn`: 搜索个人资料内容。
    *   `ExecuteSearch`: 在网络上搜索该人物。
3.  **抓取与总结**: 如果发现搜索结果，则抓取并总结内容。
4.  **聚合**: 合并 LinkedIn 数据和网页摘要。
5.  **WriteReport**: 生成最终的 Markdown 报告。

## 原项目

*   [MeirKaD/pepolehub](https://github.com/MeirKaD/pepolehub)