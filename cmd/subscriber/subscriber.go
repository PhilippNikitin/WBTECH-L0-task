package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	stan "github.com/nats-io/stan.go"

	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/cache"
	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/logging"
	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/models"
	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/validation"
)

func runServerProcessesExceptNATSStreaming() {

	// Подключение к PostgreSQL
	db, err := models.GetDBConnection()
	if err != nil {
		logging.Logger.Fatalf("%v", err)
	}
	defer db.Close()

	// вывод текущего размера кэша до восстановления
	log.Printf("Размер кэша до восстановления: %d\n", cache.GetOrderCacheSize(&cache.OrderCache))

	// проводим восстановление кэша из базы данных
	err = cache.RestoreCacheFromDb(db, &cache.OrderCache)
	if err != nil {
		logging.Logger.Fatalf("Ошибка при восстановлении кэша из базы данных PostgreSQL: %v", err)
	}

	// вывод текущего размера кэша после восстановления
	log.Printf("Размер кэша после восстановления: %d\n", cache.GetOrderCacheSize(&cache.OrderCache))

	// Определяем хэндлер для получения данных о заказе из кэша по переданному с клиента OrderUID
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Подгружаем index.html
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "cmd/subscriber/index.html")
			return
		}
		// получаем orderUID из запроса
		orderUID := r.URL.Query().Get("order_uid")
		// ищем заказ в кэше
		orderJSON, ok := cache.OrderCache.Load(orderUID)
		// обрабатываем ошибку, если заказ не найден в кэше
		if !ok {
			err := fmt.Errorf("заказ не найден")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(orderJSON.([]byte))
		// обрабатываем ошибку, если произошла ошибка во время записи заказа в response
		if err != nil {
			err := fmt.Errorf("error writing response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Запуск HTTP-сервера
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func main() {

	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "subscriber")
	if err != nil {
		logging.Logger.Printf("Ошибка подключения к NATS Streaming серверу: %v", err)
		runServerProcessesExceptNATSStreaming()
		return
	}
	defer sc.Close()

	// Подключение к PostgreSQL
	db, err := models.GetDBConnection()
	if err != nil {
		logging.Logger.Fatalf("%v", err)
	}
	defer db.Close()

	// Инициализируем функцию для обработки сообщений в канале NATS Streaming
	handleOrder := func(msg *stan.Msg) {
		// Десериализуем данные из заказа
		var order models.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			logging.Logger.Printf("Ошибка при десериализации: %v", err)
			return
		}

		// Валидация полученных данных
		err = validation.ValidateOrder(order)
		if err != nil {
			logging.Logger.Printf("Заказ не прошел валидацию: %v", err)
			return
		}

		// Сохраняем данные по заказу в кэш в случае, если этих данных там нет
		_, found := cache.OrderCache.Load(order.OrderUID)
		if !found {
			cache.OrderCache.Store(order.OrderUID, msg.Data)
			log.Printf("Заказ %s сохранен в кэш.\n", order.OrderUID)
			logging.Logger.Printf("Заказ %s сохранен в кэш.\n", order.OrderUID)
			logging.Logger.Printf("Текущий размер кэша: %d\n", cache.GetOrderCacheSize(&cache.OrderCache))
		}

		// Сохраняем заказ в базе данных PostgreSQL
		err = models.SaveOrderInDatabase(db, order)
		if err != nil {
			logging.Logger.Fatalf("Ошибка при сохранении заказа в базу данных: %v", err)
		}
	}

	// Подписка на канал NATS Streaming
	sub, err := sc.Subscribe("orders", handleOrder, stan.DurableName("durable"))
	if err != nil {
		logging.Logger.Fatalf("Ошибка при создании подписки (subscribing) на NATS Streaming сервер: %v", err)
	}
	defer sub.Unsubscribe()

	// Запускаем процессы сервера, не связанные с NATS-streaming
	runServerProcessesExceptNATSStreaming()

}
