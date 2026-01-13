import { useQuery } from '@tanstack/react-query'
import { logApi } from '@/api/log'
import { AttackEventQuery, AttackEventResponse } from '@/types/log'

export const useAttackEvents = (query: AttackEventQuery) => {
    return useQuery<AttackEventResponse, Error, AttackEventResponse, [string, AttackEventQuery]>({
        queryKey: ['attackEvents', query],
        queryFn: () => logApi.getAttackEvents(query),
        placeholderData: (previousData) => previousData,
    })
}