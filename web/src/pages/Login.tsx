import { useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { Alert, Button, Card, Form, Input, Space, Typography } from "antd";
import { useAuth } from "../contexts/AuthContext";
import "./Login.css";

const { Title, Text } = Typography;

const Login = () => {
  const { user, login, loading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm();
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState<boolean>(false);

  if (user) {
    return <Navigate to="/" replace />;
  }

  const targetPath =
    (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/";

  const handleFinish = async (values: { username: string; password: string }) => {
    setError(null);
    setSubmitting(true);
    try {
      await login(values.username.trim(), values.password);
      navigate(targetPath, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败，请稍后重试");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="login-shell">
      <div className="login-hero">
        <div>
          <Text className="login-kicker">Shushu UI Plan System</Text>
          <Title>数枢-APP景区UI方案系统</Title>
          <Text type="secondary">
            面向景区内容收集的协作工作台，统一草稿、媒体规范与同步流程。
          </Text>
          <div className="login-feature-list">
            <div className="login-feature">
              <Text>任务分工清晰化，支持协助与操作追溯。</Text>
            </div>
            <div className="login-feature">
              <Text>媒体规则集中配置，上传即校验并生成合规版本。</Text>
            </div>
            <div className="login-feature">
              <Text>提交历史保留差异，确保最终同步可信。</Text>
            </div>
          </div>
        </div>
        <Space direction="vertical" size={4}>
          <Text type="secondary">建议使用桌面端浏览器进行录入与预览。</Text>
          <Text type="secondary">如需开通账号，请联系管理员创建。</Text>
        </Space>
      </div>

      <Card className="login-card" bordered={false}>
        <Space direction="vertical" size={20} style={{ width: "100%" }}>
          <Space direction="vertical" size={6}>
            <Title level={4} style={{ margin: 0 }}>
              进入工作台
            </Title>
            <Text type="secondary">使用账号登录以继续任务协作。</Text>
          </Space>

          {error ? <Alert message={error} type="error" showIcon /> : null}

          <Form form={form} layout="vertical" onFinish={handleFinish} requiredMark={false}>
            <Form.Item
              label="账号"
              name="username"
              rules={[{ required: true, message: "请输入账号" }]}
            >
              <Input placeholder="输入用户名" autoComplete="username" />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: "请输入密码" }]}
            >
              <Input.Password placeholder="输入密码" autoComplete="current-password" />
            </Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={submitting || loading}
              block
              style={{ height: 44 }}
            >
              登录
            </Button>
          </Form>

          <Space direction="vertical" size={4}>
            <Text type="secondary">默认管理员账号：admin / 123456</Text>
            <Text type="secondary">初次登录后请及时修改密码。</Text>
          </Space>
        </Space>
      </Card>
    </div>
  );
};

export default Login;
