package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func unsplashSearch(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("unsplash.usage"))
		return err
	}

	query := args
	limit := 5

	if strings.Contains(args, ";") {
		parts := strings.SplitN(args, ";", 2)
		query = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			if n, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil && n > 0 {
				limit = n
				if limit > 10 {
					limit = 10
				}
			}
		}
	}

	msg, _ := eOR(m, locales.Tr("unsplash.searching"))

	photos := fetchUnsplashPhotoURLs(query, limit)
	if len(photos) == 0 {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("unsplash.no_results"), query))
		return err
	}

	msg.Delete()

	var downloadedFiles []string
	for i, imgURL := range photos {
		tmpFile := fmt.Sprintf("/tmp/unsplash_%d_%d.jpg", time.Now().UnixNano(), i)
		if err := downloadFileToPath(imgURL, tmpFile); err != nil {
			continue
		}
		downloadedFiles = append(downloadedFiles, tmpFile)
	}

	if len(downloadedFiles) == 0 {
		_, err := m.Respond(fmt.Sprintf(locales.Tr("unsplash.download_error"), query))
		return err
	}

	defer func() {
		for _, file := range downloadedFiles {
			os.Remove(file)
		}
	}()

	caption := fmt.Sprintf(locales.Tr("unsplash.caption"), query)

	if len(downloadedFiles) > 1 {
		_, err := m.RespondAlbum(downloadedFiles, &telegram.MediaOptions{
			Caption: caption,
		})
		if err != nil {
			return err
		}
	} else {
		_, err := m.RespondMedia(downloadedFiles[0], &telegram.MediaOptions{
			Caption: caption,
		})
		if err != nil {
			return err
		}
	}

	_, err := m.Respond(fmt.Sprintf(locales.Tr("unsplash.uploaded"), len(downloadedFiles)))
	return err
}

func fetchUnsplashPhotoURLs(query string, limit int) []string {

	query = strings.ReplaceAll(query, " ", "-")
	searchURL := fmt.Sprintf("https://unsplash.com/s/photos/%s", url.PathEscape(query))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	srcPattern := regexp.MustCompile(`src="(https://images\.unsplash\.com/photo[^"]+)"`)
	matches := srcPattern.FindAllStringSubmatch(string(body), -1)

	var urls []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			imgURL := match[1]

			if idx := strings.Index(imgURL, "?"); idx != -1 {
				baseURL := imgURL[:idx]

				imgURL = baseURL + "?w=800&q=80"
			}

			if !seen[imgURL] {
				seen[imgURL] = true
				urls = append(urls, imgURL)
			}
		}
	}

	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	if len(urls) > limit {
		urls = urls[:limit]
	}

	return urls
}

func downloadFileToPath(urlStr, filepath string) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func LoadUnsplashModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "unsplash", Description: "Search and download Unsplash images", Func: unsplashSearch, ModuleName: "Unsplash"},
	}
	AddHandlers(handlers, c)
}
