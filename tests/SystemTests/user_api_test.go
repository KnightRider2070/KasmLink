package SystemTests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"
	"kasmlink/pkg/api/user"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, true)
	kApi := user.NewUserService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	username := "neo42"
	firstName := "Luke"
	lastName := "Skywalker"
	phone := "1701"
	password := "redpill42"

	targetUser := models.TargetUser{
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(targetUser)
	assert.NoError(t, err)
	assert.Equal(t, username, response.Username)
	assert.Equal(t, firstName, response.FirstName)
	assert.Equal(t, lastName, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)
}

func TestGetUser(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, true)
	kApi := user.NewUserService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	username := "neo42"
	firstName := "Luke"
	lastName := "Skywalker"
	phone := "1701"
	password := "redpill42"

	targetUser := models.TargetUser{
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(targetUser)
	assert.NoError(t, err)
	assert.Equal(t, username, response.Username)
	assert.Equal(t, firstName, response.FirstName)
	assert.Equal(t, lastName, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	getUser, err := kApi.GetUser(response.UserID, "")
	assert.NoError(t, err)
	assert.Equal(t, response.UserID, getUser.UserID)
	assert.Equal(t, username, getUser.Username)
	assert.Equal(t, firstName, getUser.FirstName)
	assert.Equal(t, lastName, getUser.LastName)
}

func TestUpdateUser(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, true)
	kApi := user.NewUserService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	username := "neo42"
	firstName := "Luke"
	lastName := "Skywalker"
	phone := "1701"
	password := "redpill42"

	targetUser := models.TargetUser{
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(targetUser)
	assert.NoError(t, err)
	assert.Equal(t, username, response.Username)
	assert.Equal(t, firstName, response.FirstName)
	assert.Equal(t, lastName, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	updatedUser := models.TargetUser{
		UserID:    response.UserID,
		Username:  username,
		FirstName: lastName,
		LastName:  firstName,
	}

	updateResponse, err := kApi.UpdateUser(updatedUser)
	assert.NoError(t, err)
	assert.Equal(t, username, updateResponse.Username)
	assert.Equal(t, firstName, updateResponse.LastName)
	assert.Equal(t, lastName, updateResponse.FirstName)
}

func TestDeleteUser(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, true)
	kApi := user.NewUserService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	username := "neo42"
	firstName := "Luke"
	lastName := "Skywalker"
	phone := "1701"
	password := "redpill42"

	targetUser := models.TargetUser{
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Disabled:  false,
		Phone:     phone,
		Password:  password,
	}

	response, err := kApi.CreateUser(targetUser)
	assert.NoError(t, err)
	assert.Equal(t, username, response.Username)
	assert.Equal(t, firstName, response.FirstName)
	assert.Equal(t, lastName, response.LastName)
	assert.Equal(t, phone, response.Phone)
	assert.NotEmpty(t, response.UserID)

	err = kApi.DeleteUser(response.UserID, true)
	assert.NoError(t, err)
}
