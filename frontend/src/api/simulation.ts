import type {
  AttackConfig,
  FaultConfig,
  SimulationRequestOptions,
  SimulationResponse,
  SimulatorActionResponse,
  SimulatorDescription,
  SimulatorEvent,
  SimulatorEventsPayload,
  SimulatorParam,
  SimulatorParamsPayload,
  SimulatorState,
  SimulatorStateExportPayload,
  SnapshotInfo,
  WSCommand,
  WSMessage,
} from '@/types/visualizationDomain'

/**
 * SimulationClient 负责和实验环境中的 simulations 服务通信。
 * 这里统一校验业务返回码，并在 WebSocket 不可用时优雅降级到 HTTP 轮询。
 */
export class SimulationClient {
  private readonly baseUrl: string
  private readonly wsUrl: string
  private ws: WebSocket | null = null
  private readonly eventHandlers = new Map<string, Array<(data: unknown) => void>>()
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = 5
  private readonly reconnectDelay = 1000
  private manualDisconnect = false
  private websocketUnavailable = false

  constructor(baseUrl?: string, websocketUrl?: string) {
    this.baseUrl = baseUrl || `${window.location.protocol}//${window.location.hostname}:8080`

    if (websocketUrl) {
      this.wsUrl = websocketUrl
      return
    }

    if (this.baseUrl.startsWith('/')) {
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      this.wsUrl = `${wsProtocol}//${window.location.host}${this.baseUrl}/ws/simulator`
      return
    }

    this.wsUrl = this.baseUrl.replace(/^http/i, 'ws') + '/ws/simulator'
  }

  async listSimulators(): Promise<SimulatorDescription[]> {
    return this.request<SimulatorDescription[]>('/api/simulators')
  }

  async getMeta(): Promise<SimulatorDescription | null> {
    return this.request<SimulatorDescription | null>('/api/meta')
  }

  async init(module: string, params?: Record<string, unknown>, nodeCount?: number): Promise<void> {
    await this.request('/api/simulator/init', {
      method: 'POST',
      body: { module, params, node_count: nodeCount },
    })
  }

  async start(): Promise<void> {
    await this.request('/api/simulator/start', { method: 'POST' })
  }

  async stop(): Promise<void> {
    await this.request('/api/simulator/stop', { method: 'POST' })
  }

  async pause(): Promise<void> {
    await this.request('/api/simulator/pause', { method: 'POST' })
  }

  async resume(): Promise<void> {
    await this.request('/api/simulator/resume', { method: 'POST' })
  }

  async step(): Promise<SimulatorState> {
    return this.request<SimulatorState>('/api/simulator/step', { method: 'POST' })
  }

  async reset(): Promise<void> {
    await this.request('/api/simulator/reset', { method: 'POST' })
  }

  async switch(module: string, preserveState = false): Promise<void> {
    await this.request('/api/simulator/switch', {
      method: 'POST',
      body: { module, preserve_state: preserveState },
    })
  }

  async getState(): Promise<SimulatorState> {
    return this.request<SimulatorState>('/api/simulator/state')
  }

  async getEvents(since = 0, limit = 100): Promise<SimulatorEvent[]> {
    const data = await this.request<SimulatorEventsPayload>(`/api/simulator/events?since=${since}&limit=${limit}`)
    return data?.events || []
  }

  async getParams(): Promise<SimulatorParam[]> {
    const data = await this.request<SimulatorParamsPayload>('/api/simulator/params')
    return Array.isArray(data) ? data : Object.values(data || {})
  }

  async setParam(key: string, value: unknown): Promise<void> {
    await this.request(`/api/simulator/params/${key}`, {
      method: 'PUT',
      body: { value },
    })
  }

  async setSpeed(speed: number): Promise<void> {
    await this.request('/api/simulator/speed', {
      method: 'PUT',
      body: { speed },
    })
  }

  async executeAction(action: string, params?: Record<string, unknown>): Promise<SimulatorActionResponse | null> {
    return this.request<SimulatorActionResponse | null>('/api/simulator/action', {
      method: 'POST',
      body: { action, params },
    })
  }

  async injectFault(fault: FaultConfig): Promise<string> {
    const data = await this.request<{ fault_id?: string }>('/api/simulator/fault', {
      method: 'POST',
      body: fault as unknown as Record<string, unknown>,
    })
    return data?.fault_id || ''
  }

  async removeFault(faultId: string): Promise<void> {
    await this.request(`/api/simulator/fault/${faultId}`, { method: 'DELETE' })
  }

  async clearFaults(): Promise<void> {
    await this.request('/api/simulator/faults', { method: 'DELETE' })
  }

  async injectAttack(attack: AttackConfig): Promise<string> {
    const data = await this.request<{ attack_id?: string }>('/api/simulator/attack', {
      method: 'POST',
      body: attack as unknown as Record<string, unknown>,
    })
    return data?.attack_id || ''
  }

  async removeAttack(attackId: string): Promise<void> {
    await this.request(`/api/simulator/attack/${attackId}`, { method: 'DELETE' })
  }

  async clearAttacks(): Promise<void> {
    await this.request('/api/simulator/attacks', { method: 'DELETE' })
  }

  async saveSnapshot(name: string): Promise<void> {
    await this.request('/api/simulator/snapshot', {
      method: 'POST',
      body: { name },
    })
  }

  async listSnapshots(): Promise<SnapshotInfo[]> {
    return this.request<SnapshotInfo[]>('/api/simulator/snapshots')
  }

  async loadSnapshot(name: string): Promise<void> {
    await this.request('/api/simulator/snapshot/load', {
      method: 'POST',
      body: { name },
    })
  }

  async deleteSnapshot(snapshotId: string): Promise<void> {
    await this.request(`/api/simulator/snapshot/${snapshotId}`, { method: 'DELETE' })
  }

  async exportState(): Promise<Record<string, unknown>> {
    const data = await this.request<SimulatorStateExportPayload>('/api/simulator/state/export')
    return data?.state || {}
  }

  async importState(state: Record<string, unknown>): Promise<void> {
    await this.request('/api/simulator/state/import', {
      method: 'POST',
      body: { state },
    })
  }

  connect(): Promise<void> {
    if (this.websocketUnavailable) {
      return Promise.reject(new Error('当前环境不支持 WebSocket，已切换为轮询模式。'))
    }

    return new Promise((resolve, reject) => {
      try {
        this.manualDisconnect = false
        this.ws = new WebSocket(this.wsUrl)
        let settled = false
        let opened = false

        this.ws.onopen = () => {
          opened = true
          settled = true
          this.reconnectAttempts = 0
          this.websocketUnavailable = false
          this.emit('connected', null)
          resolve()
        }

        this.ws.onclose = () => {
          this.emit('disconnected', null)

          if (!opened && !settled && !this.manualDisconnect) {
            this.websocketUnavailable = true
            settled = true
            reject(new Error('当前环境不支持 WebSocket，已自动切换为轮询模式。'))
            return
          }

          if (!this.manualDisconnect && !this.websocketUnavailable) {
            this.attemptReconnect()
          }
        }

        this.ws.onerror = () => {
          const wsError = new Error('WebSocket 连接不可用，已自动切换为轮询模式。')
          this.websocketUnavailable = true
          this.emit('error', wsError)
          if (!settled) {
            settled = true
            reject(wsError)
          }
        }

        this.ws.onmessage = (event) => {
          try {
            const messages = String(event.data).split('\n')
            for (const rawMessage of messages) {
              if (!rawMessage.trim()) {
                continue
              }
              const message = JSON.parse(rawMessage) as WSMessage
              this.handleMessage(message)
            }
          } catch (error) {
            console.error('Simulation WebSocket 消息解析失败:', error)
          }
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  disconnect(): void {
    if (!this.ws) {
      return
    }
    this.manualDisconnect = true
    this.ws.close()
    this.ws = null
  }

  send(action: string, params?: Record<string, unknown>): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket 尚未连接')
    }
    const command: WSCommand = { action, params }
    this.ws.send(JSON.stringify(command))
  }

  on(event: string, handler: (data: unknown) => void): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, [])
    }
    this.eventHandlers.get(event)?.push(handler)
  }

  off(event: string, handler: (data: unknown) => void): void {
    const handlers = this.eventHandlers.get(event)
    if (!handlers) {
      return
    }
    const index = handlers.indexOf(handler)
    if (index >= 0) {
      handlers.splice(index, 1)
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }

  private async request<T = void>(path: string, options: SimulationRequestOptions = {}): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: options.method || 'GET',
      headers: options.body ? { 'Content-Type': 'application/json' } : undefined,
      body: options.body ? JSON.stringify(options.body) : undefined,
      credentials: 'include',
    })

    let payload: SimulationResponse<T> | null = null
    try {
      payload = (await response.json()) as SimulationResponse<T>
    } catch {
      if (!response.ok) {
        throw new Error(`请求失败: ${response.status}`)
      }
      return undefined as T
    }

    if (!response.ok) {
      throw new Error(payload?.message || `请求失败: ${response.status}`)
    }

    if (payload && payload.code !== 0) {
      throw new Error(payload.message || '模拟器返回业务错误')
    }

    return payload?.data as T
  }

  private emit(event: string, data: unknown): void {
    this.eventHandlers.get(event)?.forEach((handler) => handler(data))
  }

  private handleMessage(message: WSMessage): void {
    switch (message.type) {
      case 'state_update':
        this.emit('state', message.data)
        break
      case 'event':
        this.emit('event', message.data)
        break
      case 'error':
        this.emit('error', message.data)
        break
      default:
        this.emit(message.type, message.data)
    }
  }

  private attemptReconnect(): void {
    if (this.websocketUnavailable || this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit('reconnect_failed', null)
      return
    }

    this.reconnectAttempts += 1
    window.setTimeout(() => {
      this.connect().catch(() => {
        // 重连失败时继续由递归逻辑接管，这里不重复抛错。
      })
    }, this.reconnectDelay * this.reconnectAttempts)
  }
}
