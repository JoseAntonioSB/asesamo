import difflib

# Patch for .env
with open('/Users/jasbotello/devel/asesamo/.env', 'r') as f:
    env_orig = f.readlines()

env_new = []
for line in env_orig:
    if line.strip() == 'APP_PORT=8080':
        env_new.append('API_PORT=8080\n')
        env_new.append(line)
    elif line.strip() == 'PORT=3000':
        env_new.append('WEB_PORT=3000\n')
        env_new.append(line)
    else:
        env_new.append(line)

env_patch = ''.join(difflib.unified_diff(env_orig, env_new, fromfile='a/.env', tofile='b/.env'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env.patch', 'w') as f:
    f.write(env_patch)

# Patch for .env.example
with open('/Users/jasbotello/devel/asesamo/.env.example', 'r') as f:
    example_orig = f.readlines()

example_new = []
for line in example_orig:
    if line.strip() == 'APP_PORT=8080':
        example_new.append('API_PORT=8080\n')
        example_new.append(line)
    elif line.strip() == 'PORT=3000':
        example_new.append('WEB_PORT=3000\n')
        example_new.append(line)
    else:
        example_new.append(line)

example_patch = ''.join(difflib.unified_diff(example_orig, example_new, fromfile='a/.env.example', tofile='b/.env.example'))
with open('/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env_example.patch', 'w') as f:
    f.write(example_patch)

print('Updated env and env.example patches!')
