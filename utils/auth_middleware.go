package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gadhittana01/go-modules/constant"
)

const (
	HeaderContentType           = "Content-Type"
	HeaderUserAgent             = "User-Agent"
	HeaderAuthorization         = "Authorization"
	HeaderAuthorizationCustomer = "Authorization-Customer"
	HeaderXFAuthorization       = "X-Forwarded-Authorization"
	HeaderXUserID               = "X-User-Id"
	UserSession                 = "user-session"
)

type AuthPayload struct {
	UserID string `json:"userId"`
}

type AuthMiddleware interface {
	CheckIsAuthenticated(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc
}

type AuthMiddlewareImpl struct {
	config *BaseConfig
	token  TokenClient
}

func NewAuthMiddleware(config *BaseConfig, token TokenClient) AuthMiddleware {
	return &AuthMiddlewareImpl{
		config: config,
		token:  token,
	}
}

func (m *AuthMiddlewareImpl) CheckIsAuthenticated(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(HeaderXFAuthorization)
		if header == "" {
			header = r.Header.Get(HeaderAuthorization)
		}

		if header == "" || !strings.Contains(header, "Bearer ") {
			PanicIfError(CustomError("unauthorized", 401))
		}
		authToken := strings.Split(header, " ")[1]

		res, err := m.token.DecodeToken(DecodeTokenReq{
			Token: authToken,
		})
		if err != nil {
			LogInfo(fmt.Sprintf("failed when decode token, error: %v", err))
			PanicIfError(CustomError("unauthorized", 401))
		}

		authCtx := AppendRequestCtx(r, constant.UserSession, &AuthPayload{
			UserID: res.UserID,
		})
		handler(w, r.WithContext(authCtx))
	}
}

func AppendRequestCtx(r *http.Request, key constant.ContextKeyType, input interface{}) context.Context {
	return context.WithValue(r.Context(), constant.UserSession, input)
}

func SetRequestContext(userID string) context.Context {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return AppendRequestCtx(req, constant.UserSession, &AuthPayload{
		UserID: userID,
	})
}

func GetRequestCtx(ctx context.Context, ctxKey constant.ContextKeyType) *AuthPayload {
	ctxVal := ctx.Value(ctxKey)
	if ctxVal != nil {
		return ctxVal.(*AuthPayload)
	}

	return &AuthPayload{}
}
