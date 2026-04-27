import { SearchFilter } from '@/components/common'
import type { ChallengeSearchFilterProps } from '@/types/presentation'
import type { SearchFilterItem } from '@/types/presentation'
import { FILTER_OPTIONS } from '@/utils/constants'

export default function ChallengeSearchFilter({
  values,
  showDifficulty = true,
  onChange,
  onSearch,
  onReset,
}: ChallengeSearchFilterProps) {
  const filters: SearchFilterItem[] = [
    { key: 'keyword', type: 'input', placeholder: '搜索题目名称', label: '关键词' },
    { key: 'category', type: 'select', placeholder: '分类', options: [...FILTER_OPTIONS.CHALLENGE_CATEGORY], label: '分类' },
    ...(showDifficulty
      ? [{ key: 'difficulty', type: 'select', placeholder: '难度', options: [...FILTER_OPTIONS.CHALLENGE_DIFFICULTY], label: '难度' } as SearchFilterItem]
      : []),
  ]

  return (
    <SearchFilter
      filters={filters}
      values={values}
      onChange={(next) => onChange({
        keyword: typeof next.keyword === 'string' ? next.keyword : '',
        category: typeof next.category === 'string' ? next.category : '',
        difficulty: typeof next.difficulty === 'string' ? next.difficulty : '',
      })}
      onSearch={onSearch}
      onReset={onReset}
    />
  )
}
