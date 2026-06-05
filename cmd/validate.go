package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/JunLang-7/novel2script/internal/formatters"
	"github.com/JunLang-7/novel2script/internal/models"

	"gopkg.in/yaml.v3"
)

// ValidateCommand 定义 validate 子命令。
type ValidateCommand struct {
	flagSet *flag.FlagSet
}

// NewValidateCommand 创建 validate 子命令。
func NewValidateCommand() *ValidateCommand {
	return &ValidateCommand{
		flagSet: flag.NewFlagSet("validate", flag.ExitOnError),
	}
}

// Name 返回子命令名称。
func (c *ValidateCommand) Name() string { return "validate" }

// Usage 返回使用说明。
func (c *ValidateCommand) Usage() string {
	return `用法: novel2script validate <YAML文件>

验证YAML剧本文件是否符合novel2script schema。

示例:
  novel2script validate script.yaml`
}

// Run 执行 validate 命令。
func (c *ValidateCommand) Run(args []string) error {
	c.flagSet.Parse(args)

	inputArgs := c.flagSet.Args()
	if len(inputArgs) < 1 {
		return fmt.Errorf("请指定YAML文件路径")
	}
	inputPath := inputArgs[0]

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var script models.Script
	if err := yaml.Unmarshal(data, &script); err != nil {
		fmt.Printf("YAML解析失败: %v\n", err)
		os.Exit(1)
	}

	if script.ScriptTitle == "" {
		fmt.Println("错误: script_title 为空")
		os.Exit(1)
	}
	if script.SourceNovel == "" {
		fmt.Println("错误: source_novel 为空")
		os.Exit(1)
	}
	if len(script.Characters) == 0 {
		fmt.Println("警告: 角色表为空")
	}
	if len(script.Acts) == 0 {
		fmt.Println("警告: 剧本正文为空")
	}

	warnings := formatters.ValidateScript(&script)
	if len(warnings) == 0 {
		fmt.Printf("验证通过: %s\n", script.ScriptTitle)
		fmt.Printf("  - 角色数: %d\n", len(script.Characters))
		sceneCount := 0
		for _, act := range script.Acts {
			sceneCount += len(act.Scenes)
		}
		fmt.Printf("  - 幕数: %d\n", len(script.Acts))
		fmt.Printf("  - 场景数: %d\n", sceneCount)
		return nil
	}

	fmt.Printf("发现 %d 个警告:\n", len(warnings))
	for _, w := range warnings {
		fmt.Printf("  - %s\n", w)
	}

	return nil
}
