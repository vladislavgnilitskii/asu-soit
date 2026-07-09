import { Skeleton } from "@/components/ui/skeleton"

interface TableSkeletonProps {
  columns: number
  rows?: number
}

// Заглушка-«скелет» на время загрузки таблицы: рамка + мерцающие ячейки.
export function TableSkeleton({ columns, rows = 6 }: TableSkeletonProps) {
  return (
    <div className="rounded-md border divide-y">
      {Array.from({ length: rows }).map((_, r) => (
        <div key={r} className="flex items-center gap-4 px-4 py-3">
          {Array.from({ length: columns }).map((_, c) => (
            <Skeleton
              key={c}
              className="h-4"
              style={{ width: c === 0 ? "40%" : `${Math.max(12, 22 - c * 2)}%` }}
            />
          ))}
        </div>
      ))}
    </div>
  )
}
