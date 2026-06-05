package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// SchemaCommand 定义 schema 子命令。
type SchemaCommand struct {
	flagSet *flag.FlagSet
	output  string
}

// NewSchemaCommand 创建 schema 子命令。
func NewSchemaCommand() *SchemaCommand {
	c := &SchemaCommand{}
	c.flagSet = flag.NewFlagSet("schema", flag.ExitOnError)
	c.flagSet.StringVar(&c.output, "o", "", "输出schema到文件（默认stdout）")
	c.flagSet.StringVar(&c.output, "output", "", "输出schema到文件（默认stdout）")
	return c
}

// Name 返回子命令名称。
func (c *SchemaCommand) Name() string { return "schema" }

// Usage 返回使用说明。
func (c *SchemaCommand) Usage() string {
	return `用法: novel2script schema [选项]

输出JSON Schema定义。

选项:
  -o, --output  输出schema到文件（默认stdout）

示例:
  novel2script schema
  novel2script schema -o script-schema.json`
}

// scriptSchemaJSON 包含剧本的 JSON Schema 定义。
const scriptSchemaJSON = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://novel2script.dev/schemas/script-v1.json",
  "title": "novel2script Script Schema",
  "description": "AI小说转剧本工具输出的YAML剧本格式定义",
  "type": "object",
  "required": ["script_title", "source_novel", "characters", "acts"],
  "properties": {
    "script_title": {
      "type": "string",
      "description": "剧本名称"
    },
    "source_novel": {
      "type": "string",
      "description": "原著名称"
    },
    "source_author": {
      "type": "string",
      "description": "原著作者"
    },
    "adaptor": {
      "type": "string",
      "description": "改编工具及版本",
      "default": "novel2script v1.0"
    },
    "generated_at": {
      "type": "string",
      "format": "date-time",
      "description": "生成时间（ISO 8601）"
    },
    "version": {
      "type": "string",
      "description": "格式版本号"
    },
    "metadata": {
      "$ref": "#/$defs/metadata"
    },
    "characters": {
      "type": "array",
      "description": "角色表",
      "items": { "$ref": "#/$defs/character" }
    },
    "acts": {
      "type": "array",
      "description": "剧本正文（按幕组织）",
      "items": { "$ref": "#/$defs/act" }
    }
  },
  "$defs": {
    "metadata": {
      "type": "object",
      "properties": {
        "genre": { "type": "array", "items": { "type": "string" }},
        "original_language": { "type": "string" },
        "target_format": { "type": "string", "default": "screenplay" },
        "total_novel_chapters": { "type": "integer" },
        "total_novel_chars": { "type": "integer" },
        "adaptation_coverage": { "type": "string" },
        "synopsis": { "type": "string" },
        "estimated_total_scenes": { "type": "integer" }
      }
    }
  }
}`

// Run 执行 schema 命令。
func (c *SchemaCommand) Run(args []string) error {
	c.flagSet.Parse(args)

	var schemaData map[string]interface{}
	if err := json.Unmarshal([]byte(scriptSchemaJSON), &schemaData); err != nil {
		return fmt.Errorf("内部JSON Schema解析失败: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(schemaData, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON格式化失败: %w", err)
	}

	if c.output != "" {
		return os.WriteFile(c.output, prettyJSON, 0644)
	}

	fmt.Println(string(prettyJSON))
	return nil
}
