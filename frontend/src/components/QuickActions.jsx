import { motion } from 'framer-motion'
import { Sparkles, Shuffle, Film, Tv, Play, History } from 'lucide-react'

const filterOptions = [
  { id: 'all', label: 'ALL', icon: null },
  { id: 'movie', label: 'FILMS', icon: Film },
  { id: 'tv', label: 'SERIES', icon: Tv },
  { id: 'anime', label: 'ANIME', icon: Play },
]

export default function QuickActions({
  onHiddenGems,
  onSurpriseMe,
  onToggleHistory,
  activeFilter,
  onFilterChange,
  isLoading,
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.5 }}
      className="flex flex-col md:flex-row items-center justify-center gap-4 mt-8"
    >
      {/* Main action buttons */}
      <div className="flex items-center gap-3">
        {/* Hidden Gems */}
        <motion.button
          onClick={onHiddenGems}
          disabled={isLoading}
          whileHover={{ scale: 1.05, y: -2 }}
          whileTap={{ scale: 0.95 }}
          className="
            group relative flex items-center gap-2 px-5 py-3
            bg-void border-2 border-neon-green text-neon-green
            font-mono text-sm font-bold tracking-wider
            hover:bg-neon-green hover:text-void
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-all duration-300
          "
        >
          <motion.div
            animate={{ rotate: [0, 15, -15, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            <Sparkles className="w-4 h-4" />
          </motion.div>
          <span>HIDDEN GEMS</span>

          {/* Glow effect on hover */}
          <div className="
            absolute inset-0 opacity-0 group-hover:opacity-100
            transition-opacity duration-300 pointer-events-none
          " style={{
            boxShadow: '0 0 20px rgba(57, 255, 20, 0.5), inset 0 0 20px rgba(57, 255, 20, 0.1)',
          }} />
        </motion.button>

        {/* Surprise Me */}
        <motion.button
          onClick={onSurpriseMe}
          disabled={isLoading}
          whileHover={{ scale: 1.05, y: -2 }}
          whileTap={{ scale: 0.95 }}
          className="
            group relative flex items-center gap-2 px-5 py-3
            bg-void border-2 border-neon-pink text-neon-pink
            font-mono text-sm font-bold tracking-wider
            hover:bg-neon-pink hover:text-void
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-all duration-300
          "
        >
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 4, repeat: Infinity, ease: 'linear' }}
          >
            <Shuffle className="w-4 h-4" />
          </motion.div>
          <span>SURPRISE ME</span>

          <div className="
            absolute inset-0 opacity-0 group-hover:opacity-100
            transition-opacity duration-300 pointer-events-none
          " style={{
            boxShadow: '0 0 20px rgba(255, 0, 110, 0.5), inset 0 0 20px rgba(255, 0, 110, 0.1)',
          }} />
        </motion.button>

        {/* Watch History */}
        <motion.button
          onClick={onToggleHistory}
          disabled={isLoading}
          whileHover={{ scale: 1.05, y: -2 }}
          whileTap={{ scale: 0.95 }}
          className="
            group relative flex items-center gap-2 px-5 py-3
            bg-void border-2 border-neon-cyan text-neon-cyan
            font-mono text-sm font-bold tracking-wider
            hover:bg-neon-cyan hover:text-void
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-all duration-300
          "
        >
          <History className="w-4 h-4" />
          <span>HISTORY</span>

          <div className="
            absolute inset-0 opacity-0 group-hover:opacity-100
            transition-opacity duration-300 pointer-events-none
          " style={{
            boxShadow: '0 0 20px rgba(0, 255, 255, 0.5), inset 0 0 20px rgba(0, 255, 255, 0.1)',
          }} />
        </motion.button>
      </div>

      {/* Divider */}
      <div className="hidden md:block w-px h-8 bg-muted/30" />

      {/* Filter toggles */}
      <div className="flex items-center gap-1 p-1 bg-void-lighter border border-muted/30">
        {filterOptions.map((option) => (
          <motion.button
            key={option.id}
            onClick={() => onFilterChange(option.id)}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className={`
              flex items-center gap-1.5 px-4 py-2 font-mono text-xs font-bold tracking-wider
              transition-all duration-300
              ${activeFilter === option.id
                ? 'bg-neon-cyan text-void'
                : 'text-muted hover:text-warm-white'
              }
            `}
          >
            {option.icon && <option.icon className="w-3 h-3" />}
            <span>{option.label}</span>
          </motion.button>
        ))}
      </div>
    </motion.div>
  )
}
