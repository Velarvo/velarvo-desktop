-- +goose Up
-- +goose StatementBegin

INSERT INTO app_meta (key, value) VALUES
    ('ui.language', 'en')
    ON CONFLICT(key) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM app_meta WHERE key = 'ui.language';
-- +goose StatementEnd
