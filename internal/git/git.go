package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Options Git操作的选项
type Options struct {
	RepoPath string // Git仓库路径
	Author   string // 作者名称，用于筛选提交
}

// NewGitOptions 创建新的Git选项
func NewGitOptions(repoPath string) *Options {
	// 如果没有指定路径，使用当前目录
	if repoPath == "" {
		repoPath = "."
	}

	// 创建Git选项
	opts := &Options{
		RepoPath: repoPath,
	}

	// 获取当前用户的Git用户名
	author, err := GetGitUserName(repoPath)
	if err == nil && author != "" {
		opts.Author = author
	}

	return opts
}

// CommitInfo 表示一个Git提交的信息
type CommitInfo struct {
	Hash         string
	Author       string
	Date         time.Time
	Message      string
	Branches     []string // 分支信息
	ChangedFiles []string
}

// GetCommitsBetween 获取指定时间范围内的所有提交
func GetCommitsBetween(fromDate, toDate time.Time, opts *Options) ([]CommitInfo, error) {
	// 格式化日期为git log可接受的格式
	fromStr := fromDate.Format("2006-01-02")
	toStr := toDate.Format("2006-01-02")

	// 构建git log命令的参数列表
	args := []string{
		"log",
		"--all",                            // 获取所有分支的提交
		"--pretty=format:%H|%an|%ad|%s|%D", // 添加%D获取分支信息
		"--date=iso",
		"--after=" + fromStr,
		"--before=" + toStr,
	}

	// 如果指定了作者，添加作者筛选条件
	if opts != nil && opts.Author != "" {
		args = append(args, "--author="+opts.Author)
	}

	// 构建git log命令
	cmd := exec.Command("git", args...)

	// 设置工作目录
	if opts != nil {
		cmd.Dir = opts.RepoPath
	}

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行git log失败: %w", err)
	}

	// 解析输出
	return parseCommits(string(output))
}

// GetCommitsThisWeek 获取本周的所有提交
func GetCommitsThisWeek(opts *Options) ([]CommitInfo, error) {
	// 计算本周一和下周一的日期
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}

	// 计算本周一的日期
	monday := now.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	// 计算下周一的日期
	nextMonday := monday.AddDate(0, 0, 7)

	return GetCommitsBetween(monday, nextMonday, opts)
}

// GetCommitDetails 获取指定提交的详细信息
func GetCommitDetails(hash string, opts *Options) (*CommitInfo, error) {
	// 获取提交的基本信息
	cmd := exec.Command("git", "show",
		"--pretty=format:%H|%an|%ad|%s|%D",
		"--date=iso",
		hash)

	// 设置工作目录
	if opts != nil {
		cmd.Dir = opts.RepoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取提交详情失败: %w", err)
	}

	// 解析提交信息
	commits, err := parseCommits(string(output))
	if err != nil || len(commits) == 0 {
		return nil, fmt.Errorf("解析提交详情失败: %w", err)
	}

	commit := commits[0]

	// 获取变更的文件列表
	cmdFiles := exec.Command("git", "show", "--name-only", "--pretty=format:", hash)

	// 设置工作目录
	if opts != nil {
		cmdFiles.Dir = opts.RepoPath
	}

	outputFiles, err := cmdFiles.Output()
	if err != nil {
		return nil, fmt.Errorf("获取变更文件列表失败: %w", err)
	}

	// 解析变更文件列表
	files := strings.Split(strings.TrimSpace(string(outputFiles)), "\n")
	commit.ChangedFiles = files

	return &commit, nil
}

// GetGitUserName 获取Git用户名
func GetGitUserName(repoPath string) (string, error) {
	// 构建git config命令获取用户名
	cmd := exec.Command("git", "config", "user.name")

	// 设置工作目录
	if repoPath != "" {
		cmd.Dir = repoPath
	}

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		// 如果获取失败，尝试获取全局用户名
		cmdGlobal := exec.Command("git", "config", "--global", "user.name")
		output, err = cmdGlobal.Output()
		if err != nil {
			return "", fmt.Errorf("获取Git用户名失败: %w", err)
		}
	}

	// 去除空白字符
	return strings.TrimSpace(string(output)), nil
}

// parseCommits 解析git log的输出
func parseCommits(output string) ([]CommitInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5) // 增加了分支信息字段
		if len(parts) < 5 {
			continue
		}

		hash := parts[0]
		author := parts[1]
		dateStr := parts[2]
		message := parts[3]
		refNames := parts[4]

		// 解析日期
		date, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			return nil, fmt.Errorf("解析日期失败: %w", err)
		}

		// 解析分支信息
		var branches []string
		if refNames != "" {
			// 分割引用名称（如HEAD -> main, origin/main）
			refs := strings.Split(refNames, ",")
			for _, ref := range refs {
				ref = strings.TrimSpace(ref)
				// 只保留分支名称，去除tag和HEAD指针
				switch {
				case strings.Contains(ref, "refs/heads/"):
					branch := strings.TrimPrefix(ref, "refs/heads/")
					branches = append(branches, branch)
				case strings.Contains(ref, "HEAD -> "):
					branch := strings.TrimPrefix(ref, "HEAD -> ")
					branches = append(branches, branch)
				case !strings.Contains(ref, "tag:") && !strings.HasPrefix(ref, "HEAD"):
					// 去除远程分支前缀
					if strings.Contains(ref, "/") {
						parts := strings.SplitN(ref, "/", 2)
						if len(parts) > 1 {
							branches = append(branches, parts[1])
						}
					} else {
						branches = append(branches, ref)
					}
				}
			}
		}

		// 去除重复的分支名
		uniqueBranches := make([]string, 0)
		branchMap := make(map[string]bool)
		for _, branch := range branches {
			if !branchMap[branch] {
				branchMap[branch] = true
				uniqueBranches = append(uniqueBranches, branch)
			}
		}

		commits = append(commits, CommitInfo{
			Hash:     hash,
			Author:   author,
			Date:     date,
			Message:  message,
			Branches: uniqueBranches,
		})
	}

	return commits, nil
}
