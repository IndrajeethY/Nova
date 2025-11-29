package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"regexp"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	defaultUpstreamRepo   = "https://github.com/IndrajeethY/Nova.git"
	defaultUpstreamBranch = "main"
)

var validBranchName = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

var validRepoURL = regexp.MustCompile(`^https?://[a-zA-Z0-9._@:/-]+\.git$`)

func sanitizeBranch(branch string) (string, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return defaultUpstreamBranch, nil
	}
	if !validBranchName.MatchString(branch) {
		return "", fmt.Errorf("invalid branch name")
	}
	return branch, nil
}

func sanitizeRepoURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return defaultUpstreamRepo, nil
	}
	if !validRepoURL.MatchString(url) {
		return "", fmt.Errorf("invalid repository URL")
	}
	return url, nil
}

func getRepoURL() (string, error) {
	gitToken := db.Get("GIT_TOKEN")
	upstreamRepo := db.Get("UPSTREAM_REPO")

	if upstreamRepo == "" {
		upstreamRepo = defaultUpstreamRepo
	}

	sanitized, err := sanitizeRepoURL(upstreamRepo)
	if err != nil {
		return "", err
	}

	if gitToken != "" && len(sanitized) > 8 && sanitized[:8] == "https://" {
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(gitToken) {
			return "", fmt.Errorf("invalid git token")
		}
		return fmt.Sprintf("https://%s@%s", gitToken, sanitized[8:]), nil
	}
	return sanitized, nil
}

func getUpstreamBranch() (string, error) {
	branch := db.Get("UPSTREAM_BRANCH")
	return sanitizeBranch(branch)
}

func checkForUpstreamChanges() (string, error) {
	repoURL, err := getRepoURL()
	if err != nil {
		return "", fmt.Errorf("invalid repo URL: %w", err)
	}
	branch, err := getUpstreamBranch()
	if err != nil {
		return "", fmt.Errorf("invalid branch: %w", err)
	}

	if _, err := utils.RunCommand(fmt.Sprintf("git fetch %s", repoURL)); err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}

	diffOutput, err := utils.RunCommand(fmt.Sprintf("git diff HEAD origin/%s", branch))
	if err != nil {
		return "", fmt.Errorf("diff failed: %w", err)
	}

	return diffOutput, nil
}

func resetAndPullLatest() error {
	repoURL, err := getRepoURL()
	if err != nil {
		return fmt.Errorf("invalid repo URL: %w", err)
	}
	branch, err := getUpstreamBranch()
	if err != nil {
		return fmt.Errorf("invalid branch: %w", err)
	}

	if _, err := utils.RunCommand("git reset --hard"); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	if _, err := utils.RunCommand(fmt.Sprintf("git pull %s %s", repoURL, branch)); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}

func checkUpdateCmd(m *telegram.NewMessage) error {
	msg, err := eOR(m, locales.Tr("updater.checking"))
	if err != nil {
		return err
	}

	diff, err := checkForUpstreamChanges()
	if err != nil {
		_, err = msg.Edit(fmt.Sprintf(locales.Tr("updater.check_error"), err.Error()))
		return err
	}

	if strings.TrimSpace(diff) == "" {
		_, err = msg.Edit(locales.Tr("updater.up_to_date"))
		return err
	}

	_, err = msg.Edit(locales.Tr("updater.updates_available"))
	return err
}

func updateCmd(m *telegram.NewMessage) error {
	msg, err := eOR(m, locales.Tr("updater.updating"))
	if err != nil {
		return err
	}

	err = resetAndPullLatest()
	if err != nil {
		_, err = msg.Edit(fmt.Sprintf(locales.Tr("updater.update_error"), err.Error()))
		return err
	}

	_, err = msg.Edit(locales.Tr("updater.update_success"))
	return err
}

func LoadUpdaterModule(c *telegram.Client) {
	handlers := []*Handler{
		{ModuleName: "Updater", Command: "checkupdate", Description: "Check for updates", Func: checkUpdateCmd},
		{ModuleName: "Updater", Command: "update", Description: "Update to latest version", Func: updateCmd, DisAllowSudos: true},
	}
	AddHandlers(handlers, c)
}
