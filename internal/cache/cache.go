package cache

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/logging"
	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/models"
)

var OrderCache sync.Map

func GetOrderCacheSize(m *sync.Map) int {
	count := 0
	m.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func RestoreCacheFromDb(db *sql.DB, cache *sync.Map) error {
	// выполняем запрос к базе данных для получения всех OrderUID из базы данных
	rows, err := db.Query("SELECT order_uid FROM orders")
	if err != nil {
		return err
	}
	defer rows.Close()
	// итерируемся по всей базе данных, проверяем каждый заказ на наличие в кэше
	for rows.Next() {
		var orderUID string
		err := rows.Scan(&orderUID)
		if err != nil {
			return err
		}
		// проверяем, если OrderUID заказа есть в OrderCache
		_, found := cache.Load(orderUID)
		if !found {
			// если ключ не найден, загружаем полные данные заказа из базы данных

			order := &models.Order{OrderUID: orderUID}

			// Получение данных из таблицы orders
			err := db.QueryRow("SELECT track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1", orderUID).
				Scan(&order.Track_number, &order.Entry, &order.Locale, &order.Internal_signature, &order.Customer_id, &order.Delivery_service, &order.Shardkey, &order.Sm_id, &order.Date_created, &order.Oof_shard)
			if err != nil {
				return err
			}

			// Получение данных из таблицы delivery
			err = db.QueryRow("SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1", orderUID).
				Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
			if err != nil {
				return err
			}

			// Получение данных из таблицы payment
			err = db.QueryRow("SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1", orderUID).
				Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
			if err != nil {
				return err
			}

			// Получение данных из таблицы items
			err = db.QueryRow("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", orderUID).
				Scan(&order.Items.ChrtID, &order.Items.TrackNumber, &order.Items.Price, &order.Items.Rid, &order.Items.Name, &order.Items.Sale, &order.Items.Size, &order.Items.Total_price, &order.Items.Nm_id, &order.Items.Brand, &order.Items.Status)
			if err != nil {
				return err
			}

			// маршалим order в JSON
			orderJson, err := json.Marshal(order)
			if err != nil {
				return err
			}

			// сохраняем заказ в кэш
			cache.Store(orderUID, orderJson)
			log.Printf("Заказ %s добавлен в кэш.\n", orderUID)
			logging.Logger.Printf("Заказ %s добавлен в кэш.\n", orderUID)
		}

	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
