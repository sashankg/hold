-- +goose Up
CREATE TABLE `collections` (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE `collection_fields` (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    ref INTEGER,
    is_list INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE collections;
DROP TABLE collection_field;
