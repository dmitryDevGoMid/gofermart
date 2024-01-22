package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	mocks "github.com/dmitryDevGoMid/gofermart/internal/repository/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	User repository.User
	jwt.RegisteredClaims
}

func GetTokenForTest() (string, error) {

	expirationTimeSecond := 60 * 1

	expirationTime := time.Now().Add(time.Duration(expirationTimeSecond) * time.Second)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		User: repository.User{ID: 1, Login: "opsegorsmall@email.ro", Password: "123wafde"},
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
		return "", err
	}

	return tokenString, nil
}

func TestHandlerRegiser(t *testing.T) {
	type mockBehavior func(ctx context.Context, mocks *mocks.MockRepository)
	user := repository.User{ID: 1, Login: "opsegorsmall@email.ro", Password: "123wafde"}

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		statusCode   int
		counterValue string
	}{
		{
			name: "get balance 200",
			mockBehavior: func(ctx context.Context, mocks *mocks.MockRepository) {

				var sumWithDraw float32 = 0
				var sumAccrual float32 = 0

				sumAccrual = 55.5
				sumWithDraw = 10.5

				mocks.EXPECT().SelectWithdrawByUserSum(context.Background(), &user).Return(sumWithDraw, nil).AnyTimes()
				mocks.EXPECT().SelectAccrualByUserSum(context.Background(), &user).Return(sumAccrual, nil).AnyTimes()
			},
			statusCode:   200,
			counterValue: `{"current":45,"withdrawn":10.5}`,
		},
		/*{
			name: "get ping database connect to posgress 500",
			mockBehavior: func(ctx context.Context, mocks *mocks.MockRepository) {
				//mocks.EXPECT().PingDatabase(ctx).Return(errors.New("Bad")).AnyTimes()
				mocks.EXPECT().GetCatalogData(ctx).Return(errors.New("Bad")).AnyTimes()
			},
			statusCode:   500,
			counterValue: "Failed to ping database",
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			s := mocks.NewMockRepository(c)

			services := s

			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			handler := goferHandler{cfg: cfg, repository: services}

			//Init Point Handlers
			r := gin.Default()
			//Create request
			//w := httptest.NewRecorder()

			//_, r := gin.CreateTestContext(w)

			/*r.Use(func(c *gin.Context) {
				c.SetCookie("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VyIjp7IklEIjoxLCJMb2dpbiI6Im9wc2Vnb3JzbWFsbEBlbWFpbC5ybyIsIlBhc3N3b3JkIjoiJDJhJDEwJEZYNGhpYlA5RDBUdzBFNEVxRVJRL2VBN21NdDVIWEZnNTlGMmxIUmZTN1pFT0l6dnBwZXZ5In0sImV4cCI6MTcwNTQwNDUxOX0.PoGfAuob7JtT4b7vga5U7_mRRR6iENlPw0nVD_a9ENE", 3600, "/", "localhost", false, true)
			})*/

			r.GET("/balance/", func(c *gin.Context) {
				tt.mockBehavior(c, s)
			}, handler.Balance)

			w := httptest.NewRecorder()

			token, err := GetTokenForTest()

			if err != nil {
				fmt.Println(err)
			}

			cookie := http.Cookie{
				Name:     "token",
				Value:    token,
				Domain:   "localhost",
				Path:     "/",
				MaxAge:   60 * 60,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			req := httptest.NewRequest("GET", "/balance/", nil)
			req.AddCookie(&cookie)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}
