package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kway-teow/git-work-log/internal/ai"
	"github.com/kway-teow/git-work-log/internal/git"
	"github.com/kway-teow/git-work-log/internal/report"
	"github.com/spf13/cobra"
)

// 版本信息，由 GoReleaser 在构建时注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	// 命令行参数
	fromDate     string
	toDate       string
	outputFormat string
	outputFile   string
	repoPath     string // Git仓库路径
	reposPath    string // 仓库目录路径，分析该目录下的所有Git仓库
	modelName    string // Gemini模型名称
	authorName   string // Git作者名称
	timeRange    string // 时间范围类型：day(天)、week(周)、month(月)、year(年)
	customDate   string // 指定具体日期 (YYYY-MM-DD 格式)
	promptType   string // 提示词类型：basic(基础)、detailed(详细)、targeted(针对性) 或自定义提示词文件路径 (如: kpi.md 或 /path/to/custom.txt)
)

// rootCmd 表示根命令

var rootCmd = &cobra.Command{
	Use:   "git-work-log",
	Short: "基于Git提交记录自动生成报告",
	Long: `git-work-log 是一个基于Git提交记录自动生成报告的工具。

它使用Google Gemini AI对提交记录进行智能总结，生成格式化的报告。
支持多种时间范围：天(day)、周(week)、月(month)、年(year)或自定义日期。
支持单个仓库分析(--repo)或目录下所有仓库分析(--repos)。
默认生成本周的报告。`,
	Run: func(_ *cobra.Command, _ []string) {
		// 执行生成报告的操作
		generateReport()
	},
}

// 初始化命令行参数
// 版本子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("git-work-log 版本: %s\n", version)
		fmt.Printf("提交哈希: %s\n", commit)
		fmt.Printf("构建日期: %s\n", date)
	},
}

func init() {
	// 添加版本子命令
	rootCmd.AddCommand(versionCmd)

	// 添加命令行参数
	rootCmd.PersistentFlags().StringVar(&fromDate, "from", "", "开始日期 (YYYY-MM-DD 格式)，与--range和--date参数互斥")
	rootCmd.PersistentFlags().StringVar(&toDate, "to", "", "结束日期 (YYYY-MM-DD 格式)，与--range和--date参数互斥")
	rootCmd.PersistentFlags().StringVar(&timeRange, "range", "week", "时间范围 (day, week, month, year)，默认为week")
	rootCmd.PersistentFlags().StringVar(&customDate, "date", "", "指定具体日期 (YYYY-MM-DD 格式)，与--range、--from和--to参数互斥")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "报告格式 (text 或 markdown)")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "", "输出文件路径 (默认为标准输出)")
	rootCmd.PersistentFlags().StringVar(&repoPath, "repo", "", "Git仓库路径 (默认为当前目录)")
	rootCmd.PersistentFlags().StringVar(&reposPath, "repos", "", "仓库目录路径，分析该目录下的所有Git仓库")
	rootCmd.PersistentFlags().StringVar(&modelName, "model", "", "Gemini模型名称 (默认为gemini-2.5-flash-preview-05-20)")
	rootCmd.PersistentFlags().StringVar(&authorName, "author", "", "Git作者名称")
	rootCmd.PersistentFlags().StringVar(&promptType, "prompt", "basic", "提示词类型 (basic=基础, detailed=详细, targeted=针对性) 或自定义提示词文件路径 (如: kpi.md 或 /path/to/custom.txt)")
}

func main() {
	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// calculateTimeRange 根据时间范围类型计算开始和结束日期
func calculateTimeRange(rangeType string) (time.Time, time.Time) {
	now := time.Now()
	var from, to time.Time

	switch rangeType {
	case "day":
		// 今天（从今天0点到现在）
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		to = now
	case "week":
		// 过去7天（从7天前到现在）
		from = now.AddDate(0, 0, -7)
		to = now
	case "month":
		// 过去30天（从30天前到现在）
		from = now.AddDate(0, 0, -30)
		to = now
	case "year":
		// 过去365天（从365天前到现在）
		from = now.AddDate(0, 0, -365)
		to = now
	default:
		// 默认使用过去7天
		from = now.AddDate(0, 0, -7)
		to = now
	}

	return from, to
}

// generateReport 生成报告
func generateReport() {
	// 检查环境变量
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("错误: 未设置GEMINI_API_KEY环境变量")
		os.Exit(1)
	}

	// 创建Gemini客户端
	geminiClient, err := ai.NewGeminiClientWithModel(modelName)
	if err != nil {
		fmt.Printf("错误: 创建Gemini客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 判断使用何种时间范围
	var from, to time.Time
	var err1, err2 error

	switch {
	case fromDate != "" && toDate != "":
		// 使用自定义时间范围（从某天到某天）
		// 解析指定的日期范围
		from, err1 = time.Parse("2006-01-02", fromDate)
		to, err2 = time.Parse("2006-01-02", toDate)
		if err1 != nil || err2 != nil {
			fmt.Println("错误: 日期格式不正确，请使用YYYY-MM-DD格式")
			os.Exit(1)
		}
		// 调整结束日期为当天结束
		to = to.Add(24*time.Hour - time.Second)
		fmt.Printf("使用自定义时间范围: %s 到 %s\n", fromDate, toDate)
	case customDate != "":
		// 使用指定日期
		// 解析指定的日期
		specificDate, err := time.Parse("2006-01-02", customDate)
		if err != nil {
			fmt.Println("错误: 日期格式不正确，请使用YYYY-MM-DD格式")
			os.Exit(1)
		}
		// 设置为指定日期的0点到次日0点
		from = time.Date(specificDate.Year(), specificDate.Month(), specificDate.Day(), 0, 0, 0, 0, specificDate.Location())
		to = from.AddDate(0, 0, 1)
		fmt.Printf("使用指定日期: %s\n", customDate)
	default:
		// 使用预定义的时间范围
		from, to = calculateTimeRange(timeRange)
		fmt.Printf("使用预定义时间范围 %s: %s 到 %s\n", timeRange, from.Format("2006-01-02"), to.Format("2006-01-02 15:04"))
	}

	// 判断使用何种分析模式：单仓库还是多仓库
	var repoPaths []string
	var discoveryErr error

	switch {
	case reposPath != "":
		// 多仓库模式：发现指定目录下的所有Git仓库
		repoPaths, discoveryErr = git.DiscoverGitRepos(reposPath)
		if discoveryErr != nil {
			fmt.Printf("错误: 发现Git仓库失败: %v\n", discoveryErr)
			return
		}

		if len(repoPaths) == 0 {
			fmt.Printf("在目录 %s 下没有发现任何Git仓库\n", reposPath)
			return
		}
	case repoPath != "":
		// 单仓库模式：使用指定的仓库路径
		repoPaths = []string{repoPath}
	default:
		// 默认模式：使用当前目录
		repoPaths = []string{"."}
	}

	// 收集所有仓库的提交记录
	var allCommits []git.CommitInfo
	repoCommitCounts := make(map[string]int)

	fmt.Printf("\n处理 %d 个仓库:\n", len(repoPaths))

	for _, currentRepoPath := range repoPaths {
		fmt.Printf("正在分析仓库: %s\n", currentRepoPath)

		// 创建Git选项
		gitOpts := git.NewGitOptions(currentRepoPath)

		// 如果命令行指定了作者名称，覆盖自动检测的用户名
		if authorName != "" {
			gitOpts.Author = authorName
		}

		// 获取提交记录
		commits, commitErr := git.GetCommitsBetween(from, to, gitOpts)
		if commitErr != nil {
			fmt.Printf("  警告: 仓库 %s 获取Git提交记录失败: %v\n", currentRepoPath, commitErr)
			continue
		}

		// 记录每个仓库的提交数量
		repoCommitCounts[currentRepoPath] = len(commits)

		// 为每个提交添加仓库信息
		for i := range commits {
			commits[i].RepoPath = currentRepoPath
		}

		// 合并到总的提交列表
		allCommits = append(allCommits, commits...)

		fmt.Printf("  找到 %d 条提交记录\n", len(commits))
	}

	// 显示汇总统计信息
	fmt.Printf("\n=== 提交记录统计 ===\n")
	totalCommits := 0
	for repoPath, count := range repoCommitCounts {
		// 显示相对路径，更清晰
		displayPath := repoPath
		if reposPath != "" {
			if rel, err := filepath.Rel(reposPath, repoPath); err == nil {
				displayPath = rel
			}
		}
		fmt.Printf("  %s: %d 条提交\n", displayPath, count)
		totalCommits += count
	}
	fmt.Printf("总计: %d 条提交\n\n", totalCommits)

	if len(allCommits) == 0 {
		fmt.Printf("指定时间范围 %s 到 %s 在所有仓库中都没有找到提交记录\n", from.Format("2006-01-02"), to.Format("2006-01-02"))
		return
	}

	// 显示作者信息
	if authorName != "" {
		fmt.Printf("筛选作者: %s\n", authorName)
	} else {
		fmt.Println("获取所有作者的提交")
	}

	fmt.Printf("正在生成报告...\n")

	// 使用AI生成报告
	fmt.Println("使用AI生成摘要...")

	// 重用之前创建的客户端变量
	if geminiClient == nil {
		geminiClient, err = ai.NewGeminiClientWithModel(modelName)
		if err != nil {
			fmt.Printf("错误: 创建Gemini客户端失败: %v\n", err)
			return
		}
	}
	defer geminiClient.Close()

	// 根据选择的提示词类型确定使用哪种提示词
	aiPromptType := ai.GetPromptTypeFromString(promptType)

	if ai.IsCustomPrompt(aiPromptType) {
		fmt.Printf("使用自定义提示词文件: %s\n", promptType)
	} else {
		switch aiPromptType {
		case ai.BasicPrompt:
			fmt.Println("使用基础提示词生成报告")
		case ai.DetailedPrompt:
			fmt.Println("使用详细提示词生成报告")
		case ai.TargetedPrompt:
			fmt.Println("使用针对性提示词生成报告")
		default:
			fmt.Println("使用默认基础提示词生成报告")
		}
	}

	// 使用AI生成报告
	reportSummary, err := geminiClient.SummarizeCommitsWithPrompt(allCommits, aiPromptType)
	if err != nil {
		fmt.Printf("错误: 生成报告摘要失败: %v\n", err)
		return
	}

	// 决定输出目标
	var output io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建输出文件失败: %v\n", err)
			return
		}
		defer file.Close()
		output = file
	}

	// 创建报告生成器
	reportFormat := report.Format(outputFormat)
	reportGenerator := report.NewGenerator(reportFormat, output)

	// 生成并输出报告
	err = reportGenerator.GenerateReport(reportSummary, allCommits, from, to)
	if err != nil {
		fmt.Printf("错误: 输出报告失败: %v\n", err)
		return
	}

	// 根据时间范围类型显示不同的完成消息
	reportType := getReportTypeShort()

	fmt.Printf("%s生成完成！\n", reportType)
}

// getReportTypeShort 获取报告类型的简短描述
func getReportTypeShort() string {
	switch {
	case fromDate != "" && toDate != "":
		return "自定义时间范围报告"
	case customDate != "":
		return "日报"
	case timeRange == "day":
		return "日报"
	case timeRange == "week":
		return "周报"
	case timeRange == "month":
		return "月报"
	case timeRange == "year":
		return "年报"
	default:
		return "报告"
	}
}
