package modules

import (
	"NovaUserbot/utils"
	"fmt"
)

const (
	repoURL = "https://github.com/IndrajeethY/Nova.git"
	branch  = "dev"
)

func checkForUpstreamChanges() (string, error) {
	fetchCmd := fmt.Sprintf("git fetch %s", repoURL)
	_, err := utils.RunCommand(fetchCmd)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from upstream: %w", err)
	}

	diffCmd := fmt.Sprintf("git diff HEAD origin/%s", branch)
	diffOutput, err := utils.RunCommand(diffCmd)
	if err != nil {
		return "", fmt.Errorf("failed to check for differences: %w", err)
	}

	return diffOutput, nil
}

func resetAndPullLatest() error {
	resetCmd := "git reset --hard"
	_, err := utils.RunCommand(resetCmd)
	if err != nil {
		return fmt.Errorf("failed to reset local changes: %w", err)
	}

	pullCmd := fmt.Sprintf("git pull %s %s", repoURL, branch)
	_, err = utils.RunCommand(pullCmd)
	if err != nil {
		return fmt.Errorf("failed to pull latest updates: %w", err)
	}

	return nil
}
