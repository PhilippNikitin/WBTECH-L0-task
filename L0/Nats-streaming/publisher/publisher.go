package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/nats-io/stan.go"

	"sample-app/models"
)

// Структура заказа задана в файле models.go

// Определяем функцию для создания случайной строки
func makeSampleString(length int, t string) string {

	// Создаем случайный генератор
	rand.Seed(time.Now().UnixNano())

	// Определяем набор допустимых символов
	chars := ""
	switch t {
	case "alfa":
		chars = "abcdefghijklmnopqrstuvwxyz"
	case "num":
		chars = "0123456789"
	case "alfanum":
		chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	}

	// Генерируем случайную последовательность
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return fmt.Sprint(string(result))
}

// Определяем функцию для создания фейкового заказа
func makeFakeOrder() *models.Order {
	startTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC)

	deliveryInfo := models.Delivery{
		Name:    gofakeit.FirstName() + " " + gofakeit.LastName(),
		Phone:   gofakeit.Phone(),
		Zip:     gofakeit.Zip(),
		City:    gofakeit.City(),
		Address: gofakeit.Address().Street,
		Region:  gofakeit.Address().State,
		Email:   gofakeit.Email(),
	}

	paymentInfo := models.Payment{
		Transaction:  makeSampleString(19, "alfanum"),
		RequestID:    "",
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       gofakeit.Number(1, 10000),
		PaymentDt:    gofakeit.Number(1000000000, 9999999999),
		Bank:         makeSampleString(5, "alfa"),
		DeliveryCost: gofakeit.Number(100, 10000),
		GoodsTotal:   gofakeit.Number(1, 1000),
		CustomFee:    0,
	}

	itemsInfo := models.Items{
		ChrtID:      gofakeit.Number(1000000, 9999999),
		TrackNumber: "WBILMTESTTRACK",
		Price:       float64(gofakeit.Number(1, 10000)),
		Rid:         makeSampleString(21, "alfanum"),
		Name:        gofakeit.Word(),
		Sale:        gofakeit.Number(1, 99),
		Size:        makeSampleString(1, "num"),
		Total_price: float64(gofakeit.Number(1, 10000000)),
		Nm_id:       gofakeit.Number(1, 10000000),
		Brand:       gofakeit.Company(),
		Status:      gofakeit.Number(1, 1000),
	}

	order := &models.Order{
		OrderUID:           makeSampleString(19, "alfanum"),
		Track_number:       "WBILMTESTTRACK",
		Entry:              "WBIL",
		Locale:             "en",
		Internal_signature: "",
		Customer_id:        "test",
		Delivery_service:   makeSampleString(5, "alfa"),
		Shardkey:           makeSampleString(1, "num"),
		Sm_id:              gofakeit.Number(1, 99),
		Date_created:       gofakeit.DateRange(startTime, endTime),
		Oof_shard:          makeSampleString(1, "num"),
		Delivery:           deliveryInfo, // deliveryInfo - структура, ранее созданная в текущей функции
		Payment:            paymentInfo,
		Items:              itemsInfo,
	}

	return order
}

func main() {
	// Подключение к Nats-streaming
	sc, err := stan.Connect("test-cluster", "publisher")
	if err != nil {
		log.Fatalf("Can't connect: %v", err)
	}
	defer sc.Close()

	// Публикация заказа в Nats-streaming
	data, err := json.Marshal(makeFakeOrder())
	if err != nil {
		log.Fatalf("Failed to marshal order: %v", err)
	}

	err = sc.Publish("orders", data)
	if err != nil {
		log.Fatalf("Failed to publish order: %v", err)
	}

	fmt.Println("Тестовый заказ был отправлен в Nats-streaming")
}
