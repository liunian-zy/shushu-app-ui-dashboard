import { Form, Input, Modal } from "antd";

const { TextArea } = Input;

type TaskAssistModalProps = {
  open: boolean;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (note?: string) => void;
};

const TaskAssistModal = ({ open, submitting, onCancel, onSubmit }: TaskAssistModalProps) => {
  const [form] = Form.useForm<{ note?: string }>();

  const handleOk = async () => {
    const values = await form.validateFields();
    onSubmit(values.note);
  };

  return (
    <Modal
      title="协作提交"
      open={open}
      onCancel={onCancel}
      width={720}
      onOk={handleOk}
      okButtonProps={{ loading: submitting }}
      styles={{ body: { maxHeight: "70vh", overflowY: "auto" } }}
      destroyOnClose
    >
      <Form form={form} layout="vertical" preserve={false}>
        <Form.Item label="协作说明" name="note">
          <TextArea rows={4} placeholder="补充你完成的部分或操作说明" />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default TaskAssistModal;
