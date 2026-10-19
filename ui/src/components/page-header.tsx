import type { ReactNode } from 'react'

export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string
  description?: ReactNode
  actions?: ReactNode
}) {
  return (
    <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
      </div>
      {actions}
    </div>
  )
}

export function QueryError({ error, onRetry }: { error: Error; onRetry?: () => void }) {
  return (
    <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm">
      <p className="font-medium text-destructive">Falha ao consultar o main</p>
      <p className="mt-1 text-muted-foreground">{error.message}</p>
      {onRetry && (
        <button
          type="button"
          className="mt-2 text-sm underline underline-offset-4"
          onClick={onRetry}
        >
          Tentar novamente
        </button>
      )}
    </div>
  )
}
