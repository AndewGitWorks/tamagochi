UPDATE tamagochi.orders
SET status = 'CANCELLED'
WHERE status = 'AWAITING_PAYMENT';

ALTER TABLE tamagochi.orders
    DROP CONSTRAINT orders_status_check;

ALTER TABLE tamagochi.orders
    ADD CONSTRAINT orders_status_check CHECK (status IN (
        'PENDING',
        'PAID',
        'PAYMENT_FAILED',
        'DELIVERING',
        'DELIVERED',
        'CANCELLED'
    ));

ALTER TABLE tamagochi.orders
    DROP COLUMN delivery_address;
