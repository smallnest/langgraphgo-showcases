# LangManus 快速开始指南

## 🚀 快速开始

### 1. 设置 API 密钥

```bash
# 必需：OpenAI API 密钥
export OPENAI_API_KEY="your-api-key-here"

# 可选但推荐：Tavily 搜索 API
export SEARCH_API_KEY="your-tavily-api-key"
```

获取 Tavily API 密钥（免费）：https://tavily.com/

### 2. 运行示例

```bash
# 使用默认查询
./langmanus

# 自定义查询
./langmanus "研究 2025 年 AI 发展趋势"

# 使用测试脚本
./run_example.sh "你的查询"
```

## 🔍 理解工作流

LangManus 使用多智能体工作流：

```
用户查询
    ↓
Coordinator（分析任务）
    ↓
Planner（创建执行计划）
    ↓
Supervisor（分配任务）
    ↓
Researcher/Coder/Browser（执行任务）
    ↓
Reporter（生成最终报告）
```

## 📊 输出说明

### 正常输出

```
=== COORDINATOR Agent Executing ===
Calling LLM (small)...
LLM Response: ...
NEXT_AGENT: planner

=== PLANNER Agent Executing ===
Calling LLM (small)...
Created 2 tasks

=== SUPERVISOR Agent Executing ===
Assigning task to researcher

=== RESEARCHER Agent Executing ===
Search query: 2025 machine learning trends
✓ Research completed: 5 sources found

=== REPORTER Agent Executing ===
Generating final report...
```

### 问题诊断

#### 问题 1: 没有研究结果（Research Results: 0）

**原因**: `SEARCH_API_KEY` 未设置

**解决方案**:
```bash
export SEARCH_API_KEY="your-tavily-api-key"
```

或者系统会显示警告：
```
⚠️  WARNING: SEARCH_API_KEY not set, skipping web search
```

#### 问题 2: 没有创建任务（Tasks: 0）

**原因**: Coordinator 跳过了 Planner，直接路由到其他智能体

**解决方案**:
- 现在已修复，Coordinator 会优先路由到 Planner
- 确保查询包含研究类关键词："研究"、"调查"、"分析"等

#### 问题 3: 报告内容太简单

**原因**:
1. 没有搜索结果（API 密钥问题）
2. LLM 响应格式不符合预期

**解决方案**:
- 设置 `SEARCH_API_KEY` 启用搜索
- 检查 LLM 响应（verbose 模式会显示）
- 确保使用支持的模型

## 🛠️ 配置选项

### 基础配置

```bash
# LLM 配置
export OPENAI_API_KEY="sk-..."
export OPENAI_BASE_URL="https://qianfan.baidubce.com/v2"  # 或其他兼容端点
export OPENAI_MODEL="deepseek-v3"                          # 主模型
export OPENAI_MODEL_SMALL="deepseek-v3"                    # 小模型（节省成本）

# 搜索配置
export SEARCH_API_KEY="tvly-..."
export SEARCH_ENGINE="tavily"

# 代码执行
export ENABLE_CODE_EXECUTION="true"
export CODE_TIMEOUT="60"

# 调试
export VERBOSE="true"
```

### 高级配置

```bash
# 最大迭代次数
export MAX_ITERATIONS="15"

# 并发任务数
export MAX_CONCURRENT_TASKS="3"

# 温度参数
export TEMPERATURE="0.7"
```

## 📋 示例查询

### 研究类任务
```bash
./langmanus "研究量子计算的最新进展"
./langmanus "分析 2025 年 AI 安全趋势"
./langmanus "调查大语言模型的应用场景"
```

### 代码类任务
```bash
./langmanus "写一个 Python 脚本分析 CSV 文件"
./langmanus "创建一个数据可视化示例"
```

### 综合任务
```bash
./langmanus "研究 Python 测试框架，编写示例代码，并创建对比报告"
./langmanus "分析机器学习算法，用 Python 实现并测试"
```

## 🐛 调试技巧

### 启用详细输出

```bash
export VERBOSE=true
./langmanus "你的查询"
```

详细输出会显示：
- 每个智能体的执行
- LLM 调用和响应
- 搜索查询和结果
- 任务创建和状态

### 查看完整日志

默认情况下，LLM 响应会被截断到 500 字符。要查看完整响应，可以修改 `agents.go` 中的 `truncate` 函数。

### 检查路由决策

注意输出中的 `NEXT_AGENT:` 行，这显示了智能体路由决策：

```
LLM Response:
ANALYSIS: This is a research task requiring information gathering
NEXT_AGENT: planner
REASON: Need to create a structured research plan
```

## 💡 最佳实践

### 1. 明确的查询

**好的查询**:
- ✅ "研究 2025 年 AI 趋势并创建摘要报告"
- ✅ "分析量子计算应用，重点关注密码学领域"

**不好的查询**:
- ❌ "AI"
- ❌ "告诉我一些东西"

### 2. 配置搜索 API

虽然搜索 API 是可选的，但强烈推荐设置，否则：
- Researcher 无法获取最新信息
- 报告质量会大幅下降

### 3. 选择合适的模型

- **主模型** (`OPENAI_MODEL`): 用于复杂任务（研究、代码、报告）
  - 推荐: GPT-4, DeepSeek-V3, Claude

- **小模型** (`OPENAI_MODEL_SMALL`): 用于简单任务（路由、解析）
  - 推荐: GPT-4o-mini, DeepSeek-V3（便宜）

## 🔒 安全注意事项

### 代码执行

默认启用代码执行。代码会在临时文件中执行，具有以下限制：

- 60 秒超时（可配置）
- 在当前用户权限下运行
- 执行 Python 3 和 Bash

**生产环境建议**:
```bash
# 禁用代码执行
export ENABLE_CODE_EXECUTION="false"
```

或使用容器/沙箱环境。

## 📞 获取帮助

遇到问题？

1. 检查环境变量是否正确设置
2. 启用 `VERBOSE=true` 查看详细日志
3. 查看 README.md 和 README_CN.md
4. 提交 Issue 到 GitHub

## 🎯 下一步

- 尝试不同类型的查询
- 探索代码执行功能
- 集成到你的应用中
- 自定义提示词模板
- 添加新的工具和智能体

祝你使用愉快！🚀
