import { readdirSync, readFileSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
function snapshot() {
  const dir = 'apps/core/internal/database/generated'
  const files = [
    'packages/api-client/src/schema.d.ts',
    ...readdirSync(dir)
      .sort()
      .map((file) => `${dir}/${file}`),
  ]
  return createHash('sha256')
    .update(
      files.map((file) => `${file}\n${readFileSync(file, 'utf8')}`).join('\n'),
    )
    .digest('hex')
}
const before = snapshot()
execFileSync('pnpm', ['codegen'], { stdio: 'inherit' })
if (before !== snapshot()) {
  console.error(
    'Generated artifacts were stale. Review and commit regeneration.',
  )
  process.exit(1)
}
console.log('Generated artifacts are current.')
