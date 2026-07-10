import { useState } from "react"
import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import type { Device, DeviceType, Page } from "@/lib/types"
import { formatDate } from "@/lib/format"
import { Pagination } from "@/components/Pagination"
import { CreateDeviceDialog } from "@/components/forms/CreateDeviceDialog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableSkeleton } from "@/components/TableSkeleton"

export function DevicesPage() {
  const { user } = useAuth()
  const canCreate = user?.role === "admin" || user?.role === "manager"

  const limit = 20
  const [offset, setOffset] = useState(0)
  const devices = useQuery({
    queryKey: ["devices", offset],
    queryFn: () =>
      api.get<Page<Device>>(`/devices?limit=${limit}&offset=${offset}`),
    placeholderData: keepPreviousData,
  })
  const types = useQuery({
    queryKey: ["device-types"],
    queryFn: () => api.get<DeviceType[]>("/device-types"),
  })

  const typeName = (id: string) =>
    types.data?.find((t) => t.id === id)?.name ?? "—"

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Устройства</h1>
          <p className="text-sm text-muted-foreground">
            Техника клиентов, принятая в ремонт.
          </p>
        </div>
        {canCreate && <CreateDeviceDialog />}
      </div>

      {devices.isLoading && <TableSkeleton columns={6} />}
      {devices.isError && (
        <p className="text-sm text-destructive">
          Не удалось загрузить устройства: {(devices.error as Error).message}
        </p>
      )}

      {devices.data && (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Тип</TableHead>
                <TableHead>Бренд</TableHead>
                <TableHead>Модель</TableHead>
                <TableHead>Серийный №</TableHead>
                <TableHead>Добавлено</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {devices.data.items.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground"
                  >
                    Устройств нет
                  </TableCell>
                </TableRow>
              )}
              {devices.data.items.map((d) => (
                <TableRow key={d.id}>
                  <TableCell>{typeName(d.device_type_id)}</TableCell>
                  <TableCell>{d.brand}</TableCell>
                  <TableCell>{d.model}</TableCell>
                  <TableCell>{d.serial_number || "—"}</TableCell>
                  <TableCell>{formatDate(d.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {devices.data && (
        <Pagination
          total={devices.data.total}
          limit={devices.data.limit}
          offset={devices.data.offset}
          onChange={setOffset}
        />
      )}
    </div>
  )
}
