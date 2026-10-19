# Contrato UI ↔ main

A UI conversa **somente** com o `main` (API gateway). Todas as chamadas usam o prefixo `/api`, que é repassado ao `main` pelo proxy do Vite (`vp dev`) ou pelo nginx (Docker). O `main` é o único que fala com os outros serviços: via RabbitMQ e, no caso do catálogo, via HTTP com o storage.

```
ui ──HTTP/SSE──▶ main ──RabbitMQ──▶ storage / payments / shipping / promotions
                  └──HTTP GET /products──▶ storage
```

Legenda: ✅ implementado no main · 🚧 esperado pela UI e ainda não implementado

Erros seguem o formato `{"error": "mensagem"}`. A UI mostra essa mensagem ao usuário.

---

## Produtos

### ✅ `GET /products`

Produtos com estoque disponível. O main busca no storage e responde:

```json
[{ "id": "PAP-001", "name": "Caderno universitario 96 folhas", "amount": 120 }]
```

Campos opcionais que a UI já sabe exibir, se forem adicionados: `price` (number, em BRL) e `category` (string).

## Pedidos

### ✅ `GET /orders?clientId={id}&page={n}&size={n}`

```json
{
  "items": [
    {
      "id": "0199…",
      "clientId": 1,
      "createdAt": "2026-10-19T12:00:00Z",
      "status": "Estoque confirmado",
      "items": [{ "productId": "PAP-001", "quantity": 2 }],
      "paymentUrl": "🚧 opcional"
    }
  ],
  "total": 12,
  "page": 1
}
```

Os valores de `status` vêm de `main/internal/order/status.go`, espelhado em `ui/src/lib/order-status.ts`.

### ✅ `POST /orders`

Requisição:

```json
{ "clientId": 1, "items": [{ "productId": "PAP-001", "quantity": 2 }] }
```

Resposta `201`: o pedido criado, no mesmo formato de um item da listagem.

🚧 **Pagamento**: se a resposta trouxer `paymentUrl`, a UI abre essa URL em uma nova aba logo após o clique em "Finalizar pedido". Se o navegador bloquear a aba, a UI mostra um botão para abrir manualmente. A URL precisa ser servida pelo **main**, por exemplo `/api/payments/{orderId}`, que exibe o mock com os botões aprovar/recusar e publica `pagamento.aprovado` ou `pagamento.recusado` no RabbitMQ. O `paymentUrl` também pode vir na listagem ou no evento SSE. Nesses casos, a UI mostra o botão "Pagar" enquanto o status for `Estoque confirmado`.

### ✅ `POST /orders/{id}/cancel`

Requisição:

```json
{ "clientId": 1 }
```

Resposta `200`: o pedido com `status: "Cancelado"`, no mesmo formato de um item da listagem. O main publica `pedido.excluido` para o storage devolver o estoque reservado.

Pode cancelar em qualquer status, exceto `Enviado` e `Cancelado`. Erros: `400` (id ou corpo inválido), `404` (pedido não existe para esse `clientId`) e `409` (status não permite cancelamento).

### 🚧 `GET /orders/events?clientId={id}` (SSE)

`Content-Type: text/event-stream`. Cada mudança de status de um pedido do cliente é enviada como:

```
event: order-status
data: {"orderId":"0199…","status":"Pagamento aprovado","updatedAt":"2026-10-19T12:00:05Z"}

```

- `paymentUrl` pode vir no payload. A UI também aceita eventos sem nome (`message`) com o mesmo payload.
- Comentários de keep-alive (`: ping`) são recomendados a cada ~15 s.
- Na conexão e em cada evento, a UI atualiza o cache, mostra uma notificação e refaz a listagem.
- Enquanto o endpoint não existir, a UI tenta reconectar com backoff (2 s → 60 s) e usa polling de 5 s na página de pedidos. A atualização dos status já funciona assim hoje.

O que o main vai precisar para implementar:

- Despachar o evento para os clientes conectados no `HandleStatusEvent` (`internal/order/service.go`), depois de atualizar o banco. Isso exige descobrir o `clientId` do pedido.
- Nessa rota, contornar o middleware `timeout(60s)` e o `WriteTimeout: 30s` do `http.Server` (`cmd/api.go`), por exemplo com `http.NewResponseController(w).SetWriteDeadline(time.Time{})`.
- Fazer `Flush()` a cada evento e encerrar quando `c.Request.Context()` terminar.
- O nginx da UI já está configurado para SSE (`proxy_buffering off`, `proxy_read_timeout 1h`).

## Promoções

### ✅ `GET /promotions?clientId={id}&page={n}&size={n}`

```json
{ "items": [{ "email": "voce@exemplo.com", "createdAt": "…", "categories": ["🚧 opcional"] }], "total": 1, "page": 1 }
```

> Hoje nenhum serviço grava em `promotion_registrations`, então a lista vem vazia mesmo depois de uma inscrição.

### ✅ `POST /promotions` → `204`

```json
{ "clientId": 1, "email": "voce@exemplo.com", "categories": ["papelaria", "escolar"] }
```

As categorias vêm de uma lista fixa em `ui/src/lib/categories.ts`. Se o main passar a expor um `GET /categories`, basta trocar essa lista por uma query.

### ✅ `DELETE /promotions` → `204`

```json
{ "clientId": 1, "email": "voce@exemplo.com" }
```
