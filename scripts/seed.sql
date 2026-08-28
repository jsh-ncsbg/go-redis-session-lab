\set ON_ERROR_STOP on

\if :{?benchmark_count}
\else
\set benchmark_count 10000
\endif

BEGIN;

-- reset only rows owned by this benchmark seed.
DELETE FROM users
WHERE LEFT(username, 10) = 'benchmark_'
	AND RIGHT(email, LENGTH('@benchmark.invalid')) = '@benchmark.invalid';

WITH generated_users AS MATERIALIZED (
	SELECT
		g,
		'benchmark_' || g || '_' || LEFT(md5('username:' || g), 8) AS username,
		'benchmark_' || g || '_' || LEFT(md5('email:' || g), 12) || '@benchmark.invalid' AS email,
		'benchmark-token-' || g AS raw_token
	FROM generate_series(1, :'benchmark_count'::integer) AS series(g)
),
inserted_users AS (
	INSERT INTO users (username, email, password_hash, is_active)
	SELECT
		username,
		email,
		encode(sha256(convert_to('benchmark-password:' || g, 'UTF8')), 'hex'),
		TRUE
	FROM generated_users
	RETURNING id, username
)
INSERT INTO sessions (token_hash, user_id, csrf_token_hash, created_at, expires_at)
SELECT
	encode(sha256(convert_to(g.raw_token, 'UTF8')), 'hex'),
	u.id,
	encode(sha256(convert_to('benchmark-csrf:' || g.g, 'UTF8')), 'hex'),
	NOW(),
	NOW() + INTERVAL '7 days'
FROM generated_users g
JOIN inserted_users u USING (username);

COMMIT;

SELECT COUNT(*) AS benchmark_sessions
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE LEFT(u.username, 10) = 'benchmark_'
	AND RIGHT(u.email, LENGTH('@benchmark.invalid')) = '@benchmark.invalid';
