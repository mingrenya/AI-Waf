import { useQuery } from '@tanstack/react-query'
import { logApi } from '@/api/log'
import { AttackLogQuery, AttackLogResponse } from '@/types/log'

export const useAttackLogs = (query: AttackLogQuery) => {
  return useQuery<AttackLogResponse, Error, AttackLogResponse, [string, AttackLogQuery]>({
    queryKey: ['attackLogs', query],
    queryFn: () => logApi.getAttackLogs(query),
    placeholderData: (previousData) => previousData,
  })
} 