package common

import (
	"fmt"
	"net/smtp"
	"os"

	// 阿里云短信依赖
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// 邮箱发送验证码
func sendEmail(to string, code string) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	// 邮件服务器域名 邮箱提供商决定
	smtpHost := os.Getenv("SMTP_HOST")
	// 由采用的加密通信协议决定
	smtpPort := os.Getenv("SMTP_PORT")
	// Subject 告诉邮件服务器和客户端邮箱“这是邮件的标题”
	// \r\n 换行符 邮件解析器依靠第一个连续的空行（\r\n\r\n）来界定
	// 空行前面的内容全部当作元数据（标题、格式等），空行后面的内容全部当作邮件正文
	body := fmt.Sprintf("Subject: 验证码\r\n\r\n您的验证码是：%s，5分钟内有效。", code)
	// PLAiN 机制身份认证器
	// 第一个参数身份标识 表示以登录用户的身份发送 通常为空
	auth := smtp.PlainAuth("", from, password, smtpHost)

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(body))
}

// 短信发送验证码
func sendSMS(phone string, code string) error {
	// 配置客户端
	config := &openapi.Config{
		AccessKeyId:     tea.String(os.Getenv("ALIYUN_AK_ID")),
		AccessKeySecret: tea.String(os.Getenv("ALIYUN_AK_SECRET")),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}
	client, err := dysmsapi.NewClient(config)
	if err != nil {
		return err
	}
	// 组装请求体
	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String("你的签名"),          // 需在阿里云控制台申请
		TemplateCode:  tea.String("SMS_XXXXXXXXX"), // 需在阿里云控制台申请
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s"}`, code)),
	}
	// 发起调用并拦截
	resp, err := client.SendSmsWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return err
	}
	if *resp.Body.Code != "OK" {
		return fmt.Errorf("短信发送失败: %s", *resp.Body.Message)
	}
	return nil
}
