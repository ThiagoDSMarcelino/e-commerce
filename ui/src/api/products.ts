import { queryOptions } from '@tanstack/react-query'

import { apiFetch } from './client'
import type { Product } from './types'

export const productsKeys = {
  all: ['products'] as const,
}

export const productsQueryOptions = () =>
  queryOptions({
    queryKey: productsKeys.all,
    queryFn: ({ signal }) => apiFetch<Product[]>('/products', { signal }),
    staleTime: 30_000,
  })
