import { motion, AnimatePresence } from 'framer-motion'
import { X, Compass, Download, Trash2 } from 'lucide-react'
import { mediaTypeColors, mediaTypeLabels } from '../lib/vibes'

export default function WatchHistory({
  isOpen,
  onClose,
  history,
  onFindSimilar,
  isLoading,
}) {
  const handleExport = () => {
    const dataStr = JSON.stringify(history, null, 2)
    const blob = new Blob([dataStr], { media_type: 'application/json' })
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
            onClick={onClose}
            className="fixed inset-0 bg-void/80 backdrop-blur-sm z-40"
          />

          {/* Drawer */}
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ media_type: 'spring', damping: 25, stiffness: 200 }}
            className="fixed top-0 right-0 h-full w-full max-w-md bg-void-light border-l-2 border-neon-cyan z-50 overflow-hidden"
          >
            {/* Scanlines effect */}
            <div className="absolute inset-0 pointer-events-none scanlines opacity-30" />

            {/* Header */}
            <div className="p-6 border-b border-muted/30">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-2xl font-bold font-mono text-neon-cyan text-glow-cyan">
                    WATCH HISTORY
                  </h2>
                  <p className="text-muted text-sm font-mono mt-1">
                    // {history?.length || 0} ENTRIES LOGGED
                  </p>
                </div>
                <motion.button
                  onClick={onClose}
                  whileHover={{ scale: 1.1, rotate: 90 }}
                  whileTap={{ scale: 0.9 }}
                  className="p-2 text-muted hover:text-neon-pink transition-colors"
                >
                  <X className="w-6 h-6" />
                </motion.button>
              </div>

              {/* Actions */}
              {history?.length > 0 && (
                <div className="flex items-center gap-2 mt-4">
                  <motion.button
                    onClick={handleExport}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    className="flex items-center gap-2 px-3 py-1.5 text-xs font-mono border border-muted/30 text-muted hover:border-neon-green hover:text-neon-green transition-colors"
                  >
                    <Download className="w-3 h-3" />
                    EXPORT JSON
                  </motion.button>
                </div>
              )}
            </div>

            {/* Content */}
            <div className="p-6 overflow-y-auto" style={{ height: 'calc(100% - 140px)' }}>
              {isLoading ? (
                <div className="flex items-center justify-center h-40">
                  <motion.div
                    animate={{ rotate: 360 }}
                    transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                    className="w-8 h-8 border-2 border-neon-cyan border-t-transparent"
                  />
                </div>
              ) : history?.length === 0 ? (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="text-center py-20"
                >
                  <div className="text-6xl mb-4">
                    <span className="text-glow-pink">{'{ }'}</span>
                  </div>
                  <p className="text-muted font-mono mb-2">ARCHIVE EMPTY</p>
                  <p className="text-muted-light text-sm">
                    Your vibe library is empty.
                    <br />
                    Let's find something sick.
                  </p>
                </motion.div>
              ) : (
                <div className="space-y-3">
                  {history?.map((item, index) => (
                    <motion.div
                      key={item.id || index}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.05 }}
                      className="group relative p-4 bg-void border border-muted/20 hover:border-neon-cyan transition-all duration-300"
                    >
                      {/* Type indicator */}
                      <div
                        className="absolute left-0 top-0 bottom-0 w-1"
                        style={{ backgroundColor: mediaTypeColors[item.media_type] || '#00FFFF' }}
                      />

                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1 min-w-0 pl-3">
                          {/* Type badge */}
                          <span
                            className="inline-block px-2 py-0.5 text-[10px] font-mono font-bold tracking-wider mb-2"
                            style={{
                              backgroundColor: `${mediaTypeColors[item.media_type]}20`,
                              color: mediaTypeColors[item.media_type],
                            }}
                          >
                            {mediaTypeLabels[item.media_type] || 'MEDIA'}
                          </span>

                          {/* Title */}
                          <h4 className="font-semibold text-warm-white truncate">
                            {item.title}
                          </h4>

                          {/* Year */}
                          {item.year && (
                            <p className="text-muted text-sm font-mono">{item.year}</p>
                          )}
                        </div>

                        {/* Find similar button */}
                        <motion.button
                          onClick={() => {
                            onFindSimilar(item.id)
                            onClose()
                          }}
                          whileHover={{ scale: 1.1 }}
                          whileTap={{ scale: 0.9 }}
                          className="p-2 text-muted hover:text-neon-pink opacity-0 group-hover:opacity-100 transition-all"
                          title="Find similar vibes"
                        >
                          <Compass className="w-5 h-5" />
                        </motion.button>
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </div>

            {/* Decorative elements */}
            <div className="absolute bottom-0 left-0 right-0 h-20 bg-gradient-to-t from-void-light to-transparent pointer-events-none" />
            <div className="absolute top-0 left-0 w-4 h-4 border-t-2 border-l-2 border-neon-cyan" />
            <div className="absolute bottom-0 left-0 w-4 h-4 border-b-2 border-l-2 border-neon-pink" />
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}
