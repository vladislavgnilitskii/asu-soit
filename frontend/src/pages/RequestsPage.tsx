import { useNavigate } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import type { RepairRequest, RequestStatus } from "@/lib/types"
import { formatDate, formatMoney } from "@/lib/format"
import { Badge } from "@/components/ui/badge"
import { CreateRequestDialog } from "@/components/forms/CreateRequestDialog"
import { RequestStatusSelect } from "@/components/forms/RequestStatusSelect"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export function RequestsPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const canCreate = user?.role === "admin" || user?.role === "manager"
  const canChangeStatus =
    user?.role === "admin" ||
    user?.role === "manager" ||
    user?.role === "engineer"

  const requests = useQuery({
    queryKey: ["requests"],
    queryFn: () => api.get<RepairRequest[]>("/requests"),
  })
  const statuses = useQuery({
    queryKey: ["request-statuses"],
    queryFn: () => api.get<RequestStatus[]>("/request-statuses"),
  })

  const statusName = (id: string) =>
    statuses.data?.find((s) => s.id === id)?.name ?? "—"

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Заявки</h1>
          <p className="text-sm text-muted-foreground">
            Инженер видит только назначенные ему заявки (RLS на уровне БД).
          </p>
        </div>
        {canCreate && <CreateRequestDialog />}
      </div>

      {requests.isLoading && (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      )}
      {requests.isError && (
        <p className="text-sm text-destructive">
          Не удалось загрузить заявки: {(requests.error as Error).message}
        </p>
      )}

      {requests.data && (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Неисправность</TableHead>
                <TableHead>Статус</TableHead>
                <TableHead className="text-right">Оценка</TableHead>
                <TableHead className="text-right">Итог</TableHead>
                <TableHead>Создана</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {requests.data.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground"
                  >
                    Заявок нет
                  </TableCell>
                </TableRow>
              )}
              {requests.data.map((r) => (
                <TableRow
                  key={r.id}
                  className="cursor-pointer"
                  onClick={() => navigate(`/requests/${r.id}`)}
                >
                  <TableCell className="max-w-md truncate">
                    {r.problem_description}
                  </TableCell>
                  {/* stopPropagation: клик по селекту не должен открывать карточку */}
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    {canChangeStatus && statuses.data ? (
                      <RequestStatusSelect
                        requestId={r.id}
                        statusId={r.status_id}
                        statuses={statuses.data}
                      />
                    ) : (
                      <Badge variant="secondary">
                        {statusName(r.status_id)}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatMoney(r.estimated_cost)}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatMoney(r.final_cost)}
                  </TableCell>
                  <TableCell>{formatDate(r.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
