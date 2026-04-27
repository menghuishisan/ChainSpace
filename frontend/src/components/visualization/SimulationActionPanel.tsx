import { Button, Input, InputNumber, Select } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import type { SimulationActionPanelProps } from '@/types/visualizationDomain'

/**
 * 动作面板改为“先选动作，再执行”的紧凑模式，
 * 避免一长列表单占满可视化舞台。
 */
export default function SimulationActionPanel({ actions, onExecute }: SimulationActionPanelProps) {
  const [pendingActionKey, setPendingActionKey] = useState<string>('')
  const [selectedActionKey, setSelectedActionKey] = useState<string>('')
  const [formValues, setFormValues] = useState<Record<string, Record<string, unknown>>>({})

  const groupedActions = useMemo(() => {
    const groups = new Map<string, SimulationActionPanelProps['actions']>()

    for (const action of actions) {
      const groupName = action.group || '实验动作'
      const current = groups.get(groupName) || []
      current.push(action)
      groups.set(groupName, current)
    }

    return Array.from(groups.entries())
  }, [actions])

  useEffect(() => {
    if (actions.length === 0) {
      setSelectedActionKey('')
      return
    }

    if (selectedActionKey && actions.some((action) => action.key === selectedActionKey)) {
      return
    }

    setSelectedActionKey(actions[0].key)
  }, [actions, selectedActionKey])

  const actionOptions = useMemo(() => groupedActions.map(([groupName, groupActions]) => ({
    label: groupName,
    options: groupActions.map((action) => ({
      label: action.label,
      value: action.key,
    })),
  })), [groupedActions])

  const selectedAction = useMemo(
    () => actions.find((action) => action.key === selectedActionKey) || actions[0],
    [actions, selectedActionKey],
  )

  const buildActionValues = (
    actionKey: string,
    fields?: SimulationActionPanelProps['actions'][number]['fields'],
  ) => {
    const currentValues = formValues[actionKey] || {}
    if (!fields || fields.length === 0) {
      return currentValues
    }

    return fields.reduce<Record<string, unknown>>((result, field) => {
      if (currentValues[field.key] !== undefined) {
        result[field.key] = currentValues[field.key]
        return result
      }

      if (field.defaultValue !== undefined) {
        result[field.key] = field.defaultValue
      }

      return result
    }, { ...currentValues })
  }

  if (actions.length === 0 || !selectedAction) {
    return null
  }

  const values = buildActionValues(selectedAction.key, selectedAction.fields)

  return (
    <div className="rounded-2xl border border-slate-200 bg-[linear-gradient(180deg,#ffffff_0%,#f8fafc_100%)] p-3 shadow-sm">
      <div className="mb-2 text-sm font-medium text-slate-900">交互动作</div>
      <div className="mb-3">
        <div className="mb-1 text-xs font-medium text-slate-600">选择动作</div>
        <Select
          className="w-full"
          value={selectedAction.key}
          options={actionOptions}
          onChange={(value) => setSelectedActionKey(String(value))}
        />
      </div>

      <div className="rounded-2xl border border-slate-200 bg-[linear-gradient(180deg,#f8fbff_0%,#f1f5f9_100%)] p-3">
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="text-sm font-medium text-slate-900">{selectedAction.label}</div>
            <div className="mt-1 text-xs leading-5 text-slate-600">{selectedAction.description}</div>
          </div>
          {selectedAction.overlayLabel ? (
            <span className="rounded-full border border-amber-200 bg-amber-50 px-2 py-1 text-[11px] text-amber-800">
              {selectedAction.overlayLabel}
            </span>
          ) : null}
        </div>

        {selectedAction.fields && selectedAction.fields.length > 0 ? (
          <div className="mt-3 space-y-2">
            {selectedAction.fields.map((field) => {
              const currentValue = values[field.key] ?? field.defaultValue

              if (field.type === 'select') {
                return (
                  <div key={field.key}>
                    <div className="mb-1 text-xs font-medium text-slate-700">{field.label}</div>
                    <Select
                      className="w-full"
                      value={String(currentValue ?? '')}
                      options={(field.options || []).map((option) => ({
                        label: option.label,
                        value: option.value,
                      }))}
                      onChange={(value) => {
                        setFormValues((previous) => ({
                          ...previous,
                          [selectedAction.key]: {
                            ...previous[selectedAction.key],
                            [field.key]: value,
                          },
                        }))
                      }}
                    />
                  </div>
                )
              }

              if (field.type === 'number') {
                return (
                  <div key={field.key}>
                    <div className="mb-1 text-xs font-medium text-slate-700">{field.label}</div>
                    <InputNumber
                      className="w-full"
                      min={field.min}
                      max={field.max}
                      value={Number(currentValue ?? 0)}
                      onChange={(value) => {
                        setFormValues((previous) => ({
                          ...previous,
                          [selectedAction.key]: {
                            ...previous[selectedAction.key],
                            [field.key]: value ?? field.defaultValue ?? 0,
                          },
                        }))
                      }}
                    />
                  </div>
                )
              }

              return (
                <div key={field.key}>
                  <div className="mb-1 text-xs font-medium text-slate-700">{field.label}</div>
                  <Input
                    value={String(currentValue ?? '')}
                    onChange={(event) => {
                      setFormValues((previous) => ({
                        ...previous,
                        [selectedAction.key]: {
                          ...previous[selectedAction.key],
                          [field.key]: event.target.value,
                        },
                      }))
                    }}
                  />
                </div>
              )
            })}
          </div>
        ) : null}

        <Button
          className="mt-3"
          type="primary"
          loading={pendingActionKey === selectedAction.key}
          onClick={async () => {
            setPendingActionKey(selectedAction.key)
            try {
              await onExecute(selectedAction, values)
            } finally {
              setPendingActionKey('')
            }
          }}
        >
          执行当前动作
        </Button>
      </div>
    </div>
  )
}
