CREATE UNIQUE INDEX orders_one_pending_cart_per_user_idx
    ON tamagochi.orders (user_id)
    WHERE status = 'PENDING';
