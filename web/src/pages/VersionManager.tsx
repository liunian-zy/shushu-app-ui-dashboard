import { useEffect, useMemo, useState } from "react";
import { Button, Card, Empty, Input, Modal, Space, Table, Tag, Typography, message } from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useAuth } from "../contexts/AuthContext";
import VersionEditorModal, { VersionEditorValues } from "./version/VersionEditorModal";
import { DraftVersion, formatDate } from "./version/constants";

const { Title, Text } = Typography;
const { Search } = Input;

const VersionManager = () => {
  const { token, user } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [versions, setVersions] = useState<DraftVersion[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [editingVersion, setEditingVersion] = useState<DraftVersion | null>(null);
  const [syncingId, setSyncingId] = useState<number | null>(null);

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
    return data as T;
  };

  const loadVersions = async () => {
    setLoading(true);
    try {
      const res = await request<{ data: DraftVersion[] }>("/api/draft/version-names");
      setVersions(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取版本失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadVersions();
  }, []);

  const filteredVersions = useMemo(() => {
    if (!searchText.trim()) {
      return versions;
    }
    const keyword = searchText.trim();
    return versions.filter((item) => {
      const locationName = item.location_name ?? "";
      const versionName = item.app_version_name ?? "";
      return locationName.includes(keyword) || versionName.includes(keyword);
    });
  }, [versions, searchText]);

  const openEditor = (version?: DraftVersion) => {
    setEditingVersion(version ?? null);
    setEditorOpen(true);
  };

  const handleSubmit = async (values: VersionEditorValues) => {
    try {
      setEditorSubmitting(true);
      const payload: Record<string, unknown> = {
        location_name: values.location_name,
        ai_modal: values.ai_modal || "SD"
      };
      if (values.app_version_name && values.app_version_name.trim()) {
        payload.app_version_name = values.app_version_name.trim();
      }
      if (values.feishu_field_names && values.feishu_field_names.length > 0) {
        payload.feishu_field_names = values.feishu_field_names;
      }
      if (editingVersion) {
        await request(`/api/draft/version-names/${editingVersion.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        messageApi.success("版本已更新");
      } else {
        await request("/api/draft/version-names", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        messageApi.success("版本已创建");
      }
      setEditorOpen(false);
      setEditingVersion(null);
      void loadVersions();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setEditorSubmitting(false);
    }
  };

  const handleDelete = (version: DraftVersion) => {
    Modal.confirm({
      title: "确认删除版本？",
      content: `版本 ${version.location_name || version.app_version_name || ""} 将被移除。`,
      okText: "确认删除",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await request(`/api/draft/version-names/${version.id}`, { method: "DELETE" });
          messageApi.success("已删除");
          void loadVersions();
        } catch (error) {
          if (error instanceof Error) {
            messageApi.error(error.message);
          }
        }
      }
    });
  };

  const handleSync = async (version: DraftVersion, confirm = false) => {
    if (!token) {
      messageApi.warning("缺少登录凭证");
      return;
    }
    if (!user?.id) {
      messageApi.warning("缺少提交人信息");
      return;
    }
    setSyncingId(version.id);
    try {
      const response = await fetch("/api/sync", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          draft_version_id: version.id,
          trigger_by: user.id,
          confirm,
          modules: ["version_names"]
        })
      });
      const data = await response.json().catch(() => ({}));
      if (response.status === 409 && (data as { need_confirm?: boolean }).need_confirm) {
        setSyncingId(null);
        Modal.confirm({
          title: "检测到线上已有版本",
          content: "继续同步将覆盖线上同名景区的版本配置。",
          okText: "继续同步",
          cancelText: "取消",
          onOk: () => handleSync(version, true)
        });
        return;
      }
      if (!response.ok) {
        throw new Error((data as { error?: string }).error || "同步失败");
      }
      messageApi.success("版本配置已同步");
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setSyncingId(null);
    }
  };

  const renderFields = (fields?: string[] | null) => {
    if (!fields || fields.length === 0) {
      return <Text type="secondary">未配置</Text>;
    }
    const visible = fields.slice(0, 4);
    const remaining = fields.length - visible.length;
    return (
      <Space wrap size={4}>
        {visible.map((item) => (
          <Tag key={item}>{item}</Tag>
        ))}
        {remaining > 0 ? <Tag>+{remaining}</Tag> : null}
      </Space>
    );
  };

  const columns = [
    {
      title: "景区",
      dataIndex: "location_name",
      key: "location_name",
      render: (value: string) => <Text strong>{value || "未命名"}</Text>
    },
    {
      title: "版本名",
      dataIndex: "app_version_name",
      key: "app_version_name",
      render: (value: string) => <Tag color="orange">{value || "自动生成"}</Tag>
    },
    {
      title: "模型",
      dataIndex: "ai_modal",
      key: "ai_modal",
      render: (value: string) => <Tag color="geekblue">{value || "SD"}</Tag>
    },
    {
      title: "Feishu 字段",
      dataIndex: "feishu_field_list",
      key: "feishu_field_list",
      render: (_: string, record: DraftVersion) => renderFields(record.feishu_field_list || [])
    },
    {
      title: "提交版本",
      dataIndex: "submit_version",
      key: "submit_version",
      render: (value: number) => <Text>{value ?? "-"}</Text>
    },
    {
      title: "最近提交",
      dataIndex: "last_submit_at",
      key: "last_submit_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: DraftVersion) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEditor(record)}>
            编辑
          </Button>
          <Button
            size="small"
            type="primary"
            loading={syncingId === record.id}
            onClick={() => handleSync(record, false)}
          >
            同步
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
      {contextHolder}
      <Title level={3} style={{ margin: 0 }}>
        版本配置
      </Title>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <Text type="secondary">
            创建景区版本、自动生成版本名、配置 AI 模型与 Feishu 字段。
          </Text>
          <Space wrap>
            <Search
              placeholder="搜索景区名称或版本名"
              allowClear
              onSearch={(value) => setSearchText(value)}
              onChange={(event) => setSearchText(event.target.value)}
              style={{ width: 240 }}
            />
            <Button icon={<ReloadOutlined />} onClick={loadVersions}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>
              新建版本
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        {filteredVersions.length ? (
          <Table
            rowKey="id"
            columns={columns}
            dataSource={filteredVersions}
            loading={loading}
            pagination={{ pageSize: 6 }}
          />
        ) : (
          <Empty description="暂无版本记录" />
        )}
      </Card>

      <VersionEditorModal
        open={editorOpen}
        submitting={editorSubmitting}
        version={editingVersion}
        onCancel={() => {
          setEditorOpen(false);
          setEditingVersion(null);
        }}
        onSubmit={handleSubmit}
      />
    </Space>
  );
};

export default VersionManager;
