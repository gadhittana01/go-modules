package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type MyCustomClaims struct {
	UserID string `json:"userId"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type GenerateTokenReq struct {
	UserID string `json:"user_id"`
}

type DecodeTokenReq struct {
	Token string `json:"token"`
}

type DecodeTokenRes struct {
	UserID string  `json:"user_id"`
	Type   string  `json:"type"`
	IAT    float64 `json:"iat"`
	EXP    float64 `json:"exp"`
}

type GenerateTokenResp struct {
	Token    string `json:"token"`
	ExpToken int64  `json:"expToken"`
}

type TokenClient interface {
	GenerateToken(req GenerateTokenReq) (GenerateTokenResp, error)
	DecodeToken(req DecodeTokenReq) (DecodeTokenRes, error)
}

type TokenRepositoryImpl struct {
	config *BaseConfig
}

func NewToken(config *BaseConfig) TokenClient {
	return &TokenRepositoryImpl{
		config: config,
	}
}

func (r *TokenRepositoryImpl) GenerateToken(req GenerateTokenReq) (GenerateTokenResp, error) {
	var resp GenerateTokenResp
	tokenType := "accessToken"

	exp := jwt.NewNumericDate(time.Now().Add(r.config.AccessTokenDuration))
	iss := jwt.NewNumericDate(time.Now())

	// for payload
	claims := MyCustomClaims{
		UserID: req.UserID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: exp,
			IssuedAt:  iss,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	res, err := token.SignedString([]byte(r.config.JWTKey))
	if err != nil {
		return resp, err
	}

	return GenerateTokenResp{
		Token:    res,
		ExpToken: exp.Unix(),
	}, nil
}

func (r *TokenRepositoryImpl) DecodeToken(req DecodeTokenReq) (DecodeTokenRes, error) {
	var res DecodeTokenRes

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method invalid")
		} else if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("signing method invalid")
		}
		return []byte(r.config.JWTKey), nil
	})
	if err != nil {
		return res, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return res, err
	}

	res = DecodeTokenRes{
		UserID: claims["userId"].(string),
		Type:   claims["type"].(string),
		IAT:    claims["iat"].(float64),
		EXP:    claims["exp"].(float64),
	}

	return res, nil
}
