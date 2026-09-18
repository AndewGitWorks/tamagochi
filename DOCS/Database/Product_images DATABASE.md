
| Name        | Type         | Specificator                       | Query Key                   |
| ----------- | ------------ | ---------------------------------- | --------------------------- |
| id          | SERIAL       | PRIMARY KEY                        | Primary                     |
| product_id  | INTEGER      | NOT NULL                           | FOREIGN KEY → `products.id` |
| storage_key | VARCHAR(500) | NOT NULL UNIQUE                    | Storage identifier/path     |
| sort_order  | INTEGER      | NOT NULL DEFAULT 0                 | CHECK `sort_order >= 0`     |
| created_at  | TIMESTAMP    | NOT NULL DEFAULT CURRENT_TIMESTAMP |                             |
```SQL
CREATE TABLE tamagochi.product_images (
    id SERIAL PRIMARY KEY,

    product_id INTEGER NOT NULL,

    storage_key VARCHAR(500) NOT NULL UNIQUE,

    sort_order INTEGER NOT NULL DEFAULT 0
        CHECK (sort_order >= 0),

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_product_images_product
        FOREIGN KEY (product_id)
        REFERENCES tamagochi.products(id)
        ON DELETE CASCADE
);
```