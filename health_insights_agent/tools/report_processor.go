package tools

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ReportProcessor 报告处理器
type ReportProcessor struct {
	maxSizeMB int
}

// NewReportProcessor 创建新的报告处理器
func NewReportProcessor(maxSizeMB int) *ReportProcessor {
	return &ReportProcessor{
		maxSizeMB: maxSizeMB,
	}
}

// ProcessFile 处理文件（自动检测类型）
func (p *ReportProcessor) ProcessFile(filePath string) (string, error) {
	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("无法读取文件信息: %w", err)
	}

	sizeMB := float64(fileInfo.Size()) / 1024 / 1024
	if sizeMB > float64(p.maxSizeMB) {
		return "", fmt.Errorf("文件过大: %.2fMB (最大允许 %dMB)", sizeMB, p.maxSizeMB)
	}

	// 根据文件扩展名判断类型
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return p.ProcessPDFFile(filePath)
	case ".txt", ".text":
		return p.ProcessTextFile(filePath)
	default:
		// 默认按文本文件处理
		return p.ProcessTextFile(filePath)
	}
}

// ProcessTextFile 处理文本文件
func (p *ReportProcessor) ProcessTextFile(filePath string) (string, error) {
	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("无法读取文件信息: %w", err)
	}

	sizeMB := float64(fileInfo.Size()) / 1024 / 1024
	if sizeMB > float64(p.maxSizeMB) {
		return "", fmt.Errorf("文件过大: %.2fMB (最大允许 %dMB)", sizeMB, p.maxSizeMB)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	text := string(content)

	// 清理文本
	text = p.CleanText(text)

	return text, nil
}

// ProcessPDFFile 处理PDF文件
func (p *ReportProcessor) ProcessPDFFile(filePath string) (string, error) {
	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("无法读取PDF文件信息: %w", err)
	}

	sizeMB := float64(fileInfo.Size()) / 1024 / 1024
	if sizeMB > float64(p.maxSizeMB) {
		return "", fmt.Errorf("PDF文件过大: %.2fMB (最大允许 %dMB)", sizeMB, p.maxSizeMB)
	}

	// 打开PDF文件
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开PDF文件失败: %w", err)
	}
	defer f.Close()

	// 提取所有页面的文本
	var buf bytes.Buffer
	totalPages := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		// 提取页面文本
		text, err := p.GetPlainText(nil)
		if err != nil {
			// 如果提取失败，尝试继续处理其他页面
			continue
		}

		buf.WriteString(text)
		buf.WriteString("\n")
	}

	extractedText := buf.String()

	// 清理文本
	extractedText = p.CleanText(extractedText)

	if strings.TrimSpace(extractedText) == "" {
		return "", fmt.Errorf("PDF文件未能提取到有效文本，可能是扫描版PDF")
	}

	return extractedText, nil
}

// ProcessPDFBytes 处理PDF字节流
func (p *ReportProcessor) ProcessPDFBytes(data []byte) (string, error) {
	sizeMB := float64(len(data)) / 1024 / 1024
	if sizeMB > float64(p.maxSizeMB) {
		return "", fmt.Errorf("PDF数据过大: %.2fMB (最大允许 %dMB)", sizeMB, p.maxSizeMB)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "report_*.pdf")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 写入数据
	if _, err := tmpFile.Write(data); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 处理PDF文件
	return p.ProcessPDFFile(tmpFile.Name())
}

// ProcessText 处理原始文本
func (p *ReportProcessor) ProcessText(text string) string {
	return p.CleanText(text)
}

// CleanText 清理文本
func (p *ReportProcessor) CleanText(text string) string {
	// 移除多余的空白
	text = strings.TrimSpace(text)

	// 规范化换行符
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// 移除多余的空行
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	prevEmpty := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !prevEmpty {
				cleanedLines = append(cleanedLines, "")
				prevEmpty = true
			}
		} else {
			cleanedLines = append(cleanedLines, trimmed)
			prevEmpty = false
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// ValidateReport 验证报告内容
func (p *ReportProcessor) ValidateReport(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("报告内容为空")
	}

	// 基本检查：报告应该包含一些常见的医疗术语或参数
	keywords := []string{
		"血", "red", "white", "hemoglobin", "WBC", "RBC",
		"ALT", "AST", "glucose", "cholesterol",
		"报告", "检查", "结果", "分析",
	}

	textLower := strings.ToLower(text)
	hasKeyword := false
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			hasKeyword = true
			break
		}
	}

	if !hasKeyword {
		return fmt.Errorf("报告内容似乎不是有效的医疗报告")
	}

	return nil
}

// ExtractMetadata 提取报告元数据
func (p *ReportProcessor) ExtractMetadata(text string) map[string]any {
	metadata := make(map[string]any)

	lines := strings.Split(text, "\n")
	metadata["line_count"] = len(lines)
	metadata["char_count"] = len(text)
	metadata["word_count"] = len(strings.Fields(text))

	// 尝试提取日期（简单模式匹配）
	for _, line := range lines {
		if strings.Contains(line, "日期") || strings.Contains(line, "Date") || strings.Contains(line, "date") {
			metadata["report_date"] = strings.TrimSpace(line)
			break
		}
	}

	// 尝试提取患者信息
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "姓名") || strings.Contains(lower, "name") {
			metadata["patient_info"] = strings.TrimSpace(line)
			break
		}
	}

	return metadata
}

// FormatReport 格式化报告以便更好地显示
func (p *ReportProcessor) FormatReport(text string) string {
	lines := strings.Split(text, "\n")
	var formatted []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			formatted = append(formatted, "")
			continue
		}

		// 如果是标题行（通常较短且可能包含特定字符）
		if len(trimmed) < 50 && (strings.Contains(trimmed, "报告") ||
			strings.Contains(trimmed, "检查") || strings.Contains(trimmed, "分析")) {
			formatted = append(formatted, "")
			formatted = append(formatted, "## "+trimmed)
			formatted = append(formatted, "")
		} else {
			formatted = append(formatted, trimmed)
		}
	}

	return strings.Join(formatted, "\n")
}

// SampleReport 返回一个示例报告用于测试
func SampleReport() string {
	return `血液检查报告

患者信息：
姓名：张三
性别：男
年龄：45岁
检查日期：2024-12-01

检查项目及结果：

血常规：
- 白细胞计数 (WBC): 7.5 × 10^9/L [正常范围: 4.0-10.0]
- 红细胞计数 (RBC): 4.2 × 10^12/L [正常范围: 4.0-5.5]
- 血红蛋白 (HGB): 125 g/L [正常范围: 120-160] (L)
- 血小板计数 (PLT): 180 × 10^9/L [正常范围: 100-300]

肝功能：
- 丙氨酸氨基转移酶 (ALT): 65 U/L [正常范围: 0-40] (H)
- 天冬氨酸氨基转移酶 (AST): 55 U/L [正常范围: 0-40] (H)
- 总胆红素 (TBIL): 18 μmol/L [正常范围: 5-21]
- 直接胆红素 (DBIL): 6 μmol/L [正常范围: 0-7]

肾功能：
- 尿素氮 (BUN): 5.5 mmol/L [正常范围: 2.9-7.1]
- 肌酐 (CRE): 85 μmol/L [正常范围: 44-133]

血脂：
- 总胆固醇 (TC): 6.2 mmol/L [正常范围: 3.1-5.7] (H)
- 甘油三酯 (TG): 2.5 mmol/L [正常范围: 0.4-1.7] (H)
- 高密度脂蛋白 (HDL-C): 0.9 mmol/L [正常范围: 1.0-1.5] (L)
- 低密度脂蛋白 (LDL-C): 4.1 mmol/L [正常范围: 2.1-3.4] (H)

血糖：
- 空腹血糖 (GLU): 6.5 mmol/L [正常范围: 3.9-6.1] (H)

注：H=高于正常范围，L=低于正常范围

检查医师：李医生
报告日期：2024-12-01`
}
