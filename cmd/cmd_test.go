package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertCommand_NameAndUsage(t *testing.T) {
	c := NewConvertCommand()
	if c.Name() != "convert" {
		t.Errorf("Name() = %q, want %q", c.Name(), "convert")
	}
	usage := c.Usage()
	if !strings.Contains(usage, "convert") {
		t.Error("Usage() should mention convert")
	}
	if !strings.Contains(usage, "--dry-run") {
		t.Error("Usage() should mention --dry-run")
	}
}

func TestConvertCommand_MissingInput(t *testing.T) {
	c := NewConvertCommand()
	err := c.Run([]string{})
	if err == nil {
		t.Error("expected error for missing input file")
	}
}

func TestConvertCommand_MissingAPIKey(t *testing.T) {
	// Create a minimal temp file
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "test.txt")
	os.WriteFile(inputPath, []byte("第一章 测试内容\n这是一段测试文本。"), 0644)

	// Clear API key to trigger the missing key error
	oldKey := os.Getenv("NOVEL2SCRIPT_API_KEY")
	os.Unsetenv("NOVEL2SCRIPT_API_KEY")
	defer func() {
		if oldKey != "" {
			os.Setenv("NOVEL2SCRIPT_API_KEY", oldKey)
		}
	}()

	c := NewConvertCommand()
	err := c.Run([]string{inputPath})
	if err == nil {
		t.Error("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "NOVEL2SCRIPT_API_KEY") {
		t.Errorf("error should mention NOVEL2SCRIPT_API_KEY: %v", err)
	}
}

func TestConvertCommand_ChapterRangeOutOfBounds(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "test.txt")
	content := "第一章 初入修仙\n韩立站在神手谷中。\n\n" +
		"第二章 药园初见\n墨大夫打量着韩立。\n"
	os.WriteFile(inputPath, []byte(content), 0644)

	oldKey := os.Getenv("NOVEL2SCRIPT_API_KEY")
	os.Setenv("NOVEL2SCRIPT_API_KEY", "test-key")
	defer func() {
		if oldKey != "" {
			os.Setenv("NOVEL2SCRIPT_API_KEY", oldKey)
		} else {
			os.Unsetenv("NOVEL2SCRIPT_API_KEY")
		}
	}()

	// --start beyond total chapters
	c := NewConvertCommand()
	err := c.Run([]string{"--start", "100", inputPath})
	if err == nil {
		t.Error("expected error for start chapter beyond total")
	}
}

func TestConvertCommand_DryRun(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "test.txt")
	content := "第一章 初入修仙\n" +
		"韩立站在神手谷中，望着远处的云雾，心中感慨万千。\n" +
		"这是他第一次独自面对修仙世界的危险。他握紧手中的长剑，深吸一口气。\n\n" +
		"第二章 药园初见\n" +
		"三个月后，韩立在药园中遇见了改变他一生的人——墨大夫。\n" +
		"墨大夫看上去四十余岁，一身青衫，面容清瘦但精神矍铄。\n" +
		"「你就是新来的药童？」墨大夫打量着韩立。\n" +
		"「是的，弟子韩立，见过墨大夫。」韩立恭恭敬敬地行了一礼。\n\n" +
		"第三章 长春功\n" +
		"夜深人静，韩立盘膝坐在床上，按照墨大夫所传的口诀运转长春功。\n" +
		"一缕温热的气流在丹田中缓缓生起，他心中暗喜：成了！\n" +
		"然而他不知道的是，墨大夫正在暗处观察着他的一举一动。\n\n" +
		"第四章 灵根测试\n" +
		"翌日清晨，墨大夫将韩立叫到跟前，取出一块晶莹剔透的灵石。\n" +
		"「握住它，让灵力流转。」墨大夫淡淡地说道。\n" +
		"韩立依言而行，灵石瞬间亮起了耀眼的青色光芒。\n\n" +
		"第五章 离别\n" +
		"「你的灵根很不错。」墨大夫满意地点点头，「不过修行之路漫长而艰险，」\n" +
		"他的目光变得深邃，「你可准备好了？」\n" +
		"韩立郑重地跪下行礼：「弟子决不辜负师父期望。」\n"
	os.WriteFile(inputPath, []byte(content), 0644)

	oldKey := os.Getenv("NOVEL2SCRIPT_API_KEY")
	os.Setenv("NOVEL2SCRIPT_API_KEY", "test-key")
	defer func() {
		if oldKey != "" {
			os.Setenv("NOVEL2SCRIPT_API_KEY", oldKey)
		} else {
			os.Unsetenv("NOVEL2SCRIPT_API_KEY")
		}
	}()

	c := NewConvertCommand()
	err := c.Run([]string{"--dry-run", inputPath})
	if err != nil {
		t.Fatalf("dry-run should not fail: %v", err)
	}
}

func TestAnalyzeCommand_NameAndUsage(t *testing.T) {
	c := NewAnalyzeCommand()
	if c.Name() != "analyze" {
		t.Errorf("Name() = %q, want %q", c.Name(), "analyze")
	}
	usage := c.Usage()
	if !strings.Contains(usage, "analyze") {
		t.Error("Usage() should mention analyze")
	}
}

func TestAnalyzeCommand_MissingInput(t *testing.T) {
	c := NewAnalyzeCommand()
	err := c.Run([]string{})
	if err == nil {
		t.Error("expected error for missing input file")
	}
}

func TestValidateCommand_NameAndUsage(t *testing.T) {
	c := NewValidateCommand()
	if c.Name() != "validate" {
		t.Errorf("Name() = %q, want %q", c.Name(), "validate")
	}
	usage := c.Usage()
	if !strings.Contains(usage, "validate") {
		t.Error("Usage() should mention validate")
	}
}

func TestValidateCommand_MissingInput(t *testing.T) {
	c := NewValidateCommand()
	err := c.Run([]string{})
	if err == nil {
		t.Error("expected error for missing YAML file")
	}
}

func TestValidateCommand_FileNotFound(t *testing.T) {
	c := NewValidateCommand()
	err := c.Run([]string{"/nonexistent/path/script.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidateCommand_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "test.yaml")
	validYAML := `script_title: "测试剧本"
source_novel: "测试小说"
version: "1.0"
characters:
  - id: "c1"
    name: "韩立"
    role: "protagonist"
    importance_rank: 1
acts:
  - id: "a1"
    title: "第一幕"
    scenes:
      - id: "s1"
        title: "测试场景"
        type: "scene"
        sequence: 1
        setting:
          location: "测试地点"
        elements:
          - id: "e1"
            type: "action"
            content: "测试动作"
        characters_present:
          - id: "c1"
            state: "standing"
`
	os.WriteFile(yamlPath, []byte(validYAML), 0644)

	c := NewValidateCommand()
	err := c.Run([]string{yamlPath})
	if err != nil {
		t.Fatalf("validate should succeed for valid YAML: %v", err)
	}
}

func TestSchemaCommand_NameAndUsage(t *testing.T) {
	c := NewSchemaCommand()
	if c.Name() != "schema" {
		t.Errorf("Name() = %q, want %q", c.Name(), "schema")
	}
	usage := c.Usage()
	if !strings.Contains(usage, "schema") {
		t.Error("Usage() should mention schema")
	}
}

func TestSchemaCommand_Stdout(t *testing.T) {
	c := NewSchemaCommand()
	err := c.Run([]string{})
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}
}

func TestSchemaCommand_OutputToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "schema.json")

	c := NewSchemaCommand()
	err := c.Run([]string{"-o", outPath})
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(data), "script_title") {
		t.Error("output should contain script_title")
	}
}
