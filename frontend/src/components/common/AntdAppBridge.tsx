import { useEffect } from 'react'
import { App, Modal, message, notification } from 'antd'

const noop = () => {}

export default function AntdAppBridge() {
  const app = App.useApp()

  useEffect(() => {
    Object.assign(message, app.message)
    Object.assign(notification, app.notification)
    Object.assign(Modal, {
      info: app.modal.info,
      success: app.modal.success,
      error: app.modal.error,
      warning: app.modal.warning,
      confirm: app.modal.confirm,
    })

    return () => {
      Object.assign(message, {
        open: noop,
        success: noop,
        error: noop,
        warning: noop,
        info: noop,
        loading: noop,
        destroy: noop,
      })
      Object.assign(notification, {
        open: noop,
        success: noop,
        error: noop,
        warning: noop,
        info: noop,
        destroy: noop,
      })
      Object.assign(Modal, {
        info: noop,
        success: noop,
        error: noop,
        warning: noop,
        confirm: noop,
      })
    }
  }, [app.message, app.modal, app.notification])

  return null
}
