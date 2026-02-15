import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Search, Zap } from 'lucide-react'
import { getRandomPlaceholder } from '../lib/vibes'

export default function SearchBar({ onSearch, isLoading }) {
  const [query, setQuery] = useState('')
  const [placeholder, setPlaceholder] = useState('')
  const [isFocused, setIsFocused] = useState(false)
  const inputRef = useRef(null)

  useEffect(() => {
    setPlaceholder(getRandomPlaceholder())
    const interval = setInterval(() => {
      if (!isFocused && !query) {
        setPlaceholder(getRandomPlaceholder())
      }
    }, 4000)
    return () => clearInterval(interval)
  }, [isFocused, query])

  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === '/' && document.activeElement !== inputRef.current) {
        e.preventDefault()
        inputRef.current?.focus()
      }
      if (e.key === 'Escape') {
        inputRef.current?.blur()
        setQuery('')
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  const handleSubmit = (e) => {
    e.preventDefault()
    if (query.trim() && !isLoading) {
      onSearch(query.trim())
    }
  }

  return (
    <div className="w-full max-w-3xl mx-auto">
      {/* Title */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center mb-10"
      >
        <h1 className="text-5xl md:text-7xl font-black font-mono tracking-tighter mb-1">
          <span className="text-neon-cyan text-glow-cyan">WHAT</span>
          <span className="text-warm-white mx-2 opacity-60">2</span>
          <span className="text-neon-pink text-glow-pink">WATCH</span>
        </h1>
        <p className="text-muted font-mono text-xs tracking-[0.3em] mt-2">
          DIGITAL CURATOR v2.049
        </p>
      </motion.div>

      {/* Search Container */}
      <motion.form
        onSubmit={handleSubmit}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ delay: 0.2 }}
        className="relative"
      >
        {/* Glow effect behind input */}
        <AnimatePresence>
          {(isFocused || query) && (
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.9 }}
              transition={{ duration: 0.3, ease: 'easeOut' }}
              className="absolute inset-0 -z-10 rounded-lg"
              style={{
                background: 'radial-gradient(ellipse at center, rgba(0,255,255,0.1) 0%, transparent 70%)',
                filter: 'blur(24px)',
              }}
            />
          )}
        </AnimatePresence>

        {/* Main input container */}
        <div
          className={`
            relative flex items-center gap-3 p-3 md:p-4
            bg-void-lighter/80 backdrop-blur-sm rounded-lg
            border transition-all duration-300 ease-out
            ${isFocused ? 'border-neon-cyan/50 box-glow-cyan' : 'border-white/[0.08]'}
            ${isLoading ? 'animate-pulse' : ''}
          `}
        >
          {/* Search icon */}
          <div className={`transition-colors duration-300 flex-shrink-0 ${isFocused ? 'text-neon-cyan' : 'text-muted'}`}>
            {isLoading ? (
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
              >
                <Zap className="w-5 h-5" />
              </motion.div>
            ) : (
              <Search className="w-5 h-5" />
            )}
          </div>

          {/* Input field */}
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            placeholder={placeholder}
            disabled={isLoading}
            className="
              flex-1 bg-transparent text-warm-white text-base md:text-lg
              font-sans placeholder:text-muted/40 placeholder:italic
              focus:outline-none disabled:opacity-50
            "
          />

          {/* Keyboard shortcut hint */}
          <div className="hidden md:flex items-center text-muted/40 text-xs font-mono flex-shrink-0">
            <kbd className="px-1.5 py-0.5 bg-void rounded border border-white/[0.06] text-[10px]">/</kbd>
          </div>

          {/* Submit button */}
          <motion.button
            type="submit"
            disabled={!query.trim() || isLoading}
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            className={`
              px-5 py-2 font-mono font-bold text-xs tracking-wider rounded
              transition-all duration-300 flex-shrink-0
              ${query.trim() && !isLoading
                ? 'bg-neon-cyan text-void hover:bg-neon-pink'
                : 'bg-white/[0.04] text-muted/40 cursor-not-allowed'
              }
            `}
          >
            {isLoading ? 'SCANNING...' : 'SEARCH'}
          </motion.button>
        </div>
      </motion.form>

      {/* Hint text */}
      <motion.p
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.4 }}
        className="text-center text-muted/50 text-xs mt-4 font-mono tracking-wider"
      >
        DESCRIBE YOUR VIBE. BE SPECIFIC. BE WEIRD. WE UNDERSTAND.
      </motion.p>
    </div>
  )
}
