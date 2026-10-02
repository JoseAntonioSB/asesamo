import difflib

with open('/Users/jasbotello/devel/asesamo/Makefile', 'r') as f:
    orig = f.readlines()

new = list(orig)

# docker-up
for i, line in enumerate(new):
    if line.strip() == '@echo "==> Waiting for database readiness..."':
        new[i] = '\t@echo "==> Waiting for database readiness (max 30s)..."\n'
        new[i+1] = '\t@for i in $$(seq 1 30); do \\\n\t\tif docker compose exec -T db pg_isready -U "$(DB_USER)" -d "$(DB_NAME)" > /dev/null 2>&1; then \\\n\t\t\techo "✔ Docker services started and database is ready."; \\\n\t\t\texit 0; \\\n\t\tfi; \\\n\t\tsleep 1; \\\n\tdone; \\\n\techo "✖ Database failed to become ready within 30 seconds."; \\\n\texit 1\n'
        # remove until loop lines
        del new[i+2:i+5]
        break

# migrate-force
for i, line in enumerate(new):
    if line.strip() == 'migrate-force:':
        new[i+1] = '\t@if [ -z "$(V)" ] || ! echo "$(V)" | grep -qE \'^[0-9]+$$\'; then \\\n\t\techo "✖ Error: Please specify numeric version, e.g. make migrate-force V=1"; \\\n\t\texit 1; \\\n\tfi\n'
        new[i+3] = '\tdocker compose run --rm migrate force "$(V)"\n'
        break

# seed
for i, line in enumerate(new):
    if 'docker compose exec -T -e PGPASSWORD=$(DB_PASSWORD)' in line:
        new[i] = '\t\tdocker compose exec -T -e PGPASSWORD="$(DB_PASSWORD)" db psql -U "$(DB_USER)" -d "$(DB_NAME)" < db/seeds/000001_seed_dev_data.sql;\n'
        break

# db-shell
for i, line in enumerate(new):
    if line.strip() == 'db-shell:':
        new[i+1] = '\tdocker compose exec -e PGPASSWORD="$(DB_PASSWORD)" db psql -U "$(DB_USER)" -d "$(DB_NAME)"\n'
        break

patch = ''.join(difflib.unified_diff(orig, new, fromfile='a/Makefile', tofile='b/Makefile'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/advisory_makefile_remediation.patch', 'w') as f:
    f.write(patch)

print("Generated advisory_makefile_remediation.patch!")
