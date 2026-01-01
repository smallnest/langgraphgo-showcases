package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// CLIConfig CLI配置
type CLIConfig struct {
	Command     string
	InputFile   string
	InputText   string
	OutputFile  string
	Verbose     bool
	DetailLevel string
	UseSample   bool
	Model       string
	Temperature float64
	MaxTokens   int
}

// ParseFlags 解析命令行参数
func ParseFlags() *CLIConfig {
	config := &CLIConfig{}

	flag.StringVar(&config.Command, "cmd", "analyze", "命令：analyze(分析报告), sample(使用示例报告), help(帮助)")
	flag.StringVar(&config.InputFile, "file", "", "输入文件路径（文本文件）")
	flag.StringVar(&config.InputText, "text", "", "直接输入报告文本")
	flag.StringVar(&config.OutputFile, "output", "", "输出文件路径（JSON格式）")
	flag.BoolVar(&config.Verbose, "verbose", false, "显示详细日志")
	flag.StringVar(&config.DetailLevel, "detail", "Standard", "详细程度：Basic, Standard, Comprehensive")
	flag.BoolVar(&config.UseSample, "sample", false, "使用示例报告")
	flag.StringVar(&config.Model, "model", "", "指定LLM模型（覆盖环境变量）")
	flag.Float64Var(&config.Temperature, "temperature", 0.3, "模型温度参数（0.0-2.0）")
	flag.IntVar(&config.MaxTokens, "max-tokens", 4000, "最大token数")

	flag.Parse()

	return config
}

// PrintHelp 打印帮助信息
func PrintHelp() {
	fmt.Println(`
健康洞察代理 - 血液报告AI分析工具

用法:
  health-insights-agent [选项]

命令:
  -cmd analyze    分析血液报告（默认）
  -cmd sample     使用示例报告进行分析
  -cmd help       显示帮助信息

选项:
  -file <路径>           输入文件路径（支持 .txt 和 .pdf 格式）
  -text "<文本>"         直接输入报告文本
  -output <路径>         输出文件路径（JSON格式），不指定则输出到控制台
  -verbose              显示详细日志
  -detail <级别>         详细程度：Basic, Standard, Comprehensive（默认：Standard）
  -sample               使用内置示例报告
  -model <模型>         指定LLM模型（如：gpt-4, gpt-3.5-turbo）
  -temperature <值>     模型温度参数 0.0-2.0（默认：0.3）
  -max-tokens <数量>    最大token数（默认：4000）

环境变量:
  OPENAI_API_KEY       OpenAI API密钥（必需）
  OPENAI_API_BASE      OpenAI API基础URL（可选）
  LLM_MODEL            默认使用的模型（可选）

支持的文件格式:
  .txt, .text          纯文本文件
  .pdf                 PDF文档（会自动提取文本）

示例:
  # 使用示例报告进行分析
  health-insights-agent -sample -verbose

  # 分析文本文件
  health-insights-agent -file report.txt -verbose

  # 分析PDF文件
  health-insights-agent -file report.pdf -verbose

  # 直接输入文本分析
  health-insights-agent -text "血液报告内容..." -verbose

  # 分析并保存结果到文件
  health-insights-agent -file report.pdf -output result.json

  # 使用特定模型和参数
  health-insights-agent -sample -model gpt-4 -temperature 0.5

注意事项:
  - PDF文件需要是可提取文本的格式（非扫描版）
  - 如果是扫描版PDF，请先使用OCR工具转换为文本
  - PDF文件最大支持 20MB（可通过环境变量 MAX_PDF_SIZE_MB 修改）

详细文档请查看 README_CN.md
`)
}

// FormatOutput 格式化输出
func FormatOutput(result map[string]any, verbose bool) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🩺 健康洞察分析报告")
	fmt.Println(strings.Repeat("=", 80))

	// 提取分析结果
	analysis, ok := result["analysis"].(map[string]any)
	if !ok {
		fmt.Println("\n❌ 无法解析分析结果")
		if verbose {
			fmt.Printf("\n原始结果：\n%+v\n", result)
		}
		return
	}

	// 免责声明
	if disclaimer, ok := analysis["disclaimer"].(string); ok {
		fmt.Printf("\n⚠️  免责声明\n%s\n", disclaimer)
	}

	// 总体评估
	if assessment, ok := analysis["overall_assessment"].(string); ok {
		fmt.Printf("\n📊 总体评估\n%s\n", assessment)
	}

	// 置信度
	if confidence, ok := analysis["confidence"].(float64); ok {
		fmt.Printf("\n🎯 分析置信度: %.1f%%\n", confidence*100)
	}

	// 潜在风险
	if risks, ok := analysis["potential_risks"].([]any); ok && len(risks) > 0 {
		fmt.Println("\n⚠️  潜在健康风险")
		fmt.Println(strings.Repeat("-", 80))
		for i, risk := range risks {
			if riskMap, ok := risk.(map[string]any); ok {
				fmt.Printf("\n%d. %s", i+1, riskMap["condition"])
				if level, ok := riskMap["risk_level"].(string); ok {
					fmt.Printf(" [风险等级: %s]", getRiskLevelEmoji(level))
				}
				if desc, ok := riskMap["description"].(string); ok {
					fmt.Printf("\n   %s", desc)
				}
				if evidence, ok := riskMap["supporting_evidence"].([]any); ok && len(evidence) > 0 {
					fmt.Print("\n   支持证据: ")
					for j, ev := range evidence {
						if j > 0 {
							fmt.Print(", ")
						}
						fmt.Print(ev)
					}
				}
				fmt.Println()
			}
		}
	}

	// 详细发现
	if findings, ok := analysis["detailed_findings"].([]any); ok && len(findings) > 0 {
		fmt.Println("\n🔬 详细检查发现")
		fmt.Println(strings.Repeat("-", 80))
		for i, finding := range findings {
			if findingMap, ok := finding.(map[string]any); ok {
				param := findingMap["parameter"]
				value := findingMap["value"]
				normalRange := findingMap["normal_range"]
				status := findingMap["status"]

				fmt.Printf("\n%d. %s: %s [正常范围: %s] %s",
					i+1, param, value, normalRange, getStatusEmoji(fmt.Sprintf("%v", status)))

				if interpretation, ok := findingMap["interpretation"].(string); ok {
					fmt.Printf("\n   解释: %s", interpretation)
				}
				fmt.Println()
			}
		}
	}

	// 建议
	if recommendations, ok := analysis["recommendations"].([]any); ok && len(recommendations) > 0 {
		fmt.Println("\n💡 健康建议")
		fmt.Println(strings.Repeat("-", 80))

		categories := map[string][]any{
			"Lifestyle": {},
			"Diet":      {},
			"Medical":   {},
			"Followup":  {},
		}

		for _, rec := range recommendations {
			if recMap, ok := rec.(map[string]any); ok {
				if cat, ok := recMap["category"].(string); ok {
					categories[cat] = append(categories[cat], rec)
				}
			}
		}

		for _, cat := range []string{"Lifestyle", "Diet", "Medical", "Followup"} {
			recs := categories[cat]
			if len(recs) > 0 {
				fmt.Printf("\n%s：\n", getCategoryName(cat))
				for i, rec := range recs {
					if recMap, ok := rec.(map[string]any); ok {
						title := recMap["title"]
						desc := recMap["description"]
						priority := recMap["priority"]

						fmt.Printf("%d. [%s] %s\n", i+1, getPriorityEmoji(fmt.Sprintf("%v", priority)), title)
						fmt.Printf("   %s\n", desc)
					}
				}
			}
		}
	}

	// 处理时间
	if processingTime, ok := result["processing_time_ms"].(int64); ok {
		fmt.Printf("\n⏱️  处理时间: %dms\n", processingTime)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// SaveToFile 保存结果到文件
func SaveToFile(result map[string]any, filePath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Printf("\n✅ 结果已保存到: %s\n", filePath)
	return nil
}

// Helper functions

func getRiskLevelEmoji(level string) string {
	switch level {
	case "Low":
		return "🟢 低"
	case "Medium":
		return "🟡 中"
	case "High":
		return "🔴 高"
	default:
		return level
	}
}

func getStatusEmoji(status string) string {
	switch status {
	case "Normal":
		return "✅"
	case "Low":
		return "⬇️"
	case "High":
		return "⬆️"
	case "Critical":
		return "🚨"
	default:
		return ""
	}
}

func getPriorityEmoji(priority string) string {
	switch priority {
	case "Low":
		return "🔵"
	case "Medium":
		return "🟡"
	case "High":
		return "🟠"
	case "Urgent":
		return "🔴"
	default:
		return "⚪"
	}
}

func getCategoryName(category string) string {
	names := map[string]string{
		"Lifestyle": "生活方式调整",
		"Diet":      "饮食建议",
		"Medical":   "医疗建议",
		"Followup":  "后续跟进",
	}
	if name, ok := names[category]; ok {
		return name
	}
	return category
}
