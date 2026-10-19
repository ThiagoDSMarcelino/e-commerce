import { UserRound } from 'lucide-react'
import { useId, useState } from 'react'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { setClientId, useClientId } from '@/stores/user'

export function UserSelector() {
  const clientId = useClientId()
  const id = useId()

  return (
    <div className="flex items-center gap-2">
      <Label htmlFor={id} className="flex items-center gap-1 text-sm text-muted-foreground">
        <UserRound className="size-4" />
        <span className="hidden sm:inline">Usuário</span>
      </Label>
      {/* A key reinicia o rascunho sempre que o usuário selecionado mudar. */}
      <ClientIdInput key={clientId} id={id} clientId={clientId} />
    </div>
  )
}

function ClientIdInput({ id, clientId }: { id: string; clientId: number }) {
  const [draft, setDraft] = useState(String(clientId))

  const commit = () => {
    const value = Number(draft)
    if (Number.isInteger(value) && value > 0) setClientId(value)
    else setDraft(String(clientId))
  }

  return (
    <Input
      id={id}
      type="number"
      min={1}
      step={1}
      inputMode="numeric"
      className="h-8 w-20"
      value={draft}
      onChange={(e) => setDraft(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') commit()
      }}
    />
  )
}
