package services

import (
  "context"
  "fmt"
  "net/url"
  "strings"
  "time"

  "github.com/aliyun/aliyun-oss-go-sdk/oss"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
)

type OSSService struct {
  bucket         *oss.Bucket
  internalBucket *oss.Bucket
  bucketName     string
  redisClient    *redis.Client
  signTTL        time.Duration
}

// NewOSSService creates a new OSS service instance.
// Args:
//   cfg: App config instance with OSS settings.
//   redisClient: Redis client for signed URL cache.
// Returns:
//   *OSSService: Initialized OSS service.
//   error: Error when config or OSS client initialization fails.
func NewOSSService(cfg *config.Config, redisClient *redis.Client) (*OSSService, error) {
  if cfg.OssEndpoint == "" || cfg.OssAccessKey == "" || cfg.OssSecret == "" || cfg.OssBucket == "" {
    return nil, fmt.Errorf("oss config is incomplete")
  }

  client, err := oss.New(cfg.OssEndpoint, cfg.OssAccessKey, cfg.OssSecret, oss.UseCname(true))
  if err != nil {
    return nil, err
  }

  bucket, err := client.Bucket(cfg.OssBucket)
  if err != nil {
    return nil, err
  }

  var internalBucket *oss.Bucket
  if cfg.OssInternal != "" {
    internalClient, err := oss.New(cfg.OssInternal, cfg.OssAccessKey, cfg.OssSecret, oss.UseCname(true))
    if err == nil {
      internalBucket, _ = internalClient.Bucket(cfg.OssBucket)
    }
  }

  ttl := time.Duration(cfg.OssSignTTL) * time.Second
  if ttl <= 0 {
    ttl = time.Hour
  }

  return &OSSService{
    bucket:         bucket,
    internalBucket: internalBucket,
    bucketName:     cfg.OssBucket,
    redisClient:    redisClient,
    signTTL:        ttl,
  }, nil
}

// GetUploadPreSignedURL builds a pre-signed PUT URL for uploads.
// Args:
//   path: Object path to upload.
//   ttlSeconds: URL expiration seconds. Use config default when <= 0.
// Returns:
//   string: Pre-signed URL.
//   error: Error when generating URL.
func (s *OSSService) GetUploadPreSignedURL(path string, ttlSeconds int64) (string, error) {
  if path == "" {
    return "", fmt.Errorf("path is required")
  }

  ttl := ttlSeconds
  if ttl <= 0 {
    ttl = int64(s.signTTL.Seconds())
  }

  objectKey := s.buildObjectKey(path)
  signedURL, err := s.bucket.SignURL(objectKey, oss.HTTPPut, ttl)
  if err != nil {
    return "", err
  }

  return s.unescapeSignedURL(signedURL), nil
}

// GetSignedURL builds a signed GET URL for an object path.
// Args:
//   path: Object path to sign.
//   internal: Use internal endpoint if available.
//   styleProcess: OSS style process string.
// Returns:
//   string: Signed URL or empty string when path is empty.
//   error: Error when generating URL.
func (s *OSSService) GetSignedURL(path string, internal bool, styleProcess string) (string, error) {
  if path == "" {
    return "", nil
  }

  if strings.HasPrefix(path, "http") {
    return path, nil
  }

  cacheKey := fmt.Sprintf("oss_signed:%t:%s:%s", internal, styleProcess, path)
  if cached, ok := s.getSignedURLFromCache(cacheKey); ok {
    return cached, nil
  }

  objectKey := s.buildObjectKey(path)
  options := []oss.Option{}
  if styleProcess != "" {
    options = append(options, oss.Process(styleProcess))
  }

  bucket := s.bucket
  if internal && s.internalBucket != nil {
    bucket = s.internalBucket
  }

  signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, int64(s.signTTL.Seconds()), options...)
  if err != nil {
    return "", err
  }

  signedURL = s.unescapeSignedURL(signedURL)
  s.setSignedURLCache(cacheKey, signedURL)
  return signedURL, nil
}

// DownloadToFile downloads an OSS object to a local file.
// Args:
//   objectPath: OSS object path.
//   localPath: Local file path.
// Returns:
//   error: Error when download fails.
func (s *OSSService) DownloadToFile(objectPath, localPath string) error {
  if strings.TrimSpace(objectPath) == "" || strings.TrimSpace(localPath) == "" {
    return fmt.Errorf("object path and local path are required")
  }
  objectKey := s.buildObjectKey(objectPath)
  return s.bucket.GetObjectToFile(objectKey, localPath)
}

// UploadFileFromPath uploads a local file to OSS.
// Args:
//   objectPath: OSS object path.
//   localPath: Local file path.
// Returns:
//   error: Error when upload fails.
func (s *OSSService) UploadFileFromPath(objectPath, localPath string) error {
  if strings.TrimSpace(objectPath) == "" || strings.TrimSpace(localPath) == "" {
    return fmt.Errorf("object path and local path are required")
  }
  objectKey := s.buildObjectKey(objectPath)
  return s.bucket.PutObjectFromFile(objectKey, localPath)
}

// buildObjectKey builds the OSS object key with bucket prefix.
// Args:
//   path: Object path without bucket prefix.
// Returns:
//   string: Full object key.
func (s *OSSService) buildObjectKey(path string) string {
  if strings.HasPrefix(path, "/") {
    return s.bucketName + path
  }
  return s.bucketName + "/" + path
}

// unescapeSignedURL restores the path portion of a signed URL.
// Args:
//   raw: Signed URL value.
// Returns:
//   string: URL with unescaped path.
func (s *OSSService) unescapeSignedURL(raw string) string {
  if idx := strings.Index(raw, "?"); idx != -1 {
    pathPart := raw[:idx]
    queryPart := raw[idx:]
    pathPart, _ = url.PathUnescape(pathPart)
    return pathPart + queryPart
  }
  decoded, _ := url.PathUnescape(raw)
  return decoded
}

// getSignedURLFromCache reads a signed URL from cache.
// Args:
//   key: Cache key.
// Returns:
//   string: Cached URL.
//   bool: True when cache hit.
func (s *OSSService) getSignedURLFromCache(key string) (string, bool) {
  if s.redisClient == nil {
    return "", false
  }
  ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
  defer cancel()

  val, err := s.redisClient.Get(ctx, key).Result()
  if err != nil {
    return "", false
  }
  return val, true
}

// setSignedURLCache stores a signed URL in cache.
// Args:
//   key: Cache key.
//   value: Signed URL value.
// Returns:
//   None.
func (s *OSSService) setSignedURLCache(key, value string) {
  if s.redisClient == nil {
    return
  }
  ttl := s.signTTL - 5*time.Minute
  if ttl <= 0 {
    ttl = s.signTTL
  }

  ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
  defer cancel()
  _ = s.redisClient.Set(ctx, key, value, ttl).Err()
}
