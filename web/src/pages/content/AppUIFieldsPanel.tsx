import { useEffect, useState } from "react";
import { Button, Card, Empty, Form, Input, Space, Switch, Typography } from "antd";
import { ReloadOutlined, SaveOutlined } from "@ant-design/icons";
import UploadField from "./UploadField";
import SubmissionActions from "./SubmissionActions";
import TTSInlinePanel from "./TTSInlinePanel";
import { AppUIFields, DraftVersion, formatDate, submitStatusLabels } from "./constants";
import type { Notify, RequestFn, TTSPreset, TTSFn, TTSResult, UploadFn } from "./utils";
import { buildLocalDraftKey, generateTTSBatch, loadLocalDraft, saveLocalDraft } from "./utils";

const { Title, Text } = Typography;

type AppUIFieldsPanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  generateTTS: TTSFn;
  notify: Notify;
  operatorId?: number | null;
  ttsPresets?: TTSPreset[];
  refreshTtsPresets?: () => void;
};

type AppUIFormValues = {
  home_title_left?: string;
  home_title_right?: string;
  home_subtitle?: string;
  start_experience?: string;
  step1_music_text?: string;
  step1_title?: string;
  step1_music?: string;
  step2_music_text?: string;
  step2_title?: string;
  step2_music?: string;
  status?: boolean;
  print_wait?: string;
};

const AppUIFieldsPanel = ({ version, request, uploadFile, generateTTS, notify, operatorId, ttsPresets, refreshTtsPresets }: AppUIFieldsPanelProps) => {
  const [form] = Form.useForm<AppUIFormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [step1Preview, setStep1Preview] = useState<string | null>(null);
  const [step2Preview, setStep2Preview] = useState<string | null>(null);
  const [printPreview, setPrintPreview] = useState<string | null>(null);
  const [hasRecord, setHasRecord] = useState(false);
  const [recordId, setRecordId] = useState<number | null>(null);
  const [submitStatus, setSubmitStatus] = useState<string | null>(null);
  const [lastSubmitAt, setLastSubmitAt] = useState<string | null>(null);
  const step1Value = Form.useWatch("step1_music", form);
  const step2Value = Form.useWatch("step2_music", form);
  const printValue = Form.useWatch("print_wait", form);
  const step1TextValue = Form.useWatch("step1_music_text", form);
  const step2TextValue = Form.useWatch("step2_music_text", form);
  const draftKey = version?.id ? buildLocalDraftKey("app_ui_fields", version.id, "single") : "";

  const loadData = async () => {
    if (!version?.id) {
      setHasRecord(false);
      setRecordId(null);
      setSubmitStatus(null);
      setLastSubmitAt(null);
      form.resetFields();
      return;
    }
    setLoading(true);
    try {
      const res = await request<{ data: AppUIFields | null }>(
        `/api/draft/app-ui-fields?draft_version_id=${version.id}&app_version_name_id=${version.id}`
      );
      const data = res.data ?? null;
      setHasRecord(Boolean(data?.id));
      setRecordId(data?.id ?? null);
      setSubmitStatus(data?.submit_status ?? null);
      setLastSubmitAt(data?.last_submit_at ?? null);
      form.setFieldsValue({
        home_title_left: data?.home_title_left ?? undefined,
        home_title_right: data?.home_title_right ?? undefined,
        home_subtitle: data?.home_subtitle ?? undefined,
        start_experience: data?.start_experience ?? undefined,
        step1_music_text: data?.step1_music_text ?? undefined,
        step1_title: data?.step1_title ?? undefined,
        step1_music: data?.step1_music ?? undefined,
        step2_music_text: data?.step2_music_text ?? undefined,
        step2_title: data?.step2_title ?? undefined,
        step2_music: data?.step2_music ?? undefined,
        status: (data?.status ?? 1) === 1,
        print_wait: data?.print_wait ?? undefined
      });
      setStep1Preview(data?.step1_music_url ?? null);
      setStep2Preview(data?.step2_music_url ?? null);
      setPrintPreview(data?.print_wait_url ?? null);
    } catch (error) {
      notify.error(error instanceof Error ? error.message : "获取页面配置失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, [version?.id]);

  const handleSave = async () => {
    if (!version?.id) {
      return;
    }
    try {
      const values = await form.validateFields();
      setSaving(true);
      const payload: Record<string, unknown> = {
        draft_version_id: version.id,
        app_version_name_id: version.id,
        home_title_left: values.home_title_left?.trim() || undefined,
        home_title_right: values.home_title_right?.trim() || undefined,
        home_subtitle: values.home_subtitle?.trim() || undefined,
        start_experience: values.start_experience?.trim() || undefined,
        step1_music_text: values.step1_music_text?.trim() || undefined,
        step1_title: values.step1_title?.trim() || undefined,
        step1_music: values.step1_music || undefined,
        step2_music_text: values.step2_music_text?.trim() || undefined,
        step2_title: values.step2_title?.trim() || undefined,
        step2_music: values.step2_music || undefined,
        status: values.status ? 1 : 0,
        print_wait: values.print_wait || undefined,
        updated_by: operatorId ?? undefined
      };
      if (!hasRecord && operatorId) {
        payload.created_by = operatorId;
      }
      await request("/api/draft/app-ui-fields", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      notify.success("页面配置已保存");
      void loadData();
    } catch (error) {
      if (error instanceof Error) {
        notify.error(error.message);
      }
    } finally {
      setSaving(false);
    }
  };

  const handleSaveLocalDraft = () => {
    if (!draftKey) {
      notify.warning("缺少草稿上下文");
      return;
    }
    saveLocalDraft(draftKey, form.getFieldsValue());
    notify.success("本地草稿已保存");
  };

  const handleLoadLocalDraft = () => {
    if (!draftKey) {
      notify.warning("缺少草稿上下文");
      return;
    }
    const data = loadLocalDraft(draftKey);
    if (!data) {
      notify.warning("未找到本地草稿");
      return;
    }
    form.setFieldsValue(data);
    notify.success("本地草稿已恢复");
  };

  const buildSubmissionPayload = () => {
    const values = form.getFieldsValue();
    return {
      id: recordId ?? undefined,
      draft_version_id: version?.id ?? undefined,
      app_version_name_id: version?.id ?? undefined,
      home_title_left: values.home_title_left?.trim() || undefined,
      home_title_right: values.home_title_right?.trim() || undefined,
      home_subtitle: values.home_subtitle?.trim() || undefined,
      start_experience: values.start_experience?.trim() || undefined,
      step1_music_text: values.step1_music_text?.trim() || undefined,
      step1_title: values.step1_title?.trim() || undefined,
      step1_music: values.step1_music || undefined,
      step2_music_text: values.step2_music_text?.trim() || undefined,
      step2_title: values.step2_title?.trim() || undefined,
      step2_music: values.step2_music || undefined,
      status: values.status ? 1 : 0,
      print_wait: values.print_wait || undefined
    };
  };

  if (!version?.id) {
    return <Empty description="请选择景区版本后配置页面字段" />;
  }

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Text type="secondary">配置首页标题、按钮文案、语音与打印中视频。</Text>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadData} loading={loading}>
              刷新
            </Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
              保存配置
            </Button>
            <Button onClick={handleSaveLocalDraft}>保存草稿</Button>
            <Button onClick={handleLoadLocalDraft}>恢复草稿</Button>
            {recordId ? (
              <SubmissionActions
                draftVersionId={version.id}
                moduleKey="app_ui_fields"
                entityTable="app_db_app_ui_fields"
                entityId={recordId}
                operatorId={operatorId ?? null}
                request={request}
                notify={notify}
                getPayload={buildSubmissionPayload}
              />
            ) : null}
          </Space>
          <Space>
            {submitStatus ? (
              (() => {
                const config = submitStatusLabels[submitStatus];
                if (!config) {
                  return <Text type="secondary">{submitStatus}</Text>;
                }
                return <Text type="secondary">{config.label}</Text>;
              })()
            ) : (
              <Text type="secondary">未提交</Text>
            )}
            <Text type="secondary">上次提交：{formatDate(lastSubmitAt)}</Text>
          </Space>
        </Space>
      </Card>

      <Card style={{ borderRadius: 20 }}>
        <Form form={form} layout="vertical" preserve={false}>
          <Title level={5} style={{ marginTop: 0 }}>
            首页文案
          </Title>
          <Form.Item label="标题左侧" name="home_title_left">
            <Input placeholder="如：数枢" />
          </Form.Item>
          <Form.Item label="标题右侧" name="home_title_right">
            <Input placeholder="如：景区" />
          </Form.Item>
          <Form.Item label="副标题" name="home_subtitle">
            <Input placeholder="如：沉浸式 AI 影像体验" />
          </Form.Item>
          <Form.Item label="开始打卡按钮文案" name="start_experience">
            <Input placeholder="如：开始打卡" />
          </Form.Item>

          <Title level={5}>步骤一语音</Title>
          <Form.Item label="语音文案" name="step1_music_text">
            <Input.TextArea rows={2} placeholder="语音播报文案" />
          </Form.Item>
          <TTSInlinePanel
            title="语音生成"
            text={step1TextValue ?? ""}
            disabled={!version?.id}
            presets={ttsPresets}
            request={request}
            onPresetsReload={refreshTtsPresets}
            onGenerate={(count, options) => {
              if (!version?.id) {
                return Promise.resolve([]);
              }
              return generateTTSBatch(generateTTS, step1TextValue?.trim() || "", "app-ui-step1", version.id, count, options);
            }}
            onSelect={(result: TTSResult) => {
              form.setFieldValue("step1_music", result.audio_path);
              setStep1Preview(result.audio_url ?? null);
              notify.success("语音已选择");
            }}
          />
          <Form.Item label="步骤一标题" name="step1_title">
            <Input placeholder="如：选择场景" />
          </Form.Item>
          <Form.Item label="语音文件路径" name="step1_music">
            <Input placeholder="上传或生成后自动填充" />
          </Form.Item>
          <UploadField
            label="步骤一语音文件"
            accept="audio/*"
            value={step1Value}
            previewUrl={step1Preview}
            previewType="audio"
            mediaType="audio"
            moduleKey="app-ui-step1"
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            enableValidation={false}
            enableSmartCompress={false}
            onUpload={async (file) => uploadFile(file, "app-ui-step1", version.id)}
            onChange={(path, url) => {
              form.setFieldValue("step1_music", path);
              setStep1Preview(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("step1_music", undefined);
              setStep1Preview(null);
            }}
          />

          <Title level={5}>步骤二语音</Title>
          <Form.Item label="语音文案" name="step2_music_text">
            <Input.TextArea rows={2} placeholder="语音播报文案" />
          </Form.Item>
          <TTSInlinePanel
            title="语音生成"
            text={step2TextValue ?? ""}
            disabled={!version?.id}
            presets={ttsPresets}
            request={request}
            onPresetsReload={refreshTtsPresets}
            onGenerate={(count, options) => {
              if (!version?.id) {
                return Promise.resolve([]);
              }
              return generateTTSBatch(generateTTS, step2TextValue?.trim() || "", "app-ui-step2", version.id, count, options);
            }}
            onSelect={(result: TTSResult) => {
              form.setFieldValue("step2_music", result.audio_path);
              setStep2Preview(result.audio_url ?? null);
              notify.success("语音已选择");
            }}
          />
          <Form.Item label="步骤二标题" name="step2_title">
            <Input placeholder="如：确认身份" />
          </Form.Item>
          <Form.Item label="语音文件路径" name="step2_music">
            <Input placeholder="上传或生成后自动填充" />
          </Form.Item>
          <UploadField
            label="步骤二语音文件"
            accept="audio/*"
            value={step2Value}
            previewUrl={step2Preview}
            previewType="audio"
            mediaType="audio"
            moduleKey="app-ui-step2"
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            enableValidation={false}
            enableSmartCompress={false}
            onUpload={async (file) => uploadFile(file, "app-ui-step2", version.id)}
            onChange={(path, url) => {
              form.setFieldValue("step2_music", path);
              setStep2Preview(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("step2_music", undefined);
              setStep2Preview(null);
            }}
          />

          <Title level={5}>打印中视频</Title>
          <Form.Item label="打印中视频路径" name="print_wait">
            <Input placeholder="上传后自动填充" />
          </Form.Item>
          <UploadField
            label="打印中视频"
            accept="video/mp4,video/x-m4v"
            value={printValue}
            previewUrl={printPreview}
            previewType="video"
            mediaType="video"
            moduleKey="app_ui_fields:print_wait"
            draftVersionId={version?.id}
            operatorId={operatorId ?? null}
            request={request}
            notify={notify}
            onUpload={async (file) => uploadFile(file, "app_ui_fields:print_wait", version.id)}
            onChange={(path, url) => {
              form.setFieldValue("print_wait", path);
              setPrintPreview(url ?? null);
            }}
            onClear={() => {
              form.setFieldValue("print_wait", undefined);
              setPrintPreview(null);
            }}
          />

          <Form.Item label="启用状态" name="status" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Card>

    </Space>
  );
};

export default AppUIFieldsPanel;
