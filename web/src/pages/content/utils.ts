export type RequestFn = <T>(path: string, options?: RequestInit) => Promise<T>;

export type UploadResult = {
  path: string;
  url?: string;
};

export type UploadFn = (
  file: File,
  moduleKey: string,
  draftVersionId: number
) => Promise<UploadResult>;

export type TTSResult = {
  audio_path: string;
  audio_url?: string;
};

export type TTSPreset = {
  id: number;
  name?: string | null;
  voice_id?: string | null;
  volume?: number | null;
  speed?: number | null;
  pitch?: number | null;
  stability?: number | null;
  similarity?: number | null;
  exaggeration?: number | null;
  status?: number | null;
  is_default?: number | null;
};

export type TTSOptions = {
  preset_id?: number;
  voice_id?: string;
  volume?: number;
  speed?: number;
  pitch?: number;
  stability?: number;
  similarity?: number;
  exaggeration?: number;
};

export type TTSFn = (
  text: string,
  moduleKey: string,
  draftVersionId: number,
  options?: TTSOptions
) => Promise<TTSResult>;

export const generateTTSBatch = async (
  generate: TTSFn,
  text: string,
  moduleKey: string,
  draftVersionId: number,
  count: number,
  options?: TTSOptions
) => {
  const safeCount = Math.max(1, Math.min(6, count || 1));
  const tasks = Array.from({ length: safeCount }, () => generate(text, moduleKey, draftVersionId, options));
  return Promise.all(tasks);
};

export type Notify = {
  success: (message: string) => void;
  error: (message: string) => void;
  warning: (message: string) => void;
};

export const buildLocalDraftKey = (moduleKey: string, draftVersionId: number, itemKey?: number | string) => {
  const safeModule = moduleKey || "default";
  const safeVersion = Number.isFinite(draftVersionId) ? draftVersionId : 0;
  const keyPart = itemKey === undefined || itemKey === null || itemKey === "" ? "new" : String(itemKey);
  return `shushu_local_draft:${safeModule}:${safeVersion}:${keyPart}`;
};

export const saveLocalDraft = (key: string, payload: Record<string, unknown>) => {
  if (!key) {
    return;
  }
  localStorage.setItem(key, JSON.stringify(payload));
};

export const loadLocalDraft = (key: string) => {
  if (!key) {
    return null;
  }
  const raw = localStorage.getItem(key);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return null;
  }
};

export const buildLocalPreviewUrl = (path?: string | null) => {
  if (!path) {
    return null;
  }
  if (path.startsWith("local://")) {
    const trimmed = path.replace(/^local:\/\//, "").replace(/^\/+/, "");
    return `/api/local-files/${trimmed}`;
  }
  return path;
};

export const sanitizeSubmissionPayload = (payload: Record<string, unknown>) => {
  const cleaned: Record<string, unknown> = {};
  Object.entries(payload).forEach(([key, value]) => {
    if (key.endsWith("_url")) {
      return;
    }
    if (key === "created_at" || key === "updated_at" || key === "deleted_at") {
      return;
    }
    cleaned[key] = value;
  });
  return cleaned;
};
