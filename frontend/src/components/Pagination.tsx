import { ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"

interface PaginationProps {
  total: number
  limit: number
  offset: number
  onChange: (offset: number) => void
}

// Постраничная навигация «Назад/Вперёд» с подписью «X–Y из T».
export function Pagination({ total, limit, offset, onChange }: PaginationProps) {
  if (total === 0) return null

  const from = offset + 1
  const to = Math.min(offset + limit, total)
  const canPrev = offset > 0
  const canNext = offset + limit < total

  return (
    <div className="flex items-center justify-between px-1 text-sm text-muted-foreground">
      <span className="tabular-nums">
        {from}–{to} из {total}
      </span>
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size="sm"
          disabled={!canPrev}
          onClick={() => onChange(Math.max(0, offset - limit))}
        >
          <ChevronLeft className="size-4" />
          Назад
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={!canNext}
          onClick={() => onChange(offset + limit)}
        >
          Вперёд
          <ChevronRight className="size-4" />
        </Button>
      </div>
    </div>
  )
}
