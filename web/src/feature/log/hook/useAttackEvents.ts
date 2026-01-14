import { useQuery } from '@tanstack/react-query'
import { logApi } from '@/api/log'
import { AttackEventQuery, AttackEventResponse } from '@/types/log'

const defaultData: AttackEventResponse = {
    results: [],
    totalCount: 0,
    pageSize: 10,
    currentPage: 1,
    totalPages: 0
}

export const useAttackEvents = (query: AttackEventQuery) => {
    return useQuery<AttackEventResponse, Error, AttackEventResponse, [string, AttackEventQuery]>({
        queryKey: ['attackEvents', query],
        queryFn: () => logApi.getAttackEvents(query),
        refetchOnMount: 'always',
        staleTime: 0,
        placeholderData: defaultData,
    })
}