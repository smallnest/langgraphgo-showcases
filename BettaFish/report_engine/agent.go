package report_engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
)

// ReportEngineNode generates the final report file.
func ReportEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.BettaFishState)
	fmt.Println("ReportEngine: 正在生成最终报告...")

	// Use the new template-based generation
	reportContent, err := GenerateReport(s)
	if err != nil {
		fmt.Printf("ReportEngine: 生成报告内容失败: %v\n", err)
		// Fallback to simple concatenation if template fails
		reportContent = "生成报告失败，请检查日志。"
	}

	s.FinalReport = reportContent

	// Save to file
	filename := fmt.Sprintf("report_%s_%s.md", strings.ReplaceAll(s.Query, " ", "_"), time.Now().Format("20060102_150405"))
	err = os.WriteFile(filename, []byte(s.FinalReport), 0600)
	if err != nil {
		fmt.Printf("ReportEngine: 保存报告失败: %v\n", err)
	} else {
		fmt.Printf("ReportEngine: 报告已保存至 %s\n", filename)
	}

	return s, nil
}
