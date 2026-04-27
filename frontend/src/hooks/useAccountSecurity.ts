import { useCallback } from 'react'

import { changePassword } from '@/api/auth'

export function useAccountSecurity() {
  const updatePassword = useCallback(async (oldPassword: string, newPassword: string) => {
    await changePassword(oldPassword, newPassword)
  }, [])

  return {
    updatePassword,
  }
}
