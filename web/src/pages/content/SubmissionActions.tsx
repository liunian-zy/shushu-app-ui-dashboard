import { useState } from "react";
import { Button, Drawer, Empty, Modal, Space, Table, Tag, Typography } from "antd";
import type { Notify, RequestFn } from "./utils";
import { sanitizeSubmissionPayload } from "./utils";
import { formatDate } from "./constants";

const { Text } = Typography;

type DiffItem = {
  field: string;
  old: unknown;
  new: unknown;
};

type SubmissionItem = {
  id: number;
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
  diff?: DiffItem[];
};

type SubmissionActionsProps = {
  draftVersionId: number;
  moduleKey: string;
  entityTable: string;
  entityId: number;
  operatorId?: number | null;
  request: RequestFn;
  notify: Notify;
  getPayload: () => Record<string, unknown>;
  disabled?: boolean;
};

const SubmissionActions = ({
  draftVersionId,
  moduleKey,
  entityTable,
  entityId,
  operatorId,
  request,
  notify,
  getPayload,
  disabled
}: SubmissionActionsProps) => {
  const [submitting, setSubmitting] = useState(false);
  const [diffOpen, setDiffOpen] = useState(false);
  const [diffItems, setDiffItems] = useState<DiffItem[]>([]);
  const [needConfirm, setNeedConfirm] = useState(false);
  const [currentSubmissionId, setCurrentSubmissionId] = useState<number | null>(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyItems, setHistoryItems] = useState<SubmissionItem[]>([]);

  const handleSubmit = async () => {
    if (!operatorId) {
      notify.warning("缺少提交人信息");
      return;
    }
    setSubmitting(true);
    try {
      const payload = sanitizeSubmissionPayload(getPayload());
      const res = await request<{ submission_id: number; need_confirm: boolean; diff: DiffItem[] }>("/api/draft/submit", {
        method: "POST",
        body: JSON.stringify({
          draft_version_id: draftVersionId,
          module_key: moduleKey,
          entity_table: entityTable,
          entity_id: entityId,
          submit_by: operatorId,
          payload
        })
      });
      setCurrentSubmissionId(res.submission_id);
      setNeedConfirm(res.need_confirm);
      setDiffItems(res.diff || []);
      setDiffOpen(true);
      notify.success(res.need_confirm ? "提交完成，等待确认" : "提交完成");
    } catch (error) {
      if (error instanceof Error) {
        notify.error(error.message);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleConfirm = async (submissionId: number) => {
    if (!operatorId) {
      notify.warning("缺少确认人信息");
      return;
    }
    try {
      await request("/api/draft/confirm", {
        method: "POST",
        body: JSON.stringify({
          submission_id: submissionId,
          confirmed_by: operatorId
        })
      });
      notify.success("已确认");
      if (historyOpen) {
        void loadHistory();
      }
    } catch (error) {
      if (error instanceof Error) {
        notify.error(error.message);
      }
    }
  };

  const loadHistory = async () => {
    setHistoryLoading(true);
    try {
      const params = new URLSearchParams({
        draft_version_id: String(draftVersionId),
        module_key: moduleKey,
        entity_table: entityTable,
        entity_id: String(entityId)
      });
      const res = await request<{ data: SubmissionItem[] }>(`/api/draft/submissions?${params.toString()}`);
      setHistoryItems(res.data || []);
    } catch (error) {
      if (error instanceof Error) {
        notify.error(error.message);
      }
    } finally {
      setHistoryLoading(false);
    }
  };

  const openHistory = () => {
    setHistoryOpen(true);
    void loadHistory();
  };

  const diffColumns = [
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
  ];

  const historyColumns = [
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
      title: "操作",
      key: "actions",
      render: (_: string, record: SubmissionItem) => (
        <Space>
          <Button
            size="small"
            onClick={() => {
              setDiffItems(record.diff || []);
              setNeedConfirm(record.need_confirm);
              setCurrentSubmissionId(record.id);
              setDiffOpen(true);
            }}
          >
            查看差异
          </Button>
          {record.status === "pending_confirm" ? (
            <Button size="small" type="primary" onClick={() => handleConfirm(record.id)}>
              确认
            </Button>
          ) : null}
        </Space>
      )
    }
  ];

  return (
    <Space size={6} wrap>
      <Button size="small" onClick={handleSubmit} loading={submitting} disabled={disabled}>
        提交
      </Button>
      <Button size="small" onClick={openHistory} disabled={disabled}>
        历史
      </Button>

      <Modal
        title="提交差异"
        open={diffOpen}
        onCancel={() => setDiffOpen(false)}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        footer={[
          <Button key="close" onClick={() => setDiffOpen(false)}>
            关闭
          </Button>,
          needConfirm && currentSubmissionId ? (
            <Button key="confirm" type="primary" onClick={() => handleConfirm(currentSubmissionId)}>
              确认提交
            </Button>
          ) : null
        ]}
      >
        {diffItems.length ? (
          <Table rowKey={(record, index) => `${record.field}-${index}`} columns={diffColumns} dataSource={diffItems} pagination={false} />
        ) : (
          <Empty description="暂无差异" />
        )}
      </Modal>

      <Drawer
        title="提交历史"
        open={historyOpen}
        onClose={() => setHistoryOpen(false)}
        width={720}
      >
        {historyItems.length ? (
          <Table
            rowKey="id"
            columns={historyColumns}
            dataSource={historyItems}
            loading={historyLoading}
            pagination={{ pageSize: 8 }}
          />
        ) : (
          <Empty description="暂无提交记录" />
        )}
      </Drawer>
    </Space>
  );
};

export default SubmissionActions;
