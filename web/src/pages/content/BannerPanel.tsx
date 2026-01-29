import { useEffect, useMemo, useRef, useState } from "react";
import type { ChangeEvent, DragEvent, Key } from "react";
import {
  Button,
  Card,
  Col,
  Empty,
  Form,
  Image,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography
} from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined, UploadOutlined } from "@ant-design/icons";
import UploadField from "./UploadField";
import SubmissionActions from "./SubmissionActions";
import { BannerItem, DraftVersion, bannerTypeOptions, formatDate, statusOptions, submitStatusLabels } from "./constants";
import type { Notify, RequestFn, UploadFn } from "./utils";
import { buildLocalDraftKey, loadLocalDraft, saveLocalDraft, sanitizeSubmissionPayload } from "./utils";

const { Text } = Typography;

type BannerPanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  notify: Notify;
  operatorId?: number | null;
};

type BannerFormValues = {
  image?: string;
  sort?: number;
  is_active?: boolean;
  type?: number;
};

type BatchBannerItem = {
  uid: string;
  name: string;
  path?: string;
  url?: string;
  type?: number;
  sort?: number;
  is_active?: boolean;
  valid?: boolean | null;
  violations?: Array<{ field: string; rule: unknown; actual: unknown }>;
  meta?: BatchMediaMeta;
  error?: string;
};

type BatchMediaRule = {
  max_size_kb?: number | null;
  min_width?: number | null;
  max_width?: number | null;
  min_height?: number | null;
  max_height?: number | null;
  ratio_width?: number | null;
  ratio_height?: number | null;
  allow_formats?: string | null;
  resize_mode?: string | null;
  target_format?: string | null;
  compress_quality?: number | null;
};

type BatchMediaMeta = {
  SizeBytes?: number;
  Width?: number;
  Height?: number;
  DurationMS?: number;
  Format?: string;
  FileExt?: string;
  size_bytes?: number;
  width?: number;
  height?: number;
  duration_ms?: number;
  format?: string;
  file_ext?: string;
};

type BatchSizePreset = {
  value: string;
  label: string;
  width: number;
  height: number;
};

const resolveBannerModuleKey = (type?: number | null) => {
  if (type === 1) {
    return "banners:left_top";
  }
  if (type === 2) {
    return "banners:left_bottom";
  }
  if (type === 3) {
    return "banners:right";
  }
  return "banners";
};

const formatKB = (value: number) => {
  if (value <= 0) {
    return "";
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(value % 1024 === 0 ? 0 : 1)}MB`;
  }
  return `${value}KB`;
};

const formatBytes = (value: number) => {
  if (value <= 0) {
    return "0KB";
  }
  const kb = value / 1024;
  if (kb >= 1024) {
    return `${(kb / 1024).toFixed(kb % 1024 === 0 ? 0 : 2)}MB`;
  }
  return `${Math.round(kb)}KB`;
};

const readMetaValue = (meta: BatchMediaMeta | undefined, primary: keyof BatchMediaMeta, fallback: keyof BatchMediaMeta) => {
  const value = meta?.[primary];
  if (typeof value === "number" || typeof value === "string") {
    return value;
  }
  return meta?.[fallback];
};

const BannerPanel = ({ version, request, uploadFile, notify, operatorId }: BannerPanelProps) => {
  const [items, setItems] = useState<BannerItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<BannerItem | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [batchOpen, setBatchOpen] = useState(false);
  const [batchItems, setBatchItems] = useState<BatchBannerItem[]>([]);
  const [batchUploading, setBatchUploading] = useState(false);
  const [batchValidating, setBatchValidating] = useState(false);
  const [batchCompressing, setBatchCompressing] = useState(false);
  const [batchSubmitting, setBatchSubmitting] = useState(false);
  const [batchSelectedRowKeys, setBatchSelectedRowKeys] = useState<Key[]>([]);
  const [batchSelectedRows, setBatchSelectedRows] = useState<BatchBannerItem[]>([]);
  const [batchTypeValue, setBatchTypeValue] = useState<number | null>(null);
  const [batchSelectedType, setBatchSelectedType] = useState<number | null>(null);
  const [batchTypeConflict, setBatchTypeConflict] = useState(false);
  const [batchRule, setBatchRule] = useState<BatchMediaRule | null>(null);
  const [batchRuleLoading, setBatchRuleLoading] = useState(false);
  const [batchRuleMissing, setBatchRuleMissing] = useState(false);
  const [batchResizeMode, setBatchResizeMode] = useState<string>("contain");
  const [batchCompressQuality, setBatchCompressQuality] = useState<number>(85);
  const [batchTargetFormat, setBatchTargetFormat] = useState<string>("jpg");
  const [batchResizeWidth, setBatchResizeWidth] = useState<number | null>(null);
  const [batchResizeHeight, setBatchResizeHeight] = useState<number | null>(null);
  const [batchPresetKey, setBatchPresetKey] = useState<string | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<BannerItem[]>([]);
  const [batchSubmitLoading, setBatchSubmitLoading] = useState(false);
  const [filterType, setFilterType] = useState<number | null>(null);
  const [filterSubmitStatus, setFilterSubmitStatus] = useState<string | null>(null);
  const [filterActiveStatus, setFilterActiveStatus] = useState<number | null>(null);
  const [form] = Form.useForm<BannerFormValues>();
  const imageValue = Form.useWatch("image", form);
  const typeValue = Form.useWatch("type", form);
  const batchInputRef = useRef<HTMLInputElement | null>(null);

  const canEdit = Boolean(version?.id);
  const draftKey = version?.id ? buildLocalDraftKey("banners", version.id, editingItem?.id) : "";

  const loadItems = async () => {
    if (!version?.id) {
      setItems([]);
      return;
    }
    setLoading(true);
    try {
      const res = await request<{ data: BannerItem[] }>(`/api/draft/banners?draft_version_id=${version.id}`);
      setItems(res.data || []);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取轮播图失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadItems();
  }, [version?.id]);

  useEffect(() => {
    setSelectedRowKeys([]);
    setSelectedRows([]);
  }, [version?.id]);

  useEffect(() => {
    setBatchItems([]);
    setBatchOpen(false);
    setBatchSelectedRowKeys([]);
    setBatchSelectedRows([]);
    setBatchTypeValue(null);
  }, [version?.id]);

  useEffect(() => {
    if (!batchItems.length) {
      setBatchSelectedRowKeys([]);
      setBatchSelectedRows([]);
    }
  }, [batchItems.length]);

  useEffect(() => {
    if (!batchSelectedRowKeys.length) {
      return;
    }
    const rowMap = new Map(batchItems.map((item) => [item.uid, item]));
    const nextRows = batchSelectedRowKeys
      .map((key) => rowMap.get(String(key)))
      .filter((item): item is BatchBannerItem => Boolean(item));
    setBatchSelectedRows(nextRows);
  }, [batchItems, batchSelectedRowKeys]);

  useEffect(() => {
    if (!batchSelectedRows.length) {
      setBatchSelectedType(null);
      setBatchTypeConflict(false);
      setBatchRule(null);
      setBatchRuleMissing(false);
      return;
    }
    const uniqueTypes = Array.from(new Set(batchSelectedRows.map((item) => item.type ?? 1)));
    if (uniqueTypes.length > 1) {
      setBatchSelectedType(null);
      setBatchTypeConflict(true);
      setBatchRule(null);
      setBatchRuleMissing(false);
      return;
    }
    setBatchTypeConflict(false);
    setBatchSelectedType(uniqueTypes[0] ?? 1);
  }, [batchSelectedRows]);

  useEffect(() => {
    if (!batchSelectedRows.length || batchTypeConflict || !batchSelectedType) {
      setBatchRule(null);
      setBatchRuleMissing(false);
      return;
    }
    const candidate = batchSelectedRows.find((item) => item.path);
    if (!candidate?.path) {
      setBatchRule(null);
      setBatchRuleMissing(false);
      return;
    }
    const loadRule = async () => {
      setBatchRuleLoading(true);
      try {
        const res = await request<{
          rule?: BatchMediaRule | null;
          rule_missing?: boolean;
          warning?: string;
        }>("/api/media/validate", {
          method: "POST",
          body: JSON.stringify({
            module_key: resolveBannerModuleKey(batchSelectedType),
            media_type: "image",
            path: candidate.path
          })
        });
        setBatchRule(res.rule ?? null);
        setBatchRuleMissing(Boolean(res.rule_missing || res.warning === "no_rule"));
      } catch {
        setBatchRule(null);
        setBatchRuleMissing(false);
      } finally {
        setBatchRuleLoading(false);
      }
    };
    void loadRule();
  }, [batchSelectedRows, batchSelectedType, batchTypeConflict]);

  const buildBatchId = () => `batch-${Date.now()}-${Math.random().toString(16).slice(2)}`;

  const batchMetaSummary = useMemo(() => {
    const metas = batchSelectedRows
      .map((item) => item.meta)
      .filter((item): item is BatchMediaMeta => Boolean(item));
    const widths = metas
      .map((meta) => Number(readMetaValue(meta, "Width", "width") || 0))
      .filter((value) => value > 0);
    const heights = metas
      .map((meta) => Number(readMetaValue(meta, "Height", "height") || 0))
      .filter((value) => value > 0);
    const sizes = metas
      .map((meta) => Number(readMetaValue(meta, "SizeBytes", "size_bytes") || 0))
      .filter((value) => value > 0);
    const width = widths.length ? Math.min(...widths) : 0;
    const height = heights.length ? Math.min(...heights) : 0;
    const sizeBytes = sizes.length ? Math.min(...sizes) : 0;
    const format = metas.length
      ? String(readMetaValue(metas[0], "FileExt", "file_ext") || readMetaValue(metas[0], "Format", "format") || "-")
      : "-";
    return {
      width,
      height,
      sizeBytes,
      format
    };
  }, [batchSelectedRows]);

  const batchHasRatioViolation = useMemo(
    () => batchSelectedRows.some((item) => item.violations?.some((violation) => violation.field === "ratio")),
    [batchSelectedRows]
  );

  const batchHasMinViolation = useMemo(
    () =>
      batchSelectedRows.some((item) =>
        item.violations?.some((violation) => violation.field === "width_min" || violation.field === "height_min")
      ),
    [batchSelectedRows]
  );

  const batchRuleSummary = useMemo(() => {
    if (batchRuleMissing) {
      return "未配置规则";
    }
    if (!batchRule) {
      return "未获取规则";
    }
    const toNumber = (value: unknown) => {
      const parsed = Number(value);
      return Number.isFinite(parsed) ? parsed : 0;
    };
    const minW = toNumber(batchRule.min_width);
    const minH = toNumber(batchRule.min_height);
    const maxW = toNumber(batchRule.max_width);
    const maxH = toNumber(batchRule.max_height);
    const hasSizeLimit = minW > 0 || minH > 0 || maxW > 0 || maxH > 0;
    const parts: string[] = [];
    if (hasSizeLimit) {
      parts.push(`${minW > 0 ? minW : "-"}×${minH > 0 ? minH : "-"} ~ ${maxW > 0 ? maxW : "-"}×${maxH > 0 ? maxH : "-"}`);
    }
    if (batchRule.max_size_kb) {
      parts.push(formatKB(batchRule.max_size_kb));
    }
    if (batchRule.ratio_width && batchRule.ratio_height) {
      parts.push(`比例 ${batchRule.ratio_width}:${batchRule.ratio_height}`);
    }
    if (!parts.length) {
      return "规则未限定尺寸或大小";
    }
    return `规则 ${parts.join(" · ")}`;
  }, [batchRule, batchRuleMissing]);

  const batchRatioConfig = useMemo(() => {
    if (!batchRule?.ratio_width || !batchRule?.ratio_height) {
      return null;
    }
    const gcd = (a: number, b: number): number => (b === 0 ? a : gcd(b, a % b));
    const baseGcd = gcd(batchRule.ratio_width, batchRule.ratio_height);
    const baseW = Math.max(1, Math.floor(batchRule.ratio_width / baseGcd));
    const baseH = Math.max(1, Math.floor(batchRule.ratio_height / baseGcd));

    let minScale = 1;
    if ((batchRule.min_width ?? 0) > 0) {
      minScale = Math.max(minScale, Math.ceil((batchRule.min_width ?? 0) / baseW));
    }
    if ((batchRule.min_height ?? 0) > 0) {
      minScale = Math.max(minScale, Math.ceil((batchRule.min_height ?? 0) / baseH));
    }

    let maxScale = Number.POSITIVE_INFINITY;
    if ((batchRule.max_width ?? 0) > 0) {
      maxScale = Math.min(maxScale, Math.floor((batchRule.max_width ?? 0) / baseW));
    }
    if ((batchRule.max_height ?? 0) > 0) {
      maxScale = Math.min(maxScale, Math.floor((batchRule.max_height ?? 0) / baseH));
    }

    if (batchMetaSummary.width > 0 && batchMetaSummary.height > 0) {
      const maxByOriginal = Math.min(
        Math.floor(batchMetaSummary.width / baseW),
        Math.floor(batchMetaSummary.height / baseH)
      );
      if (maxByOriginal > 0) {
        maxScale = Math.min(maxScale, maxByOriginal);
      }
    }

    if (!Number.isFinite(maxScale)) {
      maxScale = minScale;
    }

    return {
      baseW,
      baseH,
      minScale,
      maxScale
    };
  }, [batchMetaSummary.height, batchMetaSummary.width, batchRule]);

  const batchSizePresets = useMemo(() => {
    if (!batchRule || !batchMetaSummary.width || !batchMetaSummary.height) {
      return [];
    }
    const presets: BatchSizePreset[] = [];

    if (batchRatioConfig) {
      const { baseW, baseH, minScale, maxScale } = batchRatioConfig;
      if (maxScale < minScale || maxScale <= 0) {
        return [];
      }
      const span = maxScale - minScale;
      const steps = 4;
      const candidates: number[] = [];
      for (let i = 0; i <= steps; i += 1) {
        const scale = Math.round(maxScale - (span * i) / steps);
        candidates.push(scale);
      }
      const unique = Array.from(new Set(candidates))
        .filter((scale) => scale >= minScale && scale <= maxScale)
        .sort((a, b) => b - a);
      let cursor = maxScale;
      while (unique.length < 5 && cursor >= minScale) {
        if (!unique.includes(cursor)) {
          unique.push(cursor);
        }
        cursor -= 1;
      }
      unique.slice(0, 5).forEach((scale, index) => {
        const width = baseW * scale;
        const height = baseH * scale;
        presets.push({
          value: `ratio_${scale}`,
          label: `推荐${index + 1} ${width}×${height}`,
          width,
          height
        });
      });
      return presets;
    }

    const clamp = (width: number, height: number) => {
      let nextWidth = width;
      let nextHeight = height;
      if (batchMetaSummary.width > 0 && nextWidth > batchMetaSummary.width) {
        nextWidth = batchMetaSummary.width;
      }
      if (batchMetaSummary.height > 0 && nextHeight > batchMetaSummary.height) {
        nextHeight = batchMetaSummary.height;
      }
      return { width: nextWidth, height: nextHeight };
    };

    presets.push({
      value: "baseline",
      label: `基准尺寸 ${batchMetaSummary.width}×${batchMetaSummary.height}`,
      width: batchMetaSummary.width,
      height: batchMetaSummary.height
    });

    if ((batchRule.max_width ?? 0) > 0 && (batchRule.max_height ?? 0) > 0) {
      const maxSize = clamp(batchRule.max_width ?? 0, batchRule.max_height ?? 0);
      presets.push({
        value: "rule_max",
        label: `规则最大 ${maxSize.width}×${maxSize.height}`,
        width: maxSize.width,
        height: maxSize.height
      });
    }
    if ((batchRule.min_width ?? 0) > 0 && (batchRule.min_height ?? 0) > 0) {
      const minSize = clamp(batchRule.min_width ?? 0, batchRule.min_height ?? 0);
      presets.push({
        value: "rule_min",
        label: `规则最小 ${minSize.width}×${minSize.height}`,
        width: minSize.width,
        height: minSize.height
      });
    }
    if ((batchRule.max_width ?? 0) > 0 && (batchRule.min_width ?? 0) > 0) {
      const midWidth = Math.round(((batchRule.max_width ?? 0) + (batchRule.min_width ?? 0)) / 2);
      const midHeight = Math.round(((batchRule.max_height ?? 0) + (batchRule.min_height ?? 0)) / 2);
      const midSize = clamp(midWidth, midHeight);
      presets.push({
        value: "rule_mid",
        label: `推荐尺寸 ${midSize.width}×${midSize.height}`,
        width: midSize.width,
        height: midSize.height
      });
    }

    return presets;
  }, [batchMetaSummary.height, batchMetaSummary.width, batchRatioConfig, batchRule]);

  const applyBatchPreset = (key: string) => {
    const preset = batchSizePresets.find((item) => item.value === key);
    if (!preset) {
      return;
    }
    let nextWidth = preset.width;
    let nextHeight = preset.height;

    if (batchRule?.ratio_width && batchRule?.ratio_height && nextWidth > 0 && nextHeight > 0) {
      const ratio = batchRule.ratio_width / batchRule.ratio_height;
      const fitHeight = Math.round(nextWidth / ratio);
      if (fitHeight <= nextHeight) {
        nextHeight = fitHeight;
      } else {
        nextWidth = Math.round(nextHeight * ratio);
      }
    }

    if (batchMetaSummary.width > 0 && batchMetaSummary.height > 0 && nextWidth > 0 && nextHeight > 0) {
      const scale = Math.min(batchMetaSummary.width / nextWidth, batchMetaSummary.height / nextHeight, 1);
      nextWidth = Math.max(1, Math.round(nextWidth * scale));
      nextHeight = Math.max(1, Math.round(nextHeight * scale));
    }

    setBatchResizeWidth(nextWidth);
    setBatchResizeHeight(nextHeight);
  };

  const handleBatchResizeWidth = (value?: number | null) => {
    const next = value ?? null;
    if (!next) {
      setBatchResizeWidth(null);
      return;
    }
    if (batchRatioConfig) {
      let scale = Math.round(next / batchRatioConfig.baseW);
      if (scale < batchRatioConfig.minScale) {
        scale = batchRatioConfig.minScale;
      }
      if (batchRatioConfig.maxScale > 0) {
        scale = Math.min(scale, batchRatioConfig.maxScale);
      }
      const width = batchRatioConfig.baseW * scale;
      const height = batchRatioConfig.baseH * scale;
      setBatchResizeWidth(width);
      setBatchResizeHeight(height);
      return;
    }
    setBatchResizeWidth(next);
    if (batchResizeMode !== "fill" && batchMetaSummary.width > 0 && batchMetaSummary.height > 0) {
      const computed = Math.round((next * batchMetaSummary.height) / batchMetaSummary.width);
      setBatchResizeHeight(computed);
    }
  };

  const handleBatchResizeHeight = (value?: number | null) => {
    const next = value ?? null;
    if (!next) {
      setBatchResizeHeight(null);
      return;
    }
    if (batchRatioConfig) {
      let scale = Math.round(next / batchRatioConfig.baseH);
      if (scale < batchRatioConfig.minScale) {
        scale = batchRatioConfig.minScale;
      }
      if (batchRatioConfig.maxScale > 0) {
        scale = Math.min(scale, batchRatioConfig.maxScale);
      }
      const width = batchRatioConfig.baseW * scale;
      const height = batchRatioConfig.baseH * scale;
      setBatchResizeWidth(width);
      setBatchResizeHeight(height);
      return;
    }
    setBatchResizeHeight(next);
    if (batchResizeMode !== "fill" && batchMetaSummary.width > 0 && batchMetaSummary.height > 0) {
      const computed = Math.round((next * batchMetaSummary.width) / batchMetaSummary.height);
      setBatchResizeWidth(computed);
    }
  };

  useEffect(() => {
    if (!batchRule) {
      setBatchPresetKey(null);
      setBatchResizeWidth(null);
      setBatchResizeHeight(null);
      return;
    }
    const defaultQuality = batchRule.compress_quality && batchRule.compress_quality > 0 ? batchRule.compress_quality : 85;
    setBatchCompressQuality(defaultQuality);
    const defaultTarget = batchRule.target_format ? batchRule.target_format : "jpg";
    setBatchTargetFormat(defaultTarget);
    let defaultResizeMode =
      batchRule.resize_mode
        ? batchRule.resize_mode
        : batchRule.ratio_width && batchRule.ratio_height
          ? "cover"
          : "contain";
    if (batchRatioConfig && batchHasRatioViolation && defaultResizeMode === "contain") {
      defaultResizeMode = "cover";
    }
    setBatchResizeMode(defaultResizeMode);
    setBatchPresetKey(null);
    setBatchResizeWidth(null);
    setBatchResizeHeight(null);
  }, [batchRule, batchRatioConfig, batchHasRatioViolation]);

  useEffect(() => {
    if (!batchSizePresets.length || batchPresetKey) {
      return;
    }
    const first = batchSizePresets[0];
    setBatchPresetKey(first.value);
    setBatchResizeWidth(first.width);
    setBatchResizeHeight(first.height);
  }, [batchPresetKey, batchSizePresets]);

  useEffect(() => {
    if (batchRatioConfig && batchHasRatioViolation && batchResizeMode === "contain") {
      setBatchResizeMode("cover");
    }
  }, [batchHasRatioViolation, batchRatioConfig, batchResizeMode]);

  const openBatch = () => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    setBatchOpen(true);
  };

  const updateBatchItem = (uid: string, patch: Partial<BatchBannerItem>) => {
    setBatchItems((prev) => prev.map((item) => (item.uid === uid ? { ...item, ...patch } : item)));
  };

  const processBatchFiles = async (files: File[]) => {
    if (!files.length) {
      return;
    }
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    setBatchUploading(true);
    const nextItems: BatchBannerItem[] = [];
    for (const file of files) {
      const uid = buildBatchId();
      try {
        const result = await uploadFile(file, "banners", version.id);
        let valid: boolean | null = null;
        let violations: Array<{ field: string; rule: unknown; actual: unknown }> = [];
        let meta: BatchMediaMeta | undefined;
        try {
          const validateRes = await request<{
            valid: boolean;
            violations: Array<{ field: string; rule: unknown; actual: unknown }>;
            meta?: BatchMediaMeta;
          }>(
            "/api/media/validate",
            {
              method: "POST",
              body: JSON.stringify({
                module_key: resolveBannerModuleKey(1),
                media_type: "image",
                path: result.path
              })
            }
          );
          valid = validateRes.valid;
          violations = validateRes.violations ?? [];
          meta = validateRes.meta;
        } catch {
          valid = null;
        }
        nextItems.push({
          uid,
          name: file.name,
          path: result.path,
          url: result.url,
          type: 1,
          sort: 0,
          is_active: true,
          valid,
          violations,
          meta
        });
      } catch (error) {
        nextItems.push({
          uid,
          name: file.name,
          type: 1,
          sort: 0,
          is_active: true,
          valid: null,
          violations: [],
          error: error instanceof Error ? error.message : "上传失败"
        });
      }
    }
    const failed = nextItems.filter((item) => item.error).length;
    if (failed > 0) {
      notify.warning(`有 ${failed} 张轮播图上传失败`);
    }
    setBatchItems((prev) => [...prev, ...nextItems]);
    setBatchUploading(false);
  };

  const handleBatchSelect = async (event: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files ?? []);
    event.target.value = "";
    await processBatchFiles(files);
  };

  const handleBatchDrop = async (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    if (batchUploading) {
      return;
    }
    const files = Array.from(event.dataTransfer.files ?? []);
    await processBatchFiles(files);
  };

  const handleBatchValidate = async () => {
    if (!batchItems.length) {
      notify.warning("请先选择轮播图文件");
      return;
    }
    setBatchValidating(true);
    const updated = await Promise.all(
      batchItems.map(async (item) => {
        if (!item.path) {
          return { ...item, valid: false, error: item.error ?? "未上传" };
        }
        try {
          const res = await request<{
            valid: boolean;
            violations: Array<{ field: string; rule: unknown; actual: unknown }>;
            meta?: BatchMediaMeta;
          }>(
            "/api/media/validate",
            {
              method: "POST",
              body: JSON.stringify({
                module_key: resolveBannerModuleKey(item.type),
                media_type: "image",
                path: item.path
              })
            }
          );
          return { ...item, valid: res.valid, violations: res.violations ?? [], meta: res.meta, error: undefined };
        } catch (error) {
          return {
            ...item,
            valid: false,
            error: error instanceof Error ? error.message : "校验失败"
          };
        }
      })
    );
    setBatchItems(updated);
    setBatchValidating(false);
  };

  const handleBatchCompress = async () => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    if (!operatorId) {
      notify.warning("缺少提交人信息");
      return;
    }
    if (!batchItems.length) {
      notify.warning("暂无可压缩的轮播图");
      return;
    }
    if (!batchSelectedRows.length) {
      notify.warning("请先选择要压缩的轮播图");
      return;
    }
    if (batchTypeConflict) {
      notify.warning("批量压缩仅支持同位置轮播图");
      return;
    }
    setBatchCompressing(true);
    const updated = [...batchItems];
    const baseRuleOverride: Record<string, unknown> = {};
    if (batchResizeMode) {
      baseRuleOverride.resize_mode = batchResizeMode === "contain" ? "" : batchResizeMode;
    }
    if (batchTargetFormat && batchTargetFormat !== "keep") {
      baseRuleOverride.target_format = batchTargetFormat;
    }
    if (batchRule?.allow_formats) {
      baseRuleOverride.allow_formats = batchRule.allow_formats;
    }
    baseRuleOverride.compress_quality = batchCompressQuality;
    for (let index = 0; index < updated.length; index += 1) {
      const item = updated[index];
      if (!batchSelectedRowKeys.includes(item.uid)) {
        continue;
      }
      if (!item.path) {
        updated[index] = { ...item, error: item.error ?? "未上传" };
        continue;
      }
      try {
        const ruleOverride: Record<string, unknown> = { ...baseRuleOverride };
        if (batchResizeWidth && batchResizeHeight) {
          let targetWidth = batchResizeWidth;
          let targetHeight = batchResizeHeight;
          const metaWidth = Number(readMetaValue(item.meta, "Width", "width") || 0);
          const metaHeight = Number(readMetaValue(item.meta, "Height", "height") || 0);
          if (batchResizeMode !== "fill" && metaWidth > 0 && targetWidth > metaWidth) {
            targetWidth = metaWidth;
          }
          if (batchResizeMode !== "fill" && metaHeight > 0 && targetHeight > metaHeight) {
            targetHeight = metaHeight;
          }
          ruleOverride.max_width = targetWidth;
          ruleOverride.max_height = targetHeight;
        }
        const transformRes = await request<{ path: string; url?: string }>("/api/media/transform", {
          method: "POST",
          body: JSON.stringify({
            draft_version_id: version.id,
            module_key: resolveBannerModuleKey(item.type),
            media_type: "image",
            path: item.path,
            operator_id: operatorId,
            rule: ruleOverride
          })
        });
        const nextPath = transformRes.path;
        const nextUrl = transformRes.url;
        let valid: boolean | null = null;
        let violations: Array<{ field: string; rule: unknown; actual: unknown }> = [];
        let meta: BatchMediaMeta | undefined;
        try {
          const validateRes = await request<{
            valid: boolean;
            violations: Array<{ field: string; rule: unknown; actual: unknown }>;
            meta?: BatchMediaMeta;
          }>(
            "/api/media/validate",
            {
              method: "POST",
              body: JSON.stringify({
                module_key: resolveBannerModuleKey(item.type),
                media_type: "image",
                path: nextPath
              })
            }
          );
          valid = validateRes.valid;
          violations = validateRes.violations ?? [];
          meta = validateRes.meta;
        } catch {
          valid = null;
        }
        updated[index] = {
          ...item,
          path: nextPath,
          url: nextUrl ?? item.url,
          valid,
          violations,
          meta,
          error: undefined
        };
      } catch (error) {
        updated[index] = {
          ...item,
          error: error instanceof Error ? error.message : "压缩失败"
        };
      }
    }
    setBatchItems(updated);
    setBatchCompressing(false);
  };

  const handleBatchSetType = () => {
    if (batchTypeValue == null) {
      notify.warning("请选择要设置的位置");
      return;
    }
    if (!batchSelectedRows.length) {
      notify.warning("请先选择要设置位置的轮播图");
      return;
    }
    setBatchItems((prev) =>
      prev.map((item) =>
        batchSelectedRowKeys.includes(item.uid)
          ? { ...item, type: batchTypeValue }
          : item
      )
    );
    notify.success("已批量设置位置");
  };

  const handleBatchSubmit = async () => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    if (!batchItems.length) {
      notify.warning("暂无可提交的轮播图");
      return;
    }
    setBatchSubmitting(true);
    let success = 0;
    let failed = 0;
    const updated = [...batchItems];
    for (let index = 0; index < updated.length; index += 1) {
      const item = updated[index];
      if (!item.path) {
        continue;
      }
      try {
        await request("/api/draft/banners", {
          method: "POST",
          body: JSON.stringify({
            draft_version_id: version.id,
            app_version_name: version.app_version_name || undefined,
            image: item.path,
            sort: item.sort ?? 0,
            is_active: item.is_active ? 1 : 0,
            type: item.type ?? 1,
            created_by: operatorId ?? undefined,
            updated_by: operatorId ?? undefined
          })
        });
        updated[index] = { ...item, error: undefined };
        success += 1;
      } catch (error) {
        failed += 1;
        updated[index] = {
          ...item,
          error: error instanceof Error ? error.message : "创建失败"
        };
      }
    }
    setBatchItems(updated);
    if (success > 0) {
      notify.success(`已创建 ${success} 条轮播图`);
      void loadItems();
    }
    if (failed > 0) {
      notify.warning(`有 ${failed} 条轮播图未能创建`);
    }
    if (failed == 0) {
      setBatchOpen(false);
      setBatchItems([]);
    }
    setBatchSubmitting(false);
  };

  const handleBatchSubmission = async () => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    if (!operatorId) {
      notify.warning("缺少提交人信息");
      return;
    }
    if (!selectedRows.length) {
      notify.warning("请先选择要提交的记录");
      return;
    }
    setBatchSubmitLoading(true);
    try {
      const results = await Promise.allSettled(
        selectedRows.map((item) =>
          request("/api/draft/submit", {
            method: "POST",
            body: JSON.stringify({
              draft_version_id: version.id,
              module_key: "banners",
              entity_table: "app_db_banners",
              entity_id: item.id,
              submit_by: operatorId,
              payload: sanitizeSubmissionPayload(item as Record<string, unknown>)
            })
          })
        )
      );
      const failed = results.filter((item) => item.status === "rejected");
      if (failed.length > 0) {
        notify.warning(`批量提交完成，失败 ${failed.length} 条`);
      } else {
        notify.success("批量提交完成");
      }
      setSelectedRowKeys([]);
      setSelectedRows([]);
      void loadItems();
    } finally {
      setBatchSubmitLoading(false);
    }
  };

  const openEditor = (item?: BannerItem) => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    setEditingItem(item ?? null);
    setPreviewUrl(item?.image_url ?? null);
    setEditorOpen(true);
  };

  const syncEditorForm = (item: BannerItem | null) => {
    form.setFieldsValue({
      image: item?.image ?? undefined,
      sort: item?.sort ?? 0,
      is_active: (item?.is_active ?? 1) === 1,
      type: item?.type ?? 1
    });
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (!version?.id) {
        return;
      }
      setEditorSubmitting(true);
      const payload: Record<string, unknown> = {
        draft_version_id: version.id,
        app_version_name: version.app_version_name || undefined,
        image: values.image || undefined,
        sort: values.sort ?? 0,
        is_active: values.is_active ? 1 : 0,
        type: values.type ?? 1,
        updated_by: operatorId ?? undefined
      };
      if (!editingItem && operatorId) {
        payload.created_by = operatorId;
      }
      if (editingItem) {
        await request(`/api/draft/banners/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        notify.success("轮播图已更新");
      } else {
        await request("/api/draft/banners", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        notify.success("轮播图已创建");
      }
      setEditorOpen(false);
      setEditingItem(null);
      setPreviewUrl(null);
      form.resetFields();
      void loadItems();
    } catch (error) {
      if (error instanceof Error) {
        notify.error(error.message);
      }
    } finally {
      setEditorSubmitting(false);
    }
  };

  const handleDelete = (item: BannerItem) => {
    Modal.confirm({
      title: "确认删除轮播图？",
      content: "该轮播图将被移除。",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/draft/banners/${item.id}`, { method: "DELETE" });
          notify.success("已删除");
          void loadItems();
        } catch (error) {
          if (error instanceof Error) {
            notify.error(error.message);
          }
        }
      }
    });
  };

  const handleSaveLocalDraft = () => {
    if (!draftKey) {
      notify.warning("缺少草稿上下文");
      return;
    }
    saveLocalDraft(draftKey, form.getFieldsValue());
    notify.success("本地草稿已保存");
  };

  const handleLoadLocalDraft = () => {
    if (!draftKey) {
      notify.warning("缺少草稿上下文");
      return;
    }
    const data = loadLocalDraft(draftKey);
    if (!data) {
      notify.warning("未找到本地草稿");
      return;
    }
    form.setFieldsValue(data);
    notify.success("本地草稿已恢复");
  };

  const typeLabelMap = useMemo(() => {
    const map = new Map<number, string>();
    bannerTypeOptions.forEach((option) => map.set(option.value, option.label));
    return map;
  }, []);

  const submitStatusOptions = useMemo(
    () => [
      { value: "none", label: "未提交" },
      ...Object.entries(submitStatusLabels).map(([value, config]) => ({
        value,
        label: config.label
      }))
    ],
    []
  );

  const filteredItems = useMemo(() => {
    return items.filter((item) => {
      if (filterType != null && (item.type ?? 1) !== filterType) {
        return false;
      }
      if (filterActiveStatus != null && (item.is_active ?? 1) !== filterActiveStatus) {
        return false;
      }
      if (filterSubmitStatus) {
        if (filterSubmitStatus === "none") {
          if (item.submit_status) {
            return false;
          }
        } else if (item.submit_status !== filterSubmitStatus) {
          return false;
        }
      }
      return true;
    });
  }, [filterActiveStatus, filterSubmitStatus, filterType, items]);

  const columns = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      width: 80
    },
    {
      title: "位置",
      dataIndex: "type",
      key: "type",
      render: (value: number) => <Tag>{typeLabelMap.get(value) || `类型${value ?? "-"}`}</Tag>
    },
    {
      title: "排序",
      dataIndex: "sort",
      key: "sort",
      render: (value: number) => <Text>{value ?? 0}</Text>
    },
    {
      title: "状态",
      dataIndex: "is_active",
      key: "is_active",
      render: (value: number) => (value === 0 ? <Tag>停用</Tag> : <Tag color="green">启用</Tag>)
    },
    {
      title: "提交状态",
      dataIndex: "submit_status",
      key: "submit_status",
      render: (value?: string | null) => {
        if (!value) {
          return <Tag>未提交</Tag>;
        }
        const config = submitStatusLabels[value];
        if (!config) {
          return <Tag>{value}</Tag>;
        }
        return <Tag color={config.color}>{config.label}</Tag>;
      }
    },
    {
      title: "上次提交",
      dataIndex: "last_submit_at",
      key: "last_submit_at",
      render: (value?: string | null) => <Text type="secondary">{formatDate(value ?? null)}</Text>
    },
    {
      title: "预览",
      dataIndex: "image_url",
      key: "image_url",
      render: (value: string) =>
        value ? <Image src={value} width={80} style={{ borderRadius: 8 }} /> : "-"
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: BannerItem) => (
        <Space>
          <SubmissionActions
            draftVersionId={version?.id || 0}
            moduleKey="banners"
            entityTable="app_db_banners"
            entityId={record.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            getPayload={() => record as Record<string, unknown>}
            disabled={!version?.id}
          />
          <Button size="small" icon={<EditOutlined />} onClick={() => openEditor(record)}>
            编辑
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      )
    }
  ];

  const batchColumns = [
    {
      title: "预览",
      dataIndex: "url",
      key: "preview",
      render: (value: string) =>
        value ? <Image src={value} width={72} style={{ borderRadius: 8 }} /> : "-"
    },
    {
      title: "文件",
      dataIndex: "name",
      key: "name",
      render: (value: string, record: BatchBannerItem) => (
        <Space direction="vertical" size={0}>
          <Text>{value}</Text>
          {record.path ? (
            <Text type="secondary" style={{ maxWidth: 220 }} ellipsis>
              {record.path}
            </Text>
          ) : null}
        </Space>
      )
    },
    {
      title: "位置",
      dataIndex: "type",
      key: "type",
      render: (value: number, record: BatchBannerItem) => (
        <Select
          value={value ?? 1}
          options={bannerTypeOptions}
          onChange={(next) => updateBatchItem(record.uid, { type: next })}
          style={{ width: 120 }}
        />
      )
    },
    {
      title: "排序",
      dataIndex: "sort",
      key: "sort",
      render: (value: number, record: BatchBannerItem) => (
        <InputNumber
          min={0}
          max={999}
          value={value ?? 0}
          onChange={(next) => updateBatchItem(record.uid, { sort: typeof next === "number" ? next : 0 })}
        />
      )
    },
    {
      title: "启用",
      dataIndex: "is_active",
      key: "is_active",
      render: (value: boolean, record: BatchBannerItem) => (
        <Switch checked={value ?? true} onChange={(checked) => updateBatchItem(record.uid, { is_active: checked })} />
      )
    },
    {
      title: "校验",
      key: "validate",
      render: (_: string, record: BatchBannerItem) => {
        if (record.error) {
          return <Text type="danger">{record.error}</Text>;
        }
        if (record.valid == null) {
          return <Text type="secondary">未检测</Text>;
        }
        if (record.valid) {
          return <Tag color="green">通过</Tag>;
        }
        const count = record.violations?.length ?? 0;
        return (
          <Space direction="vertical" size={0}>
            <Tag color="red">未通过</Tag>
            {count ? <Text type="secondary">违规 {count} 项</Text> : null}
          </Space>
        );
      }
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: BatchBannerItem) => (
        <Button
          size="small"
          onClick={() => setBatchItems((prev) => prev.filter((item) => item.uid !== record.uid))}
        >
          移除
        </Button>
      )
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">轮播图需要区分左上、左下与右侧三种位置。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadItems} disabled={!canEdit}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()} disabled={!canEdit}>
              新建轮播图
            </Button>
            <Button icon={<UploadOutlined />} onClick={openBatch} disabled={!canEdit}>
              批量上传
            </Button>
            <Button onClick={handleBatchSubmission} loading={batchSubmitLoading} disabled={!selectedRowKeys.length || !canEdit}>
              批量提交
            </Button>
          </Space>
          <Space wrap>
            <Text type="secondary">筛选</Text>
            <Select
              placeholder="位置"
              style={{ width: 140 }}
              value={filterType ?? undefined}
              onChange={(value) => setFilterType(value)}
              options={bannerTypeOptions}
              allowClear
            />
            <Select
              placeholder="提交状态"
              style={{ width: 160 }}
              value={filterSubmitStatus ?? undefined}
              onChange={(value) => setFilterSubmitStatus(value)}
              options={submitStatusOptions}
              allowClear
            />
            <Select
              placeholder="启用状态"
              style={{ width: 140 }}
              value={filterActiveStatus ?? undefined}
              onChange={(value) => setFilterActiveStatus(value)}
              options={statusOptions.map((item) => ({ value: item.value, label: item.label }))}
              allowClear
            />
            <Button
              onClick={() => {
                setFilterType(null);
                setFilterSubmitStatus(null);
                setFilterActiveStatus(null);
              }}
              disabled={!filterType && !filterSubmitStatus && filterActiveStatus == null}
            >
              清空筛选
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        {version?.id ? (
          <Table
            rowKey="id"
            columns={columns}
            dataSource={filteredItems}
            loading={loading}
            pagination={{ pageSize: 10, showSizeChanger: true, pageSizeOptions: ["10", "20", "50"] }}
            rowSelection={{
              selectedRowKeys,
              onChange: (keys, rows) => {
                setSelectedRowKeys(keys);
                setSelectedRows(rows as BannerItem[]);
              }
            }}
          />
        ) : (
          <Empty description="请选择景区版本后录入轮播图" />
        )}
      </Card>

      <Modal
        title={editingItem ? "编辑轮播图" : "新建轮播图"}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false);
          setEditingItem(null);
          setPreviewUrl(null);
          form.resetFields();
        }}
        afterOpenChange={(open) => {
          if (open) {
            syncEditorForm(editingItem);
          }
        }}
        width={900}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        footer={[
          <Button key="draft-save" onClick={handleSaveLocalDraft}>
            保存草稿
          </Button>,
          <Button key="draft-load" onClick={handleLoadLocalDraft}>
            恢复草稿
          </Button>,
          <Button key="cancel" onClick={() => {
            setEditorOpen(false);
            setEditingItem(null);
            setPreviewUrl(null);
            form.resetFields();
          }}>
            取消
          </Button>,
          <Button key="submit" type="primary" onClick={handleSubmit} loading={editorSubmitting}>
            保存
          </Button>
        ]}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Row gutter={12}>
            <Col span={8}>
              <Form.Item label="位置" name="type" rules={[{ required: true, message: "请选择轮播图位置" }]}>
                <Select options={bannerTypeOptions} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="排序" name="sort">
                <InputNumber min={0} max={999} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="启用状态" name="is_active" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="轮播图路径" name="image" rules={[{ required: true, message: "请上传轮播图" }]}>
            <Input placeholder="上传后自动填充" />
          </Form.Item>
          <UploadField
            label="轮播图文件"
            accept="image/png,image/jpeg"
            value={imageValue}
            previewUrl={previewUrl}
            previewType="image"
            mediaType="image"
            moduleKey={resolveBannerModuleKey(typeValue ?? editingItem?.type)}
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            onUpload={async (file) => {
              if (!version?.id) {
                throw new Error("缺少版本信息");
              }
              return uploadFile(file, "banners", version.id);
            }}
            onChange={(path, url) => {
              form.setFieldValue("image", path);
              setPreviewUrl(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("image", undefined);
              setPreviewUrl(null);
            }}
          />
        </Form>
      </Modal>

      <Modal
        title="批量上传轮播图"
        open={batchOpen}
        onCancel={() => setBatchOpen(false)}
        footer={null}
        width={960}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        destroyOnClose
      >
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <Space wrap>
            <Button icon={<UploadOutlined />} onClick={() => batchInputRef.current?.click()} loading={batchUploading}>
              选择文件
            </Button>
            <Button onClick={handleBatchValidate} loading={batchValidating} disabled={!batchItems.length}>
              批量检测
            </Button>
            <Button
              onClick={handleBatchCompress}
              loading={batchCompressing}
              disabled={!batchItems.length || !batchSelectedRowKeys.length || batchTypeConflict}
            >
              批量智能压缩
            </Button>
            <Select
              style={{ width: 180 }}
              placeholder="批量设置位置"
              value={batchTypeValue ?? undefined}
              onChange={(value) => setBatchTypeValue(value)}
              options={bannerTypeOptions}
              allowClear
            />
            <Button onClick={handleBatchSetType} disabled={!batchItems.length}>
              应用位置
            </Button>
            <Button type="primary" onClick={handleBatchSubmit} loading={batchSubmitting} disabled={!batchItems.length}>
              批量保存
            </Button>
            <Button onClick={() => setBatchItems([])} disabled={!batchItems.length}>
              清空列表
            </Button>
          </Space>
          <div
            onDragOver={(event) => {
              event.preventDefault();
            }}
            onDrop={handleBatchDrop}
            style={{
              border: "1px dashed #d9d9d9",
              borderRadius: 12,
              padding: "12px 16px",
              textAlign: "center",
              background: "#fafafa",
              color: "#595959"
            }}
          >
            拖拽图片到此处可批量上传
          </div>
          {batchSelectedRows.length ? (
            <Card size="small" style={{ borderRadius: 12 }}>
              <Space direction="vertical" size={12} style={{ width: "100%" }}>
                <Space wrap>
                  <Text type="secondary">批量智能压缩参数</Text>
                  {batchRuleLoading ? <Text type="secondary">规则加载中…</Text> : null}
                  {batchSelectedType ? (
                    <Tag color="blue">{typeLabelMap.get(batchSelectedType) || `类型${batchSelectedType}`}</Tag>
                  ) : null}
                </Space>
                {batchTypeConflict ? (
                  <Text type="danger">已选轮播图包含不同位置，请先统一位置再批量压缩。</Text>
                ) : null}
                <Space direction="vertical" size={4}>
                  <Text type="secondary">{batchRuleSummary}</Text>
                  {batchRule?.ratio_width && batchRule?.ratio_height ? (
                    <Text type="secondary">
                      目标比例 {batchRule.ratio_width}:{batchRule.ratio_height}
                    </Text>
                  ) : null}
                  {batchMetaSummary.width > 0 && batchMetaSummary.height > 0 ? (
                    <Text type="secondary">
                      参考尺寸 {batchMetaSummary.width}×{batchMetaSummary.height} · 大小 {formatBytes(batchMetaSummary.sizeBytes)}
                    </Text>
                  ) : (
                    <Text type="secondary">尚未获取图片尺寸，推荐尺寸不可用。</Text>
                  )}
                  {batchHasRatioViolation ? <Text type="secondary">比例不匹配，请选择合适的调整方式。</Text> : null}
                  {batchHasMinViolation ? <Text type="secondary">尺寸小于最小限制，智能压缩不会放大。</Text> : null}
                  {batchRuleMissing ? <Text type="secondary">当前模块未配置规则，将按默认策略压缩。</Text> : null}
                </Space>
                {batchSizePresets.length ? (
                  <Select
                    value={batchPresetKey ?? undefined}
                    placeholder="推荐尺寸"
                    options={batchSizePresets.map((item) => ({ value: item.value, label: item.label }))}
                    onChange={(value) => {
                      setBatchPresetKey(value);
                      applyBatchPreset(value);
                    }}
                    style={{ width: "100%" }}
                  />
                ) : null}
                <Space wrap>
                  <Text type="secondary">目标尺寸</Text>
                  <InputNumber min={1} value={batchResizeWidth ?? undefined} onChange={handleBatchResizeWidth} />
                  <Text type="secondary">×</Text>
                  <InputNumber min={1} value={batchResizeHeight ?? undefined} onChange={handleBatchResizeHeight} />
                </Space>
                {batchRatioConfig ? <Text type="secondary">按比例自动换算尺寸，严格匹配规则。</Text> : null}
                <Space wrap>
                  <Text type="secondary">调整方式</Text>
                  <Select
                    value={batchResizeMode}
                    onChange={(value) => setBatchResizeMode(value)}
                    options={[
                      { value: "contain", label: "等比缩放（不裁剪）", disabled: !!batchRatioConfig && batchHasRatioViolation },
                      { value: "cover", label: "裁剪填充（保持比例）" },
                      { value: "fill", label: "拉伸到目标尺寸" }
                    ]}
                    style={{ minWidth: 220 }}
                  />
                </Space>
                <Space wrap>
                  <Text type="secondary">质量(%)</Text>
                  <InputNumber min={1} max={100} value={batchCompressQuality} onChange={(value) => setBatchCompressQuality(value ?? 85)} />
                </Space>
                <Select
                  value={batchTargetFormat}
                  onChange={(value) => setBatchTargetFormat(value)}
                  options={[
                    { value: "jpg", label: "转为 JPG" },
                    { value: "keep", label: "保持原格式" }
                  ]}
                  style={{ width: "100%" }}
                />
              </Space>
            </Card>
          ) : (
            <Text type="secondary">选择要压缩的图片后，可在此设置智能压缩参数。</Text>
          )}
          <Table
            rowKey="uid"
            columns={batchColumns}
            dataSource={batchItems}
            pagination={false}
            scroll={{ x: 900 }}
            rowSelection={{
              selectedRowKeys: batchSelectedRowKeys,
              onChange: (keys, rows) => {
                setBatchSelectedRowKeys(keys);
                setBatchSelectedRows(rows as BatchBannerItem[]);
              }
            }}
          />
        </Space>
        <input
          ref={batchInputRef}
          type="file"
          accept="image/*"
          multiple
          onChange={handleBatchSelect}
          style={{ display: "none" }}
        />
      </Modal>
    </Space>
  );
};

export default BannerPanel;
