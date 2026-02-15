import { motion, AnimatePresence } from 'framer-motion'
import { X, Compass, Download, Trash2 } from 'lucide-react'
import { mediaTypeColors, mediaTypeLabels } from '../lib/vibes'

function formatDate(dateString) {
  if (!dateString) return null
  try {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now - date
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) return 'Today'
    if (diffDays === 1) return 'Yesterday'
    if (diffDays < 7) return `${diffDays}d ago`
    if (diffDays < 30) return `${Math.floor(diffDays / 7)}w ago`
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  } catch {
    return null
  }
}

export default function WatchHistory({
  isOpen,
  onClose,
  history,
  onFindSimilar,
  onRemove,
  isLoading,
}) {
  const handleExport = () => {
    const dataStr = JSON.stringify(history, null, 2)
    const blob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `vibe-history-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            onClick={onClose}
            className="fixed inset-0 bg-void/80 backdrop-blur-sm z-40"
          />

          {/* Drawer */}
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 30, stiffness: 300 }}
            className="fixed top-0 right-0 h-full w-full max-w-md bg-void-light border-l border-white/[0.06] z-50 overflow-hidden flex flex-col"
          >
            {/* Header */}
            <div className="p-5 border-b border-white/[0.06] flex-shrink-0">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-bold font-mono text-neon-cyan">
                    WATCH HISTORY
                  </h2>
                  <p className="text-muted/60 text-xs font-mono mt-0.5">
                    {history?.length || 0} entries logged
                  </p>
                </div>
                <motion.button
                  onClick={onClose}
                  whileHover={{ scale: 1.1, rotate: 90 }}
                  whileTap={{ scale: 0.9 }}
                  className="p-1.5 text-muted/60 hover:text-neon-pink transition-colors rounded"
                >
                  <X className="w-5 h-5" />
                </motion.button>
              </div>

              {/* Actions */}
              {history?.length > 0 && (
                <div className="flex items-center gap-2 mt-3">
                  <motion.button
                    onClick={handleExport}
                    whileHover={{ scale: 1.03 }}
                    whileTap={{ scale: 0.97 }}
                    className="flex items-center gap-1.5 px-2.5 py-1.5 text-[10px] font-mono font-bold tracking-wider rounded border border-white/[0.08] text-muted/60 hover:border-neon-green/40 hover:text-neon-green transition-all duration-300"
                  >
                    <Download className="w-3 h-3" />
                    EXPORT
                  </motion.button>
                </div>
              )}
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-4">
              {isLoading ? (
                <div className="flex items-center justify-center h-40">
                  <motion.div
                    animate={{ rotate: 360 }}
                    transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                    className="w-6 h-6 border-2 border-neon-cyan/40 border-t-neon-cyan rounded-full"
                  />
                </div>
              ) : history?.length === 0 ? (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="text-center py-20"
                >
                  <div className="text-4xl mb-4 font-mono text-muted/20">
                    {'{ }'}
                  </div>
                  <p className="text-muted/60 font-mono text-sm mb-1">ARCHIVE EMPTY</p>
                  <p className="text-muted/40 text-xs">
                    Mark something as seen to start building your history.
                  </p>
                </motion.div>
              ) : (
                <div className="space-y-2">
                  {history?.map((item, index) => {
                    const typeColor = mediaTypeColors[item.media_type] || '#00FFFF'
                    const watchedDate = formatDate(item.watched_at || item.created_at)

                    return (
                      <motion.div
                        key={item.id || index}
                        initial={{ opacity: 0, x: 15 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.03 }}
                        className="group relative p-3 bg-void/50 rounded-lg border border-white/[0.04] hover:border-white/[0.1] transition-all duration-300"
                      >
                        {/* Left accent */}
                        <div
                          className="absolute left-0 top-2 bottom-2 w-0.5 rounded-full"
                          style={{ backgroundColor: typeColor }}
                        />

                        <div className="flex items-start justify-between gap-3 pl-3">
                          <div className="flex-1 min-w-0">
                            {/* Type + Year row */}
                            <div className="flex items-center gap-2 mb-1">
                              <span
                                className="text-[9px] font-mono font-bold tracking-wider px-1.5 py-0.5 rounded-full"
                                style={{
                                  backgroundColor: `${typeColor}12`,
                                  color: typeColor,
                                }}
                              >
                                {mediaTypeLabels[item.media_type] || 'MEDIA'}
                              </span>
                              {item.year && (
                                <span className="text-muted/40 text-[10px] font-mono">{item.year}</span>
                              )}
                              {watchedDate && (
                                <span className="text-muted/30 text-[10px] font-mono ml-auto">{watchedDate}</span>
                              )}
                            </div>

                            {/* Title */}
                            <h4 className="font-semibold text-warm-white text-sm truncate">
                              {item.title}
                            </h4>
                          </div>

                          {/* Action buttons */}
                          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex-shrink-0">
                            <motion.button
                              onClick={() => {
                                onFindSimilar(item.id)
                                onClose()
                              }}
                              whileHover={{ scale: 1.1 }}
                              whileTap={{ scale: 0.9 }}
                              className="p-1.5 text-muted/40 hover:text-neon-pink transition-colors rounded"
                              title="Find similar"
                            >
                              <Compass className="w-4 h-4" />
                            </motion.button>
                            {onRemove && (
                              <motion.button
                                onClick={() => onRemove(item.id)}
                                whileHover={{ scale: 1.1 }}
                                whileTap={{ scale: 0.9 }}
                                className="p-1.5 text-muted/40 hover:text-neon-pink transition-colors rounded"
                                title="Remove from history"
                              >
                                <Trash2 className="w-4 h-4" />
                              </motion.button>
                            )}
                          </div>
                        </div>
                      </motion.div>
                    )
                  })}
                </div>
              )}
            </div>

            {/* Bottom fade */}
            <div className="absolute bottom-0 left-0 right-0 h-16 bg-gradient-to-t from-void-light to-transparent pointer-events-none" />
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}
