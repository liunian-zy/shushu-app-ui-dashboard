import { Card, Drawer, Empty, Space, Typography } from "antd";
import { formatDate, resolveActionLabel, TaskAction, UserItem } from "./constants";

const { Text } = Typography;

type TaskActionsDrawerProps = {
  open: boolean;
  loading: boolean;
  actions: TaskAction[];
  users: UserItem[];
  onClose: () => void;
};

const TaskActionsDrawer = ({ open, loading, actions, users, onClose }: TaskActionsDrawerProps) => {
  const userMap = new Map<number, UserItem>();
  users.forEach((item) => userMap.set(item.id, item));

  return (
    <Drawer title="任务操作历史" open={open} onClose={onClose} width={420}>
      {loading ? (
        <Text type="secondary">正在加载...</Text>
      ) : actions.length ? (
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          {actions.map((item) => {
            const actor = item.actor_id ? userMap.get(item.actor_id) : null;
            const actorLabel =
              item.actor_name ||
              item.actor_username ||
              actor?.display_name ||
              actor?.username ||
              (item.actor_id ? `用户${item.actor_id}` : "-");
            return (
              <Card key={item.id} size="small">
                <Space direction="vertical" size={6}>
                  <Text strong>{resolveActionLabel(item.action)}</Text>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {formatDate(item.created_at)}
                  </Text>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    操作人：
                    {actorLabel}
                  </Text>
                  {item.detail ? (
                    <pre style={{ margin: 0, whiteSpace: "pre-wrap", fontSize: 12 }}>
                      {JSON.stringify(item.detail, null, 2)}
                    </pre>
                  ) : null}
                </Space>
              </Card>
            );
          })}
        </Space>
      ) : (
        <Empty description="暂无操作记录" />
      )}
    </Drawer>
  );
};

export default TaskActionsDrawer;
