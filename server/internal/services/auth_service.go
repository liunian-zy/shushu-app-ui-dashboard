package services

import (
  "errors"
  "time"

  "github.com/golang-jwt/jwt/v5"
  "golang.org/x/crypto/bcrypt"

  "shushu-app-ui-dashboard/internal/config"
)

type AuthService struct {
  secret       []byte
  issuer       string
  expireWindow time.Duration
}

type AuthUser struct {
  ID          int64
  Username    string
  DisplayName string
  Role        string
}

type AuthClaims struct {
  UserID      int64  `json:"user_id"`
  Username    string `json:"username"`
  DisplayName string `json:"display_name"`
  Role        string `json:"role"`
  jwt.RegisteredClaims
}

// NewAuthService creates an auth service instance.
// Args:
//   cfg: App config instance.
// Returns:
//   *AuthService: Initialized service.
//   error: Error when config is invalid.
func NewAuthService(cfg *config.Config) (*AuthService, error) {
  if cfg == nil || cfg.JwtSecret == "" {
    return nil, errors.New("jwt secret is required")
  }
  issuer := cfg.JwtIssuer
  if issuer == "" {
    issuer = "shushu-app-ui-dashboard"
  }
  expireHours := cfg.JwtExpireHours
  if expireHours <= 0 {
    expireHours = 24
  }
  return &AuthService{
    secret:       []byte(cfg.JwtSecret),
    issuer:       issuer,
    expireWindow: time.Duration(expireHours) * time.Hour,
  }, nil
}

// HashPassword hashes a password using bcrypt.
// Args:
//   password: Raw password string.
// Returns:
//   string: Password hash.
//   error: Error when hashing fails.
func (s *AuthService) HashPassword(password string) (string, error) {
  if password == "" {
    return "", errors.New("password is required")
  }
  raw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    return "", err
  }
  return string(raw), nil
}

// VerifyPassword compares a hash with a password.
// Args:
//   hash: Password hash.
//   password: Raw password string.
// Returns:
//   bool: True when match.
func (s *AuthService) VerifyPassword(hash, password string) bool {
  if hash == "" || password == "" {
    return false
  }
  return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// IssueToken issues a JWT for a user.
// Args:
//   user: Auth user data.
// Returns:
//   string: Signed token.
//   time.Time: Expiration time.
//   error: Error when signing fails.
func (s *AuthService) IssueToken(user *AuthUser) (string, time.Time, error) {
  if user == nil || user.ID <= 0 || user.Username == "" {
    return "", time.Time{}, errors.New("invalid user")
  }

  expiresAt := time.Now().Add(s.expireWindow)
  claims := AuthClaims{
    UserID:      user.ID,
    Username:    user.Username,
    DisplayName: user.DisplayName,
    Role:        user.Role,
    RegisteredClaims: jwt.RegisteredClaims{
      Issuer:    s.issuer,
      Subject:   user.Username,
      ExpiresAt: jwt.NewNumericDate(expiresAt),
      IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  signed, err := token.SignedString(s.secret)
  if err != nil {
    return "", time.Time{}, err
  }
  return signed, expiresAt, nil
}

// ParseToken parses and validates a JWT.
// Args:
//   tokenStr: JWT string.
// Returns:
//   *AuthClaims: Parsed claims.
//   error: Error when token is invalid.
func (s *AuthService) ParseToken(tokenStr string) (*AuthClaims, error) {
  if tokenStr == "" {
    return nil, errors.New("token is required")
  }

  parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
  token, err := parser.ParseWithClaims(tokenStr, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
    return s.secret, nil
  })
  if err != nil {
    return nil, err
  }
  claims, ok := token.Claims.(*AuthClaims)
  if !ok || !token.Valid {
    return nil, errors.New("invalid token")
  }
  if claims.Issuer != s.issuer {
    return nil, errors.New("invalid issuer")
  }
  return claims, nil
}
