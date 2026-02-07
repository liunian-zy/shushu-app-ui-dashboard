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
import { DraftVersion, IdentityItem, identityNameOptions, formatDate, statusOptions, submitStatusLabels } from "./constants";
import type { Notify, RequestFn, UploadFn } from "./utils";
import { buildLocalDraftKey, loadLocalDraft, saveLocalDraft, sanitizeSubmissionPayload } from "./utils";

const { Text } = Typography;

type IdentityPanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  notify: Notify;
  operatorId?: number | null;
};

type IdentityFormValues = {
  name?: string;
  image?: string;
  sort?: number;
  status?: boolean;
};

const IdentityPanel = ({ version, request, uploadFile, notify, operatorId }: IdentityPanelProps) => {
  const [items, setItems] = useState<IdentityItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [templates, setTemplates] = useState<Array<{ id: number; name?: string | null }>>([]);
  const [templateLoading, setTemplateLoading] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingItem, setEditingItem] = useState<IdentityItem | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<IdentityItem[]>([]);
  const [batchSubmitLoading, setBatchSubmitLoading] = useState(false);
  const [sortMode, setSortMode] = useState(false);
  const [sortSaving, setSortSaving] = useState(false);
  const [sortDraft, setSortDraft] = useState<IdentityItem[]>([]);
  const [draggingSortId, setDraggingSortId] = useState<number | null>(null);
  const [statusUpdating, setStatusUpdating] = useState<Record<number, boolean>>({});
  const [filterName, setFilterName] = useState<string>("");
  const [filterSubmitStatus, setFilterSubmitStatus] = useState<string | null>(null);
  const [filterStatus, setFilterStatus] = useState<number | null>(null);
  const [form] = Form.useForm<IdentityFormValues>();
  const imageValue = Form.useWatch("image", form);
  const draftKey = version?.id ? buildLocalDraftKey("identities", version.id, editingItem?.id) : "";

  const loadItems = async () => {
    if (!version?.id) {
      setItems([]);
      return;
    }
    setLoading(true);
    try {
      const res = await request<{ data: IdentityItem[] }>(`/api/draft/identities?draft_version_id=${version.id}`);
      setItems(res.data || []);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取身份信息失败");
    } finally {
      setLoading(false);
    }
  };

  const loadTemplates = async () => {
    setTemplateLoading(true);
    try {
      const res = await request<{ data: Array<{ id: number; name?: string | null }> }>("/api/identity-templates");
      setTemplates(res.data || []);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取模板失败");
    } finally {
      setTemplateLoading(false);
    }
  };

  useEffect(() => {
    void loadItems();
  }, [version?.id]);

  useEffect(() => {
    setSelectedRowKeys([]);
    setSelectedRows([]);
  }, [version?.id]);

  useEffect(() => {
    void loadTemplates();
  }, []);

  const openEditor = (item?: IdentityItem) => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    setEditingItem(item ?? null);
    setPreviewUrl(item?.image_url ?? null);
    setEditorOpen(true);
  };

  const syncEditorForm = (item: IdentityItem | null) => {
    form.setFieldsValue({
      name: item?.name ?? undefined,
      image: item?.image ?? undefined,
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
        image: values.image || undefined,
        sort: values.sort ?? 0,
        status: values.status ? 1 : 0,
        updated_by: operatorId ?? undefined
      };
      if (!editingItem && operatorId) {
        payload.created_by = operatorId;
      }
      if (editingItem) {
        await request(`/api/draft/identities/${editingItem.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        notify.success("身份信息已更新");
      } else {
        await request("/api/draft/identities", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        notify.success("身份信息已创建");
      }
      setEditorOpen(false);
      setEditingItem(null);
      setPreviewUrl(null);
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

  const handleDelete = (item: IdentityItem) => {
    Modal.confirm({
      title: "确认删除身份？",
      content: item.name || "该身份将被移除。",
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/draft/identities/${item.id}`, { method: "DELETE" });
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
              module_key: "identities",
              entity_table: "app_db_identities",
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

  const handleApplyTemplate = () => {
    if (!version?.id) {
      notify.warning("请先选择景区版本");
      return;
    }
    if (!selectedTemplateId) {
      notify.warning("请选择身份模板");
      return;
    }
    Modal.confirm({
      title: "确认套用模板？",
      content: "将用模板内容覆盖当前身份信息。",
      okText: "确认套用",
      cancelText: "取消",
      onOk: async () => {
        try {
          await request("/api/draft/identities/apply-template", {
            method: "POST",
            body: JSON.stringify({
              draft_version_id: version.id,
              template_id: selectedTemplateId,
              replace: true
            })
          });
          notify.success("模板已应用");
          void loadItems();
        } catch (error) {
          if (error instanceof Error) {
            notify.error(error.message);
          }
        }
      }
    });
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

  const moveSortItem = (dragId: number, hoverId: number) => {
    if (dragId === hoverId) {
      return;
    }
    setSortDraft((prev) => {
      const from = prev.findIndex((item) => item.id === dragId);
      const to = prev.findIndex((item) => item.id === hoverId);
      if (from < 0 || to < 0) {
        return prev;
      }
      const next = [...prev];
      const [moved] = next.splice(from, 1);
      next.splice(to, 0, moved);
      return next;
    });
  };

  const enterSortMode = () => {
    if (!filteredItems.length) {
      notify.warning("没有可排序的数据");
      return;
    }
    setSortDraft(filteredItems);
    setSortMode(true);
  };

  const cancelSortMode = () => {
    setSortMode(false);
    setDraggingSortId(null);
    setSortDraft([]);
  };

  const applySortMode = async () => {
    if (!sortDraft.length) {
      cancelSortMode();
      return;
    }
    setSortSaving(true);
    try {
      const total = sortDraft.length;
      const results = await Promise.allSettled(
        sortDraft.map((item, index) =>
          request(`/api/draft/identities/${item.id}`, {
            method: "PUT",
            body: JSON.stringify({
              sort: total - index,
              updated_by: operatorId ?? undefined
            })
          })
        )
      );
      const failed = results.filter((item) => item.status === "rejected").length;
      if (failed > 0) {
        notify.warning(`批量排序完成，失败 ${failed} 条`);
      } else {
        notify.success("批量排序已应用");
      }
      cancelSortMode();
      void loadItems();
    } finally {
      setSortSaving(false);
    }
  };

  const handleInlineStatusChange = async (record: IdentityItem, checked: boolean) => {
    setStatusUpdating((prev) => ({ ...prev, [record.id]: true }));
    try {
      await request(`/api/draft/identities/${record.id}`, {
        method: "PUT",
        body: JSON.stringify({
          status: checked ? 1 : 0,
          updated_by: operatorId ?? undefined
        })
      });
      notify.success("状态已更新");
      void loadItems();
    } catch (error) {
      if (error instanceof Error) {
        notify.error(error.message);
      }
    } finally {
      setStatusUpdating((prev) => ({ ...prev, [record.id]: false }));
    }
  };

  const filteredItems = useMemo(() => {
    const keyword = filterName.trim();
    return items
      .filter((item) => {
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
      })
      .sort((a, b) => (b.sort ?? 0) - (a.sort ?? 0) || a.id - b.id);
  }, [filterName, filterStatus, filterSubmitStatus, items]);

  const columns = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      width: 80
    },
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
      render: (value: number, record: IdentityItem) => (
        <Switch
          checked={(value ?? 1) === 1}
          disabled={sortMode}
          loading={!!statusUpdating[record.id]}
          onChange={(checked) => handleInlineStatusChange(record, checked)}
        />
      )
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
        value ? <Image src={value} width={72} style={{ borderRadius: 8 }} /> : "-"
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: IdentityItem) => (
        <Space>
          <SubmissionActions
            draftVersionId={version?.id || 0}
            moduleKey="identities"
            entityTable="app_db_identities"
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
          <Text type="secondary">身份默认为男士、女士、男孩、女孩，可按需启用或停用。</Text>
          <Space wrap>
            <Select
              style={{ minWidth: 220 }}
              placeholder="选择身份模板"
              loading={templateLoading}
              options={templates.map((item) => ({
                value: item.id,
                label: item.name || `模板${item.id}`
              }))}
              value={selectedTemplateId ?? undefined}
              onChange={(value) => setSelectedTemplateId(value)}
              allowClear
            />
            <Button onClick={handleApplyTemplate} disabled={!version?.id}>
              使用模板
            </Button>
          </Space>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadItems} disabled={!version?.id}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()} disabled={!version?.id}>
              新建身份
            </Button>
            {!sortMode ? (
              <Button onClick={enterSortMode} disabled={!version?.id || !filteredItems.length}>
                进入排序模式
              </Button>
            ) : (
              <>
                <Button type="primary" onClick={applySortMode} loading={sortSaving}>
                  应用排序
                </Button>
                <Button onClick={cancelSortMode}>取消排序</Button>
              </>
            )}
            <Button onClick={handleBatchSubmission} loading={batchSubmitLoading} disabled={!selectedRowKeys.length || !version?.id}>
              批量提交
            </Button>
          </Space>
          <Space wrap>
            <Text type="secondary">筛选</Text>
            <Input
              placeholder="身份名称"
              value={filterName}
              onChange={(event) => setFilterName(event.target.value)}
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
            dataSource={sortMode ? sortDraft : filteredItems}
            loading={loading}
            pagination={{ pageSize: 10, showSizeChanger: true, pageSizeOptions: ["10", "20", "50"] }}
            rowSelection={
              sortMode
                ? undefined
                : {
                    selectedRowKeys,
                    onChange: (keys, rows) => {
                      setSelectedRowKeys(keys);
                      setSelectedRows(rows as IdentityItem[]);
                    }
                  }
            }
            onRow={(record) =>
              sortMode
                ? {
                    draggable: true,
                    style: { cursor: "move" },
                    onDragStart: () => setDraggingSortId(record.id),
                    onDragOver: (event) => {
                      event.preventDefault();
                      if (draggingSortId != null) {
                        moveSortItem(draggingSortId, record.id);
                      }
                    },
                    onDrop: (event) => {
                      event.preventDefault();
                      setDraggingSortId(null);
                    },
                    onDragEnd: () => setDraggingSortId(null)
                  }
                : {}
            }
          />
        ) : (
          <Empty description="请选择景区版本后录入身份信息" />
        )}
      </Card>


      <Modal
        title={editingItem ? "编辑身份" : "新建身份"}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false);
          setEditingItem(null);
          setPreviewUrl(null);
          form.resetFields();
        }}
        afterOpenChange={(open) => {
          if (open) {
            syncEditorForm(editingItem);
          }
        }}
        width={860}
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
              setPreviewUrl(null);
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
          <Form.Item label="身份名称" name="name" rules={[{ required: true, message: "请输入身份名称" }]}>
            <Input placeholder="如：男士" />
          </Form.Item>
          <Space wrap>
            {identityNameOptions.map((item) => (
              <Button key={item} size="small" onClick={() => form.setFieldValue("name", item)}>
                {item}
              </Button>
            ))}
          </Space>
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
          <Form.Item label="头像路径" name="image" rules={[{ required: true, message: "请上传头像图片" }]}>
            <Input placeholder="上传后自动填充" />
          </Form.Item>
          <UploadField
            label="头像图片"
            accept="image/png,image/jpeg"
            value={imageValue}
            previewUrl={previewUrl}
            previewType="image"
            mediaType="image"
            moduleKey="identities"
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            onUpload={async (file) => {
              if (!version?.id) {
                throw new Error("缺少版本信息");
              }
              return uploadFile(file, "identities", version.id);
            }}
            onChange={(path, url) => {
              form.setFieldValue("image", path);
              setPreviewUrl(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("image", undefined);
              setPreviewUrl(null);
            }}
          />
        </Form>
      </Modal>
    </Space>
  );
};

export default IdentityPanel;
