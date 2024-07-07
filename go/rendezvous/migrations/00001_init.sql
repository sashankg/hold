-- +goose Up
CREATE TABLE `addresses` (
    peer STRING PRIMARY KEY,
    ma TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE peers;
DROP TABLE addresses;
