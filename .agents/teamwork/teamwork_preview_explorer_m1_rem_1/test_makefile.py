import subprocess

makefile = '/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/proposed_Makefile'

# Test migrate-force
res = subprocess.run(['make', '-f', makefile, '-n', 'migrate-force', 'V=1; echo INJECTED'], capture_output=True, text=True)
print("migrate-force output:", res.stdout)
assert 'INJECTED' not in res.stdout or '"1; echo INJECTED"' in res.stdout, 'Unquoted injection detected'

# Test seed
res = subprocess.run(['make', '-f', makefile, '-n', 'seed', 'DB_PASSWORD=pass; echo INJECTED'], capture_output=True, text=True)
print("seed output:", res.stdout)
assert 'PGPASSWORD="pass; echo INJECTED"' in res.stdout, 'Password not properly quoted'

# Test db-shell
res = subprocess.run(['make', '-f', makefile, '-n', 'db-shell', 'DB_PASSWORD=pass; echo INJECTED'], capture_output=True, text=True)
print("db-shell output:", res.stdout)
assert 'PGPASSWORD="pass; echo INJECTED"' in res.stdout, 'db-shell password not properly quoted'

print("All Makefile security assertions PASSED!")
