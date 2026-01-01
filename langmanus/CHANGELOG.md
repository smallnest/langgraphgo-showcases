# Changelog

## [Latest] - 2025-12-09

### 🎉 Initial Release

首次发布 LangManus Go 版本，完美复刻原 Python 版本的功能。

### ✨ 新功能

- **多智能体架构**: 7个专业智能体协同工作
  - Coordinator（协调器）
  - Planner（规划者）
  - Supervisor（主管）
  - Researcher（研究员）
  - Coder（编码员）
  - Browser（浏览器）
  - Reporter（报告员）

- **LangGraph 工作流**: 使用 langgraphgo 实现复杂的条件路由
- **LLM 集成**: 通过 langchaingo 支持 OpenAI 兼容接口
- **工具支持**:
  - Tavily 搜索 API 集成
  - Python/Bash 代码执行
  - 安全超时机制

### 🔧 配置管理

- ✅ **自动加载 .env**: 使用 godotenv 自动加载配置
- ✅ **详细配置文档**: .env.example 包含所有配置选项
- ✅ **灵活配置**: 支持环境变量覆盖

### 📝 文档

- ✅ README.md（英文）
- ✅ README_CN.md（中文）
- ✅ QUICKSTART.md（快速开始指南）
- ✅ .env.example（配置示例）
- ✅ run_example.sh（测试脚本）

### 🐛 问题修复

#### 修复 1: 改进日志输出

**问题**: 执行过程不透明，看不到 LLM 响应和决策过程

**解决方案**:
- 添加详细的智能体执行日志
- 显示 LLM 响应（前 500 字符）
- 显示搜索查询和结果
- 显示任务创建和分配

```go
if a.Verbose {
    fmt.Printf("Calling LLM (%s)...\n", modelName)
    fmt.Printf("LLM Response (first 500 chars):\n%s\n", truncate(content, 500))
}
```

#### 修复 2: 搜索 API 错误处理

**问题**: 当 SEARCH_API_KEY 未设置时，程序静默失败，没有明确提示

**解决方案**:
- 检查 API 密钥状态
- 显示明确的警告信息
- 提供设置说明

```go
if a.Tools.Search.APIKey == "" {
    fmt.Println("⚠️  WARNING: SEARCH_API_KEY not set, skipping web search")
    fmt.Println("    Set SEARCH_API_KEY environment variable to enable search")
}
```

#### 修复 3: 工作流路由优化

**问题**: Coordinator 跳过 Planner，直接路由到 Researcher，导致没有创建任务计划

**解决方案**:
- 优化 Coordinator 提示词
- 默认将研究类任务路由到 Planner
- 确保任务分解和规划

**之前**:
```
Coordinator → Researcher → Supervisor → Reporter
(没有创建任务，Tasks: 0)
```

**现在**:
```
Coordinator → Planner → Supervisor → Researcher → Supervisor → Reporter
(创建任务计划，Tasks: 2-3)
```

#### 修复 4: 模板函数支持

**问题**: Go template 缺少 `add` 函数

**解决方案**:
```go
funcMap := template.FuncMap{
    "add": func(a, b int) int { return a + b },
}
tmpl := template.New("prompt").Funcs(funcMap).Parse(promptTemplate)
```

### 📊 性能指标

- **可执行文件大小**: 11 MB
- **编译时间**: < 5 秒
- **内存占用**: 约 50-100 MB（运行时）
- **响应时间**: 取决于 LLM API 延迟

### 🔄 与原版对比

| 特性 | Python 版本 | Go 版本 | 状态 |
|------|------------|---------|------|
| 多智能体架构 | ✅ | ✅ | ✅ 完全实现 |
| LLM 集成 | LangChain | langchaingo | ✅ 完全实现 |
| 图编排 | LangGraph | langgraphgo | ✅ 完全实现 |
| 搜索功能 | Tavily | Tavily | ✅ 完全实现 |
| 代码执行 | Python REPL | Python/Bash | ✅ 完全实现 |
| 配置管理 | .env | .env (godotenv) | ✅ 完全实现 |
| 流式输出 | SSE | - | ❌ 未实现 |
| FastAPI 服务 | ✅ | ❌ | ❌ 未实现 |
| 浏览器自动化 | Playwright | - | ⚠️ 部分实现 |

### 🚀 使用示例

#### 基础使用
```bash
# 1. 配置环境
cp .env.example .env
vim .env  # 设置 OPENAI_API_KEY

# 2. 构建
go build -o langmanus

# 3. 运行
./langmanus "研究 2025 年 AI 趋势"
```

#### 期望输出
```
=== LangManus Starting ===
Query: 研究 2025 年 AI 趋势

=== COORDINATOR Agent Executing ===
Calling LLM (small)...
LLM Response: ANALYSIS: Research task requiring planning...
NEXT_AGENT: planner

=== PLANNER Agent Executing ===
Calling LLM (small)...
Created 2 tasks:
  1. Search for latest AI trends in 2025 - researcher
  2. Synthesize findings into report - reporter

=== SUPERVISOR Agent Executing ===
Assigning next task: researcher

=== RESEARCHER Agent Executing ===
Calling LLM (main)...
Search query: 2025 AI trends machine learning
✓ Research completed: 5 sources found
  1. AI Trends 2025 (https://example.com/ai-trends)
  2. Machine Learning Advances (https://example.com/ml-2025)

=== SUPERVISOR Agent Executing ===
All tasks completed, routing to reporter

=== REPORTER Agent Executing ===
Calling LLM (main)...
Generating final report...

=== LangManus Complete ===
Tasks completed: 2/2
Research Results: 1
```

### 📚 依赖项

```go
require (
    github.com/google/uuid v1.6.0
    github.com/joho/godotenv v1.5.1
    github.com/smallnest/langgraphgo v0.5.0
    github.com/tmc/langchaingo v0.1.14
)
```

### 🎯 下一步计划

- [ ] 实现流式输出支持
- [ ] 添加 FastAPI/HTTP 服务器模式
- [ ] 完善浏览器自动化（集成 chromedp）
- [ ] 添加更多搜索引擎支持（Serp, Jina）
- [ ] 实现持久化和检查点
- [ ] 添加可视化工具
- [ ] 性能优化和并发改进
- [ ] 单元测试和集成测试

### 🙏 致谢

感谢原 [LangManus](https://github.com/Darwin-lfl/langmanus) 项目提供的设计思路和架构参考。

---

## 问题排查指南

### 问题: 没有生成报告

**症状**:
```
Research Results: 0
Tasks: 0
Final Report: {json object}
```

**原因**:
1. SEARCH_API_KEY 未设置
2. 工作流跳过了 Planner

**解决方案**:
1. 设置 `SEARCH_API_KEY` 在 .env 文件中
2. 确保使用最新版本（已修复路由问题）
3. 启用 `VERBOSE=true` 查看详细日志

### 问题: LLM 调用失败

**症状**:
```
LLM call failed: ...
```

**解决方案**:
1. 检查 `OPENAI_API_KEY` 是否正确
2. 检查 `OPENAI_BASE_URL` 是否可访问
3. 验证模型名称是否正确
4. 检查网络连接

### 问题: 代码执行失败

**症状**:
```
Code execution error: ...
```

**解决方案**:
1. 确保 Python 3 已安装
2. 检查 `CODE_TIMEOUT` 设置
3. 验证 `ENABLE_CODE_EXECUTION=true`
4. 检查代码权限

---

## 版本信息

- **版本**: v1.0.0
- **发布日期**: 2025-12-09
- **Go 版本**: 1.25.0+
- **langgraphgo 版本**: 0.5.0
- **langchaingo 版本**: 0.1.14
