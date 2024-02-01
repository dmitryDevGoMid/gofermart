package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/jaeger"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/security"
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

	//Период жизни токена
	expirationTimeSecond := 60 * 1

	expirationTime := time.Now().Add(time.Duration(expirationTimeSecond) * time.Second)
	//Формируем заявку с полезными данными
	claims := &Claims{
		User: repository.User{ID: 1, Login: "opsegorsmall@email.ro", Password: "123wafde"},
		RegisteredClaims: jwt.RegisteredClaims{
			// Устанавливаем срок действия токена
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Устанавливаем слгоритм и условия
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Токен
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func TestHandlerBalance(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)
	//Тестируемый пользователь
	user := repository.User{ID: 1, Login: "opsegorsmall@email.ro", Password: "123wafde"}

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: true,
			name:          "get 200",
			mockBehavior: func(mocks *mocks.MockRepository) {

				var sumWithDraw float32 = 0
				var sumAccrual float32 = 0

				//Данняе которые вернет заглушка
				sumAccrual = 55.5
				sumWithDraw = 10.5

				//Определяем ожидаемое поведения заглушки для репозитария
				mocks.EXPECT().SelectWithdrawByUserSum(context.Background(), &user).Return(sumWithDraw, nil).AnyTimes()
				mocks.EXPECT().SelectAccrualByUserSum(context.Background(), &user).Return(sumAccrual, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /blance/
			counterValue: `{"current":45,"withdrawn":10.5}`,
		},
		{
			name:          "get 401",
			authorisation: false,
			mockBehavior:  func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 401,
			//Ожидаемые данные от poin: /blance/
			counterValue: ``,
		},
		{
			name:          "get 500",
			authorisation: true,
			mockBehavior: func(mocks *mocks.MockRepository) {
				var sumWithDraw float32 = 0
				var sumAccrual float32 = 0

				mocks.EXPECT().SelectWithdrawByUserSum(context.Background(), &user).Return(sumWithDraw, errors.New("Bad")).AnyTimes()
				mocks.EXPECT().SelectAccrualByUserSum(context.Background(), &user).Return(sumAccrual, errors.New("Bad")).AnyTimes()
			},
			statusCode:   500,
			counterValue: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			//Инициализируем действия заглушки
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.GET("/balance/", func(c *gin.Context) {
			}, handler.Balance)

			w := httptest.NewRecorder()

			//Получаем токен для авторизации
			token, err := GetTokenForTest()

			if err != nil {
				fmt.Println(err)
			}

			// Куки для проверки аутентификации
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

			//Требуется авторизация или нет
			if tt.authorisation {
				req.AddCookie(&cookie)
			}

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}

/*
200 — пользователь успешно зарегистрирован и аутентифицирован;
400 — неверный формат запроса;
409 — логин уже занят;
500 — внутренняя ошибка сервера.
*/
func TestHandlerRegister(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	user := repository.User{
		ID:       1,
		Login:    "vasiliy",
		Password: "123wafde",
	}

	userData, err := json.Marshal(user)
	fmt.Println(userData)
	if err != nil {
		fmt.Println("Couldn't marshal metrics by test:", err)
	}

	tests := []struct {
		name          string
		body          string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 200",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(nil, nil).AnyTimes()
				mocks.EXPECT().InsertUser(context.Background(), &user).Return(&user, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /register/
			counterValue: `{"code":200,"message":"Success"}`,
		},
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 409",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(&user, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 409,
			//Ожидаемые данные от poin: /register/
			counterValue: ``,
		},
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 500",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(nil, errors.New("Internal server 500")).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 500,
			//Ожидаемые данные от poin: /register/
			counterValue: ``,
		},
		{
			authorisation: false,
			//Кривой состав json data
			body:         `{"login": "login", "fgfgfgfg": "password}`,
			name:         "get 400",
			mockBehavior: func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 400,
			//Ожидаемые данные от poin: /register/
			counterValue: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование
			cfg.Server.TestingEnabled = true

			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.POST("/register/", func(c *gin.Context) {
			}, handler.Register)

			w := httptest.NewRecorder()

			if err != nil {
				fmt.Println(err)
			}

			req := httptest.NewRequest("POST", "/register/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}

func TestHandlerLogin(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	user := repository.User{
		ID:       1,
		Login:    "vasiliy",
		Password: "123wafde",
	}

	userData, err := json.Marshal(user)
	fmt.Println(userData)
	if err != nil {
		fmt.Println("Couldn't marshal metrics by test:", err)
	}

	tests := []struct {
		name          string
		body          string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 200",
			mockBehavior: func(mocks *mocks.MockRepository) {

				encryptPasswword, err := security.EncryptPassword(user.Password)
				if err != nil {
					fmt.Println(err)
				}

				userWithEncryptPassword := user
				userWithEncryptPassword.Password = encryptPasswword

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(&userWithEncryptPassword, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /login/
			counterValue: `{"code":200,"message":"Success"}`,
		},
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 401",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(nil, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 401,
			//Ожидаемые данные от poin: /login/
			counterValue: ``,
		},
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 500",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(nil, errors.New("Internal server 500")).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 500,
			//Ожидаемые данные от poin: /login/
			counterValue: ``,
		},
		{
			authorisation: false,
			//Кривой состав json data
			body:         `{"login": "login", "fgfgfgfg": "password}`,
			name:         "get 400",
			mockBehavior: func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 400,
			//Ожидаемые данные от poin: /register/
			counterValue: `{"code":400,"message":"Bad Request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование

			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.POST("/login/", func(c *gin.Context) {
			}, handler.Login)

			w := httptest.NewRecorder()

			if err != nil {
				fmt.Println(err)
			}

			req := httptest.NewRequest("POST", "/login/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}

func TestHandlerLoginCookies(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	user := repository.User{
		ID:       1,
		Login:    "vasiliy",
		Password: "123wafde",
	}

	userData, err := json.Marshal(user)
	fmt.Println(userData)
	if err != nil {
		fmt.Println("Couldn't marshal metrics by test:", err)
	}

	tests := []struct {
		name          string
		body          string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: false,
			body:          string(userData),
			name:          "get 200",
			mockBehavior: func(mocks *mocks.MockRepository) {

				encryptPasswword, err := security.EncryptPassword(user.Password)
				if err != nil {
					fmt.Println(err)
				}

				userWithEncryptPassword := user
				userWithEncryptPassword.Password = encryptPasswword

				mocks.EXPECT().SelectUserByEmail(context.Background(), &user).Return(&userWithEncryptPassword, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /login/
			counterValue: `{"code":200,"message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование

			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.POST("/login/", func(c *gin.Context) {
			}, handler.Login)

			w := httptest.NewRecorder()

			if err != nil {
				fmt.Println(err)
			}

			req := httptest.NewRequest("POST", "/login/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
			assert.NotEmpty(t, w.Result().Cookies())
		})
	}
}

/*200 — номер заказа уже был загружен этим пользователем;
202 — новый номер заказа принят в обработку;
400 — неверный формат запроса;
401 — пользователь не аутентифицирован;
409 — номер заказа уже был загружен другим пользователем;
422 — неверный формат номера заказа;
500 — внутренняя ошибка сервера.
*/

func TestHandlerOrders(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	lunaNumber := luna.Generate(12)

	contentType := "text/plain"

	accrual := repository.Accrual{
		ID:         1,
		IDUser:     2,
		IDorder:    lunaNumber,
		Accrual:    54.5,
		IDStatus:   1,
		UploadedAt: time.Now(),
	}

	tests := []struct {
		name          string
		body          string
		contentType   string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: true,
			body:          lunaNumber,
			name:          "get 202",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByIDorder(context.Background(), &repository.Accrual{IDorder: lunaNumber}).Return(nil, nil).AnyTimes()
				mocks.EXPECT().InsertAccrual(context.Background(), &repository.Accrual{IDUser: 1, IDorder: lunaNumber}).Return(nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 202,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: true,
			body:          lunaNumber,
			name:          "get 409",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByIDorder(context.Background(), &repository.Accrual{IDorder: lunaNumber}).Return(&accrual, nil).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 409,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: true,
			body:          lunaNumber,
			name:          "get 200",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByIDorder(context.Background(), &repository.Accrual{IDorder: lunaNumber}).Return(&repository.Accrual{IDUser: 1, IDorder: lunaNumber}, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: false,
			body:          lunaNumber,
			name:          "get 401",
			contentType:   contentType,
			mockBehavior:  func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 401,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: true,
			body:          lunaNumber,
			name:          "get 500",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByIDorder(context.Background(), &repository.Accrual{IDorder: lunaNumber}).Return(nil, errors.New("Bad 500")).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 500,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: true,
			//Кривой состав json data
			body:         lunaNumber,
			name:         "get 400",
			contentType:  "application/json",
			mockBehavior: func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 400,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: true,
			//Кривой состав json data
			body:         "9386548568383456",
			name:         "get 422",
			contentType:  contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 422,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование
			cfg.Server.TestingEnabled = false
			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.POST("/orders/", func(c *gin.Context) {
			}, handler.OrdersPost)

			w := httptest.NewRecorder()

			// Получаем токен для авторизации
			token, err := GetTokenForTest()

			if err != nil {
				fmt.Println(err)
			}

			//  Куки для проверки аутентификации
			cookie := http.Cookie{
				Name:     "token",
				Value:    token,
				Domain:   "localhost",
				Path:     "/",
				MaxAge:   60 * 60,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			req := httptest.NewRequest("POST", "/orders/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			if tt.authorisation {
				req.AddCookie(&cookie)
			}

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}

/*
200 — успешная обработка запроса
204 — нет данных для ответа
401 — пользователь не аутентифицирован;
500 — внутренняя ошибка сервера.
*/
func TestHandlerOrdersList(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	lunaNumber1 := luna.Generate(12)
	lunaNumber2 := luna.Generate(12)

	contentType := "text/plain"

	user := repository.User{
		ID:       1,
		Login:    "opsegorsmall@email.ro",
		Password: "123wafde",
	}

	listAccrual := []repository.AccrualList{
		{IDorder: lunaNumber1, Accrual: 10.4, Status: "NEW", UploadedAt: time.Now()},
		{IDorder: lunaNumber2, Accrual: 5.4, Status: "NEW", UploadedAt: time.Now()},
	}

	listAccrualMarshal, err := json.Marshal(listAccrual)

	if err != nil {
		fmt.Println(err)
	}

	_ = repository.Accrual{
		ID:         1,
		IDUser:     2,
		IDorder:    lunaNumber1,
		Accrual:    54.5,
		IDStatus:   1,
		UploadedAt: time.Now(),
	}

	tests := []struct {
		name          string
		contentType   string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: true,
			name:          "get 200",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByUser(context.Background(), &user).Return(listAccrual, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /orders/
			counterValue: string(listAccrualMarshal),
		},
		{
			authorisation: true,
			name:          "get 204",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByUser(context.Background(), &user).Return([]repository.AccrualList{}, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 204,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: false,
			name:          "get 401",
			contentType:   contentType,
			mockBehavior:  func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 401,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
		{
			authorisation: true,
			name:          "get 500",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectAccrualByUser(context.Background(), &user).Return(nil, errors.New("Bad 500")).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 500,
			//Ожидаемые данные от poin: /orders/
			counterValue: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование
			cfg.Server.TestingEnabled = false
			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.GET("/orders/", func(c *gin.Context) {
			}, handler.OrdersGet)

			w := httptest.NewRecorder()

			// Получаем токен для авторизации
			token, err := GetTokenForTest()

			if err != nil {
				fmt.Println(err)
			}

			//  Куки для проверки аутентификации
			cookie := http.Cookie{
				Name:     "token",
				Value:    token,
				Domain:   "localhost",
				Path:     "/",
				MaxAge:   60 * 60,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			req := httptest.NewRequest("GET", "/orders/", nil)
			//req.Header.Set("Content-Type", tt.contentType)

			if tt.authorisation {
				req.AddCookie(&cookie)
			}

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}

/*200 — успешная обработка запросаж
401 — пользователь не аутентифицирован;
402 — на счету не достаточно средств
422 — неверный  номера заказа;
500 — внутренняя ошибка сервера.
*/

func TestHandlerWithdraw(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	lunaNumber := luna.Generate(12)

	user := repository.User{
		ID:       1,
		Login:    "opsegorsmall@email.ro",
		Password: "123wafde",
	}

	//Тело запроса с корректным номером заказа
	body := struct {
		Order string
		Sum   float32
	}{
		Order: lunaNumber,
		Sum:   13.4,
	}

	//Тело запроса с кривым номеро заказа
	requestBody, err := json.Marshal(body)

	if err != nil {
		fmt.Println(err)
	}

	bodBady := struct {
		Order string
		Sum   float32
	}{
		Order: "23423423423",
		Sum:   13.4,
	}

	requestBodyBady, err := json.Marshal(bodBady)

	if err != nil {
		fmt.Println(err)
	}

	contentType := "application/json"

	withdraw := repository.Withdraw{
		IDUser:  1,
		IDorder: lunaNumber,
		Sum:     13.4,
	}

	tests := []struct {
		name          string
		body          string
		contentType   string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: true,
			body:          string(requestBody),
			name:          "get 200",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {
				var sumWithDraw float32 = 0
				var sumAccrual float32 = 0

				//Данняе которые вернет заглушка
				sumAccrual = 55.5
				sumWithDraw = 10.5

				//Определяем ожидаемое поведения заглушки для репозитария
				mocks.EXPECT().SelectWithdrawByUserSum(context.Background(), &user).Return(sumWithDraw, nil).AnyTimes()
				mocks.EXPECT().SelectAccrualByUserSum(context.Background(), &user).Return(sumAccrual, nil).AnyTimes()

				mocks.EXPECT().InsertWithdraw(context.Background(), &withdraw).Return(nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /balance/withdraw/
			counterValue: ``,
		},
		{
			authorisation: true,
			body:          string(requestBody),
			name:          "get 402",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {
				var sumWithDraw float32 = 0
				var sumAccrual float32 = 0

				//Данняе которые вернет заглушка
				sumAccrual = 10.5
				sumWithDraw = 5.3

				//Определяем ожидаемое поведения заглушки для репозитария
				mocks.EXPECT().SelectWithdrawByUserSum(context.Background(), &user).Return(sumWithDraw, nil).AnyTimes()
				mocks.EXPECT().SelectAccrualByUserSum(context.Background(), &user).Return(sumAccrual, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 402,
			//Ожидаемые данные от poin: /balance/withdraw/
			counterValue: ``,
		},
		{
			authorisation: true,
			//Кривой состав json data
			body:         string(requestBodyBady),
			name:         "get 422",
			contentType:  contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 422,
			//Ожидаемые данные от poin: /balance/withdraw/
			counterValue: ``,
		},
		{
			authorisation: false,
			body:          string(requestBody),
			name:          "get 401",
			contentType:   contentType,
			mockBehavior:  func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 401,
			//Ожидаемые данные от poin: /balance/withdraw/
			counterValue: ``,
		},
		{
			authorisation: true,
			body:          string(requestBody),
			name:          "get 500",
			contentType:   contentType,
			mockBehavior: func(mocks *mocks.MockRepository) {
				var sumWithDraw float32 = 0
				var sumAccrual float32 = 0

				//Данняе которые вернет заглушка
				sumAccrual = 55.5

				//Определяем ожидаемое поведения заглушки для репозитария
				mocks.EXPECT().SelectWithdrawByUserSum(context.Background(), &user).Return(sumWithDraw, errors.New("Bad Request 500")).AnyTimes()
				mocks.EXPECT().SelectAccrualByUserSum(context.Background(), &user).Return(sumAccrual, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 500,
			//Ожидаемые данные от poin: /balance/withdraw/
			counterValue: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование
			cfg.Server.TestingEnabled = false
			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.POST("/balance/withdraw/", func(c *gin.Context) {
			}, handler.BalanceWithDraw)

			w := httptest.NewRecorder()

			// Получаем токен для авторизации
			token, err := GetTokenForTest()

			if err != nil {
				fmt.Println(err)
			}

			//  Куки для проверки аутентификации
			cookie := http.Cookie{
				Name:     "token",
				Value:    token,
				Domain:   "localhost",
				Path:     "/",
				MaxAge:   60 * 60,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			req := httptest.NewRequest("POST", "/balance/withdraw/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			if tt.authorisation {
				req.AddCookie(&cookie)
			}

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}

/*
200 — успешная обработка запроса
204 — нет ни одного списания
401 — пользователь не аутентифицирован;
500 — внутренняя ошибка сервера.
*/

func TestHandlerWithdrawals(t *testing.T) {
	//Обьявляем тип функция для вызова во время выполнения запроса
	type mockBehavior func(mocks *mocks.MockRepository)

	lunaNumber1 := luna.Generate(12)
	lunaNumber2 := luna.Generate(12)

	user := repository.User{
		ID:       1,
		Login:    "opsegorsmall@email.ro",
		Password: "123wafde",
	}

	listWithdrawals := []repository.WithdrawList{
		{IDorder: lunaNumber1, Sum: 10.4, ProcessedAt: time.Now()},
		{IDorder: lunaNumber2, Sum: 5.4, ProcessedAt: time.Now()},
	}

	listWithdrawalsMarshal, err := json.Marshal(listWithdrawals)

	if err != nil {
		fmt.Println(err)
	}

	tests := []struct {
		name          string
		contentType   string
		mockBehavior  mockBehavior
		statusCode    int
		counterValue  string
		authorisation bool
	}{
		{
			authorisation: true,
			name:          "get 200",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectWithdrawByUsers(context.Background(), &user).Return(listWithdrawals, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 200,
			//Ожидаемые данные от poin: /withdrawals/
			counterValue: string(listWithdrawalsMarshal),
		},
		{
			authorisation: true,
			name:          "get 204",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectWithdrawByUsers(context.Background(), &user).Return([]repository.WithdrawList{}, nil).AnyTimes()

			},
			//Ожидаемый статус ответа
			statusCode: 204,
			//Ожидаемые данные от poin: /withdrawals/
			counterValue: ``,
		},
		{
			authorisation: false,
			name:          "get 401",
			mockBehavior:  func(mocks *mocks.MockRepository) {},
			//Ожидаемый статус ответа
			statusCode: 401,
			//Ожидаемые данные от poin: /withdrawals/
			counterValue: ``,
		},
		{
			authorisation: true,
			name:          "get 500",
			mockBehavior: func(mocks *mocks.MockRepository) {

				mocks.EXPECT().SelectWithdrawByUsers(context.Background(), &user).Return(nil, errors.New("Bad 500")).AnyTimes()
			},
			//Ожидаемый статус ответа
			statusCode: 500,
			//Ожидаемые данные от poin: /withdrawals/
			counterValue: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Иницаиалзация тестирования
			c := gomock.NewController(t)
			defer c.Finish()

			//Мокинг репозитарий
			repository := mocks.NewMockRepository(c)
			tt.mockBehavior(repository)

			//Конфиг данные
			cfg, err := config.ParseConfig()

			if err != nil {
				fmt.Println("Config", err)
			}

			//На вермя тестирования отключаем трассировку
			cfg.Server.TracingEnabled = false
			//На время включаем тестирование
			cfg.Server.TestingEnabled = false
			//Получаем объект трассировки
			tracing := jaeger.NewTracing(cfg)

			//Инициализируем обработчик данныим: конфиг, репозитарий
			handler := goferHandler{cfg: cfg, repository: repository, tracing: tracing}

			//Имитируем response/request handler
			r := gin.Default()

			r.GET("/withdrawals/", func(c *gin.Context) {
			}, handler.WithDrawals)

			w := httptest.NewRecorder()

			// Получаем токен для авторизации
			token, err := GetTokenForTest()

			if err != nil {
				fmt.Println(err)
			}

			//  Куки для проверки аутентификации
			cookie := http.Cookie{
				Name:     "token",
				Value:    token,
				Domain:   "localhost",
				Path:     "/",
				MaxAge:   60 * 60,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			req := httptest.NewRequest("GET", "/withdrawals/", nil)

			if tt.authorisation {
				req.AddCookie(&cookie)
			}

			//  Выполняем на gin метод, которые соответствует интерфейсу type Handler interface { ServeHTTP(ResponseWriter, *Request)}
			r.ServeHTTP(w, req)

			//Выполняем сравнение полученного(w - response) c ожидаемым (tt.)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.counterValue, w.Body.String())
		})
	}
}
