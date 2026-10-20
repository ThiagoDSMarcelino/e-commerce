import { useQuery } from '@tanstack/react-query'
import {
  createColumnHelper,
  type PaginationState,
  rowExpandingFeature,
  rowPaginationFeature,
  tableFeatures,
  type Updater,
  useTable,
} from '@tanstack/react-table'
import { createFileRoute, Link } from '@tanstack/react-router'
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CreditCard,
  Loader2,
  RefreshCw,
  X,
} from 'lucide-react'
import { useMemo } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'

import { ordersQueryOptions, useCancelOrder } from '@/api/orders'
import { productsQueryOptions } from '@/api/products'
import type { Order } from '@/api/types'
import { DataTable } from '@/components/data-table'
import { OrderStatusBadge } from '@/components/order-status-badge'
import { PageHeader, QueryError } from '@/components/page-header'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { formatDateTime, shortId } from '@/lib/format'
import { canCancel, canPay } from '@/lib/order-status'
import { openPayment } from '@/lib/payment'
import { cn } from '@/lib/utils'
import { useClientId } from '@/stores/user'

const PAGE_SIZES = [5, 10, 20, 50] as const

const searchSchema = z.object({
  page: z.number().int().min(1).default(1).catch(1),
  size: z.number().int().min(1).max(100).default(10).catch(10),
})

export const Route = createFileRoute('/orders')({
  validateSearch: searchSchema,
  component: OrdersPage,
})

const features = tableFeatures({ rowPaginationFeature, rowExpandingFeature })

const helper = createColumnHelper<typeof features, Order>()

const columns = helper.columns([
  helper.display({
    id: 'expand',
    header: () => <span className="sr-only">Detalhes</span>,
    cell: ({ row }) => (
      <Button
        variant="ghost"
        size="icon"
        aria-label={row.getIsExpanded() ? 'Ocultar itens' : 'Mostrar itens'}
        onClick={() => row.toggleExpanded()}
      >
        <ChevronDown
          className={cn('size-4 transition-transform', row.getIsExpanded() && 'rotate-180')}
        />
      </Button>
    ),
  }),
  helper.accessor('id', {
    header: 'Pedido',
    cell: ({ getValue }) => (
      <span className="font-mono text-xs" title={getValue()}>
        {shortId(getValue())}
      </span>
    ),
  }),
  helper.accessor('createdAt', {
    header: 'Criado em',
    cell: ({ getValue }) => <span className="tabular-nums">{formatDateTime(getValue())}</span>,
  }),
  helper.accessor('status', {
    header: 'Status',
    cell: ({ getValue }) => <OrderStatusBadge status={getValue()} />,
  }),
  helper.accessor((order) => order.items.reduce((sum, it) => sum + it.quantity, 0), {
    id: 'units',
    header: 'Itens',
    cell: ({ getValue, row }) => (
      <span className="tabular-nums">
        {getValue()} un. · {row.original.items.length} produto(s)
      </span>
    ),
  }),
  helper.display({
    id: 'actions',
    header: () => <span className="sr-only">Ações</span>,
    cell: ({ row }) => {
      const { paymentUrl, status } = row.original
      return (
        <div className="flex items-center justify-end gap-2">
          {paymentUrl && canPay(status) && (
            <Button size="sm" variant="outline" onClick={() => openPayment(paymentUrl)}>
              <CreditCard className="size-4" /> Pagar
            </Button>
          )}
          {canCancel(status) && <CancelOrderButton order={row.original} />}
        </div>
      )
    },
  }),
])

const EMPTY_ORDERS: Order[] = []

function CancelOrderButton({ order }: { order: Order }) {
  const clientId = useClientId()
  const cancel = useCancelOrder()

  const onConfirm = () =>
    cancel.mutate(
      { orderId: order.id, clientId },
      {
        onSuccess: () => toast.success(`Pedido ${shortId(order.id)} cancelado`),
        onError: (err) =>
          toast.error('Não foi possível cancelar o pedido', { description: err.message }),
      },
    )

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button size="sm" variant="ghost" disabled={cancel.isPending}>
          {cancel.isPending ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <X className="size-4" />
          )}
          Cancelar
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Cancelar o pedido {shortId(order.id)}?</AlertDialogTitle>
          <AlertDialogDescription>
            Os itens reservados voltam para o estoque. Essa ação não pode ser desfeita.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Voltar</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={onConfirm}>
            Cancelar pedido
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

function OrdersPage() {
  const clientId = useClientId()
  const { page, size } = Route.useSearch()
  const navigate = Route.useNavigate()

  const ordersQuery = useQuery(ordersQueryOptions(clientId, page, size))
  const productsQuery = useQuery(productsQueryOptions())

  const productNames = useMemo(
    () => new Map((productsQuery.data ?? []).map((p) => [p.id, p.name])),
    [productsQuery.data],
  )

  const total = ordersQuery.data?.total ?? 0
  const pageCount = Math.max(1, Math.ceil(total / size))
  const pagination: PaginationState = { pageIndex: page - 1, pageSize: size }

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const next = typeof updater === 'function' ? updater(pagination) : updater
    void navigate({ search: { page: next.pageIndex + 1, size: next.pageSize } })
  }

  const table = useTable({
    features,
    columns,
    data: ordersQuery.data?.items ?? EMPTY_ORDERS,
    getRowId: (row) => row.id,
    getRowCanExpand: () => true,
    manualPagination: true,
    rowCount: total,
    state: { pagination },
    onPaginationChange,
  })

  return (
    <>
      <PageHeader
        title="Pedidos"
        description={`Pedidos do usuário #${clientId}. Os status são atualizados automaticamente.`}
        actions={
          <Button
            variant="outline"
            size="sm"
            disabled={ordersQuery.isFetching}
            onClick={() => void ordersQuery.refetch()}
          >
            <RefreshCw className={cn('size-4', ordersQuery.isFetching && 'animate-spin')} />
            Atualizar
          </Button>
        }
      />

      {ordersQuery.isPending ? (
        <div className="flex flex-col gap-2">
          {Array.from({ length: 5 }, (_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      ) : ordersQuery.isError ? (
        <QueryError error={ordersQuery.error} onRetry={() => void ordersQuery.refetch()} />
      ) : (
        <>
          <DataTable
            table={table}
            emptyMessage={
              <>
                Nenhum pedido ainda.{' '}
                <Link to="/catalog" className="underline underline-offset-4">
                  Ir para o catálogo
                </Link>
              </>
            }
            renderDetail={(row) =>
              row.getIsExpanded() ? (
                <ul className="grid gap-1 text-sm sm:grid-cols-2">
                  {row.original.items.map((item) => (
                    <li key={item.productId} className="flex justify-between gap-4">
                      <span>
                        <span className="font-mono text-xs text-muted-foreground">
                          {item.productId}
                        </span>{' '}
                        {productNames.get(item.productId)}
                      </span>
                      <span className="tabular-nums">{item.quantity}x</span>
                    </li>
                  ))}
                </ul>
              ) : null
            }
          />

          <div className="mt-4 flex flex-wrap items-center justify-between gap-4 text-sm">
            <div className="flex items-center gap-2 text-muted-foreground">
              Por página
              <Select
                value={String(size)}
                onValueChange={(value) => table.setPageSize(Number(value))}
              >
                <SelectTrigger size="sm" className="w-20">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PAGE_SIZES.map((n) => (
                    <SelectItem key={n} value={String(n)}>
                      {n}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-2">
              <span className="text-muted-foreground tabular-nums">
                Página {page} de {pageCount} · {total} pedido(s)
              </span>
              <Button
                variant="outline"
                size="icon"
                aria-label="Página anterior"
                disabled={!table.getCanPreviousPage()}
                onClick={() => table.previousPage()}
              >
                <ChevronLeft className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                aria-label="Próxima página"
                disabled={!table.getCanNextPage()}
                onClick={() => table.nextPage()}
              >
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        </>
      )}
    </>
  )
}
