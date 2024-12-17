package Tests

var alternateValues = true

var baseUrl = "https://192.168.56.103"
var apiSecret = "rBmnE9mHLmzd"
var apiKeySecret = "8A1xrZLhm14WFuk9Gf9AerVMljdPzIhk"
var hostIp = "192.168.56.103"
var user = "thor"
var password = "stark"
var knwHosts = "C:\\Users\\cjhue\\.ssh\\known_hosts"

// Alternate Values
var altBaseUrl = "https://192.168.120.5"
var altApiSecret = "QpuCkIfRl2Mj"
var altApiKeySecret = "fsd6xERiqRXoFTAPXhiRSX20s4gBdiwN"
var altHostIp = "192.168.120.5"
var altUser = "thor"
var altPassword = "stark"
var altKnwHosts = "C:\\Users\\thor\\.ssh\\known_hosts"

func init() {
	if alternateValues {
		baseUrl = altBaseUrl
		apiSecret = altApiSecret
		apiKeySecret = altApiKeySecret
		hostIp = altHostIp
		user = altUser
		password = altPassword
		knwHosts = altKnwHosts
	}
}

//Create ssh config
//sshConfig, _ := sshmanager.NewSSHConfig(user, password, hostIp, 22, knwHosts, 10*time.Second)

//Create KASM API
//kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)
