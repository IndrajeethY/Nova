package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"golang.org/x/oauth2"
)

const (
	gdriveAuthURL    = "https://accounts.google.com/o/oauth2/auth"
	gdriveTokenURL   = "https://oauth2.googleapis.com/token"
	gdriveUploadURL  = "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	gdriveFilesURL   = "https://www.googleapis.com/drive/v3/files"
)

var gdriveScopes = []string{
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/drive",
}

type GDriveConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenExpiry  int64  `json:"token_expiry"`
}

type GDriveFile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	MimeType       string `json:"mimeType"`
	Size           string `json:"size"`
	WebViewLink    string `json:"webViewLink"`
	WebContentLink string `json:"webContentLink"`
}

type GDriveFileList struct {
	Files         []GDriveFile `json:"files"`
	NextPageToken string       `json:"nextPageToken"`
}

func getGDriveConfig() (*GDriveConfig, error) {
	data := db.Get("GDRIVE_CONFIG")
	if data == "" {
		return nil, fmt.Errorf("gdrive not configured")
	}
	var config GDriveConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func saveGDriveConfig(config *GDriveConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return db.Set("GDRIVE_CONFIG", string(data))
}

func getGDriveOAuth2Config(config *GDriveConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  gdriveAuthURL,
			TokenURL: gdriveTokenURL,
		},
		RedirectURL: "urn:ietf:wg:oauth:2.0:oob",
		Scopes:      gdriveScopes,
	}
}

func refreshGDriveToken(config *GDriveConfig) error {
	if config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	oauth2Config := getGDriveOAuth2Config(config)
	token := &oauth2.Token{
		RefreshToken: config.RefreshToken,
	}

	tokenSource := oauth2Config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return err
	}

	config.AccessToken = newToken.AccessToken
	config.TokenExpiry = newToken.Expiry.Unix()
	if newToken.RefreshToken != "" {
		config.RefreshToken = newToken.RefreshToken
	}

	return saveGDriveConfig(config)
}

func getValidGDriveToken(config *GDriveConfig) (string, error) {
	if config.TokenExpiry > 0 && time.Now().Unix() >= config.TokenExpiry-300 {
		if err := refreshGDriveToken(config); err != nil {
			return "", err
		}
	}
	return config.AccessToken, nil
}

func gdriveSetupCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("gdrive.setup_usage"))
		return err
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		_, err := eOR(m, locales.Tr("gdrive.setup_usage"))
		return err
	}

	clientID := strings.TrimSpace(parts[0])
	clientSecret := strings.TrimSpace(parts[1])

	config := &GDriveConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	if err := saveGDriveConfig(config); err != nil {
		_, err := eOR(m, locales.Tr("gdrive.setup_error"))
		return err
	}

	oauth2Config := getGDriveOAuth2Config(config)
	authURL := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	_, err := eOR(m, fmt.Sprintf(locales.Tr("gdrive.auth_url"), authURL))
	return err
}

func gdriveAuthCommand(m *telegram.NewMessage) error {
	code := strings.TrimSpace(m.Args())
	if code == "" {
		_, err := eOR(m, locales.Tr("gdrive.auth_usage"))
		return err
	}

	config, err := getGDriveConfig()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.not_configured"))
		return err
	}

	oauth2Config := getGDriveOAuth2Config(config)
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("gdrive.auth_error"), err.Error()))
		return err
	}

	config.AccessToken = token.AccessToken
	config.RefreshToken = token.RefreshToken
	config.TokenExpiry = token.Expiry.Unix()

	if err := saveGDriveConfig(config); err != nil {
		_, err := eOR(m, locales.Tr("gdrive.setup_error"))
		return err
	}

	_, err = eOR(m, locales.Tr("gdrive.auth_success"))
	return err
}

func gdriveUploadCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("gdrive.reply_to_file"))
		return err
	}

	config, err := getGDriveConfig()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.not_configured"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("gdrive.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gdrive.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.download_error"))
		return err
	}
	defer os.Remove(filePath)

	msg.Edit(locales.Tr("gdrive.uploading"))

	token, err := getValidGDriveToken(config)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.token_error"), err.Error()))
		return err
	}

	fileName := filepath.Base(filePath)
	customName := strings.TrimSpace(m.Args())
	if customName != "" {
		fileName = customName
	}

	fileLink, err := uploadToGDrive(filePath, fileName, token)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.upload_error"), err.Error()))
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("gdrive.upload_success"), fileName, fileLink))
	return err
}

func uploadToGDrive(filePath, fileName, token string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	metadataObj := map[string]string{"name": fileName}
	metadataBytes, err := json.Marshal(metadataObj)
	if err != nil {
		return "", err
	}

	boundary := "boundary"
	body := fmt.Sprintf("--%s\r\nContent-Type: application/json; charset=UTF-8\r\n\r\n%s\r\n--%s\r\nContent-Type: application/octet-stream\r\n\r\n",
		boundary, string(metadataBytes), boundary)

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	fullBody := body + string(fileContent) + fmt.Sprintf("\r\n--%s--", boundary)

	req, err := http.NewRequest("POST", gdriveUploadURL, strings.NewReader(fullBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/related; boundary=%s", boundary))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(fullBody)))

	timeoutMinutes := fileInfo.Size()/1024/1024 + 1
	if timeoutMinutes > 30 {
		timeoutMinutes = 30
	}
	httpClient := &http.Client{Timeout: time.Duration(timeoutMinutes) * time.Minute}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(respBody))
	}

	var result GDriveFile
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	permReq, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/permissions", gdriveFilesURL, result.ID),
		strings.NewReader(`{"role": "reader", "type": "anyone"}`))
	if err == nil {
		permReq.Header.Set("Authorization", "Bearer "+token)
		permReq.Header.Set("Content-Type", "application/json")
		permResp, permErr := httpClient.Do(permReq)
		if permErr != nil {
			logger.Warnf("Failed to set file permissions: %v", permErr)
		} else {
			permResp.Body.Close()
		}
	}

	return fmt.Sprintf("https://drive.google.com/file/d/%s/view", result.ID), nil
}

func gdriveListCommand(m *telegram.NewMessage) error {
	config, err := getGDriveConfig()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.not_configured"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gdrive.fetching"))

	token, err := getValidGDriveToken(config)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.token_error"), err.Error()))
		return err
	}

	query := strings.TrimSpace(m.Args())
	apiURL := gdriveFilesURL + "?pageSize=10&fields=files(id,name,mimeType,size,webViewLink)"
	if query != "" {
		escapedQuery := strings.ReplaceAll(query, "'", "\\'")
		escapedQuery = strings.ReplaceAll(escapedQuery, "\\", "\\\\")
		apiURL += "&q=" + url.QueryEscape(fmt.Sprintf("name contains '%s'", escapedQuery))
	}

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.list_error"))
		return err
	}
	defer resp.Body.Close()

	var result GDriveFileList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.list_error"))
		return err
	}

	if len(result.Files) == 0 {
		_, err := msg.Edit(locales.Tr("gdrive.no_files"))
		return err
	}

	text := locales.Tr("gdrive.list_header")
	for _, file := range result.Files {
		text += fmt.Sprintf("\n• <a href='%s'>%s</a>", file.WebViewLink, file.Name)
	}

	_, err = msg.Edit(text, &telegram.SendOptions{ParseMode: "HTML", LinkPreview: false})
	return err
}

func gdriveSearchCommand(m *telegram.NewMessage) error {
	query := strings.TrimSpace(m.Args())
	if query == "" {
		_, err := eOR(m, locales.Tr("gdrive.search_usage"))
		return err
	}

	config, err := getGDriveConfig()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.not_configured"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gdrive.searching"))

	token, err := getValidGDriveToken(config)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.token_error"), err.Error()))
		return err
	}

	escapedQuery := strings.ReplaceAll(query, "'", "\\'")
	escapedQuery = strings.ReplaceAll(escapedQuery, "\\", "\\\\")
	apiURL := gdriveFilesURL + "?pageSize=15&fields=files(id,name,mimeType,size,webViewLink)&q=" +
		url.QueryEscape(fmt.Sprintf("name contains '%s'", escapedQuery))

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.search_error"))
		return err
	}
	defer resp.Body.Close()

	var result GDriveFileList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.search_error"))
		return err
	}

	if len(result.Files) == 0 {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.no_results"), query))
		return err
	}

	text := fmt.Sprintf(locales.Tr("gdrive.search_header"), query)
	for _, file := range result.Files {
		text += fmt.Sprintf("\n• <a href='%s'>%s</a>", file.WebViewLink, file.Name)
	}

	_, err = msg.Edit(text, &telegram.SendOptions{ParseMode: "HTML", LinkPreview: false})
	return err
}

func gdriveDownloadCommand(m *telegram.NewMessage) error {
	fileID := strings.TrimSpace(m.Args())
	if fileID == "" {
		_, err := eOR(m, locales.Tr("gdrive.download_usage"))
		return err
	}

	if strings.Contains(fileID, "drive.google.com") {
		parts := strings.Split(fileID, "/")
		for i, part := range parts {
			if part == "d" && i+1 < len(parts) {
				fileID = parts[i+1]
				break
			}
		}
	}

	config, err := getGDriveConfig()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.not_configured"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gdrive.downloading"))

	token, err := getValidGDriveToken(config)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.token_error"), err.Error()))
		return err
	}

	metaReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s?fields=name,size", gdriveFilesURL, fileID), nil)
	metaReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	metaResp, err := client.Do(metaReq)
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.file_not_found"))
		return err
	}
	defer metaResp.Body.Close()

	var fileMeta GDriveFile
	if err := json.NewDecoder(metaResp.Body).Decode(&fileMeta); err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.file_not_found"))
		return err
	}

	downloadReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s?alt=media", gdriveFilesURL, fileID), nil)
	downloadReq.Header.Set("Authorization", "Bearer "+token)

	downloadClient := &http.Client{Timeout: 10 * time.Minute}
	downloadResp, err := downloadClient.Do(downloadReq)
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.download_failed"))
		return err
	}
	defer downloadResp.Body.Close()

	tmpFile := filepath.Join("/tmp", fileMeta.Name)
	outFile, err := os.Create(tmpFile)
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.download_failed"))
		return err
	}

	_, err = io.Copy(outFile, downloadResp.Body)
	outFile.Close()
	if err != nil {
		os.Remove(tmpFile)
		_, err := msg.Edit(locales.Tr("gdrive.download_failed"))
		return err
	}

	msg.Edit(locales.Tr("gdrive.uploading_telegram"))

	_, err = m.Respond(fmt.Sprintf(locales.Tr("gdrive.file_caption"), fileMeta.Name), &telegram.SendOptions{
		Media: tmpFile,
	})
	os.Remove(tmpFile)

	if err != nil {
		logger.Errorf("Failed to send file: %v", err)
		_, err := msg.Edit(locales.Tr("gdrive.send_error"))
		return err
	}

	msg.Delete()
	return nil
}

func gdriveDeleteCommand(m *telegram.NewMessage) error {
	fileID := strings.TrimSpace(m.Args())
	if fileID == "" {
		_, err := eOR(m, locales.Tr("gdrive.delete_usage"))
		return err
	}

	if strings.Contains(fileID, "drive.google.com") {
		parts := strings.Split(fileID, "/")
		for i, part := range parts {
			if part == "d" && i+1 < len(parts) {
				fileID = parts[i+1]
				break
			}
		}
	}

	config, err := getGDriveConfig()
	if err != nil {
		_, err := eOR(m, locales.Tr("gdrive.not_configured"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gdrive.deleting"))

	token, err := getValidGDriveToken(config)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("gdrive.token_error"), err.Error()))
		return err
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", gdriveFilesURL, fileID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		_, err := msg.Edit(locales.Tr("gdrive.delete_error"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		_, err := msg.Edit(locales.Tr("gdrive.delete_error"))
		return err
	}

	_, err = msg.Edit(locales.Tr("gdrive.delete_success"))
	return err
}

func LoadGDriveModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "gsetup", Func: gdriveSetupCommand, Description: "Setup Google Drive (client_id client_secret)", ModuleName: "GDrive", DisAllowSudos: true},
		{Command: "gauth", Func: gdriveAuthCommand, Description: "Authorize Google Drive with auth code", ModuleName: "GDrive", DisAllowSudos: true},
		{Command: "gupload", Func: gdriveUploadCommand, Description: "Upload replied file to Google Drive", ModuleName: "GDrive", DisAllowSudos: true},
		{Command: "glist", Func: gdriveListCommand, Description: "List files in Google Drive", ModuleName: "GDrive", DisAllowSudos: true},
		{Command: "gsearch", Func: gdriveSearchCommand, Description: "Search files in Google Drive", ModuleName: "GDrive", DisAllowSudos: true},
		{Command: "gdown", Func: gdriveDownloadCommand, Description: "Download file from Google Drive", ModuleName: "GDrive", DisAllowSudos: true},
		{Command: "gdelete", Func: gdriveDeleteCommand, Description: "Delete file from Google Drive", ModuleName: "GDrive", DisAllowSudos: true},
	}
	AddHandlers(handlers, c)
}
