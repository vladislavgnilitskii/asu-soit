import { useState, type FormEvent } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import type { CreateEmployeeDTO, Employee, Role } from "@/lib/types"
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

export function EmployeesPage() {
  const queryClient = useQueryClient()

  const employees = useQuery({
    queryKey: ["employees"],
    queryFn: () => api.get<Employee[]>("/employees"),
  })
  const roles = useQuery({
    queryKey: ["roles"],
    queryFn: () => api.get<Role[]>("/roles"),
  })

  const roleName = (id: string) =>
    roles.data?.find((r) => r.id === id)?.name ?? "—"

  const toggleActive = useMutation({
    mutationFn: (emp: Employee) =>
      api.patch<Employee>(`/employees/${emp.id}`, { is_active: !emp.is_active }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["employees"] })
      toast.success("Сохранено")
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Сотрудники</h1>
          <p className="text-sm text-muted-foreground">
            Учётные записи и роли. Доступно только администратору.
          </p>
        </div>
        <CreateEmployeeDialog roles={roles.data ?? []} />
      </div>

      {employees.isLoading && (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      )}
      {employees.isError && (
        <p className="text-sm text-destructive">
          Не удалось загрузить сотрудников: {(employees.error as Error).message}
        </p>
      )}

      {employees.data && (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ФИО</TableHead>
                <TableHead>Логин</TableHead>
                <TableHead>Роль</TableHead>
                <TableHead>Статус</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {employees.data.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="font-medium">
                    {e.last_name} {e.first_name} {e.middle_name ?? ""}
                  </TableCell>
                  <TableCell>{e.login}</TableCell>
                  <TableCell>{roleName(e.role_id)}</TableCell>
                  <TableCell>
                    <Badge variant={e.is_active ? "secondary" : "destructive"}>
                      {e.is_active ? "Активен" : "Отключён"}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => toggleActive.mutate(e)}
                      disabled={toggleActive.isPending}
                    >
                      {e.is_active ? "Отключить" : "Включить"}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}

const EMPTY = {
  role_id: "",
  last_name: "",
  first_name: "",
  middle_name: "",
  login: "",
  password: "",
}

function CreateEmployeeDialog({ roles }: { roles: Role[] }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ ...EMPTY })

  const set = (key: keyof typeof EMPTY, value: string) =>
    setForm((f) => ({ ...f, [key]: value }))

  const items = roles.map((r) => ({ value: r.id, label: r.name }))

  const mutation = useMutation({
    mutationFn: (dto: CreateEmployeeDTO) =>
      api.post<Employee>("/employees", dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["employees"] })
      toast.success("Сотрудник создан")
      setForm({ ...EMPTY })
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    mutation.mutate({
      role_id: form.role_id,
      last_name: form.last_name,
      first_name: form.first_name,
      middle_name: form.middle_name || undefined,
      login: form.login,
      password: form.password,
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button>Добавить сотрудника</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новый сотрудник</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="ln">Фамилия *</Label>
              <Input
                id="ln"
                value={form.last_name}
                onChange={(e) => set("last_name", e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="fn">Имя *</Label>
              <Input
                id="fn"
                value={form.first_name}
                onChange={(e) => set("first_name", e.target.value)}
                required
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="mn">Отчество</Label>
            <Input
              id="mn"
              value={form.middle_name}
              onChange={(e) => set("middle_name", e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label>Роль *</Label>
            <Select
              items={items}
              value={form.role_id || null}
              onValueChange={(v) => set("role_id", v ?? "")}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Выберите роль" />
              </SelectTrigger>
              <SelectContent>
                {items.map((r) => (
                  <SelectItem key={r.value} value={r.value}>
                    {r.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="lg">Логин *</Label>
              <Input
                id="lg"
                value={form.login}
                onChange={(e) => set("login", e.target.value)}
                autoComplete="off"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="pw">Пароль * (мин. 6)</Label>
              <Input
                id="pw"
                type="password"
                minLength={6}
                value={form.password}
                onChange={(e) => set("password", e.target.value)}
                autoComplete="new-password"
                required
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={mutation.isPending || !form.role_id}>
              {mutation.isPending ? "Создание…" : "Создать"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
