/**
 * CTF竞赛相关 API
 */
import { get, post, put, del } from './request'
import type { 
  Contest, 
  Challenge,
  ContestChallenge,
  ChallengeEnv,
  Team,
  Scoreboard,
  AgentBattleScore,
  FlagSubmitResult,
  AgentBattleEvent,
  AgentBattleStatus,
  SpectateData,
  CurrentRoundInfo,
  PaginatedData,
  CreateContestRequest,
  ContestRecord,
} from '@/types'
import type { TeamWithContest } from '@/types/presentation'
import { API_SERVER_ORIGIN } from '@/utils/constants'

// ====== 竞赛管理 ======

/**
 * 获取竞赛列表
 * 后端路由: GET /contests
 */
export function getContests(params?: {
  page?: number
  page_size?: number
  type?: string
  level?: string
  status?: string
  keyword?: string
  is_public?: boolean
}): Promise<PaginatedData<Contest>> {
  return get<PaginatedData<Contest>>('/contests', params)
}

/**
 * 创建竞赛
 * 后端路由: POST /contests
 */
export function createContest(data: CreateContestRequest): Promise<Contest> {
  return post<Contest>('/contests', data)
}

/**
 * 获取竞赛详情
 * 后端路由: GET /contests/:id
 */
export function getContest(id: number): Promise<Contest> {
  return get<Contest>(`/contests/${id}`)
}

/**
 * 更新竞赛
 * 后端路由: PUT /contests/:id
 */
export function updateContest(id: number, data: Partial<CreateContestRequest>): Promise<Contest> {
  return put<Contest>(`/contests/${id}`, data)
}

/**
 * 发布竞赛
 * 后端路由: PUT /contests/:id/publish
 */
export function publishContest(id: number): Promise<void> {
  return put(`/contests/${id}/publish`)
}

/**
 * 删除竞赛
 * 后端路由: DELETE /contests/:id
 */
export function deleteContest(id: number): Promise<void> {
  return del(`/contests/${id}`)
}

// ====== 队伍管理 ======

/**
 * 创建队伍
 * 后端路由: POST /teams
 */
export function createTeam(data: { name: string; contest_id?: number }): Promise<Team> {
  return post<Team>('/teams', data)
}

/**
 * 加入队伍
 * 后端路由: POST /teams/join
 */
export function joinTeam(inviteCode: string): Promise<Team> {
  return post<Team>('/teams/join', { invite_code: inviteCode })
}

/**
 * 提交 Flag
 * 后端路由: POST /contests/:id/submit
 */
export function submitFlag(contestId: number, challengeId: number, flag: string): Promise<FlagSubmitResult> {
  return post<FlagSubmitResult>(`/contests/${contestId}/submit`, { challenge_id: challengeId, flag })
}

/**
 * 获取排行榜
 * 后端路由: GET /contests/:id/scoreboard
 */
export function getScoreboard(contestId: number): Promise<Scoreboard> {
  return get<Scoreboard>(`/contests/${contestId}/scoreboard`)
}

// ====== 题目环境 ======

function mapAgentBattleScoreboard(scores: AgentBattleScore[]): Scoreboard {
  return {
    list: scores.map((score) => ({
      rank: score.rank,
      team_id: score.team_id,
      team_name: score.team_name,
      total_score: score.score,
      solve_count: score.success_count,
    })),
  }
}

function mapRoundEventsPage(data: PaginatedData<AgentBattleEvent>): AgentBattleEvent[] {
  return data.list || []
}

function resolveContestRuntimeAccessUrl(url?: string): string | undefined {
  if (!url) {
    return url
  }
  if (/^https?:\/\//i.test(url)) {
    return url
  }
  return `${API_SERVER_ORIGIN}${url}`
}

function normalizeChallengeEnv(env: ChallengeEnv): ChallengeEnv {
  return {
    ...env,
    access_url: resolveContestRuntimeAccessUrl(env.access_url),
    tools: (env.tools || []).map((tool) => ({
      ...tool,
      route: resolveContestRuntimeAccessUrl(tool.route) || tool.route,
      instance_route: resolveContestRuntimeAccessUrl(tool.instance_route) || tool.instance_route,
    })),
    service_entries: (env.service_entries || []).map((entry) => ({
      ...entry,
      access_url: resolveContestRuntimeAccessUrl(entry.access_url),
    })),
  }
}

/**
 * 启动题目环境
 * 后端路由: POST /contests/:id/challenges/:cid/env
 */
export function startChallengeEnv(contestId: number, challengeId: number): Promise<ChallengeEnv> {
  return post<ChallengeEnv>(`/contests/${contestId}/challenges/${challengeId}/env`).then(normalizeChallengeEnv)
}

/**
 * 获取题目环境状态
 * 后端路由: GET /contests/:id/challenges/:cid/env
 */
export function getChallengeEnvStatus(contestId: number, challengeId: number): Promise<ChallengeEnv | null> {
  return get<ChallengeEnv | null>(`/contests/${contestId}/challenges/${challengeId}/env`).then((env) => (env ? normalizeChallengeEnv(env) : null))
}

export function getChallengeAttachmentAccessUrl(
  contestId: number,
  challengeId: number,
  attachmentIndex: number,
): Promise<string> {
  return get<{ url: string }>(
    `/contests/${contestId}/challenges/${challengeId}/attachments/${attachmentIndex}/access-url`,
  ).then((payload) => payload.url)
}

/**
 * 停止题目环境
 * 后端路由: DELETE /contests/:id/challenges/:cid/env
 */
export function stopChallengeEnv(contestId: number, challengeId: number): Promise<void> {
  return del(`/contests/${contestId}/challenges/${challengeId}/env`)
}

// ====== 智能体对战（Agent Battle）======
// 后端路由: /agent-battle/*

/**
 * 获取轮次列表
 * 后端路由: GET /agent-battle/contests/:contest_id/rounds
 */
export function getAgentBattleRounds(contestId: number): Promise<Array<{
  id: number
  round_number: number
  status: string
  phase?: string
  upgrade_window_end?: string
  start_time?: string
  end_time?: string
}>> {
  return get<PaginatedData<{
    id: number
    round_number: number
    status: string
    phase?: string
    upgrade_window_end?: string
    start_time?: string
    end_time?: string
  }>>(`/agent-battle/contests/${contestId}/rounds`).then((data) => data.list || [])
}

/**
 * 获取当前轮次
 * 后端路由: GET /agent-battle/contests/:contest_id/current-round
 */
export function getCurrentRound(contestId: number): Promise<CurrentRoundInfo | null> {
  return get<CurrentRoundInfo | null>(`/agent-battle/contests/${contestId}/current-round`)
}

/**
 * 获取智能体对战排行榜
 * 后端路由: GET /agent-battle/contests/:contest_id/scoreboard
 */
export function getAgentBattleScoreboard(contestId: number): Promise<Scoreboard> {
  return get<AgentBattleScore[]>(`/agent-battle/contests/${contestId}/scoreboard`).then(mapAgentBattleScoreboard)
}

/**
 * 部署合约
 * 后端路由: POST /agent-battle/contracts/deploy
 */
export function deployContract(data: {
  contest_id: number
  source_code: string
}): Promise<{
  id: number
  contract_address: string
  status: string
  version: number
  deployed_at?: string
}> {
  return post('/agent-battle/contracts/deploy', data)
}

/**
 * 升级合约
 * 后端路由: POST /agent-battle/contracts/upgrade
 */
export function upgradeContract(data: {
  contest_id: number
  new_implementation: string
}): Promise<{
  id: number
  contract_address: string
  status: string
  version?: number
  deployed_at?: string
}> {
  return post('/agent-battle/contracts/upgrade', data)
}

/**
 * 获取队伍合约
 * 后端路由: GET /agent-battle/contests/:contest_id/teams/:team_id/contract
 */
export function getTeamContract(contestId: number, teamId: number): Promise<{
  id: number
  contest_id: number
  team_id: number
  team_name: string
  contract_address: string
  status: string
  version: number
  deployed_at?: string
} | null> {
  return get(`/agent-battle/contests/${contestId}/teams/${teamId}/contract`)
}

/**
 * 获取轮次事件
 * 后端路由: GET /agent-battle/rounds/:round_id/events
 */
export function getRoundEvents(roundId: number): Promise<AgentBattleEvent[]> {
  return get<PaginatedData<AgentBattleEvent>>(`/agent-battle/rounds/${roundId}/events`).then(mapRoundEventsPage)
}

/**
 * 创建轮次（管理员）
 * 后端路由: POST /agent-battle/contests/:contest_id/rounds
 */
export function createRound(contestId: number, data: {
  round_number: number
  start_time: string
  end_time: string
  description?: string
}): Promise<{
  id: number
  contest_id: number
  round_number: number
  status: string
  start_time: string
  end_time: string
}> {
  return post(`/agent-battle/contests/${contestId}/rounds`, data)
}

/**
 * 开始轮次（管理员）
 * 后端路由: POST /agent-battle/rounds/:round_id/start
 */

/**
 * 结束轮次（管理员）
 * 后端路由: POST /agent-battle/rounds/:round_id/end
 */

// ====== 竞赛报名 ======

/**
 * 报名竞赛
 * 后端路由: POST /contests/:id/register
 */
export function registerContest(contestId: number): Promise<void> {
  return post(`/contests/${contestId}/register`)
}

/**
 * 获取我在竞赛中的队伍
 * 后端路由: GET /contests/:id/my-team
 */
export function getMyTeam(contestId: number): Promise<Team | null> {
  return get<Team | null>(`/contests/${contestId}/my-team`, undefined, { _silent: true })
}

/**
 * 上传智能体代码
 * 后端路由: POST /contests/:id/agent/upload
 */
export function uploadAgentCode(contestId: number, file: File): Promise<{ version: string }> {
  const formData = new FormData()
  formData.append('file', file)
  return post(`/contests/${contestId}/agent/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/**
 * 获取战斗事件
 * 后端路由: GET /agent-battle/contests/:contest_id/events
 */
export function getBattleEvents(contestId: number, params?: {
  round_id?: number
  from_block?: number
  limit?: number
}): Promise<AgentBattleEvent[]> {
  return get<AgentBattleEvent[]>(`/agent-battle/contests/${contestId}/events`, params)
}

/**
 * 获取回放数据
 * 后端路由: GET /agent-battle/contests/:contest_id/replay
 */
export function getReplayData(contestId: number, params: {
  round_id: number
}): Promise<{
  round_id: number
  start_block: number
  end_block: number
  snapshots: Array<{
    block: number
    teams: Array<{
      team_id: number
      score: number
      resource: number
    }>
    events: Array<{
      event_type: string
      actor_team?: string
      target_team?: string
      action_result?: string
      score_delta?: number
      resource_delta?: number
      description: string
    }>
  }>
}> {
  return get(`/agent-battle/contests/${contestId}/replay`, params)
}

/**
 * 获取我的竞赛记录
 * 后端路由: GET /contests/my-records
 */
export function getMyContestRecords(params?: {
  page?: number
  page_size?: number
}): Promise<PaginatedData<ContestRecord>> {
  return get('/contests/my-records', params)
}

/**
 * 获取我的所有队伍列表
 * 后端路由: GET /teams/my
 */
export function getMyTeams(): Promise<TeamWithContest[]> {
  return get<TeamWithContest[]>('/teams/my')
}

/**
 * 离开队伍
 * 后端路由: POST /teams/:team_id/leave
 */
export function leaveTeam(teamId: number): Promise<void> {
  return post(`/teams/${teamId}/leave`)
}

/**
 * 邀请队伍成员
 * 后端路由: POST /teams/:team_id/invite
 */
export function inviteTeamMember(teamId: number, userId: number): Promise<void> {
  return post(`/teams/${teamId}/invite`, { user_id: userId })
}

/**
 * 移除队伍成员
 * 后端路由: DELETE /teams/:team_id/members/:user_id
 */
export function removeTeamMember(teamId: number, userId: number): Promise<void> {
  return del(`/teams/${teamId}/members/${userId}`)
}

/**
 * 获取竞赛题目（参赛用）
 * 后端路由: GET /contests/:id/challenges
 */
export function getPlayingChallenges(contestId: number): Promise<Challenge[]> {
  return get<Challenge[]>(`/contests/${contestId}/challenges`)
}

export function getContestReviewChallenges(contestId: number): Promise<Challenge[]> {
  return get<Challenge[]>(`/contests/${contestId}/review-challenges`)
}

// ====== 竞赛题目管理（管理员/教师）======

/**
 * 获取竞赛题目列表（管理视角）
 * 后端路由: GET /contests/:id/challenges/admin
 */
export function getContestChallengesAdmin(contestId: number): Promise<ContestChallenge[]> {
  return get<ContestChallenge[]>(`/contests/${contestId}/challenges/admin`)
}

/**
 * 添加题目到竞赛
 * 后端路由: POST /contests/:id/challenges
 */
export function addChallengeToContest(contestId: number, data: {
  challenge_id: number
  points?: number
  sort_order?: number
  is_visible?: boolean
}): Promise<ContestChallenge> {
  return post<ContestChallenge>(`/contests/${contestId}/challenges`, data)
}

/**
 * 从竞赛移除题目
 * 后端路由: DELETE /contests/:id/challenges/:challengeId
 */
export function removeChallengeFromContest(contestId: number, challengeId: number): Promise<void> {
  return del(`/contests/${contestId}/challenges/${challengeId}`)
}

/**
 * 获取对抗赛状态
 * 后端路由: GET /agent-battle/contests/:contest_id/status
 */
export function getAgentBattleStatus(contestId: number): Promise<AgentBattleStatus> {
  return get<AgentBattleStatus>(`/agent-battle/contests/${contestId}/status`)
}

/**
 * 获取观战数据
 * 后端路由: GET /agent-battle/contests/:contest_id/spectate
 */
export function getSpectateData(contestId: number): Promise<SpectateData> {
  return get<SpectateData>(`/agent-battle/contests/${contestId}/spectate`).then((data) => ({
    ...data,
    teams: (data.teams || []).map((team) => {
      const legacyTeam = team as typeof team & { score?: number }
      return {
        ...team,
        total_score: team.total_score ?? legacyTeam.score ?? 0,
        resource_held: team.resource_held ?? 0,
        is_alive: team.is_alive ?? true,
      }
    }),
  }))
}

/**
 * 获取对抗赛最终排名
 * 后端路由: GET /agent-battle/contests/:contest_id/final-rank
 */
export function getFinalRank(contestId: number): Promise<Array<{
  rank: number
  team_id: number
  team_name: string
  total_score: number
}>> {
  return get(`/agent-battle/contests/${contestId}/final-rank`)
}

// ====== 对抗赛队伍工作区 ======

export interface TeamWorkspace {
  team_id: number
  team_name: string
  env_id?: string
  pod_name: string
  status: string
  access_url?: string
  chain_rpc_url?: string
  tools?: Array<{
    key: string
    label: string
    kind?: string
    target?: string
    port: number
    route: string
  }>
}

function normalizeTeamWorkspace(workspace: TeamWorkspace): TeamWorkspace {
  const tools = (workspace.tools || []).map((tool) => ({
    ...tool,
    route: resolveContestRuntimeAccessUrl(tool.route) || tool.route,
  }))
  return {
    ...workspace,
    access_url: resolveContestRuntimeAccessUrl(workspace.access_url),
    chain_rpc_url: resolveContestRuntimeAccessUrl(workspace.chain_rpc_url),
    tools,
  }
}

/**
 * 获取当前用户所在队伍的工作区
 * 后端路由: GET /agent-battle/contests/:contest_id/workspace
 */
export function getTeamWorkspace(contestId: number): Promise<TeamWorkspace> {
  return get<TeamWorkspace>(`/agent-battle/contests/${contestId}/workspace`).then(normalizeTeamWorkspace)
}

/**
 * 创建或获取队伍工作区
 * 后端路由: POST /agent-battle/contests/:contest_id/workspace
 */
export function createTeamWorkspace(contestId: number): Promise<TeamWorkspace> {
  return post<TeamWorkspace>(`/agent-battle/contests/${contestId}/workspace`).then(normalizeTeamWorkspace)
}

/**
 * 停止队伍工作区
 * 后端路由: DELETE /agent-battle/contests/:contest_id/workspace
 */
export function stopTeamWorkspace(contestId: number): Promise<void> {
  return del(`/agent-battle/contests/${contestId}/workspace`)
}
