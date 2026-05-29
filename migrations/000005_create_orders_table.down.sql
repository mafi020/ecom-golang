DROP TABLE IF EXISTS orders;
DROP TYPE  IF EXISTS order_status;

DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_status;

DROP INDEX IF EXISTS idx_orders_tracking_number;
DROP INDEX IF EXISTS idx_orders_courier_partner;