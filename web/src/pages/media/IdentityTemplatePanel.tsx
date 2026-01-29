import { useEffect, useMemo, useState } from "react";
import { Button, Card, Col, Empty, Form, Image, Input, InputNumber, Modal, Row, Select, Space, Table, Tag, Typography, message } from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import UploadField from "../content/UploadField";
import type { UploadFn } from "../content/utils";

const { Text } = Typography;

type TemplateItem = {
  id: number;
  name?: string | null;
  image?: string | null;
  image_url?: string | null;
  sort?: number | null;
  status?: number | null;
};

type Template = {
  id: number;
  name?: string | null;
  description?: string | null;
  status?: number | null;
  item_count?: number | null;
};

type IdentityTemplatePanelProps = {
  uploadFile: UploadFn;
  request: <T>(path: string, options?: RequestInit) => Promise<T>;
};

type TemplateFormValues = {
  name?: string;
  description?: string;
  status?: number;
};

type TemplateItemFormValues = {
  name?: string;
  image?: string;
  sort?: number;
  status?: number;
};

const statusOptions = [
  { value: 1, label: "启用" },
  { value: 0, label: "停用" }
];

const IdentityTemplatePanel = ({ uploadFile, request }: IdentityTemplatePanelProps) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [templateLoading, setTemplateLoading] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null);
  const [items, setItems] = useState<TemplateItem[]>([]);
  const [itemsLoading, setItemsLoading] = useState(false);
  const [templateModalOpen, setTemplateModalOpen] = useState(false);
  const [templateSubmitting, setTemplateSubmitting] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null);
  const [itemModalOpen, setItemModalOpen] = useState(false);
  const [itemSubmitting, setItemSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<TemplateItem | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [templateForm] = Form.useForm<TemplateFormValues>();
  const [itemForm] = Form.useForm<TemplateItemFormValues>();
  const imageValue = Form.useWatch("image", itemForm);

  const loadTemplates = async () => {
    setTemplateLoading(true);
    try {
      const res = await request<{ data: Template[] }>("/api/identity-templates");
      setTemplates(res.data || []);
      if (!selectedTemplateId && res.data?.length) {
        setSelectedTemplateId(res.data[0].id);
      }
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取模板失败");
    } finally {
      setTemplateLoading(false);
    }
  };

  const loadItems = async (templateId?: number | null) => {
    if (!templateId) {
      setItems([]);
      return;
    }
    setItemsLoading(true);
    try {
      const res = await request<{ data: TemplateItem[] }>(`/api/identity-templates/${templateId}/items`);
      setItems(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取模板明细失败");
    } finally {
      setItemsLoading(false);
    }
  };

  useEffect(() => {
    void loadTemplates();
  }, []);

  useEffect(() => {
    void loadItems(selectedTemplateId);
  }, [selectedTemplateId]);

  const openTemplateModal = (template?: Template) => {
    setEditingTemplate(template ?? null);
    setTemplateModalOpen(true);
  };

  const syncTemplateForm = (item: Template | null) => {
    templateForm.setFieldsValue({
      name: item?.name ?? undefined,
      description: item?.description ?? undefined,
      status: item?.status ?? 1
    });
  };

  const submitTemplate = async () => {
    try {
      const values = await templateForm.validateFields();
      setTemplateSubmitting(true);
      const payload = {
        name: values.name?.trim(),
        description: values.description?.trim(),
        status: values.status ?? 1
      };
      if (editingTemplate) {
        await request(`/api/identity-templates/${editingTemplate.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        messageApi.success("模板已更新");
      } else {
        await request("/api/identity-templates", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        messageApi.success("模板已创建");
      }
      setTemplateModalOpen(false);
      setEditingTemplate(null);
      templateForm.resetFields();
      void loadTemplates();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setTemplateSubmitting(false);
    }
  };

  const deleteTemplate = (template: Template) => {
    Modal.confirm({
      title: "确认删除模板？",
      content: "模板及其明细将被移除。",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/identity-templates/${template.id}`, { method: "DELETE" });
          messageApi.success("模板已删除");
          if (selectedTemplateId === template.id) {
            setSelectedTemplateId(null);
            setItems([]);
          }
          void loadTemplates();
        } catch (error) {
          if (error instanceof Error) {
            messageApi.error(error.message);
          }
        }
      }
    });
  };

  const openItemModal = (item?: TemplateItem) => {
    if (!selectedTemplateId) {
      messageApi.warning("请先选择模板");
      return;
    }
    setEditingItem(item ?? null);
    setImagePreview(item?.image_url ?? null);
    setItemModalOpen(true);
  };

  const syncItemForm = (item: TemplateItem | null) => {
    itemForm.setFieldsValue({
      name: item?.name ?? undefined,
      image: item?.image ?? undefined,
      sort: item?.sort ?? 0,
      status: item?.status ?? 1
    });
  };

  const submitItem = async () => {
    if (!selectedTemplateId) {
      return;
    }
    try {
      const values = await itemForm.validateFields();
      setItemSubmitting(true);
      const payload = {
        name: values.name?.trim(),
        image: values.image,
        sort: values.sort ?? 0,
        status: values.status ?? 1
      };
      if (editingItem) {
        await request(`/api/identity-template-items/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        messageApi.success("模板项已更新");
      } else {
        await request(`/api/identity-templates/${selectedTemplateId}/items`, {
          method: "POST",
          body: JSON.stringify(payload)
        });
        messageApi.success("模板项已创建");
      }
      setItemModalOpen(false);
      setEditingItem(null);
      setImagePreview(null);
      itemForm.resetFields();
      void loadItems(selectedTemplateId);
      void loadTemplates();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setItemSubmitting(false);
    }
  };

  const deleteItem = (item: TemplateItem) => {
    Modal.confirm({
      title: "确认删除模板项？",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/identity-template-items/${item.id}`, { method: "DELETE" });
          messageApi.success("模板项已删除");
          void loadItems(selectedTemplateId);
          void loadTemplates();
        } catch (error) {
          if (error instanceof Error) {
            messageApi.error(error.message);
          }
        }
      }
    });
  };

  const templateOptions = useMemo(
    () =>
      templates.map((item) => ({
        value: item.id,
        label: item.name || `模板${item.id}`
      })),
    [templates]
  );

  const templateColumns = [
    {
      title: "模板名称",
      dataIndex: "name",
      key: "name",
      render: (value: string) => <Text>{value || "-"}</Text>
    },
    {
      title: "描述",
      dataIndex: "description",
      key: "description",
      render: (value: string) => <Text type="secondary">{value || "-"}</Text>
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: number) => (value === 0 ? <Tag>停用</Tag> : <Tag color="green">启用</Tag>)
    },
    {
      title: "明细数",
      dataIndex: "item_count",
      key: "item_count",
      render: (value: number) => <Text>{value ?? 0}</Text>
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: Template) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openTemplateModal(record)}>
            编辑
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => deleteTemplate(record)}>
            删除
          </Button>
        </Space>
      )
    }
  ];

  const itemColumns = [
    {
      title: "身份",
      dataIndex: "name",
      key: "name",
      render: (value: string) => <Text>{value || "-"}</Text>
    },
    {
      title: "排序",
      dataIndex: "sort",
      key: "sort",
      render: (value: number) => <Text>{value ?? 0}</Text>
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: number) => (value === 0 ? <Tag>停用</Tag> : <Tag color="green">启用</Tag>)
    },
    {
      title: "预览",
      dataIndex: "image",
      key: "image",
      render: (_: string, record: TemplateItem) =>
        record.image_url ? <Image src={record.image_url} width={72} style={{ borderRadius: 8 }} /> : "-"
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: TemplateItem) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openItemModal(record)}>
            编辑
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => deleteItem(record)}>
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
          <Text type="secondary">配置身份默认模板，支持不同组合与启用状态。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadTemplates} loading={templateLoading}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openTemplateModal()}>
              新建模板
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        <Table rowKey="id" columns={templateColumns} dataSource={templates} loading={templateLoading} pagination={{ pageSize: 6 }} />
      </Card>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Space wrap>
            <Select
              style={{ minWidth: 220 }}
              placeholder="选择模板查看明细"
              options={templateOptions}
              value={selectedTemplateId ?? undefined}
              onChange={(value) => setSelectedTemplateId(value)}
              allowClear
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openItemModal()} disabled={!selectedTemplateId}>
              新建身份项
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        {selectedTemplateId ? (
          <Table rowKey="id" columns={itemColumns} dataSource={items} loading={itemsLoading} pagination={{ pageSize: 6 }} />
        ) : (
          <Empty description="请选择模板后编辑身份项" />
        )}
      </Card>

      <Modal
        title={editingTemplate ? "编辑模板" : "新建模板"}
        open={templateModalOpen}
        onCancel={() => {
          setTemplateModalOpen(false);
          setEditingTemplate(null);
          templateForm.resetFields();
        }}
        afterOpenChange={(open) => {
          if (open) {
            syncTemplateForm(editingTemplate);
          }
        }}
        onOk={submitTemplate}
        width={820}
        okButtonProps={{ loading: templateSubmitting }}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        destroyOnClose
      >
        <Form form={templateForm} layout="vertical" preserve={false} initialValues={{ status: 1 }}>
          <Row gutter={12}>
            <Col span={16}>
              <Form.Item label="模板名称" name="name" rules={[{ required: true, message: "请输入模板名称" }]}>
                <Input placeholder="如：标准四身份" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="状态" name="status">
                <Select options={statusOptions} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="描述" name="description">
            <Input placeholder="可选：模板说明" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingItem ? "编辑身份项" : "新建身份项"}
        open={itemModalOpen}
        onCancel={() => {
          setItemModalOpen(false);
          setEditingItem(null);
          setImagePreview(null);
          itemForm.resetFields();
        }}
        afterOpenChange={(open) => {
          if (open) {
            syncItemForm(editingItem);
          }
        }}
        onOk={submitItem}
        width={860}
        okButtonProps={{ loading: itemSubmitting }}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        destroyOnClose
      >
        <Form form={itemForm} layout="vertical" preserve={false} initialValues={{ status: 1 }}>
          <Form.Item label="身份名称" name="name" rules={[{ required: true, message: "请输入身份名称" }]}>
            <Input placeholder="如：男士" />
          </Form.Item>
          <Form.Item label="图片路径" name="image" rules={[{ required: true, message: "请上传图片" }]}>
            <Input placeholder="上传后自动填充" />
          </Form.Item>
          <UploadField
            label="模板图片"
            accept="image/png,image/jpeg"
            value={imageValue}
            previewUrl={imagePreview}
            previewType="image"
            mediaType="image"
            moduleKey="identity-templates"
            request={request}
            onUpload={(file) => uploadFile(file, "identity-templates", 0)}
            onChange={(path, url) => {
              itemForm.setFieldValue("image", path);
              setImagePreview(url ?? null);
            }}
            onClear={() => {
              itemForm.setFieldValue("image", undefined);
              setImagePreview(null);
            }}
          />
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item label="排序" name="sort">
                <InputNumber min={0} max={999} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="状态" name="status">
                <Select options={statusOptions} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </Space>
  );
};

export default IdentityTemplatePanel;
