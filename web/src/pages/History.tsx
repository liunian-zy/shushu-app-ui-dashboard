import { Button, Card, Input, Modal, Select, Space, Table, Tabs, Tag, Typography, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { DraftVersion, formatDate } from "./content/constants";

const { Title, Text } = Typography;

type SubmissionItem = {
  id: number;
  module_key: string;
  entity_table: string;
  entity_id: number;
  submit_version: number;
  submit_by: number;
  submit_name?: string | null;
  submit_username?: string | null;
  need_confirm: boolean;
  status: string;
  created_at?: string;
  confirmed_by?: number | null;
  confirmed_name?: string | null;
  confirmed_username?: string | null;
  confirmed_at?: string | null;
  diff?: Array<{ field: string; old: unknown; new: unknown }>;
};

type FieldHistoryItem = {
  id: number;
  entity_table: string;
  entity_id: number;
  field_name: string;
  old_value?: string | null;
  new_value?: string | null;
  submit_id?: number | null;
  changed_by?: number | null;
  changed_name?: string | null;
  changed_username?: string | null;
  created_at?: string | null;
};

type AuditLogItem = {
  id: number;
  draft_version_id?: number | null;
  entity_table?: string | null;
  entity_id?: number | null;
  action?: string | null;
  actor_id?: number | null;
  actor_name?: string | null;
  actor_username?: string | null;
  detail?: unknown;
  created_at?: string | null;
};

type MediaVersionItem = {
  id: number;
  asset_id?: number | null;
  version_no?: number | null;
  file_url?: string | null;
  file_url_signed?: string | null;
  file_size?: number | null;
  width?: number | null;
  height?: number | null;
  duration_ms?: number | null;
  format?: string | null;
  compress_profile?: string | null;
  module_key?: string | null;
  media_type?: string | null;
  created_at?: string | null;
  origin_url_signed?: string | null;
};

const moduleOptions = [
  { value: "banners", label: "轮播图", table: "app_db_banners" },
  { value: "identities", label: "身份信息", table: "app_db_identities" },
  { value: "scenes", label: "场景信息", table: "app_db_scenes" },
  { value: "app_ui_fields", label: "页面配置", table: "app_db_app_ui_fields" },
  { value: "config_extra_steps", label: "额外配置", table: "app_db_config_extra_steps" },
  { value: "clothes_categories", label: "服饰偏好", table: "app_db_clothes_categories" },
  { value: "photo_hobbies", label: "拍摄偏好", table: "app_db_photo_hobbies" }
];

const History = () => {
  const { token } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [versions, setVersions] = useState<DraftVersion[]>([]);
  const [versionLoading, setVersionLoading] = useState(false);
  const [selectedVersionId, setSelectedVersionId] = useState<number | null>(null);
  const [moduleKey, setModuleKey] = useState<string>("");
  const [entityTable, setEntityTable] = useState<string>("");
  const [entityId, setEntityId] = useState<string>("");
  const [actionFilter, setActionFilter] = useState<string>("");
  const [fieldFilter, setFieldFilter] = useState<string>("");
  const [mediaType, setMediaType] = useState<string>("");
  const [submissions, setSubmissions] = useState<SubmissionItem[]>([]);
  const [submissionLoading, setSubmissionLoading] = useState(false);
  const [fieldHistory, setFieldHistory] = useState<FieldHistoryItem[]>([]);
  const [fieldLoading, setFieldLoading] = useState(false);
  const [auditLogs, setAuditLogs] = useState<AuditLogItem[]>([]);
  const [auditLoading, setAuditLoading] = useState(false);
  const [mediaVersions, setMediaVersions] = useState<MediaVersionItem[]>([]);
  const [mediaLoading, setMediaLoading] = useState(false);
  const [diffOpen, setDiffOpen] = useState(false);
  const [diffItems, setDiffItems] = useState<Array<{ field: string; old: unknown; new: unknown }>>([]);

  const request = async <T,>(path: string, options: RequestInit = {}): Promise<T> => {
    if (!token) {
      throw new Error("缺少登录凭证");
    }
    const headers = new Headers(options.headers);
    headers.set("Authorization", `Bearer ${token}`);
    if (options.body && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }
    const response = await fetch(path, { ...options, headers });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error((data as { error?: string }).error || "请求失败");
    }
    return data;
  };

  const loadVersions = async () => {
    setVersionLoading(true);
    try {
      const res = await request<{ data: DraftVersion[] }>("/api/draft/version-names");
      setVersions(res.data || []);
      if (!selectedVersionId && res.data?.length) {
        setSelectedVersionId(res.data[0].id);
      }
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取版本失败");
    } finally {
      setVersionLoading(false);
    }
  };

  useEffect(() => {
    void loadVersions();
  }, []);

  const moduleMap = useMemo(() => {
    const map = new Map<string, string>();
    moduleOptions.forEach((item) => map.set(item.value, item.table));
    return map;
  }, []);

  useEffect(() => {
    const table = moduleMap.get(moduleKey) || "";
    setEntityTable(table);
  }, [moduleKey, moduleMap]);

  const versionOptions = useMemo(
    () =>
      versions.map((item) => ({
        value: item.id,
        label: `${item.location_name || "未命名景区"} / ${item.app_version_name || "未生成版本"}`
      })),
    [versions]
  );

  const loadSubmissions = async () => {
    if (!selectedVersionId || !moduleKey || !entityTable) {
      messageApi.warning("请选择版本与模块");
      return;
    }
    setSubmissionLoading(true);
    try {
      const params = new URLSearchParams({
        draft_version_id: String(selectedVersionId),
        module_key: moduleKey,
        entity_table: entityTable
      });
      if (entityId.trim()) {
        params.set("entity_id", entityId.trim());
      }
      const res = await request<{ data: SubmissionItem[] }>(`/api/draft/submissions?${params.toString()}`);
      setSubmissions(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取提交记录失败");
    } finally {
      setSubmissionLoading(false);
    }
  };

  const loadFieldHistory = async () => {
    if (!selectedVersionId || !entityTable) {
      messageApi.warning("请选择版本与数据表");
      return;
    }
    setFieldLoading(true);
    try {
      const params = new URLSearchParams({
        draft_version_id: String(selectedVersionId),
        entity_table: entityTable
      });
      if (entityId.trim()) {
        params.set("entity_id", entityId.trim());
      }
      if (fieldFilter.trim()) {
        params.set("field_name", fieldFilter.trim());
      }
      const res = await request<{ data: FieldHistoryItem[] }>(`/api/field-history?${params.toString()}`);
      setFieldHistory(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取字段历史失败");
    } finally {
      setFieldLoading(false);
    }
  };

  const loadAuditLogs = async () => {
    if (!selectedVersionId) {
      messageApi.warning("请选择版本");
      return;
    }
    setAuditLoading(true);
    try {
      const params = new URLSearchParams({
        draft_version_id: String(selectedVersionId)
      });
      if (entityTable) {
        params.set("entity_table", entityTable);
      }
      if (entityId.trim()) {
        params.set("entity_id", entityId.trim());
      }
      if (actionFilter.trim()) {
        params.set("action", actionFilter.trim());
      }
      const res = await request<{ data: AuditLogItem[] }>(`/api/audit/logs?${params.toString()}`);
      setAuditLogs(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取审计日志失败");
    } finally {
      setAuditLoading(false);
    }
  };

  const loadMediaVersions = async () => {
    if (!selectedVersionId) {
      messageApi.warning("请选择版本");
      return;
    }
    setMediaLoading(true);
    try {
      const params = new URLSearchParams({
        draft_version_id: String(selectedVersionId)
      });
      if (moduleKey) {
        params.set("module_key", moduleKey);
      }
      if (mediaType.trim()) {
        params.set("media_type", mediaType.trim());
      }
      const res = await request<{ data: MediaVersionItem[] }>(`/api/media/versions?${params.toString()}`);
      setMediaVersions(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取媒体版本失败");
    } finally {
      setMediaLoading(false);
    }
  };

  const baseFilterControls = (
    <Space wrap>
      <Select
        style={{ minWidth: 260 }}
        placeholder="选择景区版本"
        options={versionOptions}
        loading={versionLoading}
        value={selectedVersionId ?? undefined}
        onChange={(value) => setSelectedVersionId(value)}
        allowClear
      />
      <Select
        style={{ minWidth: 200 }}
        placeholder="选择模块"
        options={moduleOptions.map((item) => ({ value: item.value, label: item.label }))}
        value={moduleKey || undefined}
        onChange={(value) => setModuleKey(value)}
        allowClear
      />
      <Input
        style={{ width: 160 }}
        placeholder="实体ID(可选)"
        value={entityId}
        onChange={(event) => setEntityId(event.target.value)}
      />
    </Space>
  );

  const submissionColumns = [
    { title: "版本", dataIndex: "submit_version", key: "submit_version" },
    {
      title: "提交人",
      key: "submit_by",
      render: (_: string, record: SubmissionItem) => (
        <Text>{record.submit_name || record.submit_username || record.submit_by}</Text>
      )
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: string) => {
        if (value === "confirmed") {
          return <Tag color="green">已确认</Tag>;
        }
        if (value === "pending_confirm") {
          return <Tag color="gold">待确认</Tag>;
        }
        return <Tag>已提交</Tag>;
      }
    },
    {
      title: "提交时间",
      dataIndex: "created_at",
      key: "created_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "差异",
      key: "diff",
      render: (_: string, record: SubmissionItem) => (
        <Button
          size="small"
          onClick={() => {
            setDiffItems(record.diff || []);
            setDiffOpen(true);
          }}
        >
          查看
        </Button>
      )
    }
  ];

  const fieldColumns = [
    { title: "字段", dataIndex: "field_name", key: "field_name" },
    {
      title: "旧值",
      dataIndex: "old_value",
      key: "old_value",
      render: (value: string) => <Text type="secondary">{value || "-"}</Text>
    },
    {
      title: "新值",
      dataIndex: "new_value",
      key: "new_value",
      render: (value: string) => <Text type="secondary">{value || "-"}</Text>
    },
    { title: "提交ID", dataIndex: "submit_id", key: "submit_id" },
    {
      title: "操作者",
      key: "operator",
      render: (_: string, record: FieldHistoryItem) => (
        <Text>{record.changed_name || record.changed_username || record.changed_by || "-"}</Text>
      )
    },
    {
      title: "时间",
      dataIndex: "created_at",
      key: "created_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    }
  ];

  const auditColumns = [
    { title: "动作", dataIndex: "action", key: "action" },
    {
      title: "操作者",
      key: "actor",
      render: (_: string, record: AuditLogItem) => (
        <Text>{record.actor_name || record.actor_username || record.actor_id || "-"}</Text>
      )
    },
    {
      title: "实体",
      key: "entity",
      render: (_: string, record: AuditLogItem) => (
        <Text type="secondary">
          {record.entity_table || "-"}#{record.entity_id || "-"}
        </Text>
      )
    },
    {
      title: "详情",
      key: "detail",
      render: (_: string, record: AuditLogItem) => (
        <Text type="secondary" ellipsis={{ tooltip: JSON.stringify(record.detail) }}>
          {record.detail ? JSON.stringify(record.detail) : "-"}
        </Text>
      )
    },
    {
      title: "时间",
      dataIndex: "created_at",
      key: "created_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    }
  ];

  const mediaColumns = [
    { title: "模块", dataIndex: "module_key", key: "module_key" },
    { title: "类型", dataIndex: "media_type", key: "media_type" },
    { title: "版本号", dataIndex: "version_no", key: "version_no" },
    {
      title: "文件",
      key: "file",
      render: (_: string, record: MediaVersionItem) =>
        record.file_url_signed ? (
          <a href={record.file_url_signed} target="_blank" rel="noreferrer">
            查看
          </a>
        ) : (
          "-"
        )
    },
    {
      title: "大小",
      dataIndex: "file_size",
      key: "file_size",
      render: (value: number) => <Text type="secondary">{value ? `${value}B` : "-"}</Text>
    },
    {
      title: "尺寸",
      key: "size",
      render: (_: string, record: MediaVersionItem) => (
        <Text type="secondary">
          {record.width || 0}x{record.height || 0}
        </Text>
      )
    },
    {
      title: "时间",
      dataIndex: "created_at",
      key: "created_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Title level={3} style={{ margin: 0 }}>
        操作历史
      </Title>
      <Card style={{ borderRadius: 20 }}>
        <Text type="secondary">
          追溯每次提交、差异比对与媒体处理的完整记录。
        </Text>
      </Card>
      <Tabs
        items={[
          {
            key: "submissions",
            label: "提交记录",
            children: (
              <Space direction="vertical" size={16} style={{ width: "100%" }}>
                <Card style={{ borderRadius: 20 }}>
                  <Space direction="vertical" size={12} style={{ width: "100%" }}>
                    {baseFilterControls}
                    <Button onClick={loadSubmissions} loading={submissionLoading}>
                      查询提交记录
                    </Button>
                  </Space>
                </Card>
                <Card style={{ borderRadius: 20 }}>
                  <Table
                    rowKey="id"
                    columns={submissionColumns}
                    dataSource={submissions}
                    loading={submissionLoading}
                    pagination={{ pageSize: 8 }}
                  />
                </Card>
              </Space>
            )
          },
          {
            key: "fields",
            label: "字段历史",
            children: (
              <Space direction="vertical" size={16} style={{ width: "100%" }}>
                <Card style={{ borderRadius: 20 }}>
                  <Space direction="vertical" size={12} style={{ width: "100%" }}>
                    {baseFilterControls}
                    <Input
                      style={{ width: 200 }}
                      placeholder="字段名过滤(可选)"
                      value={fieldFilter}
                      onChange={(event) => setFieldFilter(event.target.value)}
                    />
                    <Button onClick={loadFieldHistory} loading={fieldLoading}>
                      查询字段历史
                    </Button>
                  </Space>
                </Card>
                <Card style={{ borderRadius: 20 }}>
                  <Table
                    rowKey="id"
                    columns={fieldColumns}
                    dataSource={fieldHistory}
                    loading={fieldLoading}
                    pagination={{ pageSize: 8 }}
                  />
                </Card>
              </Space>
            )
          },
          {
            key: "media",
            label: "媒体版本",
            children: (
              <Space direction="vertical" size={16} style={{ width: "100%" }}>
                <Card style={{ borderRadius: 20 }}>
                  <Space direction="vertical" size={12} style={{ width: "100%" }}>
                    {baseFilterControls}
                    <Input
                      style={{ width: 160 }}
                      placeholder="媒体类型(可选)"
                      value={mediaType}
                      onChange={(event) => setMediaType(event.target.value)}
                    />
                    <Button onClick={loadMediaVersions} loading={mediaLoading}>
                      查询媒体版本
                    </Button>
                  </Space>
                </Card>
                <Card style={{ borderRadius: 20 }}>
                  <Table
                    rowKey="id"
                    columns={mediaColumns}
                    dataSource={mediaVersions}
                    loading={mediaLoading}
                    pagination={{ pageSize: 8 }}
                  />
                </Card>
              </Space>
            )
          },
          {
            key: "audit",
            label: "审计日志",
            children: (
              <Space direction="vertical" size={16} style={{ width: "100%" }}>
                <Card style={{ borderRadius: 20 }}>
                  <Space direction="vertical" size={12} style={{ width: "100%" }}>
                    {baseFilterControls}
                    <Input
                      style={{ width: 200 }}
                      placeholder="动作过滤(可选)"
                      value={actionFilter}
                      onChange={(event) => setActionFilter(event.target.value)}
                    />
                    <Button onClick={loadAuditLogs} loading={auditLoading}>
                      查询审计日志
                    </Button>
                  </Space>
                </Card>
                <Card style={{ borderRadius: 20 }}>
                  <Table
                    rowKey="id"
                    columns={auditColumns}
                    dataSource={auditLogs}
                    loading={auditLoading}
                    pagination={{ pageSize: 8 }}
                  />
                </Card>
              </Space>
            )
          }
        ]}
      />
      <Card style={{ borderRadius: 20 }}>
        <Text type="secondary">
          提示：提交记录与字段历史需要选择模块与版本；媒体版本与审计日志只需版本即可查询。
        </Text>
      </Card>
      <Modal
        title="差异对比"
        open={diffOpen}
        onCancel={() => setDiffOpen(false)}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        footer={[
          <Button key="close" onClick={() => setDiffOpen(false)}>
            关闭
          </Button>
        ]}
      >
        <Table
          rowKey={(record, index) => `${record.field}-${index}`}
          dataSource={diffItems}
          pagination={false}
          columns={[
            { title: "字段", dataIndex: "field", key: "field" },
            {
              title: "旧值",
              dataIndex: "old",
              key: "old",
              render: (value: unknown) => <Text type="secondary">{JSON.stringify(value)}</Text>
            },
            {
              title: "新值",
              dataIndex: "new",
              key: "new",
              render: (value: unknown) => <Text type="secondary">{JSON.stringify(value)}</Text>
            }
          ]}
        />
      </Modal>
    </Space>
  );
};

export default History;
