package authentication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/golang-jwt/jwt/v4"
)

type CheckJWTToken struct {
}

func (chain CheckJWTToken) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Cookes Cookes")
	data := result.(*service.Data)

	cookies, err := data.Default.Ctx.Cookie("token")
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusUnauthorized)
		}
		return []pipeline.Message{data}, err
	}

	fmt.Println("Cookies:", cookies)

	// Get the JWT string from the cookie
	tknStr := cookies

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusUnauthorized)
		}
	}
	if tkn != nil {
		if !tkn.Valid {
			data.Default.Response = func() {
				data.Default.Ctx.Status(http.StatusUnauthorized)
			}
			return []pipeline.Message{data}, err
		}
	} else {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusUnauthorized)
		}
		return []pipeline.Message{data}, err
	}

	data.User.User = claims.User

	// Finally, return the welcome message to the user, along with their
	// username given in the token
	//w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
	return []pipeline.Message{data}, nil
}
