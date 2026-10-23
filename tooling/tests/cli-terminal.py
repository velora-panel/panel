"""Linux PTY regression test; only use a disposable *_test PostgreSQL database."""
import json
import os
import pty
import select
import signal
import subprocess
import time
import uuid
from urllib.parse import urlparse, parse_qsl, urlencode, urlunparse, unquote

url = os.environ.get('TEST_DATABASE_URL', '')
assert urlparse(url).path.endswith('_test'), 'Dedicated TEST_DATABASE_URL required'
parsed = urlparse(url)
schema = 'cli_' + uuid.uuid4().hex
pg_env = dict(os.environ, PGHOST=parsed.hostname or '', PGPORT=str(parsed.port or 5432), PGUSER=unquote(parsed.username or ''), PGPASSWORD=unquote(parsed.password or ''), PGDATABASE=parsed.path[1:])
def sql(statement):
    subprocess.run(['psql', '-X', '-v', 'ON_ERROR_STOP=1', '-qtAc', statement], env=pg_env, check=True)
sql('CREATE SCHEMA ' + schema)
import atexit
atexit.register(lambda: sql('DROP SCHEMA ' + schema + ' CASCADE'))
query = dict(parse_qsl(parsed.query)); query['search_path'] = schema
isolated = urlunparse(parsed._replace(query=urlencode(query)))
env = dict(os.environ, DATABASE_URL=isolated, MIGRATION_DATABASE_URL=isolated)
result = subprocess.run(['pnpm', 'velora', 'admin', 'create'], env=env, input='', capture_output=True, text=True)
assert result.returncode != 0 and 'schema is not initialized' in result.stderr, result.stderr.replace(isolated, '[test database]')

subprocess.run(['pnpm', 'velora', 'migrate', 'up'], env=env, check=True)
subprocess.run(['apps/cli/bin/velora', 'migrate', 'up'], env=env, check=True)
result = subprocess.run(['apps/cli/bin/velora', 'migrate', 'version'], env=env, capture_output=True, text=True, check=True)
assert result.stdout.strip() == '3'
result = subprocess.run(['apps/cli/bin/velora', 'migrate', 'status'], env=env, capture_output=True, text=True, check=True)
assert len(result.stdout.splitlines()) == 3 and 'pending' not in result.stdout
email = f'terminal-{uuid.uuid4()}@example.com'
password = f'PTY-{uuid.uuid4()}'
pid, fd = pty.fork()
if pid == 0:
    os.execvpe('pnpm', ['pnpm', 'velora', 'admin', 'create'], env)
transcript = bytearray()
def until(marker):
    deadline = time.monotonic() + 30
    while marker not in transcript:
        assert time.monotonic() < deadline, 'CLI prompt timed out'
        if select.select([fd], [], [], 1)[0]:
            transcript.extend(os.read(fd, 4096))
try:
    until(b'Email: ')
    os.write(fd, (email + '\n').encode())
    until(b'Password: ')
    os.write(fd, (password + '\n').encode())
    until(b'Confirm password: ')
    os.write(fd, (password + '\n').encode())
    until(b'Account updated successfully.')
    _, status = os.waitpid(pid, 0)
    assert os.waitstatus_to_exitcode(status) == 0
    assert password.encode() not in transcript, 'Password echoed by terminal'
finally:
    os.close(fd)
    try:
        os.kill(pid, signal.SIGTERM)
    except ProcessLookupError:
        pass
result = subprocess.run(['apps/cli/bin/velora', 'admin', 'list'], env=env, capture_output=True, text=True, check=True)
assert email in result.stdout, 'Native CLI did not find the terminal-created administrator'
result = subprocess.run(['pnpm', 'velora', 'admin', 'create'], env=env, input='', capture_output=True, text=True)
assert result.returncode != 0 and 'interactive terminal required' in result.stderr
result = subprocess.run(['pnpm', 'velora', 'unknown'], env=env, capture_output=True, text=True)
assert result.returncode != 0 and 'unknown command' in result.stderr
# Exercise privileged reset and last-admin protection without secret command arguments.
reset = json.dumps(dict(email=email, password=password, password_confirmation=password))
result = subprocess.run(['apps/cli/bin/velora', 'admin', 'reset-password', '--password-stdin'], env=env, input=reset, capture_output=True, text=True, check=True)
assert password not in result.stdout + result.stderr
result = subprocess.run(['apps/cli/bin/velora', 'user', 'disable', '--email', email], env=env, capture_output=True, text=True)
assert result.returncode != 0 and 'last active administrator' in result.stderr.lower()
member = 'member-' + uuid.uuid4().hex + '@example.com'
value = json.dumps(dict(email=member, password=password, password_confirmation=password))
subprocess.run(['apps/cli/bin/velora', 'user', 'create', '--password-stdin'], env=env, input=value, capture_output=True, text=True, check=True)
for action in ['disable', 'enable']:
    subprocess.run(['apps/cli/bin/velora', 'user', action, '--email', member], env=env, capture_output=True, text=True, check=True)
    result = subprocess.run(['apps/cli/bin/velora', 'user', 'list'], env=env, capture_output=True, text=True, check=True)
    line = next(line for line in result.stdout.splitlines() if member in line)
    assert ('disabled=true' in line) == (action == 'disable')
result = subprocess.run(['apps/cli/bin/velora', 'admin', 'list'], env=env, capture_output=True, text=True, check=True)
assert member not in result.stdout
print('PASS: fresh schema, migration status/version/idempotence, PTY no-echo prompts, native persistence/reset/status changes, last-admin guard, noninteractive rejection and exit codes')
