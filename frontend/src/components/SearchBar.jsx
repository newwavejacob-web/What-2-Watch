import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Search, Zap, Command } from 'lucide-react'
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
        className="text-center mb-8"
      >
        <h1 className="text-5xl md:text-7xl font-bold font-mono tracking-tighter mb-2">
          <span className="text-neon-cyan text-glow-cyan">VIBE</span>
          <span className="text-warm-white">.</span>
          <span className="text-neon-pink text-glow-pink">SEARCH</span>
        </h1>
        <p className="text-muted-light font-mono text-sm tracking-widest">
          // DIGITAL CURATOR v2.049
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
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8 }}
              className="absolute inset-0 -z-10"
              style={{
                background: 'radial-gradient(ellipse at center, rgba(0,255,255,0.15) 0%, transparent 70%)',
                filter: 'blur(20px)',
              }}
            />
          )}
        </AnimatePresence>

        {/* Ripple effect when typing */}
        <AnimatePresence>
          {query && (
            <motion.div
              key={query.length}
              initial={{ scale: 1, opacity: 0.3 }}
              animate={{ scale: 2, opacity: 0 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.6 }}
              className="absolute inset-0 border-2 border-neon-cyan -z-10"
              style={{ borderRadius: 0 }}
            />
          )}
        </AnimatePresence>

        {/* Main input container */}
        <div
          className={`
            relative flex items-center gap-4 p-4 md:p-6
            bg-void-lighter border-2 transition-all duration-300
            ${isFocused ? 'border-neon-cyan box-glow-cyan' : 'border-muted/30'}
            ${isLoading ? 'animate-pulse' : ''}
          `}
        >
          {/* Search icon */}
          <div className={`transition-colors duration-300 ${isFocused ? 'text-neon-cyan' : 'text-muted'}`}>
            {isLoading ? (
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
              >
                <Zap className="w-6 h-6" />
              </motion.div>
            ) : (
              <Search className="w-6 h-6" />
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
              flex-1 bg-transparent text-warm-white text-lg md:text-xl
              font-sans placeholder:text-muted/50 placeholder:italic
              focus:outline-none disabled:opacity-50
            "
          />

          {/* Keyboard shortcut hint */}
          <div className="hidden md:flex items-center gap-1 text-muted text-sm font-mono">
            <kbd className="px-2 py-1 bg-void border border-muted/30 text-xs">/</kbd>
            <span>to focus</span>
          </div>

          {/* Submit button */}
          <motion.button
            type="submit"
            disabled={!query.trim() || isLoading}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className={`
              px-6 py-3 font-mono font-bold text-sm tracking-wider
              transition-all duration-300
              ${query.trim() && !isLoading
                ? 'bg-neon-cyan text-void hover:bg-neon-pink'
                : 'bg-muted/20 text-muted cursor-not-allowed'
              }
            `}
          >
            {isLoading ? 'SCANNING...' : 'SEARCH'}
          </motion.button>
        </div>

        {/* Decorative corner accents */}
        <div className="absolute top-0 left-0 w-4 h-4 border-t-2 border-l-2 border-neon-cyan -translate-x-1 -translate-y-1" />
        <div className="absolute top-0 right-0 w-4 h-4 border-t-2 border-r-2 border-neon-cyan translate-x-1 -translate-y-1" />
        <div className="absolute bottom-0 left-0 w-4 h-4 border-b-2 border-l-2 border-neon-pink -translate-x-1 translate-y-1" />
        <div className="absolute bottom-0 right-0 w-4 h-4 border-b-2 border-r-2 border-neon-pink translate-x-1 translate-y-1" />
      </motion.form>

      {/* Hint text */}
      <motion.p
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.4 }}
        className="text-center text-muted text-sm mt-4 font-mono"
      >
        DESCRIBE YOUR VIBE. BE SPECIFIC. BE WEIRD. WE UNDERSTAND.
      </motion.p>
    </div>
  )
}
