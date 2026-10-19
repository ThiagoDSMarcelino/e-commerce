import { queryOptions, useMutation, useQueryClient } from '@tanstack/react-query'

import { apiFetch } from './client'
import type { CancelPromotionInput, Paginated, RegisterPromotionInput, Registration } from './types'

export const promotionsKeys = {
  byClient: (clientId: number) => ['promotions', clientId] as const,
}

export const registrationsQueryOptions = (clientId: number) =>
  queryOptions({
    queryKey: promotionsKeys.byClient(clientId),
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ clientId: String(clientId), page: '1', size: '100' })
      return apiFetch<Paginated<Registration>>(`/promotions?${params}`, { signal })
    },
  })

export function useRegisterPromotion() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: RegisterPromotionInput) =>
      apiFetch<void>('/promotions', { method: 'POST', json: input }),
    onSuccess: (_, input) =>
      queryClient.invalidateQueries({ queryKey: promotionsKeys.byClient(input.clientId) }),
  })
}

export function useCancelPromotion() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CancelPromotionInput) =>
      apiFetch<void>('/promotions', { method: 'DELETE', json: input }),
    onSuccess: (_, input) =>
      queryClient.invalidateQueries({ queryKey: promotionsKeys.byClient(input.clientId) }),
  })
}
