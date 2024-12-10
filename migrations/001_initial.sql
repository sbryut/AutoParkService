--таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) CHECK (role IN ('admin', 'user')) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица персонала
CREATE TABLE IF NOT EXISTS auto_personal (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    father_name VARCHAR(20) NOT NULL
);

-- Таблица автомобилей
CREATE TABLE IF NOT EXISTS auto (
    id SERIAL PRIMARY KEY,
    num VARCHAR(20) UNIQUE NOT NULL,
    color VARCHAR(20) NOT NULL,
    mark VARCHAR(20) NOT NULL,
    personal_id INTEGER NOT NULL,
    CONSTRAINT fk_auto_personal FOREIGN KEY (personal_id) REFERENCES auto_personal(id) ON DELETE CASCADE
);

-- Таблица маршрутов
CREATE TABLE IF NOT EXISTS routes (
    id SERIAL PRIMARY KEY,
    start_point VARCHAR(50) NOT NULL,
    end_point VARCHAR(50) NOT NULL
);

-- Таблица журнала рейсов
CREATE TABLE IF NOT EXISTS journal (
    id SERIAL PRIMARY KEY,
    time_out TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    time_in TIMESTAMP WITHOUT TIME ZONE,
    route_id INTEGER NOT NULL,
    auto_id INTEGER NOT NULL,
    CONSTRAINT fk_journal_routes FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
    CONSTRAINT fk_journal_auto FOREIGN KEY (auto_id) REFERENCES auto(id) ON DELETE CASCADE,
    CONSTRAINT chk_time_valid CHECK (time_in IS NULL OR time_in >= time_out)
);

-- Представление для журнала
CREATE OR REPLACE VIEW journal_view AS
SELECT
    j.id AS journal_id,
    j.time_out,
    j.time_in,
    r.start_point,
    r.end_point,
    a.num AS auto_number,
    a.mark AS auto_mark,
    p.first_name || ' ' || p.last_name AS driver_name
FROM journal j
         INNER JOIN routes r ON j.route_id = r.id
         INNER JOIN auto a ON j.auto_id = a.id
         INNER JOIN auto_personal p ON a.personal_id = p.id;

-- Триггер: запрет отправки в рейс водителя, который еще не вернулся
CREATE OR REPLACE FUNCTION check_driver_availability()
    RETURNS TRIGGER AS $$
DECLARE
    active_count INT;
    driver_id INT;
BEGIN
    SELECT personal_id INTO driver_id
    FROM auto
    WHERE id = NEW.auto_id;

    SELECT COUNT(*) INTO active_count
    FROM journal j
    WHERE j.auto_id IN (
        SELECT id FROM auto WHERE personal_id = driver_id
    )
      AND j.time_in IS NULL;

    IF active_count > 0 THEN
        RAISE EXCEPTION 'Водитель с ID % не может быть отправлен в рейс, пока не вернется с предыдущего маршрута.', driver_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_driver_double_booking
    BEFORE INSERT ON journal
    FOR EACH ROW
EXECUTE FUNCTION check_driver_availability();

-- Триггер: запрет отправки автомобиля, который еще не вернулся
CREATE OR REPLACE FUNCTION TIME_IN_CHECK()
    RETURNS TRIGGER AS
$$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM JOURNAL
        WHERE AUTO_ID = NEW.AUTO_ID
          AND TIME_IN IS NULL
    ) THEN
        RAISE EXCEPTION 'Автомобиль % еще не вернулся в парк, отправка невозможна', NEW.AUTO_ID;
    END IF;
    RETURN NEW;
END;
$$
    LANGUAGE PLPGSQL;

CREATE TRIGGER PREVENT_AUTO_SENDING
    BEFORE INSERT ON JOURNAL
    FOR EACH ROW
EXECUTE FUNCTION TIME_IN_CHECK();

--Триггер: время прибытия не может быть меньше времени отправления
CREATE OR REPLACE FUNCTION CHECK_ARRIVAL_TIME()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.time_in < NEW.time_out THEN
        RAISE EXCEPTION 'Время прибытия не может быть меньше времени отправления: time_in = %, time_out = %', NEW.time_in, NEW.time_out;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER PREVENT_INVALID_ARRIVAL_TIME
    BEFORE INSERT OR UPDATE ON journal
    FOR EACH ROW
EXECUTE FUNCTION CHECK_ARRIVAL_TIME();

--триггер: запрет удаления водителя при наличии ссылок
CREATE OR REPLACE FUNCTION UNDELETE()
    RETURNS TRIGGER AS
$$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM AUTO
        WHERE PERSONAL_ID = OLD.ID
    ) THEN
        RAISE EXCEPTION 'Невозможно удалить водителя с ID %, так как на него существуют ссылки', OLD.ID;
    END IF;
    RETURN OLD;
END;
$$
    LANGUAGE PLPGSQL;

CREATE TRIGGER TRIGGER_UNDELETE
    BEFORE DELETE ON AUTO_PERSONAL
    FOR EACH ROW
EXECUTE FUNCTION UNDELETE();

-- Функция получения количества машин на каждом маршруте за все время
CREATE OR REPLACE FUNCTION GET_ROUTES_VEHICLE_COUNT()
    RETURNS TABLE(
                     ROUTE_NAME TEXT,
                     VEHICLE_COUNT BIGINT
                 )
    LANGUAGE PLPGSQL
AS $$
BEGIN
    RETURN QUERY
        SELECT
            CONCAT(r.start_point, ' - ', r.end_point) AS ROUTE_NAME,
            COUNT(DISTINCT j.auto_id) AS VEHICLE_COUNT
        FROM routes r
                 JOIN journal j ON r.id = j.route_id
        GROUP BY r.start_point, r.end_point
        ORDER BY VEHICLE_COUNT DESC;
END;
$$;

