import difflib

with open('/Users/jasbotello/devel/asesamo/docker-compose.yml', 'r') as f:
    orig = f.readlines()

new = list(orig)
for i, line in enumerate(new):
    if 'PGPASSWORD=' in line:
        new[i] = '      - PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" -f /seeds/000001_seed_dev_data.sql\n'
    elif 'ports:' in line and '${APP_PORT:-8080}:8080' in new[i+1]:
        new[i+1] = '      - "${API_PORT:-8080}:8080"\n'
    elif 'APP_PORT: ${APP_PORT:-8080}' in line:
        new[i] = '      APP_PORT: 8080\n'
    elif 'ports:' in line and '${PORT:-3000}:3000' in new[i+1]:
        new[i+1] = '      - "${WEB_PORT:-3000}:3000"\n'
    elif 'PORT: ${PORT:-3000}' in line:
        new[i] = '      PORT: 3000\n'

patch = ''.join(difflib.unified_diff(orig, new, fromfile='a/docker-compose.yml', tofile='b/docker-compose.yml'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_docker_compose.patch', 'w') as f:
    f.write(patch)

# Also generate patch for .env
with open('/Users/jasbotello/devel/asesamo/.env', 'r') as f:
    env_orig = f.readlines()

env_new = list(env_orig)
for i, line in enumerate(env_new):
    if line.strip() == 'APP_PORT=8080':
        env_new[i] = 'API_PORT=8080\nAPP_PORT=8080\n'
    elif line.strip() == 'PORT=3000':
        env_new[i] = 'WEB_PORT=3000\nPORT=3000\n'

env_patch = ''.join(difflib.unified_diff(env_orig, env_new, fromfile='a/.env', tofile='b/.env'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env.patch', 'w') as f:
    f.write(env_patch)

# Also generate patch for .env.example
with open('/Users/jasbotello/devel/asesamo/.env.example', 'r') as f:
    example_orig = f.readlines()

example_new = list(example_orig)
for i, line in enumerate(example_new):
    if line.strip() == 'APP_PORT=8080':
        example_new[i] = 'API_PORT=8080\nAPP_PORT=8080\n'
    elif line.strip() == 'PORT=3000':
        example_new[i] = 'WEB_PORT=3000\nPORT=3000\n'

example_patch = ''.join(difflib.unified_diff(example_orig, example_new, fromfile='a/.env.example', tofile='b/.env.example'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env_example.patch', 'w') as f:
    f.write(example_patch)

print('Generated all 3 patches cleanly!')
