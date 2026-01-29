# 任务清单: tts-presets

目录: `helloagents/plan/202601282017_tts-presets/`

---

## 任务状态符号说明

| 符号 | 状态 | 说明 |
|------|------|------|
| `[ ]` | pending | 待执行 |
| `[√]` | completed | 已完成 |
| `[X]` | failed | 执行失败 |
| `[-]` | skipped | 已跳过 |
| `[?]` | uncertain | 待确认 |

---

## 执行状态
```yaml
总任务: 7
已完成: 7
完成率: 100%
```

---

## 任务列表

### 1. 数据库与默认预设

- [√] 1.1 新增 `server/migrations/005_app_db_tts_presets.sql` 创建预设表并写入默认预设
  - 验证: migrations 执行无报错，表结构与默认值正确

### 2. 后端接口

- [√] 2.1 新增 TTS 预设 CRUD Handler（列表/新增/更新/删除）
  - 文件: `server/internal/http/handlers/tts_presets.go`
  - 验证: 接口返回结构正确，默认预设优先

- [√] 2.2 路由注册并加管理员权限
  - 文件: `server/internal/http/router.go`
  - 依赖: 2.1

- [√] 2.3 增加参数范围校验与默认值填充辅助函数
  - 文件: `server/internal/http/handlers/tts_presets.go`
  - 验证: 非法范围返回错误

### 3. 前端预设配置

- [√] 3.1 新增 TTS 预设管理面板（媒体规则页新 Tab）
  - 文件: `web/src/pages/media/TtsPresetPanel.tsx`, `web/src/pages/MediaRules.tsx`
  - 验证: 可新增/编辑/停用/删除预设

### 4. TTS 生成参数

- [√] 4.1 扩展 TTS 生成方法与参数类型
  - 文件: `web/src/pages/content/utils.ts`, `web/src/pages/ContentEntry.tsx`
  - 验证: 请求体携带参数且不影响旧流程

- [√] 4.2 TTS 生成面板支持预设选择与滑动条参数
  - 文件: `web/src/pages/content/TTSInlinePanel.tsx`（含场景/配置/偏好/页面配置调用）
  - 依赖: 4.1

### 5. 测试

- [√] 5.1 新增 TTS 预设范围校验与默认值测试
  - 文件: `tests/handlers/tts_presets_test.go`

---

## 执行备注

> 执行过程中的重要记录

| 任务 | 状态 | 备注 |
|------|------|------|
