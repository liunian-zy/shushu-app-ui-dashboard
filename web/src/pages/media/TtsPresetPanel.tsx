import { useEffect, useState } from "react";
import {
  AutoComplete,
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
import { DeleteOutlined, EditOutlined, PlayCircleOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { formatDate } from "../content/constants";

const { Text } = Typography;
const { TextArea } = Input;

type TTSPreset = {
  id: number;
  name?: string | null;
  voice_id?: string | null;
  emotion_name?: string | null;
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
  emotion_name?: string;
  volume?: number;
  speed?: number;
  pitch?: number;
  stability?: number;
  similarity?: number;
  exaggeration?: number;
  status?: number;
  is_default?: boolean;
};

type TTSTestResult = {
  audio_path: string;
  audio_url?: string;
};

type TtsPresetPanelProps = {
  request: <T>(path: string, options?: RequestInit) => Promise<T>;
};

const defaultPreset = {
  voice_id: "70eb6772-4cd1-11f0-9276-00163e0fe4f9",
  emotion_name: "Happy",
  volume: 58,
  speed: 1,
  pitch: 56,
  stability: 50,
  similarity: 95,
  exaggeration: 0
};

const testTextPresets = [
  "打卡普济寺60年一开的山门，站上皇帝都错过的C位。这一刻，你不是游客，是历史的VIP！",
  "与猴共影：和黔灵山首席模特，完成一次完美合作，用一张合影，刷爆你的朋友圈吧",
  "AI会为你扫描整个定海古城，精准锁定古城所有绝美机位，一键生成专属文旅大片",
  "选择后，AI将会带你穿越万岁山武侠江湖，为您随机匹配江湖身份，快来看看你在万岁山武侠江湖的身份卡吧"
];

const testTextOptions = testTextPresets.map((text, index) => ({
  value: text,
  label: `预设${index + 1}：${text.length > 20 ? `${text.slice(0, 20)}...` : text}`
}));

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
  const [emotionOptions, setEmotionOptions] = useState<{ value: string; label: string }[]>([]);
  const [emotionLoading, setEmotionLoading] = useState(false);
  const [testOpen, setTestOpen] = useState(false);
  const [testLoading, setTestLoading] = useState(false);
  const [testResults, setTestResults] = useState<TTSTestResult[]>([]);
  const [testText, setTestText] = useState(testTextPresets[0]);
  const [testCount, setTestCount] = useState(3);
  const [testPreset, setTestPreset] = useState<TTSPreset | null>(null);
  const [testFromForm, setTestFromForm] = useState(false);
  const [form] = Form.useForm<PresetFormValues>();

  const voiceIdValue = Form.useWatch("voice_id", form) ?? defaultPreset.voice_id;
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

  const loadVoiceDetail = async (voiceId: string) => {
    const trimmed = voiceId?.trim();
    if (!trimmed) {
      setEmotionOptions([]);
      return;
    }
    setEmotionLoading(true);
    try {
      const res = await request<{ data: { emotion?: { name: string; show_name?: string }[] } }>(
        "/api/tts/voice-detail",
        {
          method: "POST",
          body: JSON.stringify({ voice_id: trimmed, slang_id: 18 })
        }
      );
      const emotions = res.data?.emotion || [];
      setEmotionOptions(
        emotions.map((item) => ({
          value: item.name,
          label: item.show_name ? `${item.show_name} (${item.name})` : item.name
        }))
      );
    } catch (error) {
      setEmotionOptions([]);
      messageApi.error(error instanceof Error ? error.message : "获取情绪列表失败");
    } finally {
      setEmotionLoading(false);
    }
  };

  useEffect(() => {
    void loadItems();
  }, []);

  useEffect(() => {
    if (!editorOpen) {
      return;
    }
    void loadVoiceDetail(voiceIdValue);
  }, [editorOpen, voiceIdValue]);

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
      emotion_name: item?.emotion_name ?? defaultPreset.emotion_name,
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
        emotion_name: values.emotion_name?.trim() || undefined,
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

  const openTestFromItem = (item: TTSPreset) => {
    setTestFromForm(false);
    setTestPreset(item);
    setTestResults([]);
    setTestText(testTextPresets[0]);
    setTestCount(3);
    setTestOpen(true);
  };

  const openTestFromForm = () => {
    setTestFromForm(true);
    setTestPreset(null);
    setTestResults([]);
    setTestText(testTextPresets[0]);
    setTestCount(3);
    setTestOpen(true);
  };

  const buildTestParams = () => {
    if (testFromForm) {
      const values = form.getFieldsValue();
      return {
        voice_id: values.voice_id?.trim() || defaultPreset.voice_id,
        emotion_name: values.emotion_name?.trim() || undefined,
        volume: values.volume ?? defaultPreset.volume,
        speed: values.speed ?? defaultPreset.speed,
        pitch: values.pitch ?? defaultPreset.pitch,
        stability: values.stability ?? defaultPreset.stability,
        similarity: values.similarity ?? defaultPreset.similarity,
        exaggeration: values.exaggeration ?? defaultPreset.exaggeration
      };
    }
    const preset = testPreset;
    return {
      voice_id: preset?.voice_id?.trim() || defaultPreset.voice_id,
      emotion_name: preset?.emotion_name?.trim() || undefined,
      volume: preset?.volume ?? defaultPreset.volume,
      speed: preset?.speed ?? defaultPreset.speed,
      pitch: preset?.pitch ?? defaultPreset.pitch,
      stability: preset?.stability ?? defaultPreset.stability,
      similarity: preset?.similarity ?? defaultPreset.similarity,
      exaggeration: preset?.exaggeration ?? defaultPreset.exaggeration
    };
  };

  const runTest = async () => {
    const text = testText?.trim();
    if (!text) {
      messageApi.warning("请输入测试文案");
      return;
    }
    const safeCount = Math.max(1, Math.min(6, Number(testCount) || 1));
    const params = buildTestParams();
    setTestLoading(true);
    try {
      const tasks = Array.from({ length: safeCount }, () =>
        request<TTSTestResult>("/api/tts/convert", {
          method: "POST",
          body: JSON.stringify({
            text,
            module_key: "tts-presets-test",
            draft_version_id: 0,
            voice_id: params.voice_id,
            emotion_name: params.emotion_name,
            volume: params.volume,
            speed: params.speed,
            pitch: params.pitch,
            stability: params.stability,
            similarity: params.similarity,
            exaggeration: params.exaggeration
          })
        })
      );
      const results = await Promise.all(tasks);
      setTestResults(results);
      if (!results.length) {
        messageApi.warning("未生成测试语音");
      }
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "测试失败");
    } finally {
      setTestLoading(false);
    }
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
          <Text type="secondary">情绪 {record.emotion_name || "-"}</Text>
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
          <Button size="small" icon={<PlayCircleOutlined />} onClick={() => openTestFromItem(record)}>
            测试
          </Button>
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

  const testParams = buildTestParams();

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
        footer={[
          <Button key="test" onClick={openTestFromForm}>
            测试语音
          </Button>,
          <Button key="cancel" onClick={closeEditor}>
            取消
          </Button>,
          <Button key="submit" type="primary" onClick={submitEditor} loading={editorSubmitting}>
            保存
          </Button>
        ]}
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
          <Form.Item label="情绪参数" name="emotion_name">
            <AutoComplete
              options={emotionOptions}
              placeholder="输入或选择情绪"
              allowClear
              filterOption={(inputValue, option) =>
                String(option?.value ?? "").toLowerCase().includes(inputValue.toLowerCase())
              }
            />
            {emotionLoading ? <Text type="secondary">正在加载情绪列表...</Text> : null}
          </Form.Item>
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
      <Modal
        title="语音测试"
        open={testOpen}
        onCancel={() => setTestOpen(false)}
        footer={[
          <Button key="clear" onClick={() => setTestResults([])} disabled={!testResults.length}>
            清空结果
          </Button>,
          <Button key="cancel" onClick={() => setTestOpen(false)}>
            关闭
          </Button>,
          <Button key="run" type="primary" onClick={runTest} loading={testLoading}>
            生成测试语音
          </Button>
        ]}
        width={760}
        destroyOnClose
      >
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Space direction="vertical" size={4}>
            <Text type="secondary">
              当前音色：{testParams.voice_id}
              {testParams.emotion_name ? ` / 情绪：${testParams.emotion_name}` : ""}
            </Text>
            <Text type="secondary">
              参数：音量 {testParams.volume}，语速 {testParams.speed}，音高 {testParams.pitch}，稳定 {testParams.stability}，
              相似 {testParams.similarity}，夸张 {testParams.exaggeration}
            </Text>
          </Space>
          <Space wrap>
            <Select
              placeholder="选择预设文案"
              options={testTextOptions}
              onChange={(value) => setTestText(value)}
              value={testTextOptions.some((item) => item.value === testText) ? testText : undefined}
              style={{ minWidth: 260 }}
              allowClear
            />
            <InputNumber min={1} max={6} value={testCount} onChange={(value) => setTestCount(value ?? 1)} />
            <Text type="secondary">条并发生成</Text>
          </Space>
          <TextArea
            rows={4}
            value={testText}
            onChange={(event) => setTestText(event.target.value)}
            placeholder="输入或编辑测试文案"
          />
          <Space direction="vertical" size={12} style={{ width: "100%" }}>
            {testResults.map((item, index) => (
              <Card key={`${item.audio_path}-${index}`} size="small">
                <Space direction="vertical" size={8} style={{ width: "100%" }}>
                  <Text strong>候选 {index + 1}</Text>
                  {item.audio_url ? <audio controls src={item.audio_url} style={{ width: "100%" }} /> : null}
                  <Text type="secondary">{item.audio_path}</Text>
                </Space>
              </Card>
            ))}
          </Space>
        </Space>
      </Modal>
    </Space>
  );
};

export default TtsPresetPanel;
