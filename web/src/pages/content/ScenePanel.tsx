import { useEffect, useMemo, useState } from "react";
import type { Key } from "react";
import {
  Button,
  Card,
  Col,
  Empty,
  Form,
  Image,
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
import { DraftVersion, SceneItem, formatDate, statusOptions, submitStatusLabels } from "./constants";
import type { Notify, RequestFn, TTSPreset, TTSFn, TTSResult, UploadFn } from "./utils";
import { buildLocalDraftKey, generateTTSBatch, loadLocalDraft, saveLocalDraft, sanitizeSubmissionPayload } from "./utils";

const { Text } = Typography;
const { TextArea } = Input;

type ScenePanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  generateTTS: TTSFn;
  notify: Notify;
  operatorId?: number | null;
  ttsPresets?: TTSPreset[];
};

type SceneFormValues = {
  name?: string;
  desc?: string;
  image?: string;
  music?: string;
  sort?: number;
  status?: boolean;
};

const ScenePanel = ({ version, request, uploadFile, generateTTS, notify, operatorId, ttsPresets }: ScenePanelProps) => {
  const [items, setItems] = useState<SceneItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<SceneItem | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [musicPreview, setMusicPreview] = useState<string | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<SceneItem[]>([]);
  const [batchSubmitLoading, setBatchSubmitLoading] = useState(false);
  const [filterName, setFilterName] = useState<string>("");
  const [filterSubmitStatus, setFilterSubmitStatus] = useState<string | null>(null);
  const [filterStatus, setFilterStatus] = useState<number | null>(null);
  const [form] = Form.useForm<SceneFormValues>();
  const imageValue = Form.useWatch("image", form);
  const musicValue = Form.useWatch("music", form);
  const descValue = Form.useWatch("desc", form);
  const draftKey = version?.id ? buildLocalDraftKey("scenes", version.id, editingItem?.id) : "";

  const loadItems = async () => {
    if (!version?.id) {
      setItems([]);
      return;
    }
    setLoading(true);
    try {
      const res = await request<{ data: SceneItem[] }>(`/api/draft/scenes?draft_version_id=${version.id}`);
      setItems(res.data || []);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取场景失败");
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

  const openEditor = (item?: SceneItem) => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    setEditingItem(item ?? null);
    setImagePreview(item?.image_url ?? null);
    setMusicPreview(item?.music_url ?? null);
    setEditorOpen(true);
  };

  const syncEditorForm = (item: SceneItem | null) => {
    form.setFieldsValue({
      name: item?.name ?? undefined,
      desc: item?.desc ?? undefined,
      image: item?.image ?? undefined,
      music: item?.music ?? undefined,
      sort: item?.sort ?? 0,
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
        app_version_name: version.app_version_name || undefined,
        name: values.name?.trim() || undefined,
        desc: values.desc?.trim() || undefined,
        image: values.image || undefined,
        music: values.music || undefined,
        sort: values.sort ?? 0,
        status: values.status ? 1 : 0,
        updated_by: operatorId ?? undefined
      };
      if (!editingItem && operatorId) {
        payload.created_by = operatorId;
      }
      if (editingItem) {
        await request(`/api/draft/scenes/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        notify.success("场景已更新");
      } else {
        await request("/api/draft/scenes", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        notify.success("场景已创建");
      }
      setEditorOpen(false);
      setEditingItem(null);
      setImagePreview(null);
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

  const handleDelete = (item: SceneItem) => {
    Modal.confirm({
      title: "确认删除场景？",
      content: item.name || "该场景将被移除。",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/draft/scenes/${item.id}`, { method: "DELETE" });
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
    const keyword = filterName.trim();
    return items.filter((item) => {
      if (keyword && !(item.name || "").includes(keyword)) {
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
  }, [filterName, filterStatus, filterSubmitStatus, items]);

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
              module_key: "scenes",
              entity_table: "app_db_scenes",
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
      title: "场景",
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
      title: "预览",
      dataIndex: "image_url",
      key: "image_url",
      render: (value: string) =>
        value ? <Image src={value} width={80} style={{ borderRadius: 8 }} /> : "-"
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: SceneItem) => (
        <Space>
          <SubmissionActions
            draftVersionId={version?.id || 0}
            moduleKey="scenes"
            entityTable="app_db_scenes"
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
          <Text type="secondary">场景包含图片、介绍与语音文件。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadItems} disabled={!version?.id}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()} disabled={!version?.id}>
              新建场景
            </Button>
            <Button onClick={handleBatchSubmission} loading={batchSubmitLoading} disabled={!selectedRowKeys.length || !version?.id}>
              批量提交
            </Button>
          </Space>
          <Space wrap>
            <Text type="secondary">筛选</Text>
            <Input
              placeholder="场景名称"
              value={filterName}
              onChange={(event) => setFilterName(event.target.value)}
              allowClear
              style={{ width: 180 }}
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
                setFilterName("");
                setFilterSubmitStatus(null);
                setFilterStatus(null);
              }}
              disabled={!filterName && !filterSubmitStatus && filterStatus == null}
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
                setSelectedRows(rows as SceneItem[]);
              }
            }}
          />
        ) : (
          <Empty description="请选择景区版本后录入场景信息" />
        )}
      </Card>

      <Modal
        title={editingItem ? "编辑场景" : "新建场景"}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false);
          setEditingItem(null);
          setImagePreview(null);
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
              setImagePreview(null);
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
          <Form.Item label="场景名称" name="name" rules={[{ required: true, message: "请输入场景名称" }]}>
            <Input placeholder="如：展厅入口" />
          </Form.Item>
          <Form.Item label="场景介绍/语音文案" name="desc">
            <TextArea rows={3} placeholder="用于场景介绍或语音播报" />
          </Form.Item>
          <TTSInlinePanel
            title="语音生成"
            text={descValue ?? ""}
            disabled={!version?.id}
            presets={ttsPresets}
            onGenerate={(count, options) => {
              if (!version?.id) {
                return Promise.resolve([]);
              }
              return generateTTSBatch(generateTTS, descValue?.trim() || "", "scenes", version.id, count, options);
            }}
            onSelect={handleSelectTTS}
          />
          <Form.Item label="场景图片路径" name="image" rules={[{ required: true, message: "请上传场景图片" }]}>
            <Input placeholder="上传后自动填充" />
          </Form.Item>
          <UploadField
            label="场景图片"
            accept="image/png,image/jpeg"
            value={imageValue}
            previewUrl={imagePreview}
            previewType="image"
            mediaType="image"
            moduleKey="scenes"
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            onUpload={async (file) => {
              if (!version?.id) {
                throw new Error("缺少版本信息");
              }
              return uploadFile(file, "scenes", version.id);
            }}
            onChange={(path, url) => {
              form.setFieldValue("image", path);
              setImagePreview(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("image", undefined);
              setImagePreview(null);
            }}
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
            moduleKey="scene-audio"
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
              return uploadFile(file, "scene-audio", version.id);
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
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item label="排序" name="sort">
                <InputNumber min={0} max={999} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="启用状态" name="status" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </Space>
  );
};

export default ScenePanel;
