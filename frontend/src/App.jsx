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

  const {
    loading,
    error,
    search,
    markAsSeen,
    getWatchHistory,
    getHiddenGems,
    getSimilar,
    clearError,
  } = useApi()

  // Filter results by media type
  const filteredResults = results.filter((item) => {
    if (activeFilter === 'all') return true
    return item.type === activeFilter
  })

  // Handle search
  const handleSearch = useCallback(async (query) => {
    setCurrentQuery(query)
    setHasSearched(true)
    clearError()

    try {
      const data = await search(query)
      setResults(data.recommendations || data || [])

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

  // Handle mark as seen
  const handleMarkSeen = useCallback(async (mediaId) => {
    try {
      await markAsSeen(mediaId)
      setSeenIds((prev) => [...new Set([...prev, mediaId])])
    } catch (err) {
      console.error('Failed to mark as seen:', err)
      throw err
    }
  }, [markAsSeen, setSeenIds])

  // Handle find similar
  const handleFindSimilar = useCallback(async (mediaId) => {
    setHasSearched(true)
    clearError()

    try {
      const data = await getSimilar(mediaId)
      setResults(data.recommendations || data || [])
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
      setResults(data.recommendations || data || [])
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
        setWatchHistory(data.history || data || [])
      } catch (err) {
        console.error('Failed to load watch history:', err)
        setWatchHistory([])
      } finally {
        setHistoryLoading(false)
      }
    }
    setHistoryOpen(!historyOpen)
  }, [historyOpen, getWatchHistory])

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
          <header className="mb-12">
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
                    <span className="text-muted font-mono text-sm">QUERY:</span>
                    <span className="text-neon-cyan font-mono">"{currentQuery}"</span>
                  </div>
                  <span className="text-muted font-mono text-sm">
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

            {/* Empty state - initial */}
            {!loading && !error && !hasSearched && (
              <EmptyState type="initial" />
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
                className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
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
          <footer className="mt-20 text-center">
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 1 }}
              className="inline-block"
            >
              <p className="text-muted/50 font-mono text-xs tracking-widest">
                WHAT 2 WATCH // POWERED BY AI EMBEDDINGS + LLM RERANKING
              </p>
              <div className="flex items-center justify-center gap-2 mt-2">
                <div className="w-2 h-2 bg-neon-green animate-pulse" />
                <span className="text-muted/30 font-mono text-xs">SYSTEM ONLINE</span>
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
        isLoading={historyLoading}
      />

      {/* Keyboard shortcuts info */}
      <div className="fixed bottom-4 left-4 hidden md:block">
        <div className="flex items-center gap-4 text-muted/30 font-mono text-xs">
          <span>
            <kbd className="px-1.5 py-0.5 bg-void-lighter border border-muted/20 mr-1">/</kbd>
            Search
          </span>
          <span>
            <kbd className="px-1.5 py-0.5 bg-void-lighter border border-muted/20 mr-1">ESC</kbd>
            Clear
          </span>
        </div>
      </div>
    </div>
  )
}
