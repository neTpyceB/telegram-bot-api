package tgbotapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// SendProtectedAudio sends an audio file to a chat with content protection to prevent forwarding.
func (bot *BotAPI) SendProtectedAudio(chatID int64, audioPath string) (Message, error) {
	// API URL for sending audio
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendAudio", bot.Token)

	// Open the audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to open audio file: %v", err)
	}
	defer file.Close()

	// Create form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the chat_id and protect_content field
	_ = writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	_ = writer.WriteField("protect_content", "true") // Prevent forwarding

	// Create the form file part for the audio
	part, err := writer.CreateFormFile("audio", filepath.Base(audioPath))
	if err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to create form file: %v", err)
	}

	// Copy the file contents into the form
	_, err = io.Copy(part, file)
	if err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to copy audio file: %v", err)
	}

	// Close the writer (this is important for multipart encoding)
	writer.Close()

	// Send the request to the Telegram API
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to create request: %v", err)
	}

	// Set content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to send audio message: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response into an APIResponse
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to decode response: %v", err)
	}

	// Check if the API response is successful
	if !apiResp.Ok {
		return Message{}, fmt.Errorf("[ERROR] Telegram API error: %s", apiResp.Description)
	}

	// Return the sent message from the API response
	var message Message
	if err := json.Unmarshal(apiResp.Result, &message); err != nil {
		return Message{}, fmt.Errorf("[ERROR] Failed to unmarshal result: %v", err)
	}

	return message, nil
}
