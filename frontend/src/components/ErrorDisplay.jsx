import { motion } from 'framer-motion'
import { AlertTriangle, RefreshCw, Wifi, Server } from 'lucide-react'

export default function ErrorDisplay({ error, onRetry }) {
  const isNetworkError = error?.toLowerCase().includes('network') || error?.toLowerCase().includes('fetch')
  const isServerError = error?.toLowerCase().includes('500') || error?.toLowerCase().includes('server')

  const Icon = isNetworkError ? Wifi : isServerError ? Server : AlertTriangle
  const title = isNetworkError ? 'CONNECTION LOST' : isServerError ? 'SERVER ERROR' : 'SYSTEM MALFUNCTION'

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="relative max-w-lg mx-auto p-8 bg-void-lighter border-2 border-neon-pink"
    >
      {/* Glitch overlay */}
      <motion.div
        animate={{
          opacity: [0, 0.5, 0],
          x: [-2, 2, -2],
        }}
        transition={{ duration: 0.3, repeat: Infinity }}
        className="absolute inset-0 bg-neon-pink/5 pointer-events-none"
      />

      {/* Scanlines */}
      <div className="absolute inset-0 pointer-events-none scanlines opacity-30" />

      {/* Content */}
      <div className="relative text-center">
        {/* Icon */}
        <motion.div
          animate={{
            rotate: [0, 5, -5, 0],
            scale: [1, 1.05, 1],
          }}
          transition={{ duration: 2, repeat: Infinity }}
          className="inline-flex items-center justify-center w-20 h-20 mb-6 border-2 border-neon-pink"
          style={{ boxShadow: '0 0 30px rgba(255, 0, 110, 0.3)' }}
        >
          <Icon className="w-10 h-10 text-neon-pink" />
        </motion.div>

        {/* Title */}
        <h3 className="text-2xl font-bold font-mono text-neon-pink text-glow-pink mb-2 glitch" data-text={title}>
          {title}
        </h3>

        {/* Error code */}
        <p className="font-mono text-muted text-sm mb-4">
          // ERR_{Math.random().toString(36).substring(2, 8).toUpperCase()}
        </p>

        {/* Error message */}
        <div className="p-4 bg-void border-l-4 border-neon-pink mb-6">
          <p className="text-muted-light text-sm font-mono break-words">
            {error || 'An unexpected error occurred. The matrix has you.'}
          </p>
        </div>

        {/* Retry button */}
        {onRetry && (
          <motion.button
            onClick={onRetry}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="
              inline-flex items-center gap-2 px-6 py-3
              bg-neon-pink text-void font-mono font-bold
              hover:bg-neon-cyan transition-colors duration-300
            "
          >
            <RefreshCw className="w-4 h-4" />
            RETRY CONNECTION
          </motion.button>
        )}

        {/* Decorative ASCII art */}
        <div className="mt-8 text-muted/30 font-mono text-xs">
          <pre className="leading-tight">
{`    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
    ▓  SIGNAL INTERRUPTED  ▓
    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓`}
          </pre>
        </div>
      </div>

      {/* Corner accents */}
      <div className="absolute top-0 left-0 w-6 h-6 border-t-2 border-l-2 border-neon-pink -translate-x-1 -translate-y-1" />
      <div className="absolute top-0 right-0 w-6 h-6 border-t-2 border-r-2 border-neon-pink translate-x-1 -translate-y-1" />
      <div className="absolute bottom-0 left-0 w-6 h-6 border-b-2 border-l-2 border-neon-pink -translate-x-1 translate-y-1" />
      <div className="absolute bottom-0 right-0 w-6 h-6 border-b-2 border-r-2 border-neon-pink translate-x-1 translate-y-1" />
    </motion.div>
  )
}
