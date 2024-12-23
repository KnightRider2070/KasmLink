package SystemTests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/webApi"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

}
func TestGetUser(t *testing.T) {
	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	getUser, err := kApi.GetUser(ctx, response.UserID, "")

	assert.NoError(t, err)

	assert.Equal(t, response.UserID, getUser.UserID)
	assert.Equal(t, response.Username, username)
	assert.Equal(t, response.FirstName, first_name)
	assert.Equal(t, response.LastName, last_name)
	assert.Equal(t, response.Organization, "")
}

func TestGetUsers(t *testing.T) {
	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	users, err := kApi.GetUsers(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	users, err = kApi.GetUsers(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(users))

}

func TestUpdateUser(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	updatedUser := webApi.TargetUser{
		UserID:    response.UserID,
		Username:  username,
		LastName:  first_name,
		FirstName: last_name,
	}

	updateUser, err := kApi.UpdateUser(ctx, updatedUser)

	assert.NoError(t, err)

	assert.Equal(t, username, updateUser.Username)
	assert.Equal(t, first_name, updateUser.LastName)
	assert.Equal(t, last_name, updateUser.FirstName)
	assert.Equal(t, phone, updateUser.Phone)
	assert.NotEmpty(t, updateUser.UserID)

}

func TestDeleteUser(t *testing.T) {
	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	users, err := kApi.GetUsers(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(users))

	err = kApi.DeleteUser(ctx, response.UserID, true)

	assert.NoError(t, err)

	users, err = kApi.GetUsers(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))

}

func TestGetUserAttributes(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	attributes, err := kApi.GetUserAttributes(ctx, response.UserID)
	assert.NoError(t, err)
	assert.Equal(t, response.UserID, attributes.UserID)
	assert.Equal(t, "", attributes.DefaultImageId)

}

func TestUpdateUserAttributes(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	attributes, err := kApi.GetUserAttributes(ctx, response.UserID)
	assert.NoError(t, err)
	assert.Equal(t, response.UserID, attributes.UserID)
	assert.Equal(t, "", attributes.DefaultImageId)

	// Modify user attribute
	attributes.DefaultImageId = "6a335ca1505a4e0eb966930823bcc691"

	err = kApi.UpdateUserAttributes(ctx, *attributes)
	assert.NoError(t, err)

	attributes, err = kApi.GetUserAttributes(ctx, response.UserID)
	assert.NoError(t, err)
	assert.Equal(t, response.UserID, attributes.UserID)
	assert.Equal(t, "6a335ca1505a4e0eb966930823bcc691", attributes.DefaultImageId)

}

func TestLogoutUser(t *testing.T) {
	//TODO: Think about how to test the for logout, idea maybe request a kasm then logout and then check if the kasm is terminated.
}

func hasGroupWithID(user webApi.UserResponse, targetGroupID string) bool {
	for _, group := range user.Groups {
		if group.GroupID == targetGroupID {
			return true
		}
	}
	return false
}

func TestAddUserToGroup(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	adminGroupId := "65ae90f8aebf46f29993b52c580364b8"

	err = kApi.AddUserToGroup(ctx, response.UserID, adminGroupId)
	assert.NoError(t, err)

	userGet, err := kApi.GetUser(ctx, response.UserID, "")

	assert.True(t, hasGroupWithID(*userGet, adminGroupId))
}

func TestRemoveUserFromGroup(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	adminGroupId := "65ae90f8aebf46f29993b52c580364b8"

	err = kApi.AddUserToGroup(ctx, response.UserID, adminGroupId)
	assert.NoError(t, err)

	userGet, err := kApi.GetUser(ctx, response.UserID, "")

	assert.True(t, hasGroupWithID(*userGet, adminGroupId))

	err = kApi.RemoveUserFromGroup(ctx, response.UserID, adminGroupId)

	assert.NoError(t, err)
}

func TestGenerateLoginLink(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	username := "neo42"
	first_name := "Luke"
	last_name := "Skywalker"
	phone := "1701"
	password := "redpill42"

	user := webApi.TargetUser{
		Username:  username,
		FirstName: first_name,
		LastName:  last_name,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(ctx, user)

	assert.NoError(t, err)

	assert.Equal(t, username, response.Username)
	assert.Equal(t, first_name, response.FirstName)
	assert.Equal(t, last_name, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	link, err := kApi.GenerateLoginLink(ctx, response.UserID)

	assert.NoError(t, err)

	assert.NotEmpty(t, link)

}
