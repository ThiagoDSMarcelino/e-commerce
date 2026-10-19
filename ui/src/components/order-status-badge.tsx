import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { type StatusTone, statusMeta } from '@/lib/order-status'

const TONE_CLASSES: Record<StatusTone, string> = {
  neutral: 'bg-muted text-muted-foreground',
  info: 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-200',
  success: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200',
  danger: 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-200',
}

export function OrderStatusBadge({ status }: { status: string }) {
  const { tone, final } = statusMeta(status)

  return (
    <Badge variant="outline" className={cn('border-transparent', TONE_CLASSES[tone])}>
      {!final && <span className="size-1.5 animate-pulse rounded-full bg-current" />}
      {status}
    </Badge>
  )
}
