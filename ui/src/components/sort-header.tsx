import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-react'

import { Button } from '@/components/ui/button'

type SortableColumn = {
  getIsSorted: () => false | 'asc' | 'desc'
  toggleSorting: (desc?: boolean) => void
}

export function SortHeader({ column, label }: { column: SortableColumn; label: string }) {
  const sorted = column.getIsSorted()
  const Icon = sorted === 'asc' ? ArrowUp : sorted === 'desc' ? ArrowDown : ArrowUpDown

  return (
    <Button
      variant="ghost"
      size="sm"
      className="-ml-2"
      onClick={() => column.toggleSorting(sorted === 'asc')}
    >
      {label}
      <Icon className="size-3.5 opacity-60" />
    </Button>
  )
}
