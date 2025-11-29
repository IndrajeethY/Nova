package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

const maxPasteSize = 10 * 1024 * 1024

type GistFile struct {
	Content string `json:"content"`
}

type GistRequest struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

type GistResponse struct {
	HTMLURL string `json:"html_url"`
	ID      string `json:"id"`
}

func pasteToNekobin(content string) (string, error) {
	payload := map[string]string{"content": content}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://nekobin.com/api/documents", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			Key string `json:"key"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.Ok || result.Result.Key == "" {
		return "", fmt.Errorf("nekobin paste failed")
	}

	return fmt.Sprintf("https://nekobin.com/%s", result.Result.Key), nil
}

func pasteToSpacebin(content string) (string, error) {
	payload := map[string]string{
		"content":   content,
		"extension": "txt",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://spaceb.in/api/v1/documents", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Payload struct {
			ID string `json:"id"`
		} `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Payload.ID == "" {
		return "", fmt.Errorf("spacebin paste failed")
	}

	return fmt.Sprintf("https://spaceb.in/%s", result.Payload.ID), nil
}

func pasteToPasteSr(content string) (string, error) {
	payload := map[string]interface{}{
		"contents": content,
		"filename": "paste.txt",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://paste.sr.ht/api/pastes", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		resultURL := strings.TrimSpace(string(body))
		if strings.HasPrefix(resultURL, "http") {
			return resultURL, nil
		}
		return "", err
	}

	if result.URL == "" {
		return "", fmt.Errorf("paste.sr.ht paste failed")
	}

	return result.URL, nil
}

func pasteToDpaste(content string) (string, error) {
	data := url.Values{}
	data.Set("content", content)
	data.Set("syntax", "text")
	data.Set("expiry_days", "30")

	req, err := http.NewRequest("POST", "https://dpaste.org/api/", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("dpaste paste failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	resultURL := strings.TrimSpace(string(body))
	if strings.HasPrefix(resultURL, "http") {
		return resultURL, nil
	}

	return "", fmt.Errorf("dpaste paste failed")
}

func pasteTo0x0(content string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "paste.txt")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, strings.NewReader(content))
	if err != nil {
		return "", err
	}
	writer.Close()

	req, err := http.NewRequest("POST", "https://0x0.st", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("0x0.st paste failed: %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	resultURL := strings.TrimSpace(string(respBody))
	if strings.HasPrefix(resultURL, "http") {
		return resultURL, nil
	}

	return "", fmt.Errorf("0x0.st paste failed")
}

func pasteToPasteRs(content string) (string, error) {
	req, err := http.NewRequest("POST", "https://paste.rs/", strings.NewReader(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("paste.rs paste failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	resultURL := strings.TrimSpace(string(body))
	if strings.HasPrefix(resultURL, "http") {
		return resultURL, nil
	}

	return "", fmt.Errorf("paste.rs paste failed")
}

func pasteToPastesDev(content string) (string, error) {
	req, err := http.NewRequest("POST", "https://api.pastes.dev/post", strings.NewReader(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("pastes.dev paste failed: %d", resp.StatusCode)
	}

	var result struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Key == "" {
		return "", fmt.Errorf("pastes.dev paste failed")
	}

	return fmt.Sprintf("https://pastes.dev/%s", result.Key), nil
}

func pasteToGist(content, filename, description string) (string, error) {
	ghToken := db.Get("GIT_TOKEN")
	if ghToken == "" {
		return "", fmt.Errorf("GIT_TOKEN not set")
	}

	gistReq := GistRequest{
		Description: description,
		Public:      false,
		Files:       map[string]GistFile{filename: {Content: content}},
	}
	jsonData, err := json.Marshal(gistReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gist request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+ghToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gist creation failed: %s", string(body))
	}

	var result GistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.HTMLURL, nil
}

func tryPasteServices(content string) (string, string, error) {
	pasteURL, err := pasteToPastesDev(content)
	if err == nil {
		return pasteURL, "pastes.dev", nil
	}

	pasteURL, err = pasteToPasteRs(content)
	if err == nil {
		return pasteURL, "paste.rs", nil
	}

	pasteURL, err = pasteToNekobin(content)
	if err == nil {
		return pasteURL, "Nekobin", nil
	}

	pasteURL, err = pasteToSpacebin(content)
	if err == nil {
		return pasteURL, "Spacebin", nil
	}

	pasteURL, err = pasteToDpaste(content)
	if err == nil {
		return pasteURL, "dpaste.org", nil
	}

	pasteURL, err = pasteTo0x0(content)
	if err == nil {
		return pasteURL, "0x0.st", nil
	}

	pasteURL, err = pasteToPasteSr(content)
	if err == nil {
		return pasteURL, "paste.sr.ht", nil
	}

	return "", "", fmt.Errorf("all paste services failed")
}

func parseProviderFlag(args string) (provider string, remaining string) {
	args = strings.TrimSpace(args)
	if strings.HasPrefix(args, "-p ") {
		parts := strings.SplitN(args, " ", 3)
		if len(parts) >= 2 {
			provider = strings.ToLower(parts[1])
			if len(parts) >= 3 {
				remaining = parts[2]
			}
		}
	} else {
		remaining = args
	}
	return
}

func pasteToProvider(content, filename, provider string) (string, string, error) {
	switch provider {
	case "gist", "github":
		url, err := pasteToGist(content, filename, "Pasted via NovaUserbot")
		return url, "GitHub Gist", err
	case "nekobin", "neko":
		url, err := pasteToNekobin(content)
		return url, "Nekobin", err
	case "spacebin", "space":
		url, err := pasteToSpacebin(content)
		return url, "Spacebin", err
	case "pastesr", "srht":
		url, err := pasteToPasteSr(content)
		return url, "paste.sr.ht", err
	case "dpaste":
		url, err := pasteToDpaste(content)
		return url, "dpaste.org", err
	case "0x0":
		url, err := pasteTo0x0(content)
		return url, "0x0.st", err
	case "pasters", "paste.rs":
		url, err := pasteToPasteRs(content)
		return url, "paste.rs", err
	case "pastesdev", "pastes.dev":
		url, err := pasteToPastesDev(content)
		return url, "pastes.dev", err
	default:
		return tryPasteServices(content)
	}
}

func pasteCommand(m *telegram.NewMessage) error {
	var content string
	var filename string
	var msg *telegram.NewMessage

	args := strings.TrimSpace(m.Args())
	provider, textContent := parseProviderFlag(args)

	if m.IsReply() {
		reply, err := m.GetReplyMessage()
		if err != nil {
			_, err := eOR(m, locales.Tr("paste.fetch_error"))
			return err
		}

		if reply.Media() != nil {
			doc := reply.Document()
			if doc == nil {
				_, err := eOR(m, locales.Tr("paste.unsupported_media"))
				return err
			}

			if doc.Size > maxPasteSize {
				_, err := eOR(m, locales.Tr("paste.file_too_large"))
				return err
			}

			msg, _ = eOR(m, locales.Tr("paste.downloading"))

			filePath, err := reply.Download()
			if err != nil {
				if msg != nil {
					msg.Edit(locales.Tr("paste.download_error"))
				}
				return err
			}
			defer os.Remove(filePath)

			data, err := os.ReadFile(filePath)
			if err != nil {
				if msg != nil {
					msg.Edit(locales.Tr("paste.read_error"))
				}
				return err
			}
			content = string(data)
			filename = filepath.Base(filePath)

			if msg != nil {
				msg.Edit(locales.Tr("paste.uploading"))
			}
		} else if reply.Text() != "" {
			content = reply.Text()
			filename = "paste.txt"
		} else {
			_, err := eOR(m, locales.Tr("paste.no_content"))
			return err
		}
	} else {
		if textContent == "" {
			_, err := eOR(m, locales.Tr("paste.usage"))
			return err
		}
		content = textContent
		filename = "paste.txt"
	}

	if len(content) > maxPasteSize {
		_, err := eOR(m, locales.Tr("paste.content_too_large"))
		return err
	}

	if msg == nil {
		msg, _ = eOR(m, locales.Tr("paste.uploading"))
	}

	var pasteURL string
	var err error
	var service string

	if provider != "" {
		pasteURL, service, err = pasteToProvider(content, filename, provider)
	} else if db.Get("GIT_TOKEN") != "" && (strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".py") || strings.HasSuffix(filename, ".go")) {
		pasteURL, err = pasteToGist(content, filename, "Pasted via NovaUserbot")
		service = "GitHub Gist"
		if err != nil {
			pasteURL, service, err = tryPasteServices(content)
		}
	} else {
		pasteURL, service, err = tryPasteServices(content)
	}

	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("paste.upload_error"), err.Error()))
		}
		return err
	}

	if msg != nil {
		_, err = msg.Edit(fmt.Sprintf(locales.Tr("paste.success"), service, pasteURL))
	}
	return err
}

func readFileCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("read.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("read.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("read.no_file"))
		return err
	}

	doc := reply.Document()
	if doc == nil {
		_, err := eOR(m, locales.Tr("read.no_file"))
		return err
	}

	if doc.Size > maxPasteSize {
		_, err := eOR(m, locales.Tr("read.file_too_large"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("read.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("read.download_error"))
		}
		return err
	}
	defer os.Remove(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("read.read_error"))
		}
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	args := strings.TrimSpace(m.Args())
	numLines := 0
	if strings.HasPrefix(args, "-n") {
		numStr := strings.TrimPrefix(args, "-n")
		numStr = strings.TrimSpace(numStr)
		if n, parseErr := strconv.Atoi(numStr); parseErr == nil && n > 0 {
			numLines = n
		}
	} else if args != "" {
		if n, parseErr := strconv.Atoi(args); parseErr == nil && n > 0 {
			numLines = n
		}
	}

	if numLines > 0 && numLines < len(lines) {
		lines = lines[:numLines]
	}

	output := strings.Join(lines, "\n")
	if len(output) > 4000 {
		output = output[:4000] + "\n...[truncated]"
	}

	filename := filepath.Base(filePath)
	result := fmt.Sprintf(locales.Tr("read.result"), filename, len(lines), output)

	if msg != nil {
		_, err = msg.Edit(result, &telegram.SendOptions{ParseMode: "HTML"})
	}
	return err
}

func LoadPasteModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "paste", Func: pasteCommand, Description: "Paste to Nekobin (-p for other services)", ModuleName: "Paste"},
		{Command: "read", Func: readFileCommand, Description: "Read file content (-n for line limit)", ModuleName: "Paste"},
	}
	AddHandlers(handlers, c)
}
