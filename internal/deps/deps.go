package deps

import (
	"net/http"

	"go.uber.org/zap"
)

type Deps struct {
	Logger     *zap.Logger
	HTTPClient *http.Client
	ErrMsg     string
}

func NewDeps(logger *zap.Logger, httpClient *http.Client, errMsg string) *Deps {
	return &Deps{Logger: logger, HTTPClient: httpClient, ErrMsg: errMsg}
}
