package profile

// ProfileResponse 个人资料接口响应数据结构
type ProfileResponse struct {
	UserID   any    `json:"userId"`
	Username any    `json:"username"`
	Remark   string `json:"remark"`
}
