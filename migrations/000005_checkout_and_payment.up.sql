ALTER TABLE tamagochi.orders
    ADD COLUMN delivery_address TEXT;

ALTER TABLE tamagochi.orders
    DROP CONSTRAINT orders_status_check;

ALTER TABLE tamagochi.orders
    ADD CONSTRAINT orders_status_check CHECK (status IN (
        'PENDING',
        'AWAITING_PAYMENT',
        'PAID',
        'PAYMENT_FAILED',
        'DELIVERING',
        'DELIVERED',
        'CANCELLED'
    ));
