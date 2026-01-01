package graph

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func SearchNode(ctx context.Context, s any) (any, error) {
	state, ok := s.(*State)
	if !ok {
		return nil, fmt.Errorf("无效的状态类型: %T", s)
	}

	logf(state, "正在搜索 %s ...", state.Username)

	// 使用文本输出而不是 JSON
	ctx, cancel := context.WithTimeout(ctx, 8*time.Minute)
	defer cancel()
	// nolint:gosec // G204: Username is validated input for the social-analyzer tool
	cmd := exec.CommandContext(ctx, "social-analyzer", "--username", state.Username, "--top", "500")

	// 获取 stdout 管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return state, fmt.Errorf("获取 stdout 管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return state, fmt.Errorf("启动 social-analyzer 失败: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	var profileBuilder strings.Builder
	isProfileSection := false
	checkCount := 0

	// 用于解析 Result
	var currentResult *Result
	var results []Result

	for scanner.Scan() {
		line := scanner.Text()

		// 1. 处理日志输出
		if strings.HasPrefix(line, "[init]") || strings.HasPrefix(line, "[Init]") || strings.HasPrefix(line, "[Info]") {
			logf(state, "%s", line)
		} else if strings.HasPrefix(line, "[Checking]") {
			checkCount++
			if checkCount%100 == 0 {
				logf(state, "搜索了 %d 个网站", checkCount)
			}
		} else if strings.HasPrefix(line, "[Detected]") {
			logf(state, "%s", line)
			isProfileSection = true
		}

		// 2. 捕获 Profile 数据
		if isProfileSection {
			profileBuilder.WriteString(line + "\n")

			// 尝试解析 Result 用于前端显示 (SocialData)
			if strings.HasPrefix(line, "found") {
				currentResult = &Result{Exists: true}
			} else if strings.HasPrefix(line, "link") && currentResult != nil {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					currentResult.Link = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "title") && currentResult != nil {
				// title 也可以作为 name
				// 简单起见，我们从 link 中提取 name，或者使用 title
				// 这里我们暂时不提取 name，后面统一处理
				// Parts would be split here if needed: parts := strings.SplitN(line, ":", 2)
			} else if strings.HasPrefix(line, "-----------------------") {
				if currentResult != nil && currentResult.Link != "" {
					linkLower := strings.ToLower(currentResult.Link)

					// 过滤不需要的网站
					if strings.Contains(linkLower, "xvideos") {
						currentResult = nil
						continue
					}

					// 过滤 Habr Career (俄罗斯技术招聘网站)
					if strings.Contains(linkLower, "habr") || strings.Contains(linkLower, "career.habr") {
						currentResult = nil
						continue
					}

					// 从 link 中提取 name
					// e.g., https://github.com/smallnest -> github.com
					// 简单处理
					linkParts := strings.Split(currentResult.Link, "/")
					if len(linkParts) > 2 {
						currentResult.Name = linkParts[2]
					} else {
						currentResult.Name = currentResult.Link
					}
					results = append(results, *currentResult)
					currentResult = nil
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		// social-analyzer 可能会返回非零退出码，即使部分成功
		// 我们记录错误但继续，只要我们捕获了一些数据
		logf(state, "social-analyzer 完成 (可能带有警告): %v", err)
	}

	state.ProfileData = profileBuilder.String()
	state.SocialData = results

	logf(state, "搜索完成，捕获了 %d 字节的画像数据", len(state.ProfileData))
	return state, nil
}

func ProfileNode(ctx context.Context, s any) (any, error) {
	state, ok := s.(*State)
	if !ok {
		return nil, fmt.Errorf("无效的状态类型: %T", s)
	}

	logf(state, "正在为 %s 生成画像...", state.Username)
	defer logf(state, "画像生成完成")

	llm, err := openai.New()
	if err != nil {
		return state, fmt.Errorf("创建 LLM 失败: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**任务：社交媒体用户深度画像分析**\n\n目标用户: %s\n\n", state.Username))
	sb.WriteString("请基于提供的社交媒体数据，对目标用户进行全面、深入的画像分析。直接输出分析结果，无需说明数据来源。\n\n")
	sb.WriteString("**分析维度与要求：**\n\n")
	sb.WriteString("## 1. 基础身份信息\n- 用户名/昵称及其含义解读\n- 性别 (男、女或者未知)\n- 地理位置（国家/地区/城市）\n- 语言使用习惯\n- 账号创建时间与活跃年限\n- 时区与在线活动时间特征\n\n")
	sb.WriteString("## 2. 职业与专业领域\n- 职业身份识别（开发者/设计师/创作者/学生等）\n- 专业技能栈与核心能力\n- 代表性作品/项目及其影响力指标\n- 行业地位与影响力评估\n- 职业发展轨迹（如有时间线数据）\n\n")
	sb.WriteString("## 3. 技术特征（针对技术类用户）\n- 主要技术栈与编程语言偏好\n- 开源贡献情况（项目数量、stars、forks）\n- 技术社区活跃度（GitHub/Stack Overflow/技术论坛等）\n- 技术方向演变趋势\n- 学习与分享习惯\n\n")
	sb.WriteString("## 4. 社交行为模式\n- **平台分布**：主要活跃平台类型与使用频率\n- **内容偏好**：发布内容的主题、类型、风格\n- **互动模式**：followers/following比例，互动频率\n- **社交圈层**：专业社交 vs 娱乐社交的比重\n- **活跃时段**：发帖时间、更新频率\n\n")
	sb.WriteString("## 5. 内容创作特征\n- 内容产出类型（代码/文章/视频/图片等）\n- 创作主题与领域聚焦度\n- 更新频率与持续性\n- 内容质量与受欢迎程度\n- 写作/表达风格\n\n")
	sb.WriteString("## 6. 兴趣爱好与生活方式\n- 专业外的兴趣领域\n- 娱乐休闲偏好\n- 学习与自我提升倾向\n- 社区参与情况\n- 消费偏好（如有相关数据）\n\n")
	sb.WriteString("## 7. 价值观与个性特征\n- 个人签名/简介的价值观解读\n- 内容中体现的思维方式\n- 对开源/分享/协作的态度\n- 社交风格（专业型/活跃型/低调型）\n- 表达习惯与沟通特点\n\n")
	sb.WriteString("## 8. 影响力评估\n- **定量指标**：\n  - 粉丝数量与增长趋势\n  - 内容传播数据（stars/likes/retweets）\n  - 项目使用量与下载量\n- **定性评估**：\n  - 行业认可度\n  - 社区口碑\n  - 影响力范围（本地/全国/国际）\n\n")
	sb.WriteString("## 9. 数字足迹特征\n- 平台注册密度与分布特点\n- 账号活跃度差异分析\n- 弃用平台与迁移趋势\n- 隐私保护意识评估\n\n")
	sb.WriteString("## 10. 发展趋势与预测\n- 技能发展方向\n- 兴趣演变轨迹\n- 职业发展可能性\n- 潜在合作机会领域\n\n")
	sb.WriteString("## 11. 异常点与特殊标记\n- 不符合常规模式的数据点\n- 账号冲突或矛盾信息\n- 潜在的隐私或安全风险标记\n- 需要进一步验证的信息\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("一点澄清：Go 不要翻译成国际象棋或者围棋， 因为它可能代表的是 Go 语言\n\n")
	sb.WriteString("当前时间：" + time.Now().Format(time.RFC3339) + "\n\n")
	sb.WriteString("**输出格式要求：**\n1. 使用结构化的标题和子标题\n2.选取可信度最高的avatar加入在报告的顶部, 一定不要采用 Gravatar 的头像图片\n3. 关键数据用**加粗**突出\n4. 重要发现用单独段落强调\n5. 最后提供一段100-150字的**综合画像总结**\n6. 标注**可信度等级**：高/中/低（基于数据完整性）\n\n")
	sb.WriteString("**分析原则：**\n- 客观中立，基于数据事实\n- 区分确定信息与推测信息\n- 注意文化背景差异\n- 保护用户隐私，不做不当推测\n- 发现数据矛盾时明确指出\n\n")
	sb.WriteString("发现的账户信息(Raw Data):\n")

	if state.ProfileData == "" {
		sb.WriteString("未明确找到特定的社交媒体账户（或者它们是私密/受保护的）。\n")
		sb.WriteString("请基于用户名本身和拥有此句柄的用户的共同特征生成一个假设性的画像，但要明确说明这是推测性的。\n")
	} else {
		profileData := state.ProfileData
		if len(profileData) > 90*1024 { // make sure tokens < 90k
			profileData = profileData[:90*1024]
		}
		sb.WriteString(profileData)
	}

	completion, err := llm.Call(ctx, sb.String(), llms.WithTemperature(0.7))
	if err != nil {
		return state, fmt.Errorf("生成画像失败: %w", err)
	}

	state.ProfileText = completion
	return state, nil
}

// AccountNode 使用 Tavily API 搜索用户 ID，并使用 LLM 提取
func AccountNode(ctx context.Context, s any) (any, error) {
	state, ok := s.(*State)
	if !ok {
		return nil, fmt.Errorf("无效的状态类型: %T", s)
	}

	logf(state, "正在为 %s 搜索 User ID...", state.Username)

	// 1. 调用 Tavily API 搜索用户信息
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		logf(state, "警告: TAVILY_API_KEY 未设置，跳过 Tavily 搜索，使用用户名作为 User ID")
		state.UserID = state.Username
		return state, nil
	}

	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s user id social media account", state.Username)

	// 准备 Tavily API 请求
	requestBody := map[string]any{
		"api_key":      apiKey,
		"query":        searchQuery,
		"search_depth": "basic",
		"max_results":  5,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logf(state, "构建 Tavily 请求失败: %v", err)
		state.UserID = state.Username
		return state, nil
	}

	// 发送请求到 Tavily API
	resp, err := http.Post("https://api.tavily.com/search", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logf(state, "Tavily API 请求失败: %v，使用用户名作为 User ID", err)
		state.UserID = state.Username
		return state, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logf(state, "Tavily API 返回错误状态 %d: %s，使用用户名作为 User ID", resp.StatusCode, string(body))
		state.UserID = state.Username
		return state, nil
	}

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logf(state, "读取 Tavily 响应失败: %v，使用用户名作为 User ID", err)
		state.UserID = state.Username
		return state, nil
	}

	// 保存原始搜索结果
	state.TavilySearchData = string(responseBody)
	logf(state, "Tavily 搜索完成，获得 %d 字节数据", len(responseBody))

	// 2. 使用 LLM 从搜索结果中提取 User ID
	llm, err := openai.New()
	if err != nil {
		logf(state, "创建 LLM 失败: %v，使用用户名作为 User ID", err)
		state.UserID = state.Username
		return state, nil
	}

	// 构建提示词
	var promptBuilder strings.Builder
	promptBuilder.WriteString("**任务：从搜索结果中提取准确的 User ID**\n\n")
	promptBuilder.WriteString(fmt.Sprintf("目标用户名: %s\n\n", state.Username))
	promptBuilder.WriteString("请仔细分析以下 Tavily 搜索结果，找出该用户在各个社交媒体平台上的真实 User ID。\n\n")
	promptBuilder.WriteString("**要求：**\n")
	promptBuilder.WriteString("1. User ID 通常是用户在平台上的唯一标识符（不是显示名称）\n")
	promptBuilder.WriteString("2. 优先查找 GitHub、Twitter、LinkedIn 等主流平台的 User ID\n")
	promptBuilder.WriteString("3. 如果找到明确的 User ID，直接返回它（只返回 ID 本身，不要其他解释）\n")
	promptBuilder.WriteString("4. 如果找到多个可能的 ID，返回最可信的一个\n")
	promptBuilder.WriteString(fmt.Sprintf("5. 如果搜索结果中没有找到明确的 User ID，或者不确定，请直接返回: %s\n\n", state.Username))
	promptBuilder.WriteString("**搜索结果数据：**\n")

	// 限制数据大小以避免超出 token 限制
	searchData := state.TavilySearchData
	if len(searchData) > 30*1024 {
		searchData = searchData[:30*1024]
	}
	promptBuilder.WriteString(searchData)
	promptBuilder.WriteString("\n\n请只返回 User ID，不要其他内容。")

	// 调用 LLM
	extractedID, err := llm.Call(ctx, promptBuilder.String(), llms.WithTemperature(0.2))
	if err != nil {
		logf(state, "LLM 提取 User ID 失败: %v，使用用户名作为 User ID", err)
		state.UserID = state.Username
		return state, nil
	}

	// 清理提取的 ID（去除空格和换行）
	extractedID = strings.TrimSpace(extractedID)

	if extractedID == "" {
		logf(state, "LLM 未能提取 User ID，使用用户名作为 User ID")
		state.UserID = state.Username
	} else {
		state.UserID = extractedID
		logf(state, "成功提取 User ID: %s", state.UserID)
	}

	return state, nil
}

func logf(state *State, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
	if state.LogChan != nil {
		state.LogChan <- msg
	}
}
