package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// LLMMessage là cấu trúc tin nhắn gửi đến API
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AdvancedLLMClient là client nâng cao để giao tiếp với các mô hình ngôn ngữ lớn
type AdvancedLLMClient struct {
	ApiKey      string
	ApiEndpoint string
	Model       string
	Functions   []FunctionDefinition
}

// FunctionDefinition định nghĩa một function có thể gọi bởi LLM
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// AdvancedLLMRequest là cấu trúc yêu cầu gửi đến API với hỗ trợ function calling
type AdvancedLLMRequest struct {
	Model       string                   `json:"model"`
	Messages    []LLMMessage             `json:"messages"`
	Temperature float64                  `json:"temperature"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Tools       []map[string]interface{} `json:"tools,omitempty"`
	ToolChoice  interface{}              `json:"tool_choice,omitempty"`
}

// AdvancedLLMResponse là cấu trúc phản hồi từ API với hỗ trợ function calling
type AdvancedLLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
}

// NewAdvancedLLMClient tạo một client nâng cao mới
func NewAdvancedLLMClient() *AdvancedLLMClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "your-api-key" // Thay thế bằng API key thực tế hoặc sử dụng biến môi trường
	}
	
	// Định nghĩa các function có thể gọi
	budgetOptimizationFunction := FunctionDefinition{
		Name:        "optimize_budget",
		Description: "Tối ưu hóa ngân sách dựa trên mẫu chi tiêu và mục tiêu tài chính",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category_recommendations": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"category": map[string]interface{}{
								"type": "string",
								"description": "Tên danh mục chi tiêu",
							},
							"current_budget": map[string]interface{}{
								"type": "integer",
								"description": "Ngân sách hiện tại cho danh mục này",
							},
							"recommended_budget": map[string]interface{}{
								"type": "integer",
								"description": "Ngân sách được đề xuất cho danh mục này",
							},
							"reason": map[string]interface{}{
								"type": "string",
								"description": "Lý do đề xuất thay đổi ngân sách",
							},
						},
						"required": []string{"category", "current_budget", "recommended_budget", "reason"},
					},
				},
				"saving_goals": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type": "string",
								"description": "Tiêu đề mục tiêu tiết kiệm",
							},
							"amount": map[string]interface{}{
								"type": "integer",
								"description": "Số tiền mục tiêu",
							},
							"duration": map[string]interface{}{
								"type": "string",
								"description": "Thời gian để đạt mục tiêu (ví dụ: '3 tháng')",
							},
							"description": map[string]interface{}{
								"type": "string",
								"description": "Mô tả chi tiết về mục tiêu tiết kiệm",
							},
						},
						"required": []string{"title", "amount", "duration", "description"},
					},
				},
			},
			"required": []string{"category_recommendations", "saving_goals"},
		},
	}
	
	savingRecommendationFunction := FunctionDefinition{
		Name:        "recommend_savings",
		Description: "Đề xuất các cách tiết kiệm chi tiêu dựa trên phân tích dữ liệu chi tiêu",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"recommendations": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type": "string",
								"description": "Tiêu đề ngắn gọn của đề xuất",
							},
							"description": map[string]interface{}{
								"type": "string",
								"description": "Mô tả chi tiết về cách thực hiện đề xuất",
							},
							"saving_amount": map[string]interface{}{
								"type": "integer",
								"description": "Ước tính số tiền có thể tiết kiệm được (VND)",
							},
							"difficulty": map[string]interface{}{
								"type": "string",
								"enum": []string{"dễ", "trung bình", "khó"},
								"description": "Mức độ khó thực hiện đề xuất",
							},
							"category": map[string]interface{}{
								"type": "string",
								"description": "Danh mục chi tiêu liên quan đến đề xuất",
							},
							"timeframe": map[string]interface{}{
								"type": "string",
								"description": "Khung thời gian để thấy kết quả (ví dụ: 'ngay lập tức', 'trong 1 tháng')",
							},
							"impact_level": map[string]interface{}{
								"type": "string",
								"enum": []string{"thấp", "trung bình", "cao"},
								"description": "Mức độ tác động đến tài chính tổng thể",
							},
						},
						"required": []string{"title", "description", "saving_amount", "difficulty", "category", "timeframe", "impact_level"},
					},
				},
			},
			"required": []string{"recommendations"},
		},
	}
	
	financialInsightFunction := FunctionDefinition{
		Name:        "analyze_financial_health",
		Description: "Phân tích sức khỏe tài chính và đưa ra các nhận định sâu sắc",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"financial_health_score": map[string]interface{}{
					"type": "integer",
					"description": "Điểm sức khỏe tài chính (0-100)",
				},
				"spending_income_ratio": map[string]interface{}{
					"type": "number",
					"description": "Tỷ lệ chi tiêu/thu nhập",
				},
				"key_insights": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type": "string",
								"description": "Tiêu đề nhận định",
							},
							"description": map[string]interface{}{
								"type": "string",
								"description": "Mô tả chi tiết về nhận định",
							},
							"impact": map[string]interface{}{
								"type": "string",
								"enum": []string{"cao", "trung bình", "thấp"},
								"description": "Mức độ tác động của nhận định",
							},
							"trend": map[string]interface{}{
								"type": "string",
								"enum": []string{"tăng", "giảm", "ổn định"},
								"description": "Xu hướng của chỉ số này",
							},
						},
						"required": []string{"title", "description", "impact", "trend"},
					},
				},
				"recommendations": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type": "string",
								"description": "Tiêu đề đề xuất",
							},
							"description": map[string]interface{}{
								"type": "string",
								"description": "Mô tả chi tiết về đề xuất",
							},
							"priority": map[string]interface{}{
								"type": "string",
								"enum": []string{"cao", "trung bình", "thấp"},
								"description": "Mức độ ưu tiên của đề xuất",
							},
							"expected_impact": map[string]interface{}{
								"type": "string",
								"description": "Tác động dự kiến khi thực hiện đề xuất",
							},
						},
						"required": []string{"title", "description", "priority", "expected_impact"},
					},
				},
				"risk_factors": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"factor": map[string]interface{}{
								"type": "string",
								"description": "Yếu tố rủi ro",
							},
							"severity": map[string]interface{}{
								"type": "string",
								"enum": []string{"cao", "trung bình", "thấp"},
								"description": "Mức độ nghiêm trọng của rủi ro",
							},
							"mitigation": map[string]interface{}{
								"type": "string",
								"description": "Cách giảm thiểu rủi ro",
							},
						},
						"required": []string{"factor", "severity", "mitigation"},
					},
				},
			},
			"required": []string{"financial_health_score", "spending_income_ratio", "key_insights", "recommendations", "risk_factors"},
		},
	}
	
	return &AdvancedLLMClient{
		ApiKey:      apiKey,
		ApiEndpoint: "https://api.openai.com/v1/chat/completions",
		Model:       "gpt-4o",
		Functions:   []FunctionDefinition{budgetOptimizationFunction, savingRecommendationFunction, financialInsightFunction},
	}
}

// GetSmartSavingRecommendation tạo đề xuất tiết kiệm thông minh với function calling
func (c *AdvancedLLMClient) GetSmartSavingRecommendation(userExpenseData string, userProfile string) (string, error) {
	prompt := fmt.Sprintf(`
Dưới đây là dữ liệu chi tiêu của người dùng trong 3 tháng qua:

%s

Và đây là thông tin hồ sơ người dùng:

%s

Dựa trên dữ liệu này, hãy phân tích và đưa ra các đề xuất cụ thể để giúp người dùng tiết kiệm chi tiêu.
Hãy sử dụng function recommend_savings để trả về kết quả.
`, userExpenseData, userProfile)

	// Chuyển đổi các function thành định dạng tools cho API
	tools := make([]map[string]interface{}, 0)
	for _, fn := range c.Functions {
		if fn.Name == "recommend_savings" {
			tools = append(tools, map[string]interface{}{
				"type": "function",
				"function": fn,
			})
		}
	}

	// Tạo tool choice với kiểu dữ liệu đúng
	toolChoice := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name": "recommend_savings",
		},
	}

	request := AdvancedLLMRequest{
		Model: c.Model,
		Messages: []LLMMessage{
			{
				Role:    "system",
				Content: "Bạn là một trợ lý tài chính thông minh, giúp phân tích chi tiêu và đưa ra đề xuất tiết kiệm cá nhân hóa. Hãy đưa ra các đề xuất cụ thể, thực tế và có thể thực hiện được.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		Tools:       tools,
		ToolChoice:  toolChoice,
	}

	return c.sendRequest(request)
}

// GetAdvancedBudgetOptimization tạo đề xuất tối ưu hóa ngân sách nâng cao
func (c *AdvancedLLMClient) GetAdvancedBudgetOptimization(userBudgetData string, userExpenseData string, userGoals string) (string, error) {
	prompt := fmt.Sprintf(`
Dưới đây là dữ liệu ngân sách hiện tại của người dùng:

%s

Đây là dữ liệu chi tiêu thực tế trong 3 tháng qua:

%s

Và đây là các mục tiêu tài chính của người dùng:

%s

Dựa trên dữ liệu này, hãy phân tích và đưa ra đề xuất để tối ưu hóa ngân sách của người dùng.
Hãy sử dụng function optimize_budget để trả về kết quả.
`, userBudgetData, userExpenseData, userGoals)

	// Chuyển đổi các function thành định dạng tools cho API
	tools := make([]map[string]interface{}, 0)
	for _, fn := range c.Functions {
		if fn.Name == "optimize_budget" {
			tools = append(tools, map[string]interface{}{
				"type": "function",
				"function": fn,
			})
		}
	}

	// Tạo tool choice với kiểu dữ liệu đúng
	toolChoice := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name": "optimize_budget",
		},
	}

	request := AdvancedLLMRequest{
		Model: c.Model,
		Messages: []LLMMessage{
			{
				Role:    "system",
				Content: "Bạn là một chuyên gia tài chính cá nhân, giúp tối ưu hóa ngân sách và đề xuất mục tiêu tiết kiệm phù hợp. Hãy cân nhắc cả nhu cầu chi tiêu hiện tại và mục tiêu tài chính dài hạn.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		Tools:       tools,
		ToolChoice:  toolChoice,
	}

	return c.sendRequest(request)
}

// GetComprehensiveFinancialInsights phân tích sâu về tình hình tài chính với nhiều yếu tố
func (c *AdvancedLLMClient) GetComprehensiveFinancialInsights(userFinancialData string, marketTrends string) (string, error) {
	prompt := fmt.Sprintf(`
Dưới đây là dữ liệu tài chính tổng hợp của người dùng:

%s

Và đây là thông tin về xu hướng thị trường hiện tại:

%s

Dựa trên dữ liệu này, hãy phân tích sâu và đưa ra những nhận định về tình hình tài chính của người dùng.
Hãy sử dụng function analyze_financial_health để trả về kết quả.
`, userFinancialData, marketTrends)

	// Chuyển đổi các function thành định dạng tools cho API
	tools := make([]map[string]interface{}, 0)
	for _, fn := range c.Functions {
		if fn.Name == "analyze_financial_health" {
			tools = append(tools, map[string]interface{}{
				"type": "function",
				"function": fn,
			})
		}
	}

	// Tạo tool choice với kiểu dữ liệu đúng
	toolChoice := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name": "analyze_financial_health",
		},
	}

	request := AdvancedLLMRequest{
		Model: c.Model,
		Messages: []LLMMessage{
			{
				Role:    "system",
				Content: "Bạn là một chuyên gia phân tích tài chính cá nhân, giúp người dùng hiểu rõ tình hình tài chính và đưa ra các đề xuất cải thiện. Hãy cân nhắc cả yếu tố cá nhân và xu hướng thị trường.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		Tools:       tools,
		ToolChoice:  toolChoice,
	}

	return c.sendRequest(request)
}

// sendRequest gửi yêu cầu đến API và xử lý phản hồi
func (c *AdvancedLLMClient) sendRequest(request AdvancedLLMRequest) (string, error) {
	// Chuyển đổi request thành JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("lỗi khi chuyển đổi request thành JSON: %v", err)
	}

	// Tạo HTTP request
	req, err := http.NewRequest("POST", c.ApiEndpoint, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", fmt.Errorf("lỗi khi tạo HTTP request: %v", err)
	}

	// Thêm headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	// Gửi request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lỗi khi gửi request: %v", err)
	}
	defer resp.Body.Close()

	// Đọc phản hồi
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("lỗi khi đọc phản hồi: %v", err)
	}

	// Kiểm tra mã trạng thái
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API trả về lỗi: %s", string(body))
	}

	// Phân tích phản hồi
	var response AdvancedLLMResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("lỗi khi phân tích phản hồi: %v", err)
	}

	// Kiểm tra xem có phản hồi không
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("không có phản hồi từ API")
	}

	// Kiểm tra xem có tool calls không
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		// Lấy kết quả từ function call
		functionCall := response.Choices[0].Message.ToolCalls[0].Function
		return functionCall.Arguments, nil
	}

	// Nếu không có tool calls, trả về nội dung
	return response.Choices[0].Message.Content, nil
}

// ExtractJSONFromResponse trích xuất phần JSON từ phản hồi
func ExtractJSONFromResponse(response string) string {
	// Tìm vị trí bắt đầu của JSON (dấu {)
	startIndex := strings.Index(response, "{")
	if startIndex == -1 {
		return ""
	}

	// Tìm vị trí kết thúc của JSON (dấu })
	endIndex := strings.LastIndex(response, "}")
	if endIndex == -1 || endIndex < startIndex {
		return ""
	}

	// Trích xuất phần JSON
	jsonStr := response[startIndex : endIndex+1]

	// Kiểm tra xem chuỗi có phải là JSON hợp lệ không
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
		return ""
	}

	return jsonStr
}