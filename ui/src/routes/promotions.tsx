import { useForm } from '@tanstack/react-form'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2, MailCheck, X } from 'lucide-react'
import { toast } from 'sonner'
import { z } from 'zod'

import {
  registrationsQueryOptions,
  useCancelPromotion,
  useRegisterPromotion,
} from '@/api/promotions'
import { PageHeader, QueryError } from '@/components/page-header'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { PROMOTION_CATEGORIES } from '@/lib/categories'
import { formatDateTime } from '@/lib/format'
import { useClientId } from '@/stores/user'

export const Route = createFileRoute('/promotions')({
  component: PromotionsPage,
})

const emailSchema = z.email('Informe um e-mail válido')
const categoriesSchema = z.array(z.string()).min(1, 'Selecione ao menos uma categoria')

function errorMessages(errors: unknown[]) {
  return errors
    .map((err) =>
      typeof err === 'string' ? err : (err as { message?: string } | undefined)?.message,
    )
    .filter(Boolean)
    .join(', ')
}

function PromotionsPage() {
  const clientId = useClientId()

  return (
    <>
      <PageHeader
        title="Promoções"
        description="Cadastre seu e-mail e escolha as categorias de interesse para receber ofertas."
      />

      <div className="grid gap-6 lg:grid-cols-2">
        <SubscribeForm clientId={clientId} />
        <Registrations clientId={clientId} />
      </div>
    </>
  )
}

function SubscribeForm({ clientId }: { clientId: number }) {
  const register = useRegisterPromotion()

  const form = useForm({
    defaultValues: { email: '', categories: [] as string[] },
    onSubmit: async ({ value, formApi }) => {
      try {
        await register.mutateAsync({
          clientId,
          email: value.email.trim(),
          categories: value.categories,
        })
        toast.success('Inscrição enviada', {
          description: `Ofertas serão enviadas para ${value.email}.`,
        })
        formApi.reset()
      } catch (err) {
        toast.error('Não foi possível concluir a inscrição', {
          description: (err as Error).message,
        })
      }
    },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Nova inscrição</CardTitle>
        <CardDescription>Usuário #{clientId}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          className="flex flex-col gap-5"
          noValidate
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            void form.handleSubmit()
          }}
        >
          <form.Field name="email" validators={{ onBlur: emailSchema, onSubmit: emailSchema }}>
            {(field) => (
              <div className="flex flex-col gap-2">
                <Label htmlFor={field.name}>E-mail</Label>
                <Input
                  id={field.name}
                  type="email"
                  autoComplete="email"
                  placeholder="voce@exemplo.com"
                  value={field.state.value}
                  aria-invalid={field.state.meta.errors.length > 0}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
                {field.state.meta.errors.length > 0 && (
                  <p className="text-xs text-destructive">
                    {errorMessages(field.state.meta.errors)}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          <form.Field
            name="categories"
            mode="array"
            validators={{ onChange: categoriesSchema, onSubmit: categoriesSchema }}
          >
            {(field) => (
              <fieldset className="flex flex-col gap-2">
                <legend className="mb-2 text-sm font-medium">Categorias de interesse</legend>
                <div className="grid gap-2 sm:grid-cols-2">
                  {PROMOTION_CATEGORIES.map((category) => {
                    const id = `category-${category.id}`
                    const checked = field.state.value.includes(category.id)
                    return (
                      <div key={category.id} className="flex items-center gap-2">
                        <Checkbox
                          id={id}
                          checked={checked}
                          onCheckedChange={(next) =>
                            field.handleChange(
                              next === true
                                ? [...field.state.value, category.id]
                                : field.state.value.filter((c) => c !== category.id),
                            )
                          }
                        />
                        <Label htmlFor={id} className="font-normal">
                          {category.label}
                        </Label>
                      </div>
                    )
                  })}
                </div>
                {field.state.meta.errors.length > 0 && (
                  <p className="text-xs text-destructive">
                    {errorMessages(field.state.meta.errors)}
                  </p>
                )}
              </fieldset>
            )}
          </form.Field>

          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <Button type="submit" disabled={isSubmitting} className="self-start">
                {isSubmitting ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <MailCheck className="size-4" />
                )}
                Inscrever-se
              </Button>
            )}
          </form.Subscribe>
        </form>
      </CardContent>
    </Card>
  )
}

function Registrations({ clientId }: { clientId: number }) {
  const registrationsQuery = useQuery(registrationsQueryOptions(clientId))
  const cancel = useCancelPromotion()

  const onCancel = (email: string) =>
    cancel.mutate(
      { clientId, email },
      {
        onSuccess: () => toast.success(`Inscrição de ${email} cancelada`),
        onError: (err) => toast.error('Não foi possível cancelar', { description: err.message }),
      },
    )

  return (
    <Card>
      <CardHeader>
        <CardTitle>Minhas inscrições</CardTitle>
        <CardDescription>E-mails cadastrados para receber ofertas.</CardDescription>
      </CardHeader>
      <CardContent>
        {registrationsQuery.isPending ? (
          <div className="flex flex-col gap-2">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : registrationsQuery.isError ? (
          <QueryError
            error={registrationsQuery.error}
            onRetry={() => void registrationsQuery.refetch()}
          />
        ) : registrationsQuery.data.items.length === 0 ? (
          <p className="text-sm text-muted-foreground">Nenhuma inscrição registrada.</p>
        ) : (
          <ul className="flex flex-col divide-y">
            {registrationsQuery.data.items.map((registration) => (
              <li key={registration.email} className="flex items-center gap-3 py-2">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{registration.email}</p>
                  <p className="text-xs text-muted-foreground">
                    Desde {formatDateTime(registration.createdAt)}
                    {registration.categories?.length
                      ? ` · ${registration.categories.join(', ')}`
                      : ''}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={cancel.isPending && cancel.variables?.email === registration.email}
                  onClick={() => onCancel(registration.email)}
                >
                  <X className="size-4" /> Cancelar
                </Button>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
