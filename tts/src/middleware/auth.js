export const apiKeyAuth = (req, res, next) => {
  const apiKey = req.headers['x-api-key'];

  if (!apiKey) {
    return res.status(401).json({
      success: false,
      message: '缺少API密钥'
    });
  }

  if (apiKey !== process.env.API_KEY) {
    return res.status(403).json({
      success: false,
      message: 'API密钥无效'
    });
  }

  next();
};
