# novel2script

AI 辅助小说转剧本工具 —— 将中文小说自动转换为结构化 YAML 剧本初稿。

## 特性

- **自动改编** — 将 3 章以上小说文本转换为完整的剧本（角色表 + 分幕场景 + 元素序列）
- **智能分析** — 多级渐进式角色提取、事件驱动场景分割、叙事→剧本元素转换
- **并行处理** — goroutine 并发 LLM 调用，支持大长篇（1000+ 章）
- **断点续传** — SQLite 缓存中间结果，中断后可恢复
- **多格式输出** — YAML（可编辑剧本）和 Markdown（可读草稿）
- **多 LLM 支持** — Anthropic Claude 和 OpenAI 兼容接口

## 安装

```bash
go install github.com/JunLang-7/novel2script@latest
```

或从源码编译：

```bash
git clone https://github.com/JunLang-7/novel2script.git
cd novel2script
go build -o novel2script .
```

## 快速开始


```bash
# Anthropic 兼容接口(默认)
export NOVEL2SCRIPT_API_KEY=sk-ant-xxx

# OpenAI 兼容接口
export NOVEL2SCRIPT_API_KEY=sk-xxx
export NOVEL2SCRIPT_PROVIDER=openai
export NOVEL2SCRIPT_BASE_URL=https://api.openai.com/chat/completions
export NOVEL2SCRIPT_MODEL=xxx

# 预览成本（不调用 LLM）
novel2script convert 小说.txt --dry-run

# 输出 Markdown 草稿
novel2script convert 小说.txt -f md -o draft.md

# 仅分析角色和场景
novel2script analyze 小说.txt -o analysis.json

# 验证生成的剧本
novel2script validate script.yaml
```

也可在项目根目录创建 `.env` 文件（参考 `.env.example`）：

```bash
NOVEL2SCRIPT_API_KEY=sk-ant-xxx
NOVEL2SCRIPT_PROVIDER=anthropic
NOVEL2SCRIPT_MODEL=claude-sonnet-4-20250514
NOVEL2SCRIPT_PARALLEL=5
```

## 命令

| 命令 | 说明 |
|------|------|
| `convert` | 将小说转换为完整剧本（主命令） |
| `analyze` | 仅分析角色和场景，不做剧本转换 |
| `validate` | 验证 YAML 剧本文件格式 |
| `schema` | 输出 JSON Schema 定义 |

### convert 选项

```
novel2script convert <输入文件> [选项]

选项:
  -o, --output     输出文件路径 (默认: script.yaml)
  -f, --format     输出格式: yaml | md (默认: yaml)
  -s, --start      起始章节号（从 1 开始）
  -e, --end        结束章节号
  -m, --model      LLM 模型名称
  -p, --parallel   并行 LLM 调用数 (默认: 5)
  -n, --dry-run    仅分析不转换：输出分块统计和预估成本
  -v, --verbose    详细日志输出
  -r, --resume     从上次中断处继续
```

### analyze 选项

```
novel2script analyze <输入文件> [选项]

选项:
  -o, --output     输出文件路径 (默认: analysis.json)
  -s, --start      起始章节号（从 1 开始）
  -e, --end        结束章节号
  -m, --model      LLM 模型名称
  -p, --parallel   并行 LLM 调用数 (默认: 5)
  -r, --resume     从上次中断处继续
  -v, --verbose    详细日志输出
```

### 断点续传

使用 `--resume` 启用 SQLite 缓存，中断后可跳过已完成的工作：

```bash
# 首次运行（失败中断后重新执行，跳过已完成的 chunk）
novel2script convert 凡人修仙传.txt --resume -v -o script.yaml
```

缓存内容：
- **角色表** — 按小说内容哈希缓存，同一小说不同文件可复用
- **场景分割结果** — 按 chunk 缓存，已完成的 chunk 直接恢复

缓存目录默认为 `~/.novel2script/cache`，可通过 `NOVEL2SCRIPT_CACHE_DIR` 环境变量修改。

## 配置

通过环境变量配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NOVEL2SCRIPT_API_KEY` | LLM API 密钥（必填） | - |
| `NOVEL2SCRIPT_MODEL` | 模型名称 | `claude-sonnet-4-20250514` |
| `NOVEL2SCRIPT_PROVIDER` | LLM 提供商 (`anthropic` 或 `openai`) | `anthropic` |
| `NOVEL2SCRIPT_BASE_URL` | API 基础 URL | 提供商默认地址 |
| `NOVEL2SCRIPT_PARALLEL` | 并行 LLM 调用数 | `5` |
| `NOVEL2SCRIPT_CACHE_DIR` | 缓存目录 | `~/.novel2script/cache` |

## 输入格式

支持的输入格式：

- **纯文本** (`.txt`) — UTF-8 或 GBK 编码，自动检测
- **Markdown** (`.md`) — 自动识别标题作为章节标记

章节标记识别模式：
- `第X章` / `第X回` / `第X节`
- `Chapter X` / `CHAPTER X`
- Markdown `#` 标题
- 卷标记 `卷一` / `第一部`

## 输出格式

### YAML（默认）

结构化剧本，包含三个区块：

```yaml
# 文件头 - 元数据
script_title: "凡人修仙传·剧本改编"
source_novel: "凡人修仙传"
source_author: "忘语"
...
metadata:
  genre: ["仙侠", "玄幻"]
  total_novel_chapters: 5
  synopsis: "山村少年韩立偶然踏入修仙之路..."
...

# 角色表 - 完整角色信息
characters:
  - id: char_hanli
    name: 韩立
    role: protagonist
    traits: [谨慎, 坚韧]
    relationships:
      - target_id: char_mojuren
        type: 师徒
...

# 剧本正文 - 分幕组织
acts:
  - id: act_1
    title: "第一幕"
    scenes:
      - id: scene_1_1
        title: "七玄门·神手谷"
        setting:
          location: "神手谷"
          time_of_day: "黄昏"
        elements:
          - type: action
            content: "韩立盘膝坐在蒲团上，双手结印。"
          - type: dialogue
            speaker_name: "韩立"
            content: "这长春功果然难练。"
          - type: internal_monologue
            speaker_name: "韩立"
            content: "若不能突破，七日后的试炼恐怕..."
```

### Markdown

人类可读的剧本草稿，适合审阅和分享。

## 处理管道

```
输入文件(.txt/.md)
    │
    ▼
Step 1  章节检测 ─── 正则匹配章边界，按 ~15000 tokens 分块
    │
    ▼
Step 2  角色提取 ─── 多级渐进式 + 跨块去重合并（--resume 缓存角色表）
    │   ├─ Pass 1: 前 3 章精细提取核心角色
    │   └─ Pass 2: 后续每块批量提取新角色
    │
    ├─ fillTargetIDs ─── 自动补全关系中的 target_id（name/alias 索引匹配）
    │
    ▼
Step 2.5 元数据提取 ─── 推断 source_author、genre、synopsis
    │
    ▼
Step 3  场景分割 ─── 地点/时间/视角/事件驱动，注入已知角色 ID
    │         （--resume 跳过已完成的 chunk，从 SQLite 缓存恢复）
    │
    ▼
Step 4  剧本转换 ─── 叙事→动作，对话→标注对白，场景聚焦约束
    │
    ▼
Step 5  分幕构建 + Step 6 组装 Script → YAML/Markdown 输出
```

## 项目结构

```
novel2script/
├── main.go                      # CLI 入口
├── cmd/                         # 子命令
│   ├── convert.go               # convert（主命令）
│   ├── analyze.go               # analyze（仅分析）
│   ├── validate.go              # validate（验证 YAML）
│   └── schema.go                # schema（导出 JSON Schema）
├── internal/
│   ├── models/                  # 数据模型（Script, Scene, Character）
│   ├── pipeline/                # 管道编排、分块、合并、角色提取、场景分割
│   ├── formatters/              # YAML / Markdown 输出
│   ├── llm/                     # LLM 客户端、提示词、token 估算
│   ├── text/                    # 章节检测、编码识别
│   ├── storage/                 # SQLite 缓存（断点续传）
│   └── config/                  # 环境变量配置
├── schema/                      # JSON Schema + 设计文档
├── tests/fixtures/              # 测试样本
└── examples/output/             # 示例输出
```

## License

MIT
