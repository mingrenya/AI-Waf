import * as React from "react"
import { CalendarIcon } from "lucide-react"
import { format } from "date-fns"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area"

// 定义选择器类型
export type DateTimePickerType = "date" | "dateHour" | "dateHourMinute" | "dateHourMinuteSecond"

interface DateTimePicker24hProps {
    value?: Date
    onChange?: (date: Date | undefined) => void
    className?: string
    // 添加类型属性，默认为完整的日期时间选择
    type?: DateTimePickerType
}

export function DateTimePicker24h({
    value,
    onChange,
    className,
    type = "dateHourMinute" // 默认为日期+小时+分钟
}: DateTimePicker24hProps) {
    // 使用内部状态管理弹出框的开关状态
    const [isOpen, setIsOpen] = React.useState(false)

    const hours = Array.from({ length: 24 }, (_, i) => i)
    const minutes = Array.from({ length: 12 }, (_, i) => i * 5)
    const seconds = Array.from({ length: 12 }, (_, i) => i * 5)

    const handleDateSelect = (selectedDate: Date | undefined) => {
        if (selectedDate && onChange) {
            // 保留之前选择的时间
            if (value) {
                selectedDate.setHours(value.getHours(), value.getMinutes(), value.getSeconds())
            }
            onChange(selectedDate)
        }
    }

    const handleTimeChange = (
        timeType: "hour" | "minute" | "second",
        valueStr: string
    ) => {
        if (value && onChange) {
            const newDate = new Date(value)
            if (timeType === "hour") {
                newDate.setHours(parseInt(valueStr))
            } else if (timeType === "minute") {
                newDate.setMinutes(parseInt(valueStr))
            } else if (timeType === "second") {
                newDate.setSeconds(parseInt(valueStr))
            }
            onChange(newDate)
        } else if (onChange) {
            // 如果尚未选择日期，则创建当前日期并设置时间
            const newDate = new Date()
            if (timeType === "hour") {
                newDate.setHours(parseInt(valueStr))
            } else if (timeType === "minute") {
                newDate.setMinutes(parseInt(valueStr))
            } else if (timeType === "second") {
                newDate.setSeconds(parseInt(valueStr))
            }
            onChange(newDate)
        }
    }

    // 根据类型决定显示格式
    const getFormatString = () => {
        switch (type) {
            case "date":
                return "yyyy/MM/dd"
            case "dateHour":
                return "yyyy/MM/dd HH"
            case "dateHourMinute":
                return "yyyy/MM/dd HH:mm"
            case "dateHourMinuteSecond":
                return "yyyy/MM/dd HH:mm:ss"
            default:
                return "yyyy/MM/dd HH:mm"
        }
    }

    // 根据类型决定占位符文本
    const getPlaceholderText = () => {
        switch (type) {
            case "date":
                return "YYYY/MM/DD"
            case "dateHour":
                return "YYYY/MM/DD HH"
            case "dateHourMinute":
                return "YYYY/MM/DD HH:MM"
            case "dateHourMinuteSecond":
                return "YYYY/MM/DD HH:MM:SS"
            default:
                return "YYYY/MM/DD HH:MM"
        }
    }

    return (
        <Popover open={isOpen} onOpenChange={setIsOpen}>
            <PopoverTrigger asChild>
                <Button
                    variant="outline"
                    className={cn(
                        "w-full justify-start text-left font-normal",
                        !value && "text-muted-foreground",
                        className
                    )}
                >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {value ? (
                        <span className="truncate">
                            {format(value, getFormatString())}
                        </span>
                    ) : (
                        <span className="truncate text-muted-foreground">
                            {getPlaceholderText()}
                        </span>
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0">
                <div className="sm:flex">
                    <Calendar
                        mode="single"
                        selected={value}
                        onSelect={handleDateSelect}
                        initialFocus
                    />

                    {/* 根据类型条件渲染时间选择器 */}
                    {type !== "date" && (
                        <div className="flex flex-col sm:flex-row sm:h-[300px] divide-y sm:divide-y-0 sm:divide-x">
                            {/* 小时选择器 - 对于所有包含小时的类型 */}
                            <ScrollArea className="w-64 sm:w-auto">
                                <div className="flex sm:flex-col p-2">
                                    {hours.map((hour) => (
                                        <Button
                                            key={hour}
                                            size="icon"
                                            variant={value && value.getHours() === hour ? "default" : "ghost"}
                                            className="sm:w-full shrink-0 aspect-square"
                                            onClick={() => handleTimeChange("hour", hour.toString())}
                                        >
                                            {hour.toString().padStart(2, '0')}
                                        </Button>
                                    ))}
                                </div>
                                <ScrollBar orientation="horizontal" className="sm:hidden" />
                            </ScrollArea>

                            {/* 分钟选择器 - 仅对包含分钟的类型 */}
                            {(type === "dateHourMinute" || type === "dateHourMinuteSecond") && (
                                <ScrollArea className="w-64 sm:w-auto">
                                    <div className="flex sm:flex-col p-2">
                                        {minutes.map((minute) => (
                                            <Button
                                                key={minute}
                                                size="icon"
                                                variant={value && value.getMinutes() === minute ? "default" : "ghost"}
                                                className="sm:w-full shrink-0 aspect-square"
                                                onClick={() => handleTimeChange("minute", minute.toString())}
                                            >
                                                {minute.toString().padStart(2, '0')}
                                            </Button>
                                        ))}
                                    </div>
                                    <ScrollBar orientation="horizontal" className="sm:hidden" />
                                </ScrollArea>
                            )}

                            {/* 秒选择器 - 仅对包含秒的类型 */}
                            {type === "dateHourMinuteSecond" && (
                                <ScrollArea className="w-64 sm:w-auto">
                                    <div className="flex sm:flex-col p-2">
                                        {seconds.map((second) => (
                                            <Button
                                                key={second}
                                                size="icon"
                                                variant={value && value.getSeconds() === second ? "default" : "ghost"}
                                                className="sm:w-full shrink-0 aspect-square"
                                                onClick={() => handleTimeChange("second", second.toString())}
                                            >
                                                {second.toString().padStart(2, '0')}
                                            </Button>
                                        ))}
                                    </div>
                                    <ScrollBar orientation="horizontal" className="sm:hidden" />
                                </ScrollArea>
                            )}
                        </div>
                    )}
                </div>
            </PopoverContent>
        </Popover>
    )
}