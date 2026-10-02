import difflib

with open('/Users/jasbotello/devel/asesamo/Makefile', 'r') as f:
    lines = f.readlines()

new_lines = []
i = 0
while i < len(lines):
    line = lines[i]
    if line.startswith('docker-up:'):
        new_lines.append(line)
        new_lines.append('\t@echo "==> Starting ASESAMO Docker services..."\n')
        new_lines.append('\tdocker compose up -d\n')
        new_lines.append('\t@echo "==> Waiting for database readiness (max 30s)..."\n')
        new_lines.append('\t@for i in $$(seq 1 30); do \\\n')
        new_lines.append('\t\tif docker compose exec -T db pg_isready -U "$(DB_USER)" -d "$(DB_NAME)" > /dev/null 2>&1; then \\\n')
        new_lines.append('\t\t\techo "✔ Docker services started and database is ready."; \\\n')
        new_lines.append('\t\t\texit 0; \\\n')
        new_lines.append('\t\tfi; \\\n')
        new_lines.append('\t\tsleep 1; \\\n')
        new_lines.append('\tdone; \\\n')
        new_lines.append('\techo "✖ Database failed to become ready within 30 seconds."; \\\n')
        new_lines.append('\texit 1\n')
        # skip old docker-up recipe (until '## Stop docker-compose stack')
        i += 1
        while i < len(lines) and not lines[i].startswith('## Stop docker-compose'):
            i += 1
        continue
    elif line.startswith('migrate-force:'):
        new_lines.append(line)
        new_lines.append('\t@if [ -z "$(V)" ] || ! echo "$(V)" | grep -qE \'^[0-9]+$$\'; then \\\n')
        new_lines.append('\t\techo "✖ Error: Please specify numeric version, e.g. make migrate-force V=1"; \\\n')
        new_lines.append('\t\texit 1; \\\n')
        new_lines.append('\tfi\n')
        new_lines.append('\t@echo "==> Forcing migration version to $(V)..."\n')
        new_lines.append('\tdocker compose run --rm migrate force "$(V)"\n')
        new_lines.append('\t@echo "✔ Migration version forced to $(V)."\n')
        i += 1
        while i < len(lines) and not lines[i].startswith('## Seed database'):
            i += 1
        continue
    elif 'PGPASSWORD=$(DB_PASSWORD)' in line:
        new_lines.append(line.replace('PGPASSWORD=$(DB_PASSWORD)', 'PGPASSWORD="$(DB_PASSWORD)"').replace('-U $(DB_USER) -d $(DB_NAME)', '-U "$(DB_USER)" -d "$(DB_NAME)"'))
        i += 1
        continue
    else:
        new_lines.append(line)
        i += 1

with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/proposed_Makefile', 'w') as f:
    f.writelines(new_lines)

patch = ''.join(difflib.unified_diff(lines, new_lines, fromfile='a/Makefile', tofile='b/Makefile'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/advisory_makefile_remediation.patch', 'w') as f:
    f.write(patch)

print('Generated proposed_Makefile and patch successfully!')
