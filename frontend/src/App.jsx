import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import ParticleBackground from './components/ParticleBackground'
import SearchBar from './components/SearchBar'
import RecommendationCard from './components/RecommendationCard'
import QuickActions from './components/QuickActions'
import WatchHistory from './components/WatchHistory'
import LoadingState from './components/LoadingState'
import ErrorDisplay from './components/ErrorDisplay'
import EmptyState from './components/EmptyState'
import { useApi } from './hooks/useApi'
import { useLocalStorage } from './hooks/useLocalStorage'
import { getRandomVibe } from './lib/vibes'

function Toast({ message, type = 'error', onDismiss }) {
  useEffect(() => {
    const timer = setTimeout(onDismiss, 3000)
    return () => clearTimeout(timer)
  }, [onDismiss])

  return (
    <motion.div
      initial={{ opacity: 0, y: 50, scale: 0.9 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: 20, scale: 0.9 }}
      className={`
        fixed bottom-6 right-6 z-50 px-4 py-3 rounded-lg
        font-mono text-xs tracking-wider backdrop-blur-sm
        border
        ${type === 'error'
          ? 'bg-neon-pink/10 border-neon-pink/30 text-neon-pink'
          : 'bg-neon-green/10 border-neon-green/30 text-neon-green'
        }
      `}
    >
      {message}
    </motion.div>
  )
}

export default function App() {
  const [results, setResults] = useState([])
  const [seenIds, setSeenIds] = useLocalStorage('vibe_seen_ids', [])
  const [searchHistory, setSearchHistory] = useLocalStorage('vibe_search_history', [])
  const [historyOpen, setHistoryOpen] = useState(false)
  const [watchHistory, setWatchHistory] = useState([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [activeFilter, setActiveFilter] = useState('all')
  const [currentQuery, setCurrentQuery] = useState('')
  const [hasSearched, setHasSearched] = useState(false)
  const [toast, setToast] = useState(null)

  const {
    loading,
    error,
    search,
    markAsSeen,
    removeSeen,
    getWatchHistory,
    getHiddenGems,
    getSimilar,
    clearError,
  } = useApi()

  // Sync seen IDs from backend on initial load
  useEffect(() => {
    const syncSeenIds = async () => {
      try {
        const data = await getWatchHistory()
        if (data.seen && Array.isArray(data.seen)) {
          const backendIds = data.seen.map(item => item.id || item.media_id).filter(Boolean)
          setSeenIds(prev => [...new Set([...prev, ...backendIds])])
        }
      } catch {
        // Backend unavailable — localStorage cache is fine
      }
    }
    syncSeenIds()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Filter results by media type
  const filteredResults = results.filter((item) => {
    if (activeFilter === 'all') return true
    return item.media_type === activeFilter
  })

  // Normalize a recommendation from the API into the flat structure RecommendationCard expects
  const normalizeRecommendation = (rec) => {
    if (rec.media) {
      return {
        ...rec.media,
        match_score: rec.vibe_score,
        explanation: rec.explanation,
        rank: rec.rank,
      }
    }
    return rec
  }

  // Handle search
  const handleSearch = useCallback(async (query) => {
    setCurrentQuery(query)
    setHasSearched(true)
    clearError()

    try {
      const data = await search(query)
      setResults((data.recommendations || []).map(normalizeRecommendation))

      // Add to search history
      setSearchHistory((prev) => {
        const newHistory = [
          { query, timestamp: Date.now() },
          ...prev.filter((h) => h.query !== query),
        ].slice(0, 20)
        return newHistory
      })
    } catch (err) {
      console.error('Search failed:', err)
      setResults([])
    }
  }, [search, clearError, setSearchHistory])

  // Handle mark as seen — optimistic update with rollback
  const handleMarkSeen = useCallback(async (mediaId) => {
    // Optimistic update
    setSeenIds((prev) => [...new Set([...prev, mediaId])])

    try {
      await markAsSeen(mediaId)
    } catch (err) {
      // Rollback on failure
      setSeenIds((prev) => prev.filter(id => id !== mediaId))
      setToast({ message: 'FAILED TO SYNC — RETRYING...', type: 'error' })
      console.error('Failed to mark as seen:', err)
      throw err
    }
  }, [markAsSeen, setSeenIds])

  // Handle remove from seen
  const handleRemoveSeen = useCallback(async (mediaId) => {
    try {
      await removeSeen(mediaId)
      setSeenIds((prev) => prev.filter(id => id !== mediaId))
      // Update watch history list
      setWatchHistory(prev => prev.filter(item => item.id !== mediaId))
    } catch (err) {
      setToast({ message: 'FAILED TO REMOVE', type: 'error' })
      console.error('Failed to remove from seen:', err)
    }
  }, [removeSeen, setSeenIds])

  // Handle find similar
  const handleFindSimilar = useCallback(async (mediaId) => {
    setHasSearched(true)
    clearError()

    try {
      const data = await getSimilar(mediaId)
      setResults((data.recommendations || []).map(normalizeRecommendation))
      setCurrentQuery(`Similar to: ${mediaId}`)
    } catch (err) {
      console.error('Failed to find similar:', err)
    }
  }, [getSimilar, clearError])

  // Handle hidden gems
  const handleHiddenGems = useCallback(async () => {
    setHasSearched(true)
    clearError()

    try {
      const data = await getHiddenGems()
      setResults((data.hidden_gems || []).map(gem => ({
        ...normalizeRecommendation(gem),
        is_hidden_gem: true,
      })))
      setCurrentQuery('Hidden Gems')
    } catch (err) {
      console.error('Failed to get hidden gems:', err)
    }
  }, [getHiddenGems, clearError])

  // Handle surprise me
  const handleSurpriseMe = useCallback(() => {
    const randomVibe = getRandomVibe()
    handleSearch(randomVibe)
  }, [handleSearch])

  // Toggle history drawer
  const handleToggleHistory = useCallback(async () => {
    if (!historyOpen) {
      setHistoryLoading(true)
      try {
        const data = await getWatchHistory()
        setWatchHistory(data.seen || [])
      } catch (err) {
        console.error('Failed to load watch history:', err)
        setWatchHistory([])
      } finally {
        setHistoryLoading(false)
      }
    }
    setHistoryOpen(!historyOpen)
  }, [historyOpen, getWatchHistory])

  // Clear a single search history entry
  const handleClearSearchHistoryItem = useCallback((query) => {
    setSearchHistory(prev => prev.filter(h => h.query !== query))
  }, [setSearchHistory])

  // Clear all search history
  const handleClearAllSearchHistory = useCallback(() => {
    setSearchHistory([])
  }, [setSearchHistory])

  // Retry on error
  const handleRetry = useCallback(() => {
    if (currentQuery) {
      handleSearch(currentQuery)
    } else {
      clearError()
    }
  }, [currentQuery, handleSearch, clearError])

  return (
    <div className="min-h-screen relative">
      {/* Particle background */}
      <ParticleBackground />

      {/* Main content */}
      <div className="relative z-10 px-4 py-8 md:py-16">
        <div className="max-w-6xl mx-auto">
          {/* Hero section with search */}
          <header className="mb-16">
            <SearchBar onSearch={handleSearch} isLoading={loading} />
            <QuickActions
              onHiddenGems={handleHiddenGems}
              onSurpriseMe={handleSurpriseMe}
              onToggleHistory={handleToggleHistory}
              activeFilter={activeFilter}
              onFilterChange={setActiveFilter}
              isLoading={loading}
            />
          </header>

          {/* Results section */}
          <main>
            {/* Current query indicator */}
            <AnimatePresence>
              {currentQuery && hasSearched && !loading && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="mb-8 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-muted/60 font-mono text-xs tracking-wider">QUERY:</span>
                    <span className="text-neon-cyan font-mono text-sm">"{currentQuery}"</span>
                  </div>
                  <span className="text-muted/60 font-mono text-xs tracking-wider">
                    {filteredResults.length} RESULTS
                  </span>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Loading state */}
            {loading && <LoadingState />}

            {/* Error state */}
            {error && !loading && (
              <ErrorDisplay error={error} onRetry={handleRetry} />
            )}

            {/* Empty state - initial with search history */}
            {!loading && !error && !hasSearched && (
              <EmptyState
                type="initial"
                searchHistory={searchHistory}
                onSearchHistoryClick={handleSearch}
                onClearSearchHistoryItem={handleClearSearchHistoryItem}
                onClearAllSearchHistory={handleClearAllSearchHistory}
              />
            )}

            {/* Empty state - no results */}
            {!loading && !error && hasSearched && filteredResults.length === 0 && (
              <EmptyState type="no-results" />
            )}

            {/* Results grid */}
            {!loading && !error && filteredResults.length > 0 && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5"
              >
                {filteredResults.map((item, index) => (
                  <RecommendationCard
                    key={item.id || index}
                    item={item}
                    index={index}
                    onMarkSeen={handleMarkSeen}
                    onFindSimilar={handleFindSimilar}
                    isSeen={seenIds.includes(item.id)}
                  />
                ))}
              </motion.div>
            )}
          </main>

          {/* Footer */}
          <footer className="mt-24 text-center">
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 1 }}
              className="inline-block"
            >
              <p className="text-muted/30 font-mono text-[10px] tracking-[0.3em]">
                WHAT 2 WATCH // POWERED BY AI EMBEDDINGS + LLM RERANKING
              </p>
              <div className="flex items-center justify-center gap-2 mt-2">
                <div className="w-1.5 h-1.5 rounded-full bg-neon-green/60 animate-pulse" />
                <span className="text-muted/20 font-mono text-[10px] tracking-wider">SYSTEM ONLINE</span>
              </div>
            </motion.div>
          </footer>
        </div>
      </div>

      {/* Watch history drawer */}
      <WatchHistory
        isOpen={historyOpen}
        onClose={() => setHistoryOpen(false)}
        history={watchHistory}
        onFindSimilar={handleFindSimilar}
        onRemove={handleRemoveSeen}
        isLoading={historyLoading}
      />

      {/* Toast notifications */}
      <AnimatePresence>
        {toast && (
          <Toast
            message={toast.message}
            type={toast.type}
            onDismiss={() => setToast(null)}
          />
        )}
      </AnimatePresence>

      {/* Keyboard shortcuts info */}
      <div className="fixed bottom-4 left-4 hidden md:block">
        <div className="flex items-center gap-3 text-muted/20 font-mono text-[10px]">
          <span>
            <kbd className="px-1 py-0.5 bg-void-lighter/50 border border-white/[0.06] rounded mr-1">/</kbd>
            Search
          </span>
          <span>
            <kbd className="px-1 py-0.5 bg-void-lighter/50 border border-white/[0.06] rounded mr-1">ESC</kbd>
            Clear
          </span>
        </div>
      </div>
    </div>
  )
}
