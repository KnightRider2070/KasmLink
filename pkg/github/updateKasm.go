package shadowhub

import (
	"fmt"
	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	URLRegexPattern    = `https?://[^\s]+`
	VariablePatternStr = `(?m)^\s*%s\s*=\s*['"]?([^'"\s]+)['"]?`
)

var (
	urlRegex = regexp.MustCompile(URLRegexPattern)
)

// GitHubClient encapsulates GitHub API logic.
type GitHubClient struct {
	client *github.Client
}

// NewGitHubClient initializes a GitHub client, optionally authenticated.
func NewGitHubClient(token string) *GitHubClient {
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		log.Debug().Msg("Creating authenticated GitHub client")
		return &GitHubClient{client: github.NewClient(tc)}
	}
	log.Debug().Msg("Creating unauthenticated GitHub client")
	return &GitHubClient{client: github.NewClient(nil)}
}

// FetchFileContent retrieves the content of a file from a GitHub repository.
func (gh *GitHubClient) FetchFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	log.Info().Str("owner", owner).Str("repo", repo).Str("path", path).Msg("Fetching file from GitHub")

	content, _, _, err := gh.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file from GitHub: %w", err)
	}

	fileContent, err := content.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode file content from GitHub: %w", err)
	}

	log.Debug().Str("path", path).Msg("Fetched file content successfully")
	return fileContent, nil
}

// ScriptProcessor processes and updates scripts within a workspace.
type ScriptProcessor struct {
	githubClient      *GitHubClient
	variablesToCheck  []string
	workspaceBasePath string
	repoOwner         string
	repoName          string
}

// NewScriptProcessor initializes a new ScriptProcessor.
func NewScriptProcessor(client *GitHubClient, variables []string, workspaceBasePath, owner, repo string) *ScriptProcessor {
	return &ScriptProcessor{
		githubClient:      client,
		variablesToCheck:  variables,
		workspaceBasePath: workspaceBasePath,
		repoOwner:         owner,
		repoName:          repo,
	}
}

// ProcessScripts iterates through workspace scripts and updates them based on upstream changes.
func (sp *ScriptProcessor) ProcessScripts(ctx context.Context) {
	err := filepath.Walk(sp.workspaceBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error accessing file")
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sh") {
			sp.processSingleScript(ctx, path)
		}
		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("workspace_path", sp.workspaceBasePath).Msg("Error walking workspace path")
	}
}

// processSingleScript processes a single script and updates it if necessary.
func (sp *ScriptProcessor) processSingleScript(ctx context.Context, scriptPath string) {
	log.Info().Str("path", scriptPath).Msg("Processing script")

	localContent, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Error().Err(err).Str("path", scriptPath).Msg("Failed to read local script")
		return
	}

	upstreamPath := sp.remapLocalToUpstream(scriptPath)
	upstreamContent, err := sp.githubClient.FetchFileContent(ctx, sp.repoOwner, sp.repoName, upstreamPath)
	if err != nil {
		log.Error().Err(err).Str("upstream_path", upstreamPath).Msg("Failed to fetch upstream script")
		return
	}

	updatedContent, hasChanges := sp.updateScript(string(localContent), upstreamContent)
	if hasChanges {
		err := os.WriteFile(scriptPath, []byte(updatedContent), 0644)
		if err != nil {
			log.Error().Err(err).Str("path", scriptPath).Msg("Failed to write updated script")
		} else {
			log.Info().Str("path", scriptPath).Msg("Script updated with detected changes")
		}
	} else {
		log.Info().Str("path", scriptPath).Msg("No changes detected in script")
	}
}

// remapLocalToUpstream maps a local script path to its upstream GitHub path.
func (sp *ScriptProcessor) remapLocalToUpstream(localPath string) string {
	upstreamPath := strings.ReplaceAll(localPath, `\`, `/`)
	upstreamPath = strings.TrimPrefix(upstreamPath, sp.workspaceBasePath+"/")
	log.Debug().Str("local_path", localPath).Str("upstream_path", upstreamPath).Msg("Mapped local to upstream path")
	return upstreamPath
}

// updateScript checks for URL and variable changes and applies updates to the script content.
func (sp *ScriptProcessor) updateScript(localContent, upstreamContent string) (string, bool) {
	localURLs := extractURLs(localContent)
	upstreamURLs := extractURLs(upstreamContent)
	urlsChanged := checkURLChanges(localURLs, upstreamURLs)

	updatedContent, varsChanged := updateCustomVariables(localContent, upstreamContent, sp.variablesToCheck)

	return updatedContent, urlsChanged || varsChanged
}

// extractURLs extracts valid URLs from the script content.
func extractURLs(scriptContent string) []string {
	allURLs := urlRegex.FindAllString(scriptContent, -1)
	var validURLs []string
	for _, url := range allURLs {
		if !strings.Contains(url, "${") {
			validURLs = append(validURLs, url)
		}
	}
	return validURLs
}

// checkURLChanges detects URL changes between local and upstream scripts.
func checkURLChanges(localURLs, upstreamURLs []string) bool {
	changed := false
	for _, localURL := range localURLs {
		bestMatch, matchLength := findLongestMatchingURL(localURL, upstreamURLs)
		if matchLength != len(localURL) || matchLength != len(bestMatch) {
			log.Warn().Str("local_url", localURL).Str("best_match", bestMatch).Msg("URL has changed")
			changed = true
		}
	}
	return changed
}

// updateCustomVariables updates specific variables in the script content.
func updateCustomVariables(localContent, upstreamContent string, variables []string) (string, bool) {
	localVars := extractVariables(localContent, variables)
	upstreamVars := extractVariables(upstreamContent, variables)
	changed := false

	for name, upstreamValue := range upstreamVars {
		if localValue, exists := localVars[name]; exists && localValue != upstreamValue {
			localContent = updateVariable(localContent, name, upstreamValue)
			changed = true
		}
	}
	return localContent, changed
}

// extractVariables extracts specified variables from script content.
func extractVariables(content string, variables []string) map[string]string {
	values := make(map[string]string)
	for _, name := range variables {
		pattern := fmt.Sprintf(VariablePatternStr, regexp.QuoteMeta(name))
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			values[name] = match[1]
		}
	}
	return values
}

// updateVariable updates a single variable in the script content.
func updateVariable(content, name, value string) string {
	pattern := fmt.Sprintf(`(?m)^%s\s*=\s*['"]?([^'"\s]+)['"]?`, regexp.QuoteMeta(name))
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, fmt.Sprintf(`%s='%s'`, name, value))
}

// findLongestMatchingURL identifies the best matching URL based on prefix length.
func findLongestMatchingURL(localURL string, upstreamURLs []string) (string, int) {
	bestMatch := ""
	longestMatch := 0
	for _, url := range upstreamURLs {
		matchLength := len(strings.SplitAfterN(localURL, url, 2))
		if matchLength > longestMatch {
			bestMatch = url
			longestMatch = matchLength
		}
	}
	return bestMatch, longestMatch
}
