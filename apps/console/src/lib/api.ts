import { createClient, type paths } from '@velora/api-client'
export const api = createClient<paths>({
  credentials: 'same-origin',
  headers: { 'Content-Type': 'application/json' },
})
export function unwrap<T>(result: {
  data?: T
  error?: { error: { message: string } }
}): T {
  if (result.error) throw new Error(result.error.error.message)
  if (result.data === undefined)
    throw new Error('Core returned an empty response.')
  return result.data
}
