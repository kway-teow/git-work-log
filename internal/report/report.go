package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kway-teow/git-work-log/internal/git"
)

// Format 表示报告输出格式
type Format string

const (
	// FormatText 纯文本格式
	FormatText Format = "text"
	// FormatMarkdown Markdown格式
	FormatMarkdown Format = "markdown"
)

// Generator 报告生成器
type Generator struct {
	Format Format
	Output io.Writer // 输出目标，可以是文件或标准输出
}

// NewGenerator 创建一个新的报告生成器
func NewGenerator(format Format, output io.Writer) *Generator {
	// 如果没有指定输出，默认使用标准输出
	if output == nil {
		output = os.Stdout
	}
	return &Generator{Format: format, Output: output}
}

// GenerateReport 生成周报
func (g *Generator) GenerateReport(summary string, commits []git.CommitInfo, fromDate, toDate time.Time) error {
	// 根据格式生成报告
	switch g.Format {
	case FormatMarkdown:
		return g.generateMarkdownReport(summary, commits, fromDate, toDate)
	default: // 默认使用文本格式
		return g.generateTextReport(summary, commits, fromDate, toDate)
	}
}

// generateTextReport 生成纯文本格式的报告
func (g *Generator) generateTextReport(summary string, commits []git.CommitInfo, fromDate, toDate time.Time) error {
	// 根据时间范围确定报告类型
	reportType := g.determineReportType(fromDate, toDate)

	fmt.Fprintf(g.Output, "%s (%s 至 %s)\n", reportType, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
	fmt.Fprintln(g.Output, "==================================")
	fmt.Fprintln(g.Output)

	// 统计仓库信息
	repoStats := make(map[string]int)
	for _, commit := range commits {
		if commit.RepoPath != "" {
			repoStats[commit.RepoPath]++
		}
	}

	// 如果有多个仓库，显示仓库统计
	if len(repoStats) > 1 {
		fmt.Fprintln(g.Output, "## 仓库统计")
		for repo, count := range repoStats {
			fmt.Fprintf(g.Output, "- %s: %d 条提交\n", repo, count)
		}
		fmt.Fprintln(g.Output)
	}

	fmt.Fprintln(g.Output, "## AI 总结")
	fmt.Fprintln(g.Output, summary)
	fmt.Fprintln(g.Output)
	fmt.Fprintln(g.Output, "## 提交记录")
	fmt.Fprintf(g.Output, "共有 %d 条提交记录\n\n", len(commits))

	for i, commit := range commits {
		fmt.Fprintf(g.Output, "提交 %d:\n", i+1)
		fmt.Fprintf(g.Output, "- 哈希值: %s\n", commit.Hash[:8])
		fmt.Fprintf(g.Output, "- 作者: %s\n", commit.Author)
		fmt.Fprintf(g.Output, "- 日期: %s\n", commit.Date.Format("2006-01-02 15:04:05"))

		// 显示仓库信息（如果有多个仓库）
		if len(repoStats) > 1 && commit.RepoPath != "" {
			fmt.Fprintf(g.Output, "- 仓库: %s\n", commit.RepoPath)
		}

		// 显示分支信息
		if len(commit.Branches) > 0 {
			fmt.Fprintf(g.Output, "- 分支: %s\n", strings.Join(commit.Branches, ", "))
		}

		fmt.Fprintf(g.Output, "- 消息: %s\n\n", commit.Message)
	}

	return nil
}

// generateMarkdownReport 生成Markdown格式的报告
func (g *Generator) generateMarkdownReport(summary string, commits []git.CommitInfo, fromDate, toDate time.Time) error {
	// 根据时间范围确定报告类型
	reportType := g.determineReportType(fromDate, toDate)

	// 生成文件名用于提示
	fileName := fmt.Sprintf("%s-%s-to-%s.md",
		g.getReportTypeShort(fromDate, toDate),
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"))

	// 写入标题
	fmt.Fprintf(g.Output, "# %s (%s 至 %s)\n\n",
		reportType,
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"))

	// 统计仓库信息
	repoStats := make(map[string]int)
	for _, commit := range commits {
		if commit.RepoPath != "" {
			repoStats[commit.RepoPath]++
		}
	}

	// 如果有多个仓库，显示仓库统计
	if len(repoStats) > 1 {
		fmt.Fprintln(g.Output, "## 仓库统计")
		fmt.Fprintln(g.Output)
		for repo, count := range repoStats {
			fmt.Fprintf(g.Output, "- **%s**: %d 条提交\n", repo, count)
		}
		fmt.Fprintln(g.Output)
	}

	// 写入AI总结
	fmt.Fprintln(g.Output, "## AI 总结")
	fmt.Fprintln(g.Output, summary)
	fmt.Fprintln(g.Output)

	// 写入提交记录
	fmt.Fprintln(g.Output, "## 提交记录")
	fmt.Fprintf(g.Output, "共有 %d 条提交记录\n\n", len(commits))

	for i, commit := range commits {
		fmt.Fprintf(g.Output, "### 提交 %d\n\n", i+1)
		fmt.Fprintf(g.Output, "- **哈希值**: `%s`\n", commit.Hash[:8])
		fmt.Fprintf(g.Output, "- **作者**: %s\n", commit.Author)
		fmt.Fprintf(g.Output, "- **日期**: %s\n", commit.Date.Format("2006-01-02 15:04:05"))

		// 显示仓库信息（如果有多个仓库）
		if len(repoStats) > 1 && commit.RepoPath != "" {
			fmt.Fprintf(g.Output, "- **仓库**: `%s`\n", commit.RepoPath)
		}

		// 显示分支信息
		if len(commit.Branches) > 0 {
			fmt.Fprintf(g.Output, "- **分支**: %s\n", strings.Join(commit.Branches, ", "))
		}

		fmt.Fprintf(g.Output, "- **消息**: %s\n", commit.Message)

		if len(commit.ChangedFiles) > 0 {
			fmt.Fprintln(g.Output, "- **变更文件**:")
			for _, fileName := range commit.ChangedFiles {
				fmt.Fprintf(g.Output, "  - `%s`\n", fileName)
			}
		}

		fmt.Fprintln(g.Output)
	}

	// 如果输出不是标准输出，打印提示信息
	if g.Output != os.Stdout {
		fmt.Printf("已生成Markdown报告: %s\n", fileName)
	}
	return nil
}

// determineReportType 根据时间范围确定报告类型
func (g *Generator) determineReportType(fromDate, toDate time.Time) string {
	// 计算时间范围的天数
	daysDiff := toDate.Sub(fromDate).Hours() / 24

	switch {
	case daysDiff <= 1:
		return "工作日报"
	case daysDiff <= 7:
		return "工作周报"
	case daysDiff <= 31:
		return "工作月报"
	case daysDiff <= 366:
		return "工作年报"
	default:
		return "工作报告"
	}
}

// getReportTypeShort 获取报告类型的简短形式，用于文件名
func (g *Generator) getReportTypeShort(fromDate, toDate time.Time) string {
	// 计算时间范围的天数
	daysDiff := toDate.Sub(fromDate).Hours() / 24

	switch {
	case daysDiff <= 1:
		return "daily"
	case daysDiff <= 7:
		return "weekly"
	case daysDiff <= 31:
		return "monthly"
	case daysDiff <= 366:
		return "yearly"
	default:
		return "report"
	}
}
