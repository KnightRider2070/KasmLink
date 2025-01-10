package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Directory structure constants
var (
	// CacheDir Cache Directory - Main directory for all cache-related files
	CacheDir = filepath.Join(GetUserCacheDir(), "cluster_cache")

	ClusterName = "main" // Default cluster name

	// ClusterDirectory Cluster Directory - Main directory for cluster files
	ClusterDirectory = filepath.Join("./clusters")

	// CachePaths Cache subdirectories
	CachePaths = struct {
		HelmChartCache     string // Stores Helm charts
		KubernetesCache    string // Stores Kubernetes-related resources
		BaseConfigCache    string // Stores base configurations
		TemporaryCache     string // General temporary cache
		PatchCache         string // Stores patch files
		DocumentationCache string // Stores documentation files
	}{
		HelmChartCache:     filepath.Join(CacheDir, "helm_chart_cache"),
		KubernetesCache:    filepath.Join(CacheDir, "kubernetes_resources_cache"),
		BaseConfigCache:    filepath.Join(CacheDir, "base_configuration_cache"),
		TemporaryCache:     filepath.Join(CacheDir, "temporary_cache"),
		PatchCache:         filepath.Join(CacheDir, "patch_cache"),
		DocumentationCache: filepath.Join(CacheDir, "documentation_cache"),
	}

	// ConfigPaths Cluster and Talos-related paths
	ConfigPaths = struct {
		ClusterEnvFilePath string // Path to the environment config file
		TalosConfigFile    string // Path to Talos config file
		TalosDir           string // Path to Talos directory
		KubernetesDir      string // Path to Kubernetes directory
		GeneratedDir       string // Path to the directory containing generated files
		SecretsFile        string // Path to the secrets file
		AgeKeyFile         string // Path to the Age key file
		SopsFile           string // Path to the Sopss file
	}{
		ClusterEnvFilePath: filepath.Join(ClusterDirectory, "clusterenv.yaml"),
		TalosConfigFile:    filepath.Join(ClusterDirectory, "talos", "talconfig.yaml"),
		TalosDir:           filepath.Join(ClusterDirectory, "talos"),
		KubernetesDir:      filepath.Join(ClusterDirectory, "kubernetes"),
		GeneratedDir:       filepath.Join(ClusterDirectory, "talos", "generated"),
		SecretsFile:        filepath.Join(ClusterDirectory, "talos", "generated", "talsecret.yaml"),
		AgeKeyFile:         filepath.Join(ClusterDirectory, "talos", "age.agekey"),
		SopsFile:           filepath.Join(ClusterDirectory, ".sops.yaml"),
	}

	// IndexCacheDir Miscellaneous paths
	IndexCacheDir   = "./index_cache" // Indexing cache directory
	GpgKeyDirectory = ".cr-gpg"       // Directory for storing GPG keys

	// AllClusterIPs IP Lists
	AllClusterIPs   = []string{} // All cluster IP addresses
	ControlPlaneIPs = []string{} // IPs for control plane nodes (master nodes)
	WorkerNodeIPs   = []string{} // IPs for worker nodes (non-master nodes)
)

// GetUserCacheDir returns the user's cache directory for the application.
func GetUserCacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Println("Error getting user cache directory:", err)
		return "./" // Fallback to the current directory if there is an error
	}
	return userCacheDir
}

// PrintPaths outputs the directory paths to the console for easy inspection (could be useful for debugging).
func PrintPaths() {
	fmt.Println("Cache Directory:", CacheDir)
	fmt.Println("Helm Chart Cache Path:", CachePaths.HelmChartCache)
	fmt.Println("Kubernetes Cache Path:", CachePaths.KubernetesCache)
	fmt.Println("Base Config Cache Path:", CachePaths.BaseConfigCache)
	fmt.Println("Temporary Cache Path:", CachePaths.TemporaryCache)
	fmt.Println("Patch Cache Path:", CachePaths.PatchCache)
	fmt.Println("Documentation Cache Path:", CachePaths.DocumentationCache)
	fmt.Println("Cluster Directory:", ClusterDirectory)
	fmt.Println("Cluster Env File Path:", ConfigPaths.ClusterEnvFilePath)
	fmt.Println("Talos Config File Path:", ConfigPaths.TalosConfigFile)
	fmt.Println("Talos Directory Path:", ConfigPaths.TalosDir)
	fmt.Println("Kubernetes Directory Path:", ConfigPaths.KubernetesDir)
	fmt.Println("Generated Directory Path:", ConfigPaths.GeneratedDir)
	fmt.Println("Talos Secret File Path:", ConfigPaths.SecretsFile)
	fmt.Println("Index Cache Directory:", IndexCacheDir)
	fmt.Println("GPG Key Directory:", GpgKeyDirectory)
}
