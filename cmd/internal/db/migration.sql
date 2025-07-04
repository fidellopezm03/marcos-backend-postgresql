CREATE DATABASE Tienda;
CREATE TABLE IF NOT EXISTS User(
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
);
CREATE TABLE IF NOT EXISTS Content(
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(100) NOT NULL,
    user_id INTEGER REFERENCES User(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS File(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    content_id INTEGER REFERENCES Content(id) ON DELETE CASCADE,
    file_path VARCHAR(255) NOT NULL
);