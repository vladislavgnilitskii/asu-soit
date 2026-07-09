import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import type { Client } from "@/lib/types"
import { formatDate } from "@/lib/format"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export function ClientsPage() {
  const clients = useQuery({
    queryKey: ["clients"],
    queryFn: () => api.get<Client[]>("/clients"),
  })

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-semibold">Клиенты</h1>
        <p className="text-sm text-muted-foreground">
          Физические лица и организации.
        </p>
      </div>

      {clients.isLoading && (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      )}
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
              {clients.data.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={4}
                    className="text-center text-muted-foreground"
                  >
                    Клиентов нет
                  </TableCell>
                </TableRow>
              )}
              {clients.data.map((c) => (
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
    </div>
  )
}
