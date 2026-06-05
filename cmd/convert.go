package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/JunLang-7/novel2script/internal/config"
	"github.com/JunLang-7/novel2script/internal/formatters"
	"github.com/JunLang-7/novel2script/internal/llm"
	"github.com/JunLang-7/novel2script/internal/pipeline"
	"github.com/JunLang-7/novel2script/internal/text"
)

// ConvertCommand 定义 convert 子命令。
type ConvertCommand struct {
	flagSet      *flag.FlagSet
	output       string
	format       string
	startChapter int
	endChapter   int
	model        string
	parallel     int
	dryRun       bool
	verbose      bool
	resume       bool
}

// NewConvertCommand 创建 convert 子命令。
func NewConvertCommand() *ConvertCommand {
	c := &ConvertCommand{}
	c.flagSet = flag.NewFlagSet("convert", flag.ExitOnError)
	c.flagSet.StringVar(&c.output, "o", "script.yaml", "输出文件路径")
	c.flagSet.StringVar(&c.output, "output", "script.yaml", "输出文件路径")
	c.flagSet.StringVar(&c.format, "f", "yaml", "输出格式: yaml | json | md")
	c.flagSet.StringVar(&c.format, "format", "yaml", "输出格式: yaml | json | md")
	c.flagSet.IntVar(&c.startChapter, "s", 0, "起始章节号（从1开始）")
	c.flagSet.IntVar(&c.startChapter, "start", 0, "起始章节号（从1开始）")
	c.flagSet.IntVar(&c.endChapter, "e", 0, "结束章节号")
	c.flagSet.IntVar(&c.endChapter, "end", 0, "结束章节号")
	c.flagSet.StringVar(&c.model, "m", "", "LLM模型名称")
	c.flagSet.StringVar(&c.model, "model", "", "LLM模型名称")
	c.flagSet.IntVar(&c.parallel, "p", 0, "并行LLM调用数")
	c.flagSet.IntVar(&c.parallel, "parallel", 0, "并行LLM调用数")
	c.flagSet.BoolVar(&c.dryRun, "n", false, "仅分析不转换：输出分块统计和预估成本")
	c.flagSet.BoolVar(&c.dryRun, "dry-run", false, "仅分析不转换：输出分块统计和预估成本")
	c.flagSet.BoolVar(&c.verbose, "v", false, "详细日志输出")
	c.flagSet.BoolVar(&c.verbose, "verbose", false, "详细日志输出")
	c.flagSet.BoolVar(&c.resume, "r", false, "从上次中断处继续")
	c.flagSet.BoolVar(&c.resume, "resume", false, "从上次中断处继续")
	return c
}

// Name 返回子命令名称。
func (c *ConvertCommand) Name() string { return "convert" }

// Usage 返回使用说明。
func (c *ConvertCommand) Usage() string {
	return `用法: novel2script convert <输入文件> [选项]

将小说转换为结构化剧本YAML。

选项:
  -o, --output     输出文件路径 (默认: script.yaml)
  -f, --format     输出格式: yaml | md (默认: yaml)
  -s, --start      起始章节号（从1开始）
  -e, --end        结束章节号
  -m, --model      LLM模型名称
  -p, --parallel   并行LLM调用数 (默认: 5)
  -n, --dry-run    仅分析不转换
  -v, --verbose    详细日志输出
  -r, --resume     从上次中断处继续

示例:
  novel2script convert 凡人修仙传.txt -o 凡人修仙传_剧本.yaml --parallel 5
  novel2script convert sample.md -f md -o readable_draft.md
  novel2script convert big_novel.txt --start 1 --end 50 -o partial.yaml
  novel2script convert big_novel.txt --dry-run`
}

// Run 执行 convert 命令。
func (c *ConvertCommand) Run(args []string) error {
	c.flagSet.Parse(args)

	// 获取输入文件
	inputArgs := c.flagSet.Args()
	if len(inputArgs) < 1 {
		return fmt.Errorf("请指定输入文件路径，使用 -h 查看帮助")
	}
	inputPath := inputArgs[0]

	// 加载配置
	cfg := config.Load()
	if c.model != "" {
		cfg.Model = c.model
	}
	if c.parallel > 0 {
		cfg.Parallel = c.parallel
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("请设置环境变量 NOVEL2SCRIPT_API_KEY")
	}

	// 读取文件
	rawText, err := text.DetectAndReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 应用章节范围过滤
	if c.startChapter > 0 || c.endChapter > 0 {
		chapters, err := text.SplitChapters(rawText)
		if err != nil {
			return fmt.Errorf("章节检测失败: %w", err)
		}
		start := max(c.startChapter-1, 0)
		end := len(chapters)
		if c.endChapter > 0 && c.endChapter <= len(chapters) {
			end = c.endChapter
		}
		if start >= len(chapters) {
			return fmt.Errorf("起始章节 %d 超出总章节数 %d", c.startChapter, len(chapters))
		}
		selected := chapters[start:end]
		var sb strings.Builder
		for i, ch := range selected {
			if i > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(ch.Content)
		}
		rawText = sb.String()
	}

	// Dry-run 模式
	if c.dryRun {
		return runDryRun(rawText)
	}

	// 创建 LLM 客户端
	client := llm.NewClient(llm.Config{
		Provider:    cfg.Provider,
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		MaxRetries:  3,
		MaxParallel: cfg.Parallel,
	})

	// 运行管道
	orch := pipeline.NewOrchestrator(client, pipeline.OrchestratorConfig{
		TokensPerChunk: 15000,
		Parallelism:    cfg.Parallel,
		Verbose:        c.verbose,
	})

	script, stats, err := orch.Run(context.Background(), rawText)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}

	// 输出统计
	if c.verbose {
		fmt.Fprintf(os.Stderr, "处理统计:\n")
		fmt.Fprintf(os.Stderr, "  章节数: %d\n", stats.TotalChapters)
		fmt.Fprintf(os.Stderr, "  总字数: %s\n", text.FormatCharCount(stats.TotalChars))
		fmt.Fprintf(os.Stderr, "  处理批次: %d\n", stats.NumChunks)
		fmt.Fprintf(os.Stderr, "  LLM调用次数: %d\n", stats.NumLLMCalls)
		fmt.Fprintf(os.Stderr, "  耗时: %v\n", stats.Duration)
	}

	// 校验
	if c.verbose {
		warnings := formatters.ValidateScript(script)
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "警告: %s\n", w)
		}
	}

	// 输出
	switch strings.ToLower(c.format) {
	case "md", "markdown":
		f, err := os.Create(c.output)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer f.Close()
		return formatters.WriteMarkdown(f, script)
	case "json":
		return fmt.Errorf("JSON 格式暂未支持，请使用 yaml 或 md 格式")
	case "yaml", "":
		f, err := os.Create(c.output)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer f.Close()
		return formatters.WriteYAML(f, script)
	default:
		return fmt.Errorf("不支持的输出格式: %s，支持 yaml | md", c.format)
	}
}

func runDryRun(rawText string) error {
	chapters, err := text.SplitChapters(rawText)
	if err != nil {
		return err
	}

	chunks := text.GroupIntoChunks(chapters, 15000)

	fmt.Printf("=== Dry Run 分析报告 ===\n\n")
	fmt.Printf("总章节数: %d\n", len(chapters))
	fmt.Printf("总字数: %s\n", text.FormatCharCount(len([]rune(rawText))))
	fmt.Printf("处理批次: %d\n", len(chunks))

	totalTokens := 0
	for _, chunk := range chunks {
		totalTokens += chunk.TokenEst
	}

	fmt.Printf("\n--- 分块详情 ---\n")
	for _, chunk := range chunks {
		fmt.Printf("  %s: 第%d-%d章, %s, ~%d tokens\n",
			chunk.ID, chunk.ChapterStart, chunk.ChapterEnd,
			text.FormatCharCount(chunk.CharCount), chunk.TokenEst)
	}

	// 成本预估（基于 Claude Sonnet 价格）
	estCalls := len(chunks)*3 + 2 // 角色提取 + 场景分割 + 剧本转换
	estInputTokens := totalTokens * 3
	estOutputTokens := totalTokens / 2

	fmt.Printf("\n--- 成本预估 ---\n")
	fmt.Printf("预估LLM调用次数: %d\n", estCalls)
	fmt.Printf("预估输入tokens: ~%d\n", estInputTokens)
	fmt.Printf("预估输出tokens: ~%d\n", estOutputTokens)
	fmt.Printf("预估成本(Claude Sonnet): ~$%.2f\n",
		float64(estInputTokens)/1_000_000*3.0+float64(estOutputTokens)/1_000_000*15.0)

	return nil
}
