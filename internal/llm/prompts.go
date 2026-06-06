package llm

// 所有 prompt 模板均为中文，采用五段式结构：
// 系统角色 → 任务指令 → 格式约束 → 示例 → 输入

// SystemPrompt 是所有 LLM 调用的通用系统角色设定。
const SystemPrompt = `你是一位专业的剧本改编专家，擅长将中文网络小说改编为影视剧本。
你的核心能力：
1. 理解中文网文的叙事结构和节奏（起承转合、高潮铺垫）
2. 区分并正确处理：叙事、对话、内心独白、侧面描写
3. 识别网文中常见的套路和模板（升级、打脸、奇遇等）
4. 将文学化的描述转化为可视化的动作和镜头语言`

// CharacterExtractionPrompt 角色提取提示词模板。
// 占位符: {text}
const CharacterExtractionPrompt = `请分析以下小说文本，提取所有重要角色信息。

特别注意：
- 网文中常有角色"马甲"（化名、假名），请识别并归入同一角色
- 角色可能有多个称呼（如"韩立"→"韩师叔"→"厉飞雨"），全部收集到aliases中
- 关系类型请使用中文网文常用词汇（如"道侣""宿敌""护道者"）
- relationships 中的 target_id 必须填写为目标角色的 id（如 "char_hanli"），不要留空
- relationships 中的 description 必须包含目标角色的名字（如"韩父是韩立的父亲"），不要用代词替代
- role 字段取值为: protagonist | deuteragonist | antagonist | supporting | love_interest | cameo

<<输入文本>>
{text}

请以JSON数组格式返回角色列表，每个角色包含以下字段：
id (格式为 "char_<拼音名>"), name, aliases (数组), role, importance_rank (整数), description, traits (数组), character_arc, first_appearance_chapter (整数), relationships (数组，每项包含 target_id, type, description（描述中必须提及对方角色名，如"张三的父亲"而非"他的父亲"）)

只返回JSON数组，不要其他说明。`

// SceneSegmentationPrompt 场景分割提示词模板。
// 占位符: {text}
const SceneSegmentationPrompt = `请将以下小说章节分割为独立的戏剧场景。

场景分割规则：
1. 地点变化 → 新场景
2. 时间跳跃 → 新场景（如"第二天""三日后"）
3. 视角切换 → 新场景（如从主角切到反派POV）
4. 独立事件 → 新场景（如开始战斗、进入秘境）
5. 连续在同一地点的对话/动作，即使跨越多个段落，仍归为同一场景

	<<已知角色及其ID（请使用这些ID标注出场角色）>>
	{character_context}

<<原文>>
{text}

请以JSON数组格式返回场景列表，每个场景包含以下字段：
id (格式为 "scene_<序号>"), title (场景标题), sequence (整数), location (地点), location_type (地点类型，如"门派""洞府""城镇""荒野""秘境"), time_of_day (时间，如"清晨""黄昏""夜晚"), atmosphere (氛围描述), summary (场景概要), chapter_source (整数，对应原文章节号), characters_present (数组，每个元素包含 id（必须使用上方已知角色列表中的 id，格式如 char_hanli）和 state), mood (情绪基调)

只返回JSON数组，不要其他说明。`

// ScriptConversionPrompt 剧本转换提示词模板。
// 占位符: {character_context}, {scene_title}, {scene_summary}, {location}, {time}, {characters_present}, {text}
const ScriptConversionPrompt = `将以下小说场景转换为标准剧本格式。

转换规则：
1. 叙事性文字 → action类型（用现在时，描写可视化的动作和画面）
2. 对话 → dialogue类型，标注说话者(tone:语气)。对话中的"笑道""怒道""冷声道"等 → 标注为tone
3. 内心独白 → internal_monologue类型（网文特有，保留为画外音）。识别触发词："心中暗道""想道""心道""暗想""寻思"
4. 旁白/背景说明 → narration类型（保留对世界观构建重要的部分）
5. 战斗场景 → 切分为短促的动作序列（每个动作1-2句）
6. 环境描写 → 融入动作描写中，或单独作为action段落
7. "只见""但见""却见"等引导的描写 → 转换为可视化动作

<<角色表>>
{character_context}

<<场景信息>>
场景：{scene_title}
	概要：{scene_summary}
地点：{location}
时间：{time}
出现角色：{characters_present}


	重要：只转换上述场景对应的原文部分，不要转换其他场景的内容。如果原文包含多个场景的文本，请严格只提取与「{scene_title}」相关的段落进行转换。
<<原文>>
{text}

请返回JSON格式的剧本元素数组，每个元素包含以下字段：
id (格式为 "elem_<序号>"), type (取值为 action | dialogue | internal_monologue | narration), content (元素内容), speaker_id (如果是dialogue或internal_monologue), speaker_name (如果是dialogue或internal_monologue), tone (如果是dialogue，标注语气), delivery (表演指示，可选), visual_cue (镜头提示，可选)

只返回JSON数组，不要其他说明。`

// SynopsisPrompt 梗概生成提示词模板。
// 占位符: {text}
const SynopsisPrompt = `请为以下小说文本撰写一段简洁的故事梗概（200字以内），用于剧本元数据中的剧情简介。

<<原文>>
{text}

只返回故事梗概文本，不要其他说明。`

// PlotAnalysisPrompt 情节弧线分析提示词模板。
// 占位符: {text}
const PlotAnalysisPrompt = `请分析以下小说文本的情节结构，识别主要的情节弧线和关键转折点。

特别注意中文网文的典型结构：
- 开篇引入（世界观、主角初始状态）
- 成长升级（修炼突破、获得机缘）
- 冲突激化（对立势力、个人恩怨）
- 高潮对决（决战、逆转）
- 收尾铺垫（为下一阶段埋下伏笔）

<<原文>>
{text}

请以JSON格式返回，包含以下字段：
acts (数组，每项包含: title, summary, chapter_start, chapter_end, dramatic_function)

只返回JSON，不要其他说明。`
