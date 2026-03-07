CREATE TABLE IF NOT EXISTS users (
    id CHAR(26) PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    phone_number VARCHAR(20),
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    disabled BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create partial index for non-deleted users (improves performance)
CREATE INDEX IF NOT EXISTS idx_users_not_deleted 
ON users(email) 
WHERE deleted_at IS NULL;

-- Create index for deleted users (useful for audit queries)
CREATE INDEX IF NOT EXISTS idx_users_deleted 
ON users(deleted_at) 
WHERE deleted_at IS NOT NULL;
