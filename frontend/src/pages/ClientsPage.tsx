import { useState } from "react"
import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import type { Client, Page } from "@/lib/types"
import { formatDate } from "@/lib/format"
import { Badge } from "@/components/ui/badge"
import { Pagination } from "@/components/Pagination"
import { CreateClientDialog } from "@/components/forms/CreateClientDialog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableSkeleton } from "@/components/TableSkeleton"

export function ClientsPage() {
  const limit = 20
  const [offset, setOffset] = useState(0)
  const clients = useQuery({
    queryKey: ["clients", offset],
    queryFn: () =>
      api.get<Page<Client>>(`/clients?limit=${limit}&offset=${offset}`),
    placeholderData: keepPreviousData,
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Клиенты</h1>
          <p className="text-sm text-muted-foreground">
            Физические лица и организации.
          </p>
        </div>
        <CreateClientDialog />
      </div>

      {clients.isLoading && <TableSkeleton columns={5} />}
      {clients.isError && (
        <p className="text-sm text-destructive">
          Не удалось загрузить клиентов: {(clients.error as Error).message}
        </p>
      )}

      {clients.data && (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Тип</TableHead>
                <TableHead>Телефон</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Добавлен</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {clients.data.items.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={4}
                    className="text-center text-muted-foreground"
                  >
                    Клиентов нет
                  </TableCell>
                </TableRow>
              )}
              {clients.data.items.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>
                    <Badge variant="outline">
                      {c.client_type === "individual"
                        ? "Физлицо"
                        : "Организация"}
                    </Badge>
                  </TableCell>
                  <TableCell>{c.phone}</TableCell>
                  <TableCell>{c.email || "—"}</TableCell>
                  <TableCell>{formatDate(c.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {clients.data && (
        <Pagination
          total={clients.data.total}
          limit={clients.data.limit}
          offset={clients.data.offset}
          onChange={setOffset}
        />
      )}
    </div>
  )
}
