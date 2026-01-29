export type DraftVersion = {
  id: number;
  app_version_name?: string | null;
  location_name?: string | null;
  ai_modal?: string | null;
  draft_status?: string | null;
  submit_version?: number | null;
};

export type TaskItem = {
  id: number;
  draft_version_id: number;
  module_key?: string | null;
  title?: string | null;
  description?: string | null;
  status?: string | null;
  assigned_to?: number | null;
  allow_assist?: number | null;
  priority?: number | null;
  created_at?: string | null;
  updated_at?: string | null;
};

export type UserItem = {
  id: number;
  username?: string | null;
  display_name?: string | null;
  role?: string | null;
};

export type TaskAction = {
  id: number;
  action?: string | null;
  actor_id?: number | null;
  actor_name?: string | null;
  actor_username?: string | null;
  detail?: unknown;
  created_at?: string | null;
};

export const statusOptions = [
  { value: "open", label: "待开始", color: "default" },
  { value: "in_progress", label: "进行中", color: "blue" },
  { value: "submitted", label: "已提交", color: "purple" },
  { value: "pending_confirm", label: "待确认", color: "gold" },
  { value: "confirmed", label: "已确认", color: "green" },
  { value: "completed", label: "已完成", color: "green" },
  { value: "closed", label: "已关闭", color: "default" }
];

export const moduleOptions = [
  { value: "banners", label: "首页轮播图" },
  { value: "identities", label: "身份信息" },
  { value: "scenes", label: "场景信息" },
  { value: "app_ui_fields", label: "页面配置" },
  { value: "config_extra_steps", label: "额外配置" },
  { value: "clothes_categories", label: "服饰偏好" },
  { value: "photo_hobbies", label: "拍摄偏好" },
  { value: "print_wait", label: "打印中视频" }
];

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

export const resolveStatusMeta = (value?: string | null) =>
  statusOptions.find((item) => item.value === value) ?? {
    value: value ?? "-",
    label: value ?? "-",
    color: "default"
  };

export const resolveModuleLabel = (value?: string | null) => {
  if (!value) {
    return "-";
  }
  const found = moduleOptions.find((item) => item.value === value);
  return found?.label ?? value;
};

export const resolveActionLabel = (value?: string | null) => {
  switch (value) {
    case "create":
      return "创建任务";
    case "assign":
      return "任务指派";
    case "assist":
      return "协作提交";
    case "status_change":
      return "状态变更";
    case "allow_assist":
      return "协作开关";
    case "complete_upload":
      return "完成上传";
    default:
      return value ?? "-";
  }
};
