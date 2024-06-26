CREATE TABLE IF NOT EXISTS users_books (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    book_id bigint NOT NULL REFERENCES books ON DELETE CASCADE,
    PRIMARY KEY (user_id, book_id)
);