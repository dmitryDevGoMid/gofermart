package authentication

import (
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	User repository.User
	jwt.RegisteredClaims
}

type SetJWTToken struct{}

func (chain SetJWTToken) Process(result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	expirationTimeSecond := 60

	expirationTime := time.Now().Add(time.Duration(expirationTimeSecond) * time.Second)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		User: data.User.User,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		//w.WriteHeader(http.StatusInternalServerError)
		return []pipeline.Message{data}, err
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	/*http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})*/
	data.Default.Ctx.SetCookie("token", tokenString, expirationTimeSecond, "/", "localhost", false, true)

	return []pipeline.Message{data}, nil
}
