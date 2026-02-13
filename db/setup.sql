/*
 * schema.sql - TimescaleDB schema
 * 
 * Hypertables for sensor data (vibrations) and detected faults.
 * Automatic time-based partitioning for faster queries and easier management.
 */

-- TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Raw data from vibration sensor (GY-521 on nRF5340DK)
CREATE TABLE IF NOT EXISTS sensor_data (
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),  
    x DOUBLE PRECISION NOT NULL,                   -- X axis of accelerometer
    y DOUBLE PRECISION NOT NULL,                   -- Y axis
    z DOUBLE PRECISION NOT NULL,                   -- Z axis
    score DOUBLE PRECISION,                        -- ML result
    threshold DOUBLE PRECISION,                    -- threshold
    PRIMARY KEY (timestamp)
);

-- Hypertable for sensor data (partitioned by day)
SELECT create_hypertable(
    'sensor_data', 
    'timestamp',
    if_not_exists => TRUE,
    chunk_time_interval => interval '1 day'
);

-- Detected faults
CREATE TABLE IF NOT EXISTS fault_events (
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    score DOUBLE PRECISION NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL                    
);

SELECT create_hypertable(
    'fault_events',
    'timestamp',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_sensor_score
ON sensor_data(score DESC, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_faults_status
ON fault_events(status, timestamp DESC);