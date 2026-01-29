import PreferencePanel from "./PreferencePanel";
import { DraftVersion } from "./constants";
import type { Notify, RequestFn, TTSPreset, TTSFn, UploadFn } from "./utils";

type ClothesPanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  generateTTS: TTSFn;
  notify: Notify;
  operatorId?: number | null;
  ttsPresets?: TTSPreset[];
};

const ClothesPanel = ({ version, request, uploadFile, generateTTS, notify, operatorId, ttsPresets }: ClothesPanelProps) => {
  return (
    <PreferencePanel
      title="服饰偏好"
      hint="服饰偏好用于用户选择拍摄服饰风格，可配置图片与语音说明。"
      listEndpoint="/api/draft/clothes-categories"
      createEndpoint="/api/draft/clothes-categories"
      updateEndpoint="/api/draft/clothes-categories"
      deleteEndpoint="/api/draft/clothes-categories"
      moduleKey="clothes_categories"
      entityTable="app_db_clothes_categories"
      version={version}
      request={request}
      uploadFile={uploadFile}
      generateTTS={generateTTS}
      notify={notify}
      operatorId={operatorId}
      ttsPresets={ttsPresets}
    />
  );
};

export default ClothesPanel;
