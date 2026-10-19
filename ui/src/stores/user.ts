import { createStore, useSelector } from '@tanstack/react-store'

import { readJson, writeJson } from '@/lib/storage'

const STORAGE_KEY = 'ecommerce:clientId'

export function isValidClientId(value: unknown): value is number {
  return typeof value === 'number' && Number.isInteger(value) && value > 0
}

const stored = readJson<unknown>(STORAGE_KEY, 1)

// Sem autenticação: o consumidor apenas escolhe o número do seu id.
export const clientIdStore = createStore<number>(isValidClientId(stored) ? stored : 1)

clientIdStore.subscribe((clientId) => writeJson(STORAGE_KEY, clientId))

export function setClientId(clientId: number) {
  if (isValidClientId(clientId)) clientIdStore.setState(() => clientId)
}

export function useClientId() {
  return useSelector(clientIdStore)
}
