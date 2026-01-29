import { useEffect, useState } from "react";
import { Button, Card, InputNumber, Modal, Space, Typography } from "antd";
import type { TTSResult } from "./utils";

const { Text } = Typography;

type TTSSelectModalProps = {
  open: boolean;
  title: string;
  text: string;
  onGenerate: (count: number) => Promise<TTSResult[]>;
  onSelect: (result: TTSResult) => void;
  onCancel: () => void;
};

const TTSSelectModal = ({ open, title, text, onGenerate, onSelect, onCancel }: TTSSelectModalProps) => {
  const [count, setCount] = useState(3);
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState<TTSResult[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setOptions([]);
      setError(null);
    }
  }, [open]);

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await onGenerate(count);
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

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onCancel}
      footer={null}
      width={720}
      styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
      destroyOnClose
    >
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <Space wrap>
          <InputNumber min={1} max={6} value={count} onChange={(value) => setCount(value ?? 1)} />
          <Text type="secondary">条候选语音</Text>
          <Button type="primary" onClick={handleGenerate} loading={loading}>
            生成候选
          </Button>
          <Button onClick={() => setOptions([])} disabled={options.length === 0}>
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
    </Modal>
  );
};

export default TTSSelectModal;
