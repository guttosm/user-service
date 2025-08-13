-- +goose Up
-- +goose StatementBegin
-- Install UUID and crypto extensions for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create a custom UUID v7 function
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID AS $$
DECLARE
    unix_ts_ms BIGINT;
    uuid_bytes BYTEA;
BEGIN
    -- Get current timestamp in milliseconds since Unix epoch
    unix_ts_ms := EXTRACT(EPOCH FROM NOW()) * 1000;
    
    -- Generate UUID v7: timestamp (48 bits) + version (4 bits) + random (12 bits) + variant (2 bits) + random (62 bits)
    uuid_bytes := 
        -- 48-bit timestamp in milliseconds
        substring(int8send(unix_ts_ms), 3, 6) ||
        -- 4-bit version (7) + 12-bit random
        substring(gen_random_bytes(2), 1, 2) ||
        -- 2-bit variant (10) + 62-bit random  
        substring(gen_random_bytes(8), 1, 8);
    
    -- Set version to 7 (bits 12-15 of time_hi_and_version)
    uuid_bytes := set_byte(uuid_bytes, 6, (get_byte(uuid_bytes, 6) & 15) | 112);
    
    -- Set variant to 10 (bits 6-7 of clock_seq_hi_and_reserved)
    uuid_bytes := set_byte(uuid_bytes, 8, (get_byte(uuid_bytes, 8) & 63) | 128);
    
    RETURN encode(uuid_bytes, 'hex')::UUID;
END;
$$ LANGUAGE plpgsql VOLATILE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove custom UUID v7 function
DROP FUNCTION IF EXISTS uuid_generate_v7();
-- Remove extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
