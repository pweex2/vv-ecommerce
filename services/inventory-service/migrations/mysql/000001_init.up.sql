CREATE TABLE IF NOT EXISTS inventories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT UNSIGNED,
    sku VARCHAR(255) NOT NULL UNIQUE,
    quantity INT,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);

CREATE TABLE IF NOT EXISTS inventory_deduction_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(255),
    request_id VARCHAR(255) NOT NULL UNIQUE,
    sku VARCHAR(255),
    trace_id VARCHAR(255),
    quantity INT,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_order_id (order_id),
    INDEX idx_sku (sku),
    INDEX idx_trace_id (trace_id)
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSON NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    trace_id VARCHAR(255),
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Seed data
INSERT INTO inventories (product_id, sku, quantity) 
VALUES (1, 'PHONE-001', 100) 
ON DUPLICATE KEY UPDATE quantity = quantity;
