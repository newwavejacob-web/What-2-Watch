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
      className="relative max-w-md mx-auto p-8 rounded-lg glass-card border border-neon-pink/20"
    >
      {/* Content */}
      <div className="relative text-center">
        {/* Icon */}
        <motion.div
          animate={{
            rotate: [0, 3, -3, 0],
            scale: [1, 1.03, 1],
          }}
          transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
          className="inline-flex items-center justify-center w-16 h-16 mb-5 rounded-lg border border-neon-pink/20"
          style={{ boxShadow: '0 0 20px rgba(255, 0, 110, 0.15)' }}
        >
          <Icon className="w-8 h-8 text-neon-pink/70" />
        </motion.div>

        {/* Title */}
        <h3 className="text-lg font-bold font-mono text-neon-pink mb-1">
          {title}
        </h3>

        {/* Error code */}
        <p className="font-mono text-muted/30 text-[10px] tracking-wider mb-4">
          ERR_{Math.random().toString(36).substring(2, 8).toUpperCase()}
        </p>

        {/* Error message */}
        <div className="p-3 bg-void/50 rounded border-l-2 border-neon-pink/30 mb-6 text-left">
          <p className="text-muted-light/70 text-xs font-mono break-words leading-relaxed">
            {error || 'An unexpected error occurred. The matrix has you.'}
          </p>
        </div>

        {/* Retry button */}
        {onRetry && (
          <motion.button
            onClick={onRetry}
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            className="
              inline-flex items-center gap-2 px-5 py-2.5 rounded
              bg-neon-pink/10 text-neon-pink border border-neon-pink/30
              font-mono font-bold text-xs tracking-wider
              hover:bg-neon-pink/20 hover:border-neon-pink/50
              transition-all duration-300
            "
          >
            <RefreshCw className="w-3.5 h-3.5" />
            RETRY
          </motion.button>
        )}
      </div>
    </motion.div>
  )
}
