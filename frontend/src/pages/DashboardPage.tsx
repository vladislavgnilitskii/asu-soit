import { useQuery } from "@tanstack/react-query"
import { ClipboardList, Wrench, CheckCircle2, type LucideIcon } from "lucide-react"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import type { RequestStats } from "@/lib/types"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

interface StatCardProps {
  title: string
  value: number
  icon: LucideIcon
  loading: boolean
}

function StatCard({ title, value, icon: Icon, loading }: StatCardProps) {
  return (
    <Card className="transition-shadow hover:shadow-md">
      <CardHeader className="flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <span className="flex size-8 items-center justify-center rounded-md bg-accent text-accent-foreground">
          <Icon className="size-4" />
        </span>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-9 w-16" />
        ) : (
          <div className="text-3xl font-semibold tabular-nums">{value}</div>
        )}
      </CardContent>
    </Card>
  )
}

export function DashboardPage() {
  const { user } = useAuth()
  // ключ — потомок ["requests"], чтобы инвалидация списка заявок
  // (создание/смена статуса/закрытие) автоматически обновляла и счётчики
  const stats = useQuery({
    queryKey: ["requests", "stats"],
    queryFn: () => api.get<RequestStats>("/requests/stats"),
  })

  const total = stats.data?.total ?? 0
  const open = stats.data?.open ?? 0
  const closed = stats.data?.closed ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Обзор</h1>
        <p className="text-sm text-muted-foreground">
          Добро пожаловать, {user?.login}.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard
          title="Всего заявок"
          value={total}
          icon={ClipboardList}
          loading={stats.isLoading}
        />
        <StatCard
          title="В работе"
          value={open}
          icon={Wrench}
          loading={stats.isLoading}
        />
        <StatCard
          title="Закрыто"
          value={closed}
          icon={CheckCircle2}
          loading={stats.isLoading}
        />
      </div>
    </div>
  )
}
