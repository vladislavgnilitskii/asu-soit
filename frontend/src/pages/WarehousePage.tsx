import { useState, type FormEvent } from "react"
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { formatMoney } from "@/lib/format"
import type {
  CreateSparePartDTO,
  Page,
  PartCategory,
  SparePart,
} from "@/lib/types"
import { Pagination } from "@/components/Pagination"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
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
import { TableSkeleton } from "@/components/TableSkeleton"

export function WarehousePage() {
  const { user } = useAuth()
  const canWrite = user?.role === "admin" || user?.role === "storekeeper"

  const limit = 20
  const [offset, setOffset] = useState(0)
  const parts = useQuery({
    queryKey: ["spare-parts", offset],
    queryFn: () =>
      api.get<Page<SparePart>>(`/spare-parts?limit=${limit}&offset=${offset}`),
    placeholderData: keepPreviousData,
  })
  const categories = useQuery({
    queryKey: ["part-categories"],
    queryFn: () => api.get<PartCategory[]>("/part-categories"),
  })

  const categoryName = (id: string) =>
    categories.data?.find((c) => c.id === id)?.name ?? "—"

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Склад</h1>
          <p className="text-sm text-muted-foreground">
            Запчасти: остатки, приход, списание.
          </p>
        </div>
        {canWrite && <CreatePartDialog categories={categories.data ?? []} />}
      </div>

      {parts.isLoading && <TableSkeleton columns={8} />}
      {parts.isError && (
        <p className="text-sm text-destructive">
          Не удалось загрузить склад: {(parts.error as Error).message}
        </p>
      )}

      {parts.data && (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Наименование</TableHead>
                <TableHead>Категория</TableHead>
                <TableHead>Артикул</TableHead>
                <TableHead className="text-right">Закупка</TableHead>
                <TableHead className="text-right">Продажа</TableHead>
                <TableHead className="text-right">Остаток</TableHead>
                {canWrite && <TableHead />}
              </TableRow>
            </TableHeader>
            <TableBody>
              {parts.data.items.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={canWrite ? 7 : 6}
                    className="text-center text-muted-foreground"
                  >
                    Склад пуст
                  </TableCell>
                </TableRow>
              )}
              {parts.data.items.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>{categoryName(p.category_id)}</TableCell>
                  <TableCell>{p.sku || "—"}</TableCell>
                  <TableCell className="text-right">
                    {formatMoney(p.purchase_price)}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatMoney(p.sale_price)}
                  </TableCell>
                  <TableCell className="text-right">
                    <Badge
                      variant={p.quantity_in_stock > 0 ? "secondary" : "destructive"}
                    >
                      {p.quantity_in_stock}
                    </Badge>
                  </TableCell>
                  {canWrite && (
                    <TableCell className="text-right space-x-1">
                      <StockDialog part={p} mode="receive" />
                      <StockDialog part={p} mode="writeoff" />
                    </TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {parts.data && (
        <Pagination
          total={parts.data.total}
          limit={parts.data.limit}
          offset={parts.data.offset}
          onChange={setOffset}
        />
      )}
    </div>
  )
}

// ── Новая позиция каталога ──────────────────────────────────────────────────

const EMPTY_PART = {
  category_id: "",
  name: "",
  sku: "",
  purchase_price: "",
  sale_price: "",
}

function CreatePartDialog({ categories }: { categories: PartCategory[] }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ ...EMPTY_PART })

  const set = (key: keyof typeof EMPTY_PART, value: string) =>
    setForm((f) => ({ ...f, [key]: value }))

  const items = categories.map((c) => ({ value: c.id, label: c.name }))

  const mutation = useMutation({
    mutationFn: (dto: CreateSparePartDTO) =>
      api.post<SparePart>("/spare-parts", dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["spare-parts"] })
      toast.success("Позиция добавлена")
      setForm({ ...EMPTY_PART })
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    mutation.mutate({
      category_id: form.category_id,
      name: form.name,
      sku: form.sku || undefined,
      purchase_price: Number(form.purchase_price),
      sale_price: Number(form.sale_price),
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button>Добавить позицию</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новая позиция каталога</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Категория *</Label>
            <Select
              items={items}
              value={form.category_id || null}
              onValueChange={(v) => set("category_id", v ?? "")}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Выберите категорию" />
              </SelectTrigger>
              <SelectContent>
                {items.map((c) => (
                  <SelectItem key={c.value} value={c.value}>
                    {c.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="pname">Наименование *</Label>
            <Input
              id="pname"
              value={form.name}
              onChange={(e) => set("name", e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="sku">Артикул</Label>
            <Input
              id="sku"
              value={form.sku}
              onChange={(e) => set("sku", e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="pp">Закупочная цена *</Label>
              <Input
                id="pp"
                type="number"
                min="0"
                step="0.01"
                value={form.purchase_price}
                onChange={(e) => set("purchase_price", e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="sp">Цена продажи *</Label>
              <Input
                id="sp"
                type="number"
                min="0"
                step="0.01"
                value={form.sale_price}
                onChange={(e) => set("sale_price", e.target.value)}
                required
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              type="submit"
              disabled={mutation.isPending || !form.category_id}
            >
              {mutation.isPending ? "Сохранение…" : "Сохранить"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Приход / списание ───────────────────────────────────────────────────────

function StockDialog({
  part,
  mode,
}: {
  part: SparePart
  mode: "receive" | "writeoff"
}) {
  const isReceive = mode === "receive"
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [quantity, setQuantity] = useState("1")
  const [unitPrice, setUnitPrice] = useState(part.purchase_price.toString())
  const [invoiceNumber, setInvoiceNumber] = useState("")
  const [note, setNote] = useState("")

  const mutation = useMutation({
    mutationFn: () =>
      api.post<SparePart>(
        `/spare-parts/${part.id}/${mode}`,
        isReceive
          ? {
              quantity: Number(quantity),
              unit_price: Number(unitPrice),
              invoice_number: invoiceNumber || undefined,
              note: note || undefined,
            }
          : { quantity: Number(quantity), note: note || undefined },
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["spare-parts"] })
      toast.success(isReceive ? "Приход оформлен" : "Списание оформлено")
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={
          <Button size="sm" variant={isReceive ? "outline" : "ghost"}>
            {isReceive ? "Приход" : "Списание"}
          </Button>
        }
      />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isReceive ? "Приход" : "Списание"}: {part.name}
          </DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            mutation.mutate()
          }}
          className="space-y-4"
        >
          <div className="space-y-2">
            <Label htmlFor="q">Количество *</Label>
            <Input
              id="q"
              type="number"
              min="1"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              required
            />
          </div>
          {isReceive && (
            <>
              <div className="space-y-2">
                <Label htmlFor="up">Цена за единицу *</Label>
                <Input
                  id="up"
                  type="number"
                  min="0"
                  step="0.01"
                  value={unitPrice}
                  onChange={(e) => setUnitPrice(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="inv">Номер накладной</Label>
                <Input
                  id="inv"
                  value={invoiceNumber}
                  onChange={(e) => setInvoiceNumber(e.target.value)}
                />
              </div>
            </>
          )}
          <div className="space-y-2">
            <Label htmlFor="note">Примечание</Label>
            <Input
              id="note"
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending
                ? "Оформление…"
                : isReceive
                  ? "Оформить приход"
                  : "Списать"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
