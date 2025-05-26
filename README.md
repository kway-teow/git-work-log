# Git Report Generator

一个基于Git提交记录自动生成报告的工具，使用Google Gemini AI进行智能总结。支持日报、周报、月报、年报和自定义日期范围报告。

## 功能

- 自动获取Git提交记录
- 使用Google Gemini AI自动总结提交内容
- 支持多种提示词模板：
  - 基础提示词：简洁的工作摘要（默认选项）
  - 详细提示词：结构化的详细报告
  - 针对性提示词：面向不同受众的报告
- 支持多种时间范围：
  - 日报（当天）
  - 周报（本周，默认选项）
  - 月报（本月）
  - 年报（本年）
  - 自定义日期范围
  - 指定具体日期
- 生成格式化的报告（支持文本和Markdown格式）
- 支持指定Git仓库目录，可在任意位置运行
- 支持输出到文件或标准输出
- 自动获取当前Git用户的提交记录，也可指定作者

## 安装

```bash
go install github.com/kway-teow/git-work-log/cmd/git-work-log@latest
```

## 使用方法

```bash
# 设置Gemini API密钥
export GEMINI_API_KEY="your-api-key"

# 生成本周的周报（默认）
git-work-log

# 生成日报（今天的报告）
git-work-log --range day

# 生成周报
git-work-log --range week

# 生成月报（本月的报告）
git-work-log --range month

# 生成年报（本年的报告）
git-work-log --range year

# 生成指定日期的报告
git-work-log --date 2025-05-25

# 生成指定日期范围的报告
git-work-log --from 2025-05-19 --to 2025-05-26

# 指定输出格式
git-work-log --format markdown

# 指定输出文件
git-work-log --output my-weekly-report.md

# 指定Git仓库目录
git-work-log --repo /path/to/your/repo

# 指定Gemini模型
git-work-log --model gemini-pro

# 指定作者名称
git-work-log --author "Your Name"

# 使用不同的提示词类型
git-work-log --prompt basic     # 基础提示词（默认）：简洁的工作摘要
git-work-log --prompt detailed  # 详细提示词：结构化的详细报告
git-work-log --prompt targeted  # 针对性提示词：面向不同受众的报告

# 结合使用多个选项
git-work-log --from 2025-05-19 --to 2025-05-26 --format markdown --output report.md --repo /path/to/repo --model gemini-pro --author "Your Name" --prompt detailed
```

## 命令行选项

```
Usage:
  git-work-log [flags]

Flags:
  --author string   Git作者名称 (默认使用当前用户名)
  --date string     指定具体日期 (YYYY-MM-DD 格式)，与--range、--from和--to参数互斥
  --from string     开始日期 (YYYY-MM-DD 格式)，与--range和--date参数互斥
  --format string   报告格式 (text 或 markdown) (default "text")
  -h, --help         显示帮助信息
  --model string    Gemini模型名称 (默认为gemini-2.5-flash-preview-05-20)
  --output string   输出文件路径 (默认为标准输出)
  --prompt string   提示词类型 (basic=基础, detailed=详细, targeted=针对性) (default "basic")
  --range string    时间范围 (day, week, month, year)，默认为week，与--date、--from和--to参数互斥
  --repo string     Git仓库路径 (默认为当前目录)
  --to string       结束日期 (YYYY-MM-DD 格式)，与--range和--date参数互斥
```

## 配置

### API密钥

工具会从环境变量中读取Gemini API密钥：

```bash
export GEMINI_API_KEY="your-api-key"
```

### Gemini模型

默认使用 `gemini-2.5-flash-preview-05-20` 模型，但您可以使用 `--model` 参数指定其他模型，如：

- `gemini-pro`: 文本模型，免费层级可用
- `gemini-pro-vision`: 多模态模型，免费层级可用
- `gemini-1.5-pro`: 付费模型
- `gemini-1.5-flash`: 付费模型

例如：
```bash
git-work-log --model gemini-pro
```

## 许可证

MIT
