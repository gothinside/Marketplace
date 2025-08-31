CREATE TABLE IF NOT EXISTS users(
    ID       INT PRIMARY KEY
    EMAIL    VARCHAR(2000) UNIQUE,
    USERNAME VARCHAR(2000) UNIQUE,
    PASSWORD VARCHAR(2000)
)

CREATE TABLE IF NOT EXISTS roles(
    role_id INT PRIMARY KEY,
    role_name VARCHAR(200),
)

CREATE TABLE IF NOT EXISTS user_role(
    user_id INT REFERENCES users(id)
    role_id INT REFERENCES roles(role_id)
)


INSERT INTO roles VALUES
(1, 'user'),
(2, 'admin')
(3, 'superuser')

INSERT INTO users VALUES
(1, 'admin@example.com', 'admin', 'pass'),

INSERT INTO user_role VALUES 
(1, 2)