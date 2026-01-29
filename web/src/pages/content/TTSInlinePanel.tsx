import { useEffect, useMemo, useState } from "react";
import { Button, Card, Col, Input, InputNumber, Modal, Row, Select, Slider, Space, Typography, message } from "antd";
import type { RequestFn, TTSPreset, TTSOptions, TTSResult } from "./utils";

const { Text } = Typography;

type TTSInlinePanelProps = {
  title?: string;
  text?: string | null;
  presets?: TTSPreset[];
  onGenerate: (count: number, options: TTSOptions) => Promise<TTSResult[]>;
  onSelect: (result: TTSResult) => void;
  request?: RequestFn;
  onPresetsReload?: () => void;
  disabled?: boolean;
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

const TTSInlinePanel = ({ title, text, presets, onGenerate, onSelect, request, onPresetsReload, disabled }: TTSInlinePanelProps) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [count, setCount] = useState(3);
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState<TTSResult[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [presetId, setPresetId] = useState<number | null>(null);
  const [voiceId, setVoiceId] = useState<string>(defaultPreset.voice_id);
  const [emotionName, setEmotionName] = useState<string>(defaultPreset.emotion_name);
  const [volume, setVolume] = useState<number>(defaultPreset.volume);
  const [speed, setSpeed] = useState<number>(defaultPreset.speed);
  const [pitch, setPitch] = useState<number>(defaultPreset.pitch);
  const [stability, setStability] = useState<number>(defaultPreset.stability);
  const [similarity, setSimilarity] = useState<number>(defaultPreset.similarity);
  const [exaggeration, setExaggeration] = useState<number>(defaultPreset.exaggeration);
  const [createOpen, setCreateOpen] = useState(false);
  const [createName, setCreateName] = useState("");
  const [createLoading, setCreateLoading] = useState(false);

  useEffect(() => {
    setOptions([]);
    setError(null);
  }, [text]);

  const presetOptions = useMemo(
    () =>
      (presets || [])
        .filter((item) => (item.status ?? 1) === 1)
        .map((item) => ({
          value: item.id,
          label: item.name || `预设${item.id}`
        })),
    [presets]
  );

  const selectedPreset = useMemo(
    () => (presets || []).find((item) => item.id === presetId) ?? null,
    [presets, presetId]
  );

  const applyPreset = (preset: TTSPreset | null) => {
    if (!preset) {
      setVoiceId(defaultPreset.voice_id);
      setEmotionName(defaultPreset.emotion_name);
      setVolume(defaultPreset.volume);
      setSpeed(defaultPreset.speed);
      setPitch(defaultPreset.pitch);
      setStability(defaultPreset.stability);
      setSimilarity(defaultPreset.similarity);
      setExaggeration(defaultPreset.exaggeration);
      return;
    }
    setVoiceId(preset.voice_id || defaultPreset.voice_id);
    setEmotionName(preset.emotion_name || defaultPreset.emotion_name);
    setVolume(preset.volume ?? defaultPreset.volume);
    setSpeed(preset.speed ?? defaultPreset.speed);
    setPitch(preset.pitch ?? defaultPreset.pitch);
    setStability(preset.stability ?? defaultPreset.stability);
    setSimilarity(preset.similarity ?? defaultPreset.similarity);
    setExaggeration(preset.exaggeration ?? defaultPreset.exaggeration);
  };

  useEffect(() => {
    if (!presetOptions.length) {
      setPresetId(null);
      applyPreset(null);
      return;
    }
    if (presetId && presetOptions.some((item) => item.value === presetId)) {
      return;
    }
    const fallback = (presets || []).find((item) => (item.is_default ?? 0) === 1) ?? presets?.[0] ?? null;
    setPresetId(fallback?.id ?? null);
    applyPreset(fallback ?? null);
  }, [presetOptions, presetId, presets]);

  useEffect(() => {
    if (!selectedPreset) {
      return;
    }
    applyPreset(selectedPreset);
  }, [selectedPreset?.id]);

  const handleGenerate = async () => {
    if (!text?.trim()) {
      setError("请先填写语音文案");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const result = await onGenerate(count, {
        preset_id: presetId ?? undefined,
        voice_id: voiceId.trim() || undefined,
        emotion_name: emotionName.trim() || undefined,
        volume,
        speed,
        pitch,
        stability,
        similarity,
        exaggeration
      });
      setOptions(result);
      if (result.length === 0) {
        setError("未生成语音，请检查文案或服务状态");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "生成失败");
    } finally {
      setLoading(false);
    }
  };

  const openCreatePreset = () => {
    setCreateName("");
    setCreateOpen(true);
  };

  const submitCreatePreset = async () => {
    if (!request) {
      return;
    }
    const name = createName.trim();
    if (!name) {
      messageApi.warning("请输入预设名称");
      return;
    }
    setCreateLoading(true);
    try {
      await request("/api/tts/presets", {
        method: "POST",
        body: JSON.stringify({
          name,
          voice_id: voiceId.trim() || defaultPreset.voice_id,
          emotion_name: emotionName.trim() || undefined,
          volume,
          speed,
          pitch,
          stability,
          similarity,
          exaggeration,
          status: 1,
          is_default: 0
        })
      });
      messageApi.success("预设已创建");
      setCreateOpen(false);
      if (onPresetsReload) {
        onPresetsReload();
      }
    } catch (err) {
      messageApi.error(err instanceof Error ? err.message : "创建失败");
    } finally {
      setCreateLoading(false);
    }
  };

  return (
    <Card size="small" style={{ borderRadius: 12 }}>
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        {contextHolder}
        {title ? <Text strong>{title}</Text> : null}
        <Space direction="vertical" size={4}>
          <Text type="secondary">语音参数支持选择预设，并可在生成前微调。</Text>
          <Text type="secondary">音量/语速/音高/稳定度/相似度/夸张度均为可控参数。</Text>
        </Space>
        <Space wrap>
          <Select
            placeholder="选择语音预设"
            value={presetId ?? undefined}
            options={presetOptions}
            onChange={(value) => setPresetId(value)}
            allowClear
            style={{ minWidth: 220 }}
          />
          <Button onClick={() => applyPreset(selectedPreset)} disabled={!selectedPreset}>
            恢复预设值
          </Button>
          <Button onClick={openCreatePreset} disabled={!request}>
            新增预设
          </Button>
        </Space>
        <Space direction="vertical" size={8} style={{ width: "100%" }}>
          <Text type="secondary">voice_id：用于指定音色，留空时使用默认音色。</Text>
          <Input
            value={voiceId}
            onChange={(event) => setVoiceId(event.target.value)}
            placeholder="手动输入 voice_id"
          />
        </Space>
        <Space direction="vertical" size={8} style={{ width: "100%" }}>
          <Text type="secondary">emotion_name：可选情绪参数，留空时使用服务默认值。</Text>
          <Input
            value={emotionName}
            onChange={(event) => setEmotionName(event.target.value)}
            placeholder="例如：Happy / Default"
          />
        </Space>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Row gutter={12} align="middle">
            <Col span={6}><Text type="secondary">音量 0-100</Text></Col>
            <Col span={14}>
              <Slider min={0} max={100} value={volume} onChange={(value) => setVolume(value)} />
            </Col>
            <Col span={4}>
              <InputNumber min={0} max={100} value={volume} onChange={(value) => setVolume(value ?? 0)} />
            </Col>
          </Row>
          <Row gutter={12} align="middle">
            <Col span={6}><Text type="secondary">语速 0.5-2.0</Text></Col>
            <Col span={14}>
              <Slider min={0.5} max={2} step={0.1} value={speed} onChange={(value) => setSpeed(value)} />
            </Col>
            <Col span={4}>
              <InputNumber min={0.5} max={2} step={0.1} value={speed} onChange={(value) => setSpeed(value ?? 1)} />
            </Col>
          </Row>
          <Row gutter={12} align="middle">
            <Col span={6}><Text type="secondary">音高 1-100</Text></Col>
            <Col span={14}>
              <Slider min={1} max={100} value={pitch} onChange={(value) => setPitch(value)} />
            </Col>
            <Col span={4}>
              <InputNumber min={1} max={100} value={pitch} onChange={(value) => setPitch(value ?? 1)} />
            </Col>
          </Row>
          <Row gutter={12} align="middle">
            <Col span={6}><Text type="secondary">稳定度 0-100</Text></Col>
            <Col span={14}>
              <Slider min={0} max={100} value={stability} onChange={(value) => setStability(value)} />
            </Col>
            <Col span={4}>
              <InputNumber min={0} max={100} value={stability} onChange={(value) => setStability(value ?? 0)} />
            </Col>
          </Row>
          <Row gutter={12} align="middle">
            <Col span={6}><Text type="secondary">相似度 0-100</Text></Col>
            <Col span={14}>
              <Slider min={0} max={100} value={similarity} onChange={(value) => setSimilarity(value)} />
            </Col>
            <Col span={4}>
              <InputNumber min={0} max={100} value={similarity} onChange={(value) => setSimilarity(value ?? 0)} />
            </Col>
          </Row>
          <Row gutter={12} align="middle">
            <Col span={6}><Text type="secondary">夸张度 0-100</Text></Col>
            <Col span={14}>
              <Slider min={0} max={100} value={exaggeration} onChange={(value) => setExaggeration(value)} />
            </Col>
            <Col span={4}>
              <InputNumber min={0} max={100} value={exaggeration} onChange={(value) => setExaggeration(value ?? 0)} />
            </Col>
          </Row>
        </Space>
        <Space wrap>
          <InputNumber min={1} max={6} value={count} onChange={(value) => setCount(value ?? 1)} />
          <Text type="secondary">条候选语音</Text>
          <Button type="primary" onClick={handleGenerate} loading={loading} disabled={disabled}>
            生成语音
          </Button>
          <Button onClick={() => setOptions([])} disabled={!options.length}>
            清空
          </Button>
        </Space>
        {text ? (
          <Text type="secondary">当前文案：{text.length > 60 ? `${text.slice(0, 60)}...` : text}</Text>
        ) : null}
        {error ? <Text type="danger">{error}</Text> : null}
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          {options.map((item, index) => (
            <Card key={`${item.audio_path}-${index}`} size="small">
              <Space direction="vertical" size={8} style={{ width: "100%" }}>
                <Text strong>候选 {index + 1}</Text>
                {item.audio_url ? <audio src={item.audio_url} controls style={{ width: "100%" }} /> : null}
                <Space>
                  <Button type="primary" onClick={() => onSelect(item)}>
                    使用此语音
                  </Button>
                  <Text type="secondary" style={{ maxWidth: 320 }} ellipsis>
                    {item.audio_path}
                  </Text>
                </Space>
              </Space>
            </Card>
          ))}
        </Space>
      </Space>
      <Modal
        title="新增语音预设"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={submitCreatePreset}
        okText="保存"
        cancelText="取消"
        confirmLoading={createLoading}
        destroyOnClose
      >
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">将当前参数保存为新的预设名称。</Text>
          <Input
            value={createName}
            onChange={(event) => setCreateName(event.target.value)}
            placeholder="例如：场景默认女声"
          />
        </Space>
      </Modal>
    </Card>
  );
};

export default TTSInlinePanel;
