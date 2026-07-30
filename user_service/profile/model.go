package profile

// ProfileResponse 个人资料接口响应数据结构
type ProfileResponse struct {
	UserID   any    `json:"user_id"`
	Username any    `json:"username"`
	Remark   string `json:"remark"`
}
