// Espelho de main/internal/order/status.go
export const ORDER_STATUS = {
  created: 'Criado',
  stockOk: 'Estoque confirmado',
  outOfStock: 'Estoque indisponível',
  paymentOk: 'Pagamento aprovado',
  paymentFailed: 'Pagamento recusado',
  shipped: 'Enviado',
  cancelled: 'Excluído',
} as const

export type OrderStatus = (typeof ORDER_STATUS)[keyof typeof ORDER_STATUS]

export type StatusTone = 'neutral' | 'info' | 'success' | 'danger'

type StatusMeta = { tone: StatusTone; final: boolean }

const META: Record<OrderStatus, StatusMeta> = {
  [ORDER_STATUS.created]: { tone: 'neutral', final: false },
  [ORDER_STATUS.stockOk]: { tone: 'info', final: false },
  [ORDER_STATUS.paymentOk]: { tone: 'info', final: false },
  [ORDER_STATUS.shipped]: { tone: 'success', final: true },
  [ORDER_STATUS.outOfStock]: { tone: 'danger', final: true },
  [ORDER_STATUS.paymentFailed]: { tone: 'danger', final: true },
  [ORDER_STATUS.cancelled]: { tone: 'danger', final: true },
}

export function statusMeta(status: string): StatusMeta {
  return META[status as OrderStatus] ?? { tone: 'neutral', final: false }
}

export function isFinalStatus(status: string) {
  return statusMeta(status).final
}

// Pedido aguardando o consumidor pagar (quando o main passar a enviar paymentUrl).
export function canPay(status: string) {
  return status === ORDER_STATUS.stockOk
}
