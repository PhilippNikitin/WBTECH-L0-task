package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	stan "github.com/nats-io/stan.go"

	"sample-app/models"
)

func main() {
	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "subscriber")
	if err != nil {
		log.Fatalf("Error connecting to NATS Streaming: %v", err)
	}
	defer sc.Close()

	// Подключение к PostgreSQL
	db, err := sql.Open("postgres", "user=admin password=admin dbname=orders sslmode=disable")
	if err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Инициализируем функцию для обработки сообщений в канале NATS Streaming
	handleOrder := func(msg *stan.Msg) {
		var order models.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Fatalf("Error unmarshaling order: %v", err)
		}

		// Сохранение заказа в базе данных PostgreSQL
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Error starting transaction: %v", err)
		}

		_, err = tx.Exec("INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
			order.OrderUID, order.Track_number, order.Entry, order.Locale, order.Internal_signature, order.Customer_id, order.Delivery_service, order.Shardkey, order.Sm_id, order.Date_created, order.Oof_shard)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error saving order to PostgreSQL: %v", err)
		}

		_, err = tx.Exec("INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error saving delivery to PostgreSQL: %v", err)
		}

		_, err = tx.Exec("INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
			order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error saving payment to PostgreSQL: %v", err)
		}

		_, err = tx.Exec("INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			order.OrderUID, order.Items.ChrtID, order.Items.TrackNumber, order.Items.Price, order.Items.Rid, order.Items.Name, order.Items.Sale, order.Items.Size, order.Items.Total_price, order.Items.Nm_id, order.Items.Brand, order.Items.Status)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error saving payment to PostgreSQL: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatalf("Error committing transaction: %v", err)
		}

		fmt.Printf("Saved order: %+v\n", order)
	}

	// Подписка на канал NATS Streaming
	sub, err := sc.Subscribe("orders", handleOrder, stan.DurableName("my-durable"))
	if err != nil {
		log.Fatalf("Error subscribing to NATS Streaming: %v", err)
	}
	defer sub.Unsubscribe()

	// Обработка запросов на получение заказа
	http.HandleFunc("/orders/{order_uid}", func(w http.ResponseWriter, r *http.Request) {
		orderUID := r.URL.Path[len("/orders/"):]
		order, err := getOrderFromDB(db, orderUID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	})

	// Запуск HTTP-сервера
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getOrderFromDB(db *sql.DB, orderUID string) (*models.Order, error) {
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
