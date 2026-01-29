import { useEffect, useMemo, useState } from "react";
import { Button, Card, Empty, Modal, Select, Space, Switch, Table, Tag, Typography, message } from "antd";
import { EditOutlined, EyeOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useAuth } from "../contexts/AuthContext";
import TaskActionsDrawer from "./taskboard/TaskActionsDrawer";
import TaskAssistModal from "./taskboard/TaskAssistModal";
import TaskEditorModal, { TaskEditorValues } from "./taskboard/TaskEditorModal";
import {
  DraftVersion,
  TaskAction,
  TaskItem,
  UserItem,
  formatDate,
  moduleOptions,
  resolveModuleLabel,
  resolveStatusMeta,
  statusOptions
} from "./taskboard/constants";

const { Title, Text } = Typography;

const TaskBoard = () => {
  const { token, user } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [versions, setVersions] = useState<DraftVersion[]>([]);
  const [versionLoading, setVersionLoading] = useState(false);
  const [selectedVersionId, setSelectedVersionId] = useState<number | null>(null);
  const [tasks, setTasks] = useState<TaskItem[]>([]);
  const [taskLoading, setTaskLoading] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string | null>(null);
  const [moduleFilter, setModuleFilter] = useState<string | null>(null);
  const [assignedFilter, setAssignedFilter] = useState<number | null>(null);
  const [onlyMine, setOnlyMine] = useState(false);
  const [users, setUsers] = useState<UserItem[]>([]);
  const [userLoading, setUserLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<TaskItem | null>(null);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [assistOpen, setAssistOpen] = useState(false);
  const [assistSubmitting, setAssistSubmitting] = useState(false);
  const [assistTask, setAssistTask] = useState<TaskItem | null>(null);
  const [actionsOpen, setActionsOpen] = useState(false);
  const [actionsLoading, setActionsLoading] = useState(false);
  const [actions, setActions] = useState<TaskAction[]>([]);
  const [completingTaskId, setCompletingTaskId] = useState<number | null>(null);

  const isAdmin = user?.role === "admin";

  const userMap = useMemo(() => {
    const map = new Map<number, UserItem>();
    users.forEach((item) => {
      map.set(item.id, item);
    });
    return map;
  }, [users]);

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

  const loadUsers = async () => {
    if (!isAdmin) {
      return;
    }
    setUserLoading(true);
    try {
      const res = await request<{ data: UserItem[] }>("/api/users");
      setUsers(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取用户失败");
    } finally {
      setUserLoading(false);
    }
  };

  const loadTasks = async () => {
    if (!selectedVersionId) {
      setTasks([]);
      return;
    }
    setTaskLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("draft_version_id", String(selectedVersionId));
      if (statusFilter) {
        params.set("status", statusFilter);
      }
      if (moduleFilter) {
        params.set("module_key", moduleFilter);
      }
      const queryAssigned = isAdmin ? assignedFilter : onlyMine ? user?.id ?? null : null;
      if (queryAssigned) {
        params.set("assigned_to", String(queryAssigned));
      }
      const res = await request<{ data: TaskItem[] }>(`/api/tasks?${params.toString()}`);
      setTasks(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取任务失败");
    } finally {
      setTaskLoading(false);
    }
  };

  const loadActions = async (taskId: number) => {
    setActionsLoading(true);
    try {
      const res = await request<{ data: TaskAction[] }>(`/api/tasks/${taskId}/actions`);
      setActions(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取历史失败");
    } finally {
      setActionsLoading(false);
    }
  };

  useEffect(() => {
    void loadVersions();
  }, []);

  useEffect(() => {
    void loadUsers();
  }, [isAdmin]);

  useEffect(() => {
    void loadTasks();
  }, [selectedVersionId, statusFilter, moduleFilter, assignedFilter, onlyMine, isAdmin]);

  const versionOptions = useMemo(
    () =>
      versions.map((item) => ({
        value: item.id,
        label: `${item.location_name || "未命名景区"} / ${item.app_version_name || "未生成版本"}`
      })),
    [versions]
  );

  const userOptions = useMemo(
    () =>
      users.map((item) => ({
        value: item.id,
        label: item.display_name || item.username || `用户${item.id}`
      })),
    [users]
  );

  const openEditor = (task?: TaskItem) => {
    if (!selectedVersionId) {
      messageApi.warning("请先选择景区版本");
      return;
    }
    setEditingTask(task ?? null);
    setEditorOpen(true);
  };

  const handleSaveTask = async (values: TaskEditorValues) => {
    try {
      if (!selectedVersionId) {
        return;
      }
      setEditorSubmitting(true);
      const payload = {
        module_key: values.module_key,
        title: values.title,
        description: values.description,
        assigned_to: values.assigned_to ?? null,
        allow_assist: values.allow_assist ? 1 : 0,
        priority: values.priority ?? 0,
        status: values.status
      };
      if (editingTask) {
        await request(`/api/tasks/${editingTask.id}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
        messageApi.success("任务已更新");
      } else {
        await request("/api/tasks", {
          method: "POST",
          body: JSON.stringify({
            draft_version_id: selectedVersionId,
            ...payload
          })
        });
        messageApi.success("任务已创建");
      }
      setEditorOpen(false);
      setEditingTask(null);
      void loadTasks();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setEditorSubmitting(false);
    }
  };

  const handleAssist = async (note?: string) => {
    if (!assistTask) {
      return;
    }
    try {
      setAssistSubmitting(true);
      await request(`/api/tasks/${assistTask.id}/assist`, {
        method: "POST",
        body: JSON.stringify({ note })
      });
      messageApi.success("协作记录已提交");
      setAssistOpen(false);
      void loadTasks();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setAssistSubmitting(false);
    }
  };

  const openAssist = (task: TaskItem) => {
    setAssistTask(task);
    setAssistOpen(true);
  };

  const openActions = (task: TaskItem) => {
    setActionsOpen(true);
    void loadActions(task.id);
  };

  const completeTask = (task: TaskItem) => {
    Modal.confirm({
      title: "确认完成并上传？",
      content: "系统将上传该任务相关的本地文件到 OSS，并更新草稿路径。",
      okText: "开始上传",
      cancelText: "取消",
      onOk: async () => {
        try {
          setCompletingTaskId(task.id);
          await request(`/api/tasks/${task.id}/complete`, { method: "POST" });
          messageApi.success("任务已完成并上传");
          void loadTasks();
        } catch (error) {
          if (error instanceof Error) {
            messageApi.error(error.message);
          }
        } finally {
          setCompletingTaskId(null);
        }
      }
    });
  };

  const renderAssignee = (assignedTo?: number | null) => {
    if (!assignedTo) {
      return <Text type="secondary">未指派</Text>;
    }
    if (assignedTo === user?.id) {
      return <Tag color="blue">我</Tag>;
    }
    const detail = userMap.get(assignedTo);
    const label = detail?.display_name || detail?.username || `用户${assignedTo}`;
    return <Text>{label}</Text>;
  };

  const columns = [
    {
      title: "任务",
      dataIndex: "title",
      key: "title",
      render: (_: string, record: TaskItem) => (
        <Space direction="vertical" size={4}>
          <Text strong>{record.title || "未命名任务"}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.description || "暂无描述"}
          </Text>
        </Space>
      )
    },
    {
      title: "模块",
      dataIndex: "module_key",
      key: "module_key",
      render: (value: string) => <Tag>{resolveModuleLabel(value)}</Tag>
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: string) => {
        const meta = resolveStatusMeta(value);
        return <Tag color={meta.color}>{meta.label}</Tag>;
      }
    },
    {
      title: "指派",
      dataIndex: "assigned_to",
      key: "assigned_to",
      render: (value: number) => renderAssignee(value)
    },
    {
      title: "协作",
      dataIndex: "allow_assist",
      key: "allow_assist",
      render: (value: number) => (value === 0 ? <Tag>关闭</Tag> : <Tag color="green">允许</Tag>)
    },
    {
      title: "优先级",
      dataIndex: "priority",
      key: "priority",
      render: (value: number) => <Text>{value ?? 0}</Text>
    },
    {
      title: "更新",
      dataIndex: "updated_at",
      key: "updated_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "操作",
      key: "actions",
      render: (_: string, record: TaskItem) => (
        <Space>
          {isAdmin ? (
            <Button size="small" icon={<EditOutlined />} onClick={() => openEditor(record)}>
              编辑
            </Button>
          ) : null}
          {record.allow_assist !== 0 ? (
            <Button size="small" onClick={() => openAssist(record)}>
              协助
            </Button>
          ) : null}
          {record.status !== "completed" ? (
            <Button
              size="small"
              type="primary"
              loading={completingTaskId === record.id}
              onClick={() => completeTask(record)}
            >
              完成上传
            </Button>
          ) : null}
          <Button size="small" icon={<EyeOutlined />} onClick={() => openActions(record)}>
            历史
          </Button>
        </Space>
      )
    }
  ];

  const selectedVersion = versions.find((item) => item.id === selectedVersionId);

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Title level={3} style={{ margin: 0 }}>
        任务看板
      </Title>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <Text type="secondary">选择景区版本后即可查看任务分配、协作记录与状态。</Text>
          <Space wrap size={12}>
            <Select
              style={{ minWidth: 280 }}
              placeholder="选择景区版本"
              loading={versionLoading}
              options={versionOptions}
              value={selectedVersionId ?? undefined}
              onChange={(value) => setSelectedVersionId(value)}
            />
            <Select
              style={{ minWidth: 160 }}
              placeholder="任务模块"
              options={moduleOptions}
              allowClear
              value={moduleFilter ?? undefined}
              onChange={(value) => setModuleFilter(value ?? null)}
            />
            <Select
              style={{ minWidth: 140 }}
              placeholder="任务状态"
              options={statusOptions.map((item) => ({ value: item.value, label: item.label }))}
              allowClear
              value={statusFilter ?? undefined}
              onChange={(value) => setStatusFilter(value ?? null)}
            />
            {isAdmin ? (
              <Select
                style={{ minWidth: 180 }}
                placeholder="指派给"
                loading={userLoading}
                allowClear
                options={userOptions}
                value={assignedFilter ?? undefined}
                onChange={(value) => setAssignedFilter(value ?? null)}
              />
            ) : (
              <Space>
                <Text type="secondary">仅看我的任务</Text>
                <Switch checked={onlyMine} onChange={setOnlyMine} />
              </Space>
            )}
            <Button icon={<ReloadOutlined />} onClick={loadTasks}>
              刷新
            </Button>
            {isAdmin ? (
              <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>
                新建任务
              </Button>
            ) : null}
          </Space>
          {selectedVersion ? (
            <Space size={16}>
              <Tag color="geekblue">{selectedVersion.location_name || "未命名景区"}</Tag>
              <Tag color="orange">{selectedVersion.app_version_name || "未生成版本"}</Tag>
              <Tag>{selectedVersion.ai_modal || "SD"}</Tag>
              <Text type="secondary">草稿状态：{selectedVersion.draft_status || "未设置"}</Text>
            </Space>
          ) : null}
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        {selectedVersionId ? (
          <Table
            rowKey="id"
            columns={columns}
            dataSource={tasks}
            loading={taskLoading}
            pagination={{ pageSize: 8 }}
          />
        ) : (
          <Empty description="请选择景区版本后查看任务" />
        )}
      </Card>

      <TaskEditorModal
        open={editorOpen}
        task={editingTask}
        users={users}
        userLoading={userLoading}
        submitting={editorSubmitting}
        onCancel={() => {
          setEditorOpen(false);
          setEditingTask(null);
        }}
        onSubmit={handleSaveTask}
      />
      <TaskAssistModal
        open={assistOpen}
        submitting={assistSubmitting}
        onCancel={() => setAssistOpen(false)}
        onSubmit={handleAssist}
      />
      <TaskActionsDrawer
        open={actionsOpen}
        loading={actionsLoading}
        actions={actions}
        users={users}
        onClose={() => setActionsOpen(false)}
      />
    </Space>
  );
};

export default TaskBoard;
