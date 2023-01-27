package gentest

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev_forum-user/pkg/entity"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	v := make([]rune, length)
	for i := range v {
		v[i] = letters[rand.Intn(len(letters))]
	}
	return string(v)
}

// RandomUser panics on hardware error.
// It should be used ONLY for testing.
func RandomUser(nameLen, emailLen, passLen int) entity.User {
	id := uuid.Must(uuid.NewV4())

	return entity.User{
		Id:        id.String(),
		Name:      RandomString(nameLen),
		Email:     RandomString(emailLen),
		Password:  RandomString(passLen),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// RandomEntity returns a random Entity marshaled
// to JSON and panics on error.
// It should be used ONLY for testing.
func RandomJSONUser(titleLen, emailLen, passLen int) []byte {
	json, err := json.Marshal(RandomUser(titleLen, emailLen, passLen))
	if err != nil {
		panic(err)
	}
	return json
}
