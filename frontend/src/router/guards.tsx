import { useEffect, useState } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { Spin } from 'antd'

import { useUserStore } from '@/store'
import { clearLoginInfo, getAccessToken } from '@/utils/storage'

export default function AuthGuard() {
  const location = useLocation()
  const { isLoggedIn, user, fetchUser, loading, setUser } = useUserStore()
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    const checkAuth = async () => {
      const token = getAccessToken()

      if (!token) {
        if (isLoggedIn || user) {
          clearLoginInfo()
          setUser(null)
        }
        setChecking(false)
        return
      }

      if (isLoggedIn && user) {
        setChecking(false)
        return
      }

      try {
        await fetchUser()
      } catch {
        clearLoginInfo()
        setUser(null)
      } finally {
        setChecking(false)
      }
    }

    void checkAuth()
  }, [fetchUser, isLoggedIn, setUser, user])

  if (checking || loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-gradient-main">
        <Spin size="large" tip="验证登录状态中...">
          <div />
        </Spin>
      </div>
    )
  }

  if (!getAccessToken() || !isLoggedIn || !user) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <Outlet />
}

interface RoleGuardProps {
  allowedRoles: string[]
  children: React.ReactNode
}

export function RoleGuard({ allowedRoles, children }: RoleGuardProps) {
  const { user } = useUserStore()

  if (!user) {
    return <Navigate to="/login" replace />
  }

  if (!allowedRoles.includes(user.role)) {
    return <Navigate to="/403" replace />
  }

  return <>{children}</>
}

export function usePermission() {
  const { user, hasRole, isAdmin, isSchoolAdmin, isTeacher, isStudent } = useUserStore()

  return {
    user,
    hasRole,
    isAdmin: isAdmin(),
    isSchoolAdmin: isSchoolAdmin(),
    isTeacher: isTeacher(),
    isStudent: isStudent(),
    canManageCourse: () => isTeacher() || isSchoolAdmin() || isAdmin(),
    canManageContest: () => isTeacher() || isSchoolAdmin() || isAdmin(),
    canManageUser: () => isSchoolAdmin() || isAdmin(),
    canAccessAdmin: () => isAdmin(),
    canAccessSchoolAdmin: () => isSchoolAdmin() || isAdmin(),
  }
}
