import type { QueryClient } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { createRootRouteWithContext, Link, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'
import { Store } from 'lucide-react'

import { SseIndicator } from '@/components/sse-indicator'
import { Badge } from '@/components/ui/badge'
import { Toaster } from '@/components/ui/sonner'
import { TooltipProvider } from '@/components/ui/tooltip'
import { UserSelector } from '@/components/user-selector'
import { useOrderEvents } from '@/hooks/use-order-events'
import { useCartCount } from '@/stores/cart'
import { useClientId } from '@/stores/user'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: RootLayout,
  notFoundComponent: () => <p className="text-muted-foreground">Página não encontrada.</p>,
})

const navLinkClass =
  'rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground'

function RootLayout() {
  const clientId = useClientId()
  const cartCount = useCartCount(clientId)

  // Uma única conexão SSE por usuário selecionado, ativa em todas as páginas.
  useOrderEvents(clientId)

  return (
    <TooltipProvider>
      <div className="min-h-svh bg-background">
        <header className="sticky top-0 z-10 border-b bg-background/90 backdrop-blur">
          <div className="mx-auto flex h-14 max-w-6xl items-center gap-4 px-4">
            <Link to="/catalog" className="flex items-center gap-2 font-semibold">
              <Store className="size-5" />
              <span className="hidden sm:inline">E-commerce</span>
            </Link>

            <nav className="flex items-center gap-1">
              <Link
                to="/catalog"
                className={navLinkClass}
                activeProps={{ className: 'bg-muted text-foreground' }}
              >
                Catálogo
                {cartCount > 0 && (
                  <Badge className="ml-1.5 h-5 min-w-5 px-1.5 tabular-nums">{cartCount}</Badge>
                )}
              </Link>
              <Link
                to="/orders"
                className={navLinkClass}
                activeProps={{ className: 'bg-muted text-foreground' }}
              >
                Pedidos
              </Link>
              <Link
                to="/promotions"
                className={navLinkClass}
                activeProps={{ className: 'bg-muted text-foreground' }}
              >
                Promoções
              </Link>
            </nav>

            <div className="ml-auto flex items-center gap-4">
              <SseIndicator />
              <UserSelector />
            </div>
          </div>
        </header>

        <main className="mx-auto max-w-6xl px-4 py-6">
          <Outlet />
        </main>
      </div>

      <Toaster richColors position="bottom-right" />
      {import.meta.env.DEV && (
        <>
          <ReactQueryDevtools buttonPosition="bottom-left" />
          <TanStackRouterDevtools position="bottom-left" />
        </>
      )}
    </TooltipProvider>
  )
}
