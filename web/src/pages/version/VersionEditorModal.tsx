import { useMemo } from "react";
import { Button, Col, Form, Input, Modal, Row, Select, Space, Typography } from "antd";
import { baseFeishuFields, buildDefaultFeishuFields, DraftVersion, optionalFeishuFields } from "./constants";

const { Text } = Typography;

export type VersionEditorValues = {
  location_name: string;
  app_version_name?: string;
  ai_modal?: string;
  feishu_field_names?: string[];
};

type VersionEditorModalProps = {
  open: boolean;
  submitting: boolean;
  version: DraftVersion | null;
  onCancel: () => void;
  onSubmit: (values: VersionEditorValues) => void;
};

const aiOptions = [
  { value: "SD", label: "SD（默认）" },
  { value: "NANO", label: "NANO" }
];

const VersionEditorModal = ({ open, submitting, version, onCancel, onSubmit }: VersionEditorModalProps) => {
  const [form] = Form.useForm<VersionEditorValues>();
  const aiModalValue = Form.useWatch("ai_modal", form);

  const syncEditorForm = (item: DraftVersion | null) => {
    form.setFieldsValue({
      location_name: item?.location_name ?? undefined,
      app_version_name: item?.app_version_name ?? undefined,
      ai_modal: item?.ai_modal ?? "SD",
      feishu_field_names: item?.feishu_field_list ?? buildDefaultFeishuFields(item?.ai_modal ?? "SD")
    });
  };

  const feishuOptions = useMemo(() => {
    const fields = new Set<string>([...baseFeishuFields, ...optionalFeishuFields]);
    const currentList = (form.getFieldValue("feishu_field_names") || []) as string[];
    currentList.forEach((item) => fields.add(item));
    if (aiModalValue && aiModalValue.toUpperCase() === "SD") {
      fields.add("SD模式");
    }
    return Array.from(fields).map((item) => ({ value: item, label: item }));
  }, [aiModalValue, form]);

  const handleResetFields = () => {
    const defaults = buildDefaultFeishuFields(aiModalValue ?? "SD");
    form.setFieldValue("feishu_field_names", defaults);
  };

  const handleOk = async () => {
    const values = await form.validateFields();
    onSubmit(values);
  };

  return (
    <Modal
      title={version ? "编辑版本配置" : "新建版本"}
      open={open}
      onCancel={onCancel}
      width={860}
      afterOpenChange={(nextOpen) => {
        if (nextOpen) {
          syncEditorForm(version ?? null);
        }
      }}
      onOk={handleOk}
      okButtonProps={{ loading: submitting }}
      styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
      destroyOnClose
    >
      <Form form={form} layout="vertical" preserve={false}>
        <Form.Item
          label="景区名称"
          name="location_name"
          rules={[{ required: true, message: "请输入景区名称" }]}
        >
          <Input placeholder="如：博物馆" />
        </Form.Item>
        <Row gutter={12}>
          <Col span={12}>
            <Form.Item label="版本名称" name="app_version_name">
              <Input placeholder="留空将自动生成（拼音大写）" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="AI 模型" name="ai_modal">
              <Select options={aiOptions} />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item label="Feishu 字段" name="feishu_field_names">
          <Select
            mode="tags"
            placeholder="输入或选择字段"
            options={feishuOptions}
            tokenSeparators={[",", "，"]}
          />
        </Form.Item>
        <Space align="center">
          <Button onClick={handleResetFields}>重置默认字段</Button>
          <Text type="secondary">
            默认字段包含基础列，SD 模式会自动补充「SD模式」。
          </Text>
        </Space>
      </Form>
    </Modal>
  );
};

export default VersionEditorModal;
