import { useState, type FormEvent } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import type {
  Client,
  CreateDeviceDTO,
  Device,
  DeviceType,
} from "@/lib/types"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
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

const EMPTY = {
  client_id: "",
  device_type_id: "",
  brand: "",
  model: "",
  serial_number: "",
  appearance_note: "",
}

export function CreateDeviceDialog() {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ ...EMPTY })

  // справочники подгружаем только когда диалог открыт
  const clients = useQuery({
    queryKey: ["clients"],
    queryFn: () => api.get<Client[]>("/clients"),
    enabled: open,
  })
  const types = useQuery({
    queryKey: ["device-types"],
    queryFn: () => api.get<DeviceType[]>("/device-types"),
    enabled: open,
  })

  const set = (key: keyof typeof EMPTY, value: string) =>
    setForm((f) => ({ ...f, [key]: value }))

  const mutation = useMutation({
    mutationFn: (dto: CreateDeviceDTO) => api.post<Device>("/devices", dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["devices"] })
      toast.success("Устройство добавлено")
      setForm({ ...EMPTY })
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    mutation.mutate({
      client_id: form.client_id,
      device_type_id: form.device_type_id,
      brand: form.brand,
      model: form.model,
      serial_number: form.serial_number || undefined,
      appearance_note: form.appearance_note || undefined,
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button>Добавить устройство</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новое устройство</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Клиент *</Label>
            <Select
              value={form.client_id || null}
              onValueChange={(v) => set("client_id", v ?? "")}
            >
              <SelectTrigger>
                <SelectValue placeholder="Выберите клиента" />
              </SelectTrigger>
              <SelectContent>
                {clients.data?.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.phone} (
                    {c.client_type === "individual" ? "физлицо" : "организация"})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Тип устройства *</Label>
            <Select
              value={form.device_type_id || null}
              onValueChange={(v) => set("device_type_id", v ?? "")}
            >
              <SelectTrigger>
                <SelectValue placeholder="Выберите тип" />
              </SelectTrigger>
              <SelectContent>
                {types.data?.map((t) => (
                  <SelectItem key={t.id} value={t.id}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="brand">Бренд *</Label>
              <Input
                id="brand"
                value={form.brand}
                onChange={(e) => set("brand", e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="model">Модель *</Label>
              <Input
                id="model"
                value={form.model}
                onChange={(e) => set("model", e.target.value)}
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="serial">Серийный номер</Label>
            <Input
              id="serial"
              value={form.serial_number}
              onChange={(e) => set("serial_number", e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="note">Внешний вид / примечание</Label>
            <Textarea
              id="note"
              value={form.appearance_note}
              onChange={(e) => set("appearance_note", e.target.value)}
            />
          </div>

          <DialogFooter>
            <Button
              type="submit"
              disabled={mutation.isPending || !form.client_id || !form.device_type_id}
            >
              {mutation.isPending ? "Сохранение…" : "Сохранить"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
