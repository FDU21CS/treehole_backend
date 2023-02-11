package utils

import (
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
	"treehole_backend/config"
)

func SendCodeEmail(code, receiver string) error {
	credential := common.NewCredential(
		config.Config.TencentSecretID,
		config.Config.TencentSecretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ses.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, err := ses.NewClient(credential, regions.HongKong, cpf)
	if err != nil {
		return err
	}

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := ses.NewSendEmailRequest()

	request.FromEmailAddress = common.StringPtr(config.Config.EmailUrl)
	request.Destination = common.StringPtrs([]string{receiver})
	request.Template = &ses.Template{
		TemplateID:   common.Uint64Ptr(config.Config.TencentTemplateID),
		TemplateData: common.StringPtr(fmt.Sprintf("{\"code\": \"%s\"}", code)),
	}
	request.Subject = common.StringPtr("[MOSS] Verification Code")
	request.TriggerType = common.Uint64Ptr(1)

	// 返回的resp是一个SendEmailResponse的实例，与请求对象对应
	_, err = client.SendEmail(request)
	return err
}
