#!/usr/bin/env python3
import json
import os
import re
import subprocess
import sys

REPO_ROOT = "/Users/jasbotello/devel/asesamo"
COMPOSE_FILE = os.path.join(REPO_ROOT, "docker-compose.yml")
ENV_FILE = os.path.join(REPO_ROOT, ".env")
ENV_EXAMPLE_FILE = os.path.join(REPO_ROOT, ".env.example")
MAKEFILE = os.path.join(REPO_ROOT, "Makefile")
COMMON_SH = os.path.join(REPO_ROOT, "test/e2e/common.sh")
SEED_SQL = os.path.join(REPO_ROOT, "db/seeds/000001_seed_dev_data.sql")
PROJECT_MD = os.path.join(REPO_ROOT, "PROJECT.md")

print("=================================================================")
print("  ASESAMO M1 REMEDIATION VERIFICATION SUITE")
print("=================================================================\n")

# 1. Docker Compose Config Quiet
print("--- Check 1: docker compose config --quiet ---")
res = subprocess.run(["docker", "compose", "config", "--quiet"], cwd=REPO_ROOT, capture_output=True, text=True)
assert res.returncode == 0, f"docker compose config failed: {res.stderr}"
print("✔ docker compose config --quiet passed with exit code 0\n")

# 2. Port Decoupling & Seeder Quotes
print("--- Check 2: Docker Compose Port Decoupling & Seeder Quotes ---")
env = dict(os.environ)
env["PORT"] = "4000"
env["WEB_PORT"] = "4000"
env["APP_PORT"] = "9090"
env["API_PORT"] = "9090"
res = subprocess.run(["docker", "compose", "--env-file", "/dev/null", "config", "--format", "json"], env=env, cwd=REPO_ROOT, capture_output=True, text=True)
assert res.returncode == 0, f"docker compose json failed: {res.stderr}"
cfg = json.loads(res.stdout)

# Web Service
web = cfg["services"]["web"]
target_port = web["ports"][0]["target"]
published_port = web["ports"][0]["published"]
env_port = web["environment"].get("PORT")
health_test = " ".join(web["healthcheck"]["test"])
print(f"  Web Published: {published_port}, Target: {target_port}, Container PORT: {env_port}")
print(f"  Web Healthcheck: {health_test}")
assert published_port == "4000", f"Web published port mismatch: {published_port}"
assert target_port == 3000, f"Web target port should be locked to 3000, got: {target_port}"
assert env_port == "3000", f"Web container PORT should be locked to 3000, got: {env_port}"
assert str(target_port) == str(env_port), "Web target port != env PORT"
assert f":{target_port}" in health_test, f"Web healthcheck does not probe target port {target_port}"
print("✔ Web service port decoupling verified")

# API Service
api = cfg["services"]["api"]
api_target_port = api["ports"][0]["target"]
api_published_port = api["ports"][0]["published"]
api_env_port = api["environment"].get("APP_PORT")
api_health_test = " ".join(api["healthcheck"]["test"])
print(f"  API Published: {api_published_port}, Target: {api_target_port}, Container APP_PORT: {api_env_port}")
print(f"  API Healthcheck: {api_health_test}")
assert api_published_port == "9090", f"API published port mismatch: {api_published_port}"
assert api_target_port == 8080, f"API target port should be locked to 8080, got: {api_target_port}"
assert api_env_port == "8080", f"API container APP_PORT should be locked to 8080, got: {api_env_port}"
assert str(api_target_port) == str(api_env_port), "API target port != env APP_PORT"
assert f":{api_target_port}" in api_health_test, f"API healthcheck does not probe target port {api_target_port}"
print("✔ API service port decoupling verified")

# Seeder Entrypoint Quoting
seeder = cfg["services"]["seeder"]
seeder_entrypoint = " ".join(seeder["entrypoint"])
print(f"  Seeder Entrypoint: {seeder_entrypoint}")
assert 'PGPASSWORD="$${POSTGRES_PASSWORD}"' in seeder_entrypoint, "Seeder PGPASSWORD not double-quoted"
assert '-U "$${POSTGRES_USER}"' in seeder_entrypoint, "Seeder POSTGRES_USER not double-quoted"
assert '-d "$${POSTGRES_DB}"' in seeder_entrypoint, "Seeder POSTGRES_DB not double-quoted"
print("✔ Seeder credentials double-quoted in entrypoint\n")

# 3. Environment Variables in .env and .env.example
print("--- Check 3: .env and .env.example Variable & Comment Verification ---")
for fpath in [ENV_FILE, ENV_EXAMPLE_FILE]:
    fname = os.path.basename(fpath)
    with open(fpath, "r") as f:
        content = f.read()
    assert "API_PORT=8080" in content, f"API_PORT=8080 missing in {fname}"
    assert "WEB_PORT=3000" in content, f"WEB_PORT=3000 missing in {fname}"
    assert "APP_PORT=8080" in content, f"APP_PORT=8080 missing in {fname}"
    assert "PORT=3000" in content, f"PORT=3000 missing in {fname}"
    # Verify distinguishing comments
    assert "Host publishing port" in content, f"Host publishing port comment missing in {fname}"
    assert "Container / application internal listening port" in content, f"Container listening port comment missing in {fname}"
    print(f"✔ {fname} verified with API_PORT, WEB_PORT, and distinguishing comments")
print()

# 4. Makefile Help and Dry-Run Across All Targets
print("--- Check 4: Makefile Targets Dry-Run ---")
res_help = subprocess.run(["make", "help"], cwd=REPO_ROOT, capture_output=True, text=True)
assert res_help.returncode == 0, f"make help failed: {res_help.stderr}"
print("✔ make help executed successfully")

targets = [
    "help", "docker-up", "docker-down", "docker-down-v",
    "migrate-up", "migrate-down", "migrate-version", "migrate-force",
    "seed", "db-shell", "test", "test-integration", "lint", "build", "run", "sqlc"
]
for t in targets:
    res = subprocess.run(["make", "-n", t], cwd=REPO_ROOT, capture_output=True, text=True)
    assert res.returncode == 0, f"make -n {t} failed: {res.stderr}"
print(f"✔ All {len(targets)} Makefile targets passed dry-run (exit code 0)\n")

# 5. Makefile Command Injection Hardening
print("--- Check 5: Makefile Injection Protection ---")
res_inj1 = subprocess.run(["make", "-n", "migrate-force", "V=1; echo INJECTED"], cwd=REPO_ROOT, capture_output=True, text=True)
assert "INJECTED" not in res_inj1.stdout, "DEFECT: Shell injection detected in make migrate-force"
print("✔ make migrate-force injection payload rejected in dry-run output")

res_inj2 = subprocess.run(["make", "-n", "seed", "DB_PASSWORD=pass; echo INJECTED"], cwd=REPO_ROOT, capture_output=True, text=True)
assert "INJECTED" not in res_inj2.stdout, "DEFECT: Shell injection detected in make seed"
print("✔ make seed injection payload rejected in dry-run output")

res_inj3 = subprocess.run(["make", "-n", "db-shell", "DB_PASSWORD=pass; echo INJECTED"], cwd=REPO_ROOT, capture_output=True, text=True)
assert "INJECTED" not in res_inj3.stdout, "DEFECT: Shell injection detected in make db-shell"
print("✔ make db-shell injection payload rejected in dry-run output")

# 6. Makefile migrate-force Input Validation
print("\n--- Check 6: Makefile migrate-force Input Validation ---")
bad_versions = ["", "abc", "1; id", '1" || id || "', "-1", "1.5", "$(id)", "`id`", "foo bar"]
for bad_v in bad_versions:
    res = subprocess.run(["make", "migrate-force", f"V={bad_v}"], cwd=REPO_ROOT, capture_output=True, text=True)
    assert res.returncode != 0, f"Bad version V='{bad_v}' was not rejected! Returncode: {res.returncode}"
print(f"✔ All {len(bad_versions)} malicious/invalid V values properly rejected with non-zero exit code")

res_valid = subprocess.run(["make", "-n", "migrate-force", "V=1"], cwd=REPO_ROOT, capture_output=True, text=True)
assert res_valid.returncode == 0, f"Valid V=1 failed dry-run: {res_valid.stderr}"
print("✔ Valid numeric V=1 dry-run succeeded")

# 7. Makefile docker-up Bounded Loop
print("\n--- Check 7: Makefile docker-up Bounded Loop ---")
res_up = subprocess.run(["make", "-n", "docker-up"], cwd=REPO_ROOT, capture_output=True, text=True)
assert "seq 1 30" in res_up.stdout, "docker-up does not contain 30s bounded loop"
assert "exit 1" in res_up.stdout, "docker-up does not exit with 1 on timeout"
assert "until" not in res_up.stdout, "docker-up still contains unbounded until loop"
print("✔ docker-up contains bounded 30s loop with timeout exit 1")

# 8. test/e2e/common.sh Access Point UUID
print("\n--- Check 8: test/e2e/common.sh UUID Synchronization ---")
with open(COMMON_SH, "r") as f:
    common_content = f.read()

assert 'TEST_ACCESS_POINT_ID="${TEST_ACCESS_POINT_ID:-33333333-3333-3333-3333-333333333331}"' in common_content, \
    "TEST_ACCESS_POINT_ID default does not match 33333333-3333-3333-3333-333333333331 in test/e2e/common.sh"
print("✔ test/e2e/common.sh TEST_ACCESS_POINT_ID set to 33333333-3333-3333-3333-333333333331")

with open(SEED_SQL, "r") as f:
    seed_content = f.read()
assert "33333333-3333-3333-3333-333333333331" in seed_content, "Seed script missing access point UUID"
print("✔ db/seeds/000001_seed_dev_data.sql access point UUID confirmed matching")

with open(PROJECT_MD, "r") as f:
    project_content = f.read()
assert "33333333-3333-3333-3333-333333333331" in project_content, "PROJECT.md missing access point UUID"
print("✔ PROJECT.md access point UUID confirmed matching")

print("\n=================================================================")
print("  ALL 8 VERIFICATION CHECKS PASSED WITH 100% SUCCESS!")
print("=================================================================")
