package main

import (
	"fmt"
	"io"
	"os"
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
	fromDate    string
	toDate      string
	outputFormat string
	outputFile  string
	repoPath    string // Git仓库路径
	modelName   string // Gemini模型名称
	authorName  string // Git作者名称
	timeRange   string // 时间范围类型：day(天)、week(周)、month(月)、year(年)
	customDate  string // 指定具体日期 (YYYY-MM-DD 格式)
	promptType  string // 提示词类型：basic(基础)、detailed(详细)、targeted(针对性)
)

// rootCmd 表示根命令

var rootCmd = &cobra.Command{
	Use:   "git-work-log",
	Short: "基于Git提交记录自动生成报告",
	Long: `git-work-log 是一个基于Git提交记录自动生成报告的工具。

它使用Google Gemini AI对提交记录进行智能总结，生成格式化的报告。
支持多种时间范围：天(day)、周(week)、月(month)、年(year)或自定义日期。
默认生成本周的报告。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 执行生成报告的操作
		generateReport()
	},
}

// 初始化命令行参数
// 版本子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
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
	rootCmd.PersistentFlags().StringVar(&modelName, "model", "", "Gemini模型名称 (默认为gemini-2.5-flash-preview-05-20)")
	rootCmd.PersistentFlags().StringVar(&authorName, "author", "", "Git作者名称")
	rootCmd.PersistentFlags().StringVar(&promptType, "prompt", "basic", "提示词类型 (basic=基础, detailed=详细, targeted=针对性)")
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
		// 今天
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 0, 1)
	case "week":
		// 本周（周一到周日）
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		from = now.AddDate(0, 0, -int(weekday)+1)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		to = from.AddDate(0, 0, 7)
	case "month":
		// 本月
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 1, 0)
	case "year":
		// 本年
		from = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		to = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
	default:
		// 默认使用周
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		from = now.AddDate(0, 0, -int(weekday)+1)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		to = from.AddDate(0, 0, 7)
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

	if fromDate != "" && toDate != "" {
		// 使用自定义时间范围（从某天到某天）
		fmt.Println("使用自定义时间范围（从某天到某天）")
		// 解析指定的日期范围
		from, err1 = time.Parse("2006-01-02", fromDate)
		to, err2 = time.Parse("2006-01-02", toDate)
		if err1 != nil || err2 != nil {
			fmt.Println("错误: 日期格式不正确，请使用YYYY-MM-DD格式")
			os.Exit(1)
		}
		// 调整结束日期为当天结束
		to = to.Add(24*time.Hour - time.Second)
	} else if customDate != "" {
		// 使用指定日期
		fmt.Println("使用指定日期")
		// 解析指定的日期
		specificDate, err := time.Parse("2006-01-02", customDate)
		if err != nil {
			fmt.Println("错误: 日期格式不正确，请使用YYYY-MM-DD格式")
			os.Exit(1)
		}
		// 设置为指定日期的0点到次日0点
		from = time.Date(specificDate.Year(), specificDate.Month(), specificDate.Day(), 0, 0, 0, 0, specificDate.Location())
		to = from.AddDate(0, 0, 1)
	} else {
		// 使用预定义的时间范围
		fmt.Printf("使用预定义时间范围: %s\n", timeRange)
		from, to = calculateTimeRange(timeRange)
	}

	// 创建Git选项
	gitOpts := git.NewGitOptions(repoPath)
	
	// 如果命令行指定了作者名称，覆盖自动检测的用户名
	if authorName != "" {
		gitOpts.Author = authorName
		fmt.Printf("使用指定作者: %s\n", authorName)
	} else if gitOpts.Author != "" {
		fmt.Printf("使用当前 Git 用户: %s\n", gitOpts.Author)
	} else {
		fmt.Println("未指定作者，将获取所有提交")
	}

	// 获取提交记录
	commits, err := git.GetCommitsBetween(from, to, gitOpts)
	if err != nil {
		fmt.Printf("错误: 获取Git提交记录失败: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Printf("指定时间范围 %s 到 %s 没有找到提交记录\n", from.Format("2006-01-02"), to.Format("2006-01-02"))
		os.Exit(0)
	}

	fmt.Printf("找到 %d 条提交记录，正在生成报告...\n", len(commits))

	// 使用AI生成报告
	fmt.Println("使用AI生成摘要...")
	
	// 重用之前创建的客户端变量
	if geminiClient == nil {
		geminiClient, err = ai.NewGeminiClientWithModel(modelName)
		if err != nil {
			fmt.Printf("错误: 创建Gemini客户端失败: %v\n", err)
			os.Exit(1)
		}
	}
	defer geminiClient.Close()

	// 根据选择的提示词类型确定使用哪种提示词
	var aiPromptType ai.PromptType
	switch promptType {
	case "basic":
		aiPromptType = ai.BasicPrompt
		fmt.Println("使用基础提示词生成报告")
	case "detailed":
		aiPromptType = ai.DetailedPrompt
		fmt.Println("使用详细提示词生成报告")
	case "targeted":
		aiPromptType = ai.TargetedPrompt
		fmt.Println("使用针对性提示词生成报告")
	default:
		aiPromptType = ai.BasicPrompt
		fmt.Println("使用默认基础提示词生成报告")
	}

	// 使用AI生成报告
	reportSummary, err := geminiClient.SummarizeCommitsWithPrompt(commits, aiPromptType)
	if err != nil {
		fmt.Printf("错误: 生成报告摘要失败: %v\n", err)
		os.Exit(1)
	}

	// 决定输出目标
	var output io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建输出文件失败: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	// 创建报告生成器
	reportFormat := report.Format(outputFormat)
	reportGenerator := report.NewGenerator(reportFormat, output)

	// 生成并输出报告
	err = reportGenerator.GenerateReport(reportSummary, commits, from, to)
	if err != nil {
		fmt.Printf("错误: 输出报告失败: %v\n", err)
		os.Exit(1)
	}

	// 根据时间范围类型显示不同的完成消息
	var reportType string
	switch {
	case fromDate != "" && toDate != "":
		reportType = "自定义时间范围报告"
	case customDate != "":
		reportType = "日报"
	case timeRange == "day":
		reportType = "日报"
	case timeRange == "week":
		reportType = "周报"
	case timeRange == "month":
		reportType = "月报"
	case timeRange == "year":
		reportType = "年报"
	default:
		reportType = "报告"
	}

	fmt.Printf("%s生成完成！\n", reportType)
}
