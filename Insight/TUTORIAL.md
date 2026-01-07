好的，收到！我要把这份硬核的 **Insight 技术文档**，改写成一篇生动有趣、略带幽默感的“AI研究团队”组建教程。准备好，我们要让这群字节跳动开源的小特工们活起来！ 🚀

-----

## 🤖 恭喜你！你刚刚雇佣了一支“闪电研究队”！

你手上的这个东西，叫 **Insight**，是字节跳动在 **2025年5月10日** 开源的秘密武器。别看它名字像个安静的动物（鹿流？），它实际上是一个由一堆 **“超级智能体（Agent）”** 组成的高效研究公司。

它的任务很简单粗暴：**把你的一个宏大课题，从零信息状态，一路飙到一份精美的报告或播客音频！** 整个过程，你只管翘着脚喝咖啡。

-----

## 💡 认识一下你的“梦幻研究团队”

Insight 强大的秘诀在于 **LangGraph**（以及我们的 Go 语言复刻版 **langgraphgo**！）。它就像一条装配线，让每个智能体各司其职，完美协作。现在，隆重介绍一下你的新员工：

| 员工代号                        | 职能（就像人类）        | 口头禅/工作方式                                            |
| :------------------------------ | :---------------------- | :--------------------------------------------------------- |
| **项目规划师** (Planner)        | **项目经理**            | “别慌，再复杂的任务，我也能给你切成三明治！” 🥪             |
| **首席研究员** (Researcher)     | **资深信息探子**        | “上穷碧落，下黄泉，我要用搜索引擎榨干所有信息！” 🕵️‍♂️         |
| **首席编码员** (Coder) (待上岗) | **数据科学家/分析师**   | “数据？代码跑一下，图表自然就出来了。” 💻                   |
| **报告主编** (Reporter)         | **高级内容撰写师**      | “把所有碎片信息给我，我给你变出个毕业论文级别的报告！” 🎩   |
| **播客制作人** (Podcast)        | **幽默段子手/音频策划** | “严肃的东西太无聊？来，我们找两个主持人给你聊成脱口秀！” 🎙️ |

-----

## 🚀 Insight 的核心超能力

这套框架不仅能帮你完成研究，它还有一些“炫技”功能，让你看起来像个真正的科技大佬：

1.  **无限换“脑”**：它不挑食！无论是字节自己的 **豆包 1.5 Pro**，还是 **OpenAI API**，甚至各种开源模型，它都能通过 **litellm** 库无缝对接。想换哪个“大脑”就换哪个！
2.  **内容全家桶**：
      * **图文报告 (HTML)**：必须的，格式精美。
      * **PPT演示文稿**：可以直接出稿（虽然我们的复刻版说这个功能“效果不漂亮”暂且放着了 😂）。
      * **双人播客音频**：把报告内容秒变轻松对谈。想象一下，两个 AI 播客主持人用火山引擎的甜美声音给你解读《量子计算的最新进展》！
3.  **人工踢馆 & 回放模式**：
      * **“人工介入”**：就像在玩游戏时按下了 **暂停键**。你可以随时插话：“不对，那个研究方向先放放，改成这个！”
      * **“Replay 模式”**：如果你想知道那个研究员智能体是怎么做出“找到这个古怪链接”这个决定的，可以打开回放，像看电影一样回溯它的每一次思考和决策！

-----

## 🛠️ 10 分钟挑战：用 Go 语言“复刻”一个鹿流！

你可能会想：“这么强大的AI团队，不得花我半个月工资请来？” **错！**

借助 **langgraphgo**（[http://lango.rpcx.io](http://lango.rpcx.io)），一个 Go 语言版的 LangGraph 框架，我们的目标是：

> **一个人 + 一杯咖啡 + 半小时 = 一个 Insight 克隆体！**
> (注意：你甚至可能不需要写代码，让 AI 编程工具代劳即可 😉)

### 简化后的复刻功能列表 🔪

为了在半小时内搞定，我们暂时让这支队伍稍稍精简了一下（但核心能力还在！）：

  * **保留** ✅：规划器、研究员、报告员、多模型兼容、HTML 报告、播客**脚本**（音频功能留给有志网友）。
  * **暂时去除** ❌：人工介入（用户不常用，砍！）、编码员（稍后实现）、PPT生成（效果不好看，砍！）。

最终的效果：[https://insight.rpcx.io](https://insight.rpcx.io)

### 架构：Go 语言的流水线作业 ⚙️

我们的 Go 语言版 `Insight` 的核心，是一个由 **状态** 驱动的 LangGraph 图结构。

整个过程的“秘密文件袋”就是这个 **State 结构**：它记录着从用户请求、研究计划、到中间的研究结果，以及最终的报告，所有智能体都在这个袋子里读写信息。

```go
// State：我们的共享“秘密文件袋”
type State struct {
    Plan            []string // 规划师填入的研究步骤
    ResearchResults []string // 研究员填入的发现
    FinalReport     string   // 报告主编填入的 HTML 报告
    GeneratePodcast bool     // 规划师决定是否需要播客
    // ... 其他字段
}
```

-----

### 💻 核心节点（Agent）是怎么工作的？

#### 1\. 规划器 (Planner)

它会拿着你的查询，像个 CEO 一样要求 LLM 返回一个 **JSON 格式** 的详细工作计划。

**核心代码**：
```go
func PlannerNode(ctx context.Context, state any) (any, error) {
    s := state.(*State)
    llm, _ := getLLM()

    // 构建提示词，要求返回 JSON 格式的计划
    prompt := fmt.Sprintf(`你是一名研究规划师。请为以下查询创建一个分步研究计划：%s。
请以 JSON 格式返回结果，格式如下：
{
    "plan": ["步骤1", "步骤2", ...],
    "generate_podcast": true/false
}`, s.Request.Query)

    // 调用 LLM 生成计划
    completion, _ := llms.GenerateFromSinglePrompt(ctx, llm, prompt)

    // 解析 JSON 结果
    var output struct {
        Plan            []string `json:"plan"`
        GeneratePodcast bool     `json:"generate_podcast"`
    }
    json.Unmarshal([]byte(completion), &output)

    s.Plan = output.Plan
    s.GeneratePodcast = output.GeneratePodcast
    return s, nil
}
```

> **知识点**：如果 LLM 返回了不是 JSON 的东西，我们的代码会启用 **"Plan B：猜谜模式"**，尝试按行解析并猜出用户是否想要"播客"，展现了 Go 语言工程师的朴实无华和鲁棒性。

#### 2\. 研究员 (Researcher)

它会根据规划器的计划，**遍历** 每一个步骤，并对每个步骤单独向 LLM 提问，把结果堆在 `ResearchResults` 数组里。

**核心代码**：
```go
func ResearcherNode(ctx context.Context, state any) (any, error) {
    s := state.(*State)
    llm, _ := getLLM()

    var results []string
    // 遍历计划中的每个步骤
    for _, step := range s.Plan {
        // 为每个步骤构建研究提示词
        prompt := fmt.Sprintf("你是一名研究员。请为这个研究步骤查找详细信息：%s。提供发现摘要。", step)

        // 调用 LLM 研究该步骤
        completion, _ := llms.GenerateFromSinglePrompt(ctx, llm, prompt)

        // 将研究结果添加到数组
        results = append(results, fmt.Sprintf("Step: %s\nFindings: %s", step, completion))
    }

    s.ResearchResults = results
    return s, nil
}
```

#### 3\. 报告主编 (Reporter)

它把所有的研究结果（`ResearchResults`）打个包，甩给 LLM，要求返回一个 **Markdown** 报告。

**核心代码**：
```go
func ReporterNode(ctx context.Context, state any) (any, error) {
    s := state.(*State)
    llm, _ := getLLM()

    // 合并所有研究结果
    researchData := strings.Join(s.ResearchResults, "\n\n")

    // 构建报告生成提示词
    prompt := fmt.Sprintf(`你是一名资深报告撰写员。请根据以下研究结果撰写一份全面的最终报告。
使用 Markdown 格式，包含清晰的标题、要点。
研究结果：
%s

原始查询：%s`, researchData, s.Request.Query)

    // 调用 LLM 生成 Markdown 报告
    completion, _ := llms.GenerateFromSinglePrompt(ctx, llm, prompt)

    // 处理图片占位符 [IMAGE_X:Title] 替换为真实的 <img> 标签
    imgRe := regexp.MustCompile(`\[IMAGE_(\d+)[：:]([^\]]+)\]`)
    completion = imgRe.ReplaceAllStringFunc(completion, func(match string) string {
        parts := imgRe.FindStringSubmatch(match)
        idx, _ := strconv.Atoi(parts[1])
        title := parts[2]
        imgURL := s.Images[idx-1]
        return fmt.Sprintf(`<img src="%s" alt="%s" />`, imgURL, title)
    })

    // 将 Markdown 转换为 HTML
    s.FinalReport = markdownToHTML(completion)
    return s, nil
}
```

> **黑科技**：它还负责 **"图片占位符魔法"**！如果报告里有 `[IMAGE_X:Title]` 这种占位符，它会用正则表达式把它替换成真正的 HTML `<img>` 标签！

#### 4\. 播客制作人 (Podcast)

它拿着报告，要求 LLM 生成一个 **双人主持**（Host 1 和 Host 2）的对话脚本，并直接渲染成聊天气泡样式的 **HTML/CSS**。

**核心代码**：
```go
func PodcastNode(ctx context.Context, state any) (any, error) {
    s := state.(*State)
    llm, _ := getLLM()

    researchData := strings.Join(s.ResearchResults, "\n\n")

    // 构建播客脚本生成提示词
    prompt := fmt.Sprintf(`你是一名专业的播客制作人。请根据以下研究结果，创作一段引人入胜的播客对话脚本。
对话应该由两名主持人（Host 1 和 Host 2）进行，风格轻松幽默，通俗易懂。
请以 JSON 格式返回结果，格式如下：
{
    "title": "播客标题",
    "lines": [
        {"speaker": "Host 1", "content": "对话内容..."},
        {"speaker": "Host 2", "content": "对话内容..."}
    ]
}
研究结果：%s`, researchData)

    // 调用 LLM 生成播客脚本
    completion, _ := llms.GenerateFromSinglePrompt(ctx, llm, prompt)

    // 解析 JSON 脚本
    var script struct {
        Title string `json:"title"`
        Lines []struct {
            Speaker string `json:"speaker"`
            Content string `json:"content"`
        } `json:"lines"`
    }
    json.Unmarshal([]byte(completion), &script)

    // 渲染成 HTML 聊天气泡（Host 1 蓝色，Host 2 粉色）
    var html strings.Builder
    for _, line := range script.Lines {
        bgColor := "#e6f7ff"  // 蓝色
        if strings.Contains(line.Speaker, "2") {
            bgColor = "#fff0f6"  // 粉色
        }
        html.WriteString(fmt.Sprintf(`<div style="background:%s">%s: %s</div>`,
            bgColor, line.Speaker, line.Content))
    }

    s.PodcastScript = html.String()
    return s, nil
}
```

> **炫技点**：Host 1 是冷静的**蓝色气泡**，Host 2 是活泼的**粉色气泡**，让报告秒变聊天记录！

-----

## 🏃‍♀️ 部署与体验：Web 模式，才是王道！

为了让你有最佳的体验，我们建议你使用 **Web 服务器模式**：

1.  **编译 Go 程序**：
    ```bash
    export OPENAI_API_KEY="你的_OpenAI_或_DeepSeek_密钥"
    go build -o Insight .
    ```
2.  **启动服务器**：
    ```bash
    ./Insight
    ```
3.  **打开浏览器**：
    ```
    http://localhost:8085
    ```

**Web 模式的优势：**

  * **实时进度条**：使用 **SSE（Server-Sent Events）** 技术，你可以看到智能体们在屏幕上 **实时打字、实时思考** 的全过程，就像看一部紧张刺激的纪录片。
  * **深色主题**：保护你的眼睛，让你看起来更像深夜工作的黑客。
  * **结果缓存**：**“回放模式”启动！** 如果你查过“量子计算”，下次再查时，它会瞬间从本地文件读取结果，并以 **200ms 的间隔** 模拟推送给你。**秒开！省 Token！**

现在，你已经掌握了这支超级研究团队的全部秘密，快去试试让你的“Insight”跑起来吧！ **想让我直接帮你生成一个关于“人工智能如何改变烹饪”的研究报告吗？**