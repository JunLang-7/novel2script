package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JunLang-7/novel2script/internal/config"
	"github.com/JunLang-7/novel2script/internal/llm"
	"github.com/JunLang-7/novel2script/internal/pipeline"
	"github.com/JunLang-7/novel2script/internal/storage"
	"github.com/JunLang-7/novel2script/internal/text"
)

// AnalyzeCommand 定义 analyze 子命令。
type AnalyzeCommand struct {
	flagSet      *flag.FlagSet
	output       string
	model        string
	parallel     int
	verbose      bool
	resume       bool
	startChapter int
	endChapter   int
}

// NewAnalyzeCommand 创建 analyze 子命令。
func NewAnalyzeCommand() *AnalyzeCommand {
	c := &AnalyzeCommand{}
	c.flagSet = flag.NewFlagSet("analyze", flag.ExitOnError)
	c.flagSet.StringVar(&c.output, "o", "analysis.json", "分析结果输出路径")
	c.flagSet.StringVar(&c.output, "output", "analysis.json", "分析结果输出路径")
	c.flagSet.StringVar(&c.model, "m", "", "LLM模型名称")
	c.flagSet.StringVar(&c.model, "model", "", "LLM模型名称")
	c.flagSet.IntVar(&c.parallel, "p", 0, "并行LLM调用数")
	c.flagSet.IntVar(&c.parallel, "parallel", 0, "并行LLM调用数")
	c.flagSet.BoolVar(&c.verbose, "v", false, "详细日志")
	c.flagSet.BoolVar(&c.verbose, "verbose", false, "详细日志")
	c.flagSet.BoolVar(&c.resume, "r", false, "从上次中断处继续")
	c.flagSet.BoolVar(&c.resume, "resume", false, "从上次中断处继续")
	c.flagSet.IntVar(&c.startChapter, "s", 0, "起始章节号（从1开始）")
	c.flagSet.IntVar(&c.startChapter, "start", 0, "起始章节号（从1开始）")
	c.flagSet.IntVar(&c.endChapter, "e", 0, "结束章节号")
	c.flagSet.IntVar(&c.endChapter, "end", 0, "结束章节号")
	return c
}

// Name 返回子命令名称。
func (c *AnalyzeCommand) Name() string { return "analyze" }

// Usage 返回使用说明。
func (c *AnalyzeCommand) Usage() string {
	return `用法: novel2script analyze <输入文件> [选项]

仅分析小说（角色提取 + 场景分割），不做完整剧本转换。

选项:
  -o, --output     分析结果输出路径 (默认: analysis.json)
  -m, --model      LLM模型名称
  -p, --parallel   并行LLM调用数 (默认: 5)
  -s, --start      起始章节号（从1开始）
  -e, --end        结束章节号
  -r, --resume     从上次中断处继续
  -v, --verbose    详细日志输出

示例:
  novel2script analyze 小说.txt -o analysis.json
  novel2script analyze 小说.txt -s 1 -e 10 --resume`
}

// Run 执行 analyze 命令。
func (c *AnalyzeCommand) Run(args []string) error {
	c.flagSet.Parse(reorderArgs(args))

	inputArgs := c.flagSet.Args()
	if len(inputArgs) < 1 {
		return fmt.Errorf("请指定输入文件路径")
	}
	inputPath := inputArgs[0]

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

	rawText, err := text.DetectAndReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

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

	client := llm.NewClient(llm.Config{
		Provider:    cfg.Provider,
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		MaxRetries:  3,
		MaxParallel: cfg.Parallel,
	})

	var cache storage.Cache
	if c.resume {
		sqliteCache, err := storage.NewSQLiteCache(cfg.CacheDir)
		if err != nil {
			return fmt.Errorf("创建缓存失败: %w", err)
		}
		defer sqliteCache.Close()
		cache = sqliteCache
	}

	orch := pipeline.NewOrchestrator(client, pipeline.OrchestratorConfig{
		TokensPerChunk: 15000,
		Parallelism:    cfg.Parallel,
		Verbose:        c.verbose,
		Cache:          cache,
	})

	// 仅运行分析阶段（角色提取 + 场景分割），不做剧本转换
	result, stats, err := orch.Analyze(context.Background(), rawText)
	if err != nil {
		return fmt.Errorf("分析失败: %w", err)
	}

	fmt.Printf("分析完成 | 章节: %d | 字数: %s | 角色: %d | 场景: %d | LLM调用: %d | 耗时: %v\n",
		stats.TotalChapters,
		text.FormatCharCount(stats.TotalChars),
		len(result.Characters),
		len(result.Scenes),
		stats.NumLLMCalls,
		stats.Duration,
	)
	fmt.Printf("输入 tokens: %d | 输出 tokens: %d\n",
		stats.TotalInputTokens,
		stats.TotalOutputTokens,
	)

	if err := os.MkdirAll(filepath.Dir(c.output), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	f, err := os.Create(c.output)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
