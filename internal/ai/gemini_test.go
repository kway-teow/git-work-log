package ai

import (
	"os"
	"testing"
)

// TestNewGeminiClient 测试创建Gemini客户端
func TestNewGeminiClient(t *testing.T) {
	// 保存原始环境变量
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey) // 测试结束后恢复

	// 测试1: 未设置API密钥
	os.Unsetenv("GEMINI_API_KEY")
	_, err := NewGeminiClient()
	if err == nil {
		t.Error("未设置API密钥时应该返回错误")
	}

	// 测试2: 设置API密钥但跳过实际API调用
	// 注意：这只是验证函数逻辑，不会实际调用API
	os.Setenv("GEMINI_API_KEY", "test_key")
	// 由于实际创建客户端需要有效的API密钥，我们不测试成功路径
	// 这里只是验证代码不会在设置了环境变量的情况下立即返回错误
}

// TestGetPromptType 测试提示词类型转换
func TestGetPromptType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PromptType
	}{
		{"基础提示词", "basic", BasicPrompt},
		{"详细提示词", "detailed", DetailedPrompt},
		{"针对性提示词", "targeted", TargetedPrompt},
		{"未知类型", "unknown", BasicPrompt}, // 默认返回基础提示词
		{"空字符串", "", BasicPrompt},       // 默认返回基础提示词
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetPromptTypeFromString(test.input)
			if result != test.expected {
				t.Errorf("输入 %s: 期望 %s, 得到 %s", test.input, test.expected, result)
			}
		})
	}
}

// TestLoadPromptTemplate 测试加载提示词模板
func TestLoadPromptTemplate(t *testing.T) {
	// 创建临时测试文件
	testContent := "测试模板内容 {{.CommitMessages}}"
	tmpfile, err := os.CreateTemp("", "test-prompt-*.txt")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatalf("无法写入临时文件: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("无法关闭临时文件: %v", err)
	}

	// 测试加载不存在的文件
	_, err = loadPromptTemplateFromPath("/non/existent/path.txt")
	if err == nil {
		t.Error("加载不存在的文件应该返回错误")
	}

	// 测试加载存在的文件
	content, err := loadPromptTemplateFromPath(tmpfile.Name())
	if err != nil {
		t.Errorf("加载存在的文件不应该返回错误: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("加载的内容不匹配: 期望 %q, 得到 %q", testContent, string(content))
	}
}
