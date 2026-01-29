import axios from 'axios';
import FormData from 'form-data';
import CryptoJS from 'crypto-js';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { HttpsProxyAgent } from 'https-proxy-agent';
import { SocksProxyAgent } from 'socks-proxy-agent';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const TTS_API_BASE = 'https://tts-api.imyfone.com';

// AES解密配置
const AES_KEY = '123456789imyfone123456789imyfone';
const AES_IV = '123456789imyfone';

// 账号池配置
const POOL_SIZE = parseInt(process.env.ACCOUNT_POOL_SIZE) || 10; // 账号池大小
const POOL_CHECK_INTERVAL = parseInt(process.env.POOL_CHECK_INTERVAL) || 60000; // 检查间隔（毫秒）

// 日志目录
const LOG_DIR = path.join(__dirname, '../../logs');
// 数据目录（用于持久化）
const DATA_DIR = path.join(__dirname, '../../data');
// 账号池持久化文件
const ACCOUNT_POOL_FILE = path.join(DATA_DIR, 'account_pool.json');

// 确保日志目录存在
if (!fs.existsSync(LOG_DIR)) {
  fs.mkdirSync(LOG_DIR, { recursive: true });
}

// 确保数据目录存在
if (!fs.existsSync(DATA_DIR)) {
  fs.mkdirSync(DATA_DIR, { recursive: true });
}

// 格式化时间
const formatTime = (date = new Date()) => {
  return date.toISOString().replace('T', ' ').substring(0, 19);
};

// 获取日志文件路径（按天分割）
const getLogFilePath = () => {
  const date = new Date().toISOString().split('T')[0];
  return path.join(LOG_DIR, `api_${date}.log`);
};

// 写入日志
const writeLog = (type, data) => {
  const logEntry = {
    timestamp: formatTime(),
    type,
    ...data
  };

  const logLine = JSON.stringify(logEntry) + '\n';
  const logFile = getLogFilePath();

  fs.appendFileSync(logFile, logLine, 'utf8');

  // 同时输出到控制台
  console.log(`[${logEntry.timestamp}] [${type}]`, JSON.stringify(data, null, 2));
};

// 记录API请求日志
const logRequest = (method, url, headers, body) => {
  writeLog('REQUEST', {
    method,
    url,
    headers: { ...headers, 'Content-Type': headers['Content-Type'] || headers['content-type'] },
    body: body || null
  });
};

// 记录API响应日志
const logResponse = (method, url, status, data, duration) => {
  writeLog('RESPONSE', {
    method,
    url,
    status,
    duration: `${duration}ms`,
    data
  });
};

// 记录错误日志
const logError = (method, url, error, duration) => {
  writeLog('ERROR', {
    method,
    url,
    duration: `${duration}ms`,
    error: error.message,
    stack: error.stack
  });
};

// 记录账号池日志
const logPool = (action, data) => {
  writeLog('POOL', {
    action,
    ...data
  });
};

// 记录代理日志
const logProxy = (action, data) => {
  writeLog('PROXY', {
    action,
    ...data
  });
};

// ==================== 代理IP池管理 ====================

const PROXY_API_URL = process.env.PROXY_API_URL || 'https://share.proxy.qg.net/get?key=UTQNELRF&pwd=098F606CFF96';
const PROXY_USERNAME = process.env.PROXY_USERNAME || 'UTQNELRF';
const PROXY_PASSWORD = process.env.PROXY_PASSWORD || '098F606CFF96';
const PROXY_DEFAULT_PROTOCOL = (process.env.PROXY_PROTOCOL || 'http').toLowerCase();
const PROXY_REFRESH_INTERVAL = 1 * 60 * 1000; // 5分钟刷新一次代理列表
const PROXY_FETCH_COUNT = 5; // 一次获取5条代理

class ProxyPool {
  constructor() {
    this.proxies = [];
    this.currentIndex = 0;
    this.lastRefresh = 0;
    this.isRefreshing = false;
    this.refreshPromise = null; // 用于等待正在进行的刷新
  }

  // 获取代理列表
  async fetchProxies() {
    // 如果正在刷新，等待当前刷新完成
    if (this.isRefreshing && this.refreshPromise) {
      await this.refreshPromise;
      return;
    }

    this.isRefreshing = true;
    const startTime = Date.now();

    // 创建一个 Promise 供其他调用等待
    this.refreshPromise = (async () => {
      try {
        logProxy('FETCH_START', {
          count: PROXY_FETCH_COUNT
        });

        // 获取代理列表
        const response = await axios.get(PROXY_API_URL, {
          params: {
            num: PROXY_FETCH_COUNT
          },
          timeout: 10000
        });

        const fetchDuration = Date.now() - startTime;

        if (response.data.code === 'SUCCESS' && response.data.data) {
          const rawProxies = response.data.data.map(p => {
            // server 格式为 "ip:port"
            const [ip, port] = p.server.split(':');
            return {
              ip,
              port: parseInt(port),
              area: p.area,
              isp: p.isp,
              deadline: p.deadline,
              protocol: (p.protocol || p.scheme || PROXY_DEFAULT_PROTOCOL || 'http').toLowerCase(),
              username: p.username || p.user || p.account || p.uname || '',
              password: p.password || p.pass || p.pwd || ''
            };
          });

          this.proxies = rawProxies;
          this.currentIndex = 0;
          this.lastRefresh = Date.now();

          logProxy('FETCH_SUCCESS', {
            count: rawProxies.length,
            duration: `${fetchDuration}ms`
          });
        } else {
          logProxy('FETCH_FAILED', {
            code: response.data.code,
            msg: response.data.msg || response.data.message,
            duration: `${fetchDuration}ms`
          });
        }
      } catch (error) {
        const duration = Date.now() - startTime;
        logProxy('FETCH_ERROR', {
          error: error.message,
          duration: `${duration}ms`
        });
      } finally {
        this.isRefreshing = false;
        this.refreshPromise = null;
      }
    })();

    await this.refreshPromise;
  }

  // 获取下一个代理
  async getProxy() {
    // 检查是否需要刷新代理列表
    const now = Date.now();
    if (this.proxies.length === 0 || now - this.lastRefresh > PROXY_REFRESH_INTERVAL) {
      await this.fetchProxies();
    }

    if (this.proxies.length === 0) {
      logProxy('NO_PROXY', { message: '无可用代理' });
      return null;
    }

    // 轮询获取代理
    const proxy = this.proxies[this.currentIndex];
    this.currentIndex = (this.currentIndex + 1) % this.proxies.length;

    return proxy;
  }

  // 创建代理Agent
  createAgent(proxy) {
    if (!proxy) return null;

    const { ip, port } = proxy;
    const protocol = (proxy.protocol || PROXY_DEFAULT_PROTOCOL || 'http').toLowerCase();
    const username = proxy.username || PROXY_USERNAME;
    const password = proxy.password || PROXY_PASSWORD;

    try {
      const auth = username || password
        ? `${encodeURIComponent(username)}:${encodeURIComponent(password || '')}@`
        : '';

      if (protocol.startsWith('socks')) {
        const proxyUrl = `${protocol}://${auth}${ip}:${port}`;
        return new SocksProxyAgent(proxyUrl);
      }

      const proxyUrl = `http://${auth}${ip}:${port}`;
      return new HttpsProxyAgent(proxyUrl);
    } catch (error) {
      logProxy('CREATE_AGENT_ERROR', {
        proxy: `${ip}:${port}`,
        error: error.message
      });
      return null;
    }
  }

  // 标记代理失败（从列表中移除）
  markFailed(proxy) {
    if (!proxy) return;

    const index = this.proxies.findIndex(p => p.ip === proxy.ip && p.port === proxy.port);
    if (index !== -1) {
      this.proxies.splice(index, 1);
      logProxy('MARK_FAILED', {
        proxy: `${proxy.ip}:${proxy.port}`,
        remainingCount: this.proxies.length
      });
    }
  }

  // 获取代理池状态
  getStatus() {
    return {
      total: this.proxies.length,
      lastRefresh: this.lastRefresh ? new Date(this.lastRefresh).toISOString() : null
    };
  }
}

// 全局代理池实例
const proxyPool = new ProxyPool();

// ==================== 账号池管理 ====================

// 账号状态
const AccountStatus = {
  IDLE: 'idle',       // 空闲
  BUSY: 'busy',       // 使用中
  INVALID: 'invalid'  // 失效
};

// 账号池
class AccountPool {
  constructor() {
    this.accounts = [];
    this.isInitialized = false;
    this.isInitializing = false;
    this.checkTimer = null;
  }

  // 从文件加载账号池
  loadFromFile() {
    try {
      if (fs.existsSync(ACCOUNT_POOL_FILE)) {
        const data = fs.readFileSync(ACCOUNT_POOL_FILE, 'utf8');
        const savedAccounts = JSON.parse(data);

        // 过滤掉失效的账号，只加载有效的
        this.accounts = savedAccounts
          .filter(acc => acc.status !== AccountStatus.INVALID)
          .map(acc => ({
            ...acc,
            status: AccountStatus.IDLE // 重启后所有账号重置为空闲
          }));

        logPool('LOAD_FROM_FILE', {
          loadedCount: this.accounts.length,
          originalCount: savedAccounts.length,
          filteredOut: savedAccounts.length - this.accounts.length
        });

        return true;
      }
    } catch (error) {
      logPool('LOAD_FILE_ERROR', {
        error: error.message
      });
    }
    return false;
  }

  // 保存账号池到文件
  saveToFile() {
    try {
      // 只保存非失效的账号
      const accountsToSave = this.accounts
        .filter(acc => acc.status !== AccountStatus.INVALID)
        .map(acc => ({
          id: acc.id,
          tourist_id: acc.tourist_id,
          session_id: acc.session_id,
          device_id: acc.device_id,
          status: acc.status,
          createdAt: acc.createdAt,
          lastUsedAt: acc.lastUsedAt,
          useCount: acc.useCount,
          errorCount: acc.errorCount
        }));

      fs.writeFileSync(ACCOUNT_POOL_FILE, JSON.stringify(accountsToSave, null, 2), 'utf8');

      logPool('SAVE_TO_FILE', {
        savedCount: accountsToSave.length
      });
    } catch (error) {
      logPool('SAVE_FILE_ERROR', {
        error: error.message
      });
    }
  }

  // 初始化账号池
  async initialize() {
    if (this.isInitialized || this.isInitializing) return;

    this.isInitializing = true;
    logPool('INIT_START', { targetSize: POOL_SIZE });

    try {
      // 先尝试从文件加载
      const loaded = this.loadFromFile();
      const currentCount = this.accounts.length;

      if (loaded && currentCount > 0) {
        logPool('INIT_LOADED', {
          loadedCount: currentCount,
          targetSize: POOL_SIZE
        });
      }

      // 如果账号不足，补充新账号
      const needCount = Math.max(0, POOL_SIZE - currentCount);

      if (needCount > 0) {
        logPool('INIT_NEED_MORE', {
          currentCount,
          needCount
        });

        // 并发注册账号
        const promises = [];
        for (let i = 0; i < needCount; i++) {
          promises.push(this.registerAccount().catch(() => null));
        }

        const results = await Promise.allSettled(promises);
        const successCount = results.filter(r => r.status === 'fulfilled' && r.value !== null).length;

        logPool('INIT_REGISTER_COMPLETE', {
          needCount,
          successCount,
          failCount: needCount - successCount
        });
      }

      logPool('INIT_COMPLETE', {
        targetSize: POOL_SIZE,
        actualSize: this.accounts.length,
        status: this.getStatus()
      });

      this.isInitialized = true;

      // 启动定期检查
      this.startPeriodicCheck();
    } catch (error) {
      logPool('INIT_ERROR', { error: error.message });
    } finally {
      this.isInitializing = false;
    }
  }

  // 注册单个账号（使用代理IP）
  async registerAccount() {
    // 生成32位随机十六进制字符串，模拟真实设备ID
    const deviceId = Array.from({ length: 32 }, () =>
      Math.floor(Math.random() * 16).toString(16)
    ).join('');
    const url = `${TTS_API_BASE}/v3/user/tourist_id`;
    const headers = { 'Device': deviceId };
    const startTime = Date.now();

    // 获取代理
    const proxy = await proxyPool.getProxy();
    const agent = proxyPool.createAgent(proxy);
    const proxyProtocol = proxy && proxy.protocol ? proxy.protocol : PROXY_DEFAULT_PROTOCOL;
    const proxyInfo = proxy ? `${proxyProtocol}://${proxy.ip}:${proxy.port}` : 'none';

    logPool('REGISTER_START', {
      deviceId,
      proxy: proxyInfo,
      proxyAuth: !!(proxy && (proxy.username || proxy.password || PROXY_USERNAME || PROXY_PASSWORD))
    });

    try {
      const axiosConfig = {
        headers,
        timeout: 15000
      };

      // 如果有代理，添加代理配置
      if (agent) {
        axiosConfig.httpsAgent = agent;
        axiosConfig.httpAgent = agent;
        axiosConfig.proxy = false;
      }

      const response = await axios.get(url, axiosConfig);
      const duration = Date.now() - startTime;

      if (response.data.check_code === 200000) {
        const account = {
          id: `acc_${Date.now()}_${Math.random().toString(36).substring(2, 8)}`,
          tourist_id: response.data.data.tourist_id,
          session_id: response.data.data.session_id,
          device_id: deviceId,
          status: AccountStatus.IDLE,
          createdAt: Date.now(),
          lastUsedAt: null,
          useCount: 0,
          errorCount: 0
        };

        this.accounts.push(account);

        // 持久化账号池
        this.saveToFile();

        logPool('REGISTER_SUCCESS', {
          accountId: account.id,
          touristId: account.tourist_id,
          deviceId: account.device_id,
          proxy: proxyInfo,
          duration: `${duration}ms`,
          poolSize: this.accounts.length
        });
        return account;
      }

      // 注册失败，标记代理可能有问题
      if (proxy) {
        proxyPool.markFailed(proxy);
      }
      throw new Error(`注册失败: ${response.data.message || 'unknown'}`);
    } catch (error) {
      const duration = Date.now() - startTime;

      // 如果是网络错误，标记代理失败
      if (proxy && (error.code === 'ECONNREFUSED' || error.code === 'ETIMEDOUT' || error.code === 'ECONNRESET')) {
        proxyPool.markFailed(proxy);
      }

      logPool('REGISTER_ERROR', {
        deviceId,
        proxy: proxyInfo,
        error: error.message,
        errorCode: error.code,
        duration: `${duration}ms`
      });
      throw error;
    }
  }

  // 获取空闲账号
  async acquireAccount() {
    // 确保已初始化
    if (!this.isInitialized) {
      await this.initialize();
    }

    // 查找空闲账号
    const idleAccount = this.accounts.find(acc => acc.status === AccountStatus.IDLE);

    if (idleAccount) {
      idleAccount.status = AccountStatus.BUSY;
      idleAccount.lastUsedAt = Date.now();
      logPool('ACQUIRE', {
        accountId: idleAccount.id,
        touristId: idleAccount.tourist_id,
        idleRemaining: this.getIdleCount(),
        status: this.getStatus()
      });
      return idleAccount;
    }

    // 没有空闲账号，尝试注册新的
    logPool('ACQUIRE_NO_IDLE', {
      status: this.getStatus(),
      message: '无空闲账号，尝试注册新账号'
    });

    try {
      const newAccount = await this.registerAccount();
      newAccount.status = AccountStatus.BUSY;
      newAccount.lastUsedAt = Date.now();
      logPool('ACQUIRE_NEW', {
        accountId: newAccount.id,
        touristId: newAccount.tourist_id,
        status: this.getStatus()
      });
      return newAccount;
    } catch (error) {
      logPool('ACQUIRE_FAILED', {
        error: error.message,
        status: this.getStatus()
      });
      throw new Error('无可用账号，请稍后重试');
    }
  }

  // 释放账号
  releaseAccount(accountId, success = true) {
    const account = this.accounts.find(acc => acc.id === accountId);
    if (account) {
      account.useCount++;

      if (success) {
        account.status = AccountStatus.IDLE;
        account.errorCount = 0;
        logPool('RELEASE', {
          accountId: account.id,
          touristId: account.tourist_id,
          success: true,
          useCount: account.useCount,
          status: this.getStatus()
        });
      } else {
        account.errorCount++;
        // 连续失败3次标记为失效
        if (account.errorCount >= 3) {
          account.status = AccountStatus.INVALID;
          logPool('RELEASE_INVALID', {
            accountId: account.id,
            touristId: account.tourist_id,
            errorCount: account.errorCount,
            reason: '连续失败3次，标记为失效',
            status: this.getStatus()
          });
        } else {
          account.status = AccountStatus.IDLE;
          logPool('RELEASE', {
            accountId: account.id,
            touristId: account.tourist_id,
            success: false,
            errorCount: account.errorCount,
            useCount: account.useCount,
            status: this.getStatus()
          });
        }
      }

      // 持久化账号池
      this.saveToFile();
    }
  }

  // 标记账号失效
  invalidateAccount(accountId) {
    const account = this.accounts.find(acc => acc.id === accountId);
    if (account) {
      account.status = AccountStatus.INVALID;
      logPool('INVALIDATE', {
        accountId: account.id,
        touristId: account.tourist_id,
        reason: '手动标记失效',
        status: this.getStatus()
      });

      // 持久化账号池
      this.saveToFile();
    }
  }

  // 获取空闲账号数量
  getIdleCount() {
    return this.accounts.filter(acc => acc.status === AccountStatus.IDLE).length;
  }

  // 获取池状态
  getStatus() {
    return {
      total: this.accounts.length,
      idle: this.accounts.filter(acc => acc.status === AccountStatus.IDLE).length,
      busy: this.accounts.filter(acc => acc.status === AccountStatus.BUSY).length,
      invalid: this.accounts.filter(acc => acc.status === AccountStatus.INVALID).length
    };
  }

  // 定期检查和补充账号
  startPeriodicCheck() {
    if (this.checkTimer) return;

    this.checkTimer = setInterval(async () => {
      await this.checkAndReplenish();
    }, POOL_CHECK_INTERVAL);

    logPool('PERIODIC_CHECK_START', {
      interval: `${POOL_CHECK_INTERVAL}ms`,
      status: this.getStatus()
    });
  }

  // 检查并补充账号
  async checkAndReplenish() {
    logPool('CHECK_START', { status: this.getStatus() });

    // 移除失效账号
    const invalidAccounts = this.accounts.filter(acc => acc.status === AccountStatus.INVALID);
    if (invalidAccounts.length > 0) {
      const removedIds = invalidAccounts.map(acc => acc.touristId);
      this.accounts = this.accounts.filter(acc => acc.status !== AccountStatus.INVALID);
      logPool('REMOVE_INVALID', {
        removedCount: invalidAccounts.length,
        removedIds,
        status: this.getStatus()
      });

      // 移除失效账号后持久化
      this.saveToFile();
    }

    // 补充账号
    const currentCount = this.accounts.length;
    const needCount = POOL_SIZE - currentCount;

    if (needCount > 0) {
      logPool('REPLENISH_START', {
        currentCount,
        targetSize: POOL_SIZE,
        needCount
      });

      const promises = [];
      for (let i = 0; i < needCount; i++) {
        promises.push(this.registerAccount().catch(() => null));
      }

      const results = await Promise.all(promises);
      const successCount = results.filter(r => r !== null).length;

      logPool('REPLENISH_COMPLETE', {
        needCount,
        successCount,
        failCount: needCount - successCount,
        status: this.getStatus()
      });
    } else {
      logPool('CHECK_COMPLETE', {
        message: '账号池充足，无需补充',
        status: this.getStatus()
      });
    }
  }

  // 停止定期检查
  stopPeriodicCheck() {
    if (this.checkTimer) {
      clearInterval(this.checkTimer);
      this.checkTimer = null;
    }
  }
}

// 全局账号池实例
const accountPool = new AccountPool();

// ==================== TTS 服务 ====================

// AES解密
export const decryptUrl = (encryptedUrl) => {
  try {
    const key = CryptoJS.enc.Utf8.parse(AES_KEY);
    const iv = CryptoJS.enc.Utf8.parse(AES_IV);

    const decrypted = CryptoJS.AES.decrypt(encryptedUrl, key, {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7
    });

    return decrypted.toString(CryptoJS.enc.Utf8);
  } catch (error) {
    console.error('解密URL失败:', error.message);
    throw new Error('解密音频URL失败');
  }
};

// 文本转语音
export const textToSpeech = async (text, options = {}) => {
  const {
    volume = 58,
    speed = 1,
    pitch = 56,
    stability = 50,
    similarity = 95,
    exaggeration = 0,
    voice_id = '70eb6772-4cd1-11f0-9276-00163e0fe4f9',
    emotion_name = 'Happy',
    accent = 'Chinese(Mandarin)',
    country_code = 'JP'
  } = options;

  // 从账号池获取账号
  const account = await accountPool.acquireAccount();

  const makeRequest = async (acc) => {
    const url = `${TTS_API_BASE}/v5/voice/tts_tourist`;
    const formData = new FormData();
    formData.append('accent', accent);
    formData.append('emotion_name', emotion_name);
    formData.append('text', `<speak>${text}</speak>`);
    formData.append('speed', String(speed));
    formData.append('volume', String(volume));
    formData.append('voice_id', voice_id);
    formData.append('article_title', 'Unnamed');
    formData.append('session_id', acc.session_id);
    formData.append('tourist_id', acc.tourist_id);
    formData.append('is_audition', '1');
    formData.append('pitch', String(pitch));
    formData.append('stability', String(stability));
    formData.append('similarity', String(similarity));
    formData.append('exaggeration', String(exaggeration));
    formData.append('plan_type', '2');
    formData.append('country_code', country_code);

    const headers = {
      'Accept': 'application/json, text/plain, */*',
      'Accept-Language': 'zh-CN,zh;q=0.9,ja;q=0.8,ko;q=0.7,fr;q=0.6,de;q=0.5,zh-TW;q=0.4,ru;q=0.3,en;q=0.2,el;q=0.1,it;q=0.1',
      'Connection': 'keep-alive',
      'Device': acc.device_id,
      'Origin': 'https://www.topmediai.com',
      'Referer': 'https://www.topmediai.com/',
      'Sec-Fetch-Dest': 'empty',
      'Sec-Fetch-Mode': 'cors',
      'Sec-Fetch-Site': 'cross-site',
      'Site-Initializing': 'www.topmediai.com',
      'TouristCode': acc.tourist_id,
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36',
      'Web-Req': '1',
      'X-Pnab': '1',
      'X-Requested-With': 'TTS',
      ...formData.getHeaders()
    };

    // 记录请求体
    const requestBody = {
      accent,
      emotion_name,
      text: text.length > 100 ? text.substring(0, 100) + '...' : text,
      speed,
      volume,
      voice_id,
      pitch,
      stability,
      similarity,
      exaggeration,
      country_code,
      tourist_id: acc.tourist_id
    };

    const startTime = Date.now();
    logRequest('POST', url, headers, requestBody);

    try {
      const response = await axios.post(url, formData, { headers });
      const duration = Date.now() - startTime;

      logResponse('POST', url, response.status, response.data, duration);

      return response.data;
    } catch (error) {
      const duration = Date.now() - startTime;
      logError('POST', url, error, duration);
      throw error;
    }
  };

  try {
    const result = await makeRequest(account);

    if (result.check_code !== 200000) {
      const message = result.message || '';

      // 精确匹配额度不足错误，直接废弃账号
      if (message.includes('left characters is not enough')) {
        accountPool.invalidateAccount(account.id);
        logPool('QUOTA_EXCEEDED', {
          accountId: account.id,
          touristId: account.tourist_id,
          checkCode: result.check_code,
          message: message,
          reason: '额度不足，立即废弃'
        });
      } else {
        // 其他错误，按失败计数（连续3次失败才废弃）
        accountPool.releaseAccount(account.id, false);
      }

      throw new Error(result.message || 'TTS转换失败');
    }

    // 成功，释放账号
    accountPool.releaseAccount(account.id, true);

    // 解密音频URL
    const audioUrl = decryptUrl(result.data.oss_url);

    return {
      success: true,
      audioUrl,
      data: {
        id: result.data.id,
        displayName: result.data.display_name,
        voiceAvatar: result.data.voice_avatar_url
      }
    };
  } catch (error) {
    // 如果是网络错误等异常，释放账号（标记失败）
    const accountStillBusy = accountPool.accounts.find(
      acc => acc.id === account.id && acc.status === AccountStatus.BUSY
    );
    if (accountStillBusy) {
      accountPool.releaseAccount(account.id, false);
    }
    throw error;
  }
};

// 获取语音详情（含情绪/口音信息）
export const getVoiceDetail = async (voiceId, slangId = 18) => {
  if (!voiceId || !String(voiceId).trim()) {
    throw new Error('voice_id is required');
  }

  const account = await accountPool.acquireAccount();

  const makeRequest = async (acc) => {
    const url = `${TTS_API_BASE}/v3/voice/detail`;
    const formData = new FormData();
    formData.append('voice_id', String(voiceId).trim());
    formData.append('slang_id', String(slangId || 18));
    formData.append('session_id', acc.session_id);
    formData.append('tourist_id', acc.tourist_id);

    const headers = {
      'Accept': 'application/json, text/plain, */*',
      'Accept-Language': 'zh-CN,zh;q=0.9,ja;q=0.8,ko;q=0.7,fr;q=0.6,de;q=0.5,zh-TW;q=0.4,ru;q=0.3,en;q=0.2,el;q=0.1,it;q=0.1',
      'Connection': 'keep-alive',
      'Device': acc.device_id,
      'Origin': 'https://www.topmediai.com',
      'Referer': 'https://www.topmediai.com/',
      'Sec-Fetch-Dest': 'empty',
      'Sec-Fetch-Mode': 'cors',
      'Sec-Fetch-Site': 'cross-site',
      'Site-Initializing': 'www.topmediai.com',
      'TouristCode': acc.tourist_id,
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36',
      'Web-Req': '1',
      'X-Pnab': '1',
      'X-Requested-With': 'TTS',
      ...formData.getHeaders()
    };

    const requestBody = {
      voice_id: String(voiceId).trim(),
      slang_id: String(slangId || 18),
      tourist_id: acc.tourist_id
    };

    const startTime = Date.now();
    logRequest('POST', url, headers, requestBody);

    try {
      const response = await axios.post(url, formData, { headers });
      const duration = Date.now() - startTime;
      logResponse('POST', url, response.status, response.data, duration);
      return response.data;
    } catch (error) {
      const duration = Date.now() - startTime;
      logError('POST', url, error, duration);
      throw error;
    }
  };

  try {
    const result = await makeRequest(account);

    if (result.check_code !== 200000) {
      accountPool.releaseAccount(account.id, false);
      throw new Error(result.message || '获取语音详情失败');
    }

    accountPool.releaseAccount(account.id, true);
    return result.data;
  } catch (error) {
    const accountStillBusy = accountPool.accounts.find(
      acc => acc.id === account.id && acc.status === AccountStatus.BUSY
    );
    if (accountStillBusy) {
      accountPool.releaseAccount(account.id, false);
    }
    throw error;
  }
};

// 下载音频文件
export const downloadAudio = async (url) => {
  const startTime = Date.now();
  logRequest('GET', url, {});

  try {
    const response = await axios.get(url, {
      responseType: 'arraybuffer'
    });
    const duration = Date.now() - startTime;

    logResponse('GET', url, response.status, { size: response.data.length }, duration);

    return response.data;
  } catch (error) {
    const duration = Date.now() - startTime;
    logError('GET', url, error, duration);
    throw new Error('下载音频文件失败');
  }
};

// 获取账号池状态
export const getPoolStatus = () => {
  return accountPool.getStatus();
};

// 初始化账号池（服务启动时调用）
export const initializePool = async () => {
  await accountPool.initialize();
};
