CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL,
    bucket VARCHAR(255) NOT NULL,
    object_key VARCHAR(512) NOT NULL UNIQUE,
    mime_type VARCHAR(127) NOT NULL,
    size BIGINT NOT NULL CHECK (size > 0),
    file_type VARCHAR(50) NOT NULL DEFAULT 'document',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_owner_id ON files(owner_id);
CREATE INDEX idx_files_object_key ON files(object_key);
