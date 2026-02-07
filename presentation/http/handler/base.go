package handler

import (
	"errors"
	"net/http"

	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/labstack/echo/v4"
)

type BaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func HandleSuccess(c echo.Context, data any) error {
	resp := BaseResp{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
	return c.JSON(http.StatusOK, resp)
}

func HandleError(c echo.Context, err error) error {
	resp := BaseResp{}

	var appErr *errorx.AppError
	if errors.As(err, &appErr) {
		resp.Code = int(appErr.Code)
		resp.Message = appErr.Message

		status := http.StatusInternalServerError
		if appErr.Code < 500 {
			status = int(appErr.Code)
		}
		return c.JSON(status, resp)
	}

	// fallback for unexpected errors
	resp.Code = int(errorx.ErrInternal)
	resp.Message = errorx.GetErrorMessage(int(errorx.ErrInternal))
	return c.JSON(http.StatusInternalServerError, resp)
}
