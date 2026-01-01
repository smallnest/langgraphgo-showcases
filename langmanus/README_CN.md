# LangManus - Go 实现

使用 [langgraphgo](https://github.com/smallnest/langgraphgo) 和 [langchaingo](https://github.com/tmc/langchaingo) 实现的 LangManus 多智能体 AI 自动化框架的 Go 版本。

## 概述

LangManus 是一个社区驱动的 AI 自动化框架，通过多智能体架构整合语言模型和专业工具来执行复杂任务。这个 Go 实现完美复刻了原 Python 版本。

## 架构

LangManus 采用**分层多智能体架构**，包含以下角色：

- **协调器 (Coordinator)**: 入口点，分析初始请求并路由到合适的智能体
- **规划者 (Planner)**: 分析复杂任务并创建执行策略
- **主管 (Supervisor)**: 统筹工作智能体并监控任务完成情况
- **研究员 (Researcher)**: 使用网络搜索进行信息收集和数据分析
- **编码员 (Coder)**: 处理代码生成、修改和执行 (Python/Bash)
- **浏览器 (Browser)**: 执行网页交互和信息检索
- **报告员 (Reporter)**: 生成综合性最终报告和摘要

## 特性

- ✅ **多智能体编排**: 专业智能体之间的协调工作流
- ✅ **LLM 集成**: 支持 OpenAI 和兼容 API，具有多个模型层级
- ✅ **网络搜索**: 集成 Tavily API 提供研究能力
- ✅ **代码执行**: 安全的 Python 和 Bash 脚本执行
- ✅ **任务规划**: 自动分解复杂任务
- ✅ **流式支持**: 执行过程中的实时更新
- ✅ **可配置**: 基于环境变量的配置

## 安装

```bash
cd showcases/langmanus
go build
```

## 配置

LangManus 会自动从当前目录的 `.env` 文件加载配置。

### 步骤 1: 创建 `.env` 文件

```bash
# 复制示例配置
cp .env.example .env

# 编辑配置
vim .env
```

### 步骤 2: 配置你的 `.env`

```bash
# 必需
OPENAI_API_KEY=your-api-key-here

# 可选（以下是默认值）
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o
OPENAI_MODEL_SMALL=gpt-4o-mini
TEMPERATURE=0.7

# 搜索配置（推荐）
SEARCH_API_KEY=your-tavily-api-key
SEARCH_ENGINE=tavily

# 代码执行
ENABLE_CODE_EXECUTION=true
CODE_TIMEOUT=60

# 智能体配置
MAX_ITERATIONS=15
VERBOSE=true
MAX_CONCURRENT_TASKS=3
```

**注意**：`.env` 文件会在启动时自动加载。如果需要，仍然可以用环境变量覆盖配置。

## 使用方法

### 基本使用

```bash
# 使用默认查询
./langmanus

# 自定义查询
./langmanus "研究 2024 年机器学习趋势并创建摘要"

# 带代码执行的复杂任务
./langmanus "分析 HuggingFace 数据集并编写 Python 代码来可视化结果"
```

### 编程方式使用

```go
package main

import (
    "context"
    "log"
)

func main() {
    // 创建配置
    config := NewConfig()

    // 创建 LangManus 实例
    lm, err := NewLangManus(config)
    if err != nil {
        log.Fatal(err)
    }

    // 运行查询
    ctx := context.Background()
    state, err := lm.Run(ctx, "你的查询")
    if err != nil {
        log.Fatal(err)
    }

    // 访问结果
    println(state.FinalReport)
}
```

### 流式模式

```go
// 流式更新
stateChan, err := lm.Stream(ctx, "你的查询")
if err != nil {
    log.Fatal(err)
}

for state := range stateChan {
    fmt.Printf("智能体: %s\n", state.CurrentAgent)
    fmt.Println(state.Summary())
}
```

## 工作流程

1. **协调器** 接收查询并分析任务类型
2. **规划者** (如需要) 将复杂任务分解为步骤
3. **主管** 将任务分配给专业工作者：
   - **研究员** 负责信息收集
   - **编码员** 负责代码生成/执行
   - **浏览器** 负责网页交互
4. **报告员** 将所有结果综合成最终报告

```
┌─────────────┐
│   协调器    │
└──────┬──────┘
       │
       ├─────────────┐
       │             │
┌──────▼──────┐ ┌───▼────────┐
│   规划者    │ │  研究员    │
└──────┬──────┘ └────────────┘
       │
┌──────▼──────┐
│    主管     │
└──────┬──────┘
       │
       ├──────────┬──────────┐
       │          │          │
┌──────▼──────┐ ┌▼─────┐ ┌──▼────┐
│   研究员    │ │编码员│ │浏览器 │
└──────┬──────┘ └┬─────┘ └──┬────┘
       │         │          │
       └─────────┴──────────┘
                 │
          ┌──────▼──────┐
          │   报告员    │
          └─────────────┘
```

## 示例

### 示例 1: 研究任务

```bash
./langmanus "量子计算领域的最新发展有哪些？"
```

这将：
1. 路由到研究员智能体
2. 进行网络搜索
3. 分析并综合结果
4. 生成综合报告

### 示例 2: 代码分析

```bash
./langmanus "编写 Python 代码来分析 CSV 文件并创建可视化"
```

这将：
1. 路由到规划者
2. 创建带步骤的计划
3. 编码员编写 Python 代码
4. 安全执行代码
5. 报告结果

### 示例 3: 复杂多步骤任务

```bash
./langmanus "研究 Python 测试框架，编写示例代码，并创建对比报告"
```

这将：
1. 规划者创建执行策略
2. 研究员收集测试框架信息
3. 编码员编写示例代码
4. 报告员将发现和代码综合成最终报告

## 组件

### 状态管理

`State` 结构跟踪：
- 查询和消息
- 任务规划和执行
- 智能体路由历史
- 研究和代码结果
- 最终报告

### 工具

- **SearchTool**: 通过 Tavily API 进行网络搜索
- **CodeExecutor**: 带超时的安全 Python/Bash 执行
- **ToolRegistry**: 所有工具的中央注册表

### 智能体

每个智能体实现包含：
- 特定的提示词模板
- LLM 集成
- 状态转换逻辑
- 路由决策

## 与原版对比

| 功能 | 原版 (Python) | 本版 (Go) | 状态 |
|------|---------------|----------|------|
| 多智能体架构 | ✅ | ✅ | 完成 |
| LLM 集成 | ✅ | ✅ | 完成 |
| 网络搜索 (Tavily) | ✅ | ✅ | 完成 |
| 代码执行 | ✅ | ✅ | 完成 |
| 图编排 | LangGraph | langgraphgo | 完成 |
| 流式处理 | ✅ | ✅ | 完成 |
| FastAPI 服务器 | ✅ | ❌ | 未实现 |
| 浏览器自动化 | ✅ | ⚠️ | 部分实现 |

## 系统要求

- Go 1.25+
- Python 3.x (用于代码执行)
- OpenAI API 密钥或兼容端点
- Tavily API 密钥 (用于搜索功能)

## 许可证

MIT 许可证 - 与原 LangManus 项目相同

## 致谢

这是受 Darwin-lfl 的原版 [LangManus](https://github.com/Darwin-lfl/langmanus) 启发的 Go 实现。

使用以下技术构建：
- [langgraphgo](https://github.com/smallnest/langgraphgo) - 多智能体工作流编排
- [langchaingo](https://github.com/tmc/langchaingo) - LLM 集成

## 贡献

欢迎贡献！本项目遵循与原 LangManus 相同的理念 - 回馈开源社区。

## 故障排除

### 搜索不工作
- 确保设置了 `SEARCH_API_KEY`
- 检查 Tavily API 配额

### 代码执行失败
- 确保安装了 Python 3 并在 PATH 中
- 检查 `CODE_TIMEOUT` 设置
- 验证 `ENABLE_CODE_EXECUTION=true`

### LLM 错误
- 验证 `OPENAI_API_KEY` 有效
- 如使用自定义端点，检查 `OPENAI_BASE_URL`
- 确保模型名称正确

## 路线图

- [ ] 添加更多搜索引擎 (Serp, Jina)
- [ ] 实现 FastAPI 服务器模式
- [ ] 添加浏览器自动化
- [ ] 增强错误处理
- [ ] 可视化工具
- [ ] 持久化和检查点
- [ ] 多语言代码执行
