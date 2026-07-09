import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { api, ApiError } from "@/lib/api"
import type { RepairRequest, RequestStatus } from "@/lib/types"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface Props {
  requestId: string
  statusId: string
  statuses: RequestStatus[]
}

// Инлайн-смена статуса заявки прямо в таблице.
export function RequestStatusSelect({ requestId, statusId, statuses }: Props) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (newStatusId: string) =>
      api.patch<RepairRequest>(`/requests/${requestId}/status`, {
        status_id: newStatusId,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["requests"] })
      toast.success("Статус обновлён")
    },
    onError: (err) =>
      toast.error(err instanceof ApiError ? err.message : "Ошибка"),
  })

  return (
    <Select
      value={statusId}
      onValueChange={(v) => v && mutation.mutate(v)}
      disabled={mutation.isPending}
    >
      <SelectTrigger className="h-8 w-44">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {statuses.map((s) => (
          <SelectItem key={s.id} value={s.id}>
            {s.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
