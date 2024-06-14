#!/bin/bash

# Запускаем subscriber
go run cmd/subscriber/subscriber.go &

# Запускаем publisher
go run cmd/publisher/publisher.go &

# Ждем завершения процессов
wait
