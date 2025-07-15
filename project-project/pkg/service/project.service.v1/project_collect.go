package project_service_v1

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-grpc/project"
	"test.com/project-project/internal/data"
	"test.com/project-project/pkg/model"
	"time"
)

func (ps *ProjectService) UpdateCollectProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.CollectProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var err error
	if "collect" == msg.CollectType {
		pc := &data.ProjectCollection{
			ProjectCode: projectCode,
			MemberCode:  msg.MemberId,
			CreateTime:  time.Now().UnixMilli(),
		}
		err = ps.projectRepo.SaveProjectCollect(c, pc) // 收藏
	}
	if "cancel" == msg.CollectType {
		err = ps.projectRepo.DeleteProjectCollection(c, msg.MemberId, projectCode) // 取消收藏
	}
	if err != nil {
		zap.L().Error("project SaveProjectCollect error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project.CollectProjectResponse{}, nil
}
