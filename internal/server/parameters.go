package server

import (
	"context"
	"errors"

	"github.com/Hikyo-Org/hikyo/api/apigen"
	"github.com/Hikyo-Org/hikyo/internal/domain"
	"github.com/Hikyo-Org/hikyo/internal/service"
)

type environmentParameterService interface {
	Parameters(context.Context, service.Actor, domain.Scope) (map[string]string, error)
	SetParameter(context.Context, service.Actor, domain.Scope, string, string, bool) error
}

func (a *API) ListEnvironmentParameters(ctx context.Context, req apigen.ListEnvironmentParametersRequestObject) (apigen.ListEnvironmentParametersResponseObject, error) {
	s, ok := a.Environments.(environmentParameterService)
	if !ok {
		return nil, errors.New("server: environment parameter service is not wired")
	}
	out, err := s.Parameters(ctx, service.Bearer(bearer(ctx)), envScope(req.Org, req.Project, req.Environment))
	if err != nil {
		return nil, err
	}
	return apigen.ListEnvironmentParameters200JSONResponse(out), nil
}

func (a *API) ChangeEnvironmentParameter(ctx context.Context, req apigen.ChangeEnvironmentParameterRequestObject) (apigen.ChangeEnvironmentParameterResponseObject, error) {
	s, ok := a.Environments.(environmentParameterService)
	if !ok {
		return nil, errors.New("server: environment parameter service is not wired")
	}
	pattern := ""
	if req.Body.Pattern != nil {
		pattern = *req.Body.Pattern
	}
	if err := s.SetParameter(ctx, service.Bearer(bearer(ctx)), envScope(req.Org, req.Project, req.Environment), req.Body.Name, pattern, string(req.Body.Action) == "delete"); err != nil {
		return nil, err
	}
	return apigen.ChangeEnvironmentParameter204Response{}, nil
}

type parameterizedExportService interface {
	ExportWithParameters(context.Context, service.Actor, domain.Scope, int64, bool, map[string]string) ([]service.ExportedValue, int64, error)
}
