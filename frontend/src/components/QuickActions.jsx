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
      initial={{ opacity: 0, y: 15 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.5, ease: 'easeOut' }}
      className="flex flex-col md:flex-row items-center justify-center gap-4 mt-10"
    >
      {/* Main action buttons */}
      <div className="flex items-center gap-2">
        {/* Hidden Gems */}
        <motion.button
          onClick={onHiddenGems}
          disabled={isLoading}
          whileHover={{ scale: 1.04 }}
          whileTap={{ scale: 0.96 }}
          className="
            group relative flex items-center gap-2 px-4 py-2.5
            bg-transparent border border-neon-green/30 text-neon-green/80
            font-mono text-xs font-bold tracking-wider rounded
            hover:bg-neon-green/10 hover:border-neon-green/60 hover:text-neon-green
            disabled:opacity-40 disabled:cursor-not-allowed
            transition-all duration-300 ease-out
          "
        >
          <motion.div
            animate={{ rotate: [0, 12, -12, 0] }}
            transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
          >
            <Sparkles className="w-3.5 h-3.5" />
          </motion.div>
          <span>HIDDEN GEMS</span>
        </motion.button>

        {/* Surprise Me */}
        <motion.button
          onClick={onSurpriseMe}
          disabled={isLoading}
          whileHover={{ scale: 1.04 }}
          whileTap={{ scale: 0.96 }}
          className="
            group relative flex items-center gap-2 px-4 py-2.5
            bg-transparent border border-neon-pink/30 text-neon-pink/80
            font-mono text-xs font-bold tracking-wider rounded
            hover:bg-neon-pink/10 hover:border-neon-pink/60 hover:text-neon-pink
            disabled:opacity-40 disabled:cursor-not-allowed
            transition-all duration-300 ease-out
          "
        >
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 4, repeat: Infinity, ease: 'linear' }}
          >
            <Shuffle className="w-3.5 h-3.5" />
          </motion.div>
          <span>SURPRISE ME</span>
        </motion.button>

        {/* Watch History */}
        <motion.button
          onClick={onToggleHistory}
          disabled={isLoading}
          whileHover={{ scale: 1.04 }}
          whileTap={{ scale: 0.96 }}
          className="
            group relative flex items-center gap-2 px-4 py-2.5
            bg-transparent border border-neon-cyan/30 text-neon-cyan/80
            font-mono text-xs font-bold tracking-wider rounded
            hover:bg-neon-cyan/10 hover:border-neon-cyan/60 hover:text-neon-cyan
            disabled:opacity-40 disabled:cursor-not-allowed
            transition-all duration-300 ease-out
          "
        >
          <History className="w-3.5 h-3.5" />
          <span>HISTORY</span>
        </motion.button>
      </div>

      {/* Divider */}
      <div className="hidden md:block w-px h-6 bg-white/[0.08]" />

      {/* Filter toggles */}
      <div className="flex items-center gap-0.5 p-1 bg-void-lighter/60 border border-white/[0.06] rounded-lg">
        {filterOptions.map((option) => (
          <motion.button
            key={option.id}
            onClick={() => onFilterChange(option.id)}
            whileTap={{ scale: 0.97 }}
            className={`
              flex items-center gap-1.5 px-3.5 py-1.5 font-mono text-[11px] font-bold tracking-wider
              rounded transition-all duration-250 ease-out
              ${activeFilter === option.id
                ? 'bg-neon-cyan/15 text-neon-cyan'
                : 'text-muted/60 hover:text-warm-white'
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
