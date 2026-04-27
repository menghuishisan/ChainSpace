import { Badge, Card, Col, Row, Tag } from 'antd'
import type { JeopardyChallengeGridProps } from '@/types/presentation'

export default function JeopardyChallengeGrid({
  categoryGroups,
  selectedChallengeId,
  runtimeLabels,
  difficultyMap,
  compact = false,
  onSelectChallenge,
}: JeopardyChallengeGridProps) {
  return (
    <div className="space-y-5">
      {categoryGroups.map((group) => (
        <Card
          key={group.category}
          className="overflow-hidden border-0 shadow-sm"
          styles={{ body: { padding: 20 } }}
          title={(
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Tag color="blue" className="mr-0">
                  {group.label}
                </Tag>
                <span className="text-sm text-text-secondary">
                  {group.challenges.filter((challenge) => challenge.is_solved).length}/{group.challenges.length} 已解
                </span>
              </div>
              <span className="text-xs uppercase tracking-[0.2em] text-text-secondary">
                {group.category}
              </span>
            </div>
          )}
        >
          {compact ? (
            <div className="space-y-3">
              {group.challenges.map((challenge) => {
                const difficulty = difficultyMap[challenge.difficulty]
                const selected = selectedChallengeId === challenge.id

                return (
                  <button
                    type="button"
                    key={challenge.id}
                    className={[
                      'w-full rounded-2xl border px-4 py-3 text-left transition-all duration-200',
                      'hover:-translate-y-0.5 hover:shadow-sm',
                      challenge.is_solved ? 'border-emerald-300 bg-emerald-50/80' : 'border-slate-200 bg-white',
                      selected ? 'border-primary shadow-md ring-2 ring-primary/15' : '',
                    ].join(' ')}
                    onClick={() => onSelectChallenge(challenge)}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="line-clamp-2 text-sm font-semibold leading-6 text-slate-900">
                          {challenge.title}
                        </div>
                        <div className="mt-2 flex flex-wrap gap-2">
                          <Tag color={difficulty?.color} className="mr-0">{difficulty?.text}</Tag>
                          <Tag className="mr-0">{runtimeLabels[challenge.runtime_profile]}</Tag>
                          {challenge.is_solved ? <Tag color="success" className="mr-0">已完成</Tag> : null}
                        </div>
                      </div>
                      <div className="shrink-0 rounded-xl bg-slate-900 px-2.5 py-1 text-sm font-semibold text-white">
                        {challenge.points ?? challenge.base_points}
                      </div>
                    </div>
                    <div className="mt-3 flex items-center justify-between text-xs text-text-secondary">
                      <span>{challenge.solve_count ?? 0} 人解出</span>
                      <span>ID #{challenge.id}</span>
                    </div>
                  </button>
                )
              })}
            </div>
          ) : (
            <Row gutter={[16, 16]}>
              {group.challenges.map((challenge) => {
                const difficulty = difficultyMap[challenge.difficulty]
                const selected = selectedChallengeId === challenge.id

                return (
                  <Col xs={24} md={12} xxl={8} key={challenge.id}>
                    <Card
                      hoverable
                      className={[
                        'h-full cursor-pointer border transition-all duration-200',
                        'hover:-translate-y-0.5 hover:shadow-md',
                        challenge.is_solved ? 'border-emerald-300 bg-emerald-50/70' : 'border-slate-200',
                        selected ? 'border-primary shadow-lg ring-2 ring-primary/15' : '',
                      ].join(' ')}
                      styles={{ body: { padding: 18 } }}
                      onClick={() => onSelectChallenge(challenge)}
                    >
                      <div className="flex h-full flex-col gap-4">
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0">
                            <Badge count={challenge.is_solved ? <span className="text-sm text-success">✓</span> : 0} offset={[-6, 4]}>
                              <h4 className="line-clamp-2 text-base font-semibold leading-6 text-slate-900">
                                {challenge.title}
                              </h4>
                            </Badge>
                          </div>
                          <div className="shrink-0 rounded-2xl bg-slate-900 px-3 py-1 text-sm font-semibold text-white">
                            {challenge.points ?? challenge.base_points}
                          </div>
                        </div>

                        <div className="flex flex-wrap gap-2">
                          <Tag color={difficulty?.color}>{difficulty?.text}</Tag>
                          <Tag>{runtimeLabels[challenge.runtime_profile]}</Tag>
                          {challenge.is_solved ? <Tag color="success">已完成</Tag> : null}
                        </div>

                        <div className="mt-auto flex items-center justify-between text-xs text-text-secondary">
                          <span>{challenge.solve_count ?? 0} 人解出</span>
                          <span>ID #{challenge.id}</span>
                        </div>
                      </div>
                    </Card>
                  </Col>
                )
              })}
            </Row>
          )}
        </Card>
      ))}
    </div>
  )
}
