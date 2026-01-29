export type DraftVersion = {
  id: number;
  app_version_name?: string | null;
  location_name?: string | null;
  feishu_field_names?: string | null;
  feishu_field_list?: string[] | null;
  ai_modal?: string | null;
  status?: number | null;
  draft_status?: string | null;
  submit_version?: number | null;
  last_submit_by?: number | null;
  last_submit_at?: string | null;
  confirmed_by?: number | null;
  confirmed_at?: string | null;
};

export const baseFeishuFields = [
  "编号顺序",
  "场景",
  "身份",
  "提示词",
  "人像位置",
  "输入图1",
  "输入图2",
  "输入图3",
  "审核情况",
  "作废"
];

export const optionalFeishuFields = ["SD模式", "服饰偏好", "拍摄偏好"];

export const buildDefaultFeishuFields = (aiModal?: string | null) => {
  const fields = [...baseFeishuFields];
  if (!aiModal || aiModal.toUpperCase() === "SD") {
    fields.push("SD模式");
  }
  return fields;
};

export const formatDate = (value?: string | null) => {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString("zh-CN");
};
