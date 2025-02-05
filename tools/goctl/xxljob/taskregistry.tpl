package tasks

import (
	"context"
	"{{.projectPackage}}/internal/svc"

	"github.com/shippomx/zard/job/xxljob"
)

type TaskFunc func(ctx context.Context, param *xxljob.TaskRequest, svcCtx *svc.ServiceContext) (msg string)

func wrapTask(f TaskFunc, svcCtx *svc.ServiceContext) xxljob.TaskFunc {
	return func(ctx context.Context, param *xxljob.TaskRequest) (msg string) {
		return f(ctx, param, svcCtx)
	}
}

// RegisterTasks 注册所有XXL-JOB任务.
func RegisterTasks(server *xxljob.Server, svcCtx *svc.ServiceContext) {
	server.RegisterTask("sendEmail", wrapTask(SendEmail, svcCtx))
	// 在这里注册其他任务.
}
