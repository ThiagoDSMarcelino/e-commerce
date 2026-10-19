import { createStore, useSelector } from '@tanstack/react-store'

import { readJson, writeJson } from '@/lib/storage'

const STORAGE_KEY = 'ecommerce:cart'

export type CartItem = {
  productId: string
  name: string
  quantity: number
}

// Um carrinho por usuário, indexado pelo clientId.
type CartState = Record<string, CartItem[]>

const EMPTY: CartItem[] = []

export const cartStore = createStore<CartState>(readJson<CartState>(STORAGE_KEY, {}))

cartStore.subscribe((state) => writeJson(STORAGE_KEY, state))

function update(clientId: number, fn: (items: CartItem[]) => CartItem[]) {
  cartStore.setState((state) => ({ ...state, [clientId]: fn(state[clientId] ?? EMPTY) }))
}

export const cartActions = {
  add(clientId: number, product: { id: string; name: string }, quantity: number) {
    update(clientId, (items) => {
      const existing = items.find((it) => it.productId === product.id)
      if (!existing) return [...items, { productId: product.id, name: product.name, quantity }]
      return items.map((it) =>
        it.productId === product.id ? { ...it, quantity: it.quantity + quantity } : it,
      )
    })
  },
  setQuantity(clientId: number, productId: string, quantity: number) {
    update(clientId, (items) =>
      items.map((it) => (it.productId === productId ? { ...it, quantity } : it)),
    )
  },
  remove(clientId: number, productId: string) {
    update(clientId, (items) => items.filter((it) => it.productId !== productId))
  },
  clear(clientId: number) {
    update(clientId, () => EMPTY)
  },
}

export function useCart(clientId: number) {
  return useSelector(cartStore, (state) => state[clientId] ?? EMPTY)
}

export function useCartCount(clientId: number) {
  return useSelector(cartStore, (state) =>
    (state[clientId] ?? EMPTY).reduce((sum, it) => sum + it.quantity, 0),
  )
}
