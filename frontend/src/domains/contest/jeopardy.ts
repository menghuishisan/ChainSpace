import DOMPurify from 'dompurify'
import { marked } from 'marked'

import { ChallengeRuntimeProfileMap } from '@/types'
import type { CategoryGroup, Challenge, ChallengeEnv, ChallengeRuntimeProfile } from '@/types'

export function renderChallengeMarkdown(markdown: string): string {
  const html = marked.parse(markdown, { async: false })
  return DOMPurify.sanitize(typeof html === 'string' ? html : '')
}

export function getRuntimeDescription(runtimeProfile: ChallengeRuntimeProfile): string {
	switch (runtimeProfile) {
		case 'static':
			return '本题无需启动环境，请结合题面、附件和代码完成分析。'
		case 'single_chain_instance':
			return '启动后可在独立环境中完成部署、交互和验证操作。'
		case 'fork_replay':
			return '启动后可使用提供的环境和链上接口继续复现与验证。'
		case 'multi_service_lab':
			return '启动后可使用题目提供的多个服务或组件完成分析。'
		default:
			return '请结合题面说明完成分析、验证和提交。'
	}
}

export function getRuntimeHeadline(runtimeProfile: ChallengeRuntimeProfile): string {
	switch (runtimeProfile) {
		case 'static':
			return '题目说明'
		case 'single_chain_instance':
			return '环境说明'
		case 'fork_replay':
			return '环境说明'
		case 'multi_service_lab':
			return '环境说明'
		default:
			return '题目环境'
	}
}

export function getRuntimeProfileLabel(runtimeProfile: string): string {
  if (runtimeProfile in ChallengeRuntimeProfileMap) {
    return ChallengeRuntimeProfileMap[runtimeProfile as ChallengeRuntimeProfile]
  }

  return runtimeProfile
}

export function buildChallengeCategoryGroups(
  challenges: Challenge[],
  categoryMap: Record<string, string>,
): CategoryGroup[] {
  return Object.entries(categoryMap)
    .map(([category, label]) => ({
      category,
      label,
      challenges: challenges.filter((challenge) => challenge.category === category),
    }))
    .filter((group) => group.challenges.length > 0)
}

export function getSolvedChallengeCount(challenges: Challenge[]): number {
  return challenges.filter((challenge) => challenge.is_solved).length
}

export function getSolvedChallengePoints(challenges: Challenge[]): number {
  return challenges.reduce((sum, challenge) => (
    challenge.is_solved ? sum + (challenge.points ?? challenge.base_points) : sum
  ), 0)
}

export function extractContractName(code: string): string {
  const match = code.match(/contract\s+(\w+)/)
  return match ? match[1] : 'contract'
}

export function normalizeChallengeEnvState(env: ChallengeEnv | null) {
  if (!env) {
    return null
  }

  return env.status === 'running'
    || env.status === 'creating'
    || env.status === 'expired'
    || env.status === 'failed'
    ? env
    : null
}

export function getContestChallengePreviewAccess(
  contestStatus: string,
  isRegistered: boolean,
  isManager: boolean,
) {
  const showChallengePreview = contestStatus === 'ended' || (contestStatus === 'ongoing' && isRegistered) || isManager

  return {
    showChallengePreview,
    useReviewChallenges: contestStatus === 'ended',
  }
}

export function getContestStatusSummary(contestStatus: string, isRegistered: boolean): string {
  if (contestStatus === 'ended') {
    return '比赛已结束，这里展示赛后题目回顾与比赛结果信息。'
  }

  if (contestStatus === 'ongoing') {
    return isRegistered
      ? '比赛进行中，你可以进入正式赛场解题。'
      : '比赛进行中，未报名选手无法进入正式赛场。'
  }

  if (contestStatus === 'published') {
    return '比赛正在报名中，请在截止前完成报名或组队。'
  }

  return '比赛仍处于草稿或准备阶段，当前以基础信息展示为主。'
}
