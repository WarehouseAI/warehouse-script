package service_errors

import "github.com/warehouse/ai-service/internal/pkg/errors"

var (
	NodeExecError = &errors.Error{Code: 400, Reason: "Can't exec node"}
)
