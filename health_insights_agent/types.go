package main

import (
	"time"
)

// AnalysisReport 代表一份完整的健康分析报告
type AnalysisReport struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	ReportText       string          `json:"report_text"`
	Analysis         *HealthAnalysis `json:"analysis"`
	CreatedAt        time.Time       `json:"created_at"`
	Metadata         map[string]any  `json:"metadata"`
	ProcessingTimeMs int64           `json:"processing_time_ms"`
}

// HealthAnalysis 包含AI生成的健康分析结果
type HealthAnalysis struct {
	Disclaimer        string            `json:"disclaimer"`
	PotentialRisks    []HealthRisk      `json:"potential_risks"`
	Recommendations   []Recommendation  `json:"recommendations"`
	DetailedFindings  []DetailedFinding `json:"detailed_findings"`
	OverallAssessment string            `json:"overall_assessment"`
	Confidence        float64           `json:"confidence"` // 0-1之间，表示分析的置信度
}

// HealthRisk 代表一个潜在的健康风险
type HealthRisk struct {
	Condition          string   `json:"condition"`           // 疾病或健康状况名称
	RiskLevel          string   `json:"risk_level"`          // Low, Medium, High
	SupportingEvidence []string `json:"supporting_evidence"` // 支持该风险判断的血液指标
	Description        string   `json:"description"`         // 风险描述
	Severity           int      `json:"severity"`            // 1-10，严重程度
}

// Recommendation 代表一条健康建议
type Recommendation struct {
	Category    string `json:"category"`    // Lifestyle, Diet, Medical, Followup
	Title       string `json:"title"`       // 建议标题
	Description string `json:"description"` // 详细描述
	Priority    string `json:"priority"`    // Low, Medium, High, Urgent
	Actionable  bool   `json:"actionable"`  // 是否可以立即执行
}

// DetailedFinding 代表血液指标的详细发现
type DetailedFinding struct {
	Parameter            string `json:"parameter"`             // 参数名称，如 "Hemoglobin"
	Value                string `json:"value"`                 // 实际值
	NormalRange          string `json:"normal_range"`          // 正常范围
	Status               string `json:"status"`                // Normal, Low, High, Critical
	Interpretation       string `json:"interpretation"`        // 对该指标的解释
	ClinicalSignificance string `json:"clinical_significance"` // 临床意义
}

// AgentState 代表分析代理的状态
type AgentState struct {
	ReportText    string          `json:"report_text"`
	ExtractedData map[string]any  `json:"extracted_data"`
	Analysis      *HealthAnalysis `json:"analysis"`
	UserContext   *UserContext    `json:"user_context"`
	Messages      []string        `json:"messages"`
	CurrentStep   string          `json:"current_step"`
	Error         string          `json:"error,omitempty"`
}

// UserContext 用户背景信息（可选）
type UserContext struct {
	Age                int               `json:"age,omitempty"`
	Gender             string            `json:"gender,omitempty"`
	MedicalHistory     []string          `json:"medical_history,omitempty"`
	CurrentMedications []string          `json:"current_medications,omitempty"`
	Lifestyle          map[string]string `json:"lifestyle,omitempty"`
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	ReportText  string           `json:"report_text"`
	UserContext *UserContext     `json:"user_context,omitempty"`
	Options     *AnalysisOptions `json:"options,omitempty"`
}

// AnalysisOptions 分析选项
type AnalysisOptions struct {
	Verbose        bool     `json:"verbose"`         // 是否输出详细日志
	DetailLevel    string   `json:"detail_level"`    // Basic, Standard, Comprehensive
	FocusAreas     []string `json:"focus_areas"`     // 关注的特定健康领域
	IncludeHistory bool     `json:"include_history"` // 是否包含历史分析对比
	Language       string   `json:"language"`        // 输出语言，默认中文
}

// ModelConfig LLM模型配置
type ModelConfig struct {
	Provider       string   `json:"provider"`        // openai, anthropic, groq等
	PrimaryModel   string   `json:"primary_model"`   // 主模型
	FallbackModels []string `json:"fallback_models"` // 备用模型列表
	Temperature    float64  `json:"temperature"`     // 温度参数
	MaxTokens      int      `json:"max_tokens"`      // 最大token数
	TopP           float64  `json:"top_p"`           // Top-p采样参数
}

// SessionHistory 会话历史记录
type SessionHistory struct {
	UserID       string           `json:"user_id"`
	Sessions     []AnalysisReport `json:"sessions"`
	TotalReports int              `json:"total_reports"`
	LastAccessed time.Time        `json:"last_accessed"`
}

// BloodParameter 血液参数定义
type BloodParameter struct {
	Name                 string   `json:"name"`         // 参数名称
	CommonNames          []string `json:"common_names"` // 常用别名
	Unit                 string   `json:"unit"`         // 单位
	NormalRangeMale      *Range   `json:"normal_range_male,omitempty"`
	NormalRangeFemale    *Range   `json:"normal_range_female,omitempty"`
	Category             string   `json:"category"` // CBC, Liver, Kidney, Lipid等
	ClinicalSignificance string   `json:"clinical_significance"`
}

// Range 数值范围
type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ExtractionResult PDF提取结果
type ExtractionResult struct {
	Text       string               `json:"text"`
	Parameters []ExtractedParameter `json:"parameters"`
	Metadata   map[string]any       `json:"metadata"`
	Success    bool                 `json:"success"`
	Error      string               `json:"error,omitempty"`
}

// ExtractedParameter 提取的参数
type ExtractedParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
	Flag  string `json:"flag,omitempty"` // L(低), H(高), N(正常)
}
