import { useQuery } from '@tanstack/react-query'
import { logApi } from '@/api/log'
import { AttackLogQuery, AttackLogResponse } from '@/types/log'

const defaultData: AttackLogResponse = {
  results: [],
  totalCount: 0,
  pageSize: 10,
  currentPage: 1,
  totalPages: 0
}

export const useAttackLogs = (query: AttackLogQuery) => {
  return useQuery<AttackLogResponse, Error, AttackLogResponse, [string, AttackLogQuery]>({
    queryKey: ['attackLogs', query],
    queryFn: () => logApi.getAttackLogs(query),
    refetchOnMount: 'always',
    staleTime: 0,
    placeholderData: defaultData,
  })
} 