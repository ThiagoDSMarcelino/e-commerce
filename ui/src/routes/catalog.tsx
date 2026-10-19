import { useQuery } from '@tanstack/react-query'
import {
  columnFilteringFeature,
  createColumnHelper,
  createFilteredRowModel,
  createSortedRowModel,
  filterFn_includesString,
  globalFilteringFeature,
  rowSortingFeature,
  sortFn_alphanumeric,
  sortFn_text,
  tableFeatures,
  useTable,
} from '@tanstack/react-table'
import { createFileRoute } from '@tanstack/react-router'
import { Plus, Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import { toast } from 'sonner'

import { productsQueryOptions } from '@/api/products'
import type { Product } from '@/api/types'
import { CartPanel } from '@/components/cart-panel'
import { DataTable } from '@/components/data-table'
import { PageHeader, QueryError } from '@/components/page-header'
import { SortHeader } from '@/components/sort-header'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCurrency } from '@/lib/format'
import { cartActions, useCart } from '@/stores/cart'
import { useClientId } from '@/stores/user'

export const Route = createFileRoute('/catalog')({
  loader: ({ context }) => {
    // Não bloqueia a navegação: erros do catálogo são exibidos na própria página.
    void context.queryClient.prefetchQuery(productsQueryOptions())
  },
  component: CatalogPage,
})

const features = tableFeatures({
  rowSortingFeature,
  columnFilteringFeature,
  globalFilteringFeature,
  sortedRowModel: createSortedRowModel(),
  filteredRowModel: createFilteredRowModel(),
  sortFns: { alphanumeric: sortFn_alphanumeric, text: sortFn_text },
  filterFns: { includesString: filterFn_includesString },
})

const helper = createColumnHelper<typeof features, Product>()

const EMPTY_PRODUCTS: Product[] = []

// price e category ainda não são enviados pelo main; as colunas aparecem quando existirem.
function buildColumns({ price, category }: { price: boolean; category: boolean }) {
  return helper.columns([
    helper.accessor('id', {
      header: ({ column }) => <SortHeader column={column} label="Código" />,
      cell: ({ getValue }) => <span className="font-mono text-xs">{getValue()}</span>,
      sortFn: 'alphanumeric',
    }),
    helper.accessor('name', {
      header: ({ column }) => <SortHeader column={column} label="Produto" />,
      sortFn: 'text',
    }),
    ...(category
      ? [
          helper.accessor('category', {
            header: ({ column }) => <SortHeader column={column} label="Categoria" />,
            cell: ({ getValue }) => getValue() ?? '—',
          }),
        ]
      : []),
    ...(price
      ? [
          helper.accessor('price', {
            header: ({ column }) => <SortHeader column={column} label="Preço" />,
            cell: ({ getValue }) => {
              const value = getValue()
              return value === undefined ? '—' : formatCurrency(value)
            },
          }),
        ]
      : []),
    helper.accessor('amount', {
      header: ({ column }) => <SortHeader column={column} label="Disponível" />,
      cell: ({ getValue }) => <span className="tabular-nums">{getValue()}</span>,
    }),
    helper.display({
      id: 'actions',
      header: () => <span className="sr-only">Adicionar</span>,
      cell: ({ row }) => <AddToCart product={row.original} />,
    }),
  ])
}

function CatalogPage() {
  const clientId = useClientId()
  const productsQuery = useQuery(productsQueryOptions())
  const [globalFilter, setGlobalFilter] = useState('')

  const products = productsQuery.data ?? EMPTY_PRODUCTS
  const hasPrice = products.some((p) => p.price !== undefined)
  const hasCategory = products.some((p) => p.category !== undefined)
  const columns = useMemo(
    () => buildColumns({ price: hasPrice, category: hasCategory }),
    [hasPrice, hasCategory],
  )
  const stock = useMemo(() => new Map(products.map((p) => [p.id, p])), [products])

  const table = useTable({
    features,
    columns,
    data: products,
    getRowId: (row) => row.id,
    globalFilterFn: 'includesString',
    getColumnCanGlobalFilter: (column) =>
      column.id === 'id' || column.id === 'name' || column.id === 'category',
    state: { globalFilter },
    onGlobalFilterChange: (updater) =>
      setGlobalFilter((prev) =>
        String(typeof updater === 'function' ? updater(prev) : (updater ?? '')),
      ),
  })

  return (
    <>
      <PageHeader
        title="Catálogo"
        description="Produtos disponíveis no estoque, consultados via main."
        actions={
          <div className="relative w-full sm:w-72">
            <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Buscar por código ou nome"
              className="pl-8"
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
            />
          </div>
        }
      />

      <div className="grid gap-6 lg:grid-cols-[1fr_22rem]">
        <div>
          {productsQuery.isPending ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 6 }, (_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : productsQuery.isError ? (
            <QueryError error={productsQuery.error} onRetry={() => void productsQuery.refetch()} />
          ) : (
            <DataTable table={table} emptyMessage="Nenhum produto disponível." />
          )}
        </div>

        <CartPanel clientId={clientId} stock={stock} />
      </div>
    </>
  )
}

function AddToCart({ product }: { product: Product }) {
  const clientId = useClientId()
  const cart = useCart(clientId)
  const [quantity, setQuantity] = useState(1)

  const inCart = cart.find((it) => it.productId === product.id)?.quantity ?? 0
  const remaining = product.amount - inCart
  const valid = Number.isInteger(quantity) && quantity > 0 && quantity <= remaining

  const add = () => {
    cartActions.add(clientId, product, quantity)
    toast.success(`${quantity}x ${product.name} adicionado ao carrinho`)
    setQuantity(1)
  }

  return (
    <div className="flex items-center justify-end gap-2">
      <Input
        type="number"
        min={1}
        max={Math.max(remaining, 1)}
        className="h-8 w-16"
        aria-label={`Quantidade de ${product.name}`}
        value={quantity}
        disabled={remaining <= 0}
        onChange={(e) => setQuantity(Math.floor(Number(e.target.value)))}
      />
      <Button size="sm" variant="secondary" disabled={!valid} onClick={add}>
        <Plus className="size-4" />
        <span className="hidden md:inline">Adicionar</span>
      </Button>
    </div>
  )
}
