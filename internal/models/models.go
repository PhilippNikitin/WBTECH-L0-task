package models

import (
	"database/sql"
	"log"
	"time"

	"sample-app/internal/logging"
)

// Определяем структуры, которые будут отражать информацию о заказе
// Определяем тип Delivery
type Delivery struct {
	OrderUID string `json:"order_uid"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Zip      string `json:"zip"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Region   string `json:"region"`
	Email    string `json:"email"`
}

// Определяем тип Payment
type Payment struct {
	OrderUID     string `json:"order_uid"`
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

// Определяем тип Items
type Items struct {
	OrderUID    string  `json:"order_uid"`
	ChrtID      int     `json:"chrt_id"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	Rid         string  `json:"rid"`
	Name        string  `json:"name"`
	Sale        int     `json:"sale"`
	Size        string  `json:"size"`
	Total_price float64 `json:"total_price"`
	Nm_id       int     `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}

// Определяем тип Order
type Order struct {
	OrderUID           string    `json:"order_uid"`
	Track_number       string    `json:"track_number"`
	Entry              string    `json:"entry"`
	Locale             string    `json:"locale"`
	Internal_signature string    `json:"internal_signature"`
	Customer_id        string    `json:"customer_id"`
	Delivery_service   string    `json:"delivery_service"`
	Shardkey           string    `json:"shardkey"`
	Sm_id              int       `json:"sm_id"`
	Date_created       time.Time `json:"date_created"`
	Oof_shard          string    `json:"oof_shard"`
	Delivery           Delivery  `json:"delivery"`
	Payment            Payment   `json:"payment"`
	Items              Items     `json:"items"`
}

// Функция, которая будет сохранять заказ в базе данных
func SaveOrderInDatabase(db *sql.DB, order Order) error {
	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		logging.Logger.Fatalf("Ошибка при старте транзакции: %v", err)
	}

	_, err = tx.Exec("INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.OrderUID, order.Track_number, order.Entry, order.Locale, order.Internal_signature, order.Customer_id, order.Delivery_service, order.Shardkey, order.Sm_id, order.Date_created, order.Oof_shard)
	if err != nil {
		tx.Rollback()
		logging.Logger.Fatalf("Ошибка при сохранении заказа в PostgreSQL: %v", err)
	}

	_, err = tx.Exec("INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		tx.Rollback()
		logging.Logger.Fatalf("Ошибка при сохранении заказа в PostgreSQL: %v", err)
	}

	_, err = tx.Exec("INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		tx.Rollback()
		logging.Logger.Fatalf("Ошибка при сохранении заказа в PostgreSQL: %v", err)
	}

	_, err = tx.Exec("INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
		order.OrderUID, order.Items.ChrtID, order.Items.TrackNumber, order.Items.Price, order.Items.Rid, order.Items.Name, order.Items.Sale, order.Items.Size, order.Items.Total_price, order.Items.Nm_id, order.Items.Brand, order.Items.Status)
	if err != nil {
		tx.Rollback()
		logging.Logger.Fatalf("Ошибка при сохранении заказа в PostgreSQL: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		logging.Logger.Fatalf("Ошибка при попытке проведения транзакции: %v", err)
	}

	log.Printf("Заказ %s сохранен в базе данных.\n", order.OrderUID)
	logging.Logger.Printf("Заказ %s сохранен в базе данных.\n", order.OrderUID)
	return nil
}

// функция для получения данных о заказе напрямую из базы данных
/*
func GetOrderFromDB(db *sql.DB, orderUID string) (*models.Order, error) {
	order := &models.Order{OrderUID: orderUID}

	// Получение данных из таблицы orders
	err := db.QueryRow("SELECT track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1", orderUID).
		Scan(&order.Track_number, &order.Entry, &order.Locale, &order.Internal_signature, &order.Customer_id, &order.Delivery_service, &order.Shardkey, &order.Sm_id, &order.Date_created, &order.Oof_shard)
	if err != nil {
		return nil, err
	}

	// Получение данных из таблицы delivery
	err = db.QueryRow("SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1", orderUID).
		Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
	if err != nil {
		return nil, err
	}

	// Получение данных из таблицы payment
	err = db.QueryRow("SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1", orderUID).
		Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return nil, err
	}

	// Получение данных из таблицы items
	err = db.QueryRow("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", orderUID).
		Scan(&order.Items.ChrtID, &order.Items.TrackNumber, &order.Items.Price, &order.Items.Rid, &order.Items.Name, &order.Items.Sale, &order.Items.Size, &order.Items.Total_price, &order.Items.Nm_id, &order.Items.Brand, &order.Items.Status)
	if err != nil {
		return nil, err
	}

	return order, nil
}
*/
