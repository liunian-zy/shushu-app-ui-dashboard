export type DraftVersion = {
  id: number;
  app_version_name?: string | null;
  location_name?: string | null;
  ai_modal?: string | null;
};

export type BannerItem = {
  id: number;
  title?: string | null;
  image?: string | null;
  image_url?: string | null;
  sort?: number | null;
  is_active?: number | null;
  type?: number | null;
  app_version_name?: string | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export type IdentityItem = {
  id: number;
  name?: string | null;
  image?: string | null;
  image_url?: string | null;
  sort?: number | null;
  status?: number | null;
  app_version_name?: string | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export type SceneItem = {
  id: number;
  name?: string | null;
  image?: string | null;
  image_url?: string | null;
  desc?: string | null;
  music?: string | null;
  music_url?: string | null;
  sort?: number | null;
  status?: number | null;
  app_version_name?: string | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export type ClothesItem = {
  id: number;
  name?: string | null;
  image?: string | null;
  image_url?: string | null;
  sort?: number | null;
  status?: number | null;
  music?: string | null;
  music_url?: string | null;
  desc?: string | null;
  music_text?: string | null;
  app_version_name?: string | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export type PhotoHobbyItem = {
  id: number;
  name?: string | null;
  image?: string | null;
  image_url?: string | null;
  sort?: number | null;
  status?: number | null;
  music?: string | null;
  music_url?: string | null;
  desc?: string | null;
  music_text?: string | null;
  app_version_name?: string | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export type ExtraStepItem = {
  id: number;
  app_version_name_id?: number | null;
  step_index?: number | null;
  field_name?: string | null;
  label?: string | null;
  music?: string | null;
  music_url?: string | null;
  music_text?: string | null;
  status?: number | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export type AppUIFields = {
  id: number;
  app_version_name_id?: number | null;
  home_title_left?: string | null;
  home_title_right?: string | null;
  home_subtitle?: string | null;
  start_experience?: string | null;
  step1_music?: string | null;
  step1_music_url?: string | null;
  step1_music_text?: string | null;
  step1_title?: string | null;
  step2_music?: string | null;
  step2_music_url?: string | null;
  step2_music_text?: string | null;
  step2_title?: string | null;
  status?: number | null;
  print_wait?: string | null;
  print_wait_url?: string | null;
  submit_status?: string | null;
  last_submit_at?: string | null;
};

export const bannerTypeOptions = [
  { value: 1, label: "左上" },
  { value: 2, label: "左下" },
  { value: 3, label: "右侧" }
];

export const identityNameOptions = ["男士", "女士", "男孩", "女孩"];

export const statusOptions = [
  { value: 1, label: "启用", color: "green" },
  { value: 0, label: "停用", color: "default" }
];

export const submitStatusLabels: Record<string, { label: string; color: string }> = {
  submitted: { label: "已提交", color: "blue" },
  pending_confirm: { label: "待确认", color: "gold" },
  confirmed: { label: "已确认", color: "green" }
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
