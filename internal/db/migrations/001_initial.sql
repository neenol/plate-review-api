-- +goose Up
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_user_id TEXT        NOT NULL UNIQUE,
    username      TEXT        NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE plates (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plate_hash TEXT        NOT NULL UNIQUE,
    state      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE reviews (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plate_id   UUID        NOT NULL REFERENCES plates(id),
    user_id    UUID        NOT NULL REFERENCES users(id),
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX reviews_plate_id_idx ON reviews(plate_id);

CREATE TABLE replies (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID        NOT NULL REFERENCES reviews(id),
    parent_reply_id UUID        REFERENCES replies(id),
    user_id         UUID        NOT NULL REFERENCES users(id),
    body            TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX replies_review_id_idx ON replies(review_id);
CREATE INDEX replies_parent_reply_id_idx ON replies(parent_reply_id);

CREATE TABLE votes (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id),
    target_type TEXT        NOT NULL CHECK (target_type IN ('review', 'reply')),
    target_id   UUID        NOT NULL,
    value       INTEGER     NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, target_type, target_id)
);

CREATE INDEX votes_target_idx ON votes(target_type, target_id);

-- +goose Down
DROP TABLE votes;
DROP TABLE replies;
DROP TABLE reviews;
DROP TABLE plates;
DROP TABLE users;
