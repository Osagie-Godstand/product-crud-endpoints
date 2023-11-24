DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'products') THEN
        CREATE TABLE IF NOT EXISTS products (
            id UUID PRIMARY KEY,
            brand VARCHAR(255) NOT NULL,
            description VARCHAR(255) NOT NULL,
            colour VARCHAR(255) NOT NULL,
            size VARCHAR(255) NOT NULL, 
            price DOUBLE PRECISION,
            sku VARCHAR(255) UNIQUE NOT NULL
        );
    END IF;
END $$;

