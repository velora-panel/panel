import { api } from '../lib/api'
export async function requireSession() {
  let response: Response
  try {
    ;({ response } = await api.GET('/api/v1/auth/me'))
  } catch {
    return '/error'
  }
  if (response.status === 401) return '/login'
  if (!response.ok) return '/error'
  return true
}
