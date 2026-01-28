import { useState } from 'react'
import { motion } from 'framer-motion'
import { Eye, EyeOff, Sparkles, ExternalLink, Compass } from 'lucide-react'
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

  // Parse the item data - adapt to your API response structure
  const {
    id,
    title,
    year,
    type = 'movie',
    vibe_profile,
    match_score,
    quality_score,
    is_hidden_gem,
  } = item

  const typeColor = mediaTypeColors[type] || mediaTypeColors.movie
  const typeLabel = mediaTypeLabels[type] || 'MEDIA'

  const handleMarkSeen = async () => {
    if (isMarking || isSeen) return
    setIsMarking(true)

    try {
      await onMarkSeen(id)
      setJustMarked(true)
      setTimeout(() => setJustMarked(false), 1000)
    } catch (err) {
      console.error('Failed to mark as seen:', err)
    } finally {
      setIsMarking(false)
    }
  }

  const handleFindSimilar = () => {
    onFindSimilar(id)
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 50, scale: 0.9 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 25,
        delay: index * 0.1,
      }}
      whileHover={{
        y: -5,
        transition: { duration: 0.2 }
      }}
      onHoverStart={() => setIsHovered(true)}
      onHoverEnd={() => setIsHovered(false)}
      className="relative group"
      style={{
        transformStyle: 'preserve-3d',
        perspective: '1000px',
      }}
    >
      {/* Card container */}
      <div
        className={`
          relative p-6 bg-void-lighter border-2 transition-all duration-300
          ${isHovered ? 'border-neon-cyan' : 'border-muted/20'}
          ${isSeen ? 'opacity-60' : ''}
        `}
        style={{
          transform: isHovered ? 'rotateX(2deg) rotateY(-2deg)' : 'none',
          boxShadow: isHovered
            ? `0 0 30px ${typeColor}30, inset 0 0 30px ${typeColor}05`
            : 'none',
        }}
      >
        {/* Scanline effect on hover */}
        {isHovered && (
          <div className="absolute inset-0 pointer-events-none overflow-hidden">
            <motion.div
              initial={{ y: '-100%' }}
              animate={{ y: '200%' }}
              transition={{ duration: 1.5, repeat: Infinity, ease: 'linear' }}
              className="absolute inset-x-0 h-8 bg-gradient-to-b from-transparent via-neon-cyan/10 to-transparent"
            />
          </div>
        )}

        {/* Top row: Type badge + Year + Quality */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            {/* Type badge */}
            <span
              className="px-3 py-1 text-xs font-mono font-bold tracking-wider"
              style={{
                backgroundColor: `${typeColor}20`,
                color: typeColor,
                border: `1px solid ${typeColor}`,
              }}
            >
              {typeLabel}
            </span>

            {/* Year */}
            {year && (
              <span className="text-muted text-sm font-mono">
                {year}
              </span>
            )}
          </div>

          {/* Hidden gem indicator */}
          {is_hidden_gem && (
            <motion.div
              animate={{ rotate: [0, 10, -10, 0] }}
              transition={{ duration: 2, repeat: Infinity }}
              className="flex items-center gap-1 text-neon-green text-xs font-mono"
            >
              <Sparkles className="w-4 h-4" />
              <span>HIDDEN GEM</span>
            </motion.div>
          )}
        </div>

        {/* Title */}
        <h3 className="text-xl md:text-2xl font-bold text-warm-white mb-4 leading-tight">
          {title}
        </h3>

        {/* Vibe Profile - the star of the show */}
        {vibe_profile && (
          <div className="mb-6 p-4 bg-void border-l-4 border-neon-pink">
            <p className="text-muted-light italic leading-relaxed">
              "{vibe_profile}"
            </p>
          </div>
        )}

        {/* Match score bar */}
        {match_score !== undefined && (
          <div className="mb-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-mono text-muted tracking-wider">VIBE MATCH</span>
              <span className="text-sm font-mono font-bold text-neon-cyan">
                {Math.round(match_score * 100)}%
              </span>
            </div>
            <div className="h-2 bg-void border border-muted/30 overflow-hidden">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${match_score * 100}%` }}
                transition={{ duration: 0.8, delay: index * 0.1 + 0.3 }}
                className="h-full"
                style={{
                  background: `linear-gradient(90deg, ${typeColor}, #FF006E)`,
                  boxShadow: `0 0 10px ${typeColor}`,
                }}
              />
            </div>
          </div>
        )}

        {/* Quality score if hidden gem */}
        {quality_score !== undefined && (
          <div className="flex items-center gap-2 mb-4">
            <span className="text-xs font-mono text-muted">QUALITY:</span>
            <div className="flex gap-1">
              {[...Array(5)].map((_, i) => (
                <div
                  key={i}
                  className="w-2 h-2"
                  style={{
                    backgroundColor: i < Math.round(quality_score * 5) ? '#39FF14' : '#333',
                    boxShadow: i < Math.round(quality_score * 5) ? '0 0 5px #39FF14' : 'none',
                  }}
                />
              ))}
            </div>
          </div>
        )}

        {/* Action buttons */}
        <div className="flex items-center gap-3 mt-4 pt-4 border-t border-muted/20">
          {/* Mark as seen button */}
          <motion.button
            onClick={handleMarkSeen}
            disabled={isMarking || isSeen}
            whileHover={{ scale: isSeen ? 1 : 1.05 }}
            whileTap={{ scale: isSeen ? 1 : 0.95 }}
            className={`
              flex items-center gap-2 px-4 py-2 font-mono text-sm
              transition-all duration-300
              ${isSeen
                ? 'bg-neon-green/20 text-neon-green border border-neon-green/50'
                : justMarked
                  ? 'bg-neon-green text-void animate-pulse'
                  : 'border border-muted/30 text-muted hover:border-neon-green hover:text-neon-green'
              }
              ${isMarking ? 'animate-glitch' : ''}
            `}
          >
            {isSeen ? (
              <>
                <Eye className="w-4 h-4" />
                <span>SEEN</span>
              </>
            ) : isMarking ? (
              <>
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 0.5, repeat: Infinity }}
                >
                  <Eye className="w-4 h-4" />
                </motion.div>
                <span>LOGGING...</span>
              </>
            ) : (
              <>
                <EyeOff className="w-4 h-4" />
                <span>MARK SEEN</span>
              </>
            )}
          </motion.button>

          {/* Find similar button */}
          <motion.button
            onClick={handleFindSimilar}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="
              flex items-center gap-2 px-4 py-2 font-mono text-sm
              border border-muted/30 text-muted
              hover:border-neon-pink hover:text-neon-pink
              transition-all duration-300
            "
          >
            <Compass className="w-4 h-4" />
            <span>SIMILAR</span>
          </motion.button>
        </div>

        {/* Decorative corner elements */}
        <div
          className="absolute top-0 right-0 w-0 h-0 transition-all duration-300"
          style={{
            borderTop: `20px solid ${isHovered ? typeColor : 'transparent'}`,
            borderLeft: '20px solid transparent',
          }}
        />
      </div>
    </motion.div>
  )
}
