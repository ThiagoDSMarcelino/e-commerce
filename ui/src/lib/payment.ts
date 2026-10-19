import { toast } from 'sonner'

/**
 * Abre o mock de pagamento (servido pelo main) em uma nova aba.
 * Se o navegador bloquear o popup, oferece um botão para o consumidor abrir manualmente.
 */
export function openPayment(paymentUrl: string) {
  // Sem 'noopener' nas features para conseguir detectar bloqueio (com ele, window.open sempre retorna null).
  const win = window.open(paymentUrl, '_blank')
  if (win) {
    win.opener = null
    return
  }

  toast.warning('O navegador bloqueou a aba de pagamento', {
    action: { label: 'Abrir pagamento', onClick: () => window.open(paymentUrl, '_blank') },
    duration: 15_000,
  })
}
