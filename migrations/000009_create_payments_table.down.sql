DROP TABLE  IF EXISTS payments;
DROP TYPE   IF EXISTS payment_status;
DROP TYPE   IF EXISTS payment_method;


DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_payments_transaction_id;   
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_method;