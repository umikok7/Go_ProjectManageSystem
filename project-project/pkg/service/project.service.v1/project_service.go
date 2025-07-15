package project_service_v1

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"test.com/project-grpc/user/login"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/domain"
	"test.com/project-project/internal/rpc"
	"time"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-common/tms"
	"test.com/project-grpc/project"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/repo"
	"test.com/project-project/pkg/model"
)

type ProjectService struct {
	project.UnimplementedProjectServiceServer
	cache                  repo.Cache
	transaction            tran.Transaction
	menuRepo               repo.MenuRepo
	projectRepo            repo.ProjectRepo
	projectTemplateRepo    repo.ProjectTemplateRepo
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
	projectLogRepo         repo.ProjectLogRepo
	taskRepo               repo.TaskRepo
	nodeDomain             *domain.ProjectNodeDomain
	taskDomain             *domain.TaskDomain
}

func New() *ProjectService {
	return &ProjectService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransaction(),
		menuRepo:               dao.NewMenuDao(),
		projectRepo:            dao.NewProjectDao(),
		projectTemplateRepo:    dao.NewProjectTemplateDao(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		taskRepo:               dao.NewTaskDao(),
		nodeDomain:             domain.NewProjectNodeDomain(),
		taskDomain:             domain.NewTaskDomain(),
	}
}

func (p *ProjectService) Index(context.Context, *project.IndexMessage) (*project.IndexResponse, error) {
	pms, err := p.menuRepo.FindMenus(context.Background())
	if err != nil {
		zap.L().Error("index db FindMenus error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	childs := data.CovertChild(pms)
	var mms []*project.MenuMessage
	copier.Copy(&mms, childs)
	return &project.IndexResponse{
		Menus: mms,
	}, nil
}

func (p *ProjectService) FindProjectByMemId(ctx context.Context, msg *project.ProjectRpcMessage) (*project.MyProjectResponse, error) {
	memberId := msg.MemberId
	page := msg.Page
	pageSize := msg.PageSize
	var pms []*data.ProjectAndMember
	var total int64
	var err error
	if msg.SelectBy == "" || msg.SelectBy == "my" {
		pms, total, err = p.projectRepo.FindProjectByMemId(ctx, memberId, "and deleted = 0", page, pageSize)
	}
	if msg.SelectBy == "archive" {
		pms, total, err = p.projectRepo.FindProjectByMemId(ctx, memberId, "and archive = 1", page, pageSize)
	}
	if msg.SelectBy == "deleted" {
		pms, total, err = p.projectRepo.FindProjectByMemId(ctx, memberId, "and deleted = 1", page, pageSize)
	}
	if msg.SelectBy == "collect" {
		pms, total, err = p.projectRepo.FindCollectByMemId(ctx, memberId, page, pageSize)
		for _, v := range pms {
			v.Collected = model.Collected
		}
	} else {
		// 对收藏的状态进行修改
		collectPms, _, err := p.projectRepo.FindCollectByMemId(ctx, memberId, page, pageSize)
		if err != nil {
			zap.L().Error("project FindProjectByMemId::FindCollectByMemId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		var cMap = make(map[int64]*data.ProjectAndMember)
		for _, v := range collectPms {
			cMap[v.Id] = v
		}
		for _, v := range pms {
			if cMap[v.ProjectCode] != nil {
				v.Collected = model.Collected
			}
		}
	}
	if err != nil {
		zap.L().Error("project FindProjectByMemID error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return &project.MyProjectResponse{
			Pm:    []*project.ProjectMessage{},
			Total: total,
		}, nil
	}
	var pmm []*project.ProjectMessage
	copier.Copy(&pmm, pms)
	for _, v := range pmm {
		v.Code, _ = encrypts.EncryptInt64(v.ProjectCode, model.AESKey)
		pam := data.ToMap(pms)[v.Id]
		v.AccessControlType = pam.GetAccessControlType()
		v.OrganizationCode, _ = encrypts.EncryptInt64(pam.OrganizationCode, model.AESKey)
		v.JoinTime = tms.FormatByMill(pam.JoinTime)
		v.OwnerName = msg.MemberName
		v.Order = int32(pam.Sort)
		v.CreateTime = tms.FormatByMill(pam.CreateTime)
	}
	return &project.MyProjectResponse{
		Pm:    pmm,
		Total: total,
	}, nil
}

func (ps *ProjectService) FindProjectTemplate(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectTemplateResponse, error) {
	// 1. 根据viewType查询项目模版表 得到list
	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	page := msg.Page
	pageSize := msg.PageSize
	var pts []data.ProjectTemplate
	var total int64
	var err error
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if msg.ViewType == -1 {
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateAll(c, organizationCode, page, pageSize)
	}
	if msg.ViewType == 0 {
		// 查询用户自定义模板
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateCustom(c, msg.MemberId, organizationCode, page, pageSize)
	}
	if msg.ViewType == 1 {
		// 查询系统内置模板
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateSystem(c, page, pageSize)
	}
	if err != nil {
		zap.L().Error("project FindProjectTemplate error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 2. 模型转换，拿到模版id列表，得到任务步骤模版表去进行查询
	tsts, err := ps.taskStagesTemplateRepo.FindInProTemIds(c, data.ToProjectTemplateIds(pts))
	if err != nil {
		zap.L().Error("project FindInProTemIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var ptas []*data.ProjectTemplateAll
	for _, v := range pts {
		ptas = append(ptas, v.Convert(data.CovertProjectMap(tsts)[v.Id]))
	}
	// 3. 组装数据
	var pmMsgs []*project.ProjectTemplateMessage
	copier.Copy(&pmMsgs, ptas)
	return &project.ProjectTemplateResponse{
		Ptm: pmMsgs, Total: total,
	}, nil
}

func (ps *ProjectService) SaveProject(ctxs context.Context, msg *project.ProjectRpcMessage) (*project.SaveProjectMessage, error) {
	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	templateCodeStr, _ := encrypts.Decrypt(msg.TemplateCode, model.AESKey)
	templateCode, _ := strconv.ParseInt(templateCodeStr, 10, 64)
	// 获取模版信息
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stageTemplateList, err := ps.taskStagesTemplateRepo.FindByProjectTemplateId(ctx, int(templateCode))
	if err != nil {
		zap.L().Error("project FindByProjectTemplateId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 1. 保存项目表
	pr := &data.Project{
		Name:              msg.Name,
		Description:       msg.Description,
		TemplateCode:      int(templateCode),
		CreateTime:        time.Now().UnixMilli(),
		Cover:             "https://img2.baidu.com/it/u=792555388,2449797505&fm=253&fmt=auto&app=138&f=JPEG?w=667&h=500",
		Deleted:           model.NoDeleted,
		Archive:           model.NoArchive,
		OrganizationCode:  organizationCode,
		AccessControlType: model.Open,
		TaskBoardTheme:    model.Simple,
	}
	err = ps.transaction.Action(func(conn database.DbConn) error {
		err := ps.projectRepo.SaveProject(conn, ctx, pr)
		if err != nil {
			zap.L().Error("project saveProject error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 2. 保存项目和成员的关联表
	pm := &data.ProjectMember{
		ProjectCode: pr.Id,
		MemberCode:  msg.MemberId,
		JoinTime:    time.Now().UnixMilli(),
		IsOwner:     msg.MemberId,
		Authorize:   "",
	}
	err = ps.transaction.Action(func(conn database.DbConn) error {
		err := ps.projectRepo.SaveProjectMember(conn, ctx, pm)
		if err != nil {
			zap.L().Error("project saveProjectMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		// 3. 生成任务看板的任务步骤
		for index, v := range stageTemplateList {
			taskStage := &data.TaskStages{
				ProjectCode: pr.Id,
				Name:        v.Name,
				Sort:        index + 1,
				Description: "",
				CreateTime:  time.Now().UnixMilli(),
				Deleted:     model.NoDeleted,
			}
			err := ps.taskStagesRepo.SaveTaskStages(ctx, conn, taskStage)
			if err != nil {
				zap.L().Error("project SaveTaskStages error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	code, _ := encrypts.EncryptInt64(pr.Id, model.AESKey)
	rsp := &project.SaveProjectMessage{
		Id:               pr.Id,
		Code:             code,
		OrganizationCode: organizationCodeStr,
		Name:             pr.Name,
		Cover:            pr.Cover,
		CreateTime:       tms.FormatByMill(pr.CreateTime),
		TaskBoardTheme:   pr.TaskBoardTheme,
	}
	return rsp, nil

}

func (ps *ProjectService) FindProjectDetail(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectDetailMessage, error) {
	// 1. 查项目表
	// 2. 查项目和成员的关联表，查到项目的拥有者，去member表查名字
	// 3. 查收藏表，判断收藏状态
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64) // 39
	memberId := msg.MemberId                                   // 1003
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	fmt.Printf("projectCode = %d, memberId = %d", projectCode, memberId)
	projectAndMember, err := ps.projectRepo.FindProjectByPIdAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project rpc FindProjectDetail FindProjectByPIdAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if projectAndMember == nil {
		return nil, errs.GrpcError(model.ParamsError)
	}
	ownerId := projectAndMember.IsOwner
	// 作为客户端调用User模块中的方法
	member, err := rpc.LoginServiceClient.FindMemInfoById(c, &login.UserMessage{MemId: ownerId})
	if err != nil {
		zap.L().Error("project rpc FindProjectDetail FindMemInfoById error", zap.Error(err))
		return nil, err
	}
	// 去user模块去找
	log.Println(ownerId)
	// TODD 优化 收藏的时候放入redis
	isCollected, err := ps.projectRepo.FindCollectByPIdAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindCollectByPIdAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if isCollected {
		projectAndMember.Collected = model.Collected
	}
	var detaiMsg = &project.ProjectDetailMessage{}
	copier.Copy(&detaiMsg, projectAndMember)
	detaiMsg.OwnerAvatar = member.Avatar
	detaiMsg.Name = member.Name
	detaiMsg.Code, _ = encrypts.EncryptInt64(projectAndMember.ProjectCode, model.AESKey)
	detaiMsg.AccessControlType = projectAndMember.GetAccessControlType()
	detaiMsg.OrganizationCode, _ = encrypts.EncryptInt64(projectAndMember.OrganizationCode, model.AESKey)
	detaiMsg.Order = int32(projectAndMember.Sort)
	detaiMsg.CreateTime = tms.FormatByMill(projectAndMember.CreateTime)
	return detaiMsg, nil
}

func (ps *ProjectService) UpdateDeleteProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.DeleteProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := ps.projectRepo.UpdateDeleteProject(c, projectCode, msg.Deleted)
	if err != nil {
		zap.L().Error("project RecycleProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project.DeleteProjectResponse{}, nil
}

func (ps *ProjectService) UpdateProject(ctx context.Context, msg *project.UpdateProjectMessage) (*project.UpdateProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	proj := &data.Project{
		Id:                 projectCode,
		Name:               msg.Name,
		Description:        msg.Description,
		Cover:              msg.Cover,
		TaskBoardTheme:     msg.TaskBoardTheme,
		Prefix:             msg.Prefix,
		Private:            int(msg.Private),
		OpenPrefix:         int(msg.OpenPrefix),
		OpenBeginTime:      int(msg.OpenBeginTime),
		OpenTaskPrivate:    int(msg.OpenTaskPrivate),
		Schedule:           msg.Schedule,
		AutoUpdateSchedule: int(msg.AutoUpdateSchedule),
	}
	err := ps.projectRepo.UpdateProject(c, proj)
	if err != nil {
		zap.L().Error("project UpdateProject::UpdateProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project.UpdateProjectResponse{}, nil
}

func (ps *ProjectService) GetLogBySelfProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectLogResponse, error) {
	// 根据用户id查询当前用户的日志表
	projectLogs, total, err := ps.projectLogRepo.FindLogByMemberCode(context.Background(), msg.MemberId, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project GetLogBySelfProject GetLogBySelfProject::FindLogByMemberCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 查询项目信息
	pIdList := make([]int64, len(projectLogs))
	mIdList := make([]int64, len(projectLogs))
	taskIdList := make([]int64, len(projectLogs))
	for _, v := range projectLogs {
		pIdList = append(pIdList, v.ProjectCode)
		mIdList = append(mIdList, v.MemberCode)
		taskIdList = append(taskIdList, v.SourceCode)
	}

	projects, err := ps.projectRepo.FindProjectByIds(context.Background(), pIdList)
	if err != nil {
		zap.L().Error("project GetLogBySelfProject projectRepo::FindProjectByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pMap := make(map[int64]*data.Project)
	for _, v := range projects {
		pMap[v.Id] = v
	}

	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(context.Background(), &login.UserMessage{MIds: mIdList})
	if err != nil {
		zap.L().Error("project GetLogBySelfProject LoginServiceClient::FindMemInfoByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}

	tasks, err := ps.taskRepo.FindTaskByIds(context.Background(), taskIdList)
	if err != nil {
		zap.L().Error("project GetLogBySelfProject taskRepo::FindTaskByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	tMap := make(map[int64]*data.Task)
	for _, v := range tasks {
		tMap[v.Id] = v
	}

	var list []*data.IndexProjectLogDisplay
	for _, v := range projectLogs {
		display := v.ToIndexDisplay()
		display.ProjectName = pMap[v.ProjectCode].Name
		display.MemberAvatar = mMap[v.MemberCode].Avatar
		display.MemberName = mMap[v.MemberCode].Name
		display.TaskName = tMap[v.SourceCode].Name
		list = append(list, display)
	}
	var msgList []*project.ProjectLogMessage
	copier.Copy(&msgList, list)
	return &project.ProjectLogResponse{List: msgList, Total: total}, nil
}

func (ps *ProjectService) FindProjectByMemberId(ctx context.Context, msg *project.ProjectRpcMessage) (*project.FindProjectByMemberIdResponse, error) {
	isProjectCode := false
	var projectId int64
	if msg.ProjectCode != "" {
		projectId = encrypts.DecryptNoErr(msg.ProjectCode)
		isProjectCode = true
	}
	isTaskCode := false
	var taskId int64
	if msg.TaskCode != "" {
		taskId = encrypts.DecryptNoErr(msg.TaskCode)
		isTaskCode = true
	}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if !isProjectCode && isTaskCode {
		// 如果没有projectCode但是有TaskCode，那么可以通过TaskCode将projectCode查出来
		projectCode, ok, bError := ps.taskDomain.FindProjectIdByTaskId(taskId)
		if bError != nil {
			return nil, bError
		}
		if !ok {
			return &project.FindProjectByMemberIdResponse{
				Project:  nil,
				IsOwner:  false,
				IsMember: false,
			}, nil
		}
		projectId = projectCode
		isProjectCode = true
	}
	if isProjectCode {
		// 那么就根据projectCode和memberId进行查询
		pm, err := ps.projectRepo.FindProjectByPIdAndMemId(c, projectId, msg.MemberId)
		if err != nil {
			return nil, model.DBError
		}
		if pm == nil {
			return &project.FindProjectByMemberIdResponse{
				Project:  nil,
				IsOwner:  false,
				IsMember: false,
			}, nil
		}
		projectMessage := &project.ProjectMessage{}
		copier.Copy(projectMessage, pm)
		isOwner := false
		if pm.IsOwner == 1 {
			isOwner = true
		}
		return &project.FindProjectByMemberIdResponse{
			Project:  projectMessage,
			IsOwner:  isOwner,
			IsMember: true,
		}, nil
	}
	return &project.FindProjectByMemberIdResponse{}, nil
}
