import { motion } from 'framer-motion'
import { Search, Sparkles, Zap, X, Trash2 } from 'lucide-react'

export default function EmptyState({
  type = 'initial',
  searchHistory = [],
  onSearchHistoryClick,
  onClearSearchHistoryItem,
  onClearAllSearchHistory,
}) {
  if (type === 'no-results') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center py-20"
      >
        <motion.div
          animate={{ rotate: [0, 8, -8, 0] }}
          transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
          className="inline-block text-5xl mb-6 font-mono"
        >
          <span className="text-neon-pink/60">:/</span>
        </motion.div>
        <h3 className="text-xl font-bold font-mono text-warm-white mb-2">
          NO SIGNAL DETECTED
        </h3>
        <p className="text-muted/60 text-sm max-w-sm mx-auto">
          Couldn't find anything matching that vibe. Try being more specific,
          or explore a different aesthetic dimension.
        </p>
      </motion.div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: 0.6 }}
      className="text-center py-12"
    >
      {/* Search history chips */}
      {searchHistory.length > 0 && (
        <div className="mb-14">
          <div className="flex items-center justify-center gap-3 mb-4">
            <p className="text-muted/40 font-mono text-[10px] tracking-wider">
              RECENT SEARCHES
            </p>
            {onClearAllSearchHistory && searchHistory.length > 1 && (
              <button
                onClick={onClearAllSearchHistory}
                className="flex items-center gap-1 text-muted/30 hover:text-neon-pink/60 font-mono text-[10px] tracking-wider transition-colors"
              >
                <Trash2 className="w-3 h-3" />
                CLEAR ALL
              </button>
            )}
          </div>
          <div className="flex flex-wrap justify-center gap-2 max-w-2xl mx-auto">
            {searchHistory.slice(0, 8).map((entry, i) => (
              <motion.div
                key={entry.query}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.7 + i * 0.04 }}
                className="group flex items-center gap-1.5"
              >
                <button
                  onClick={() => onSearchHistoryClick?.(entry.query)}
                  className="px-3 py-1.5 bg-void-lighter/60 border border-white/[0.06] rounded
                    text-muted-light/70 text-xs font-mono
                    hover:border-neon-cyan/30 hover:text-neon-cyan
                    transition-all duration-300 cursor-pointer"
                >
                  "{entry.query}"
                </button>
                {onClearSearchHistoryItem && (
                  <button
                    onClick={() => onClearSearchHistoryItem(entry.query)}
                    className="p-0.5 text-muted/20 hover:text-neon-pink/60 transition-colors opacity-0 group-hover:opacity-100"
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </motion.div>
            ))}
          </div>
        </div>
      )}

      {/* Vibe suggestions */}
      <div className="mb-14">
        <p className="text-muted/40 font-mono text-[10px] mb-5 tracking-wider">
          TRY SOMETHING LIKE
        </p>
        <div className="flex flex-wrap justify-center gap-2 max-w-2xl mx-auto">
          {[
            'existential dread but make it beautiful',
            'cozy apocalypse vibes',
            'neon noir with heart',
            'psychological slow burn',
            'found family in space',
          ].map((suggestion, i) => (
            <motion.button
              key={suggestion}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.8 + i * 0.06 }}
              onClick={() => onSearchHistoryClick?.(suggestion)}
              className="px-3.5 py-2 bg-void-lighter/40 border border-white/[0.04] rounded
                text-muted/50 text-xs font-mono
                hover:border-neon-cyan/20 hover:text-neon-cyan/70
                transition-all duration-300 cursor-pointer"
            >
              "{suggestion}"
            </motion.button>
          ))}
        </div>
      </div>

      {/* Feature highlights */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-3xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.2 }}
          className="p-5 border border-white/[0.04] rounded-lg hover:border-neon-cyan/20 transition-colors duration-300"
        >
          <Search className="w-6 h-6 text-neon-cyan/60 mx-auto mb-3" />
          <h4 className="font-mono font-bold text-warm-white/80 text-xs tracking-wider mb-1.5">VIBE SEARCH</h4>
          <p className="text-muted/40 text-xs leading-relaxed">
            Describe the feeling you want. We'll decode your taste.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.3 }}
          className="p-5 border border-white/[0.04] rounded-lg hover:border-neon-green/20 transition-colors duration-300"
        >
          <Sparkles className="w-6 h-6 text-neon-green/60 mx-auto mb-3" />
          <h4 className="font-mono font-bold text-warm-white/80 text-xs tracking-wider mb-1.5">HIDDEN GEMS</h4>
          <p className="text-muted/40 text-xs leading-relaxed">
            Discover high-quality obscure titles you've never heard of.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.4 }}
          className="p-5 border border-white/[0.04] rounded-lg hover:border-neon-pink/20 transition-colors duration-300"
        >
          <Zap className="w-6 h-6 text-neon-pink/60 mx-auto mb-3" />
          <h4 className="font-mono font-bold text-warm-white/80 text-xs tracking-wider mb-1.5">AI POWERED</h4>
          <p className="text-muted/40 text-xs leading-relaxed">
            Embeddings + LLM reranking for uncanny accuracy.
          </p>
        </motion.div>
      </div>
    </motion.div>
  )
}
