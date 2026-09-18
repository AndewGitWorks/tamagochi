| Name       | Type          | Specificator | Query Key                   |
| ---------- | ------------- | ------------ | --------------------------- |
| id         | SERIAL        | PRIMARY KEY  | Primary                     |
| order_id   | INTEGER       | NOT NULL     | FOREIGN KEY → `orders.id`   |
| product_id | INTEGER       | NOT NULL     | FOREIGN KEY → `products.id` |
| quantity   | INTEGER       | NOT NULL     | CHECK quantity > 0          |
| price      | DECIMAL(10,2) | NOT NULL     | CHECK price >= 0            |


```Sql
CREATE TABLE tamagochi.order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
	CONSTRAINT fk_order_items_order
	        FOREIGN KEY (order_id)
	        REFERENCES tamagochi.orders(id)
	        ON DELETE CASCADE,
	
	CONSTRAINT fk_order_items_product
	        FOREIGN KEY (product_id)
	        REFERENCES tamagochi.products(id)
	        ON DELETE RESTRICT,
	
	CONSTRAINT unique_order_product
	        UNIQUE (order_id, product_id)
);
```