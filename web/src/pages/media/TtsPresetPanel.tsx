import { useEffect, useState } from "react";
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Slider,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message
} from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { formatDate } from "../content/constants";

const { Text } = Typography;

type TTSPreset = {
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
  created_at?: string | null;
  updated_at?: string | null;
};

type PresetFormValues = {
  name?: string;
  voice_id?: string;
  volume?: number;
  speed?: number;
  pitch?: number;
  stability?: number;
  similarity?: number;
  exaggeration?: number;
  status?: number;
  is_default?: boolean;
};

type TtsPresetPanelProps = {
  request: <T>(path: string, options?: RequestInit) => Promise<T>;
};

const defaultPreset = {
  voice_id: "70eb6772-4cd1-11f0-9276-00163e0fe4f9",
  volume: 58,
  speed: 1,
  pitch: 56,
  stability: 50,
  similarity: 95,
  exaggeration: 0
};

const statusOptions = [
  { value: 1, label: "启用" },
  { value: 0, label: "停用" }
];

const TtsPresetPanel = ({ request }: TtsPresetPanelProps) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [items, setItems] = useState<TTSPreset[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<TTSPreset | null>(null);
  const [form] = Form.useForm<PresetFormValues>();

  const volumeValue = Form.useWatch("volume", form) ?? defaultPreset.volume;
  const speedValue = Form.useWatch("speed", form) ?? defaultPreset.speed;
  const pitchValue = Form.useWatch("pitch", form) ?? defaultPreset.pitch;
  const stabilityValue = Form.useWatch("stability", form) ?? defaultPreset.stability;
  const similarityValue = Form.useWatch("similarity", form) ?? defaultPreset.similarity;
  const exaggerationValue = Form.useWatch("exaggeration", form) ?? defaultPreset.exaggeration;
  const isDefault = Form.useWatch("is_default", form) ?? false;

  const loadItems = async () => {
    setLoading(true);
    try {
      const res = await request<{ data: TTSPreset[] }>("/api/tts/presets?all=1");
      setItems(res.data || []);
    } catch (error) {
      setItems([]);
      messageApi.error(error instanceof Error ? error.message : "获取预设失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadItems();
  }, []);

  useEffect(() => {
    if (isDefault) {
      form.setFieldsValue({ status: 1 });
    }
  }, [isDefault, form]);

  const openEditor = (item?: TTSPreset) => {
    setEditingItem(item ?? null);
    setEditorOpen(true);
  };

  const syncForm = (item: TTSPreset | null) => {
    form.setFieldsValue({
      name: item?.name ?? "",
      voice_id: item?.voice_id ?? defaultPreset.voice_id,
      volume: item?.volume ?? defaultPreset.volume,
      speed: item?.speed ?? defaultPreset.speed,
      pitch: item?.pitch ?? defaultPreset.pitch,
      stability: item?.stability ?? defaultPreset.stability,
      similarity: item?.similarity ?? defaultPreset.similarity,
      exaggeration: item?.exaggeration ?? defaultPreset.exaggeration,
      status: item?.status ?? 1,
      is_default: (item?.is_default ?? 0) === 1
    });
  };

  useEffect(() => {
    if (!editorOpen) {
      return;
    }
    syncForm(editingItem);
  }, [editorOpen, editingItem]);

  const closeEditor = () => {
    setEditorOpen(false);
    setEditingItem(null);
    form.resetFields();
  };

  const submitEditor = async () => {
    try {
      const values = await form.validateFields();
      setEditorSubmitting(true);
      const payload = {
        name: values.name?.trim(),
        voice_id: values.voice_id?.trim(),
        volume: values.volume ?? defaultPreset.volume,
        speed: values.speed ?? defaultPreset.speed,
        pitch: values.pitch ?? defaultPreset.pitch,
        stability: values.stability ?? defaultPreset.stability,
        similarity: values.similarity ?? defaultPreset.similarity,
        exaggeration: values.exaggeration ?? defaultPreset.exaggeration,
        status: values.status ?? 1,
        is_default: values.is_default ? 1 : 0
      };
      if (editingItem) {
        await request(`/api/tts/presets/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        messageApi.success("预设已更新");
      } else {
        await request("/api/tts/presets", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        messageApi.success("预设已创建");
      }
      closeEditor();
      void loadItems();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setEditorSubmitting(false);
    }
  };

  const deleteItem = (item: TTSPreset) => {
    if ((item.is_default ?? 0) === 1) {
      messageApi.warning("默认预设不可删除");
      return;
    }
    Modal.confirm({
      title: "确认删除预设？",
      content: "删除后无法恢复。",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/tts/presets/${item.id}`, { method: "DELETE" });
          messageApi.success("预设已删除");
          void loadItems();
        } catch (error) {
          if (error instanceof Error) {
            messageApi.error(error.message);
          }
        }
      }
    });
  };

  const columns = [
    { title: "ID", dataIndex: "id", key: "id", width: 80 },
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
      render: (value: string) => <Text>{value || "-"}</Text>
    },
    {
      title: "voice_id",
      dataIndex: "voice_id",
      key: "voice_id",
      render: (value: string) => <Text type="secondary">{value || "-"}</Text>
    },
    {
      title: "参数",
      key: "params",
      render: (_: string, record: TTSPreset) => (
        <Space direction="vertical" size={0}>
          <Text type="secondary">音量 {record.volume ?? "-"}</Text>
          <Text type="secondary">语速 {record.speed ?? "-"}</Text>
          <Text type="secondary">音高 {record.pitch ?? "-"}</Text>
          <Text type="secondary">稳定 {record.stability ?? "-"}</Text>
          <Text type="secondary">相似 {record.similarity ?? "-"}</Text>
          <Text type="secondary">夸张 {record.exaggeration ?? "-"}</Text>
        </Space>
      )
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: number) => (
        <Tag color={value === 1 ? "green" : "default"}>{value === 1 ? "启用" : "停用"}</Tag>
      )
    },
    {
      title: "默认",
      dataIndex: "is_default",
      key: "is_default",
      render: (value: number) => (value === 1 ? <Tag color="gold">默认</Tag> : "-")
    },
    {
      title: "更新时间",
      dataIndex: "updated_at",
      key: "updated_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: TTSPreset) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEditor(record)}>
            编辑
          </Button>
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            disabled={(record.is_default ?? 0) === 1}
            onClick={() => deleteItem(record)}
          >
            删除
          </Button>
        </Space>
      )
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Card style={{ borderRadius: 16 }}>
        <Space direction="vertical" size={8}>
          <Text strong>语音预设</Text>
          <Text type="secondary">
            配置全局可复用的 TTS 参数组合，生成语音时可快速选择并微调。
          </Text>
        </Space>
      </Card>
      <Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>
          新建预设
        </Button>
        <Button icon={<ReloadOutlined />} onClick={() => loadItems()} loading={loading}>
          刷新
        </Button>
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={items}
        pagination={{ pageSize: 8 }}
        scroll={{ x: 1200 }}
      />
      <Modal
        title={editingItem ? "编辑语音预设" : "新建语音预设"}
        open={editorOpen}
        onCancel={closeEditor}
        onOk={submitEditor}
        okText="保存"
        cancelText="取消"
        confirmLoading={editorSubmitting}
        width={760}
        destroyOnClose
      >
        <Form layout="vertical" form={form}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="预设名称"
                name="name"
                rules={[{ required: true, message: "请输入预设名称" }]}
              >
                <Input placeholder="例如：默认女声" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="voice_id"
                name="voice_id"
                rules={[{ required: true, message: "请输入 voice_id" }]}
              >
                <Input placeholder="输入 voice_id" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} align="middle">
            <Col span={6}><Text type="secondary">音量 0-100</Text></Col>
            <Col span={12}>
              <Slider min={0} max={100} value={volumeValue} onChange={(value) => form.setFieldsValue({ volume: value })} />
            </Col>
            <Col span={6}>
              <Form.Item name="volume" rules={[{ required: true, message: "请输入音量" }]} noStyle>
                <InputNumber min={0} max={100} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} align="middle" style={{ marginTop: 12 }}>
            <Col span={6}><Text type="secondary">语速 0.5-2.0</Text></Col>
            <Col span={12}>
              <Slider
                min={0.5}
                max={2}
                step={0.1}
                value={speedValue}
                onChange={(value) => form.setFieldsValue({ speed: value })}
              />
            </Col>
            <Col span={6}>
              <Form.Item name="speed" rules={[{ required: true, message: "请输入语速" }]} noStyle>
                <InputNumber min={0.5} max={2} step={0.1} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} align="middle" style={{ marginTop: 12 }}>
            <Col span={6}><Text type="secondary">音高 1-100</Text></Col>
            <Col span={12}>
              <Slider min={1} max={100} value={pitchValue} onChange={(value) => form.setFieldsValue({ pitch: value })} />
            </Col>
            <Col span={6}>
              <Form.Item name="pitch" rules={[{ required: true, message: "请输入音高" }]} noStyle>
                <InputNumber min={1} max={100} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} align="middle" style={{ marginTop: 12 }}>
            <Col span={6}><Text type="secondary">稳定度 0-100</Text></Col>
            <Col span={12}>
              <Slider
                min={0}
                max={100}
                value={stabilityValue}
                onChange={(value) => form.setFieldsValue({ stability: value })}
              />
            </Col>
            <Col span={6}>
              <Form.Item name="stability" rules={[{ required: true, message: "请输入稳定度" }]} noStyle>
                <InputNumber min={0} max={100} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} align="middle" style={{ marginTop: 12 }}>
            <Col span={6}><Text type="secondary">相似度 0-100</Text></Col>
            <Col span={12}>
              <Slider
                min={0}
                max={100}
                value={similarityValue}
                onChange={(value) => form.setFieldsValue({ similarity: value })}
              />
            </Col>
            <Col span={6}>
              <Form.Item name="similarity" rules={[{ required: true, message: "请输入相似度" }]} noStyle>
                <InputNumber min={0} max={100} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} align="middle" style={{ marginTop: 12 }}>
            <Col span={6}><Text type="secondary">夸张度 0-100</Text></Col>
            <Col span={12}>
              <Slider
                min={0}
                max={100}
                value={exaggerationValue}
                onChange={(value) => form.setFieldsValue({ exaggeration: value })}
              />
            </Col>
            <Col span={6}>
              <Form.Item name="exaggeration" rules={[{ required: true, message: "请输入夸张度" }]} noStyle>
                <InputNumber min={0} max={100} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16} style={{ marginTop: 16 }}>
            <Col span={12}>
              <Form.Item label="状态" name="status" rules={[{ required: true, message: "请选择状态" }]}>
                <Select options={statusOptions} disabled={isDefault} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="设为默认" name="is_default" valuePropName="checked">
                <Switch checkedChildren="默认" unCheckedChildren="非默认" />
              </Form.Item>
            </Col>
          </Row>
          <Card size="small" style={{ marginTop: 8 }}>
            <Text type="secondary">
              默认预设将用于未指定预设的语音生成，且始终保持启用状态。
            </Text>
          </Card>
        </Form>
      </Modal>
    </Space>
  );
};

export default TtsPresetPanel;
