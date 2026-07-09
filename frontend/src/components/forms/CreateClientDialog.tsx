import { useState, type FormEvent } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import type { Client, ClientType, CreateClientRequest } from "@/lib/types"
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

const EMPTY = {
  client_type: "individual" as ClientType,
  phone: "",
  email: "",
  last_name: "",
  first_name: "",
  middle_name: "",
  name: "",
  inn: "",
  kpp: "",
  contact_person: "",
}

export function CreateClientDialog() {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ ...EMPTY })

  const set = (key: keyof typeof EMPTY, value: string) =>
    setForm((f) => ({ ...f, [key]: value }))

  const mutation = useMutation({
    mutationFn: (dto: CreateClientRequest) => api.post<Client>("/clients", dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clients"] })
      toast.success("Клиент добавлен")
      setForm({ ...EMPTY })
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const isInd = form.client_type === "individual"
    const dto: CreateClientRequest = {
      client_type: form.client_type,
      phone: form.phone,
      email: form.email || undefined,
      ...(isInd
        ? {
            last_name: form.last_name,
            first_name: form.first_name,
            middle_name: form.middle_name || undefined,
          }
        : {
            name: form.name,
            inn: form.inn,
            kpp: form.kpp || undefined,
            contact_person: form.contact_person || undefined,
          }),
    }
    mutation.mutate(dto)
  }

  const isInd = form.client_type === "individual"

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button>Добавить клиента</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новый клиент</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Тип</Label>
            <Select
              items={[
                { value: "individual", label: "Физлицо" },
                { value: "organization", label: "Организация" },
              ]}
              value={form.client_type}
              onValueChange={(v) => v && set("client_type", v)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="individual">Физлицо</SelectItem>
                <SelectItem value="organization">Организация</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="phone">Телефон *</Label>
            <Input
              id="phone"
              value={form.phone}
              onChange={(e) => set("phone", e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={form.email}
              onChange={(e) => set("email", e.target.value)}
            />
          </div>

          {isInd ? (
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label htmlFor="last_name">Фамилия *</Label>
                <Input
                  id="last_name"
                  value={form.last_name}
                  onChange={(e) => set("last_name", e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="first_name">Имя *</Label>
                <Input
                  id="first_name"
                  value={form.first_name}
                  onChange={(e) => set("first_name", e.target.value)}
                  required
                />
              </div>
              <div className="col-span-2 space-y-2">
                <Label htmlFor="middle_name">Отчество</Label>
                <Input
                  id="middle_name"
                  value={form.middle_name}
                  onChange={(e) => set("middle_name", e.target.value)}
                />
              </div>
            </div>
          ) : (
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="name">Название *</Label>
                <Input
                  id="name"
                  value={form.name}
                  onChange={(e) => set("name", e.target.value)}
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="inn">ИНН *</Label>
                  <Input
                    id="inn"
                    value={form.inn}
                    onChange={(e) => set("inn", e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="kpp">КПП</Label>
                  <Input
                    id="kpp"
                    value={form.kpp}
                    onChange={(e) => set("kpp", e.target.value)}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="contact_person">Контактное лицо</Label>
                <Input
                  id="contact_person"
                  value={form.contact_person}
                  onChange={(e) => set("contact_person", e.target.value)}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? "Сохранение…" : "Сохранить"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
