package repository

import "time"

type Datas struct {
	User User
}

type User struct {
	ID       int    `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

type Accrual struct {
	ID         int       `db:"id"`
	IDUser     int       `db:"id_user"`
	IDorder    string    `db:"id_order" json:"number"`
	Accrual    float32   `db:"accrual,omitempty" json:"accrual,omitempty"`
	IDStatus   int       `db:"id_status" json:"id_status"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}

type AccrualList struct {
	IDorder    string    `db:"id_order" json:"number"`
	Accrual    float32   `db:"accrual,omitempty" json:"accrual,omitempty"`
	Status     string    `db:"status" json:"status"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}

type Withdraw struct {
	ID          int       `db:"id"`
	IDUser      int       `db:"id_user"`
	IDorder     string    `db:"id_order" json:"order"`
	Sum         float32   `db:"sum" json:"sum"`
	IDStatus    int       `db:"id_status" json:"id_status"`
	ProcessedAt time.Time `db:"processed_at" json:"processed_at"`
}

type WithdrawList struct {
	IDorder     string    `db:"id_order" json:"order"`
	Sum         float32   `db:"sum,omitempty" json:"sum,omitempty"`
	ProcessedAt time.Time `db:"processed_at" json:"processed_at"`
}

type Balance struct {
	ID       int
	IDUser   int
	Accrual  float32
	Withdraw float32
}

type ResponseBalance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type CatalogData struct {
	TypeStatus map[string]int
}

/*
CREATE TABLE IF NOT EXISTS user_balance(
		id INT GENERATED ALWAYS AS IDENTITY,
		id_user INT NOT NULL,
		accrual double precision,
		withdraw double precision,
		PRIMARY KEY(id),
		CONSTRAINT fk_user_balance
			FOREIGN KEY(id_user)
				REFERENCES users(id)
	);*/

/*CREATE TABLE IF NOT EXISTS user_accrual(
	id INT GENERATED ALWAYS AS IDENTITY,
	id_user INT NOT NULL,
	id_order INT NOT NULL,
	accrual double precision,
	id_status INT NOT NULL,
	uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(id),
	CONSTRAINT fk_user_accrual
		FOREIGN KEY(id_user)
			REFERENCES users(id),
	CONSTRAINT fk_status_type
		FOREIGN KEY(id_status)
			REFERENCES type_status(id)
);
"INSERT INTO metrics_counter(id_user, id_order, accrual, id_status, uploaded_at) VALUES (@email, @password)",
*/

/*CREATE TABLE IF NOT EXISTS user_withdraw(
	id INT GENERATED ALWAYS AS IDENTITY,
	id_user INT NOT NULL,
	id_order VARCHAR(255) NOT NULL,
	sum double precision,
	id_status INT NOT NULL,
	processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(id),
	CONSTRAINT fk_user_with_draw
		FOREIGN KEY(id_user)
			REFERENCES users(id),
	CONSTRAINT fk_status_type
		FOREIGN KEY(id_status)
			REFERENCES type_status(id)
);*/
