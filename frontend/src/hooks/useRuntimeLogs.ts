import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { listRuntimeLogs } from '@/api/experimentRuntime'
import type { LogEntry } from '@/types/presentation'

export function useRuntimeLogs(accessUrl?: string) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [connected, setConnected] = useState(false)
  const [selectedSource, setSelectedSource] = useState('')
  const [levelFilter, setLevelFilter] = useState<string[]>([])
  const [searchText, setSearchText] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const containerRef = useRef<HTMLDivElement>(null)

  const refreshLogs = useCallback(async () => {
    if (!accessUrl) {
      setConnected(false)
      setLogs([])
      return
    }

    setLoading(true)
    try {
      const nextLogs = await listRuntimeLogs(accessUrl, {
        source: selectedSource || undefined,
        levels: levelFilter,
      })
      setLogs(nextLogs)
      setConnected(true)
    } catch {
      setConnected(false)
    } finally {
      setLoading(false)
    }
  }, [accessUrl, levelFilter, selectedSource])

  useEffect(() => {
    void refreshLogs()
  }, [refreshLogs])

  useEffect(() => {
    if (!accessUrl) {
      return
    }

    const timer = window.setInterval(() => {
      void refreshLogs()
    }, 5000)

    return () => {
      window.clearInterval(timer)
    }
  }, [accessUrl, refreshLogs])

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [autoScroll, logs])

  const filteredLogs = useMemo(() => logs.filter((log) => {
    if (selectedSource && log.source !== selectedSource) {
      return false
    }
    if (levelFilter.length > 0 && !levelFilter.includes(log.level)) {
      return false
    }
    if (searchText && !log.message.toLowerCase().includes(searchText.toLowerCase())) {
      return false
    }
    return true
  }), [levelFilter, logs, searchText, selectedSource])

  return {
    logs,
    loading,
    connected,
    selectedSource,
    levelFilter,
    searchText,
    autoScroll,
    containerRef,
    filteredLogs,
    setLogs,
    setSelectedSource,
    setLevelFilter,
    setSearchText,
    setAutoScroll,
    refreshLogs,
  }
}
