import { useEffect, useRef, useState } from "react";
import {
  Button,
  Card,
  Col,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message
} from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";

const { Text } = Typography;

type MediaRule = {
  id: number;
  module_key?: string | null;
  media_type?: string | null;
  max_size_kb?: number | null;
  min_width?: number | null;
  max_width?: number | null;
  min_height?: number | null;
  max_height?: number | null;
  ratio_width?: number | null;
  ratio_height?: number | null;
  min_duration_ms?: number | null;
  max_duration_ms?: number | null;
  allow_formats?: string | null;
  resize_mode?: string | null;
  target_format?: string | null;
  compress_quality?: number | null;
  status?: number | null;
};

type MediaRulePanelProps = {
  request: <T>(path: string, options?: RequestInit) => Promise<T>;
};

type RuleFormValues = {
  module_key?: string;
  media_type?: string;
  max_size_kb?: number;
  min_width?: number;
  max_width?: number;
  min_height?: number;
  max_height?: number;
  ratio_width?: number;
  ratio_height?: number;
  min_duration_ms?: number;
  max_duration_ms?: number;
  allow_formats?: string;
  resize_mode?: string;
  target_format?: string;
  compress_quality?: number;
  status?: number;
};

const mediaTypes = [
  { value: "image", label: "图片" },
  { value: "video", label: "视频" },
  { value: "audio", label: "音频" }
];

const moduleKeyOptions = [
  { value: "banners:left_top", label: "轮播图-左上" },
  { value: "banners:left_bottom", label: "轮播图-左下" },
  { value: "banners:right", label: "轮播图-右侧" },
  { value: "identities", label: "身份信息" },
  { value: "scenes", label: "场景图片" },
  { value: "app_ui_fields:print_wait", label: "打印中视频" },
  { value: "config_extra_steps", label: "额外配置" },
  { value: "clothes_categories", label: "服饰偏好" },
  { value: "photo_hobbies", label: "拍摄偏好" },
  { value: "identity-templates", label: "身份模板" }
];

const imageFormatDefault = "jpg,jpeg,png";
const videoFormatDefault = "mp4,m4v";

const rulePresets = [
  {
    value: "banner_left_top",
    label: "轮播图-左上（正方形 200-2000 / 1MB）",
    module_key: "banners:left_top",
    media_type: "image",
    max_size_kb: 1024,
    min_width: 200,
    max_width: 2000,
    min_height: 200,
    max_height: 2000,
    ratio_width: 1,
    ratio_height: 1
  },
  {
    value: "banner_left_bottom",
    label: "轮播图-左下（正方形 200-2000 / 1MB）",
    module_key: "banners:left_bottom",
    media_type: "image",
    max_size_kb: 1024,
    min_width: 200,
    max_width: 2000,
    min_height: 200,
    max_height: 2000,
    ratio_width: 1,
    ratio_height: 1
  },
  {
    value: "banner_right",
    label: "轮播图-右侧（比例 670:540 / 1MB）",
    module_key: "banners:right",
    media_type: "image",
    max_size_kb: 1024,
    min_width: 670,
    min_height: 540,
    ratio_width: 670,
    ratio_height: 540
  },
  {
    value: "scene_image",
    label: "场景图片（比例 2:3 / 最小 240×360）",
    module_key: "scenes",
    media_type: "image",
    max_size_kb: 0,
    min_width: 240,
    min_height: 360,
    ratio_width: 2,
    ratio_height: 3
  },
  {
    value: "clothes_image",
    label: "服饰偏好图片（正方形 200-2000 / 1MB）",
    module_key: "clothes_categories",
    media_type: "image",
    max_size_kb: 1024,
    min_width: 200,
    max_width: 2000,
    min_height: 200,
    max_height: 2000,
    ratio_width: 1,
    ratio_height: 1
  },
  {
    value: "photo_image",
    label: "拍摄偏好图片（正方形 200-2000 / 1MB）",
    module_key: "photo_hobbies",
    media_type: "image",
    max_size_kb: 1024,
    min_width: 200,
    max_width: 2000,
    min_height: 200,
    max_height: 2000,
    ratio_width: 1,
    ratio_height: 1
  }
];

const statusOptions = [
  { value: 1, label: "启用" },
  { value: 0, label: "停用" }
];

const MediaRulePanel = ({ request }: MediaRulePanelProps) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [items, setItems] = useState<MediaRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<MediaRule | null>(null);
  const [presetKey, setPresetKey] = useState<string | null>(null);
  const presetKeyRef = useRef<string | null>(null);
  const presetTimerRef = useRef<number | null>(null);
  const [form] = Form.useForm<RuleFormValues>();
  const selectedMediaType = Form.useWatch("media_type", form);

  const loadItems = async () => {
    setLoading(true);
    try {
      const res = await request<{ data: MediaRule[] }>("/api/media/rules");
      setItems(res.data || []);
    } catch (error) {
      setItems([]);
      messageApi.error(error instanceof Error ? error.message : "获取规则失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadItems();
  }, []);

  useEffect(() => {
    if (!editorOpen) {
      return;
    }
    if (selectedMediaType === "image") {
      if (!form.getFieldValue("allow_formats")) {
        form.setFieldValue("allow_formats", imageFormatDefault);
      }
      if (!form.getFieldValue("compress_quality")) {
        form.setFieldValue("compress_quality", 85);
      }
    }
    if (selectedMediaType === "video") {
      if (!form.getFieldValue("allow_formats")) {
        form.setFieldValue("allow_formats", videoFormatDefault);
      }
    }
    if (selectedMediaType === "audio") {
      form.setFieldValue("allow_formats", "");
    }
  }, [editorOpen, form, selectedMediaType]);

  useEffect(() => {
    if (!editorOpen) {
      if (presetTimerRef.current) {
        window.clearTimeout(presetTimerRef.current);
        presetTimerRef.current = null;
      }
      presetKeyRef.current = null;
    }
  }, [editorOpen]);

  const openEditor = (item?: MediaRule) => {
    setEditingItem(item ?? null);
    setPresetKey(null);
    presetKeyRef.current = null;
    setEditorOpen(true);
  };

  const syncEditorForm = (item: MediaRule | null) => {
    form.setFieldsValue({
      module_key: item?.module_key ?? undefined,
      media_type: item?.media_type ?? undefined,
      max_size_kb: item?.max_size_kb ?? undefined,
      min_width: item?.min_width ?? undefined,
      max_width: item?.max_width ?? undefined,
      min_height: item?.min_height ?? undefined,
      max_height: item?.max_height ?? undefined,
      ratio_width: item?.ratio_width ?? undefined,
      ratio_height: item?.ratio_height ?? undefined,
      min_duration_ms: item?.min_duration_ms ?? undefined,
      max_duration_ms: item?.max_duration_ms ?? undefined,
      allow_formats: item?.allow_formats ?? undefined,
      resize_mode: item?.resize_mode ?? undefined,
      target_format: item?.target_format ?? undefined,
      compress_quality: item?.compress_quality ?? undefined,
      status: item?.status ?? 1
    });
  };

  const applyPreset = (value: string) => {
    const preset = rulePresets.find((item) => item.value === value);
    if (!preset) {
      return;
    }
    form.setFieldsValue({
      module_key: preset.module_key,
      media_type: preset.media_type,
      max_size_kb: preset.max_size_kb,
      status: 1
    });
    presetKeyRef.current = value;
    if (presetTimerRef.current) {
      window.clearTimeout(presetTimerRef.current);
    }
    presetTimerRef.current = window.setTimeout(() => {
      if (presetKeyRef.current !== value) {
        return;
      }
      form.setFieldsValue({
        module_key: preset.module_key,
        media_type: preset.media_type,
        max_size_kb: preset.max_size_kb,
        min_width: preset.min_width,
        max_width: preset.max_width ?? 0,
        min_height: preset.min_height,
        max_height: preset.max_height ?? 0,
        ratio_width: preset.ratio_width ?? 0,
        ratio_height: preset.ratio_height ?? 0,
        allow_formats: imageFormatDefault,
        resize_mode: "contain",
        compress_quality: 85,
        status: 1
      });
    }, 0);
    setPresetKey(value);
    messageApi.success("已应用预设规则");
  };

  const submitRule = async () => {
    try {
      const values = await form.validateFields();
      setEditorSubmitting(true);
      const payload = {
        module_key: values.module_key?.trim(),
        media_type: values.media_type,
        max_size_kb: values.max_size_kb ?? 0,
        min_width: values.min_width ?? 0,
        max_width: values.max_width ?? 0,
        min_height: values.min_height ?? 0,
        max_height: values.max_height ?? 0,
        ratio_width: values.ratio_width ?? 0,
        ratio_height: values.ratio_height ?? 0,
        min_duration_ms: values.min_duration_ms ?? 0,
        max_duration_ms: values.max_duration_ms ?? 0,
        allow_formats: values.allow_formats?.trim(),
        resize_mode: values.resize_mode?.trim(),
        target_format: values.target_format?.trim(),
        compress_quality: values.compress_quality ?? 0,
        status: values.status ?? 1
      };
      if (editingItem) {
        await request(`/api/media/rules/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        messageApi.success("规则已更新");
      } else {
        await request("/api/media/rules", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        messageApi.success("规则已创建");
      }
      setEditorOpen(false);
      setEditingItem(null);
      form.resetFields();
      void loadItems();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setEditorSubmitting(false);
    }
  };

  const deleteRule = (item: MediaRule) => {
    Modal.confirm({
      title: "确认删除规则？",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        await request(`/api/media/rules/${item.id}`, { method: "DELETE" });
        messageApi.success("规则已删除");
        void loadItems();
      }
    });
  };

  const columns = [
    {
      title: "模块",
      dataIndex: "module_key",
      key: "module_key",
      render: (value: string) => <Text>{value || "-"}</Text>
    },
    {
      title: "类型",
      dataIndex: "media_type",
      key: "media_type",
      render: (value: string) => <Tag>{value || "-"}</Tag>
    },
    {
      title: "尺寸限制",
      key: "size",
      render: (_: string, record: MediaRule) => {
        const minW = Number(record.min_width || 0);
        const minH = Number(record.min_height || 0);
        const maxW = Number(record.max_width || 0);
        const maxH = Number(record.max_height || 0);
        const hasLimit = minW > 0 || minH > 0 || maxW > 0 || maxH > 0;
        if (!hasLimit) {
          return <Text type="secondary">未设置</Text>;
        }
        const formatValue = (value: number) => (value > 0 ? value : "-");
        return (
          <Text type="secondary">
            {formatValue(minW)}×{formatValue(minH)} ~ {formatValue(maxW)}×{formatValue(maxH)}
          </Text>
        );
      }
    },
    {
      title: "比例",
      key: "ratio",
      render: (_: string, record: MediaRule) => (
        <Text type="secondary">
          {record.ratio_width && record.ratio_height ? `${record.ratio_width}:${record.ratio_height}` : "-"}
        </Text>
      )
    },
    {
      title: "大小/时长",
      key: "limit",
      render: (_: string, record: MediaRule) => (
        <Text type="secondary">
          {record.max_size_kb || 0}KB / {record.max_duration_ms || 0}ms
        </Text>
      )
    },
    {
      title: "格式",
      dataIndex: "allow_formats",
      key: "allow_formats",
      render: (value: string) => <Text type="secondary">{value || "-"}</Text>
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: number) => (value === 0 ? <Tag>停用</Tag> : <Tag color="green">启用</Tag>)
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: MediaRule) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEditor(record)}>
            编辑
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => deleteRule(record)}>
            删除
          </Button>
        </Space>
      )
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">配置图片/视频/音频的尺寸、大小与格式规则。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadItems} loading={loading}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>
              新建规则
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        {items.length ? (
          <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ pageSize: 6 }} />
        ) : (
          <Empty description="暂无媒体规则" />
        )}
      </Card>

      <Modal
        title={editingItem ? "编辑媒体规则" : "新建媒体规则"}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false);
          setEditingItem(null);
          setPresetKey(null);
          form.resetFields();
        }}
        afterOpenChange={(open) => {
          if (open) {
            if (editingItem) {
              syncEditorForm(editingItem);
            } else {
              form.resetFields();
            }
          }
        }}
        onOk={submitRule}
        width={960}
        okButtonProps={{ loading: editorSubmitting }}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        destroyOnClose
      >
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Space wrap>
            <Text type="secondary">预设模板</Text>
            <Select
              style={{ minWidth: 260 }}
              placeholder="选择常用规则"
              value={presetKey ?? undefined}
              options={rulePresets.map((item) => ({ value: item.value, label: item.label }))}
              onChange={(value) => {
                setPresetKey(value);
                applyPreset(value);
              }}
              allowClear
            />
          </Space>
          <Form form={form} layout="vertical" preserve={false} initialValues={{ status: 1 }}>
            <Row gutter={12}>
              <Col span={12}>
                <Form.Item label="模块Key" name="module_key" rules={[{ required: true, message: "请输入模块Key" }]}>
                  <Select
                    mode="combobox"
                    options={moduleKeyOptions}
                    placeholder="如：banners:left_top"
                    showSearch
                    filterOption={(input, option) =>
                      (option?.label ?? "").toString().includes(input) || (option?.value ?? "").toString().includes(input)
                    }
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="媒体类型" name="media_type" rules={[{ required: true, message: "请选择媒体类型" }]}>
                  <Select options={mediaTypes} />
                </Form.Item>
              </Col>
            </Row>
            {selectedMediaType !== "audio" ? (
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item label="最大体积(KB)" name="max_size_kb">
                    <InputNumber min={0} style={{ width: "100%" }} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="状态" name="status">
                    <Select options={statusOptions} />
                  </Form.Item>
                </Col>
              </Row>
            ) : (
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item label="状态" name="status">
                    <Select options={statusOptions} />
                  </Form.Item>
                </Col>
              </Row>
            )}
            {selectedMediaType === "image" ? (
              <>
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Item label="最小宽度" name="min_width">
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="最大宽度" name="max_width">
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Item label="最小高度" name="min_height">
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="最大高度" name="max_height">
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Item label="比例宽" name="ratio_width">
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="比例高" name="ratio_height">
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Item label="允许格式" name="allow_formats">
                      <Input placeholder="如：jpg,jpeg,png" />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="缩放模式" name="resize_mode">
                      <Select
                        allowClear
                        options={[
                          { value: "contain", label: "等比缩放（不裁剪）" },
                          { value: "cover", label: "裁剪填充（保持比例）" },
                          { value: "fill", label: "拉伸到目标尺寸" }
                        ]}
                      />
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Item label="目标格式" name="target_format">
                      <Input placeholder="如：jpg" />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="压缩质量" name="compress_quality">
                      <InputNumber min={0} max={100} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
            {selectedMediaType === "video" ? (
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item label="允许格式" name="allow_formats">
                    <Input placeholder="如：mp4,m4v" />
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
          </Form>
        </Space>
      </Modal>
    </Space>
  );
};

export default MediaRulePanel;
