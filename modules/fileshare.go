package modules

import (
	"NovaUserbot/locales"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type FileHostResponse struct {
	URL     string `json:"url"`
	Link    string `json:"link"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type CatboxResponse struct {
	URL string `json:"url"`
}

type GoFileResponse struct {
	Status string `json:"status"`
	Data   struct {
		DownloadPage string `json:"downloadPage"`
	} `json:"data"`
}

type FileIOResponse struct {
	Success bool   `json:"success"`
	Link    string `json:"link"`
}

func uploadToEnvsSh(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://envs.sh", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(data)), nil
}

func uploadToCatbox(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("reqtype", "fileupload")
	part, err := writer.CreateFormFile("fileToUpload", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://catbox.moe/user/api.php", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(data)), nil
}

var validGoFileServerPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func uploadToGoFile(filePath string) (string, error) {
	serverResp, err := http.Get("https://api.gofile.io/getServer")
	if err != nil {
		return "", err
	}
	defer serverResp.Body.Close()

	var serverData struct {
		Status string `json:"status"`
		Data   struct {
			Server string `json:"server"`
		} `json:"data"`
	}
	json.NewDecoder(serverResp.Body).Decode(&serverData)

	server := serverData.Data.Server
	if server == "" || !validGoFileServerPattern.MatchString(server) {
		server = "store1"
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.Close()

	uploadURL := fmt.Sprintf("https://%s.gofile.io/uploadFile", server)
	req, _ := http.NewRequest("POST", uploadURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result GoFileResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Status != "ok" {
		return "", fmt.Errorf("upload failed")
	}

	return result.Data.DownloadPage, nil
}

func uploadToFileIO(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://file.io", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result FileIOResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.Success {
		return "", fmt.Errorf("upload failed")
	}

	return result.Link, nil
}

func uploadTo0x0(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://0x0.st", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(data)), nil
}

func shareFileCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("fileshare.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("fileshare.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("fileshare.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("fileshare.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("fileshare.download_error"))
		return err
	}
	defer os.Remove(filePath)

	msg.Edit(locales.Tr("fileshare.uploading"))

	service := strings.ToLower(strings.TrimSpace(m.Args()))
	var link string

	switch service {
	case "catbox":
		link, err = uploadToCatbox(filePath)
	case "gofile":
		link, err = uploadToGoFile(filePath)
	case "fileio":
		link, err = uploadToFileIO(filePath)
	case "0x0":
		link, err = uploadTo0x0(filePath)
	default:
		link, err = uploadToEnvsSh(filePath)
		service = "envs.sh"
	}

	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_error"), err.Error()))
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_success"), service, link))
	return err
}

func catboxCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("fileshare.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("fileshare.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("fileshare.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("fileshare.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("fileshare.download_error"))
		return err
	}
	defer os.Remove(filePath)

	msg.Edit(locales.Tr("fileshare.uploading"))

	link, err := uploadToCatbox(filePath)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_error"), err.Error()))
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_success"), "Catbox", link))
	return err
}

func gofileCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("fileshare.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("fileshare.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("fileshare.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("fileshare.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("fileshare.download_error"))
		return err
	}
	defer os.Remove(filePath)

	msg.Edit(locales.Tr("fileshare.uploading"))

	link, err := uploadToGoFile(filePath)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_error"), err.Error()))
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_success"), "GoFile", link))
	return err
}

func fileioCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("fileshare.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("fileshare.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("fileshare.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("fileshare.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("fileshare.download_error"))
		return err
	}
	defer os.Remove(filePath)

	msg.Edit(locales.Tr("fileshare.uploading"))

	link, err := uploadToFileIO(filePath)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_error"), err.Error()))
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("fileshare.upload_success"), "File.io", link))
	return err
}

func LoadFileShareModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "share", Func: shareFileCommand, Description: "Share file to hosting service (catbox/gofile/fileio/0x0)", ModuleName: "FileShare"},
		{Command: "catbox", Func: catboxCommand, Description: "Upload to Catbox.moe", ModuleName: "FileShare"},
		{Command: "gofile", Func: gofileCommand, Description: "Upload to GoFile.io", ModuleName: "FileShare"},
		{Command: "fileio", Func: fileioCommand, Description: "Upload to File.io (one-time download)", ModuleName: "FileShare"},
	}
	AddHandlers(handlers, c)
}
