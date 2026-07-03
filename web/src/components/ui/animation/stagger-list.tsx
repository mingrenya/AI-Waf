'use client'
import { motion, type Variants } from 'motion/react'
import type { ReactNode } from 'react'

const containerVariants: Variants = {
  initial: {},
  animate: { transition: { staggerChildren: 0.06, delayChildren: 0.05 } },
}
const itemVariants: Variants = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0, transition: { duration: 0.3 } },
}

export function StaggerList({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <motion.div variants={containerVariants} initial="initial" animate="animate" className={className}>
      {Array.isArray(children)
        ? children.map((child, i) => (<motion.div key={i} variants={itemVariants}>{child}</motion.div>))
        : children}
    </motion.div>
  )
}
