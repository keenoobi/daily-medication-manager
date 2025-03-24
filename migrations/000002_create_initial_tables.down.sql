-- Возвращаем оригинальные имена колонок
ALTER TABLE schedules
    RENAME COLUMN start_time TO start_date;
ALTER TABLE schedules
    RENAME COLUMN end_time TO end_date;
-- Возвращаем оригинальный тип колонок
ALTER TABLE schedules
ALTER COLUMN frequency TYPE INTEGER;
ALTER TABLE schedules
ALTER COLUMN duration TYPE INTEGER;