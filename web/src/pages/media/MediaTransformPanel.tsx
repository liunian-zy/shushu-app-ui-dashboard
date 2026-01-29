import { useEffect, useMemo, useState } from "react";
import { Button, Card, Descriptions, Empty, Input, InputNumber, Select, Space, Table, Tag, Typography, message } from "antd";
import UploadField from "../content/UploadField";
import type { Notify, RequestFn, UploadFn } from "../content/utils";

const { Text } = Typography;

type DraftVersion = {
  id: number;
  location_name?: string | null;
  app_version_name?: string | null;
};

type MediaRule = {
  id: number;
  module_key?: string | null;
  media_type?: string | null;
  max_size_kb?: number | null;
  max_width?: number | null;
  max_height?: number | null;
  min_width?: number | null;
  min_height?: number | null;
  max_duration_ms?: number | null;
  allow_formats?: string | null;
  status?: number | null;
};

type MediaTransformPanelProps = {
  request: RequestFn;
  uploadFile: UploadFn;
  notify: Notify;
  operatorId: number | null;
};

type MediaMeta = {
  SizeBytes?: number;
  Width?: number;
  Height?: number;
  DurationMS?: number;
  Format?: string;
  size_bytes?: number;
  width?: number;
  height?: number;
  duration_ms?: number;
  format?: string;
};

type ValidationResult = {
  path: string;
  valid: boolean;
  meta?: MediaMeta;
  violations?: Array<{ field: string; rule: unknown; actual: unknown }>;
  warning?: string;
  rule_missing?: boolean;
};

type TransformResult = {
  path: string;
  asset_id: number;
  version_id: number;
  url?: string;
  meta?: MediaMeta;
};

type RuleOverride = {
  max_size_kb?: number;
  min_width?: number;
  max_width?: number;
  min_height?: number;
  max_height?: number;
  min_duration_ms?: number;
  max_duration_ms?: number;
  allow_formats?: string;
  resize_mode?: string;
  target_format?: string;
  compress_quality?: number;
};

const mediaTypeOptions = [
  { label: "图片", value: "image" },
  { label: "视频", value: "video" },
  { label: "音频", value: "audio" }
];

const presetOptions = [
  { value: "custom", label: "自定义规则" },
  { value: "image_jpg_lossy_original", label: "图片 JPG 有损（原尺寸 / 85%）", mediaType: "image" },
  { value: "image_jpg_lossy_resize", label: "图片 JPG 有损（指定尺寸）", mediaType: "image" },
  { value: "video_lossless", label: "视频无损压缩", mediaType: "video" },
  { value: "audio_lossless", label: "音频无损压缩", mediaType: "audio" }
];

const buildLocalUrl = (path: string) => {
  if (!path.startsWith("local://")) {
    return path;
  }
  const trimmed = path.replace(/^local:\/\//, "").replace(/^\/+/, "");
  return `/api/local-files/${trimmed}`;
};

const readMetaValue = (meta: MediaMeta | undefined, primary: keyof MediaMeta, fallback: keyof MediaMeta) => {
  const value = meta?.[primary];
  if (typeof value === "number" || typeof value === "string") {
    return value;
  }
  return meta?.[fallback];
};

const formatMeta = (meta?: MediaMeta) => {
  const sizeBytes = Number(readMetaValue(meta, "SizeBytes", "size_bytes") || 0);
  const width = Number(readMetaValue(meta, "Width", "width") || 0);
  const height = Number(readMetaValue(meta, "Height", "height") || 0);
  const duration = Number(readMetaValue(meta, "DurationMS", "duration_ms") || 0);
  const format = String(readMetaValue(meta, "Format", "format") || "-");
  return {
    sizeKB: Math.round(sizeBytes / 1024),
    dimension: `${width} x ${height}`,
    duration,
    format
  };
};

const MediaTransformPanel = ({ request, uploadFile, notify, operatorId }: MediaTransformPanelProps) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [versions, setVersions] = useState<DraftVersion[]>([]);
  const [versionLoading, setVersionLoading] = useState(false);
  const [selectedVersionId, setSelectedVersionId] = useState<number | null>(null);
  const [presetKey, setPresetKey] = useState("custom");
  const [moduleKey, setModuleKey] = useState("");
  const [mediaType, setMediaType] = useState<string | null>(null);
  const [rules, setRules] = useState<MediaRule[]>([]);
  const [ruleLoading, setRuleLoading] = useState(false);
  const [selectedRuleId, setSelectedRuleId] = useState<number | null>(null);
  const [qualityPercent, setQualityPercent] = useState(85);
  const [resizeWidth, setResizeWidth] = useState<number | null>(null);
  const [resizeHeight, setResizeHeight] = useState<number | null>(null);
  const [filePath, setFilePath] = useState<string | null>(null);
  const [fileUrl, setFileUrl] = useState<string | null>(null);
  const [validating, setValidating] = useState(false);
  const [transforming, setTransforming] = useState(false);
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null);
  const [transformResult, setTransformResult] = useState<TransformResult | null>(null);

  const loadVersions = async () => {
    setVersionLoading(true);
    try {
      const res = await request<{ data: DraftVersion[] }>("/api/draft/version-names");
      setVersions(res.data || []);
      if (!selectedVersionId && res.data?.length) {
        setSelectedVersionId(res.data[0].id);
      }
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取版本失败");
    } finally {
      setVersionLoading(false);
    }
  };

  const loadRules = async () => {
    setRuleLoading(true);
    try {
      const params = new URLSearchParams();
      if (moduleKey.trim()) {
        params.set("module_key", moduleKey.trim());
      }
      if (mediaType) {
        params.set("media_type", mediaType);
      }
      const query = params.toString();
      const url = query ? `/api/media/rules?${query}` : "/api/media/rules";
      const res = await request<{ data: MediaRule[] }>(url);
      setRules(res.data || []);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取媒体规则失败");
      setRules([]);
    } finally {
      setRuleLoading(false);
    }
  };

  useEffect(() => {
    void loadVersions();
  }, []);

  useEffect(() => {
    if (presetKey !== "custom") {
      return;
    }
    if (moduleKey.trim() && mediaType) {
      void loadRules();
    } else {
      setRules([]);
      setSelectedRuleId(null);
    }
  }, [moduleKey, mediaType, presetKey]);

  useEffect(() => {
    if (presetKey === "custom") {
      return;
    }
    const preset = presetOptions.find((item) => item.value === presetKey);
    if (preset?.mediaType) {
      setMediaType(preset.mediaType);
    }
    if (!moduleKey.trim()) {
      setModuleKey("media_tool");
    }
    setSelectedRuleId(null);
  }, [presetKey]);

  const buildPresetRule = (): RuleOverride | null => {
    if (presetKey === "custom") {
      return null;
    }
    if (presetKey === "image_jpg_lossy_original") {
      return {
        allow_formats: "jpg,jpeg",
        target_format: "jpg",
        compress_quality: qualityPercent
      };
    }
    if (presetKey === "image_jpg_lossy_resize") {
      return {
        allow_formats: "jpg,jpeg",
        target_format: "jpg",
        compress_quality: qualityPercent,
        resize_mode: "fill",
        max_width: resizeWidth ?? 0,
        max_height: resizeHeight ?? 0
      };
    }
    if (presetKey === "video_lossless") {
      return {
        resize_mode: "lossless",
        compress_quality: 0
      };
    }
    if (presetKey === "audio_lossless") {
      return {
        resize_mode: "lossless",
        compress_quality: 0
      };
    }
    return null;
  };

  const handleValidate = async () => {
    if (!moduleKey.trim() || !mediaType) {
      messageApi.warning("请先填写模块与媒体类型");
      return;
    }
    if (!filePath) {
      messageApi.warning("请先上传媒体文件");
      return;
    }
    if (presetKey === "image_jpg_lossy_resize" && (!resizeWidth || !resizeHeight)) {
      messageApi.warning("请填写图片目标尺寸");
      return;
    }
    setValidating(true);
    try {
      const ruleOverride = buildPresetRule();
      const res = await request<ValidationResult>("/api/media/validate", {
        method: "POST",
        body: JSON.stringify({
          module_key: moduleKey.trim(),
          media_type: mediaType,
          path: filePath,
          rule_id: presetKey === "custom" ? selectedRuleId ?? 0 : 0,
          rule: ruleOverride ?? undefined
        })
      });
      setValidationResult(res);
      if (res.warning === "no_rule") {
        messageApi.warning("未配置规则，已默认通过校验");
      } else {
        messageApi.success("校验完成");
      }
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "校验失败");
    } finally {
      setValidating(false);
    }
  };

  const handleTransform = async () => {
    if (!selectedVersionId) {
      messageApi.warning("请选择版本");
      return;
    }
    if (!moduleKey.trim() || !mediaType) {
      messageApi.warning("请先填写模块与媒体类型");
      return;
    }
    if (!filePath) {
      messageApi.warning("请先上传媒体文件");
      return;
    }
    if (presetKey === "image_jpg_lossy_resize" && (!resizeWidth || !resizeHeight)) {
      messageApi.warning("请填写图片目标尺寸");
      return;
    }
    setTransforming(true);
    try {
      const ruleOverride = buildPresetRule();
      const res = await request<TransformResult>("/api/media/transform", {
        method: "POST",
        body: JSON.stringify({
          draft_version_id: selectedVersionId,
          module_key: moduleKey.trim(),
          media_type: mediaType,
          path: filePath,
          rule_id: presetKey === "custom" ? selectedRuleId ?? 0 : 0,
          operator_id: operatorId ?? 0,
          rule: ruleOverride ?? undefined
        })
      });
      setTransformResult(res);
      messageApi.success("压缩/转码完成");
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "压缩失败");
    } finally {
      setTransforming(false);
    }
  };

  const handleReset = () => {
    setFilePath(null);
    setFileUrl(null);
    setValidationResult(null);
    setTransformResult(null);
  };

  const versionOptions = useMemo(
    () =>
      versions.map((item) => ({
        value: item.id,
        label: `${item.location_name || "未命名景区"} / ${item.app_version_name || "未生成版本"}`
      })),
    [versions]
  );

  const ruleOptions = useMemo(() => {
    return rules.map((item) => ({
      value: item.id,
      label: `${item.module_key || "模块"} / ${item.media_type || "类型"} #${item.id}`
    }));
  }, [rules]);

  const violations = validationResult?.violations || [];

  const metaSummary = formatMeta(validationResult?.meta);
  const transformMetaSummary = formatMeta(transformResult?.meta);

  const outputPreviewUrl =
    transformResult?.url ??
    (transformResult?.path && transformResult.path.startsWith("local://") ? buildLocalUrl(transformResult.path) : null);

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">上传素材后进行规则校验或压缩/转码，支持本地文件路径。</Text>
          <Space wrap>
            <Select
              style={{ minWidth: 260 }}
              placeholder="选择版本（用于写入媒体资产记录）"
              options={versionOptions}
              loading={versionLoading}
              value={selectedVersionId ?? undefined}
              onChange={(value) => setSelectedVersionId(value)}
              allowClear
            />
            <Select
              style={{ minWidth: 220 }}
              placeholder="选择压缩预设"
              options={presetOptions}
              value={presetKey}
              onChange={(value) => setPresetKey(value)}
            />
            <Input
              style={{ minWidth: 200 }}
              placeholder="模块 Key，例如 banners"
              value={moduleKey}
              onChange={(event) => setModuleKey(event.target.value)}
            />
            <Select
              style={{ minWidth: 140 }}
              placeholder="媒体类型"
              options={mediaTypeOptions}
              value={mediaType ?? undefined}
              onChange={(value) => setMediaType(value)}
              allowClear
            />
            <Select
              style={{ minWidth: 220 }}
              placeholder="选择规则（可选）"
              options={ruleOptions}
              value={selectedRuleId ?? undefined}
              loading={ruleLoading}
              onChange={(value) => setSelectedRuleId(value)}
              allowClear
              disabled={presetKey !== "custom" || !ruleOptions.length}
            />
            <Button onClick={loadRules} loading={ruleLoading}>
              刷新规则
            </Button>
          </Space>
          {presetKey !== "custom" ? (
            <Space wrap>
              {presetKey.startsWith("image_jpg_lossy") ? (
                <>
                  <Space>
                    <Text type="secondary">质量(%)</Text>
                    <InputNumber min={1} max={100} value={qualityPercent} onChange={(value) => setQualityPercent(value ?? 85)} />
                  </Space>
                  {presetKey === "image_jpg_lossy_resize" ? (
                    <Space>
                      <Text type="secondary">目标尺寸</Text>
                      <InputNumber min={1} max={10000} value={resizeWidth ?? undefined} onChange={(value) => setResizeWidth(value ?? null)} />
                      <Text type="secondary">×</Text>
                      <InputNumber min={1} max={10000} value={resizeHeight ?? undefined} onChange={(value) => setResizeHeight(value ?? null)} />
                    </Space>
                  ) : null}
                </>
              ) : null}
              {presetKey === "video_lossless" ? <Text type="secondary">无损转码（CRF=0）</Text> : null}
              {presetKey === "audio_lossless" ? <Text type="secondary">无损转封装（音频流拷贝）</Text> : null}
            </Space>
          ) : (
            <Text type="secondary">自定义规则需在「规则配置」中创建。</Text>
          )}
        </Space>
      </Card>

      <Card style={{ borderRadius: 20 }}>
        <UploadField
          label="待处理媒体"
          helper="上传文件后可校验规则或执行压缩。"
          accept={
            mediaType === "image"
              ? "image/png,image/jpeg"
              : mediaType === "video"
                ? "video/mp4,video/x-m4v"
                : mediaType === "audio"
                  ? "audio/*"
                  : ""
          }
          value={filePath}
          previewUrl={fileUrl ?? undefined}
          previewType={
            mediaType === "video" ? "video" : mediaType === "audio" ? "audio" : mediaType === "image" ? "image" : "image"
          }
          enableValidation={false}
          enableSmartCompress={false}
          onUpload={(file) => uploadFile(file, moduleKey || "media-tool", selectedVersionId ?? 0)}
          onChange={(path, url) => {
            setFilePath(path);
            setFileUrl(url ?? buildLocalUrl(path));
            setValidationResult(null);
            setTransformResult(null);
          }}
          onClear={handleReset}
        />
      </Card>

      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Space wrap>
            <Button onClick={handleValidate} loading={validating}>
              校验规则
            </Button>
            <Button type="primary" onClick={handleTransform} loading={transforming}>
              压缩/转码
            </Button>
            <Button onClick={handleReset} disabled={!filePath && !validationResult && !transformResult}>
              重置
            </Button>
          </Space>

          {validationResult ? (
            <Space direction="vertical" size={12} style={{ width: "100%" }}>
              <Space>
                <Text strong>校验结果</Text>
                {validationResult.warning === "no_rule" ? (
                  <Tag color="blue">未配置规则</Tag>
                ) : validationResult.valid ? (
                  <Tag color="green">合规</Tag>
                ) : (
                  <Tag color="red">不合规</Tag>
                )}
              </Space>
              <Descriptions size="small" column={2}>
                <Descriptions.Item label="大小(KB)">{metaSummary.sizeKB}</Descriptions.Item>
                <Descriptions.Item label="尺寸">{metaSummary.dimension}</Descriptions.Item>
                <Descriptions.Item label="时长(ms)">{metaSummary.duration}</Descriptions.Item>
                <Descriptions.Item label="格式">{metaSummary.format}</Descriptions.Item>
              </Descriptions>
              {violations.length ? (
                <Table
                  rowKey={(_, index) => `violation-${index}`}
                  size="small"
                  pagination={false}
                  dataSource={violations}
                  columns={[
                    { title: "字段", dataIndex: "field", key: "field" },
                    {
                      title: "规则",
                      dataIndex: "rule",
                      key: "rule",
                      render: (value: unknown) => <Text type="secondary">{JSON.stringify(value)}</Text>
                    },
                    {
                      title: "实际值",
                      dataIndex: "actual",
                      key: "actual",
                      render: (value: unknown) => <Text type="secondary">{JSON.stringify(value)}</Text>
                    }
                  ]}
                />
              ) : (
                <Empty description="暂无违规项" />
              )}
            </Space>
          ) : null}

          {transformResult ? (
            <Space direction="vertical" size={12} style={{ width: "100%" }}>
              <Text strong>压缩结果</Text>
              <Descriptions size="small" column={2}>
                <Descriptions.Item label="资产ID">{transformResult.asset_id}</Descriptions.Item>
                <Descriptions.Item label="版本ID">{transformResult.version_id}</Descriptions.Item>
                <Descriptions.Item label="输出路径">{transformResult.path}</Descriptions.Item>
                <Descriptions.Item label="格式">{transformMetaSummary.format}</Descriptions.Item>
                <Descriptions.Item label="大小(KB)">{transformMetaSummary.sizeKB}</Descriptions.Item>
                <Descriptions.Item label="尺寸">{transformMetaSummary.dimension}</Descriptions.Item>
              </Descriptions>
              {outputPreviewUrl && (mediaType === "image" || mediaType === "video" || mediaType === "audio") ? (
                <Card size="small" style={{ borderRadius: 12 }}>
                  {mediaType === "video" ? (
                    <video src={outputPreviewUrl} controls style={{ width: "100%" }} />
                  ) : mediaType === "audio" ? (
                    <audio src={outputPreviewUrl} controls style={{ width: "100%" }} />
                  ) : (
                    <img src={outputPreviewUrl} alt="transform" style={{ width: "100%", borderRadius: 8 }} />
                  )}
                </Card>
              ) : null}
            </Space>
          ) : null}
        </Space>
      </Card>
    </Space>
  );
};

export default MediaTransformPanel;
