import express from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import ttsRoutes from './routes/tts.js';
import { apiKeyAuth } from './middleware/auth.js';
import { initializePool, getPoolStatus } from './services/ttsService.js';

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3001;

// 中间件
app.use(cors());
app.use(express.json());

// API Key 验证中间件（所有 /api 路由都需要验证）
app.use('/api', apiKeyAuth);

// 路由
app.use('/api/tts', ttsRoutes);

// 健康检查
app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: '数枢TTS服务运行中' });
});

// 账号池状态（需要API Key验证）
app.get('/api/pool/status', (req, res) => {
  const status = getPoolStatus();
  res.json({
    success: true,
    data: status
  });
});

// 启动服务
const startServer = async () => {
  // 初始化账号池
  console.log('正在初始化账号池...');
  await initializePool();

  app.listen(PORT, () => {
    console.log(`数枢TTS服务已启动，端口: ${PORT}`);
  });
};

startServer().catch(err => {
  console.error('服务启动失败:', err);
  process.exit(1);
});
