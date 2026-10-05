package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type RequestPayload struct {
	Message string `json:"message"`
}

type ResponsePayload struct {
	Answer string `json:"answer"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	chatHistory []ChatMessage
	historyMu   sync.Mutex
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/ask", handleAsk)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleAsk(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Паника сервера: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ResponsePayload{Answer: "Внутренняя ошибка сервера бэкенда."})
		}
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("YANDEX_API_KEY")
	folderID := os.Getenv("YANDEX_FOLDER_ID")
	if apiKey == "" || folderID == "" {
		http.Error(w, "Yandex API configuration missing", http.StatusInternalServerError)
		return
	}

	answer, err := askSmartYandexGPT(req.Message, apiKey, folderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResponsePayload{Answer: answer})
}

func askSmartYandexGPT(userMessage, apiKey, folderID string) (string, error) {
	baseURL := "https://ai.api.cloud.yandex.net/v1"
	modelURI := fmt.Sprintf("gpt://%s/yandexgpt", folderID)

	historyMu.Lock()
	if len(chatHistory) == 0 {
		todayStr := time.Now().Format("02.01.2006")
		chatHistory = append(chatHistory, ChatMessage{
			Role:    "system",
			Content: "Ты — профессиональный хоккейный эксперт КХЛ и ИИ-консультант. Сегодня строго " + todayStr + ". Отвечай на вопросы пользователя на основе веб-поиска. ВАЖНО: Веди живой диалог. Никогда не повторяй вводную информацию (дату, стадион, общий счет), которую ты уже называл в предыдущих ответах. Сразу переходи к сути нового вопроса пользователя.",
		})
	}
	chatHistory = append(chatHistory, ChatMessage{Role: "user", Content: userMessage})

	var promptBuilder bytes.Buffer
	for _, msg := range chatHistory {
		if msg.Role == "system" {
			promptBuilder.WriteString(msg.Content + "\n\n")
		} else if msg.Role == "user" {
			promptBuilder.WriteString("Вопрос: " + msg.Content + "\n")
		} else if msg.Role == "assistant" {
			promptBuilder.WriteString("Ответ: " + msg.Content + "\n")
		}
	}
	promptBuilder.WriteString("Ответ:")
	inputText := promptBuilder.String()
	historyMu.Unlock()

	payload := map[string]interface{}{
		"model":      modelURI,
		"background": true,
		"input":      inputText,
		"tools": []interface{}{
			map[string]interface{}{
				"type": "web_search",
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Yandex Task Creation Error: %d, Response: %s", resp.StatusCode, string(body))
		return "Ошибка создания задачи в облаке Яндекса.", nil
	}

	var task map[string]interface{}
	if err := json.Unmarshal(body, &task); err != nil {
		return "", err
	}

	taskID, _ := task["id"].(string)
	if taskID == "" {
		return "Ошибка: Облако не вернуло ID задачи.", nil
	}

	var finalAnswer string
	for {
		statusReq, err := http.NewRequest("GET", baseURL+"/responses/"+taskID, nil)
		if err != nil {
			return "", err
		}
		statusReq.Header.Set("Authorization", "Bearer "+apiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			return "", err
		}

		statusBody, err := io.ReadAll(statusResp.Body)
		statusResp.Body.Close()
		if err != nil {
			return "", err
		}

		var currentStatus map[string]interface{}
		if err := json.Unmarshal(statusBody, &currentStatus); err != nil {
			return "", err
		}

		statusStr, _ := currentStatus["status"].(string)
		if statusStr == "completed" {
			finalAnswer = findTextInJSON(currentStatus)
			if finalAnswer == "" {
				finalAnswer = "Текст успешно сгенерирован облаком, но рекурсивный парсер Go не смог извлечь его."
			}
			break
		}
		if statusStr == "failed" || statusStr == "cancelled" {
			return "Задача генерации ответа завершилась с ошибкой внутри облака Яндекса.", nil
		}

		time.Sleep(1 * time.Second)
	}

	historyMu.Lock()
	chatHistory = append(chatHistory, ChatMessage{Role: "assistant", Content: finalAnswer})
	historyMu.Unlock()

	return finalAnswer, nil
}

func findTextInJSON(v interface{}) string {
	switch val := v.(type) {
	case map[string]interface{}:
		if textVal, exists := val["text"]; exists {
			if str, ok := textVal.(string); ok && len(str) > 20 {
				return str
			}
		}
		for _, subVal := range val {
			if res := findTextInJSON(subVal); res != "" {
				return res
			}
		}
	case []interface{}:
		for _, subVal := range val {
			if res := findTextInJSON(subVal); res != "" {
				return res
			}
		}
	}
	return ""
}
