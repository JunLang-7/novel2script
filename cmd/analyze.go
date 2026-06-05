package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/JunLang-7/novel2script/internal/config"
	"github.com/JunLang-7/novel2script/internal/formatters"
	"github.com/JunLang-7/novel2script/internal/llm"
	"github.com/JunLang-7/novel2script/internal/pipeline"
	"github.com/JunLang-7/novel2script/internal/text"
)

// AnalyzeCommand 定义 analyze 子命令。
type AnalyzeCommand struct {
	flagSet  *flag.FlagSet
	output   string
	model    string
	parallel int
	verbose  bool
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
  -p, --parallel   并行LLM调用数

示例:
  novel2script analyze 小说.txt -o analysis.json`
}

// Run 执行 analyze 命令。
func (c *AnalyzeCommand) Run(args []string) error {
	c.flagSet.Parse(args)

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

	client := llm.NewClient(llm.Config{
		Provider:    cfg.Provider,
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		MaxRetries:  3,
		MaxParallel: cfg.Parallel,
	})

	orch := pipeline.NewOrchestrator(client, pipeline.OrchestratorConfig{
		TokensPerChunk: 15000,
		Parallelism:    cfg.Parallel,
		Verbose:        c.verbose,
	})

	// 仅运行分析和分割（不转换剧本元素）
	// 实际实现中，orchestrator 可以配置模式
	script, stats, err := orch.Run(context.Background(), rawText)
	if err != nil {
		return fmt.Errorf("分析失败: %w", err)
	}

	fmt.Printf("分析完成 | 章节: %d | 字数: %s | 场景: %d | 角色: %d | 耗时: %v\n",
		stats.TotalChapters,
		text.FormatCharCount(stats.TotalChars),
		stats.NumLLMCalls,
		len(script.Characters),
		stats.Duration,
	)

	f, err := os.Create(c.output)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer f.Close()
	return formatters.WriteYAML(f, script)
}
