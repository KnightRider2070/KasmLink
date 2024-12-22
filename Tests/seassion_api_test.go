package Tests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/webApi"
	"testing"
	"time"
)

//Create ssh config
//sshConfig, _ := sshmanager.NewSSHConfig("thor", "stark", "192.168.120.5", 22, "C:\\Users\\Thor\\.ssh\\known_hosts", 10*time.Second)

//Create KASM API
//kApi := webApi.NewKasmAPI("https://192.168.120.5", "C6QmU5ohTUIE", "91MRn9E7FyBSPJ5HtexWrubIG3SYLkB5", true, 50*time.Second)

func TestRequestKasm(t *testing.T) {

	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 100*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	userID := "a2b9d2932e484280bc0a64822a5c8d42" //testuser

	imageID := "6a335ca1505a4e0eb966930823bcc691" //Brave

	envArgs := map[string]string{
		"ENV_VAR": "value",
	}

	kasmResponse, err := kApi.RequestKasmSession(ctx, userID, imageID, envArgs)
	if err != nil {
		return
	}

	assert.Contains(t, kasmResponse.Status, "starting")
	assert.Contains(t, kasmResponse.UserID, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, kasmResponse.KasmID)
	assert.NotEmpty(t, kasmResponse.SessionToken)
}

func TestGetKasmStatus(t *testing.T) {

	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 100*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	userID := "a2b9d2932e484280bc0a64822a5c8d42" //testuser

	imageID := "6a335ca1505a4e0eb966930823bcc691" //Brave

	envArgs := map[string]string{
		"ENV_VAR": "value",
	}

	kasmResponse, err := kApi.RequestKasmSession(ctx, userID, imageID, envArgs)
	if err != nil {
		return
	}

	assert.Contains(t, kasmResponse.Status, "starting")
	assert.Contains(t, kasmResponse.UserID, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, kasmResponse.KasmID)
	assert.NotEmpty(t, kasmResponse.SessionToken)

	status, err := kApi.GetKasmStatus(ctx, kasmResponse.UserID, kasmResponse.KasmID, true)
	if err != nil {
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, status.Kasm.ImageID, imageID)

}

func TestDestroyKasmSession(t *testing.T) {

	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 100*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	userID := "a2b9d2932e484280bc0a64822a5c8d42" //testuser

	imageID := "6a335ca1505a4e0eb966930823bcc691" //Brave

	envArgs := map[string]string{
		"ENV_VAR": "value",
	}

	kasmResponse, err := kApi.RequestKasmSession(ctx, userID, imageID, envArgs)
	if err != nil {
		return
	}

	assert.Contains(t, kasmResponse.Status, "starting")
	assert.Contains(t, kasmResponse.UserID, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, kasmResponse.KasmID)
	assert.NotEmpty(t, kasmResponse.SessionToken)

	status, err := kApi.GetKasmStatus(ctx, kasmResponse.UserID, kasmResponse.KasmID, true)
	if err != nil {
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, status.Kasm.ImageID, imageID)

	err = kApi.DestroyKasmSession(ctx, kasmResponse.KasmID, userID)

	assert.NoError(t, err)

}

func TestExecCommand(t *testing.T) {

}
