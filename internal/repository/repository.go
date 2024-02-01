package repository

/**
Репозитарий на пять таблиц:
	users
	user_accrual
	user_withdraw
	user_balance
	type_status
*/

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opentracing/opentracing-go"
)

type Repository interface {

	//############# UPDATE
	UpdateAccrualByID(ctx context.Context, dataAccrual *Accrual) error

	//############# SELECT
	SelectWithdrawByUserSum(ctx context.Context, dataUser *User) (float32, error)
	SelectAccrualByUserSum(ctx context.Context, dataUser *User) (float32, error)

	SelectUserByEmail(ctx context.Context, dataUser *User) (*User, error)

	SelectAccrualByUser(ctx context.Context, dataUser *User) ([]AccrualList, error)
	SelectAccrualByIDorder(ctx context.Context, dataAccrual *Accrual) (*Accrual, error)
	SelectAccrualForSendLoyalty(ctx context.Context, dataAccrual *[]Accrual) (*[]Accrual, error)

	SelectWithdrawByUsers(ctx context.Context, dataUser *User) ([]WithdrawList, error)
	SelectBalanceByUser(ctx context.Context, dataUser *User) (Balance, error)

	//############# INSERT
	InsertUser(ctx context.Context, dataUser *User) (*User, error)
	InsertAccrual(ctx context.Context, accrual *Accrual) error
	InsertWithdraw(ctx context.Context, withdraw *Withdraw) error
	InsertBalance(ctx context.Context, balance *Balance) error

	//############# INIT
	InitCatalogData(ctx context.Context) error
	GetCatalogData(ctx context.Context) *CatalogData
}

type repository struct {
	db          *pgxpool.Pool
	catalogData *CatalogData
}

func NewRepository(db *pgxpool.Pool) Repository {

	rep := &repository{
		db:          db,
		catalogData: &CatalogData{},
	}

	return rep
}

func (rep *repository) GetCatalogData(ctx context.Context) *CatalogData {
	return rep.catalogData
}

// Заполняем справочники
func (rep *repository) InitCatalogData(ctx context.Context) error {

	err := rep.SelectTypeStatusAll(ctx)

	if err != nil {
		return err
	}

	return nil
}

// ################################# UPDATE ########################################
func (rep *repository) UpdateAccrualByID(ctx context.Context, dataAccrual *Accrual) error {
	sqlStatement := `UPDATE user_accrual SET accrual = $1, id_status = $2  WHERE id = $3;`
	_, err := rep.db.Exec(ctx, sqlStatement, dataAccrual.Accrual, dataAccrual.IDStatus, dataAccrual.ID)

	if err != nil {
		return fmt.Errorf("error update Accrual: %v", err)
	}

	//updateCount := res.RowsAffected()

	//fmt.Printf("Обновили: id = %d, count = %d", dataAccrual.ID, updateCount)

	return nil
}

//################################ SELECT ########################################

// Получаем запись по email клиента
func (rep *repository) SelectUserByEmail(ctx context.Context, dataUser *User) (*User, error) {
	/*span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectUserByEmail")
	defer span.Finish()

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.HandlerBalance")
	if span != nil {
		defer span.Finish()
	}*/

	// Query for a value based on a single row.
	if err := rep.db.QueryRow(ctx, "SELECT password, id FROM users WHERE login = $1",
		dataUser.Login).Scan(&dataUser.Password, &dataUser.ID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return dataUser, nil

}

// Получаем список всех начислений клиента по UserId
func (rep *repository) SelectWithdrawByUserSum(ctx context.Context, dataUser *User) (float32, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectWithdrawByUserSum")
	defer span.Finish()

	status := "PROCESSED"

	var sumSum float32 = 0

	sqlQuery := `SELECT coalesce(SUM(sum), 0.00) as sum_withdraw FROM user_withdraw inner join type_status on user_withdraw.ID_status = type_status.ID  
						Where id_user = $1 and type_status.name = $2`
	row := rep.db.QueryRow(ctx, sqlQuery, dataUser.ID, status)

	err := row.Scan(&sumSum)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("error not row: %v", err)
		}
		return 0, fmt.Errorf("error select all accrual SelectWithdrawByUserSum: %v", err)
	}

	return sumSum, nil
}

// Получаем список всех начислений клиента по UserId
func (rep *repository) SelectAccrualByUserSum(ctx context.Context, dataUser *User) (float32, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectAccrualByUserSum")
	defer span.Finish()

	status := "PROCESSED"

	var sumAccrual float32 = 0

	sqlQuery := `SELECT coalesce(SUM(accrual), 0.00) as sum_accrual FROM user_accrual inner join type_status on user_accrual.ID_status = type_status.ID
					Where id_user = $1 and type_status.name = $2`
	row := rep.db.QueryRow(ctx, sqlQuery, dataUser.ID, status)

	err := row.Scan(&sumAccrual)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("error not row: %v", err)
		}
		return 0, fmt.Errorf("error select all accrual SelectAccrualByUserSum: %v", err)
	}

	return sumAccrual, nil
}

// Получаем список всех списаний клиента по UserId
func (rep *repository) SelectWithdrawByUser(ctx context.Context, dataUser *User) ([]WithdrawList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectWithdrawByUser")
	defer span.Finish()

	withdrawCollect := []WithdrawList{}
	withdraw := WithdrawList{}

	sqlQuery := `SELECT id_order, sum, date_trunc('second',processed_at::timestamptz) FROM user_withdraw Where id_user = $1`
	rowsAccrual, err := rep.db.Query(ctx, sqlQuery, dataUser.ID)
	if err != nil {
		return nil, fmt.Errorf("error select all accrual SelectWithdrawByUser: %v", err)
	}

	// Закрываем rowsAccrual
	defer func() {
		_ = rowsAccrual.Err()
	}()

	for rowsAccrual.Next() {
		err = rowsAccrual.Scan(&withdraw.IDorder, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawCollect = append(withdrawCollect, withdraw)
	}

	return withdrawCollect, nil
}

// Получаем список всех начислений клиента по UserId
func (rep *repository) SelectAccrualByUser(ctx context.Context, dataUser *User) ([]AccrualList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectAccrualByUser")
	defer span.Finish()

	accrualCollect := []AccrualList{}
	accrual := AccrualList{}
	sqlQuery := `SELECT id_order, accrual, type_status.name as status, date_trunc('second',uploaded_at::timestamptz)
					FROM user_accrual inner join type_status on user_accrual.ID_status = type_status.ID  
						Where id_user = $1`
	rowsAccrual, err := rep.db.Query(ctx, sqlQuery, dataUser.ID)
	if err != nil {
		return nil, fmt.Errorf("error select all accrual SelectAccrualByUser: %v", err)
	}

	// Закрываем rowsAccrual
	defer func() {
		_ = rowsAccrual.Err()
	}()

	for rowsAccrual.Next() {
		err = rowsAccrual.Scan(&accrual.IDorder, &accrual.Accrual, &accrual.Status, &accrual.UploadedAt)
		if err != nil {
			return nil, err
		}
		accrualCollect = append(accrualCollect, accrual)
	}

	return accrualCollect, nil
}

// Получаем запись из таблицы по номеру заказа
func (rep *repository) SelectAccrualForSendLoyalty(ctx context.Context, dataAccrual *[]Accrual) (*[]Accrual, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectAccrualForSendLoyalty")
	defer span.Finish()

	rows, err := rep.db.Query(ctx, "SELECT * FROM user_accrual Where id_status = $1 or id_status = $2 or id_status = $3",
		rep.catalogData.TypeStatus["NEW"], rep.catalogData.TypeStatus["PROCESSING"], rep.catalogData.TypeStatus["REGISTERED"])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		log.Println("Error querying user_accrual", err)
		return nil, err
	}

	//Accrual := *dataAccrual

	accrualRows, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Accrual])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return nil, err
	}

	//Странно, приходится использователь варинат так как ошибка ErrNoRows не возвращается хотя в мануале имеет место быть
	if len(accrualRows) <= 0 {
		return nil, errors.New("no accrual rows not found")
	}

	//Забасываем данные по линку перменной dataAccrual
	for _, p := range accrualRows {
		fmt.Println(p)
		//Набиваме массив данными перед этим сам указатель разименовываем
		*dataAccrual = append(*dataAccrual, *p)
	}

	return dataAccrual, nil
}

// Получаем запись из таблицы по номеру заказа
func (rep *repository) SelectAccrualByIDorder(ctx context.Context, dataAccrual *Accrual) (*Accrual, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectAccrualByIDorder")
	defer span.Finish()

	rows, err := rep.db.Query(ctx, "SELECT * FROM user_accrual Where id_order = $1", dataAccrual.IDorder)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Println("Error querying SelectAccrualByIDorder", err)
		return nil, err
	}

	//Accrual := *dataAccrual

	accrualRows, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Accrual])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return nil, err
	}

	if len(accrualRows) <= 0 {
		return nil, nil
	}

	fmt.Println(accrualRows)

	//Забасываем данные по линку перменной dataAccrual
	for _, p := range accrualRows {
		fmt.Println(p)
		*dataAccrual = *p
	}

	return dataAccrual, nil
}

// Получаем список всех списаний клиента по UserId
func (rep *repository) SelectWithdrawByUsers(ctx context.Context, dataUser *User) ([]WithdrawList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectWithdrawByUsers")
	defer span.Finish()

	withdrawCollect := []WithdrawList{}
	withdraw := WithdrawList{}

	sqlQuery := `SELECT id_order, sum, date_trunc('second',processed_at::timestamptz) FROM user_withdraw Where id_user = $1`
	rowsWithdraw, err := rep.db.Query(ctx, sqlQuery, dataUser.ID)
	if err != nil {
		return nil, fmt.Errorf("error select all accrual SelectWithdrawByUsers: %v", err)
	}

	// Закрываем rowsAccrual
	defer func() {
		//_ = rowsWithdraw.Close()
		_ = rowsWithdraw.Err()
	}()

	for rowsWithdraw.Next() {
		err = rowsWithdraw.Scan(&withdraw.IDorder, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawCollect = append(withdrawCollect, withdraw)
	}

	return withdrawCollect, nil
}

// Получаем баланс клиента по UserId
func (rep *repository) SelectBalanceByUser(ctx context.Context, dataUser *User) (Balance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectBalanceByUser")
	defer span.Finish()

	balanceUser := Balance{}
	if err := rep.db.QueryRow(ctx, "SELECT * FROM user_balance WHERE id_user = $1",
		dataUser.ID).Scan(&balanceUser); err != nil {
		return Balance{}, err
	}

	return balanceUser, nil
}

// Получаем айди статуса по его названию
func (rep *repository) SelectTypeStatusByName(ctx context.Context, nametype string) (int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.SelectTypeStatusByName")
	defer span.Finish()

	idTypeStatus := 0

	if err := rep.db.QueryRow(ctx, "SELECT * FROM type_status WHERE name = $1",
		nametype).Scan(&idTypeStatus); err != nil {
		return 0, err
	}

	if idTypeStatus == 0 {
		return 0, errors.New("not found type status for accrual")
	}

	return idTypeStatus, nil
}

// Получаем айди статуса по его названию и сохраняем в справочнике для того чтобы не дергать базу лишний раз
func (rep *repository) SelectTypeStatusAll(ctx context.Context) error {

	fmt.Println("SelectTypeStatusAll")

	type TypeStatus struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	rep.catalogData.TypeStatus = make(map[string]int)

	typeStatus := TypeStatus{}

	rowsTypeStatus, err := rep.db.Query(ctx, "SELECT * FROM type_status")
	if err != nil {
		return err
	}

	// Закрываем rowsAccrual
	defer func() {
		///_ = rowsTypeStatus.Close()
		_ = rowsTypeStatus.Err()
	}()

	for rowsTypeStatus.Next() {
		err = rowsTypeStatus.Scan(&typeStatus.ID, &typeStatus.Name)

		if err != nil {
			return err
		}

		rep.catalogData.TypeStatus[typeStatus.Name] = typeStatus.ID
	}

	return nil
}

//################################ INSERT ########################################

// Сохраняем нового пользователя
func (rep *repository) InsertUser(ctx context.Context, dataUser *User) (*User, error) {
	fmt.Println("DB InsertUser")

	/*span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.InsertUser")
	defer span.Finish()*/

	var idInsertRow int

	fmt.Println(dataUser)

	err := rep.db.QueryRow(ctx,
		"INSERT INTO users(login, password) VALUES($1, $2) RETURNING id", dataUser.Login, dataUser.Password).Scan(&idInsertRow)

	if err != nil {
		return nil, err
	}

	dataUser.ID = idInsertRow

	return dataUser, nil
}

func (rep *repository) InsertAccrual(ctx context.Context, accrual *Accrual) error {
	fmt.Println("DB insertAccrual")
	fmt.Println("InsertAccrual:", rep.catalogData.TypeStatus["NEW"])

	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.InsertAccrual")
	defer span.Finish()

	_, err := rep.db.Exec(ctx,
		"INSERT INTO user_accrual(id_user, id_order, accrual, id_status) VALUES (@IDUser, @IDorder, @accrual, @IDStatus)",
		pgx.NamedArgs{
			"IDUser":   accrual.IDUser,
			"IDorder":  accrual.IDorder,
			"accrual":  accrual.Accrual,
			"IDStatus": rep.catalogData.TypeStatus["NEW"],
		})

	fmt.Println(rep.catalogData.TypeStatus)

	if err != nil {
		return err
	}

	return nil
}

func (rep *repository) InsertWithdraw(ctx context.Context, withdraw *Withdraw) error {
	fmt.Println("DB insertWithdraw")

	span, ctx := opentracing.StartSpanFromContext(ctx, "Repo.InsertWithdraw")
	defer span.Finish()

	_, err := rep.db.Exec(ctx,
		"INSERT INTO user_withdraw(id_user, id_order, sum, id_status) VALUES (@IDUser, @IDorder, @sum, @IDStatus)",
		pgx.NamedArgs{
			"IDUser":   withdraw.IDUser,
			"IDorder":  withdraw.IDorder,
			"sum":      withdraw.Sum,
			"IDStatus": rep.catalogData.TypeStatus["PROCESSED"],
		})

	if err != nil {
		return err
	}

	return nil
}

func (rep *repository) InsertBalance(ctx context.Context, balance *Balance) error {
	fmt.Println("DB insertBalance")

	_, err := rep.db.Exec(ctx,
		"INSERT INTO user_balance(id_user, accrual, with_draw) VALUES (@IDUser, @accrual, @withdraw)",
		pgx.NamedArgs{
			"IDUser":   balance.IDUser,
			"accrual":  balance.Accrual,
			"withdraw": balance.Withdraw,
		})

	if err != nil {
		return err
	}

	return nil
}
