| Name       | Type      | Specificator                       | Query Key                           |
| ---------- | --------- | ---------------------------------- | ----------------------------------- |
| user_id    | INTEGER   | NOT NULL                           | FOREIGN KEY → `users.id`            |
| product_id | INTEGER   | NOT NULL                           | FOREIGN KEY → `products.id`         |
| created_at | TIMESTAMP | NOT NULL DEFAULT CURRENT_TIMESTAMP |                                     |
| —          | —         | —                                  | PRIMARY KEY `(user_id, product_id)` |
```Sql
CREATE TABLE tamagochi.user_liked_products (
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_user_liked_products
        PRIMARY KEY (user_id, product_id),

    CONSTRAINT fk_user_liked_products_user
        FOREIGN KEY (user_id)
        REFERENCES tamagochi.users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_user_liked_products_product
        FOREIGN KEY (product_id)
        REFERENCES tamagochi.products(id)
        ON DELETE CASCADE
);
```