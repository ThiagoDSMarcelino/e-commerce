import { keepPreviousData, queryOptions, useMutation, useQueryClient } from '@tanstack/react-query'

import { apiFetch } from './client'
import { productsKeys } from './products'
import type { CancelOrderInput, CreateOrderInput, Order, Paginated } from './types'

export const ordersKeys = {
  all: ['orders'] as const,
  byClient: (clientId: number) => ['orders', clientId] as const,
  page: (clientId: number, page: number, size: number) =>
    ['orders', clientId, { page, size }] as const,
}

export const ordersQueryOptions = (clientId: number, page: number, size: number) =>
  queryOptions({
    queryKey: ordersKeys.page(clientId, page, size),
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({
        clientId: String(clientId),
        page: String(page),
        size: String(size),
      })
      return apiFetch<Paginated<Order>>(`/orders?${params}`, { signal })
    },
    placeholderData: keepPreviousData,
  })

export function useCreateOrder() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateOrderInput) =>
      apiFetch<Order>('/orders', { method: 'POST', json: input }),
    onSuccess: (_, input) => {
      void queryClient.invalidateQueries({ queryKey: ordersKeys.byClient(input.clientId) })
      void queryClient.invalidateQueries({ queryKey: productsKeys.all })
    },
  })
}

export function useCancelOrder() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ orderId, clientId }: CancelOrderInput) =>
      apiFetch<Order>(`/orders/${orderId}/cancel`, { method: 'POST', json: { clientId } }),
    onSuccess: (cancelled, { clientId }) => {
      queryClient.setQueriesData<Paginated<Order>>(
        { queryKey: ordersKeys.byClient(clientId) },
        (old) =>
          old && {
            ...old,
            items: old.items.map((order) => (order.id === cancelled.id ? cancelled : order)),
          },
      )
      void queryClient.invalidateQueries({ queryKey: ordersKeys.byClient(clientId) })
      void queryClient.invalidateQueries({ queryKey: productsKeys.all })
    },
  })
}
