package utils

import (
	"NovaUserbot/db"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/generative-ai-go/genai"
	"github.com/nfnt/resize"

	"slices"

	"google.golang.org/api/option"
)

func compressImage(imagePath string) ([]byte, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error opening image file: %w", err)
	}
	defer file.Close()
	defer os.Remove(imagePath)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}
	maxWidth, maxHeight := uint(1024), uint(1024)
	resizedImg := resize.Thumbnail(maxWidth, maxHeight, img, resize.Lanczos3)

	var compressedBuffer bytes.Buffer
	err = jpeg.Encode(&compressedBuffer, resizedImg, nil)
	if err != nil {
		return nil, fmt.Errorf("error encoding resized image to JPG: %w", err)
	}

	return compressedBuffer.Bytes(), nil
}

func ProcessGemini(imagePath, text string) (string, error) {
	ctx := context.Background()
	key := db.RDb.Get(ctx, "GEMINI_API_KEY").Val()
	if key == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return "", err
	}
	defer client.Close()

	var req []genai.Part

	if imagePath != "" {
		compressedImage, err := compressImage(imagePath)
		if err != nil {
			return "", err
		}
		req = append(req, genai.ImageData("png", compressedImage))
	}
	req = append(req, genai.Text(text))

	model := client.GenerativeModel("gemini-2.0-flash")
	resp, err := model.GenerateContent(ctx, req...)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates found in response")
	}
	if len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no parts found in response")
	}

	return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]), nil
}

func RunCommand(cmd string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	var proc *exec.Cmd
	if runtime.GOOS == "windows" {
		proc = exec.Command("cmd", "/C", cmd)
	} else {
		proc = exec.Command("bash", "-c", cmd)
	}

	proc.Stderr = &stderr
	proc.Stdout = &stdout
	err := proc.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

func IsIn64Array(arr []int64, val int64) bool {
	return slices.Contains(arr, val)
}

func StringToInt64(str string) int64 {
	var i int64
	fmt.Sscanf(str, "%d", &i)
	return i
}

func RemoveFrom64Array(arr []int64, val int64) []int64 {
	for i, v := range arr {
		if v == val {
			return slices.Delete(arr, i, i+1)
		}
	}
	return arr
}

func UploadFileToEnvsSh(filePath string) (string, error) {
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
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://envs.sh", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload file, status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
