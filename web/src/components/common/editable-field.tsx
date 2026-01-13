import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Input } from "@/components/ui/input"
import { PencilIcon } from "lucide-react"
import { useState } from "react"

interface EditableFieldProps {
  value: string
  onChange: (value: string) => void
  label?: string
}

export function EditableField({ value, onChange, label }: EditableFieldProps) {
  const [open, setOpen] = useState(false)
  const [inputValue, setInputValue] = useState(value)

  const handleSave = () => {
    onChange(inputValue)
    setOpen(false)
  }

  return (
    <div className="flex items-center gap-2">
      <span className="text-zinc-500">{value}</span>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button variant="ghost" size="icon" className="h-6 w-6">
            <PencilIcon className="h-4 w-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80">
          <div className="grid gap-4">
            {label && (
              <div className="space-y-2">
                <h4 className="font-medium leading-none">{label}</h4>
              </div>
            )}
            <div className="flex gap-2">
              <Input
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                className="h-8"
              />
              <Button onClick={handleSave} className="h-8">
                保存
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
} 