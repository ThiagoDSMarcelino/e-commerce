import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { type SseStatus, useSseStatus } from '@/hooks/use-order-events'
import { cn } from '@/lib/utils'

const LABELS: Record<SseStatus, { label: string; hint: string; dot: string }> = {
  open: {
    label: 'Tempo real',
    hint: 'Conectado ao main via SSE: os status dos pedidos chegam em tempo real.',
    dot: 'bg-emerald-500',
  },
  connecting: {
    label: 'Conectando…',
    hint: 'Abrindo conexão SSE com o main.',
    dot: 'bg-amber-500 animate-pulse',
  },
  unavailable: {
    label: 'Desconectado',
    hint: 'SSE indisponível no main. Tentando reconectar; os status não serão atualizados até lá.',
    dot: 'bg-muted-foreground',
  },
}

export function SseIndicator() {
  const status = useSseStatus()
  const { label, hint, dot } = LABELS[status]

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex items-center gap-2 text-xs text-muted-foreground">
          <span className={cn('size-2 rounded-full', dot)} />
          <span className="hidden sm:inline">{label}</span>
        </span>
      </TooltipTrigger>
      <TooltipContent>{hint}</TooltipContent>
    </Tooltip>
  )
}
