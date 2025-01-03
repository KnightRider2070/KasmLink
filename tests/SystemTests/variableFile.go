package SystemTests

var alternateValues = true

var baseUrl = "https://192.168.56.103"
var apiSecret = "FIRsIHNjz6UG"
var apiKeySecret = "5mPp4pyXrS32OmLQKddPZ3cljvHfga8s"
var hostIp = "192.168.56.103"
var username = "thor"
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
		username = altUser
		password = altPassword
		knwHosts = altKnwHosts
	}
}

//Create ssh config
//sshConfig, _ := shadowssh.NewSSHConfig(username, password, hostIp, 22, knwHosts, 10*time.Second)

//Create KASM API
//kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)
