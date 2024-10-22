package shadowhub

import (
	"fmt"
	"github.com/google/go-github/v43/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Function to create a GitHub client with an optional token for authentication
func createShadowHubClient(token string) *github.Client {
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		return github.NewClient(tc)
	}
	return github.NewClient(nil) // Unauthenticated client (lower rate limits)
}

// Helper to fetch the content of a GitHub file
func fetchShadowHubFile(client *github.Client, owner, repo, path string) (string, error) {
	content, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		return "", err
	}
	fileContent, err := content.GetContent()
	if err != nil {
		return "", err
	}
	return fileContent, nil
}

// Helper function to extract URLs from script, excluding those with variables
func extractURLs(scriptContent string) []string {
	// Step 1: Match all URLs using regex
	re := regexp.MustCompile(`https?://[^\s]+`)
	allURLs := re.FindAllString(scriptContent, -1)

	// Step 2: Filter out URLs that contain "${}"
	var validURLs []string
	for _, url := range allURLs {
		// Exclude URLs that contain "${}"
		if !regexp.MustCompile(`\$\{[^}]+\}`).MatchString(url) {
			validURLs = append(validURLs, url)
		}
	}

	return validURLs
}

// Helper function to extract variables from script content
func extractVariables(scriptContent string, variableNames []string) map[string]string {
	variableValues := make(map[string]string)

	for _, variableName := range variableNames {
		// Regex to match variables with single quotes, double quotes, or no quotes
		pattern := fmt.Sprintf(`(?m)^\s*%s\s*=\s*['"]?([^'"\s]+)['"]?`, variableName)
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(scriptContent)

		if len(match) > 1 {
			variableValues[variableName] = match[1] // Extract the variable's value
		}
	}

	return variableValues
}

// Function to compute the longest common prefix of two strings
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

// Function to find the upstream URL with the longest match
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

	return bestMatch, longestMatchLength
}

// Function to check if URLs have changed by finding the longest matching URL for each local URL
func checkURLChanges(localURLs, upstreamURLs []string) bool {
	changed := false

	for _, localURL := range localURLs {
		bestMatch, matchLength := findLongestMatchingURL(localURL, upstreamURLs)
		if matchLength == len(localURL) && matchLength == len(bestMatch) {
		} else if matchLength > 0 {
			changed = true
		} else {
			fmt.Printf("No match found for local URL: %s\n", localURL)
			changed = true
		}
	}

	// Now check for upstream URLs not present in local
	for _, upstreamURL := range upstreamURLs {
		bestMatch, matchLength := findLongestMatchingURL(upstreamURL, localURLs)
		if matchLength == len(upstreamURL) && matchLength == len(bestMatch) {
			continue // Perfect match exists
		} else if matchLength > 8 { //Larger 8 to exclude 'https://' from match
			fmt.Printf("New upstream URL found: %s\n", upstreamURL)
			fmt.Println("Match Length %n\n", matchLength)
			changed = true
		}
	}
	return changed
}

// Helper function to update a specific variable in the script
func updateVariable(scriptContent, variableName, newValue string) string {
	// Regex to match variables with single quotes, double quotes, or no quotes
	variablePattern := fmt.Sprintf(`(?m)^%s\s*=\s*['"]?([^'"\s]+)['"]?`, variableName)
	re := regexp.MustCompile(variablePattern)

	// Replace the value of the variable with the new one, keeping the quotes consistent
	updatedScript := re.ReplaceAllString(scriptContent, fmt.Sprintf(`%s='%s'`, variableName, newValue))

	return updatedScript
}

// Function to update custom variables based on upstream changes
func updateCustomVariables(localScriptContent, upstreamScriptContent string, variableNames []string) (string, bool) {
	changed := false
	localVars := extractVariables(localScriptContent, variableNames)
	upstreamVars := extractVariables(upstreamScriptContent, variableNames)

	for variableName, upstreamValue := range upstreamVars {
		if localValue, exists := localVars[variableName]; exists && localValue != upstreamValue {
			fmt.Printf("Updating %s: %s -> %s\n", variableName, localValue, upstreamValue)
			localScriptContent = updateVariable(localScriptContent, variableName, upstreamValue)
			changed = true
		}
	}
	return localScriptContent, changed
}

// Path remapping function to map local paths to upstream paths
func remapLocalPathToUpstream(localPath, workspaceImageFilePath string) string {

	// Ensure that the resulting path uses forward slashes for GitHub URLs
	upstreamPath := strings.ReplaceAll(localPath, `\`, `/`) // Normalize slashes for GitHub

	// Remove the "workspace-core-image/" prefix from the local path
	upstreamPath = strings.TrimPrefix(upstreamPath, "workspace-core-image/")

	return upstreamPath
}

// Function to update dependencies for all scripts in a workspace
func UpdateShadowDependencies(workspaceImageFilePath, token string) {
	// Variables to check
	variablesToCheck := []string{"COMMIT_ID", "BRANCH", "KASMVNC_VER", "SQUID_COMMIT"}

	// GitHub repository details
	owner := "kasmtech"
	repo := "workspaces-core-images"

	// Create GitHub client (with optional token)
	client := createShadowHubClient(token)

	// Iterate over all .sh scripts in the workspace directory
	err := filepath.Walk(workspaceImageFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only consider .sh files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sh") {
			fmt.Printf("Processing script: %s\n", path)

			// Read local script content
			localScriptContent, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading local script: %v\n", err)
				return nil
			}

			// Remap local path to upstream path
			upstreamPath := remapLocalPathToUpstream(path, workspaceImageFilePath)

			// Fetch the upstream script
			upstreamScriptContent, err := fetchShadowHubFile(client, owner, repo, upstreamPath)
			if err != nil {
				fmt.Printf("Error fetching upstream script: %v\n", err)
				return nil
			}

			// Check URLs
			localURLs := extractURLs(string(localScriptContent))
			upstreamURLs := extractURLs(upstreamScriptContent)
			urlsChanged := checkURLChanges(localURLs, upstreamURLs)

			// Update custom variables if they changed upstream
			updatedScriptContent, varsChanged := updateCustomVariables(string(localScriptContent), upstreamScriptContent, variablesToCheck)

			// If either URLs or variables have changed, update the script
			if urlsChanged || varsChanged {
				err = ioutil.WriteFile(path, []byte(updatedScriptContent), 0644)
				if err != nil {
					fmt.Printf("Error writing updated script: %v\n", err)
					return nil
				}
				fmt.Printf("Script updated: %s\n", path)
			} else {
				fmt.Printf("No changes detected in script: %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", workspaceImageFilePath, err)
	}
}
