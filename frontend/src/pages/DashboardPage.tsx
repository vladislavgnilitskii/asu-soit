import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import type { RepairRequest } from "@/lib/types"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export function DashboardPage() {
  const { user } = useAuth()
  const requests = useQuery({
    queryKey: ["requests"],
    queryFn: () => api.get<RepairRequest[]>("/requests"),
  })

  const total = requests.data?.length ?? 0
  const open = requests.data?.filter((r) => !r.closed_at).length ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Обзор</h1>
        <p className="text-sm text-muted-foreground">
          Добро пожаловать, {user?.login}.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Всего заявок
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-semibold">{total}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              В работе
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-semibold">{open}</div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
