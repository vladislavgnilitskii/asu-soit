import { useState, type FormEvent } from "react"
import { useParams, Link } from "react-router-dom"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { formatDate, formatMoney } from "@/lib/format"
import type {
  Device,
  EngineerListItem,
  Invoice,
  RepairRequest,
  RequestPartEntry,
  RequestStatus,
  SparePart,
  StatusHistoryEntry,
  UpdateRequestDTO,
} from "@/lib/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { RequestStatusSelect } from "@/components/forms/RequestStatusSelect"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

const INVOICE_STATUS_LABEL: Record<Invoice["status"], string> = {
  pending: "Ожидает оплаты",
  paid: "Оплачен",
  cancelled: "Отменён",
}

export function RequestDetailPage() {
  const { id = "" } = useParams()
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const isAdmin = user?.role === "admin"
  const isManager = user?.role === "manager"
  const isEngineer = user?.role === "engineer"
  const isAccountant = user?.role === "accountant"
  const isStorekeeper = user?.role === "storekeeper"

  const canManage = isAdmin || isManager // назначение, закрытие
  const canEditDetails = isAdmin || isManager || isEngineer
  const canSeeDevice = isAdmin || isManager || isEngineer
  const canSeeHistory = isAdmin || isManager || isEngineer
  const canSeeInvoice = isAdmin || isManager || isAccountant
  const canManageInvoice = isAdmin || isAccountant
  const canIssueParts = isAdmin || isStorekeeper || isEngineer

  const request = useQuery({
    queryKey: ["requests", id],
    queryFn: () => api.get<RepairRequest>(`/requests/${id}`),
  })
  const statuses = useQuery({
    queryKey: ["request-statuses"],
    queryFn: () => api.get<RequestStatus[]>("/request-statuses"),
  })
  const device = useQuery({
    queryKey: ["devices", request.data?.device_id],
    queryFn: () => api.get<Device>(`/devices/${request.data!.device_id}`),
    enabled: canSeeDevice && !!request.data?.device_id,
  })
  const history = useQuery({
    queryKey: ["requests", id, "history"],
    queryFn: () => api.get<StatusHistoryEntry[]>(`/requests/${id}/history`),
    enabled: canSeeHistory,
  })
  const parts = useQuery({
    queryKey: ["requests", id, "parts"],
    queryFn: () => api.get<RequestPartEntry[]>(`/requests/${id}/parts`),
  })
  // 404 = счёта ещё нет; это не ошибка
  const invoice = useQuery({
    queryKey: ["requests", id, "invoice"],
    queryFn: async () => {
      try {
        return await api.get<Invoice>(`/requests/${id}/invoice`)
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) return null
        throw err
      }
    },
    enabled: canSeeInvoice,
  })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["requests"] })
  }

  if (request.isLoading)
    return <p className="text-sm text-muted-foreground">Загрузка…</p>
  if (request.isError)
    return (
      <p className="text-sm text-destructive">
        Не удалось загрузить заявку: {(request.error as Error).message}
      </p>
    )
  if (!request.data) return null

  const r = request.data
  const isClosed = !!r.closed_at

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <Link
            to="/requests"
            className="text-sm text-muted-foreground hover:underline"
          >
            ← Все заявки
          </Link>
          <h1 className="text-2xl font-semibold mt-1">
            Заявка от {formatDate(r.created_at)}
          </h1>
          <p className="text-sm text-muted-foreground max-w-2xl">
            {r.problem_description}
          </p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {statuses.data &&
            (canEditDetails && !isClosed ? (
              <RequestStatusSelect
                requestId={r.id}
                statusId={r.status_id}
                statuses={statuses.data}
              />
            ) : (
              <Badge variant="secondary">
                {statuses.data.find((s) => s.id === r.status_id)?.name ?? "—"}
              </Badge>
            ))}
          {canManage && !isClosed && <CloseRequestDialog requestId={r.id} />}
          {isClosed && <Badge>Закрыта {formatDate(r.closed_at)}</Badge>}
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm text-muted-foreground">
              Устройство
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm space-y-1">
            {!canSeeDevice && <p className="text-muted-foreground">Нет доступа</p>}
            {device.data && (
              <>
                <p className="font-medium">
                  {device.data.brand} {device.data.model}
                </p>
                <p className="text-muted-foreground">
                  С/Н: {device.data.serial_number || "—"}
                </p>
                {device.data.appearance_note && (
                  <p className="text-muted-foreground">
                    {device.data.appearance_note}
                  </p>
                )}
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm text-muted-foreground">
              Сроки и стоимость
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm space-y-1">
            <p>Создана: {formatDate(r.created_at)}</p>
            <p>Плановый срок: {formatDate(r.planned_deadline)}</p>
            <p>Оценка: {formatMoney(r.estimated_cost)}</p>
            <p>Итог: {formatMoney(r.final_cost)}</p>
          </CardContent>
        </Card>

        <AssignCard
          requestId={r.id}
          assignedTo={r.assigned_to ?? null}
          canManage={canManage && !isClosed}
          onChanged={invalidate}
        />
      </div>

      {canEditDetails && !isClosed && (
        <DetailsForm request={r} onSaved={invalidate} />
      )}

      <PartsSection
        requestId={r.id}
        parts={parts.data ?? []}
        canIssue={canIssueParts && !isClosed}
      />

      {canSeeHistory && history.data && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">История статусов</CardTitle>
          </CardHeader>
          <CardContent>
            {history.data.length === 0 ? (
              <p className="text-sm text-muted-foreground">Записей нет</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Когда</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Кто</TableHead>
                    <TableHead>Комментарий</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {history.data.map((h) => (
                    <TableRow key={h.id}>
                      <TableCell>
                        {new Date(h.changed_at).toLocaleString("ru-RU")}
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{h.status_name}</Badge>
                      </TableCell>
                      <TableCell>{h.changed_by_name}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {h.comment || "—"}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      {canSeeInvoice && (
        <InvoiceSection
          requestId={r.id}
          invoice={invoice.data ?? null}
          isClosed={isClosed}
          canManage={canManageInvoice}
        />
      )}
    </div>
  )
}

// ── Назначение инженера ─────────────────────────────────────────────────────

function AssignCard({
  requestId,
  assignedTo,
  canManage,
  onChanged,
}: {
  requestId: string
  assignedTo: string | null
  canManage: boolean
  onChanged: () => void
}) {
  const engineers = useQuery({
    queryKey: ["engineers"],
    queryFn: () => api.get<EngineerListItem[]>("/engineers"),
    enabled: canManage,
  })

  const mutation = useMutation({
    mutationFn: (engineerId: string) =>
      api.patch(`/requests/${requestId}/assign`, { assigned_to: engineerId }),
    onSuccess: () => {
      onChanged()
      toast.success("Исполнитель назначен")
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  const items = (engineers.data ?? []).map((e) => ({
    value: e.id,
    label: `${e.last_name} ${e.first_name}`,
  }))
  const current = items.find((i) => i.value === assignedTo)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">
          Исполнитель
        </CardTitle>
      </CardHeader>
      <CardContent className="text-sm space-y-2">
        {canManage ? (
          <Select
            items={items}
            value={assignedTo}
            onValueChange={(v) => v && mutation.mutate(v)}
            disabled={mutation.isPending}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Назначить инженера" />
            </SelectTrigger>
            <SelectContent>
              {items.map((e) => (
                <SelectItem key={e.value} value={e.value}>
                  {e.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <p>{current?.label ?? (assignedTo ? "Назначен" : "Не назначен")}</p>
        )}
      </CardContent>
    </Card>
  )
}

// ── Диагностика и стоимость ────────────────────────────────────────────────

function DetailsForm({
  request,
  onSaved,
}: {
  request: RepairRequest
  onSaved: () => void
}) {
  const [diagnostic, setDiagnostic] = useState(request.diagnostic_result ?? "")
  const [estimated, setEstimated] = useState(
    request.estimated_cost?.toString() ?? "",
  )
  const [final, setFinal] = useState(request.final_cost?.toString() ?? "")

  const mutation = useMutation({
    mutationFn: (dto: UpdateRequestDTO) =>
      api.patch<RepairRequest>(`/requests/${request.id}`, dto),
    onSuccess: () => {
      onSaved()
      toast.success("Сохранено")
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    // отправляем только заполненные поля — COALESCE на бэке сохранит остальное
    const dto: UpdateRequestDTO = {}
    if (diagnostic) dto.diagnostic_result = diagnostic
    if (estimated) dto.estimated_cost = Number(estimated)
    if (final) dto.final_cost = Number(final)
    mutation.mutate(dto)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Диагностика и стоимость</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="diag">Результат диагностики</Label>
            <Textarea
              id="diag"
              value={diagnostic}
              onChange={(e) => setDiagnostic(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-3 max-w-md">
            <div className="space-y-2">
              <Label htmlFor="est">Оценка, ₽</Label>
              <Input
                id="est"
                type="number"
                min="0"
                step="0.01"
                value={estimated}
                onChange={(e) => setEstimated(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="fin">Итог, ₽</Label>
              <Input
                id="fin"
                type="number"
                min="0"
                step="0.01"
                value={final}
                onChange={(e) => setFinal(e.target.value)}
              />
            </div>
          </div>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? "Сохранение…" : "Сохранить"}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}

// ── Запчасти ────────────────────────────────────────────────────────────────

function PartsSection({
  requestId,
  parts,
  canIssue,
}: {
  requestId: string
  parts: RequestPartEntry[]
  canIssue: boolean
}) {
  const total = parts.reduce((sum, p) => sum + p.quantity * p.unit_price, 0)

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle className="text-base">Запчасти</CardTitle>
        {canIssue && <IssuePartDialog requestId={requestId} />}
      </CardHeader>
      <CardContent>
        {parts.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Запчасти не использовались
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Деталь</TableHead>
                <TableHead className="text-right">Кол-во</TableHead>
                <TableHead className="text-right">Цена</TableHead>
                <TableHead className="text-right">Сумма</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {parts.map((p) => (
                <TableRow key={p.id}>
                  <TableCell>{p.part_name}</TableCell>
                  <TableCell className="text-right">{p.quantity}</TableCell>
                  <TableCell className="text-right">
                    {formatMoney(p.unit_price)}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatMoney(p.quantity * p.unit_price)}
                  </TableCell>
                </TableRow>
              ))}
              <TableRow>
                <TableCell colSpan={3} className="font-medium">
                  Итого по деталям
                </TableCell>
                <TableCell className="text-right font-medium">
                  {formatMoney(total)}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}

function IssuePartDialog({ requestId }: { requestId: string }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [partId, setPartId] = useState("")
  const [quantity, setQuantity] = useState("1")

  const partsCatalog = useQuery({
    queryKey: ["spare-parts"],
    queryFn: () => api.get<SparePart[]>("/spare-parts"),
    enabled: open,
  })

  const mutation = useMutation({
    mutationFn: () =>
      api.post(`/requests/${requestId}/parts`, {
        part_id: partId,
        quantity: Number(quantity),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["requests", requestId, "parts"] })
      queryClient.invalidateQueries({ queryKey: ["spare-parts"] })
      toast.success("Деталь выдана в ремонт")
      setPartId("")
      setQuantity("1")
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  const items = (partsCatalog.data ?? [])
    .filter((p) => p.quantity_in_stock > 0)
    .map((p) => ({
      value: p.id,
      label: `${p.name} (на складе: ${p.quantity_in_stock})`,
    }))

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button size="sm">Выдать деталь</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Выдать деталь в ремонт</DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            mutation.mutate()
          }}
          className="space-y-4"
        >
          <div className="space-y-2">
            <Label>Деталь *</Label>
            <Select
              items={items}
              value={partId || null}
              onValueChange={(v) => setPartId(v ?? "")}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Выберите деталь" />
              </SelectTrigger>
              <SelectContent>
                {items.map((p) => (
                  <SelectItem key={p.value} value={p.value}>
                    {p.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="qty">Количество *</Label>
            <Input
              id="qty"
              type="number"
              min="1"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              required
            />
          </div>
          <DialogFooter>
            <Button type="submit" disabled={mutation.isPending || !partId}>
              {mutation.isPending ? "Выдача…" : "Выдать"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Закрытие заявки ─────────────────────────────────────────────────────────

function CloseRequestDialog({ requestId }: { requestId: string }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [comment, setComment] = useState("")

  const mutation = useMutation({
    mutationFn: () => api.patch(`/requests/${requestId}/close`, { comment }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["requests"] })
      toast.success("Заявка закрыта")
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant="outline">Закрыть заявку</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Закрыть заявку?</DialogTitle>
        </DialogHeader>
        <div className="space-y-2">
          <Label htmlFor="close-comment">Комментарий</Label>
          <Textarea
            id="close-comment"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="Например: ремонт завершён, устройство выдано"
          />
        </div>
        <DialogFooter>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending}>
            {mutation.isPending ? "Закрытие…" : "Закрыть заявку"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Счёт ────────────────────────────────────────────────────────────────────

function InvoiceSection({
  requestId,
  invoice,
  isClosed,
  canManage,
}: {
  requestId: string
  invoice: Invoice | null
  isClosed: boolean
  canManage: boolean
}) {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: () => api.post<Invoice>(`/requests/${requestId}/invoice`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["requests", requestId, "invoice"] })
      toast.success("Счёт выставлен")
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  const statusMutation = useMutation({
    mutationFn: (status: "paid" | "cancelled") =>
      api.patch<Invoice>(`/invoices/${invoice!.id}/status`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["requests", requestId, "invoice"] })
      toast.success("Статус счёта обновлён")
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle className="text-base">Счёт</CardTitle>
        {canManage && !invoice && (
          <Button
            size="sm"
            onClick={() => createMutation.mutate()}
            disabled={!isClosed || createMutation.isPending}
            title={!isClosed ? "Счёт можно выставить только по закрытой заявке" : ""}
          >
            {createMutation.isPending ? "Выставление…" : "Выставить счёт"}
          </Button>
        )}
      </CardHeader>
      <CardContent className="text-sm">
        {!invoice ? (
          <p className="text-muted-foreground">
            {isClosed
              ? "Счёт ещё не выставлен."
              : "Счёт выставляется после закрытия заявки."}
          </p>
        ) : (
          <div className="flex items-center gap-6">
            <div>
              <p className="text-2xl font-semibold">
                {formatMoney(invoice.total_amount)}
              </p>
              <p className="text-muted-foreground">
                Выставлен {formatDate(invoice.issued_at)}
              </p>
            </div>
            <Badge
              variant={invoice.status === "paid" ? "default" : "secondary"}
            >
              {INVOICE_STATUS_LABEL[invoice.status]}
            </Badge>
            {canManage && invoice.status === "pending" && (
              <div className="flex gap-2">
                <Button
                  size="sm"
                  onClick={() => statusMutation.mutate("paid")}
                  disabled={statusMutation.isPending}
                >
                  Отметить оплаченным
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => statusMutation.mutate("cancelled")}
                  disabled={statusMutation.isPending}
                >
                  Отменить
                </Button>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
