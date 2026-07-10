import { useState, type FormEvent } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import type {
  CreateRepairRequestDTO,
  Device,
  Page,
  RepairRequest,
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

export function CreateRequestDialog() {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [deviceId, setDeviceId] = useState("")
  const [problem, setProblem] = useState("")
  const [deadline, setDeadline] = useState("")

  const devices = useQuery({
    queryKey: ["devices"],
    queryFn: () => api.get<Page<Device>>("/devices?limit=100"),
    enabled: open,
  })

  const reset = () => {
    setDeviceId("")
    setProblem("")
    setDeadline("")
  }

  // items нужны Base UI, чтобы SelectValue показывал подпись, а не сырое value
  const deviceItems = (devices.data?.items ?? []).map((d) => ({
    value: d.id,
    label: `${d.brand} ${d.model}${d.serial_number ? ` · ${d.serial_number}` : ""}`,
  }))

  const mutation = useMutation({
    mutationFn: (dto: CreateRepairRequestDTO) =>
      api.post<RepairRequest>("/requests", dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["requests"] })
      toast.success("Заявка создана")
      reset()
      setOpen(false)
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    mutation.mutate({
      device_id: deviceId,
      problem_description: problem,
      // бэк ждёт time.Time (RFC3339) — переводим дату в ISO, иначе не шлём
      planned_deadline: deadline ? new Date(deadline).toISOString() : undefined,
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button>Создать заявку</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новая заявка</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Устройство *</Label>
            <Select
              items={deviceItems}
              value={deviceId || null}
              onValueChange={(v) => setDeviceId(v ?? "")}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Выберите устройство" />
              </SelectTrigger>
              <SelectContent>
                {deviceItems.map((d) => (
                  <SelectItem key={d.value} value={d.value}>
                    {d.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="problem">Описание неисправности *</Label>
            <Textarea
              id="problem"
              value={problem}
              onChange={(e) => setProblem(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="deadline">Плановый срок</Label>
            <Input
              id="deadline"
              type="date"
              value={deadline}
              onChange={(e) => setDeadline(e.target.value)}
            />
          </div>

          <DialogFooter>
            <Button
              type="submit"
              disabled={mutation.isPending || !deviceId || !problem}
            >
              {mutation.isPending ? "Создание…" : "Создать"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
