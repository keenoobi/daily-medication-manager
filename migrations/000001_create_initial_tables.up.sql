-- migrations/000001_create_initial_tables.up.sql
CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    medication TEXT NOT NULL,
    frequency INT NOT NULL,
    duration INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);