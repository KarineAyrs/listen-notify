create table users
(
    id         text not null,
    first_name text not null,
    last_name  text not null,
    unique (id)
);

CREATE OR REPLACE FUNCTION notify_trigger() RETURNS trigger AS
$trigger$
DECLARE
    rec     users;
    dat     users;
    payload TEXT;
BEGIN
    CASE TG_OP
        WHEN 'UPDATE' THEN rec := NEW;
                           dat := OLD;
        WHEN 'INSERT' THEN rec := NEW;
        WHEN 'DELETE' THEN rec := OLD;
        ELSE RAISE EXCEPTION 'Unknown TG_OP: "%". Should not occur!', TG_OP;
        END CASE;
    payload := json_build_object('timestamp', CURRENT_TIMESTAMP, 'action', LOWER(TG_OP), 'db_schema', TG_TABLE_SCHEMA,
                                 'table', TG_TABLE_NAME, 'record', row_to_json(rec), 'old', row_to_json(dat));

    PERFORM pg_notify('users', payload);
    RETURN rec;
END;
$trigger$ LANGUAGE 'plpgsql';

CREATE TRIGGER users_notify
    AFTER INSERT OR UPDATE OR DELETE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE notify_trigger();