import { useEffect, useState } from "react";
import { Button, Card, Col, Form, Input, Modal, Row, Select, Space, Table, Tag, Typography, message } from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useAuth } from "../contexts/AuthContext";
import { formatDate } from "./content/constants";

const { Title, Text } = Typography;

type UserItem = {
  id: number;
  username?: string | null;
  display_name?: string | null;
  role?: string | null;
  status?: number | null;
  created_at?: string | null;
  last_login_at?: string | null;
};

type UserFormValues = {
  username?: string;
  display_name?: string;
  role?: string;
  password?: string;
};

const roleOptions = [
  { value: "admin", label: "管理员" },
  { value: "user", label: "成员" }
];

const Users = () => {
  const { token } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [items, setItems] = useState<UserItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorSubmitting, setEditorSubmitting] = useState(false);
  const [form] = Form.useForm<UserFormValues>();

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

  const loadUsers = async () => {
    setLoading(true);
    try {
      const res = await request<{ data: UserItem[] }>("/api/users");
      setItems(res.data || []);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取账号失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadUsers();
  }, []);

  const openEditor = () => {
    setEditorOpen(true);
    form.resetFields();
    form.setFieldsValue({ role: "user" });
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setEditorSubmitting(true);
      await request("/api/users", {
        method: "POST",
        body: JSON.stringify({
          username: values.username?.trim(),
          display_name: values.display_name?.trim(),
          role: values.role,
          password: values.password
        })
      });
      messageApi.success("账号已创建");
      setEditorOpen(false);
      form.resetFields();
      void loadUsers();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setEditorSubmitting(false);
    }
  };

  const columns = [
    { title: "用户名", dataIndex: "username", key: "username", render: (value: string) => <Text>{value || "-"}</Text> },
    { title: "显示名", dataIndex: "display_name", key: "display_name", render: (value: string) => <Text>{value || "-"}</Text> },
    {
      title: "角色",
      dataIndex: "role",
      key: "role",
      render: (value: string) => (value === "admin" ? <Tag color="gold">管理员</Tag> : <Tag>成员</Tag>)
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (value: number) => (value === 0 ? <Tag>停用</Tag> : <Tag color="green">启用</Tag>)
    },
    {
      title: "创建时间",
      dataIndex: "created_at",
      key: "created_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    },
    {
      title: "最近登录",
      dataIndex: "last_login_at",
      key: "last_login_at",
      render: (value: string) => <Text type="secondary">{formatDate(value)}</Text>
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Title level={3} style={{ margin: 0 }}>
        账号管理
      </Title>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">创建与管理管理员/成员账号。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadUsers} loading={loading}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openEditor}>
              新建账号
            </Button>
          </Space>
        </Space>
      </Card>
      <Card style={{ borderRadius: 20 }}>
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ pageSize: 8 }} />
      </Card>

      <Modal
        title="新建账号"
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false);
          form.resetFields();
        }}
        onOk={handleSubmit}
        width={820}
        okButtonProps={{ loading: editorSubmitting }}
        styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item label="用户名" name="username" rules={[{ required: true, message: "请输入用户名" }]}>
                <Input placeholder="如：zhangsan" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="显示名" name="display_name">
                <Input placeholder="可选：用于展示的名称" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item label="角色" name="role" rules={[{ required: true, message: "请选择角色" }]}>
                <Select options={roleOptions} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="密码" name="password" rules={[{ required: true, message: "请输入密码" }]}>
                <Input.Password placeholder="至少 6 位" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </Space>
  );
};

export default Users;
