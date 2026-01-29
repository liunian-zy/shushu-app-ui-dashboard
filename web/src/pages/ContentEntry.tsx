import { useEffect, useMemo, useState } from "react";
import { Button, Card, Modal, Select, Space, Table, Tabs, Tag, Typography, message } from "antd";
import { useAuth } from "../contexts/AuthContext";
import AppUIFieldsPanel from "./content/AppUIFieldsPanel";
import BannerPanel from "./content/BannerPanel";
import ClothesPanel from "./content/ClothesPanel";
import ConfigExtraPanel from "./content/ConfigExtraPanel";
import IdentityPanel from "./content/IdentityPanel";
import PhotoPanel from "./content/PhotoPanel";
import ScenePanel from "./content/ScenePanel";
import { formatDate } from "./content/constants";
import type { DraftVersion } from "./content/constants";
import type { Notify, RequestFn, TTSPreset, TTSFn, UploadFn } from "./content/utils";

const { Title, Text } = Typography;

type SyncJob = {
  id: number;
  module_key?: string | null;
  status?: string | null;
  error_message?: string | null;
  started_at?: string | null;
  finished_at?: string | null;
  created_at?: string | null;
  trigger_by?: number | null;
  trigger_name?: string | null;
  trigger_username?: string | null;
};

const ContentEntry = () => {
  const { token, user } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [versions, setVersions] = useState<DraftVersion[]>([]);
  const [versionLoading, setVersionLoading] = useState(false);
  const [selectedVersionId, setSelectedVersionId] = useState<number | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [syncModules, setSyncModules] = useState<string[]>([]);
  const [syncJobs, setSyncJobs] = useState<SyncJob[]>([]);
  const [syncJobsLoading, setSyncJobsLoading] = useState(false);
  const [syncHistoryOpen, setSyncHistoryOpen] = useState(false);
  const [ttsPresets, setTtsPresets] = useState<TTSPreset[]>([]);

  const notify: Notify = {
    success: (msg) => messageApi.success(msg),
    error: (msg) => messageApi.error(msg),
    warning: (msg) => messageApi.warning(msg)
  };

  const request: RequestFn = async (path, options = {}) => {
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

  const uploadFile: UploadFn = async (file, moduleKey, draftVersionId) => {
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("module_key", moduleKey);
      formData.append("draft_version_id", String(draftVersionId));

      const response = await fetch("/api/local-files/upload", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token ?? ""}`
        },
        body: formData
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error((data as { error?: string }).error || "上传失败");
      }

      return { path: data.path as string, url: data.url as string };
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "上传失败");
      throw error;
    }
  };

  const generateTTS: TTSFn = async (text, moduleKey, draftVersionId, options) => {
    const res = await request<{ audio_path: string; audio_url?: string }>("/api/tts/convert", {
      method: "POST",
      body: JSON.stringify({
        text,
        module_key: moduleKey,
        draft_version_id: draftVersionId,
        preset_id: options?.preset_id,
        voice_id: options?.voice_id,
        volume: options?.volume,
        speed: options?.speed,
        pitch: options?.pitch,
        stability: options?.stability,
        similarity: options?.similarity,
        exaggeration: options?.exaggeration
      })
    });
    return res;
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
      notify.error(error instanceof Error ? error.message : "获取版本失败");
    } finally {
      setVersionLoading(false);
    }
  };

  const loadTtsPresets = async () => {
    try {
      const res = await request<{ data: TTSPreset[] }>("/api/tts/presets");
      setTtsPresets(res.data || []);
    } catch {
      setTtsPresets([]);
    }
  };

  const loadSyncJobs = async (versionId?: number | null) => {
    const targetId = versionId ?? selectedVersionId;
    if (!targetId) {
      setSyncJobs([]);
      return;
    }
    setSyncJobsLoading(true);
    try {
      const res = await request<{ data: SyncJob[] }>(`/api/sync/jobs?draft_version_id=${targetId}`);
      setSyncJobs(res.data || []);
    } catch (error) {
      setSyncJobs([]);
      notify.error(error instanceof Error ? error.message : "获取同步记录失败");
    } finally {
      setSyncJobsLoading(false);
    }
  };

  useEffect(() => {
    void loadVersions();
  }, []);

  useEffect(() => {
    if (!token) {
      setTtsPresets([]);
      return;
    }
    void loadTtsPresets();
  }, [token]);

  useEffect(() => {
    void loadSyncJobs(selectedVersionId);
  }, [selectedVersionId]);

  const selectedVersion = useMemo(
    () => versions.find((item) => item.id === selectedVersionId) ?? null,
    [versions, selectedVersionId]
  );

  const versionOptions = useMemo(
    () =>
      versions.map((item) => ({
        value: item.id,
        label: `${item.location_name || "未命名景区"} / ${item.app_version_name || "未生成版本"}`
      })),
    [versions]
  );

  const operatorId = user?.id ?? null;

  const syncModuleOptions = [
    { value: "version_names", label: "版本配置" },
    { value: "banners", label: "轮播图" },
    { value: "identities", label: "身份信息" },
    { value: "scenes", label: "场景信息" },
    { value: "app_ui_fields", label: "页面配置" },
    { value: "config_extra_steps", label: "额外配置" },
    { value: "clothes_categories", label: "服饰偏好" },
    { value: "photo_hobbies", label: "拍摄偏好" }
  ];

  const syncStatusMap: Record<string, { label: string; color: string }> = {
    running: { label: "同步中", color: "blue" },
    success: { label: "已同步", color: "green" },
    failed: { label: "失败", color: "red" },
    pending_confirm: { label: "待确认", color: "gold" }
  };

  const latestSyncByModule = useMemo(() => {
    const latest = new Map<string, SyncJob>();
    syncJobs.forEach((job) => {
      if (!job.module_key) {
        return;
      }
      if (!latest.has(job.module_key)) {
        latest.set(job.module_key, job);
      }
    });
    return syncModuleOptions.map((option) => ({
      module: option.value,
      label: option.label,
      job: latest.get(option.value)
    }));
  }, [syncJobs, syncModuleOptions]);

  const handleSync = async (confirm = false) => {
    if (!selectedVersion) {
      notify.warning("请先选择景区版本");
      return;
    }
    if (!operatorId) {
      notify.warning("缺少提交人信息");
      return;
    }
    if (!token) {
      notify.warning("缺少登录凭证");
      return;
    }

    setSyncing(true);
    try {
      const response = await fetch("/api/sync", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          draft_version_id: selectedVersion.id,
          trigger_by: operatorId,
          confirm,
          modules: syncModules
        })
      });
      const data = await response.json().catch(() => ({}));
      if (response.status === 409 && (data as { need_confirm?: boolean }).need_confirm) {
        setSyncing(false);
        Modal.confirm({
          title: "检测到线上已有版本",
          content: "继续同步将覆盖线上同名景区的数据。",
          okText: "继续同步",
          cancelText: "取消",
          onOk: () => handleSync(true)
        });
        return;
      }
      if (!response.ok) {
        throw new Error((data as { error?: string }).error || "同步失败");
      }
      notify.success("同步完成");
      void loadSyncJobs(selectedVersion.id);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "同步失败");
    } finally {
      setSyncing(false);
    }
  };

  const tabs = [
    {
      key: "banners",
      label: "轮播图",
      children: (
        <BannerPanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          notify={notify}
          operatorId={operatorId}
        />
      )
    },
    {
      key: "identities",
      label: "身份信息",
      children: (
        <IdentityPanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          notify={notify}
          operatorId={operatorId}
        />
      )
    },
    {
      key: "scenes",
      label: "场景信息",
      children: (
        <ScenePanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          generateTTS={generateTTS}
          notify={notify}
          operatorId={operatorId}
          ttsPresets={ttsPresets}
        />
      )
    },
    {
      key: "app-ui",
      label: "页面配置",
      children: (
        <AppUIFieldsPanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          generateTTS={generateTTS}
          notify={notify}
          operatorId={operatorId}
          ttsPresets={ttsPresets}
        />
      )
    },
    {
      key: "extra",
      label: "额外配置",
      children: (
        <ConfigExtraPanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          generateTTS={generateTTS}
          notify={notify}
          operatorId={operatorId}
          ttsPresets={ttsPresets}
        />
      )
    },
    {
      key: "clothes",
      label: "服饰偏好",
      children: (
        <ClothesPanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          generateTTS={generateTTS}
          notify={notify}
          operatorId={operatorId}
          ttsPresets={ttsPresets}
        />
      )
    },
    {
      key: "photo",
      label: "拍摄偏好",
      children: (
        <PhotoPanel
          version={selectedVersion}
          request={request}
          uploadFile={uploadFile}
          generateTTS={generateTTS}
          notify={notify}
          operatorId={operatorId}
          ttsPresets={ttsPresets}
        />
      )
    }
  ];

  const renderSyncStatus = (status?: string | null) => {
    if (!status) {
      return <Tag>未同步</Tag>;
    }
    const config = syncStatusMap[status];
    if (!config) {
      return <Tag>{status}</Tag>;
    }
    return <Tag color={config.color}>{config.label}</Tag>;
  };

  const syncSummaryColumns = [
    {
      title: "模块",
      dataIndex: "label",
      key: "label"
    },
    {
      title: "状态",
      key: "status",
      render: (_: string, record: { job?: SyncJob }) => renderSyncStatus(record.job?.status ?? null)
    },
    {
      title: "上次同步",
      key: "time",
      render: (_: string, record: { job?: SyncJob }) =>
        <Text type="secondary">{formatDate(record.job?.finished_at ?? record.job?.started_at ?? record.job?.created_at ?? null)}</Text>
    },
    {
      title: "触发人",
      key: "trigger",
      render: (_: string, record: { job?: SyncJob }) =>
        <Text>{record.job?.trigger_name || record.job?.trigger_username || record.job?.trigger_by || "-"}</Text>
    },
    {
      title: "备注",
      key: "note",
      render: (_: string, record: { job?: SyncJob }) =>
        <Text type="secondary">{record.job?.error_message || "-"}</Text>
    }
  ];

  const syncHistoryColumns = [
    {
      title: "模块",
      dataIndex: "module_key",
      key: "module_key",
      render: (value: string) => {
        const option = syncModuleOptions.find((item) => item.value === value);
        return <Text>{option?.label || value || "-"}</Text>;
      }
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: string) => renderSyncStatus(value)
    },
    {
      title: "开始时间",
      dataIndex: "started_at",
      key: "started_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "完成时间",
      dataIndex: "finished_at",
      key: "finished_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "触发人",
      key: "trigger",
      render: (_: string, record: SyncJob) =>
        <Text>{record.trigger_name || record.trigger_username || record.trigger_by || "-"}</Text>
    },
    {
      title: "错误",
      dataIndex: "error_message",
      key: "error_message",
      render: (value: string) => <Text type="secondary">{value || "-"}</Text>
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Title level={3} style={{ margin: 0 }}>
        内容录入
      </Title>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">轮播图、身份、场景、配置项与视频等模块集中录入与管理。</Text>
          <Select
            style={{ minWidth: 320 }}
            placeholder="选择景区版本"
            loading={versionLoading}
            options={versionOptions}
            value={selectedVersionId ?? undefined}
            onChange={(value) => setSelectedVersionId(value)}
          />
          {selectedVersion ? (
            <Space size={12}>
              <Text>当前景区：{selectedVersion.location_name || "-"}</Text>
              <Text type="secondary">版本名：{selectedVersion.app_version_name || "-"}</Text>
            </Space>
          ) : null}
          <Space wrap>
            <Select
              mode="multiple"
              style={{ minWidth: 320 }}
              placeholder="选择要同步的模块（留空则全量同步）"
              options={syncModuleOptions}
              value={syncModules.length ? syncModules : undefined}
              onChange={(values) => setSyncModules(values)}
              allowClear
            />
            <Button type="primary" loading={syncing} onClick={() => handleSync(false)} disabled={!selectedVersion}>
              同步到线上
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Space wrap>
            <Text type="secondary">模块同步状态</Text>
            <Button onClick={() => loadSyncJobs()} loading={syncJobsLoading} disabled={!selectedVersionId}>
              刷新同步记录
            </Button>
            <Button onClick={() => setSyncHistoryOpen(true)} disabled={!syncJobs.length}>
              查看同步记录
            </Button>
          </Space>
          <Table
            rowKey="module"
            columns={syncSummaryColumns}
            dataSource={latestSyncByModule}
            pagination={false}
            size="small"
          />
        </Space>
      </Card>
      <Tabs items={tabs} />
      <Modal
        title="同步记录"
        open={syncHistoryOpen}
        onCancel={() => setSyncHistoryOpen(false)}
        footer={null}
        width={900}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        destroyOnClose
      >
        <Table
          rowKey="id"
          columns={syncHistoryColumns}
          dataSource={syncJobs}
          loading={syncJobsLoading}
          pagination={{ pageSize: 8 }}
          size="small"
        />
      </Modal>
    </Space>
  );
};

export default ContentEntry;
