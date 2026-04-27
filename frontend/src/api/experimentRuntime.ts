import type { FileItem, LogEntry } from '@/types/presentation'

function buildRuntimeUrl(accessUrl: string, path: string, searchParams?: URLSearchParams): string {
  const normalizedBase = accessUrl.endsWith('/') ? accessUrl : `${accessUrl}/`
  const url = new URL(path.replace(/^\//, ''), normalizedBase)
  if (searchParams) {
    url.search = searchParams.toString()
  }
  return url.toString()
}

async function parseRuntimeResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new Error(`Runtime request failed: ${response.status}`)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}

export async function listRuntimeFiles(accessUrl: string, path: string): Promise<FileItem[]> {
  const searchParams = new URLSearchParams({ path })
  const response = await fetch(buildRuntimeUrl(accessUrl, 'api/files', searchParams))
  const data = await parseRuntimeResponse<{ files?: FileItem[] }>(response)
  return data.files || []
}

export async function createRuntimeDirectory(accessUrl: string, path: string): Promise<void> {
  const response = await fetch(buildRuntimeUrl(accessUrl, 'api/files/mkdir'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path }),
  })
  await parseRuntimeResponse<void>(response)
}

export async function deleteRuntimeFile(accessUrl: string, path: string): Promise<void> {
  const searchParams = new URLSearchParams({ path })
  const response = await fetch(buildRuntimeUrl(accessUrl, 'api/files', searchParams), {
    method: 'DELETE',
  })
  await parseRuntimeResponse<void>(response)
}

export async function listRuntimeLogs(
  accessUrl: string,
  options?: {
    source?: string
    levels?: string[]
  },
): Promise<LogEntry[]> {
  const searchParams = new URLSearchParams()
  if (options?.source) {
    searchParams.append('source', options.source)
  }
  if (options?.levels?.length) {
    searchParams.append('levels', options.levels.join(','))
  }

  const response = await fetch(buildRuntimeUrl(accessUrl, 'api/logs', searchParams))
  const data = await parseRuntimeResponse<{ logs?: LogEntry[] }>(response)
  return data.logs || []
}

export async function probeRuntimeEndpoint(url: string, signal?: AbortSignal): Promise<boolean> {
  try {
    const response = await fetch(url, {
      signal,
      credentials: 'include',
    })
    return response.ok
  } catch {
    return false
  }
}

export async function sendRuntimeRequest(accessUrl: string, options: {
  path?: string
  method: string
  headers?: Record<string, string>
  body?: string
}): Promise<{
  status: number
  statusText: string
  body: string
}> {
  const response = await fetch(buildRuntimeUrl(accessUrl, options.path || ''), {
    method: options.method,
    headers: options.headers,
    body: options.body,
  })

  return {
    status: response.status,
    statusText: response.statusText,
    body: await response.text(),
  }
}

export async function sendRuntimeJsonRpc<T>(
  accessUrl: string,
  method: string,
  params: unknown[] = [],
): Promise<T> {
  const response = await sendRuntimeRequest(accessUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      method,
      params,
      id: Date.now(),
    }),
  })

  const parsed = JSON.parse(response.body) as {
    result?: T
    error?: {
      code?: number
      message?: string
    }
  }

  if (parsed.error) {
    throw new Error(parsed.error.message || `JSON-RPC request failed: ${parsed.error.code || 'UNKNOWN'}`)
  }

  return parsed.result as T
}
