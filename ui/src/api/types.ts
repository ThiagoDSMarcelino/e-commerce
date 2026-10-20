export type Product = {
  id: string
  name: string
  amount: number
  // Campos ainda não enviados pelo main; exibidos automaticamente quando existirem.
  price?: number
  category?: string
}

export type OrderItem = {
  productId: string
  quantity: number
}

export type Order = {
  id: string
  clientId: number
  createdAt: string
  status: string
  items: OrderItem[]
  // Futuro: URL do mock de pagamento servido pelo main.
  paymentUrl?: string
}

export type Paginated<T> = {
  items: T[]
  total: number
  page: number
}

export type CreateOrderInput = {
  clientId: number
  items: OrderItem[]
}

export type CancelOrderInput = {
  orderId: string
  clientId: number
}

export type Registration = {
  email: string
  createdAt: string
  categories: string[]
}

export type RegisterPromotionInput = {
  clientId: number
  email: string
  categories: string[]
}

export type CancelPromotionInput = {
  clientId: number
  email: string
}

// Payload do evento SSE `order-status` (GET /orders/events).
export type OrderStatusEvent = {
  orderId: string
  status: string
  updatedAt?: string
  paymentUrl?: string
}
