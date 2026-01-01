package config

import "errors"

var (
	// 配置相关错误
	ErrMissingAPIKey = errors.New("缺少LLM API密钥")
	ErrInvalidConfig = errors.New("无效的配置")

	// 分析相关错误
	ErrEmptyReport         = errors.New("血液报告内容为空")
	ErrAnalysisFailed      = errors.New("分析失败")
	ErrExtractionFailed    = errors.New("数据提取失败")
	ErrInvalidReportFormat = errors.New("无效的报告格式")

	// PDF处理错误
	ErrPDFTooLarge    = errors.New("PDF文件过大")
	ErrPDFReadFailed  = errors.New("PDF读取失败")
	ErrPDFParseFailed = errors.New("PDF解析失败")

	// 模型相关错误
	ErrModelNotAvailable = errors.New("模型不可用")
	ErrAllModelsFailed   = errors.New("所有模型都失败了")
	ErrInvalidResponse   = errors.New("无效的模型响应")
	ErrRateLimitExceeded = errors.New("超出速率限制")

	// 用户相关错误
	ErrUserNotFound       = errors.New("用户未找到")
	ErrInvalidUserContext = errors.New("无效的用户背景信息")
	ErrSessionExpired     = errors.New("会话已过期")

	// 通用错误
	ErrInternal     = errors.New("内部错误")
	ErrTimeout      = errors.New("操作超时")
	ErrInvalidInput = errors.New("无效的输入")
)
