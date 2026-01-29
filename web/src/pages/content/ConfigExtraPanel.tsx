import { useEffect, useMemo, useState } from "react";
import type { Key } from "react";
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
  Switch,
  Table,
  Tag,
  Typography
} from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import UploadField from "./UploadField";
import SubmissionActions from "./SubmissionActions";
import TTSInlinePanel from "./TTSInlinePanel";
import { DraftVersion, ExtraStepItem, formatDate, statusOptions, submitStatusLabels } from "./constants";
import type { Notify, RequestFn, TTSPreset, TTSFn, TTSResult, UploadFn } from "./utils";
import { buildLocalDraftKey, generateTTSBatch, loadLocalDraft, saveLocalDraft, sanitizeSubmissionPayload } from "./utils";

const { Text } = Typography;

type ConfigExtraPanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  generateTTS: TTSFn;
  notify: Notify;
  operatorId?: number | null;
  ttsPresets?: TTSPreset[];
  refreshTtsPresets?: () => void;
};

type ExtraFormValues = {
  step_index?: number;
  field_name?: string;
  label?: string;
  music_text?: string;
  music?: string;
  status?: boolean;
};

const extraFieldOptions = [
  { value: "clothes_categories", label: "服饰偏好" },
  { value: "photo_hobbies", label: "拍摄偏好" }
];

const ConfigExtraPanel = ({ version, request, uploadFile, generateTTS, notify, operatorId, ttsPresets, refreshTtsPresets }: ConfigExtraPanelProps) => {
  const [items, setItems] = useState<ExtraStepItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<ExtraStepItem | null>(null);
  const [musicPreview, setMusicPreview] = useState<string | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<ExtraStepItem[]>([]);
  const [batchSubmitLoading, setBatchSubmitLoading] = useState(false);
  const [filterFieldName, setFilterFieldName] = useState<string | null>(null);
  const [filterSubmitStatus, setFilterSubmitStatus] = useState<string | null>(null);
  const [filterStatus, setFilterStatus] = useState<number | null>(null);
  const [form] = Form.useForm<ExtraFormValues>();
  const musicValue = Form.useWatch("music", form);
  const musicTextValue = Form.useWatch("music_text", form);
  const draftKey = version?.id ? buildLocalDraftKey("config_extra_steps", version.id, editingItem?.id) : "";

  const loadItems = async () => {
    if (!version?.id) {
      setItems([]);
      return;
    }
    setLoading(true);
    try {
      const res = await request<{ data: ExtraStepItem[] }>(
        `/api/draft/config-extra-steps?draft_version_id=${version.id}&app_version_name_id=${version.id}`
      );
      setItems(res.data || []);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取额外配置失败");
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

  const openEditor = (item?: ExtraStepItem) => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    setEditingItem(item ?? null);
    setMusicPreview(item?.music_url ?? null);
    setEditorOpen(true);
  };

  const syncEditorForm = (item: ExtraStepItem | null) => {
    form.setFieldsValue({
      step_index: item?.step_index ?? 1,
      field_name: item?.field_name ?? undefined,
      label: item?.label ?? undefined,
      music_text: item?.music_text ?? undefined,
      music: item?.music ?? undefined,
      status: (item?.status ?? 1) === 1
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
        app_version_name_id: version.id,
        step_index: values.step_index ?? 1,
        field_name: values.field_name?.trim() || undefined,
        label: values.label?.trim() || undefined,
        music_text: values.music_text?.trim() || undefined,
        music: values.music || undefined,
        status: values.status ? 1 : 0,
        updated_by: operatorId ?? undefined
      };
      if (!editingItem && operatorId) {
        payload.created_by = operatorId;
      }
      if (editingItem) {
        await request(`/api/draft/config-extra-steps/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        notify.success("额外配置已更新");
      } else {
        await request("/api/draft/config-extra-steps", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        notify.success("额外配置已创建");
      }
      setEditorOpen(false);
      setEditingItem(null);
      setMusicPreview(null);
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

  const handleDelete = (item: ExtraStepItem) => {
    Modal.confirm({
      title: "确认删除额外配置？",
      content: item.label || "该配置将被移除。",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/draft/config-extra-steps/${item.id}`, { method: "DELETE" });
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

  const handleSelectTTS = (result: TTSResult) => {
    form.setFieldValue("music", result.audio_path);
    setMusicPreview(result.audio_url ?? null);
    notify.success("语音已选择");
  };

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
      if (filterFieldName && (item.field_name || "") !== filterFieldName) {
        return false;
      }
      if (filterStatus != null && (item.status ?? 1) !== filterStatus) {
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
  }, [filterFieldName, filterStatus, filterSubmitStatus, items]);

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
              module_key: "config_extra_steps",
              entity_table: "app_db_config_extra_steps",
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

  const columns = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      width: 80
    },
    {
      title: "顺序",
      dataIndex: "step_index",
      key: "step_index",
      render: (value: number) => <Text>{value ?? "-"}</Text>
    },
    {
      title: "字段名",
      dataIndex: "field_name",
      key: "field_name",
      render: (value: string) => <Text>{value || "-"}</Text>
    },
    {
      title: "标签",
      dataIndex: "label",
      key: "label",
      render: (value: string) => <Text>{value || "-"}</Text>
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
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
      title: "操作",
      key: "actions",
      render: (_: string, record: ExtraStepItem) => (
        <Space>
          <SubmissionActions
            draftVersionId={version?.id || 0}
            moduleKey="config_extra_steps"
            entityTable="app_db_config_extra_steps"
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

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">额外配置用于启用服饰偏好或拍摄偏好等步骤。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadItems} disabled={!version?.id}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()} disabled={!version?.id}>
              新建配置
            </Button>
            <Button onClick={handleBatchSubmission} loading={batchSubmitLoading} disabled={!selectedRowKeys.length || !version?.id}>
              批量提交
            </Button>
          </Space>
          <Space wrap>
            <Text type="secondary">筛选</Text>
            <Select
              placeholder="配置类型"
              value={filterFieldName ?? undefined}
              onChange={(value) => setFilterFieldName(value)}
              options={extraFieldOptions}
              allowClear
              style={{ width: 160 }}
            />
            <Select
              placeholder="提交状态"
              value={filterSubmitStatus ?? undefined}
              onChange={(value) => setFilterSubmitStatus(value)}
              options={submitStatusOptions}
              allowClear
              style={{ width: 160 }}
            />
            <Select
              placeholder="启用状态"
              value={filterStatus ?? undefined}
              onChange={(value) => setFilterStatus(value)}
              options={statusOptions.map((item) => ({ value: item.value, label: item.label }))}
              allowClear
              style={{ width: 140 }}
            />
            <Button
              onClick={() => {
                setFilterFieldName(null);
                setFilterSubmitStatus(null);
                setFilterStatus(null);
              }}
              disabled={!filterFieldName && !filterSubmitStatus && filterStatus == null}
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
                setSelectedRows(rows as ExtraStepItem[]);
              }
            }}
          />
        ) : (
          <Empty description="请选择景区版本后配置额外步骤" />
        )}
      </Card>

      <Modal
        title={editingItem ? "编辑额外配置" : "新建额外配置"}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false);
          setEditingItem(null);
          setMusicPreview(null);
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
          <Button
            key="cancel"
            onClick={() => {
              setEditorOpen(false);
              setEditingItem(null);
              setMusicPreview(null);
              form.resetFields();
            }}
          >
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
              <Form.Item label="步骤顺序" name="step_index" rules={[{ required: true, message: "请输入步骤顺序" }]}>
                <InputNumber min={1} max={99} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
            <Col span={16}>
              <Form.Item label="字段名" name="field_name" rules={[{ required: true, message: "请选择字段名" }]}>
                <Select
                  options={extraFieldOptions}
                  placeholder="选择额外配置字段"
                  onChange={(value) => {
                    const option = extraFieldOptions.find((item) => item.value === value);
                    if (!form.getFieldValue("label") && option) {
                      form.setFieldValue("label", option.label);
                    }
                  }}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={12}>
            <Col span={16}>
              <Form.Item label="展示标签" name="label" rules={[{ required: true, message: "请输入展示标签" }]}>
                <Input placeholder="如：服饰偏好" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="启用状态" name="status" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="语音文案" name="music_text">
            <Input.TextArea rows={2} placeholder="可选：语音播报文案" />
          </Form.Item>
          <TTSInlinePanel
            title="语音生成"
            text={musicTextValue ?? ""}
            disabled={!version?.id}
            presets={ttsPresets}
            request={request}
            onPresetsReload={refreshTtsPresets}
            onGenerate={(count, options) => {
              if (!version?.id) {
                return Promise.resolve([]);
              }
              return generateTTSBatch(generateTTS, musicTextValue?.trim() || "", "config-extra", version.id, count, options);
            }}
            onSelect={handleSelectTTS}
          />
          <Form.Item label="语音文件路径" name="music">
            <Input placeholder="上传或生成后自动填充" />
          </Form.Item>
          <UploadField
            label="语音文件"
            accept="audio/*"
            value={musicValue}
            previewUrl={musicPreview}
            previewType="audio"
            mediaType="audio"
            moduleKey="config-extra-audio"
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            enableValidation={false}
            enableSmartCompress={false}
            onUpload={async (file) => {
              if (!version?.id) {
                throw new Error("缺少版本信息");
              }
              return uploadFile(file, "config-extra-audio", version.id);
            }}
            onChange={(path, url) => {
              form.setFieldValue("music", path);
              setMusicPreview(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("music", undefined);
              setMusicPreview(null);
            }}
          />
        </Form>
      </Modal>
    </Space>
  );
};

export default ConfigExtraPanel;
