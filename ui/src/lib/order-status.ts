// Espelho de main/internal/order/status.go
export const ORDER_STATUS = {
  created: 'Criado',
  stockOk: 'Estoque confirmado',
  awaitingPayment: 'Aguardando pagamento',
  outOfStock: 'Estoque indisponível',
  paymentOk: 'Pagamento aprovado',
  paymentFailed: 'Pagamento recusado',
  shipped: 'Enviado',
  cancelled: 'Cancelado',
} as const

export type OrderStatus = (typeof ORDER_STATUS)[keyof typeof ORDER_STATUS]

export type StatusTone = 'neutral' | 'info' | 'success' | 'danger'

type StatusMeta = { tone: StatusTone; final: boolean }

const META: Record<OrderStatus, StatusMeta> = {
  [ORDER_STATUS.created]: { tone: 'neutral', final: false },
  [ORDER_STATUS.stockOk]: { tone: 'info', final: false },
  [ORDER_STATUS.awaitingPayment]: { tone: 'info', final: false },
  [ORDER_STATUS.paymentOk]: { tone: 'info', final: false },
  [ORDER_STATUS.shipped]: { tone: 'success', final: true },
  [ORDER_STATUS.outOfStock]: { tone: 'danger', final: true },
  [ORDER_STATUS.paymentFailed]: { tone: 'danger', final: true },
  [ORDER_STATUS.cancelled]: { tone: 'danger', final: true },
}

export function statusMeta(status: string): StatusMeta {
  return META[status as OrderStatus] ?? { tone: 'neutral', final: false }
}

// Espelho de nonCancellableStatuses em main/internal/order/status.go
const NON_CANCELLABLE: string[] = [
  ORDER_STATUS.shipped,
  ORDER_STATUS.cancelled,
  ORDER_STATUS.outOfStock,
  ORDER_STATUS.paymentFailed,
]

export function canCancel(status: string) {
  return !NON_CANCELLABLE.includes(status)
}

// Pedido aguardando o consumidor pagar o link recebido do pagamento.
export function canPay(status: string) {
  return status === ORDER_STATUS.awaitingPayment
}
