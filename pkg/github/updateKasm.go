// shadowhub/shadowhub.go
package shadowhub

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Constants for variable extraction and URL patterns.
const (
	// URLRegexPattern matches HTTP and HTTPS URLs.
	URLRegexPattern = `https?://[^\s]+`
	// VariablePatternStr is a format string for variable extraction regex.
	VariablePatternStr = `(?m)^\s*%s\s*=\s*['"]?([^'"\s]+)['"]?`
)

// Precompiled regular expressions for performance.
var (
	urlRegex = regexp.MustCompile(URLRegexPattern)
)

// BuildLog represents the structure of Docker build log messages.
// Used for parsing JSON-formatted build logs.
type BuildLog struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

// DockerClient encapsulates the Docker client and retry configurations.
type DockerClient struct {
	cli               *github.Client
	retries           int
	initialRetryDelay int
	backoffMultiplier int
	maxRetryDelay     int
	jitterFactor      float64

	// Mutex to protect any future mutable state
	mu sync.RWMutex
}

// NewDockerClient initializes and returns a new DockerClient.
// It sets default values if provided configurations are zero-valued.
// Parameters:
// - cli: The GitHub client instance.
// - retries: Number of retry attempts for operations.
// - initialRetryDelay: Initial delay before retrying an operation (in seconds).
// - backoffMultiplier: Multiplier for exponential backoff.
// - maxRetryDelay: Maximum delay between retries (in seconds).
// - jitterFactor: Factor for adding jitter to retry delays.
func NewDockerClient(cli *github.Client, retries, initialRetryDelay, backoffMultiplier, maxRetryDelay int, jitterFactor float64) *DockerClient {
	if retries <= 0 {
		retries = 3
	}
	if initialRetryDelay <= 0 {
		initialRetryDelay = 2
	}
	if backoffMultiplier <= 0 {
		backoffMultiplier = 2
	}
	if maxRetryDelay <= 0 {
		maxRetryDelay = 16
	}
	if jitterFactor <= 0 {
		jitterFactor = 0.1 // 10% jitter
	}

	return &DockerClient{
		cli:               cli,
		retries:           retries,
		initialRetryDelay: initialRetryDelay,
		backoffMultiplier: backoffMultiplier,
		maxRetryDelay:     maxRetryDelay,
		jitterFactor:      jitterFactor,
	}
}

// createShadowHubClient creates a GitHub client with an optional token for authentication.
// Parameters:
// - token: GitHub OAuth token. If empty, an unauthenticated client is returned.
// Returns:
// - A GitHub client instance.
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
// Parameters:
// - client: GitHub client.
// - owner: Repository owner.
// - repo: Repository name.
// - path: File path within the repository.
// Returns:
// - Content of the file as a string.
// - An error if the fetch fails.
func fetchShadowHubFile(client *github.Client, owner, repo, path string) (string, error) {
	log.Info().
		Str("owner", owner).
		Str("repo", repo).
		Str("path", path).
		Msg("Fetching file from GitHub")

	content, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Error().
			Err(err).
			Str("owner", owner).
			Str("repo", repo).
			Str("path", path).
			Msg("Failed to fetch file from GitHub")
		return "", fmt.Errorf("failed to fetch file from GitHub: %w", err)
	}

	fileContent, err := content.GetContent()
	if err != nil {
		log.Error().
			Err(err).
			Str("owner", owner).
			Str("repo", repo).
			Str("path", path).
			Msg("Failed to decode file content from GitHub")
		return "", fmt.Errorf("failed to decode file content from GitHub: %w", err)
	}

	log.Debug().Msg("Fetched file content successfully")
	return fileContent, nil
}

// extractURLs extracts URLs from script content, excluding those with variables.
// Parameters:
// - scriptContent: The content of the script as a string.
// Returns:
// - A slice of valid URLs.
func extractURLs(scriptContent string) []string {
	allURLs := urlRegex.FindAllString(scriptContent, -1)

	var validURLs []string
	for _, url := range allURLs {
		if !strings.Contains(url, "${") {
			validURLs = append(validURLs, url)
		}
	}

	log.Debug().
		Int("url_count", len(validURLs)).
		Msg("Extracted valid URLs from script content")
	return validURLs
}

// extractVariables extracts specified variables from script content.
// Parameters:
// - scriptContent: The content of the script as a string.
// - variableNames: A slice of variable names to extract.
// Returns:
// - A map of variable names to their extracted values.
func extractVariables(scriptContent string, variableNames []string) map[string]string {
	variableValues := make(map[string]string)
	for _, variableName := range variableNames {
		pattern := fmt.Sprintf(VariablePatternStr, regexp.QuoteMeta(variableName))
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(scriptContent)
		if len(match) > 1 {
			variableValues[variableName] = match[1]
		}
	}
	log.Debug().
		Int("variable_count", len(variableValues)).
		Msg("Extracted variables from script content")
	return variableValues
}

// longestCommonPrefix calculates the longest common prefix length between two strings.
// Parameters:
// - s1: First string.
// - s2: Second string.
// Returns:
// - The length of the longest common prefix.
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

// findLongestMatchingURL finds the upstream URL with the longest common prefix match.
// Parameters:
// - localURL: The local URL to match.
// - upstreamURLs: A slice of upstream URLs to compare against.
// Returns:
// - The best matching upstream URL.
// - The length of the match.
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
	log.Debug().
		Str("local_url", localURL).
		Str("best_match", bestMatch).
		Int("match_length", longestMatchLength).
		Msg("Found longest matching URL")
	return bestMatch, longestMatchLength
}

// checkURLChanges checks for URL changes by comparing local URLs with upstream URLs.
// Parameters:
// - localURLs: Slice of local URLs extracted from the script.
// - upstreamURLs: Slice of upstream URLs fetched from GitHub.
// Returns:
// - A boolean indicating whether any URL changes were detected.
func checkURLChanges(localURLs, upstreamURLs []string) bool {
	changed := false
	for _, localURL := range localURLs {
		bestMatch, matchLength := findLongestMatchingURL(localURL, upstreamURLs)
		if matchLength != len(localURL) || matchLength != len(bestMatch) {
			log.Warn().
				Str("local_url", localURL).
				Str("best_match", bestMatch).
				Msg("URL has changed")
			changed = true
		}
	}
	for _, upstreamURL := range upstreamURLs {
		_, matchLength := findLongestMatchingURL(upstreamURL, localURLs)
		if matchLength <= 8 { // Arbitrary threshold for new URLs
			log.Info().
				Str("upstream_url", upstreamURL).
				Msg("New upstream URL found")
			changed = true
		}
	}
	return changed
}

// updateVariable updates a specific variable in the script content.
// Parameters:
// - scriptContent: The original script content.
// - variableName: The name of the variable to update.
// - newValue: The new value for the variable.
// Returns:
// - The updated script content.
func updateVariable(scriptContent, variableName, newValue string) string {
	variablePattern := fmt.Sprintf(`(?m)^%s\s*=\s*['"]?([^'"\s]+)['"]?`, regexp.QuoteMeta(variableName))
	re := regexp.MustCompile(variablePattern)
	updatedScript := re.ReplaceAllString(scriptContent, fmt.Sprintf(`%s='%s'`, variableName, newValue))
	log.Debug().
		Str("variable_name", variableName).
		Str("new_value", newValue).
		Msg("Updated variable in script content")
	return updatedScript
}

// updateCustomVariables updates custom variables based on upstream changes.
// Parameters:
// - localScriptContent: The content of the local script.
// - upstreamScriptContent: The content of the upstream script.
// - variableNames: Slice of variable names to update.
// Returns:
// - The updated script content.
// - A boolean indicating whether any variables were changed.
func updateCustomVariables(localScriptContent, upstreamScriptContent string, variableNames []string) (string, bool) {
	changed := false
	localVars := extractVariables(localScriptContent, variableNames)
	upstreamVars := extractVariables(upstreamScriptContent, variableNames)
	for variableName, upstreamValue := range upstreamVars {
		if localValue, exists := localVars[variableName]; exists && localValue != upstreamValue {
			log.Info().
				Str("variable_name", variableName).
				Str("old_value", localValue).
				Str("new_value", upstreamValue).
				Msg("Variable value updated")
			localScriptContent = updateVariable(localScriptContent, variableName, upstreamValue)
			changed = true
		}
	}
	return localScriptContent, changed
}

// remapLocalPathToUpstream maps local paths to upstream paths.
// Parameters:
// - localPath: The local file path.
// - workspaceImageFilePath: The base path of the workspace image files.
// Returns:
// - The corresponding upstream file path.
func remapLocalPathToUpstream(localPath, workspaceImageFilePath string) string {
	upstreamPath := strings.ReplaceAll(localPath, `\`, `/`)
	upstreamPath = strings.TrimPrefix(upstreamPath, "workspace-core-image/")
	log.Debug().
		Str("local_path", localPath).
		Str("upstream_path", upstreamPath).
		Msg("Remapped local path to upstream path")
	return upstreamPath
}

// UpdateShadowDependencies updates dependencies for all scripts in a workspace.
// It fetches the corresponding upstream scripts from GitHub, checks for URL and variable changes,
// and updates the local scripts accordingly.
// Parameters:
// - workspaceImageFilePath: The base path of the workspace image files.
// - token: GitHub OAuth token for authenticated requests.
// Returns:
// - None. Updates are performed in-place.
func UpdateShadowDependencies(workspaceImageFilePath, token string) {
	variablesToCheck := []string{"COMMIT_ID", "BRANCH", "KASMVNC_VER", "SQUID_COMMIT"}
	owner := "kasmtech"
	repo := "workspaces-core-images"

	client := createShadowHubClient(token)
	err := filepath.Walk(workspaceImageFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Error walking directory path")
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sh") {
			log.Info().
				Str("path", path).
				Msg("Processing script")

			localScriptContent, err := os.ReadFile(path)
			if err != nil {
				log.Error().
					Err(err).
					Str("path", path).
					Msg("Failed to read local script")
				return nil // Continue processing other files
			}

			upstreamPath := remapLocalPathToUpstream(path, workspaceImageFilePath)
			upstreamScriptContent, err := fetchShadowHubFile(client, owner, repo, upstreamPath)
			if err != nil {
				log.Error().
					Err(err).
					Str("upstream_path", upstreamPath).
					Msg("Failed to fetch upstream script")
				return nil // Continue processing other files
			}

			localURLs := extractURLs(string(localScriptContent))
			upstreamURLs := extractURLs(upstreamScriptContent)
			urlsChanged := checkURLChanges(localURLs, upstreamURLs)
			updatedScriptContent, varsChanged := updateCustomVariables(string(localScriptContent), upstreamScriptContent, variablesToCheck)

			if urlsChanged || varsChanged {
				err = os.WriteFile(path, []byte(updatedScriptContent), info.Mode())
				if err != nil {
					log.Error().
						Err(err).
						Str("path", path).
						Msg("Failed to write updated script")
					return nil // Continue processing other files
				}
				log.Info().
					Str("path", path).
					Msg("Script updated due to detected changes")
			} else {
				log.Info().
					Str("path", path).
					Msg("No changes detected in script")
			}
		}
		return nil
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("workspace_path", workspaceImageFilePath).
			Msg("Error walking workspace path")
	}
}
