'use client'
import { useEffect } from 'react'
import { motion, useSpring, useTransform } from 'motion/react'

interface NumberTickerProps { value: number; duration?: number; className?: string }

export function NumberTicker({ value, duration = 1.5, className }: NumberTickerProps) {
  const spring = useSpring(0, { stiffness: 80, damping: 20, duration: duration * 1000 })
  useEffect(() => { spring.set(value) }, [spring, value])
  const display = useTransform(spring, (latest) => Math.round(latest).toLocaleString())
  return <motion.span className={className}>{display}</motion.span>
}
