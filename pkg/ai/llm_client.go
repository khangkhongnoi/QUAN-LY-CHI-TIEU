package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// LLMClient là client để giao tiếp với các mô hình ngôn ngữ lớn
type LLMClient struct {
	ApiKey      string
	ApiEndpoint string
	Model       string
}

// LLMRequest là cấu trúc yêu cầu gửi đến API
type LLMRequest struct {
	Model       string            `json:"model"`
	Messages    []BasicLLMMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
}

// BasicLLMMessage là cấu trúc tin nhắn trong yêu cầu API
type BasicLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse là cấu trúc phản hồi từ API
type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewLLMClient tạo một client LLM mới
func NewLLMClient() *LLMClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "sk-your-api-key" // Thay thế bằng API key thực tế
	}
	
	return &LLMClient{
		ApiKey:      apiKey,
		ApiEndpoint: "https://api.openai.com/v1/chat/completions",
		Model:       "gpt-3.5-turbo",
	}
}

// GetSavingRecommendation lấy đề xuất tiết kiệm từ AI
func (c *LLMClient) GetSavingRecommendation(expenseData string) (string, error) {
	prompt := fmt.Sprintf(`
Dưới đây là dữ liệu chi tiêu của người dùng trong 3 tháng qua:

%s

Dựa trên dữ liệu này, hãy phân tích và đưa ra 3 đề xuất cụ thể để giúp người dùng tiết kiệm chi tiêu. 
Mỗi đề xuất cần bao gồm:
1. Tiêu đề ngắn gọn
2. Mô tả chi tiết về cách thực hiện
3. Ước tính số tiền có thể tiết kiệm được (tính bằng VND)
4. Mức độ khó thực hiện (dễ, trung bình, khó)

Trả về kết quả dưới dạng JSON với định dạng sau:
{
  "recommendations": [
    {
      "title": "Tiêu đề đề xuất",
      "description": "Mô tả chi tiết",
      "saving_amount": 500000,
      "difficulty": "dễ"
    },
    ...
  ]
}`, expenseData)

	request := LLMRequest{
		Model:    c.Model,
		Messages: []BasicLLMMessage{
			{
				Role:    "system",
				Content: "Bạn là một chuyên gia tài chính cá nhân, giúp phân tích chi tiêu và đề xuất cách tiết kiệm.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.ApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var llmResponse LLMResponse
	err = json.Unmarshal(body, &llmResponse)
	if err != nil {
		return "", err
	}

	if len(llmResponse.Choices) == 0 {
		return "", fmt.Errorf("không có phản hồi từ API")
	}

	return llmResponse.Choices[0].Message.Content, nil
}

// GetBudgetOptimization lấy đề xuất tối ưu hóa ngân sách từ AI
func (c *LLMClient) GetBudgetOptimization(budgetData, expenseData string) (string, error) {
	prompt := fmt.Sprintf(`
Dưới đây là dữ liệu ngân sách hiện tại của người dùng:

%s

Và đây là dữ liệu chi tiêu thực tế trong 3 tháng qua:

%s

Dựa trên dữ liệu này, hãy phân tích và đưa ra đề xuất để tối ưu hóa ngân sách của người dùng. 
Đề xuất cần bao gồm:
1. Các danh mục nên tăng/giảm ngân sách
2. Phân bổ ngân sách hợp lý hơn
3. Các mục tiêu tiết kiệm khả thi

Trả về kết quả dưới dạng JSON với định dạng sau:
{
  "budget_recommendations": [
    {
      "category": "Tên danh mục",
      "current_budget": 1000000,
      "recommended_budget": 800000,
      "reason": "Lý do đề xuất"
    },
    ...
  ],
  "saving_goals": [
    {
      "title": "Tiêu đề mục tiêu",
      "amount": 5000000,
      "duration": "3 tháng",
      "description": "Mô tả chi tiết"
    },
    ...
  ]
}`, budgetData, expenseData)

	request := LLMRequest{
		Model:    c.Model,
		Messages: []BasicLLMMessage{
			{
				Role:    "system",
				Content: "Bạn là một chuyên gia tài chính cá nhân, giúp tối ưu hóa ngân sách và đề xuất mục tiêu tiết kiệm phù hợp.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.ApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var llmResponse LLMResponse
	err = json.Unmarshal(body, &llmResponse)
	if err != nil {
		return "", err
	}

	if len(llmResponse.Choices) == 0 {
		return "", fmt.Errorf("không có phản hồi từ API")
	}

	return llmResponse.Choices[0].Message.Content, nil
}

// GetFinancialInsights lấy phân tích tài chính từ AI
func (c *LLMClient) GetFinancialInsights(financialData string) (string, error) {
	prompt := fmt.Sprintf(`
Dưới đây là dữ liệu tài chính của người dùng:

%s

Dựa trên dữ liệu này, hãy phân tích và đưa ra các phân tích tài chính toàn diện. 
Phân tích cần bao gồm:
1. Điểm sức khỏe tài chính (0-100)
2. Tỷ lệ chi tiêu/thu nhập
3. Các phát hiện quan trọng
4. Đề xuất cải thiện

Trả về kết quả dưới dạng JSON với định dạng sau:
{
  "financial_health_score": 75,
  "spending_income_ratio": 0.65,
  "key_insights": [
    {
      "title": "Tiêu đề phát hiện",
      "description": "Mô tả chi tiết",
      "impact": "cao/trung bình/thấp"
    },
    ...
  ],
  "recommendations": [
    {
      "title": "Tiêu đề đề xuất",
      "description": "Mô tả chi tiết",
      "priority": "cao/trung bình/thấp"
    },
    ...
  ]
}`, financialData)

	request := LLMRequest{
		Model:    c.Model,
		Messages: []BasicLLMMessage{
			{
				Role:    "system",
				Content: "Bạn là một chuyên gia tài chính cá nhân, giúp phân tích sức khỏe tài chính và đưa ra đề xuất cải thiện.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.ApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var llmResponse LLMResponse
	err = json.Unmarshal(body, &llmResponse)
	if err != nil {
		return "", err
	}

	if len(llmResponse.Choices) == 0 {
		return "", fmt.Errorf("không có phản hồi từ API")
	}

	return llmResponse.Choices[0].Message.Content, nil
}