package main

import (
	"fmt"
	"os"

	"github.com/JunLang-7/novel2script/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	var runner commandRunner

	switch subcommand {
	case "convert":
		runner = cmd.NewConvertCommand()
	case "analyze":
		runner = cmd.NewAnalyzeCommand()
	case "validate":
		runner = cmd.NewValidateCommand()
	case "schema":
		runner = cmd.NewSchemaCommand()
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err := runner.Run(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

type commandRunner interface {
	Name() string
	Usage() string
	Run(args []string) error
}

func printUsage() {
	fmt.Println(`novel2script — AI 小说转剧本工具

将中文网文转换为结构化剧本 (YAML格式)。

用法:
  novel2script <命令> [参数]

可用命令:
  convert   将小说转换为剧本
  analyze   分析小说结构（角色 + 场景）
  validate  验证YAML剧本文件
  schema    输出JSON Schema定义

使用 "novel2script <命令> -h" 查看详细帮助。

环境变量:
  NOVEL2SCRIPT_API_KEY     LLM API密钥（必填）
  NOVEL2SCRIPT_MODEL       模型名称 (默认: claude-sonnet-4-20250514)
  NOVEL2SCRIPT_PROVIDER    LLM提供商 (默认: anthropic)
  NOVEL2SCRIPT_PARALLEL    并行调用数 (默认: 5)
  NOVEL2SCRIPT_CACHE_DIR   缓存目录 (默认: ~/.novel2script/cache)

示例:
  export NOVEL2SCRIPT_API_KEY=sk-ant-xxx
  novel2script convert 凡人修仙传.txt -o 凡人修仙传_剧本.yaml --parallel 5
  novel2script validate 凡修_剧本.yaml`)
}
