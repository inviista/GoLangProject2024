CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    comment text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments_books (
    comment_id bigint NOT NULL REFERENCES comments ON DELETE CASCADE,
    book_id bigint NOT NULL REFERENCES books ON DELETE CASCADE,
    PRIMARY KEY(comment_id, book_id)
);