import { type QueryClient, useQueryClient } from '@tanstack/react-query'
import { createStore, useSelector } from '@tanstack/react-store'
import { useEffect } from 'react'
import { toast } from 'sonner'

import { apiUrl } from '@/api/client'
import { ordersKeys } from '@/api/orders'
import { productsKeys } from '@/api/products'
import type { Order, OrderStatusEvent, Paginated } from '@/api/types'
import { shortId } from '@/lib/format'
import { canPay, statusMeta } from '@/lib/order-status'
import { openPayment } from '@/lib/payment'

export type SseStatus = 'connecting' | 'open' | 'unavailable'

export const sseStatusStore = createStore<SseStatus>('connecting')

export function useSseStatus() {
  return useSelector(sseStatusStore)
}

const SSE_EVENT = 'order-status'
const RETRY_BASE_MS = 2_000
const RETRY_MAX_MS = 60_000

/**
 * Mantém uma conexão SSE com o main (GET /orders/events?clientId=) e aplica cada
 * mudança de status no cache do TanStack Query.
 *
 * Se o endpoint não existir ou cair, o EventSource fecha e tentamos novamente com
 * backoff exponencial. Enquanto isso, a tela de pedidos usa polling.
 */
export function useOrderEvents(clientId: number) {
  const queryClient = useQueryClient()

  useEffect(() => {
    let source: EventSource | null = null
    let retryTimer: ReturnType<typeof setTimeout> | undefined
    let attempts = 0
    let disposed = false

    const handleMessage = (event: MessageEvent<string>) => {
      const data = parseEvent(event.data)
      if (!data) return
      applyStatusEvent(queryClient, clientId, data)
      notifyStatusChange(data)
    }

    const connect = () => {
      if (disposed) return
      sseStatusStore.setState(() => 'connecting')

      source = new EventSource(apiUrl(`/orders/events?clientId=${clientId}`))

      source.onopen = () => {
        attempts = 0
        sseStatusStore.setState(() => 'open')
        // Ressincroniza o que possa ter mudado enquanto estávamos desconectados.
        void queryClient.invalidateQueries({ queryKey: ordersKeys.byClient(clientId) })
      }
      source.addEventListener(SSE_EVENT, handleMessage)
      source.onmessage = handleMessage
      source.onerror = () => {
        if (source?.readyState !== EventSource.CLOSED) {
          // Queda de conexão: o próprio EventSource tenta reconectar.
          sseStatusStore.setState(() => 'connecting')
          return
        }
        // Resposta inválida (404, 502...): o navegador desiste, então agendamos nova tentativa.
        source.close()
        source = null
        sseStatusStore.setState(() => 'unavailable')
        const delay = Math.min(RETRY_BASE_MS * 2 ** attempts, RETRY_MAX_MS)
        attempts++
        retryTimer = setTimeout(connect, delay)
      }
    }

    connect()

    return () => {
      disposed = true
      clearTimeout(retryTimer)
      source?.close()
    }
  }, [clientId, queryClient])
}

function parseEvent(raw: string): OrderStatusEvent | null {
  try {
    const data = JSON.parse(raw) as Partial<OrderStatusEvent>
    if (typeof data.orderId === 'string' && typeof data.status === 'string') {
      return data as OrderStatusEvent
    }
  } catch {
    // ignora eventos malformados (ex.: keep-alive)
  }
  return null
}

function applyStatusEvent(queryClient: QueryClient, clientId: number, event: OrderStatusEvent) {
  queryClient.setQueriesData<Paginated<Order>>(
    { queryKey: ordersKeys.byClient(clientId) },
    (old) =>
      old && {
        ...old,
        items: old.items.map((order) =>
          order.id === event.orderId
            ? { ...order, status: event.status, paymentUrl: event.paymentUrl ?? order.paymentUrl }
            : order,
        ),
      },
  )
  void queryClient.invalidateQueries({ queryKey: ordersKeys.byClient(clientId) })
  // Mudanças de status podem reservar ou devolver estoque.
  void queryClient.invalidateQueries({ queryKey: productsKeys.all })
}

function notifyStatusChange(event: OrderStatusEvent) {
  const title = `Pedido ${shortId(event.orderId)}: ${event.status}`
  const action =
    event.paymentUrl && canPay(event.status)
      ? { label: 'Pagar', onClick: () => openPayment(event.paymentUrl!) }
      : undefined

  switch (statusMeta(event.status).tone) {
    case 'success':
      toast.success(title, { action })
      break
    case 'danger':
      toast.error(title, { action })
      break
    default:
      toast.info(title, { action })
  }
}
