# novel2script 剧本 YAML Schema 设计文档

## 概述

novel2script 采用 YAML 作为剧本序列化格式。本文档定义完整的 Schema 结构，并阐述每个关键设计决策背后的原因。

目标用户是中文网文作者，他们需要将小说改编为可编辑、可进一步打磨的剧本初稿。Schema 设计需同时满足以下需求：

1. **结构化**：能被工具自动处理和校验
2. **可读性**：作者能直接阅读和编辑 YAML 文件
3. **完整性**：覆盖从角色到场景到台词的全链路信息
4. **可追溯**：任何剧本片段都能回溯到原文

---

## Schema 结构总览

```
Script（剧本）
├── 文件头（script_title, source_novel, source_author, ...）
├── metadata（元数据：题材、字数、梗概等）
├── characters[]（角色表）
│   ├── 角色基本信息（id, name, aliases, role, ...）
│   └── relationships[]（角色关系）
└── acts[]（幕）
    └── scenes[]（场景）
        ├── setting（环境设置）
        ├── characters_present[]（出场角色 + 状态）
        ├── elements[]（剧本元素：action | dialogue | internal_monologue | narration）
        └── transition（转场）
```

---

## 设计决策详解

### 1. 为什么用 YAML 而非 JSON？

**决策**：选择 YAML 作为主序列化格式。

**原因**：

- **注释支持**：YAML 原生支持 `#` 注释，作者可以在生成的初稿上直接添加批注、修改意见，而不影响格式合法性。JSON 不支持注释，每次编辑都需额外沟通渠道。
- **多行字符串**：剧本中的动作描述（action）经常涉及多段落的环境描写和动作序列。YAML 的 `|`（literal block scalar）让多行文本保持可读格式，JSON 则需要转义换行符 `\n`。
- **非技术人员友好**：剧本编辑者通常不是程序员。YAML 的缩进层级比 JSON 的花括号和引号更直观。缩进天然映射了"幕 → 场景 → 元素"的层级关系。
- **生态成熟**：`gopkg.in/yaml.v3` 提供稳定的 Go 实现，`ruamel.yaml`（Python）支持保留注释的往返编辑。

**代价**：YAML 的隐式类型转换（如 `yes` → `true`、`001` → `1`）是一个已知陷阱，需在解析时注意所有值采用字符串显式引号包裹。

---

### 2. 为什么冗余存储 `speaker_name` 而非仅用 `speaker_id`？

**决策**：对话元素同时包含 `speaker_id`（角色标识符）和 `speaker_name`（显示名称）。

**原因**：

- **独立场景编辑**：在实际工作流中，单个场景文件经常被分拆给不同的编剧或导演独立编辑。如果只有 `speaker_id: "char_hanli"`，阅读者必须翻回文件头部的角色表才能知道是谁在说话。冗余存储让每个场景单元自包含。
- **别名场景**：网文角色常有多个称呼（本名、道号、化名）。`speaker_name` 可以是当前上下文中使用的具体称呼（如"厉飞雨"），而 `speaker_id` 始终指向规范角色标识。
- **校验层保证一致性**：`speaker_name` 与 `characters` 表的映射关系由 ValidateScript 交叉校验。冗余不代表不可靠。

**代价**：数据冗余增加了存储大小。但对于剧本文件（通常 < 10MB），这是可以接受的。

---

### 3. 为什么 `internal_monologue` 是独立 Element 类型？

**决策**：将内心独白定义为与 dialogue 同级的独立元素类型，而非 dialogue 的子类型。

**原因**：

- **网文的特殊需求**：中文网文（尤其是仙侠/玄幻类）大量使用内心独白——"心中暗道""暗自想道""心道""寻思"——这是区别于传统文学的重要叙事特征。将内心独白专门识别，是对网文这个领域的针对性设计。
- **改编决策的灵活性**：内心独白在影视化时有多种处理方式：
  - 保留为画外音（VO, voice-over）
  - 转化为表情和肢体动作
  - 转化为与其他角色的对话
  - 直接删除
  独立类型让后期编辑可以快速 grep 所有内心独白并做批量处理决策。
- **与 dialogue 的本质区别**：dialogue 是角色之间的互动，受社会关系、场景氛围约束；internal_monologue 是角色的私人思想，不受这些约束。两者在表演指示（delivery vs visibility）、镜头语言上完全不同。

**具体字段差异**：
- `dialogue` 有 `tone`（语气）、`delivery`（表演指示）、`language_style`（语言风格）
- `internal_monologue` 有 `visibility`（呈现方式："画外音" vs "画面表现"）

---

### 4. 为什么不按章节组织场景？

**决策**：场景按"幕（Act）"组织而非按"章（Chapter）"组织，但保留 `chapter_range` / `chapter_source` 映射字段。

**原因**：

- **戏剧结构与章节结构不同**：小说的章节划分服务于阅读节奏（每章结尾留悬念），而剧本的幕/场景划分服务于戏剧性节拍（dramatic beats）。一章小说可能包含 3 个不同地点的场景，一个高潮场景可能跨越 2 章。
- **LLM 的重新识别能力**：LLM 擅长理解叙事逻辑，能够跨章节识别"这是一个完整的冲突场景"或"这里视角切换了"。强制按章节划分会浪费 LLM 的这种能力。
- **`chapter_range` 保留溯源**：虽然不按章节组织，但 `act.chapter_range` 和 `scene.chapter_source` 保留了与原文的映射。作者可以快速定位到原文验证改编准确性。

---

### 5. 为什么包含 `source_paragraph`？

**决策**：每个 ScriptElement 携带可选的 `source_paragraph` 字段，指向原文段落号。

**原因**：

- **溯源校验**：AI 可能遗漏或误读重要情节。`source_paragraph` 使得对照原文做完整性检查成为可能。工具可以自动检测"原文第 45 段的高潮对话是否被遗漏在剧本中"。
- **迭代改编**：剧本初稿生成后，作者可能会大幅修改场景。`source_paragraph` 保留了修改前的原点，可以在后续迭代中做一致性 diff。
- **对账审计**：在团队协作场景中，制片人或导演可以要求"这段台词有原文依据吗？"——`source_paragraph` 提供了可验证的答案。

**代价**：LLM 的段落号定位可能不够精确（±2 段的误差是常见的）。因此该字段为可选字段，不作为校验的强制要求。

---

### 6. 场景元素（ScriptElement）的类型分层

**决策**：通过 `type` 鉴别字段实现联合类型，而非用嵌套结构或 interface。

| type | 含义 | 特有字段 |
|------|------|---------|
| `action` | 动作/画面描述 | `visual_cue` |
| `dialogue` | 角色对话 | `speaker_id`, `speaker_name`, `tone`, `delivery`, `language_style` |
| `internal_monologue` | 内心独白 | `speaker_id`, `speaker_name`, `visibility` |
| `narration` | 旁白 | — |
| `title_card` | 字幕/标题卡 | — |

**原因**：

- **扁平化优于嵌套**：Go struct 的扁平 tag 比深层嵌套的 YAML 更易读。用 `type` 字段作为鉴别符，所有字段平铺在同一层级，人工编辑时不需要理解复杂的嵌套规则。
- **明确的字段关联**：`type: "dialogue"` 时 `speaker_id` 才有意义，`type: "action"` 时 `visual_cue` 才有意义。校验层可以明确检查这些关联，防止"action 元素误填了 speaker_id"的情况。
- **可扩展**：未来可以增加新类型（如 `song` 歌词、`montage_sequence` 蒙太奇序列），只需在 `ElementType` 枚举中添加新值，不影响现有类型。

---

### 7. 角色表的规范化设计

**决策**：角色表采用独立的 `characters` 顶层数组，角色通过 `id` 在场景中被引用。

**关键字段说明**：

| 字段 | 类型 | 设计原因 |
|------|------|---------|
| `role` | 枚举 | 标准化的角色定位（protagonist/antagonist/supporting 等）便于工具统计角色功能分布 |
| `aliases` | 数组 | 网文角色常有多个称呼（本名、道号、绰号、化名），合并在同一角色下避免 LLM 重复提取 |
| `archetype` | 字符串 | 中文网文的角色原型（"普通人崛起""重生复仇""天才陨落"等），帮助 LLM 理解角色动机模式 |
| `character_arc` | 字符串 | 角色的完整成长弧线，为分幕编排提供依据 |
| `first_appearance_chapter` | 整数 | 帮助确定角色出场顺序和分幕中的引入时机 |
| `importance_rank` | 整数 | 排序依据，确保主角排在角色表最前面 |
| `relationships` | 数组 | 角色关系网，使用中文网文常用词汇（"道侣""宿敌""护道者"等），贴近目标用户的认知 |

---

### 8. 转场处理

**决策**：转场（Transition）是场景级别的可选字段，而非元素级别的。

**原因**：

- 转场是两个场景之间的桥接，语义上属于"上一个场景如何过渡到下一个场景"，而非场景内部的内容。
- 放在场景末尾更符合剧本阅读习惯——读完一个场景，立即知道如何过渡到下一个。
- `next_scene_hint` 提供对下一个场景的预告，帮助导演做镜头规划。

---

## 版本策略

Schema 版本号采用语义化版本 `MAJOR.MINOR.PATCH`：

- **MAJOR**：不兼容的字段重命名或类型变更
- **MINOR**：新增可选字段或新 ElementType
- **PATCH**：文档修正、描述文案调整

当前版本：**1.0.0**

`script.version` 记录生成时使用的 Schema 版本，确保未来即使 Schema 升级，旧文件仍可被正确解析（通过版本号路由到对应的校验规则）。

---

## 与行业标准的映射

| novel2script | Final Draft (.fdx) | Fountain (.fountain) |
|-------------|--------------------| ---------------------|
| Script | — | — |
| Act | — | `#` (section) |
| Scene | Scene | `##` (scene heading) |
| action | Action | 纯文本行 |
| dialogue | Character + Dialogue | `@角色名` + 对话文本 |
| internal_monologue | Character + Parenthetical (V.O.) | — |
| transition | Transition | `>` 前缀 |

YAML 格式可以无损转换为 Fountain 格式（Markdown 变体），进而被大多数剧本编辑软件导入。

---

## 完整 YAML 示例

请参见 `examples/output/sample_script.yaml`（将在后续 PR 中提交）。
