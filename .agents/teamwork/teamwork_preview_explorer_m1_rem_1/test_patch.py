import subprocess
import json
import os
import sys

# Test the proposed docker-compose transformations against challenger's checks
proposed_compose = """
services:
  db:
    image: postgres:16-alpine
    container_name: asesamo-db
    restart: unless-stopped
    ports:
      - "${DB_PORT:-5432}:5432"
    environment:
      POSTGRES_USER: ${DB_USER:-asesamo_admin}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
      POSTGRES_DB: ${DB_NAME:-asesamo_db}
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - asesamo_pgdata:/var/lib/postgresql/data
      - ./db/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
      interval: 3s
      timeout: 3s
      retries: 5
      start_period: 10s
    networks:
      - asesamo-net

  migrate:
    image: migrate/migrate:v4.17.0
    restart: "no"
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ./db/migrations:/migrations:ro
    entrypoint:
      - /migrate
      - -path=/migrations
      - -database=postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
      - -verbose
    command: ["up"]
    networks:
      - asesamo-net

  seeder:
    image: postgres:16-alpine
    restart: "no"
    depends_on:
      migrate:
        condition: service_completed_successfully
    volumes:
      - ./db/seeds:/seeds:ro
    environment:
      POSTGRES_USER: ${DB_USER:-asesamo_admin}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
      POSTGRES_DB: ${DB_NAME:-asesamo_db}
    entrypoint:
      - sh
      - -c
      - PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" -f /seeds/000001_seed_dev_data.sql
    networks:
      - asesamo-net

  api:
    build:
      context: ./asesamo-api
      dockerfile: Dockerfile
    container_name: asesamo-api
    restart: unless-stopped
    ports:
      - "${API_PORT:-8080}:8080"
    depends_on:
      db:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully
      seeder:
        condition: service_completed_successfully
    environment:
      APP_ENV: ${APP_ENV:-development}
      APP_PORT: 8080
      DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET:-super_secret_jwt_key_min_32_characters_long_12345}
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:3000,http://127.0.0.1:3000,http://web:3000}
      R2_ACCOUNT_ID: ${R2_ACCOUNT_ID:-mock_account}
      R2_ACCESS_KEY_ID: ${R2_ACCESS_KEY_ID:-mock_key}
      R2_SECRET_ACCESS_KEY: ${R2_SECRET_ACCESS_KEY:-mock_secret}
      R2_BUCKET_NAME: ${R2_BUCKET_NAME:-asesamo-dev}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 10s
    networks:
      - asesamo-net

  web:
    build:
      context: ./asesamo-web
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}
    container_name: asesamo-web
    restart: unless-stopped
    ports:
      - "${WEB_PORT:-3000}:3000"
    depends_on:
      api:
        condition: service_healthy
    environment:
      NODE_ENV: production
      PORT: 3000
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}
      INTERNAL_API_URL: ${INTERNAL_API_URL:-http://api:8080/api/v1}
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 15s
    networks:
      - asesamo-net

volumes:
  asesamo_pgdata:
    name: asesamo_pgdata

networks:
  asesamo-net:
    name: asesamo-net
    driver: bridge
"""

temp_file = "/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/proposed_docker-compose.yml"
with open(temp_file, "w") as f:
    f.write(proposed_compose)

print("Validating with docker compose config...")
env = dict(os.environ)
env["PORT"] = "4000"
env["APP_PORT"] = "9090"
env["WEB_PORT"] = "4000"
env["API_PORT"] = "9090"

res = subprocess.run(
    ["docker", "compose", "-f", temp_file, "config", "--format", "json"],
    env=env,
    capture_output=True,
    text=True,
)

if res.returncode != 0:
    print("docker compose config failed:")
    print(res.stderr)
    sys.exit(1)

cfg = json.loads(res.stdout)

# Check web service
web = cfg["services"]["web"]
web_target = web["ports"][0]["target"]
web_published = web["ports"][0]["published"]
web_env_port = web["environment"].get("PORT")
web_health = " ".join(web["healthcheck"]["test"])

print(f"WEB: target={web_target}, published={web_published}, env_port={web_env_port}")
print(f"WEB healthcheck: {web_health}")
assert str(web_target) == "3000", f"Expected web target 3000, got {web_target}"
assert str(web_published) == "4000", f"Expected web published 4000, got {web_published}"
assert str(web_env_port) == "3000", f"Expected web env PORT 3000, got {web_env_port}"
assert ":3000" in web_health, "Healthcheck does not probe 3000"

# Check api service
api = cfg["services"]["api"]
api_target = api["ports"][0]["target"]
api_published = api["ports"][0]["published"]
api_env_port = api["environment"].get("APP_PORT")
api_health = " ".join(api["healthcheck"]["test"])

print(f"API: target={api_target}, published={api_published}, env_port={api_env_port}")
print(f"API healthcheck: {api_health}")
assert str(api_target) == "8080", f"Expected api target 8080, got {api_target}"
assert str(api_published) == "9090", f"Expected api published 9090, got {api_published}"
assert str(api_env_port) == "8080", f"Expected api env APP_PORT 8080, got {api_env_port}"
assert ":8080" in api_health, "Healthcheck does not probe 8080"

# Check seeder entrypoint quoting
seeder = cfg["services"]["seeder"]
seeder_entrypoint = seeder["entrypoint"]
print(f"Seeder entrypoint: {seeder_entrypoint}")
entrypoint_str = " ".join(seeder_entrypoint)
assert 'PGPASSWORD="$${POSTGRES_PASSWORD}"' in entrypoint_str or 'PGPASSWORD="${POSTGRES_PASSWORD}"' in entrypoint_str
assert '-U "$${POSTGRES_USER}"' in entrypoint_str or '-U "${POSTGRES_USER}"' in entrypoint_str
assert '-d "$${POSTGRES_DB}"' in entrypoint_str or '-d "${POSTGRES_DB}"' in entrypoint_str

print("ALL VALIDATION CHECKS PASSED!")
