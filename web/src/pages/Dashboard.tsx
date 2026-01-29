import { Card, Col, Row, Select, Space, Tag, Typography, Button, message } from "antd";
import { useNavigate } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { DraftVersion, formatDate } from "./content/constants";

const { Title, Text } = Typography;

type DashboardSummary = {
  tasks: {
    total: number;
    pending: number;
    assist_active: number;
    status_counts: Record<string, number>;
  };
  media: {
    today_checked: number;
    today_compliant: number;
    today_pending: number;
  };
  sync: {
    pending_versions: number;
    last_sync_at?: string | null;
  };
};

const Dashboard = () => {
  const navigate = useNavigate();
  const { token } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [versions, setVersions] = useState<DraftVersion[]>([]);
  const [versionLoading, setVersionLoading] = useState(false);
  const [selectedVersionId, setSelectedVersionId] = useState<number | null>(null);
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(false);

  const request = async <T,>(path: string, options: RequestInit = {}) => {
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

  const loadSummary = async (draftVersionId: number) => {
    setSummaryLoading(true);
    try {
      const res = await request<DashboardSummary>(`/api/dashboard/summary?draft_version_id=${draftVersionId}`);
      setSummary(res);
    } catch (error) {
      setSummary(null);
      messageApi.error(error instanceof Error ? error.message : "获取概览失败");
    } finally {
      setSummaryLoading(false);
    }
  };

  useEffect(() => {
    if (!token) {
      setVersions([]);
      return;
    }
    void loadVersions();
  }, [token]);

  useEffect(() => {
    if (!token) {
      setSummary(null);
      return;
    }
    if (!selectedVersionId) {
      setSummary(null);
      return;
    }
    void loadSummary(selectedVersionId);
  }, [selectedVersionId, token]);

  const versionOptions = useMemo(
    () =>
      versions.map((item) => ({
        value: item.id,
        label: `${item.location_name || "未命名景区"} / ${item.app_version_name || "未生成版本"}`
      })),
    [versions]
  );

  const taskPending = summary?.tasks.pending ?? 0;
  const assistActive = summary?.tasks.assist_active ?? 0;
  const todayChecked = summary?.media.today_checked ?? 0;
  const todayCompliant = summary?.media.today_compliant ?? 0;
  const todayPending = summary?.media.today_pending ?? 0;
  const pendingVersions = summary?.sync.pending_versions ?? 0;
  const lastSyncAt = summary?.sync.last_sync_at ?? null;

  return (
    <Space direction="vertical" size={24} style={{ width: "100%" }}>
      {contextHolder}
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={8}>
          <Title level={3} style={{ margin: 0 }}>
            今日进度概览
          </Title>
          <Text type="secondary">
            选择一个景区版本，快速查看任务分配、媒体合规与同步状态。
          </Text>
          <Select
            style={{ minWidth: 320 }}
            placeholder="选择景区版本"
            options={versionOptions}
            value={selectedVersionId ?? undefined}
            onChange={(value) => setSelectedVersionId(value)}
            loading={versionLoading}
          />
        </Space>
      </Card>
      <Row gutter={[24, 24]}>
        <Col xs={24} md={12} xl={8}>
          <Card title="任务推进" style={{ borderRadius: 20 }}>
            <Space direction="vertical">
              <Text>
                待处理任务 {summaryLoading ? "..." : taskPending} 项
              </Text>
              <Text>协作中 {summaryLoading ? "..." : assistActive} 项</Text>
              <Button type="primary" onClick={() => navigate("/tasks")}>
                进入任务看板
              </Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} md={12} xl={8}>
          <Card title="媒体合规" style={{ borderRadius: 20 }}>
            <Space direction="vertical">
              <Text>今日检测 {summaryLoading ? "..." : todayChecked} 个文件</Text>
              <Space>
                <Tag color="green">合规 {summaryLoading ? "..." : todayCompliant}</Tag>
                <Tag color="volcano">待处理 {summaryLoading ? "..." : todayPending}</Tag>
              </Space>
              <Button onClick={() => navigate("/media-rules")}>查看处理中心</Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} md={12} xl={8}>
          <Card title="同步状态" style={{ borderRadius: 20 }}>
            <Space direction="vertical">
              <Text>待同步景区 {summaryLoading ? "..." : pendingVersions} 个</Text>
              <Text type="secondary">上次同步: {formatDate(lastSyncAt)}</Text>
              <Button onClick={() => navigate("/history")}>打开同步中心</Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </Space>
  );
};

export default Dashboard;
