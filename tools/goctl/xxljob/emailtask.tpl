package tasks

import (
	"context"
	"{{.projectPackage}}/internal/svc"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/job/xxljob"
)

func SendEmail(ctx context.Context, param *xxljob.TaskRequest, svcCtx *svc.ServiceContext) (msg string) {
	// 解析任务参数,ctx中注入了traceID,WithContext(ctx)可以在日志中显示traceID.
	logx.WithContext(ctx).Infof("Received sendEmail task with params: %s", param.ExecutorParams)

	// 使用 svcCtx 访问需要的资源.
	// 例如：svcCtx.EmailClient.Send(...).

	logx.WithContext(ctx).Info("Email sent successfully")

	return "Email sent successfully"
}
