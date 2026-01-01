package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/smallnest/langgraphgo/showcases/health_insights_agent/agents"
	"github.com/smallnest/langgraphgo/showcases/health_insights_agent/cli"
	"github.com/smallnest/langgraphgo/showcases/health_insights_agent/config"
	"github.com/smallnest/langgraphgo/showcases/health_insights_agent/tools"
)

func main() {
	// 解析命令行参数
	cliConfig := cli.ParseFlags()

	// 处理help命令
	if cliConfig.Command == "help" {
		cli.PrintHelp()
		return
	}

	// 加载应用配置
	appConfig := config.DefaultConfig()

	// 从CLI参数覆盖配置
	if cliConfig.Model != "" {
		appConfig.LLMModel = cliConfig.Model
	}
	appConfig.Verbose = cliConfig.Verbose
	appConfig.LLMTemperature = cliConfig.Temperature
	appConfig.LLMMaxTokens = cliConfig.MaxTokens

	// 验证配置
	if err := appConfig.Validate(); err != nil {
		log.Fatalf("❌ 配置错误: %v\n\n提示：请设置 OPENAI_API_KEY 环境变量\n", err)
	}

	if cliConfig.Verbose {
		fmt.Println("🩺 健康洞察代理 - Health Insights Agent")
		fmt.Printf("📦 版本: %s\n", appConfig.AppVersion)
		fmt.Printf("🤖 模型: %s\n", appConfig.LLMModel)
		fmt.Printf("🌡️  温度: %.2f\n", appConfig.LLMTemperature)
		fmt.Println()
	}

	// 创建报告处理器
	processor := tools.NewReportProcessor(appConfig.MaxPDFSizeMB)

	// 获取报告文本
	var reportText string
	var err error

	switch {
	case cliConfig.UseSample || cliConfig.Command == "sample":
		reportText = tools.SampleReport()
		if cliConfig.Verbose {
			fmt.Println("📄 使用示例报告")
		}

	case cliConfig.InputFile != "":
		reportText, err = processor.ProcessFile(cliConfig.InputFile)
		if err != nil {
			log.Fatalf("❌ 读取文件失败: %v\n", err)
		}
		if cliConfig.Verbose {
			fmt.Printf("📄 已从文件读取报告: %s\n", cliConfig.InputFile)
		}

	case cliConfig.InputText != "":
		reportText = processor.ProcessText(cliConfig.InputText)
		if cliConfig.Verbose {
			fmt.Println("📄 已接收报告文本")
		}

	default:
		fmt.Println("❌ 错误: 请提供报告文本")
		fmt.Println("\n使用以下方式之一：")
		fmt.Println("  -file <文件路径>")
		fmt.Println("  -text \"<报告文本>\"")
		fmt.Println("  -sample (使用示例报告)")
		fmt.Println("\n使用 -cmd help 查看完整帮助")
		os.Exit(1)
	}

	// 验证报告
	if err := processor.ValidateReport(reportText); err != nil {
		log.Fatalf("❌ 报告验证失败: %v\n", err)
	}

	if cliConfig.Verbose {
		fmt.Printf("📊 报告长度: %d 字符\n\n", len(reportText))
	}

	// 创建健康分析代理
	agentConfig := &agents.AgentConfig{
		ModelName:   appConfig.LLMModel,
		Temperature: appConfig.LLMTemperature,
		MaxTokens:   appConfig.LLMMaxTokens,
		Timeout:     30 * time.Second,
	}

	agent, err := agents.NewHealthAnalysisAgent(
		appConfig.LLMAPIKey,
		appConfig.LLMBaseURL,
		agentConfig,
		cliConfig.Verbose,
	)
	if err != nil {
		log.Fatalf("❌ 创建分析代理失败: %v\n", err)
	}

	// 执行分析
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := agent.Analyze(ctx, reportText)
	if err != nil {
		log.Fatalf("❌ 分析失败: %v\n", err)
	}

	// 输出结果
	if cliConfig.OutputFile != "" {
		// 保存到文件
		if err := cli.SaveToFile(result, cliConfig.OutputFile); err != nil {
			log.Fatalf("❌ 保存结果失败: %v\n", err)
		}
	}

	// 格式化输出到控制台
	cli.FormatOutput(result, cliConfig.Verbose)

	fmt.Println("\n✅ 分析完成！")
	if !cliConfig.Verbose {
		fmt.Println("\n💡 提示: 使用 -verbose 选项查看详细日志")
	}
}
