package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// HealthAnalysisAgent 健康分析代理
type HealthAnalysisAgent struct {
	model   llms.Model
	config  *AgentConfig
	verbose bool
}

// AgentConfig 代理配置
type AgentConfig struct {
	ModelName   string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration
}

// NewHealthAnalysisAgent 创建新的健康分析代理
func NewHealthAnalysisAgent(apiKey, baseURL string, config *AgentConfig, verbose bool) (*HealthAnalysisAgent, error) {
	opts := []openai.Option{
		openai.WithToken(apiKey),
	}

	if baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}

	if config.ModelName != "" {
		opts = append(opts, openai.WithModel(config.ModelName))
	}

	model, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM model: %w", err)
	}

	return &HealthAnalysisAgent{
		model:   model,
		config:  config,
		verbose: verbose,
	}, nil
}

// CreateAnalysisGraph 创建分析工作流图
func (a *HealthAnalysisAgent) CreateAnalysisGraph() (*graph.StateRunnable[map[string]any], error) {
	workflow := graph.NewStateGraph[map[string]any]()

	// 定义状态schema
	schema := graph.NewMapSchema()
	schema.RegisterReducer("messages", graph.AppendReducer)
	schema.RegisterReducer("report_text", graph.OverwriteReducer)
	schema.RegisterReducer("extracted_data", graph.OverwriteReducer)
	schema.RegisterReducer("analysis", graph.OverwriteReducer)
	schema.RegisterReducer("error", graph.OverwriteReducer)
	workflow.SetSchema(schema)

	// 添加节点：数据提取
	workflow.AddNode("extract_data", "从报告文本中提取结构化数据", a.extractDataNode)

	// 添加节点：分析报告
	workflow.AddNode("analyze_report", "分析血液报告并生成健康洞察", a.analyzeReportNode)

	// 添加节点：完成
	workflow.AddNode("finish", "完成分析", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		if a.verbose {
			fmt.Println("✅ 分析完成")
		}
		return state, nil
	})

	// 定义边
	workflow.SetEntryPoint("extract_data")
	workflow.AddEdge("extract_data", "analyze_report")
	workflow.AddEdge("analyze_report", "finish")
	workflow.AddEdge("finish", graph.END)

	return workflow.Compile()
}

// extractDataNode 数据提取节点
func (a *HealthAnalysisAgent) extractDataNode(ctx context.Context, state map[string]any) (map[string]any, error) {
	reportText, ok := state["report_text"].(string)
	if !ok || reportText == "" {
		return map[string]any{
			"error": "报告文本为空",
		}, fmt.Errorf("empty report text")
	}

	if a.verbose {
		fmt.Println("📊 正在提取血液参数...")
	}

	// 构建提取提示词
	extractPrompt := buildExtractionPrompt(reportText)

	// 调用LLM提取数据
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart("你是一位专业的医疗数据提取专家。")},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(extractPrompt)},
		},
	}

	resp, err := a.model.GenerateContent(ctx, messages,
		llms.WithTemperature(0.1), // 使用较低温度确保准确性
		llms.WithMaxTokens(2000),
	)
	if err != nil {
		return map[string]any{
			"error": fmt.Sprintf("数据提取失败: %v", err),
		}, err
	}

	extractedText := resp.Choices[0].Content
	if a.verbose {
		fmt.Printf("📋 提取结果: %s\n", truncateString(extractedText, 200))
	}

	// 解析JSON
	var extracted map[string]any
	if err := json.Unmarshal([]byte(extractJSON(extractedText)), &extracted); err != nil {
		// 如果解析失败，使用原始文本
		extracted = map[string]any{
			"raw_text": extractedText,
		}
	}

	return map[string]any{
		"extracted_data": extracted,
		"messages":       []string{"数据提取完成"},
	}, nil
}

// analyzeReportNode 分析报告节点
func (a *HealthAnalysisAgent) analyzeReportNode(ctx context.Context, state map[string]any) (map[string]any, error) {
	reportText := state["report_text"].(string)
	extractedData, _ := state["extracted_data"].(map[string]any)

	if a.verbose {
		fmt.Println("🔍 正在进行健康分析...")
	}

	// 构建分析提示词
	analysisPrompt := buildAnalysisPrompt(reportText, extractedData)

	// 调用LLM进行分析
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(getSystemPrompt())},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(analysisPrompt)},
		},
	}

	resp, err := a.model.GenerateContent(ctx, messages,
		llms.WithTemperature(a.config.Temperature),
		llms.WithMaxTokens(a.config.MaxTokens),
	)
	if err != nil {
		return map[string]any{
			"error": fmt.Sprintf("分析失败: %v", err),
		}, err
	}

	analysisText := resp.Choices[0].Content
	if a.verbose {
		fmt.Printf("💡 分析生成完成，长度: %d 字符\n", len(analysisText))
	}

	// 解析分析结果
	analysis, err := parseAnalysisResult(analysisText)
	if err != nil {
		// 如果解析失败，返回原始文本
		analysis = map[string]any{
			"raw_analysis": analysisText,
			"disclaimer":   "此分析由AI生成，不应被视为专业医疗建议的替代品。请咨询医疗保健提供者以获得适当的医疗诊断和治疗。",
		}
	}

	return map[string]any{
		"analysis": analysis,
		"messages": []string{"健康分析完成"},
	}, nil
}

// Analyze 执行完整的分析流程
func (a *HealthAnalysisAgent) Analyze(ctx context.Context, reportText string) (map[string]any, error) {
	startTime := time.Now()

	if a.verbose {
		fmt.Println("\n🩺 === 开始健康分析 ===")
		fmt.Printf("📄 报告长度: %d 字符\n", len(reportText))
	}

	// 创建分析图
	analysisGraph, err := a.CreateAnalysisGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis graph: %w", err)
	}

	// 初始状态
	initialState := map[string]any{
		"report_text":    reportText,
		"extracted_data": nil,
		"analysis":       nil,
		"messages":       []string{},
	}

	// 执行分析
	result, err := analysisGraph.Invoke(ctx, initialState)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	processingTime := time.Since(startTime)

	if a.verbose {
		fmt.Printf("\n⏱️  处理时间: %v\n", processingTime)
		fmt.Println("=== 分析完成 ===\n")
	}

	result["processing_time_ms"] = processingTime.Milliseconds()

	return result, nil
}

// Helper functions

func buildExtractionPrompt(reportText string) string {
	return fmt.Sprintf(`请从以下血液报告文本中提取所有血液参数及其值。

请提取以下信息：
1. 参数名称（如：血红蛋白、白细胞计数、ALT等）
2. 数值
3. 单位（如果有）
4. 标志（如果有：L表示低于正常范围，H表示高于正常范围，N表示正常）

输出格式为JSON：
{
  "parameters": [
    {
      "name": "参数名称",
      "value": "数值",
      "unit": "单位",
      "flag": "L/H/N"
    }
  ],
  "report_date": "报告日期（如果有）",
  "patient_info": {
    "age": "年龄（如果有）",
    "gender": "性别（如果有）"
  }
}

报告文本：
%s`, reportText)
}

func buildAnalysisPrompt(reportText string, extractedData map[string]any) string {
	var dataStr string
	if extractedData != nil {
		dataBytes, _ := json.MarshalIndent(extractedData, "", "  ")
		dataStr = string(dataBytes)
	}

	return fmt.Sprintf(`血液报告原文：
%s

提取的结构化数据：
%s

请基于以上信息，提供一份全面的健康分析。

%s`, reportText, dataStr, getAnalysisFormat())
}

func getSystemPrompt() string {
	return `你是一位经验丰富的医疗分析专家，拥有实验室医学、血液学和内科学的综合知识。
你的任务是分析血液报告并提供详细的健康洞察，包括潜在风险、详细发现和可操作的建议。
请保持专业、准确，并使用通俗易懂的语言解释医学术语。`
}

func getAnalysisFormat() string {
	return `请以JSON格式输出分析结果，包含以下字段：
{
  "disclaimer": "免责声明文本",
  "potential_risks": [
    {
      "condition": "疾病名称",
      "risk_level": "Low/Medium/High",
      "supporting_evidence": ["支持证据1", "支持证据2"],
      "description": "风险描述",
      "severity": 5
    }
  ],
  "recommendations": [
    {
      "category": "Lifestyle/Diet/Medical/Followup",
      "title": "建议标题",
      "description": "详细描述",
      "priority": "Low/Medium/High/Urgent",
      "actionable": true
    }
  ],
  "detailed_findings": [
    {
      "parameter": "参数名称",
      "value": "值",
      "normal_range": "正常范围",
      "status": "Normal/Low/High/Critical",
      "interpretation": "解释",
      "clinical_significance": "临床意义"
    }
  ],
  "overall_assessment": "总体评估文本",
  "confidence": 0.85
}

请确保输出是有效的JSON格式。`
}

func parseAnalysisResult(text string) (map[string]any, error) {
	jsonStr := extractJSON(text)
	var result map[string]any
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

func extractJSON(text string) string {
	// 尝试找到JSON代码块
	start := strings.Index(text, "```json")
	if start != -1 {
		start += 7
		end := strings.Index(text[start:], "```")
		if end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	// 尝试找到普通代码块
	start = strings.Index(text, "```")
	if start != -1 {
		start += 3
		end := strings.Index(text[start:], "```")
		if end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	// 尝试找到JSON对象
	start = strings.Index(text, "{")
	if start != -1 {
		end := strings.LastIndex(text, "}")
		if end != -1 && end > start {
			return strings.TrimSpace(text[start : end+1])
		}
	}

	return text
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
