CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL,
    value TEXT NOT NULL
);

