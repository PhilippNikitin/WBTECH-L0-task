package logging

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	// Открываем или создаем файл для логирования
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Ошибка при создании или обращении к log file: %v", err)
	}

	// Создаем новый логгер, использующий файл для вывода и включающий дату и время
	Logger = log.New(logFile, "", log.Ldate|log.Ltime)
}
