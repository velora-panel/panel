import { readFileSync } from 'node:fs'
const lines = readFileSync(process.argv[2], 'utf8').trim().split('\n').slice(1)
const exemptions = [
  '/cmd/velora/',
  '/internal/app/',
  '/internal/database/generated/',
]
const scope = lines.filter(
  (line) => !exemptions.some((path) => line.includes(path)),
)
let total = 0
let covered = 0
for (const line of scope) {
  const [, statements, count] = line.split(' ')
  total += Number(statements)
  if (Number(count) > 0) covered += Number(statements)
}
console.log(
  `Go authored scope: ${covered}/${total} statements (${((100 * covered) / total).toFixed(2)}%)`,
)
if (!total || covered !== total) process.exit(1)
