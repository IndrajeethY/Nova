package modules

import (
	"NovaUserbot/locales"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

type GitHubUser struct {
	Login       string `json:"login"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Company     string `json:"company"`
	Blog        string `json:"blog"`
	Location    string `json:"location"`
	Bio         string `json:"bio"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	HTMLURL     string `json:"html_url"`
}

type IPInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

type GoogleSearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
}

func gitHubSearch(m *telegram.NewMessage) error {
	username := strings.TrimSpace(m.Args())
	if username == "" {
		_, err := eOR(m, locales.Tr("search.github_usage"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("search.fetching"))

	apiURL := fmt.Sprintf("https://api.github.com/users/%s", url.PathEscape(username))
	resp, err := http.Get(apiURL)
	if err != nil {
		_, err := msg.Edit(locales.Tr("search.github_error"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, err := msg.Edit(locales.Tr("search.github_not_found"))
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_, err := msg.Edit(locales.Tr("search.github_error"))
		return err
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		_, err := msg.Edit(locales.Tr("search.github_error"))
		return err
	}

	profilePic := fmt.Sprintf("https://avatars.githubusercontent.com/u/%d", user.ID)

	result := fmt.Sprintf(locales.Tr("search.github_result"),
		user.HTMLURL,
		user.Name,
		user.Login,
		user.ID,
		user.Company,
		user.Blog,
		user.Location,
		user.Bio,
		user.PublicRepos,
		user.Followers,
		user.Following,
	)

	_, err = msg.Edit(result, &telegram.SendOptions{
		ParseMode: "HTML",
		Media:     profilePic,
	})
	return err
}

func googleSearch(m *telegram.NewMessage) error {
	query := strings.TrimSpace(m.Args())
	if query == "" {
		_, err := eOR(m, locales.Tr("search.google_usage"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("search.searching"))

	results, err := performDuckDuckGoSearch(query)
	if err != nil || len(results) == 0 {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("search.no_results"), query))
		return err
	}

	var output strings.Builder
	for _, res := range results {
		output.WriteString(fmt.Sprintf(" 👉🏻  <a href='%s'>%s</a>\n<code>%s</code>\n\n", res.Link, res.Title, res.Description))
	}

	response := fmt.Sprintf(locales.Tr("search.google_result"), query, output.String())
	_, err = msg.Edit(response, &telegram.SendOptions{ParseMode: "HTML", LinkPreview: false})
	return err
}

type DDGResponse struct {
	AbstractText  string `json:"AbstractText"`
	AbstractURL   string `json:"AbstractURL"`
	Heading       string `json:"Heading"`
	RelatedTopics []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
}

func performDuckDuckGoSearch(query string) ([]GoogleSearchResult, error) {
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_redirect=1", url.QueryEscape(query))
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ddg DDGResponse
	if err := json.Unmarshal(body, &ddg); err != nil {
		return nil, err
	}

	var results []GoogleSearchResult

	if ddg.AbstractText != "" {
		results = append(results, GoogleSearchResult{
			Title:       ddg.Heading,
			Link:        ddg.AbstractURL,
			Description: ddg.AbstractText,
		})
	}

	for _, topic := range ddg.RelatedTopics {
		if len(results) >= 5 {
			break
		}
		if topic.Text != "" && topic.FirstURL != "" {
			title := topic.Text
			if len(title) > 50 {
				title = title[:50] + "..."
			}
			results = append(results, GoogleSearchResult{
				Title:       title,
				Link:        topic.FirstURL,
				Description: topic.Text,
			})
		}
	}

	return results, nil
}

func imageSearch(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("search.img_usage"))
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

	msg, _ := eOR(m, locales.Tr("search.searching_images"))

	images := fetchUnsplashPhotoURLs(query, limit)
	if len(images) == 0 {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("search.no_images"), query))
		return err
	}

	msg.Delete()

	for _, img := range images {
		_, err := m.Respond(fmt.Sprintf(locales.Tr("search.image_caption"), query), &telegram.SendOptions{
			Media: img,
		})
		if err != nil {
			continue
		}
	}

	return nil
}

func ipInfoSearch(m *telegram.NewMessage) error {
	ipAddr := strings.TrimSpace(m.Args())
	if ipAddr == "" {
		_, err := eOR(m, locales.Tr("search.ipinfo_usage"))
		return err
	}

	if net.ParseIP(ipAddr) == nil {
		_, err := eOR(m, locales.Tr("search.ipinfo_not_found"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("search.fetching"))

	apiURL := fmt.Sprintf("https://ipinfo.io/%s/json", url.PathEscape(ipAddr))
	resp, err := http.Get(apiURL)
	if err != nil {
		_, err := msg.Edit(locales.Tr("search.ipinfo_error"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, err := msg.Edit(locales.Tr("search.ipinfo_not_found"))
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_, err := msg.Edit(locales.Tr("search.ipinfo_error"))
		return err
	}

	var info IPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		_, err := msg.Edit(locales.Tr("search.ipinfo_error"))
		return err
	}

	result := fmt.Sprintf(locales.Tr("search.ipinfo_result"),
		html.EscapeString(info.IP),
		html.EscapeString(info.Hostname),
		html.EscapeString(info.City),
		html.EscapeString(info.Region),
		html.EscapeString(info.Country),
		html.EscapeString(info.Loc),
		html.EscapeString(info.Org),
		html.EscapeString(info.Postal),
		html.EscapeString(info.Timezone),
	)

	_, err = msg.Edit(result, &telegram.SendOptions{
		ParseMode: "HTML",
	})
	return err
}

func LoadSearchModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "github", Description: "Get GitHub user profile info", Func: gitHubSearch, ModuleName: "Search"},
		{Command: "gh", Description: "Get GitHub user profile info (alias)", Func: gitHubSearch, ModuleName: "Search"},
		{Command: "google", Description: "Search on Google/DuckDuckGo", Func: googleSearch, ModuleName: "Search"},
		{Command: "img", Description: "Search for images", Func: imageSearch, ModuleName: "Search"},
		{Command: "ipinfo", Description: "Get IP address information", Func: ipInfoSearch, ModuleName: "Search"},
	}
	AddHandlers(handlers, c)
}
