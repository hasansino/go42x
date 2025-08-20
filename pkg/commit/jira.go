package commit

import (
	"regexp"
	"strings"
)

var jiraPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^([A-Z]+-\d+)`),
	regexp.MustCompile(`^feature/([A-Z]+-\d+)(?:-.*)?$`),
	regexp.MustCompile(`^bugfix/([A-Z]+-\d+)(?:-.*)?$`),
	regexp.MustCompile(`^hotfix/([A-Z]+-\d+)(?:-.*)?$`),
	regexp.MustCompile(`^chore/([A-Z]+-\d+)(?:-.*)?$`),
	regexp.MustCompile(`/([A-Z]+-\d+)(?:-|$)`),
}

func DetectJIRAPrefix(branchName string) string {
	if branchName == "" || branchName == "main" || branchName == "master" || branchName == "develop" {
		return ""
	}

	for _, pattern := range jiraPatterns {
		matches := pattern.FindStringSubmatch(branchName)
		if len(matches) > 1 && matches[1] != "" {
			return matches[1] + ": "
		}
	}

	return ""
}

func ApplyJIRAPrefix(commitMessage, jiraPrefix string) string {
	if jiraPrefix == "" {
		return commitMessage
	}

	if strings.HasPrefix(commitMessage, jiraPrefix) {
		return commitMessage
	}

	return jiraPrefix + commitMessage
}
