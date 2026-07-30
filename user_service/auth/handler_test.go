package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	// 设置测试模式 保持测试输出干净
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/register", RegisterHandler)
	// 模拟的无效数据
	reqBody := map[string]string{
		"username": "",
	}
	// 转换成json
	bodyBytes, _ := json.Marshal(reqBody)
	// httptest 不会真的发起请求 bodyBytes是[]byte类型 bytes.NewBuffer将其包装成io.Reader流式传输
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// NewRecorder 充当响应客户端 捕获接口返回的 HTTP 状态码和 Body 内容
	w := httptest.NewRecorder()
	// ServeHTTP 内存中直接运行路由逻辑，无需真实启动网络端口服务 并将结果装入w
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	// Body.Bytes 即内存中的JSON字节流
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	head, ok := resp["head"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected head in response")
	}

	if int(head["code"].(float64)) != http.StatusBadRequest {
		t.Errorf("expected business code %d, got %v", http.StatusBadRequest, head["code"])
	}
}
