package report_engine

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
)

const reportTemplate = `
# {{.Title}}

**生成时间**: {{.Date}}

## 摘要
{{.Summary | unescape}}

## 详细分析
{{range .Paragraphs}}
### {{.Title}}
{{.Content | unescape}}
{{end}}

{{if .ShowMediaFindings}}
## 视觉和上下文信息 (MediaEngine)
{{range .MediaFindings}}
- {{.}}
{{end}}
{{end}}

{{if .ShowDiscussion}}
## 专家讨论 (ForumEngine)
{{range .Discussion}}
> {{.}}
{{end}}
{{end}}

---
*本报告由 DeepInsight 多智能体系统自动生成。*
`

type ReportData struct {
	Title            string
	Date             string
	Summary          string
	Paragraphs       []ParagraphData
	MediaFindings    []string
	ShowMediaFindings bool
	Discussion       []string
	ShowDiscussion   bool
}

type ParagraphData struct {
	Title   string
	Content string
}

// unescape converts escape sequences like \n to actual newlines
func unescape(s string) string {
	// Replace common escape sequences
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

func GenerateReport(state *schema.DeepInsightState, skipIncomplete bool) (string, error) {
	// Create template with custom function map
	funcMap := template.FuncMap{
		"unescape": unescape,
	}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportTemplate)
	if err != nil {
		return "", err
	}

	// Extract summary from the first part of ResearchResults if available, or just use a placeholder
	summary := "暂无摘要"
	if len(state.ResearchResults) > 0 {
		// In the current implementation, ResearchResults[0] is the full report from QueryEngine.
		summary = state.ResearchResults[0]
	}

	// Construct ParagraphData from state
	var paragraphs []ParagraphData
	for _, p := range state.Paragraphs {
		content := p.Research.LatestSummary
		// Skip incomplete paragraphs if requested
		if content == "" {
			if skipIncomplete {
				continue
			}
			content = "该部分未完成研究。"
		}
		paragraphs = append(paragraphs, ParagraphData{
			Title:   p.Title,
			Content: content,
		})
	}

	data := ReportData{
		Title:            fmt.Sprintf("深度洞察报告: %s", state.Query),
		Date:             time.Now().Format("2006-01-02 15:04:05"),
		Summary:          summary,
		Paragraphs:       paragraphs,
		MediaFindings:    state.MediaResults,
		ShowMediaFindings: !state.SimpleMode, // Hide in simple mode
		Discussion:       state.Discussion,
		ShowDiscussion:   !state.SimpleMode, // Hide in simple mode
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
