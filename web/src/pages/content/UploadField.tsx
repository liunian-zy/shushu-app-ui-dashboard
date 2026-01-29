import { useEffect, useMemo, useRef, useState } from "react";
import type { ChangeEvent, DragEvent } from "react";
import { Alert, Button, Card, Image, InputNumber, Modal, Select, Space, Tag, Typography } from "antd";
import { UploadOutlined } from "@ant-design/icons";
import type { Notify, RequestFn, UploadResult } from "./utils";
import { buildLocalPreviewUrl } from "./utils";

const { Text } = Typography;

type PreviewType = "image" | "video" | "audio";
type MediaType = "image" | "video" | "audio";

type MediaRule = {
  id?: number;
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

type MediaMeta = {
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

type ValidationResult = {
  path?: string;
  valid: boolean;
  meta?: MediaMeta;
  rule?: MediaRule;
  violations?: Array<{ field: string; rule: unknown; actual: unknown }>;
  warning?: string;
  rule_missing?: boolean;
};

type TransformResult = {
  path: string;
  url?: string;
  meta?: MediaMeta;
};

type PendingTransform = {
  path: string;
  url?: string;
  meta?: MediaMeta;
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

type UploadFieldProps = {
  label: string;
  value?: string | null;
  previewUrl?: string | null;
  accept?: string;
  helper?: string;
  uploadLabel?: string;
  previewType?: PreviewType;
  mediaType?: MediaType;
  moduleKey?: string;
  draftVersionId?: number;
  operatorId?: number | null;
  request?: RequestFn;
  notify?: Notify;
  enableValidation?: boolean;
  enableSmartCompress?: boolean;
  disabled?: boolean;
  onUpload: (file: File) => Promise<UploadResult>;
  onChange: (path: string, url?: string) => void;
  onClear?: () => void;
};

const UploadField = ({
  label,
  value,
  previewUrl,
  accept,
  helper,
  uploadLabel,
  previewType,
  mediaType,
  moduleKey,
  draftVersionId,
  operatorId,
  request,
  notify,
  enableValidation = true,
  enableSmartCompress = true,
  disabled,
  onUpload,
  onChange,
  onClear
}: UploadFieldProps) => {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const validationRef = useRef<HTMLDivElement | null>(null);
  const [uploading, setUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(previewUrl ?? null);
  const [dragging, setDragging] = useState(false);
  const [validating, setValidating] = useState(false);
  const [validation, setValidation] = useState<ValidationResult | null>(null);
  const [scrollToValidation, setScrollToValidation] = useState(false);
  const [compressOpen, setCompressOpen] = useState(false);
  const [compressing, setCompressing] = useState(false);
  const [compressQuality, setCompressQuality] = useState(85);
  const [resizeWidth, setResizeWidth] = useState<number | null>(null);
  const [resizeHeight, setResizeHeight] = useState<number | null>(null);
  const [resizeMode, setResizeMode] = useState<string>("contain");
  const [targetFormat, setTargetFormat] = useState<string>("jpg");
  const [presetKey, setPresetKey] = useState<string>("rule_max");
  const [sourcePreview, setSourcePreview] = useState<string | null>(null);
  const [resultPreview, setResultPreview] = useState<string | null>(null);
  const [pendingTransform, setPendingTransform] = useState<PendingTransform | null>(null);
  const [autoPresetPending, setAutoPresetPending] = useState(false);

  useEffect(() => {
    setPreview(previewUrl ?? null);
  }, [previewUrl]);

  useEffect(() => {
    if (!value) {
      setValidation(null);
    }
    setPendingTransform(null);
    setResultPreview(null);
  }, [value]);

  const resolvedMediaType = mediaType ?? previewType;
  const canValidate = Boolean(request && moduleKey && resolvedMediaType && enableValidation);
  const canCompress = Boolean(request && moduleKey && resolvedMediaType && draftVersionId && operatorId && enableSmartCompress);

  const readMetaValue = (meta: MediaMeta | undefined, primary: keyof MediaMeta, fallback: keyof MediaMeta) => {
    const value = meta?.[primary];
    if (typeof value === "number" || typeof value === "string") {
      return value;
    }
    return meta?.[fallback];
  };

  const metaSummary = useMemo(() => {
    const meta = validation?.meta;
    const sizeBytes = Number(readMetaValue(meta, "SizeBytes", "size_bytes") || 0);
    const width = Number(readMetaValue(meta, "Width", "width") || 0);
    const height = Number(readMetaValue(meta, "Height", "height") || 0);
    const duration = Number(readMetaValue(meta, "DurationMS", "duration_ms") || 0);
    const format = String(readMetaValue(meta, "FileExt", "file_ext") || readMetaValue(meta, "Format", "format") || "-");
    return {
      sizeBytes,
      sizeKB: Math.round(sizeBytes / 1024),
      width,
      height,
      duration,
      format
    };
  }, [validation]);

  const violations = validation?.violations ?? [];
  const rule = validation?.rule;
  const ruleMissing = Boolean(validation?.rule_missing || validation?.warning === "no_rule");

  const ruleSummary = useMemo(() => {
    if (ruleMissing) {
      return "未配置规则";
    }
    if (!rule) {
      return "未获取规则";
    }
    const toNumber = (value: unknown) => {
      const parsed = Number(value);
      return Number.isFinite(parsed) ? parsed : 0;
    };
    const minW = toNumber(rule.min_width);
    const minH = toNumber(rule.min_height);
    const maxW = toNumber(rule.max_width);
    const maxH = toNumber(rule.max_height);
    const hasSizeLimit = minW > 0 || minH > 0 || maxW > 0 || maxH > 0;
    const sizeParts: string[] = [];
    if (hasSizeLimit) {
      sizeParts.push(`${minW > 0 ? minW : "-"}×${minH > 0 ? minH : "-"} ~ ${maxW > 0 ? maxW : "-"}×${maxH > 0 ? maxH : "-"}`);
    }
    if (rule.max_size_kb) {
      sizeParts.push(formatKB(rule.max_size_kb));
    }
    if (rule.ratio_width && rule.ratio_height) {
      sizeParts.push(`比例 ${rule.ratio_width}:${rule.ratio_height}`);
    }
    if (!sizeParts.length) {
      return "未设置尺寸/体积限制";
    }
    return `规则：${sizeParts.join(" · ")}`;
  }, [
    rule,
    ruleMissing,
    rule?.max_height,
    rule?.max_size_kb,
    rule?.max_width,
    rule?.min_height,
    rule?.min_width,
    rule?.ratio_height,
    rule?.ratio_width
  ]);

  const ratioConfig = useMemo(() => {
    if (!rule?.ratio_width || !rule?.ratio_height) {
      return null;
    }
    const gcd = (a: number, b: number) => {
      let x = Math.abs(a);
      let y = Math.abs(b);
      while (y) {
        const next = x % y;
        x = y;
        y = next;
      }
      return x || 1;
    };
    const baseGcd = gcd(rule.ratio_width, rule.ratio_height);
    const baseW = Math.max(1, Math.floor(rule.ratio_width / baseGcd));
    const baseH = Math.max(1, Math.floor(rule.ratio_height / baseGcd));

    let minScale = 1;
    if ((rule.min_width ?? 0) > 0) {
      minScale = Math.max(minScale, Math.ceil((rule.min_width ?? 0) / baseW));
    }
    if ((rule.min_height ?? 0) > 0) {
      minScale = Math.max(minScale, Math.ceil((rule.min_height ?? 0) / baseH));
    }

    let maxScale = Number.POSITIVE_INFINITY;
    if ((rule.max_width ?? 0) > 0) {
      maxScale = Math.min(maxScale, Math.floor((rule.max_width ?? 0) / baseW));
    }
    if ((rule.max_height ?? 0) > 0) {
      maxScale = Math.min(maxScale, Math.floor((rule.max_height ?? 0) / baseH));
    }

    if (metaSummary.width > 0 && metaSummary.height > 0) {
      const maxByOriginal = Math.min(
        Math.floor(metaSummary.width / baseW),
        Math.floor(metaSummary.height / baseH)
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
  }, [metaSummary.height, metaSummary.width, rule?.max_height, rule?.max_width, rule?.min_height, rule?.min_width, rule?.ratio_height, rule?.ratio_width]);

  const sizePresets = useMemo(() => {
    if (!rule || !metaSummary.width || !metaSummary.height) {
      return [];
    }
    const presets: Array<{ value: string; label: string; width: number; height: number }> = [];

    if (ratioConfig) {
      const { baseW, baseH, minScale, maxScale } = ratioConfig;
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
      if (metaSummary.width > 0 && nextWidth > metaSummary.width) {
        nextWidth = metaSummary.width;
      }
      if (metaSummary.height > 0 && nextHeight > metaSummary.height) {
        nextHeight = metaSummary.height;
      }
      return { width: nextWidth, height: nextHeight };
    };

    presets.push({
      value: "original",
      label: `原尺寸 ${metaSummary.width}×${metaSummary.height}`,
      width: metaSummary.width,
      height: metaSummary.height
    });

    if ((rule.max_width ?? 0) > 0 && (rule.max_height ?? 0) > 0) {
      const maxSize = clamp(rule.max_width ?? 0, rule.max_height ?? 0);
      presets.push({
        value: "rule_max",
        label: `规则最大 ${maxSize.width}×${maxSize.height}`,
        width: maxSize.width,
        height: maxSize.height
      });
    }
    if ((rule.min_width ?? 0) > 0 && (rule.min_height ?? 0) > 0) {
      const minSize = clamp(rule.min_width ?? 0, rule.min_height ?? 0);
      presets.push({
        value: "rule_min",
        label: `规则最小 ${minSize.width}×${minSize.height}`,
        width: minSize.width,
        height: minSize.height
      });
    }
    if ((rule.max_width ?? 0) > 0 && (rule.min_width ?? 0) > 0) {
      const midWidth = Math.round(((rule.max_width ?? 0) + (rule.min_width ?? 0)) / 2);
      const midHeight = Math.round(((rule.max_height ?? 0) + (rule.min_height ?? 0)) / 2);
      const midSize = clamp(midWidth, midHeight);
      presets.push({
        value: "rule_mid",
        label: `推荐尺寸 ${midSize.width}×${midSize.height}`,
        width: midSize.width,
        height: midSize.height
      });
    }

    return presets;
  }, [metaSummary.height, metaSummary.width, ratioConfig, rule]);

  const hasMinViolation = violations.some(
    (item) => item.field === "width_min" || item.field === "height_min"
  );
  const hasRatioViolation = violations.some((item) => item.field === "ratio");

  const violationSummary = useMemo(() => {
    if (!violations.length) {
      return "";
    }
    const mapping: Record<string, string> = {
      size: "大小超限",
      width_min: "宽度不足",
      width_max: "宽度超限",
      height_min: "高度不足",
      height_max: "高度超限",
      format: "格式不支持",
      ratio: "比例不匹配"
    };
    return violations
      .map((item) => mapping[item.field] || item.field)
      .join("、");
  }, [violations]);

  const runValidation = async (path: string) => {
    if (!canValidate || !request || !moduleKey || !resolvedMediaType) {
      return null;
    }
    setValidating(true);
    try {
      const res = await request<ValidationResult>("/api/media/validate", {
        method: "POST",
        body: JSON.stringify({
          module_key: moduleKey,
          media_type: resolvedMediaType,
          path
        })
      });
      setValidation(res);
      setScrollToValidation(true);
      if (res.warning === "no_rule") {
        notify?.warning("未配置规则，默认通过");
        return res;
      }
      if (!res.valid) {
        notify?.warning("上传文件未通过规则校验");
      }
      return res;
    } catch (error) {
      setValidation(null);
      notify?.error(error instanceof Error ? error.message : "校验失败");
      return null;
    } finally {
      setValidating(false);
    }
  };

  const applyPreset = (key: string) => {
    const preset = sizePresets.find((item) => item.value === key);
    if (!preset) {
      return;
    }
    let nextWidth = preset.width;
    let nextHeight = preset.height;

    if (rule?.ratio_width && rule?.ratio_height && nextWidth > 0 && nextHeight > 0) {
      const ratio = rule.ratio_width / rule.ratio_height;
      const fitHeight = Math.round(nextWidth / ratio);
      if (fitHeight <= nextHeight) {
        nextHeight = fitHeight;
      } else {
        nextWidth = Math.round(nextHeight * ratio);
      }
    }

    if (metaSummary.width > 0 && metaSummary.height > 0 && nextWidth > 0 && nextHeight > 0) {
      const scale = Math.min(metaSummary.width / nextWidth, metaSummary.height / nextHeight, 1);
      nextWidth = Math.max(1, Math.round(nextWidth * scale));
      nextHeight = Math.max(1, Math.round(nextHeight * scale));
    }

    setResizeWidth(nextWidth);
    setResizeHeight(nextHeight);
  };

  const handleTrigger = () => {
    if (disabled || uploading) {
      return;
    }
    inputRef.current?.click();
  };

  const handleChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }
    setUploading(true);
    try {
      const result = await onUpload(file);
      const nextPreview = result.url ?? buildLocalPreviewUrl(result.path);
      setPreview(nextPreview ?? null);
      onChange(result.path, nextPreview ?? undefined);
      setValidation(null);
      await runValidation(result.path);
    } catch (error) {
      setPreview(previewUrl ?? null);
    } finally {
      setUploading(false);
    }
  };

  const handleDrop = async (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    if (disabled || uploading) {
      return;
    }
    setDragging(false);
    const file = event.dataTransfer.files?.[0];
    if (!file) {
      return;
    }
    setUploading(true);
    try {
      const result = await onUpload(file);
      const nextPreview = result.url ?? buildLocalPreviewUrl(result.path);
      setPreview(nextPreview ?? null);
      onChange(result.path, nextPreview ?? undefined);
      setValidation(null);
      await runValidation(result.path);
    } catch (error) {
      setPreview(previewUrl ?? null);
    } finally {
      setUploading(false);
    }
  };

  const handleClear = () => {
    if (disabled || uploading) {
      return;
    }
    setPreview(null);
    setValidation(null);
    onClear?.();
  };

  const handleResizeWidth = (value?: number | null) => {
    const next = value ?? null;
    if (!next) {
      setResizeWidth(null);
      return;
    }
    if (ratioConfig) {
      let scale = Math.round(next / ratioConfig.baseW);
      if (scale < ratioConfig.minScale) {
        scale = ratioConfig.minScale;
      }
      if (ratioConfig.maxScale > 0) {
        scale = Math.min(scale, ratioConfig.maxScale);
      }
      const width = ratioConfig.baseW * scale;
      const height = ratioConfig.baseH * scale;
      setResizeWidth(width);
      setResizeHeight(height);
      return;
    }
    setResizeWidth(next);
    if (resizeMode !== "fill" && metaSummary.width > 0 && metaSummary.height > 0) {
      const computed = Math.round((next * metaSummary.height) / metaSummary.width);
      setResizeHeight(computed);
    }
  };

  const handleResizeHeight = (value?: number | null) => {
    const next = value ?? null;
    if (!next) {
      setResizeHeight(null);
      return;
    }
    if (ratioConfig) {
      let scale = Math.round(next / ratioConfig.baseH);
      if (scale < ratioConfig.minScale) {
        scale = ratioConfig.minScale;
      }
      if (ratioConfig.maxScale > 0) {
        scale = Math.min(scale, ratioConfig.maxScale);
      }
      const width = ratioConfig.baseW * scale;
      const height = ratioConfig.baseH * scale;
      setResizeWidth(width);
      setResizeHeight(height);
      return;
    }
    setResizeHeight(next);
    if (resizeMode !== "fill" && metaSummary.width > 0 && metaSummary.height > 0) {
      const computed = Math.round((next * metaSummary.width) / metaSummary.height);
      setResizeWidth(computed);
    }
  };

  useEffect(() => {
    if (!compressOpen || resolvedMediaType !== "image") {
      return;
    }
    if (ratioConfig && hasRatioViolation && resizeMode === "contain") {
      setResizeMode("cover");
    }
  }, [compressOpen, hasRatioViolation, ratioConfig, resizeMode, resolvedMediaType]);

  const openSmartCompress = async () => {
    setSourcePreview(preview ?? null);
    setResultPreview(null);
    setPendingTransform(null);
    if (resolvedMediaType === "image") {
      const defaultQuality = rule?.compress_quality && rule.compress_quality > 0 ? rule.compress_quality : 85;
      setCompressQuality(defaultQuality);
      const defaultTarget = rule?.target_format ? rule.target_format : "jpg";
      setTargetFormat(defaultTarget);
      const defaultResizeMode =
        rule?.resize_mode
          ? rule.resize_mode
          : rule?.ratio_width && rule?.ratio_height
            ? "cover"
            : "contain";
      setResizeMode(defaultResizeMode);
      setPresetKey("");
      setAutoPresetPending(true);
    }
    if (value && canValidate && (!validation || validation.path !== value)) {
      await runValidation(value);
    }
    setCompressOpen(true);
  };

  const handleSmartCompress = async () => {
    if (!canCompress || !request || !moduleKey || !resolvedMediaType || !draftVersionId || !operatorId || !value) {
      notify?.warning("缺少压缩所需的上下文信息");
      return;
    }
    setCompressing(true);
    try {
      const ruleOverride: Record<string, unknown> = {};
      if (resolvedMediaType === "image") {
        if (resizeWidth && resizeHeight) {
          let targetWidth = resizeWidth;
          let targetHeight = resizeHeight;
          if (resizeMode !== "fill" && metaSummary.width > 0 && targetWidth > metaSummary.width) {
            targetWidth = metaSummary.width;
          }
          if (resizeMode !== "fill" && metaSummary.height > 0 && targetHeight > metaSummary.height) {
            targetHeight = metaSummary.height;
          }
          ruleOverride.max_width = targetWidth;
          ruleOverride.max_height = targetHeight;
        }
        ruleOverride.resize_mode = resizeMode === "contain" ? "" : resizeMode;
        if (targetFormat && targetFormat !== "keep") {
          ruleOverride.target_format = targetFormat;
        }
        ruleOverride.compress_quality = compressQuality;
        if (rule?.allow_formats) {
          ruleOverride.allow_formats = rule.allow_formats;
        }
      } else {
        ruleOverride.resize_mode = "lossless";
        ruleOverride.compress_quality = 0;
      }

      const res = await request<TransformResult>("/api/media/transform", {
        method: "POST",
        body: JSON.stringify({
          draft_version_id: draftVersionId,
          module_key: moduleKey,
          media_type: resolvedMediaType,
          path: value,
          rule_id: 0,
          operator_id: operatorId,
          rule: ruleOverride
        })
      });
      const nextPreview = res.url ?? buildLocalPreviewUrl(res.path);
      setResultPreview(nextPreview ?? null);
      setPendingTransform({ path: res.path, url: res.url, meta: res.meta });
      notify?.success("压缩完成，请确认应用结果");
    } catch (error) {
      notify?.error(error instanceof Error ? error.message : "压缩失败");
    } finally {
      setCompressing(false);
    }
  };

  const applyTransform = () => {
    if (!pendingTransform) {
      notify?.warning("暂无可应用的压缩结果");
      return;
    }
    const nextPreview = pendingTransform.url ?? buildLocalPreviewUrl(pendingTransform.path);
    setPreview(nextPreview ?? null);
    onChange(pendingTransform.path, nextPreview ?? undefined);
    setValidation({
      valid: true,
      meta: pendingTransform.meta,
      rule: rule
    });
    setPendingTransform(null);
    setCompressOpen(false);
    notify?.success("已应用压缩结果");
  };

  useEffect(() => {
    if (!compressOpen || !autoPresetPending) {
      return;
    }
    if (!sizePresets.length) {
      return;
    }
    const defaultPreset =
      sizePresets.find((item) => item.value.startsWith("ratio_"))?.value ??
      sizePresets.find((item) => item.value === "rule_max")?.value ??
      sizePresets.find((item) => item.value === "rule_min")?.value ??
      sizePresets[0]?.value ??
      "";
    if (!defaultPreset) {
      return;
    }
    setPresetKey(defaultPreset);
    applyPreset(defaultPreset);
    setAutoPresetPending(false);
  }, [autoPresetPending, compressOpen, sizePresets]);

  useEffect(() => {
    if (!scrollToValidation || !validation) {
      return;
    }
    const timer = window.setTimeout(() => {
      validationRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
      setScrollToValidation(false);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [scrollToValidation, validation]);

  const renderPreview = () => {
    if (!preview) {
      return null;
    }
    if (previewType === "video") {
      return (
        <video
          src={preview}
          controls
          style={{ width: "100%", maxHeight: 240, objectFit: "contain", borderRadius: 12 }}
        />
      );
    }
    if (previewType === "audio") {
      return <audio src={preview} controls style={{ width: "100%" }} />;
    }
    return (
      <Image
        src={preview}
        alt={label}
        style={{ width: "100%", maxHeight: 240, borderRadius: 12, objectFit: "contain" }}
      />
    );
  };

  return (
    <Space direction="vertical" size={8} style={{ width: "100%" }}>
      <Space direction="vertical" size={4} style={{ width: "100%" }}>
        <Text strong>{label}</Text>
        {helper ? <Text type="secondary">{helper}</Text> : null}
      </Space>
      <div
        onDragOver={(event) => {
          event.preventDefault();
          if (!disabled && !uploading) {
            setDragging(true);
          }
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={handleDrop}
        onClick={handleTrigger}
        style={{
          border: `1px dashed ${dragging ? "#1677ff" : "#d9d9d9"}`,
          borderRadius: 12,
          padding: "12px 16px",
          textAlign: "center",
          background: dragging ? "rgba(22, 119, 255, 0.08)" : "#fafafa",
          color: disabled ? "#bfbfbf" : "#595959",
          cursor: disabled || uploading ? "not-allowed" : "pointer"
        }}
      >
        <Text type="secondary">{disabled ? "上传已禁用" : "拖拽文件到此处，或点击选择"}</Text>
      </div>
      {renderPreview()}
      <Space wrap>
        <Button icon={<UploadOutlined />} onClick={handleTrigger} loading={uploading} disabled={disabled}>
          {uploadLabel || "上传文件"}
        </Button>
        {value ? (
          <Text type="secondary" style={{ maxWidth: 240 }} ellipsis>
            {value}
          </Text>
        ) : null}
        {value && onClear ? (
          <Button type="link" onClick={handleClear} disabled={disabled || uploading}>
            清空
          </Button>
        ) : null}
        {value && canValidate ? (
          <Button onClick={() => runValidation(value)} loading={validating} disabled={disabled || uploading}>
            重新校验
          </Button>
        ) : null}
        {value && canCompress ? (
          <Button type="primary" ghost onClick={openSmartCompress} disabled={disabled || uploading}>
            智能压缩
          </Button>
        ) : null}
        {validation ? (
          <Tag color={ruleMissing ? "blue" : validation.valid ? "green" : "gold"}>
            {ruleMissing ? "未配置规则" : validation.valid ? "合规" : "不合规"}
          </Tag>
        ) : null}
        {validation && violationSummary ? (
          <Text type="danger" style={{ maxWidth: 220 }} ellipsis>
            {violationSummary}
          </Text>
        ) : null}
      </Space>
      {validation ? (
        <div ref={validationRef}>
          <Alert
            showIcon
            type={ruleMissing ? "info" : validation.valid ? "success" : "warning"}
            message={
              <Space>
                <Text>{ruleMissing ? "未配置规则，默认通过" : validation.valid ? "已通过规则校验" : "未通过规则校验"}</Text>
                <Tag color={ruleMissing ? "blue" : validation.valid ? "green" : "gold"}>
                  {ruleMissing ? "未配置规则" : validation.valid ? "合规" : "不合规"}
                </Tag>
              </Space>
            }
            description={
              <Space direction="vertical" size={4}>
                <Text type="secondary">
                  大小 {formatBytes(metaSummary.sizeBytes)} · 尺寸 {metaSummary.width}×{metaSummary.height} · 格式 {metaSummary.format}
                </Text>
                {rule?.ratio_width && rule?.ratio_height ? (
                  <Text type="secondary">
                    目标比例 {rule.ratio_width}:{rule.ratio_height}
                  </Text>
                ) : null}
                {violationSummary ? <Text type="danger">{violationSummary}</Text> : null}
                {hasRatioViolation ? <Text type="secondary">比例不匹配，请选择合适的调整方式。</Text> : null}
                {hasMinViolation ? <Text type="secondary">尺寸小于最小限制，智能压缩不会放大。</Text> : null}
                {ruleMissing ? <Text type="secondary">请在规则配置中补充该模块的校验规则。</Text> : null}
              </Space>
            }
          />
        </div>
      ) : null}
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        onChange={handleChange}
        style={{ display: "none" }}
      />
      <Modal
        title="智能压缩"
        open={compressOpen}
        onCancel={() => setCompressOpen(false)}
        onOk={handleSmartCompress}
        okText="开始压缩"
        cancelText="取消"
        okButtonProps={{ loading: compressing }}
        destroyOnClose
      >
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          {resolvedMediaType === "image" ? (
            <>
              {ratioConfig && sizePresets.length < 5 ? (
                <Text type="secondary">可用推荐尺寸不足 5 个，原图尺寸或规则范围受限。</Text>
              ) : null}
              <Text type="secondary">{ruleSummary}</Text>
              {sizePresets.length ? (
                <Select
                  value={presetKey || undefined}
                  placeholder="推荐尺寸"
                  options={sizePresets.map((item) => ({ value: item.value, label: item.label }))}
                  onChange={(value) => {
                    setPresetKey(value);
                    applyPreset(value);
                  }}
                  style={{ width: "100%" }}
                />
              ) : null}
              <Space>
                <Text type="secondary">目标尺寸</Text>
                <InputNumber min={1} value={resizeWidth ?? undefined} onChange={handleResizeWidth} />
                <Text type="secondary">×</Text>
                <InputNumber min={1} value={resizeHeight ?? undefined} onChange={handleResizeHeight} />
              </Space>
              {ratioConfig ? <Text type="secondary">按比例自动换算尺寸，严格匹配规则。</Text> : null}
              <Space>
                <Text type="secondary">调整方式</Text>
                <Select
                  value={resizeMode}
                  onChange={setResizeMode}
                  options={[
                    { value: "contain", label: "等比缩放（不裁剪）", disabled: !!ratioConfig && hasRatioViolation },
                    { value: "cover", label: "裁剪填充（保持比例）" },
                    { value: "fill", label: "拉伸到目标尺寸" }
                  ]}
                  style={{ minWidth: 220 }}
                />
              </Space>
              <Space>
                <Text type="secondary">质量(%)</Text>
                <InputNumber min={1} max={100} value={compressQuality} onChange={(value) => setCompressQuality(value ?? 85)} />
              </Space>
              <Select
                value={targetFormat}
                onChange={setTargetFormat}
                options={[
                  { value: "jpg", label: "转为 JPG" },
                  { value: "keep", label: "保持原格式" }
                ]}
                style={{ width: "100%" }}
              />
            </>
          ) : (
            <Text type="secondary">将以无损方式压缩并保持原有分辨率。</Text>
          )}
          <Space direction="vertical" size={4}>
            <Text type="secondary">压缩完成后需手动确认应用到表单。</Text>
            <Button type="primary" onClick={applyTransform} disabled={!pendingTransform}>
              应用压缩结果
            </Button>
          </Space>
          {resolvedMediaType === "image" && (sourcePreview || resultPreview) ? (
            <Space size={16} wrap>
              {sourcePreview ? (
                <Card size="small" title="原图预览" style={{ width: 240 }}>
                  <Image src={sourcePreview} alt="source" style={{ width: "100%" }} />
                </Card>
              ) : null}
              {resultPreview ? (
                <Card size="small" title="压缩后预览" style={{ width: 240 }}>
                  <Image src={resultPreview} alt="result" style={{ width: "100%" }} />
                </Card>
              ) : null}
            </Space>
          ) : null}
          {resolvedMediaType === "video" && (sourcePreview || resultPreview) ? (
            <Space size={16} wrap>
              {sourcePreview ? (
                <Card size="small" title="原视频预览" style={{ width: 320 }}>
                  <video src={sourcePreview} controls style={{ width: "100%", maxHeight: 200, borderRadius: 8 }} />
                </Card>
              ) : null}
              {resultPreview ? (
                <Card size="small" title="压缩后预览" style={{ width: 320 }}>
                  <video src={resultPreview} controls style={{ width: "100%", maxHeight: 200, borderRadius: 8 }} />
                </Card>
              ) : null}
            </Space>
          ) : null}
        </Space>
      </Modal>
    </Space>
  );
};

export default UploadField;
