# устанавливаем и настраиваем NATS Streaming Server
FROM nats-streaming:latest as nats-streaming

EXPOSE 4222 8222

# строим основной образ: приложение + база данных PostgreSQL
FROM golang:1.22.2-ubuntu24.04

ENV DB_NAME=orders
ENV DB_USER=admin
ENV DB_PASSWORD=admin

# устанавливаем PostgreSQL, создаем и настраиваем базу данных, своего пользователя и таблицы
RUN sudo apt-get update
RUN sudo apt-get install postgresql postgresql-contrib
RUN sudo -u postgres psql -c "CREATE DATABASE $DB_NAME;"
RUN sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
RUN sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
RUN sudo -u postgres psql -c "GRANT CREATE ON SCHEMA public TO $DB_USER;"
RUN sudo -u postgres psql -d $DB_NAME -c "\
    CREATE TABLE orders (
        order_uid VARCHAR(19) PRIMARY KEY,
        track_number VARCHAR(14) NOT NULL,
        entry VARCHAR(4) NOT NULL,
        locale VARCHAR(3) NOT NULL,
        internal_signature VARCHAR(10),
        customer_id VARCHAR(20) NOT NULL,
        delivery_service VARCHAR(20) NOT NULL,
        shardkey VARCHAR(10) NOT NULL,
        sm_id INTEGER NOT NULL,
        date_created TIMESTAMP NOT NULL,
        oof_shard VARCHAR(10) NOT NULL
    );"
RUN sudo -u postgres psql -d $DB_NAME -c "\
    CREATE TABLE delivery (
        order_uid VARCHAR(19) PRIMARY KEY,
    	name VARCHAR(70) NOT NULL,
    	phone VARCHAR(20) NOT NULL,
    	zip VARCHAR(20) NOT NULL,
    	city VARCHAR(100) NOT NULL,
    	address VARCHAR(200) NOT NULL,
    	region VARCHAR(100) NOT NULL,
    	email VARCHAR(100) NOT NULL,
    	FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
    );"
RUN sudo -u postgres psql -d $DB_NAME -c "\
    CREATE TABLE payment (
        order_uid VARCHAR(19) PRIMARY KEY,
    	transaction VARCHAR(19) NOT NULL,
    	request_id VARCHAR(50),
    	currency VARCHAR(10) NOT NULL,
    	provider VARCHAR(20) NOT NULL,
    	amount INTEGER NOT NULL,
    	payment_dt BIGINT NOT NULL,
    	bank VARCHAR(50) NOT NULL,
    	delivery_cost INTEGER NOT NULL,
    	goods_total INTEGER NOT NULL,
    	custom_fee INTEGER NOT NULL,
    	FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
    );"
RUN sudo -u postgres psql -d $DB_NAME -c "\
    CREATE TABLE items (
        id SERIAL PRIMARY KEY,  // нет в стракте Order
    	order_uid VARCHAR(19) NOT NULL,
    	chrt_id INTEGER NOT NULL,
    	track_number VARCHAR(50) NOT NULL,
    	price DECIMAL(10,2) NOT NULL,
    	rid VARCHAR(50) NOT NULL,
    	name VARCHAR(100) NOT NULL,
    	sale INTEGER NOT NULL,
    	size VARCHAR(10) NOT NULL,
    	total_price DECIMAL(10,2) NOT NULL,
    	nm_id INTEGER NOT NULL,
    	brand VARCHAR(100) NOT NULL,
    	status INTEGER NOT NULL,
    	FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
    );"


COPY . .
COPY --from=nats-streaming /nats-streaming-server /nats-streaming-server

WORKDIR /app
RUN chmod +x /app/start.sh

# запускаем NATS Streaming Server и приложение
CMD ["/nats-streaming-server", "/app/start.sh"]
