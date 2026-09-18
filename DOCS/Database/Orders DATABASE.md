
| Name        | Type          | Specificator       | Query                                                                                                             | Key     |
| ----------- | ------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------- | ------- |
| id          | SERIAL        | SERIAL             |                                                                                                                   | Primary |
| user_id     | INTEGER       |                    |                                                                                                                   | FOREIGN |
| number      | VARCHAR(64)   | NOT NULL<br>UNIQUE | CHECK char_length(number) BETWEEN 3 AND 64                                                                        |         |
| total_price | DECIMAL(10,2) | NOT NULL           | CHECK (total_price >= 0)                                                                                          |         |
| created_at  | TIMESTAMP     | NOT NULL           | DEFAULT CURRENT_TIMESTAMP                                                                                         |         |
| status      | VARCHAR(30)   | NOT NULL           | DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PAID', 'PAYMENT_FAILED', 'DELIVERING', 'DELIVERED', 'CANCELLED')) |         |
```Sql
CREATE TABLE tamagochi.orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    number VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    total_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
	CONSTRAINT fk_orders_user
	        FOREIGN KEY (user_id)
	        REFERENCES tamagochi.users(id)
	        ON DELETE RESTRICT
);
```