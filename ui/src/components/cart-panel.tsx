import { useNavigate } from '@tanstack/react-router'
import { Loader2, ShoppingCart, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

import { useCreateOrder } from '@/api/orders'
import type { Product } from '@/api/types'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { shortId } from '@/lib/format'
import { openPayment } from '@/lib/payment'
import { cartActions, useCart } from '@/stores/cart'

type CartPanelProps = {
  clientId: number
  // Estoque atual por produto, para limitar as quantidades.
  stock: Map<string, Product>
}

export function CartPanel({ clientId, stock }: CartPanelProps) {
  const items = useCart(clientId)
  const createOrder = useCreateOrder()
  const navigate = useNavigate()

  const totalUnits = items.reduce((sum, it) => sum + it.quantity, 0)
  const unavailable = items.filter((it) => (stock.get(it.productId)?.amount ?? 0) < it.quantity)

  const checkout = () => {
    createOrder.mutate(
      {
        clientId,
        items: items.map(({ productId, quantity }) => ({ productId, quantity })),
      },
      {
        onSuccess: (order) => {
          cartActions.clear(clientId)

          if (order.paymentUrl) {
            toast.success(`Pedido ${shortId(order.id)} criado`, {
              description: 'Abrimos o pagamento em uma nova aba.',
            })
            openPayment(order.paymentUrl)
          } else {
            toast.success(`Pedido ${shortId(order.id)} criado`, {
              description: 'Acompanhe o status na página de pedidos.',
            })
          }

          void navigate({ to: '/orders' })
        },
        onError: (err) =>
          toast.error('Não foi possível finalizar o pedido', { description: err.message }),
      },
    )
  }

  return (
    <Card className="lg:sticky lg:top-20">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ShoppingCart className="size-5" /> Carrinho
        </CardTitle>
        <CardDescription>
          {totalUnits === 0
            ? 'Nenhum item adicionado.'
            : `${totalUnits} unidade(s) em ${items.length} produto(s)`}
        </CardDescription>
      </CardHeader>

      {items.length > 0 && (
        <CardContent className="flex flex-col gap-3">
          {items.map((item) => {
            const available = stock.get(item.productId)?.amount ?? 0
            return (
              <div key={item.productId} className="flex items-center gap-2">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{item.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {item.productId} · {available} disponível(is)
                  </p>
                </div>
                <Input
                  type="number"
                  min={1}
                  max={available || undefined}
                  className="h-8 w-16"
                  aria-label={`Quantidade de ${item.name}`}
                  value={item.quantity}
                  aria-invalid={item.quantity > available}
                  onChange={(e) => {
                    const value = Math.floor(Number(e.target.value))
                    if (value > 0) cartActions.setQuantity(clientId, item.productId, value)
                  }}
                />
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Remover ${item.name}`}
                  onClick={() => cartActions.remove(clientId, item.productId)}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            )
          })}

          {unavailable.length > 0 && (
            <p className="text-xs text-destructive">
              Quantidade acima do estoque para: {unavailable.map((it) => it.name).join(', ')}.
            </p>
          )}
        </CardContent>
      )}

      <CardFooter className="flex gap-2">
        <Button
          className="flex-1"
          disabled={items.length === 0 || unavailable.length > 0 || createOrder.isPending}
          onClick={checkout}
        >
          {createOrder.isPending && <Loader2 className="size-4 animate-spin" />}
          Finalizar pedido
        </Button>
        {items.length > 0 && (
          <Button variant="outline" onClick={() => cartActions.clear(clientId)}>
            Limpar
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}
