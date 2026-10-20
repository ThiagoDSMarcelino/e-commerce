package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:embed checkout.html
var checkoutHTML string

type Handlers struct {
	publicURL string
	charges   *ChargeStore
	http      *http.Client
	page      *template.Template
}

type chargeRequest struct {
	OrderId    string `json:"orderId"`
	WebhookUrl string `json:"webhookUrl"`
}

type chargeResponse struct {
	CheckoutUrl string `json:"checkoutUrl"`
}

type webhookRequest struct {
	OrderId string `json:"orderId"`
	Status  string `json:"status"`
}

type checkoutPage struct {
	OrderId string
	Found   bool
	Charge  Charge
}

func NewHandlers(settings *Settings, charges *ChargeStore) *Handlers {
	return &Handlers{
		publicURL: settings.PublicURL,
		charges:   charges,
		http:      &http.Client{Timeout: 10 * time.Second},
		page:      template.Must(template.New("checkout").Parse(checkoutHTML)),
	}
}

func (h *Handlers) CreateCharge(w http.ResponseWriter, r *http.Request) {
	var req chargeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.OrderId == "" {
		writeError(w, http.StatusBadRequest, "orderId is required")
		return
	}

	if u, err := url.Parse(req.WebhookUrl); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeError(w, http.StatusBadRequest, "webhookUrl must be an http(s) url")
		return
	}

	if !h.charges.Create(req.OrderId, req.WebhookUrl) {
		writeError(w, http.StatusConflict, "charge for this order was already processed")
		return
	}

	checkoutUrl := h.publicURL + "/checkout/" + url.PathEscape(req.OrderId)

	slog.Info("Cobranca criada", "order", req.OrderId, "checkoutUrl", checkoutUrl, "webhookUrl", req.WebhookUrl)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(chargeResponse{CheckoutUrl: checkoutUrl})
}

func (h *Handlers) ShowCheckout(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("orderId")

	charge, ok := h.charges.Get(orderId)

	status := http.StatusOK
	if !ok {
		status = http.StatusNotFound
	}

	h.render(w, status, checkoutPage{OrderId: orderId, Found: ok, Charge: charge})
}

func (h *Handlers) Pay(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("orderId")
	checkoutPath := "/checkout/" + url.PathEscape(orderId)

	status := ChargeStatus(r.FormValue("status"))
	if status != StatusApproved && status != StatusDeclined {
		http.Error(w, "status must be APROVADO or RECUSADO", http.StatusBadRequest)
		return
	}

	charge, ok := h.charges.StartProcessing(orderId)
	if !ok {
		http.Redirect(w, r, checkoutPath, http.StatusSeeOther)
		return
	}

	final, errMsg := h.notify(context.WithoutCancel(r.Context()), charge, status)
	h.charges.Finish(orderId, final, errMsg)

	http.Redirect(w, r, checkoutPath, http.StatusSeeOther)
}

func (h *Handlers) notify(ctx context.Context, charge Charge, status ChargeStatus) (ChargeStatus, string) {
	body, err := json.Marshal(webhookRequest{OrderId: charge.OrderId, Status: string(status)})
	if err != nil {
		return StatusPending, "Falha ao montar o aviso para o MS Pagamento. Tente novamente."
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, charge.WebhookUrl, bytes.NewReader(body))
	if err != nil {
		return StatusFailed, fmt.Sprintf("Webhook invalido: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.http.Do(req)
	if err != nil {
		slog.Error("Falha ao chamar o webhook", "order", charge.OrderId, "error", err)
		return StatusPending, "Nao foi possivel falar com o MS Pagamento. Tente novamente."
	}
	defer res.Body.Close()

	reason, _ := io.ReadAll(io.LimitReader(res.Body, 512))

	slog.Info("Webhook chamado", "order", charge.OrderId, "status", status, "response", res.StatusCode)

	switch {
	case res.StatusCode >= 200 && res.StatusCode <= 299:
		return status, ""
	case res.StatusCode >= 400 && res.StatusCode <= 499:
		return StatusFailed, fmt.Sprintf("O MS Pagamento recusou o aviso (%d): %s", res.StatusCode, strings.TrimSpace(string(reason)))
	default:
		return StatusPending, fmt.Sprintf("O MS Pagamento respondeu %d. Tente novamente.", res.StatusCode)
	}
}

func (h *Handlers) render(w http.ResponseWriter, status int, data checkoutPage) {
	var buf bytes.Buffer
	if err := h.page.Execute(&buf, data); err != nil {
		slog.Error("Falha ao montar a pagina", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
