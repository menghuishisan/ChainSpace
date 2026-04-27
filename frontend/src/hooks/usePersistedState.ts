/**
 * 页面状态持久化 Hooks
 * 用于在页面刷新后保持状态（如多阶段页面的当前阶段、表单数据等）
 */
import { useState, useEffect, useCallback, useRef } from 'react'

type StorageType = 'session' | 'local'

interface UsePersistedStateOptions<T> {
  /** 存储类型：session（会话期间）或 local（永久） */
  storage?: StorageType
  /** 过期时间（毫秒） */
  expireMs?: number
  /** 自定义序列化函数 */
  serialize?: (value: T) => string
  /** 自定义反序列化函数 */
  deserialize?: (value: string) => T
  /** 验证函数，用于检查存储值是否有效 */
  validate?: (value: T) => boolean
}

interface StoredValue<T> {
  value: T
  timestamp: number
  version: number
}

const STORAGE_VERSION = 1

/**
 * 获取存储对象
 */
function getStorage(type: StorageType): Storage {
  return type === 'local' ? localStorage : sessionStorage
}

/**
 * 持久化状态 Hook
 * @param key 存储键名（建议使用页面路径作为前缀）
 * @param defaultValue 默认值
 * @param options 配置选项
 * @returns [state, setState, clearState]
 */
export function usePersistedState<T>(
  key: string,
  defaultValue: T,
  options: UsePersistedStateOptions<T> = {}
): [T, (value: T | ((prev: T) => T)) => void, () => void] {
  const { 
    storage = 'session', 
    expireMs,
    serialize = JSON.stringify,
    deserialize = JSON.parse,
    validate
  } = options
  
  const storageKey = `chainspace_${key}`
  const defaultValueRef = useRef(defaultValue)

  // 从存储中读取初始值
  const getInitialValue = useCallback((): T => {
    try {
      const store = getStorage(storage)
      const item = store.getItem(storageKey)
      if (!item) return defaultValueRef.current

      const stored: StoredValue<T> = deserialize(item)
      
      // 版本检查
      if (stored.version !== STORAGE_VERSION) {
        store.removeItem(storageKey)
        return defaultValueRef.current
      }
      
      // 过期检查
      if (expireMs && Date.now() - stored.timestamp > expireMs) {
        store.removeItem(storageKey)
        return defaultValueRef.current
      }
      
      // 有效性验证
      if (validate && !validate(stored.value)) {
        store.removeItem(storageKey)
        return defaultValueRef.current
      }
      
      return stored.value
    } catch {
      return defaultValueRef.current
    }
  }, [storageKey, storage, expireMs, deserialize, validate])

  const [state, setState] = useState<T>(getInitialValue)

  // 同步存储变化（跨标签页同步，仅localStorage）
  useEffect(() => {
    if (storage !== 'local') return

    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === storageKey && e.newValue) {
        try {
          const stored: StoredValue<T> = deserialize(e.newValue)
          if (stored.version === STORAGE_VERSION) {
            setState(stored.value)
          }
        } catch {
          // 忽略解析错误
        }
      }
    }

    window.addEventListener('storage', handleStorageChange)
    return () => window.removeEventListener('storage', handleStorageChange)
  }, [storageKey, storage, deserialize])

  // 更新状态并持久化
  const setPersistedState = useCallback((value: T | ((prev: T) => T)) => {
    setState((prev) => {
      const newValue = typeof value === 'function' ? (value as (prev: T) => T)(prev) : value
      
      try {
        const store = getStorage(storage)
        const stored: StoredValue<T> = {
          value: newValue,
          timestamp: Date.now(),
          version: STORAGE_VERSION
        }
        store.setItem(storageKey, serialize(stored))
      } catch (error) {
        console.error('Failed to persist state:', error)
      }
      
      return newValue
    })
  }, [storageKey, storage, serialize])

  // 清除持久化状态
  const clearPersistedState = useCallback(() => {
    try {
      const store = getStorage(storage)
      store.removeItem(storageKey)
      setState(defaultValueRef.current)
    } catch (error) {
      console.error('Failed to clear persisted state:', error)
    }
  }, [storageKey, storage])

  return [state, setPersistedState, clearPersistedState]
}

/**
 * 多阶段页面步骤持久化 Hook
 * @param pageId 页面唯一标识
 * @param totalSteps 总步骤数
 * @param options 额外配置
 * @returns [currentStep, setStep, resetStep, goNext, goPrev]
 */
export function usePersistedStep(
  pageId: string,
  totalSteps: number,
  options: { storage?: StorageType } = {}
): {
  currentStep: number
  setStep: (step: number) => void
  resetStep: () => void
  goNext: () => boolean
  goPrev: () => boolean
  isFirst: boolean
  isLast: boolean
} {
  const { storage = 'session' } = options
  
  const [step, setStepRaw, clearStep] = usePersistedState<number>(
    `step_${pageId}`,
    0,
    { 
      storage,
      validate: (v) => typeof v === 'number' && v >= 0 && v < totalSteps
    }
  )

  // 确保步骤在有效范围内
  const currentStep = Math.min(Math.max(0, step), totalSteps - 1)

  const setStep = useCallback((newStep: number) => {
    const validStep = Math.min(Math.max(0, newStep), totalSteps - 1)
    setStepRaw(validStep)
  }, [setStepRaw, totalSteps])

  const goNext = useCallback(() => {
    if (currentStep < totalSteps - 1) {
      setStep(currentStep + 1)
      return true
    }
    return false
  }, [currentStep, totalSteps, setStep])

  const goPrev = useCallback(() => {
    if (currentStep > 0) {
      setStep(currentStep - 1)
      return true
    }
    return false
  }, [currentStep, setStep])

  return {
    currentStep,
    setStep,
    resetStep: clearStep,
    goNext,
    goPrev,
    isFirst: currentStep === 0,
    isLast: currentStep === totalSteps - 1
  }
}

/**
 * Tab 页签持久化 Hook
 * @param pageId 页面唯一标识
 * @param defaultTab 默认选中的 Tab
 * @param validTabs 有效的 Tab 值列表
 * @returns [activeTab, setActiveTab, resetTab]
 */
export function usePersistedTab(
  pageId: string,
  defaultTab: string,
  validTabs: string[]
): [string, (tab: string) => void, () => void] {
  const [tab, setTab, clearTab] = usePersistedState<string>(
    `tab_${pageId}`,
    defaultTab,
    {
      storage: 'session',
      validate: (v) => validTabs.includes(v)
    }
  )

  return [tab, setTab, clearTab]
}

export default usePersistedState
