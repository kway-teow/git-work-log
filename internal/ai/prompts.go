package ai

// PromptType 表示不同类型的提示词
type PromptType string

const (
	// BasicPrompt 基础提示词：核心摘要
	BasicPrompt PromptType = "basic"
	// DetailedPrompt 中级提示词：详细且结构化的报告
	DetailedPrompt PromptType = "detailed"
	// TargetedPrompt 高级提示词：面向角色和受众的报告
	TargetedPrompt PromptType = "targeted"
)

// GetPromptTypeFromString 根据字符串返回对应的提示词类型
func GetPromptTypeFromString(promptTypeStr string) PromptType {
	switch promptTypeStr {
	case string(BasicPrompt):
		return BasicPrompt
	case string(DetailedPrompt):
		return DetailedPrompt
	case string(TargetedPrompt):
		return TargetedPrompt
	default:
		return BasicPrompt // 默认返回基础提示词
	}
}
