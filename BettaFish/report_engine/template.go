package report_engine

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
)

const reportTemplate = `
# {{.Title}}

**生成时间**: {{.Date}}

## 摘要
{{.Summary}}

## 详细分析
{{range .Paragraphs}}
### {{.Title}}
{{.Content}}
{{end}}

## 视觉情报 (MediaEngine)
{{range .MediaFindings}}
- {{.}}
{{end}}

## 智能体论坛讨论 (ForumEngine)
{{range .Discussion}}
> {{.}}
{{end}}

---
*本报告由 BettaFish 多智能体系统自动生成。*
`

type ReportData struct {
	Title         string
	Date          string
	Summary       string
	Paragraphs    []ParagraphData
	MediaFindings []string
	Discussion    []string
}

type ParagraphData struct {
	Title   string
	Content string
}

func GenerateReport(state *schema.BettaFishState) (string, error) {
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return "", err
	}

	// Extract summary from the first part of NewsResults if available, or just use a placeholder
	summary := "暂无摘要"
	if len(state.NewsResults) > 0 {
		// In the current implementation, NewsResults[0] is the full report from QueryEngine.
		// We might want to parse it or just use it as the "Summary" if it's short,
		// but QueryEngine actually generates a full report.
		// For this enhanced version, let's assume QueryEngine's output is the "Detailed Analysis"
		// and we might want to split it.
		// However, to keep it simple and robust:
		// We will use the Paragraphs from the state directly for "Detailed Analysis"
		// and use the QueryEngine's "Final Report" (NewsResults[0]) as the "Summary" or "Overview".
		summary = state.NewsResults[0]
	}

	// Construct ParagraphData from state
	var paragraphs []ParagraphData
	for _, p := range state.Paragraphs {
		content := p.Research.LatestSummary
		if content == "" {
			content = "该部分未完成研究。"
		}
		paragraphs = append(paragraphs, ParagraphData{
			Title:   p.Title,
			Content: content,
		})
	}

	data := ReportData{
		Title:         fmt.Sprintf("深度分析报告: %s", state.Query),
		Date:          time.Now().Format("2006-01-02 15:04:05"),
		Summary:       summary, // This might be redundant if we list paragraphs, but let's keep it for now.
		Paragraphs:    paragraphs,
		MediaFindings: state.MediaResults,
		Discussion:    state.Discussion,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
