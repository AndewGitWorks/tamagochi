
Profiles table

| Name           | Type         | Specificator       | QUERY                                          | Key     |
| -------------- | ------------ | ------------------ | ---------------------------------------------- | ------- |
| id             | SERIAL       | SERIAL             |                                                | PRIMARY |
| user_id        | BIGINT       | NOT NULL<br>UNIQUE |                                                | FOREIGN |
| bio            | VARCHAR(100) | NOT NULL           | CHECK char_length(full_name) BETWEEN 3 AND 100 |         |
| address        | VARCHAR(100) | NOT NULL           | CHECK char_length(full_name) BETWEEN 3 AND 100 |         |
| orders_counter | BIGINT       | DEFAULT 0          |                                                |         |
REFERENCE

```Sql
CREATE TABLE tamagochi.profiles(
	id SERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL,
	bio TEXT,
	address TEXT,
	orders_count BIGINT DEFAULT 0,
);
ALTER TABLE tamagochi.profiles
ADD CONSTRAINT uq_profiles_user_id UNIQUE (user_id),
ADD CONSTRAINT fk_profiles_user
    FOREIGN KEY (user_id)
    REFERENCES tamagochi.users(id);

```