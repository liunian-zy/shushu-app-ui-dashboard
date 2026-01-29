import { Col, Form, Input, InputNumber, Modal, Row, Select, Switch } from "antd";
import { moduleOptions, statusOptions, TaskItem, UserItem } from "./constants";

const { TextArea } = Input;

export type TaskEditorValues = {
  module_key: string;
  title: string;
  description?: string;
  assigned_to?: number | null;
  allow_assist: boolean;
  priority?: number;
  status?: string;
};

type TaskEditorModalProps = {
  open: boolean;
  task: TaskItem | null;
  users: UserItem[];
  userLoading: boolean;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (values: TaskEditorValues) => void;
};

const TaskEditorModal = ({
  open,
  task,
  users,
  userLoading,
  submitting,
  onCancel,
  onSubmit
}: TaskEditorModalProps) => {
  const [form] = Form.useForm<TaskEditorValues>();

  const syncEditorForm = (item: TaskItem | null) => {
    form.setFieldsValue({
      module_key: item?.module_key ?? undefined,
      title: item?.title ?? undefined,
      description: item?.description ?? undefined,
      assigned_to: item?.assigned_to ?? undefined,
      allow_assist: (item?.allow_assist ?? 1) === 1,
      priority: item?.priority ?? 0,
      status: item?.status ?? "open"
    });
  };

  const handleOk = async () => {
    const values = await form.validateFields();
    onSubmit(values);
  };

  return (
    <Modal
      title={task ? "编辑任务" : "新建任务"}
      open={open}
      onCancel={onCancel}
      width={860}
      afterOpenChange={(nextOpen) => {
        if (nextOpen) {
          syncEditorForm(task);
        }
      }}
      onOk={handleOk}
      okButtonProps={{ loading: submitting }}
      styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
      destroyOnClose
    >
      <Form form={form} layout="vertical" preserve={false}>
        <Row gutter={12}>
          <Col span={12}>
            <Form.Item
              label="任务模块"
              name="module_key"
              rules={[{ required: true, message: "请输入任务模块" }]}
            >
              <Select
                placeholder="选择任务模块"
                options={moduleOptions}
                showSearch
                optionFilterProp="label"
                allowClear
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="指派给" name="assigned_to">
              <Select
                placeholder="选择负责人"
                loading={userLoading}
                allowClear
                options={users.map((item) => ({
                  value: item.id,
                  label: item.display_name || item.username || `用户${item.id}`
                }))}
              />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item label="任务标题" name="title" rules={[{ required: true, message: "请输入任务标题" }]}>
          <Input placeholder="填写任务名称" />
        </Form.Item>
        <Form.Item label="任务描述" name="description">
          <TextArea rows={3} placeholder="补充说明或注意事项" />
        </Form.Item>
        <Row gutter={12}>
          <Col span={8}>
            <Form.Item label="允许协作" name="allow_assist" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item label="优先级" name="priority">
              <InputNumber min={0} max={10} style={{ width: "100%" }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item label="状态" name="status">
              <Select options={statusOptions.map((item) => ({ value: item.value, label: item.label }))} />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Modal>
  );
};

export default TaskEditorModal;
