
| Name          | Type         | Specificator                 | Query                                            | Key     |
| ------------- | ------------ | ---------------------------- | ------------------------------------------------ | ------- |
| id            | SERIAL       | SERIAL                       |                                                  | Primary |
| title         | VARCHAR(100) | NOT NULL<br>UNIQUE           | CHECK char_length(title) BETWEEN 3 AND 100       |         |
| price         | DECIMAL      | NOT NULL                     |                                                  |         |
| seria_title   | VARCHAR(100) | NOT NULL<br>DEFAULT "NEWBIE" | CHECK char_length(seria_title) BETWEEN 3 AND 100 |         |
| bio           | TEXT         | NOT NULL                     |                                                  |         |
| description   | TEXT         | NOT NULL                     |                                                  |         |
| specification | JSONB        |                              |                                                  |         |
```Sql
CREATE TABLE tamagochi.products(
	id SERIAL PRIMARY KEY,
	title VARCHAR(100) NOT NULL UNIQUE CHECK (
		char_lenght(title) BETWEEN 3 AND 100
	),
	price DECIMAL NOT NULL,
	seria_title VARCHAR(100) NOT NULL DEFAULT 'NEWBIE' CHECK (
		char_length(seria_title) BETWEEN 3 AND 100
	),
	bio TEXT,
	description TEXT,
	specifications JSONB,
)
```