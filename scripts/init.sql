-- 1. Segments Table (The "Rules")
CREATE TABLE IF NOT EXISTS segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    rule_logic JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. User Metrics Table (The "State")
CREATE TABLE IF NOT EXISTS user_metrics (
    user_id VARCHAR(255) PRIMARY KEY,
    order_count_total INTEGER DEFAULT 0,
    orders_23d INTEGER DEFAULT 0,
    last_order_at TIMESTAMP WITH TIME ZONE,
    location_tag VARCHAR(100) DEFAULT 'unknown',
    total_spend DECIMAL(12, 2) DEFAULT 0.00,
    ltv DECIMAL(12, 2) DEFAULT 0.00,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Experiments Table (The "Payload")
CREATE TABLE IF NOT EXISTS experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID REFERENCES segments(id),
    experiment_key VARCHAR(100) NOT NULL,
    variant_data JSONB NOT NULL,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed a sample "Power User" segment for the demo
INSERT INTO segments (name, rule_logic) VALUES 
('Power User', '{"and": [{">": [{"var": "orders_23d"}, 25]}]}');