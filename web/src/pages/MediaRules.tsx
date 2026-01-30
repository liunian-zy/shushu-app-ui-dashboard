import { Card, Space, Tabs, Typography, message } from "antd";
import { useAuth } from "../contexts/AuthContext";
import MediaRulePanel from "./media/MediaRulePanel";
import MediaTransformPanel from "./media/MediaTransformPanel";
import IdentityTemplatePanel from "./media/IdentityTemplatePanel";
import TtsPresetPanel from "./media/TtsPresetPanel";
import type { Notify, RequestFn, UploadFn } from "./content/utils";

const { Title, Text } = Typography;

const MediaRules = () => {
  const { token, user } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const isAdmin = user?.role === "admin";

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

  const tabs = [
    ...(isAdmin
      ? [
          {
            key: "rules",
            label: "规则配置",
            children: <MediaRulePanel request={request} />
          },
          {
            key: "tool",
            label: "压缩工具",
            children: (
              <MediaTransformPanel request={request} uploadFile={uploadFile} notify={notify} operatorId={user?.id ?? null} />
            )
          },
          {
            key: "templates",
            label: "身份模板",
            children: <IdentityTemplatePanel request={request} uploadFile={uploadFile} />
          }
        ]
      : []),
    {
      key: "tts-presets",
      label: "TTS 预设",
      children: <TtsPresetPanel request={request} />
    }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      {contextHolder}
      <Title level={3} style={{ margin: 0 }}>
        媒体规则
      </Title>
      <Card style={{ borderRadius: 20 }}>
        <Text type="secondary">
          {isAdmin
            ? "配置图片/视频/音频的尺寸、大小与格式规则，并管理压缩方案与身份模板。"
            : "当前仅开放 TTS 预设配置。"}
        </Text>
      </Card>
      <Tabs items={tabs} />
    </Space>
  );
};

export default MediaRules;
