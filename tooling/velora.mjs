import { existsSync, mkdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'
const root = fileURLToPath(new URL('../', import.meta.url))
const envFile = `${root}.env`
if (existsSync(envFile)) process.loadEnvFile(envFile)
const directory = `${root}apps/cli`
mkdirSync(`${directory}/bin`, { recursive: true })
const executable = `${directory}/bin/velora`
const built = spawnSync('go', ['build', '-o', executable, './cmd/velora'], {
  cwd: directory,
  stdio: 'inherit',
  env: process.env,
})
if (built.error) {
  console.error(
    'Cannot execute Go. Install Go 1.27 or use the compiled velora binary.',
  )
  process.exit(1)
}
if (built.status !== 0) process.exit(built.status ?? 1)
// Linux exec preserves the terminal, stdin, signals and the native exit status.
process.execve(executable, [executable, ...process.argv.slice(2)], process.env)
