CREATE SCHEMA tamagochi;

CREATE TABLE tamagochi.users (
    id SERIAL PRIMARY KEY,

    name VARCHAR(100) NOT NULL
        CHECK (char_length(name) BETWEEN 3 AND 100),

    surname VARCHAR(100) NOT NULL
        CHECK (char_length(surname) BETWEEN 3 AND 100),

    login VARCHAR(24) NOT NULL UNIQUE
        CHECK (char_length(login) BETWEEN 3 AND 24),

    email VARCHAR(100) NOT NULL UNIQUE
        CHECK (char_length(email) BETWEEN 3 AND 100),

    password_hashed VARCHAR(64) NOT NULL,

    phone_number VARCHAR NOT NULL
        CHECK (
            char_length(phone_number) BETWEEN 10 AND 15
            AND phone_number ~ '^\+[0-9]+$'
        ),

    version BIGINT NOT NULL DEFAULT 1,

    role VARCHAR(50) NOT NULL DEFAULT 'USER'
);


CREATE TABLE tamagochi.profiles (
    id SERIAL PRIMARY KEY,

    user_id INTEGER NOT NULL,

    bio TEXT,
    address TEXT,

    orders_count BIGINT NOT NULL DEFAULT 0
        CHECK (orders_count >= 0),

    CONSTRAINT uq_profiles_user_id
        UNIQUE (user_id),

    CONSTRAINT fk_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES tamagochi.users(id)
        ON DELETE CASCADE
);


CREATE TABLE tamagochi.products (
    id SERIAL PRIMARY KEY,

    title VARCHAR(100) NOT NULL UNIQUE
        CHECK (char_length(title) BETWEEN 3 AND 100),

    price DECIMAL(10, 2) NOT NULL
        CHECK (price >= 0),

    seria_title VARCHAR(100) NOT NULL DEFAULT 'NEWBIE'
        CHECK (char_length(seria_title) BETWEEN 3 AND 100),

    bio TEXT NOT NULL,
    description TEXT NOT NULL,
    specifications JSONB
);


CREATE TABLE tamagochi.orders (
    id SERIAL PRIMARY KEY,

    user_id INTEGER NOT NULL,

    number VARCHAR(64) NOT NULL UNIQUE
        CHECK (char_length(number) BETWEEN 3 AND 64),

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    status VARCHAR(30) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN (
            'PENDING',
            'PAID',
            'PAYMENT_FAILED',
            'DELIVERING',
            'DELIVERED',
            'CANCELLED'
        )),

    total_price DECIMAL(10, 2) NOT NULL DEFAULT 0
        CHECK (total_price >= 0),

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id)
        REFERENCES tamagochi.users(id)
        ON DELETE RESTRICT
);


CREATE TABLE tamagochi.order_items (
    id SERIAL PRIMARY KEY,

    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,

    quantity INTEGER NOT NULL
        CHECK (quantity > 0),

    price DECIMAL(10, 2) NOT NULL
        CHECK (price >= 0),

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
