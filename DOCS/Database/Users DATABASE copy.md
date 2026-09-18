
USERS TABLE

| Name            | Type         | Specificator          | QUERY                                                                                 | Key     |
| --------------- | ------------ | --------------------- | ------------------------------------------------------------------------------------- | ------- |
| id              | SERIAL       | SERIAL                |                                                                                       | PRIMARY |
| name            | VARCHAR(100) | NOT NULL              | CHECK char_length(full_name) BETWEEN 3 AND 100                                        |         |
| surname         | VARCHAR(100) | NOT NULL              | CHECK char_length(full_name) BETWEEN 3 AND 100                                        |         |
| login           | VARCHAR(24)  | NOT NULL<br>UNIQUE    | CHECK char_length(full_name) BETWEEN 3 AND 100                                        |         |
| email           | VARCHAR(100) | NOT NULL<br>UNIQUE    | CHECK char_length(email) BETWEEN 3 AND 100                                            |         |
| password_hashed | VARCHAR(64)  | NOT NULL              |                                                                                       |         |
| phone_number    | VARCHAR      | NOT NULL              | CHECK char_length(phone_number) BETWEEN 10 AND 15<br>CHECK phone_number ± '^+[0-9]+$' |         |
| version         | BIGINT       | DEFAULT 1<br>NOT NULL |                                                                                       |         |
| role            | VARCHAR(50)  | NOT NULL              | DEFAULT 'USER'                                                                        |         |

### SQL CODE FOR INITIALIZATION

```Sql
CREATE TABLE tamagochi.users(
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL CHECK (char_lenght(name)) BETWEEN 3 AND 100),
	surname VARCHAR(100) NOT NULL CHECK (char_lenght(name)) BETWEEN 3 AND 100),
	login VARCHAR(24) NOT NULL UNIQUE CHECK (char_lenght(name)) BETWEEN 3 AND 24),
	email VARCHAR(100) NOT NULL UNIQUE CHECK (char_lenght(email)) BETWEEN 3 AND 100,
	password_hashed VARCHAR(64) NOT NULL,
	phone_number VARCHAR NOT NULL CHECK(
	char_length(phone_number) BETWEEN 10 AND 15
	AND phone_number ± '^+[0-9]+$'),
	version BIGINT NOT NULL DEFAULT 1,
	role VARCHAR(50) NOT NULL DEFAULT 'USER'
)
```