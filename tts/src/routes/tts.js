import express from 'express';
import { textToSpeech, downloadAudio, getVoiceDetail } from '../services/ttsService.js';

const router = express.Router();

// 文本转语音接口
router.post('/convert', async (req, res) => {
  try {
    const { text, volume, speed, pitch, stability, similarity, exaggeration, voice_id, emotion_name, accent, country_code } = req.body;

    if (!text || text.trim() === '') {
      return res.status(400).json({
        success: false,
        message: '文本内容不能为空'
      });
    }

    if (text.length > 5000) {
      return res.status(400).json({
        success: false,
        message: '文本长度不能超过5000字符'
      });
    }

    const result = await textToSpeech(text, {
      volume,
      speed,
      pitch,
      stability,
      similarity,
      exaggeration,
      voice_id,
      emotion_name,
      accent,
      country_code
    });

    res.json(result);
  } catch (error) {
    console.error('TTS转换接口错误:', error.message);
    res.status(500).json({
      success: false,
      message: error.message || 'TTS转换失败'
    });
  }
});

// 音频代理下载接口
router.get('/audio', async (req, res) => {
  try {
    const { url } = req.query;

    if (!url) {
      return res.status(400).json({
        success: false,
        message: '缺少音频URL参数'
      });
    }

    const audioBuffer = await downloadAudio(url);

    res.set({
      'Content-Type': 'audio/mpeg',
      'Content-Length': audioBuffer.length,
      'Content-Disposition': 'attachment; filename="tts_audio.mp3"'
    });

    res.send(Buffer.from(audioBuffer));
  } catch (error) {
    console.error('音频下载接口错误:', error.message);
    res.status(500).json({
      success: false,
      message: error.message || '音频下载失败'
    });
  }
});

// 获取语音详情（情绪/口音）
router.post('/voice-detail', async (req, res) => {
  try {
    const { voice_id, slang_id } = req.body;

    if (!voice_id || String(voice_id).trim() === '') {
      return res.status(400).json({
        success: false,
        message: 'voice_id不能为空'
      });
    }

    const detail = await getVoiceDetail(voice_id, slang_id);

    res.json({
      success: true,
      data: detail
    });
  } catch (error) {
    console.error('语音详情接口错误:', error.message);
    res.status(500).json({
      success: false,
      message: error.message || '获取语音详情失败'
    });
  }
});

export default router;
