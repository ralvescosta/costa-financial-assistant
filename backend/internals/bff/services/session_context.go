package services

import (
	"context"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

func projectContextFromContext(ctx context.Context, projectID, fallbackUserID string) *commonv1.ProjectContext {
	projectCtx := &commonv1.ProjectContext{ProjectId: projectID}
	if claims := bffmiddleware.ClaimsFromContext(ctx); claims != nil {
		projectCtx.UserId = claims.GetSubject()
		projectCtx.Role = claims.GetRole()
		return projectCtx
	}
	projectCtx.UserId = fallbackUserID
	return projectCtx
}

func sessionFromContext(ctx context.Context) *commonv1.Session {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil
	}
	return &commonv1.Session{
		Id:       claims.GetSubject(),
		Email:    claims.GetEmail(),
		Username: claims.GetUsername(),
	}
}

func defaultPagination(pageSize int32, pageToken string, fallback int32) *commonv1.Pagination {
	if pageSize <= 0 {
		pageSize = fallback
	}
	return &commonv1.Pagination{PageSize: pageSize, PageToken: pageToken}
}
