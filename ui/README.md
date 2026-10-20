# UI

Frontend do e-commerce em React + TanStack (Router, Query, Table, Form, Store), com Tailwind/shadcn. O toolchain é o [Vite+](https://viteplus.dev) (`vp`).

A UI se comunica **apenas com o `main`**, sempre via `/api`. Os status dos pedidos chegam em tempo real via SSE (`GET /orders/events?clientId=`).

## Desenvolvimento

```bash
bun install
bun run dev        # vp dev em http://localhost:3000, com proxy /api → http://localhost:8080 (main)
```

Para outro endereço do main: `MAIN_API_URL=http://host:porta bun run dev`.

| Script                         | O que faz                                      |
| ------------------------------ | ---------------------------------------------- |
| `bun run check`                | formatação (Oxfmt) + lint (Oxlint) + typecheck |
| `bun run fmt` / `bun run lint` | cada etapa isolada                             |
| `bun run build`                | `tsc -b` + `vp build` em `dist/`               |

## Docker

O serviço `ui` do `docker-compose.yml` gera o build e serve com nginx em http://localhost:3000. O nginx repassa `/api/*` para `main:8080`, já configurado para SSE.

## Estrutura

- `src/api/`: cliente HTTP, tipos e `queryOptions`/mutations por recurso
- `src/hooks/use-order-events.ts`: conexão SSE com reconexão e backoff
- `src/stores/`: usuário selecionado e carrinho por usuário (TanStack Store + localStorage)
- `src/routes/`: páginas (file-based routing); `routeTree.gen.ts` é gerado pelo plugin do router
