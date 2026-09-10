-- +goose Up
CREATE TABLE partners (
    uuid UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    countries TEXT[] NOT NULL DEFAULT '{}',
    device_types TEXT[] NOT NULL DEFAULT '{}',
    min_bid_floor DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (min_bid_floor >= 0),
    blocked_categories TEXT[] NOT NULL DEFAULT '{}'
);

INSERT INTO partners (
    uuid,
    code,
    name,
    endpoint,
    is_enabled,
    countries,
    device_types,
    min_bid_floor,
    blocked_categories
) VALUES
    (
        '123e4567-e89b-12d3-a456-426655440001',
        'dsp-alpha',
        'DSP Alpha',
        'mock://success',
        TRUE,
        ARRAY['RU', 'KZ'],
        ARRAY['mobile', 'desktop'],
        0.5,
        ARRAY['gambling']
    ),
    (
        '123e4567-e89b-12d3-a456-426655440002',
        'dsp-beta',
        'DSP Beta',
        'mock://error',
        TRUE,
        ARRAY['US', 'DE'],
        ARRAY['desktop'],
        1.0,
        ARRAY[]::TEXT[]
    ),
    (
        '123e4567-e89b-12d3-a456-426655440003',
        'dsp-gamma',
        'DSP Gamma',
        'mock://timeout',
        TRUE,
        ARRAY['RU'],
        ARRAY[]::TEXT[],
        1.25,
        ARRAY['adult']
    ),
    (
        '123e4567-e89b-12d3-a456-426655440004',
        'dsp-disabled',
        'DSP Disabled',
        'mock://success',
        FALSE,
        ARRAY[]::TEXT[],
        ARRAY[]::TEXT[],
        0,
        ARRAY[]::TEXT[]
    );

-- +goose Down
DROP TABLE partners;
