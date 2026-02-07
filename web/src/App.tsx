import { useState } from "react";
import type { ReactElement, ReactNode } from "react";
import { BrowserRouter, NavLink, Navigate, Outlet, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import { Avatar, Button, Dropdown, Form, Input, Layout, Menu, Modal, Result, Space, Spin, Typography, message } from "antd";
import type { MenuProps } from "antd";
import {
  AppstoreOutlined,
  AuditOutlined,
  DeploymentUnitOutlined,
  FileTextOutlined,
  SettingOutlined,
  SyncOutlined,
  TeamOutlined,
  UserOutlined
} from "@ant-design/icons";
import Dashboard from "./pages/Dashboard";
import TaskBoard from "./pages/TaskBoard";
import VersionManager from "./pages/VersionManager";
import ContentEntry from "./pages/ContentEntry";
import MediaRules from "./pages/MediaRules";
import History from "./pages/History";
import Users from "./pages/Users";
import Login from "./pages/Login";
import { useAuth } from "./contexts/AuthContext";

const { Header, Sider, Content } = Layout;
const { Title, Text } = Typography;

const menuItems: Array<{
  key: string;
  icon: ReactNode;
  label: ReactNode;
  adminOnly?: boolean;
}> = [
  {
    key: "/",
    icon: <AppstoreOutlined />,
    label: <NavLink to="/">概览</NavLink>
  },
  {
    key: "/tasks",
    icon: <DeploymentUnitOutlined />,
    label: <NavLink to="/tasks">任务看板</NavLink>
  },
  {
    key: "/versions",
    icon: <FileTextOutlined />,
    label: <NavLink to="/versions">版本配置</NavLink>,
    adminOnly: true
  },
  {
    key: "/entry",
    icon: <SettingOutlined />,
    label: <NavLink to="/entry">内容录入</NavLink>
  },
  {
    key: "/media-rules",
    icon: <AuditOutlined />,
    label: <NavLink to="/media-rules">媒体规则</NavLink>
  },
  {
    key: "/users",
    icon: <TeamOutlined />,
    label: <NavLink to="/users">账号管理</NavLink>,
    adminOnly: true
  },
  {
    key: "/history",
    icon: <SyncOutlined />,
    label: <NavLink to="/history">操作历史</NavLink>
  }
];

const FullScreenLoading = () => {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        flexDirection: "column",
        gap: 12
      }}
    >
      <Spin size="large" />
      <Text type="secondary">正在校验登录状态...</Text>
    </div>
  );
};

type ChangePasswordFormValues = {
  old_password?: string;
  new_password?: string;
  confirm_password?: string;
};

const AppLayout = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, token, logout } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [passwordOpen, setPasswordOpen] = useState(false);
  const [passwordSubmitting, setPasswordSubmitting] = useState(false);
  const [passwordForm] = Form.useForm<ChangePasswordFormValues>();

  const visibleMenuItems = menuItems
    .filter((item) => !item.adminOnly || user?.role === "admin")
    .map(({ adminOnly, ...item }) => item);
  const activeKey = visibleMenuItems.some((item) => item.key === location.pathname) ? location.pathname : "/";
  const displayName = user?.display_name?.trim() || user?.username || "未登录";
  const roleLabel = user?.role === "admin" ? "管理员" : "成员";

  const request = async (path: string, options: RequestInit = {}) => {
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

  const handleChangePassword = async () => {
    try {
      const values = await passwordForm.validateFields();
      setPasswordSubmitting(true);
      await request("/api/users/me/password", {
        method: "POST",
        body: JSON.stringify({
          old_password: values.old_password,
          new_password: values.new_password
        })
      });
      messageApi.success("密码修改成功");
      setPasswordOpen(false);
      passwordForm.resetFields();
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setPasswordSubmitting(false);
    }
  };

  const handleMenuClick: MenuProps["onClick"] = ({ key }) => {
    if (String(key) === "change-password") {
      setPasswordOpen(true);
      return;
    }
    if (String(key) === "logout") {
      logout();
      navigate("/login", { replace: true });
    }
  };

  const userMenuItems: MenuProps["items"] = [
    {
      key: "role",
      label: `角色：${roleLabel}`,
      disabled: true
    },
    {
      key: "change-password",
      label: "修改密码"
    },
    {
      type: "divider"
    },
    {
      key: "logout",
      label: "退出登录"
    }
  ];

  const userMenu: MenuProps = {
    items: userMenuItems,
    onClick: handleMenuClick
  };

  return (
    <Layout style={{ minHeight: "100vh" }}>
      {contextHolder}
      <Sider
        width={240}
        style={{
          position: "fixed",
          left: 0,
          top: 0,
          bottom: 0,
          overflowY: "auto",
          background: "rgba(255,255,255,0.78)",
          borderRight: "1px solid rgba(29,26,23,0.08)"
        }}
      >
        <div style={{ padding: "24px 20px 16px" }}>
          <Title level={4} style={{ margin: 0 }}>
            数枢 UI 方案
          </Title>
          <Text type="secondary">景区配置协作平台</Text>
        </div>
        <Menu mode="inline" items={visibleMenuItems} selectedKeys={[activeKey]} style={{ background: "transparent", borderRight: 0 }} />
      </Sider>
      <Layout style={{ marginLeft: 240, minHeight: "100vh" }}>
        <Header
          style={{
            background: "rgba(255,255,255,0.65)",
            borderBottom: "1px solid rgba(29,26,23,0.08)",
            padding: "0 28px",
            position: "sticky",
            top: 0,
            zIndex: 50
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              height: "100%",
              justifyContent: "space-between"
            }}
          >
            <div style={{ display: "flex", flexDirection: "column", gap: 2, paddingTop: 4, paddingBottom: 4 }}>
              <Title level={5} style={{ margin: 0, lineHeight: "24px" }}>
                数枢-APP景区UI方案系统
              </Title>
              <Text type="secondary" style={{ lineHeight: "16px" }}>
                多角色协作版
              </Text>
            </div>
            <Dropdown menu={userMenu} placement="bottomRight">
              <Space style={{ cursor: "pointer" }}>
                <Avatar icon={<UserOutlined />} style={{ backgroundColor: "#d47f45" }} />
                <div style={{ display: "flex", flexDirection: "column" }}>
                  <Text>{displayName}</Text>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {roleLabel}
                  </Text>
                </div>
              </Space>
            </Dropdown>
          </div>
        </Header>
        <Content style={{ padding: "28px" }}>
          <Outlet />
        </Content>
      </Layout>

      <Modal
        title="修改密码"
        open={passwordOpen}
        onCancel={() => {
          setPasswordOpen(false);
          passwordForm.resetFields();
        }}
        onOk={handleChangePassword}
        okButtonProps={{ loading: passwordSubmitting }}
        destroyOnClose
      >
        <Form form={passwordForm} layout="vertical" preserve={false}>
          <Form.Item label="旧密码" name="old_password" rules={[{ required: true, message: "请输入旧密码" }]}> 
            <Input.Password placeholder="请输入当前登录密码" />
          </Form.Item>
          <Form.Item label="新密码" name="new_password" rules={[{ required: true, message: "请输入新密码" }]}> 
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            label="确认新密码"
            name="confirm_password"
            dependencies={["new_password"]}
            rules={[
              { required: true, message: "请再次输入新密码" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("new_password") === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error("两次输入的新密码不一致"));
                }
              })
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
};

const ProtectedLayout = () => {
  const location = useLocation();
  const { user, loading } = useAuth();

  if (loading) {
    return <FullScreenLoading />;
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }
  return <AppLayout />;
};

const RequireAdmin = ({ children }: { children: ReactElement }) => {
  const navigate = useNavigate();
  const { user } = useAuth();

  if (user?.role !== "admin") {
    return (
      <div style={{ padding: "32px" }}>
        <Result
          status="403"
          title="暂无权限"
          subTitle="该功能仅管理员可用。"
          extra={
            <Button type="primary" onClick={() => navigate("/")}>
              返回概览
            </Button>
          }
        />
      </div>
    );
  }
  return children;
};

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route element={<ProtectedLayout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/tasks" element={<TaskBoard />} />
          <Route
            path="/versions"
            element={
              <RequireAdmin>
                <VersionManager />
              </RequireAdmin>
            }
          />
          <Route path="/entry" element={<ContentEntry />} />
          <Route path="/media-rules" element={<MediaRules />} />
          <Route
            path="/users"
            element={
              <RequireAdmin>
                <Users />
              </RequireAdmin>
            }
          />
          <Route path="/history" element={<History />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
