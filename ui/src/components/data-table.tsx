import {
  FlexRender,
  type ReactTable,
  type Row,
  type RowData,
  type TableFeatures,
} from '@tanstack/react-table'
import { Fragment, type ReactNode } from 'react'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

type DataTableProps<TFeatures extends TableFeatures, TData extends RowData> = {
  // oxlint-disable-next-line typescript/no-explicit-any -- o seletor de estado varia por tabela
  table: ReactTable<TFeatures, TData, any>
  emptyMessage?: ReactNode
  // Conteúdo extra abaixo de uma linha (ex.: detalhes de uma linha expandida).
  renderDetail?: (row: Row<TFeatures, TData>) => ReactNode
}

// Renderização padrão (shadcn) de uma tabela TanStack. Features e colunas ficam em cada página.
export function DataTable<TFeatures extends TableFeatures, TData extends RowData>({
  table,
  emptyMessage = 'Nenhum registro encontrado.',
  renderDetail,
}: DataTableProps<TFeatures, TData>) {
  const rows = table.getRowModel().rows
  const columnCount = table.getAllLeafColumns().length

  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((group) => (
            <TableRow key={group.id}>
              {group.headers.map((header) => (
                <TableHead key={header.id}>
                  {header.isPlaceholder ? null : <FlexRender header={header} />}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={columnCount} className="h-24 text-center text-muted-foreground">
                {emptyMessage}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => {
              const detail = renderDetail?.(row)
              return (
                <Fragment key={row.id}>
                  <TableRow>
                    {row.getAllCells().map((cell) => (
                      <TableCell key={cell.id}>
                        <FlexRender cell={cell} />
                      </TableCell>
                    ))}
                  </TableRow>
                  {detail ? (
                    <TableRow className="bg-muted/30 hover:bg-muted/30">
                      <TableCell colSpan={columnCount}>{detail}</TableCell>
                    </TableRow>
                  ) : null}
                </Fragment>
              )
            })
          )}
        </TableBody>
      </Table>
    </div>
  )
}
