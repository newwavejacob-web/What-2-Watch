import { useState } from 'react'
import { motion } from 'framer-motion'
import { Eye, EyeOff, Sparkles, Compass, ChevronDown, ChevronUp } from 'lucide-react'
import { mediaTypeColors, mediaTypeLabels } from '../lib/vibes'

export default function RecommendationCard({
  item,
  index,
  onMarkSeen,
  onFindSimilar,
  isSeen = false
}) {
  const [isMarking, setIsMarking] = useState(false)
  const [justMarked, setJustMarked] = useState(false)
  const [isHovered, setIsHovered] = useState(false)
  const [isExpanded, setIsExpanded] = useState(false)

  const {
    id,
    title,
    year,
    media_type = 'movie',
    vibe_profile,
    match_score,
    quality_score,
    is_hidden_gem,
  } = item

  const typeColor = mediaTypeColors[media_type] || mediaTypeColors.movie
  const typeLabel = mediaTypeLabels[media_type] || 'MEDIA'

  const handleMarkSeen = async () => {
    if (isMarking || isSeen) return
    setIsMarking(true)

    try {
      await onMarkSeen(id)
      setJustMarked(true)
      setTimeout(() => setJustMarked(false), 1500)
    } catch (err) {
      console.error('Failed to mark as seen:', err)
    } finally {
      setIsMarking(false)
    }
  }

  const handleFindSimilar = () => {
    onFindSimilar(id)
  }

  // Determine if vibe_profile is long enough to need truncation
  const needsTruncation = vibe_profile && vibe_profile.length > 150

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        type: 'spring',
        stiffness: 400,
        damping: 30,
        delay: index * 0.08,
      }}
      whileHover={{
        y: -4,
        transition: { duration: 0.25, ease: 'easeOut' }
      }}
      onHoverStart={() => setIsHovered(true)}
      onHoverEnd={() => setIsHovered(false)}
      className="relative group"
    >
      {/* Card container */}
      <div
        className={`
          relative overflow-hidden rounded-lg glass-card
          border transition-all duration-300 ease-out
          ${isHovered ? 'border-opacity-80' : 'border-white/[0.06]'}
          ${isSeen ? 'opacity-50' : ''}
        `}
        style={{
          borderColor: isHovered ? `${typeColor}88` : undefined,
          boxShadow: isHovered
            ? `0 8px 32px ${typeColor}15, 0 0 1px ${typeColor}30`
            : '0 2px 8px rgba(0,0,0,0.3)',
        }}
      >
        {/* Subtle top accent line */}
        <div
          className="h-[2px] w-full transition-opacity duration-300"
          style={{
            background: `linear-gradient(90deg, transparent, ${typeColor}, transparent)`,
            opacity: isHovered ? 0.8 : 0.2,
          }}
        />

        {/* Scanline effect on hover */}
        {isHovered && (
          <div className="absolute inset-0 pointer-events-none overflow-hidden">
            <motion.div
              initial={{ y: '-100%' }}
              animate={{ y: '200%' }}
              transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
              className="absolute inset-x-0 h-12 bg-gradient-to-b from-transparent via-white/[0.03] to-transparent"
            />
          </div>
        )}

        <div className="p-5 md:p-6">
          {/* Top row: Type pill + Year + Hidden gem */}
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2.5">
              {/* Type pill badge */}
              <span
                className="px-2.5 py-0.5 text-[10px] font-mono font-bold tracking-wider rounded-full"
                style={{
                  backgroundColor: `${typeColor}15`,
                  color: typeColor,
                  border: `1px solid ${typeColor}40`,
                }}
              >
                {typeLabel}
              </span>

              {/* Year */}
              {year && (
                <span className="text-muted text-xs font-mono">
                  {year}
                </span>
              )}

              {/* Seen badge */}
              {isSeen && (
                <span className="px-2 py-0.5 text-[10px] font-mono font-bold tracking-wider rounded-full bg-neon-green/10 text-neon-green border border-neon-green/30">
                  SEEN
                </span>
              )}
            </div>

            {/* Hidden gem indicator */}
            {is_hidden_gem && (
              <motion.div
                animate={{ rotate: [0, 8, -8, 0] }}
                transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
                className="flex items-center gap-1 text-neon-green"
              >
                <Sparkles className="w-3.5 h-3.5" />
                <span className="text-[10px] font-mono font-bold tracking-wider">GEM</span>
              </motion.div>
            )}
          </div>

          {/* Title */}
          <h3 className="text-lg md:text-xl font-bold text-warm-white mb-3 leading-snug tracking-tight">
            {title}
          </h3>

          {/* Vibe Profile quote */}
          {vibe_profile && (
            <div className="mb-4">
              <div
                className={`
                  relative pl-3 border-l-2 transition-colors duration-300
                  ${isHovered ? 'border-neon-pink/60' : 'border-muted/20'}
                `}
              >
                <p
                  className={`
                    text-muted-light text-sm leading-relaxed font-sans italic
                    ${needsTruncation && !isExpanded ? 'line-clamp-3' : ''}
                  `}
                >
                  "{vibe_profile}"
                </p>
                {needsTruncation && (
                  <button
                    onClick={() => setIsExpanded(!isExpanded)}
                    className="flex items-center gap-1 mt-1.5 text-[11px] font-mono text-muted hover:text-neon-cyan transition-colors duration-200"
                  >
                    {isExpanded ? (
                      <>
                        <ChevronUp className="w-3 h-3" />
                        COLLAPSE
                      </>
                    ) : (
                      <>
                        <ChevronDown className="w-3 h-3" />
                        READ MORE
                      </>
                    )}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* Match score bar */}
          {match_score !== undefined && (
            <div className="mb-4">
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-[10px] font-mono text-muted tracking-wider uppercase">Vibe Match</span>
                <span
                  className="text-xs font-mono font-bold"
                  style={{ color: typeColor }}
                >
                  {Math.round(match_score * 100)}%
                </span>
              </div>
              <div className="h-1 bg-white/[0.06] rounded-full overflow-hidden">
                <motion.div
                  initial={{ width: 0 }}
                  animate={{ width: `${match_score * 100}%` }}
                  transition={{ duration: 0.8, delay: index * 0.08 + 0.3, ease: 'easeOut' }}
                  className="h-full rounded-full"
                  style={{
                    background: `linear-gradient(90deg, ${typeColor}CC, ${typeColor})`,
                    boxShadow: `0 0 8px ${typeColor}60`,
                  }}
                />
              </div>
            </div>
          )}

          {/* Quality score dots */}
          {quality_score !== undefined && (
            <div className="flex items-center gap-2 mb-4">
              <span className="text-[10px] font-mono text-muted tracking-wider uppercase">Quality</span>
              <div className="flex gap-1">
                {[...Array(5)].map((_, i) => {
                  const filled = i < Math.round(quality_score * 5)
                  return (
                    <div
                      key={i}
                      className="w-1.5 h-1.5 rounded-full transition-all duration-300"
                      style={{
                        backgroundColor: filled ? '#39FF14' : 'rgba(255,255,255,0.08)',
                        boxShadow: filled ? '0 0 4px #39FF14' : 'none',
                      }}
                    />
                  )
                })}
              </div>
            </div>
          )}

          {/* Action buttons */}
          <div className="flex items-center gap-2 pt-3 border-t border-white/[0.06]">
            {/* Mark as seen button */}
            <motion.button
              onClick={handleMarkSeen}
              disabled={isMarking || isSeen}
              whileHover={{ scale: isSeen ? 1 : 1.03 }}
              whileTap={{ scale: isSeen ? 1 : 0.97 }}
              className={`
                flex items-center gap-1.5 px-3 py-1.5 font-mono text-xs rounded
                transition-all duration-300
                ${isSeen
                  ? 'bg-neon-green/10 text-neon-green/60 border border-neon-green/20 cursor-default'
                  : justMarked
                    ? 'bg-neon-green/20 text-neon-green border border-neon-green/40'
                    : 'border border-white/[0.08] text-muted hover:border-neon-green/40 hover:text-neon-green'
                }
                ${isMarking ? 'animate-pulse' : ''}
              `}
            >
              {isSeen ? (
                <>
                  <Eye className="w-3.5 h-3.5" />
                  <span>SEEN</span>
                </>
              ) : isMarking ? (
                <>
                  <motion.div
                    animate={{ rotate: 360 }}
                    transition={{ duration: 0.5, repeat: Infinity, ease: 'linear' }}
                  >
                    <Eye className="w-3.5 h-3.5" />
                  </motion.div>
                  <span>...</span>
                </>
              ) : (
                <>
                  <EyeOff className="w-3.5 h-3.5" />
                  <span>MARK SEEN</span>
                </>
              )}
            </motion.button>

            {/* Find similar button */}
            <motion.button
              onClick={handleFindSimilar}
              whileHover={{ scale: 1.03 }}
              whileTap={{ scale: 0.97 }}
              className="
                flex items-center gap-1.5 px-3 py-1.5 font-mono text-xs rounded
                border border-white/[0.08] text-muted
                hover:border-neon-pink/40 hover:text-neon-pink
                transition-all duration-300
              "
            >
              <Compass className="w-3.5 h-3.5" />
              <span>SIMILAR</span>
            </motion.button>
          </div>
        </div>
      </div>
    </motion.div>
  )
}
