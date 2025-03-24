-- Изменение типа колонок на BIGINT для хранения миллисекунд
ALTER TABLE schedules
ALTER COLUMN frequency TYPE BIGINT,
    ALTER COLUMN duration TYPE BIGINT;
-- Переименование колонок дат
ALTER TABLE schedules
    RENAME COLUMN start_date TO start_time;
ALTER TABLE schedules
    RENAME COLUMN end_date TO end_time;