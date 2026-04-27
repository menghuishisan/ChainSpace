import type { Contest } from '@/types'

export interface ContestParticipationPresentation {
  canRegister: boolean
  canEnter: boolean
  canViewResult: boolean
  badgeText: string
  badgeColor: string
  teamSummaryText: string
}

export function getContestParticipationPresentation(
  contest: Contest,
  options?: {
    isManager?: boolean
    hasTeam?: boolean
  },
): ContestParticipationPresentation {
  const isManager = Boolean(options?.isManager)
  const hasTeam = Boolean(options?.hasTeam)
  const isRegistered = Boolean(contest.is_registered)

  if (isManager) {
    return {
      canRegister: false,
      canEnter: false,
      canViewResult: contest.status === 'ended',
      badgeText: '',
      badgeColor: 'default',
      teamSummaryText: hasTeam ? '已配置队伍信息' : '管理视角不展示报名身份',
    }
  }

  if (contest.status === 'published') {
    return {
      canRegister: !isRegistered,
      canEnter: false,
      canViewResult: false,
      badgeText: isRegistered ? '已报名' : '',
      badgeColor: 'success',
      teamSummaryText: isRegistered
        ? (hasTeam ? '已报名，队伍信息已就绪' : '已报名，等待分配队伍信息')
        : '未报名',
    }
  }

  if (contest.status === 'ongoing') {
    return {
      canRegister: false,
      canEnter: isRegistered,
      canViewResult: false,
      badgeText: isRegistered ? '参赛中' : '',
      badgeColor: 'processing',
      teamSummaryText: isRegistered
        ? (hasTeam ? '正在参赛' : '已报名，等待分配队伍信息')
        : '未报名',
    }
  }

  if (contest.status === 'ended') {
    return {
      canRegister: false,
      canEnter: false,
      canViewResult: true,
      badgeText: isRegistered ? '已参赛' : '已结束',
      badgeColor: isRegistered ? 'purple' : 'default',
      teamSummaryText: isRegistered
        ? (hasTeam ? '我曾以队伍身份参赛' : '我曾参赛，队伍信息暂缺')
        : '未参赛',
    }
  }

  return {
    canRegister: false,
    canEnter: false,
    canViewResult: false,
    badgeText: '',
    badgeColor: 'default',
    teamSummaryText: '比赛仍在准备阶段',
  }
}
