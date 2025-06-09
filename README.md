# Git Report Generator

一个基于Git提交记录自动生成报告的工具，使用Google Gemini AI进行智能总结。支持日报、周报、月报、年报和自定义日期范围报告。

## 功能

- 自动获取Git提交记录
- 使用Google Gemini AI自动总结提交内容
- 支持多种分析模式：
  - **单仓库分析**：分析指定的单个Git仓库
  - **多仓库分析**：自动发现并分析目录下的所有Git仓库
- 支持多种提示词模板：
  - 基础提示词：简洁的工作摘要（默认选项）
  - 详细提示词：结构化的详细报告
  - 针对性提示词：面向不同受众的报告
- 支持多种时间范围：
  - 日报（今天）
  - 周报（过去7天，默认选项）
  - 月报（过去30天）
  - 年报（过去365天）
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

# 生成周报（过去7天的报告）
git-work-log --range week

# 生成月报（过去30天的报告）
git-work-log --range month

# 生成年报（过去365天的报告）
git-work-log --range year

# 生成指定日期的报告
git-work-log --date 2025-05-25

# 生成指定日期范围的报告
git-work-log --from 2025-05-19 --to 2025-05-26

# 指定输出格式
git-work-log --format markdown

# 指定输出文件
git-work-log --output my-weekly-report.md

# 指定Git仓库目录（单仓库分析）
git-work-log --repo /path/to/your/repo

# 分析指定目录下的所有Git仓库（多仓库分析）
git-work-log --repos /path/to/projects/directory

# 分析当前目录下的所有Git仓库
git-work-log --repos .

# 分析父目录下的所有Git仓库
git-work-log --repos ..

# 指定Gemini模型
git-work-log --model gemini-pro

# 指定作者名称
git-work-log --author "Your Name"

# 使用不同的提示词类型
git-work-log --prompt basic     # 基础提示词（默认）：简洁的工作摘要
git-work-log --prompt detailed  # 详细提示词：结构化的详细报告
git-work-log --prompt targeted  # 针对性提示词：面向不同受众的报告

# 使用自定义提示词文件
git-work-log --prompt kpi-prompt.md        # 使用KPI模板
git-work-log --prompt /path/to/custom.txt  # 使用自定义提示词文件
git-work-log --prompt my-template.md       # 使用相对路径的自定义模板

# 多仓库分析示例
git-work-log --repos /Users/dev/projects --range month --format markdown --output monthly-report.md

# 结合使用多个选项（单仓库）
git-work-log --from 2025-05-19 --to 2025-05-26 --format markdown --output report.md --repo /path/to/repo --model gemini-pro --author "Your Name" --prompt detailed

# 结合使用多个选项（多仓库）
git-work-log --repos /path/to/projects --range week --format markdown --output weekly-summary.md --author "Your Name" --prompt detailed
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
  --range string    时间范围 (day=今天, week=过去7天, month=过去30天, year=过去365天)，默认为week，与--date、--from和--to参数互斥
  --repo string     Git仓库路径 (默认为当前目录)
  --repos string    仓库目录路径，分析该目录下的所有Git仓库
  --to string       结束日期 (YYYY-MM-DD 格式)，与--range和--date参数互斥
```

## 自定义提示词

除了预设的三种提示词类型外，您还可以使用自定义的提示词文件：

### 使用方法

```bash
# 使用预设提示词
git-work-log --prompt basic
git-work-log --prompt detailed  
git-work-log --prompt targeted

# 使用自定义提示词文件
git-work-log --prompt kpi-prompt.md        # 相对路径
git-work-log --prompt /path/to/custom.txt  # 绝对路径
```

### 自定义提示词文件格式

自定义提示词文件应该是一个纯文本文件（支持 .txt、.md 等格式），包含您希望AI使用的提示词内容。

**重要：** 在提示词文件中，使用 `{{.CommitMessages}}` 作为占位符，系统会自动将Git提交记录插入到这个位置。

### 示例：KPI报告模板

项目中包含了一个 `kpi-prompt.md` 示例文件，展示如何创建符合KPI考核要求的报告模板：

```bash
git-work-log --prompt kpi-prompt.md --range month --format markdown
```

### 创建自定义提示词

1. 创建一个文本文件（如 `my-template.md`）
2. 编写您的提示词内容
3. 在需要插入提交记录的地方使用 `{{.CommitMessages}}`
4. 使用 `--prompt` 参数指定文件路径

示例提示词文件：
```
请根据以下提交记录生成技术总结报告：

## 技术要点
- 重点关注新技术的应用
- 说明解决的技术难点
- 量化性能改进

## 提交记录
{{.CommitMessages}}

请按照上述要求生成报告。
```

## 多仓库分析

该工具支持两种分析模式：

### 单仓库分析（使用 `--repo`）
分析指定的单个Git仓库：
```bash
git-work-log --repo /path/to/single/repo
```

### 多仓库分析（使用 `--repos`）
自动发现并分析指定目录下的所有Git仓库：
```bash
# 分析指定目录下的所有仓库
git-work-log --repos /path/to/projects

# 分析当前目录下的所有仓库
git-work-log --repos .

# 分析父目录下的所有仓库
git-work-log --repos ..
```

### 多仓库分析特性

- **自动发现**：递归扫描目录，自动发现所有包含`.git`文件夹的仓库
- **统一报告**：将所有仓库的提交记录合并生成统一报告
- **仓库统计**：显示每个仓库的提交数量统计
- **失败容错**：单个仓库分析失败不影响其他仓库
- **路径标识**：多仓库时自动标识每个提交的来源仓库

### 使用场景

- **团队协作**：分析团队项目目录下的所有仓库
- **个人项目**：生成个人所有项目的工作总结
- **定期报告**：为管理层生成跨项目的工作报告
- **代码审查**：快速了解多个项目的最新进展

示例输出：
```
发现 3 个Git仓库:
  - /Users/dev/projects/frontend
  - /Users/dev/projects/backend
  - /Users/dev/projects/mobile

=== 提交记录统计 ===
  frontend: 15 条提交
  backend: 8 条提交
  mobile: 3 条提交
总计: 26 条提交
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
