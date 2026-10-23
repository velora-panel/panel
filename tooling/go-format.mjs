import { execFileSync } from 'node:child_process'
import { existsSync } from 'node:fs'
const directories = ['cmd', 'internal', 'application', 'migrations'].filter(existsSync)
const tool = existsSync('sqlc.yaml') ? ['tool'] : ['tool', '-modfile=../core/go.mod']
const files = execFileSync('go', [...tool, 'goimports', '-l', ...directories], { encoding: 'utf8', env: { ...process.env, GOWORK: 'off' } })
if (files.trim()) { console.error(files); process.exit(1) }
