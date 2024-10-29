package shadowhub

import (
	"fmt"
	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// createShadowHubClient creates a GitHub client with an optional token for authentication.
func createShadowHubClient(token string) *github.Client {
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		log.Debug().Msg("Creating authenticated GitHub client")
		return github.NewClient(tc)
	}
	log.Debug().Msg("Creating unauthenticated GitHub client")
	return github.NewClient(nil)
}

// fetchShadowHubFile fetches the content of a GitHub file.
func fetchShadowHubFile(client *github.Client, owner, repo, path string) (string, error) {
	log.Info().Str("owner", owner).Str("repo", repo).Str("path", path).Msg("Fetching file from GitHub")
	content, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch file from GitHub")
		return "", err
	}
	fileContent, err := content.GetContent()
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode file content from GitHub")
		return "", err
	}
	log.Debug().Str("file_content", fileContent).Msg("Fetched file content")
	return fileContent, nil
}

// extractURLs extracts URLs from script content, excluding those with variables.
func extractURLs(scriptContent string) []string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	allURLs := re.FindAllString(scriptContent, -1)

	var validURLs []string
	for _, url := range allURLs {
		if !regexp.MustCompile(`\$\{[^}]+\}`).MatchString(url) {
			validURLs = append(validURLs, url)
		}
	}
	log.Debug().Int("url_count", len(validURLs)).Msg("Extracted valid URLs from script content")
	return validURLs
}

// extractVariables extracts specified variables from script content.
func extractVariables(scriptContent string, variableNames []string) map[string]string {
	variableValues := make(map[string]string)
	for _, variableName := range variableNames {
		pattern := fmt.Sprintf(`(?m)^\s*%s\s*=\s*['"]?([^'"\s]+)['"]?`, variableName)
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(scriptContent)
		if len(match) > 1 {
			variableValues[variableName] = match[1]
		}
	}
	log.Debug().Int("variable_count", len(variableValues)).Msg("Extracted variables from script content")
	return variableValues
}

// longestCommonPrefix calculates the longest common prefix between two strings.
func longestCommonPrefix(s1, s2 string) int {
	minLength := len(s1)
	if len(s2) < minLength {
		minLength = len(s2)
	}
	for i := 0; i < minLength; i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	return minLength
}

// findLongestMatchingURL finds the upstream URL with the longest match.
func findLongestMatchingURL(localURL string, upstreamURLs []string) (string, int) {
	bestMatch := ""
	longestMatchLength := 0
	for _, upstreamURL := range upstreamURLs {
		matchLength := longestCommonPrefix(localURL, upstreamURL)
		if matchLength > longestMatchLength {
			longestMatchLength = matchLength
			bestMatch = upstreamURL
		}
	}
	log.Debug().Str("local_url", localURL).Str("best_match", bestMatch).Int("match_length", longestMatchLength).Msg("Found longest matching URL")
	return bestMatch, longestMatchLength
}

// checkURLChanges checks for URL changes by comparing local URLs with upstream URLs.
func checkURLChanges(localURLs, upstreamURLs []string) bool {
	changed := false
	for _, localURL := range localURLs {
		bestMatch, matchLength := findLongestMatchingURL(localURL, upstreamURLs)
		if matchLength != len(localURL) || matchLength != len(bestMatch) {
			log.Warn().Str("local_url", localURL).Str("best_match", bestMatch).Msg("URL has changed")
			changed = true
		}
	}
	for _, upstreamURL := range upstreamURLs {
		_, matchLength := findLongestMatchingURL(upstreamURL, localURLs)
		if matchLength <= 8 {
			log.Info().Str("upstream_url", upstreamURL).Msg("New upstream URL found")
			changed = true
		}
	}
	return changed
}

// updateVariable updates a specific variable in the script content.
func updateVariable(scriptContent, variableName, newValue string) string {
	variablePattern := fmt.Sprintf(`(?m)^%s\s*=\s*['"]?([^'"\s]+)['"]?`, variableName)
	re := regexp.MustCompile(variablePattern)
	updatedScript := re.ReplaceAllString(scriptContent, fmt.Sprintf(`%s='%s'`, variableName, newValue))
	log.Debug().Str("variable_name", variableName).Str("new_value", newValue).Msg("Updated variable in script content")
	return updatedScript
}

// updateCustomVariables updates custom variables based on upstream changes.
func updateCustomVariables(localScriptContent, upstreamScriptContent string, variableNames []string) (string, bool) {
	changed := false
	localVars := extractVariables(localScriptContent, variableNames)
	upstreamVars := extractVariables(upstreamScriptContent, variableNames)
	for variableName, upstreamValue := range upstreamVars {
		if localValue, exists := localVars[variableName]; exists && localValue != upstreamValue {
			log.Info().Str("variable_name", variableName).Str("old_value", localValue).Str("new_value", upstreamValue).Msg("Variable value updated")
			localScriptContent = updateVariable(localScriptContent, variableName, upstreamValue)
			changed = true
		}
	}
	return localScriptContent, changed
}

// remapLocalPathToUpstream maps local paths to upstream paths.
func remapLocalPathToUpstream(localPath, workspaceImageFilePath string) string {
	upstreamPath := strings.ReplaceAll(localPath, `\`, `/`)
	upstreamPath = strings.TrimPrefix(upstreamPath, "workspace-core-image/")
	log.Debug().Str("local_path", localPath).Str("upstream_path", upstreamPath).Msg("Remapped local path to upstream path")
	return upstreamPath
}

// UpdateShadowDependencies updates dependencies for all scripts in a workspace.
func UpdateShadowDependencies(workspaceImageFilePath, token string) {
	variablesToCheck := []string{"COMMIT_ID", "BRANCH", "KASMVNC_VER", "SQUID_COMMIT"}
	owner := "kasmtech"
	repo := "workspaces-core-images"

	client := createShadowHubClient(token)
	err := filepath.Walk(workspaceImageFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking directory path")
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sh") {
			log.Info().Str("path", path).Msg("Processing script")

			localScriptContent, err := ioutil.ReadFile(path)
			if err != nil {
				log.Error().Err(err).Str("path", path).Msg("Failed to read local script")
				return nil
			}
			upstreamPath := remapLocalPathToUpstream(path, workspaceImageFilePath)
			upstreamScriptContent, err := fetchShadowHubFile(client, owner, repo, upstreamPath)
			if err != nil {
				log.Error().Err(err).Str("upstream_path", upstreamPath).Msg("Failed to fetch upstream script")
				return nil
			}

			localURLs := extractURLs(string(localScriptContent))
			upstreamURLs := extractURLs(upstreamScriptContent)
			urlsChanged := checkURLChanges(localURLs, upstreamURLs)
			updatedScriptContent, varsChanged := updateCustomVariables(string(localScriptContent), upstreamScriptContent, variablesToCheck)

			if urlsChanged || varsChanged {
				err = ioutil.WriteFile(path, []byte(updatedScriptContent), 0644)
				if err != nil {
					log.Error().Err(err).Str("path", path).Msg("Failed to write updated script")
					return nil
				}
				log.Info().Str("path", path).Msg("Script updated due to detected changes")
			} else {
				log.Info().Str("path", path).Msg("No changes detected in script")
			}
		}
		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("workspace_path", workspaceImageFilePath).Msg("Error walking workspace path")
	}
}
