import PreferencePanel from "./PreferencePanel";
import { DraftVersion } from "./constants";
import type { Notify, RequestFn, TTSPreset, TTSFn, UploadFn } from "./utils";

type PhotoPanelProps = {
  version: DraftVersion | null;
  request: RequestFn;
  uploadFile: UploadFn;
  generateTTS: TTSFn;
  notify: Notify;
  operatorId?: number | null;
  ttsPresets?: TTSPreset[];
  refreshTtsPresets?: () => void;
};

const PhotoPanel = ({ version, request, uploadFile, generateTTS, notify, operatorId, ttsPresets, refreshTtsPresets }: PhotoPanelProps) => {
  return (
    <PreferencePanel
      title="拍摄偏好"
      hint="拍摄偏好用于用户选择拍摄风格，可配置图片、语音与介绍文案。"
      listEndpoint="/api/draft/photo-hobbies"
      createEndpoint="/api/draft/photo-hobbies"
      updateEndpoint="/api/draft/photo-hobbies"
      deleteEndpoint="/api/draft/photo-hobbies"
      moduleKey="photo_hobbies"
      entityTable="app_db_photo_hobbies"
      version={version}
      request={request}
      uploadFile={uploadFile}
      generateTTS={generateTTS}
      notify={notify}
      operatorId={operatorId}
      ttsPresets={ttsPresets}
      refreshTtsPresets={refreshTtsPresets}
    />
  );
};

export default PhotoPanel;
