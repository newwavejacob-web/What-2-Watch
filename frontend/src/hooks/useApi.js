import { useState, useCallback } from 'react'
import { getUserId } from './useLocalStorage'

const API_BASE = '/api'

export function useApi() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const fetchWithError = useCallback(async (url, options = {}) => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      setLoading(false)
      return data
    } catch (err) {
      setError(err.message)
      setLoading(false)
      throw err
    }
  }, [])

  const search = useCallback(async (query, limit = 10) => {
    return fetchWithError(`${API_BASE}/recommend`, {
      method: 'POST',
      body: JSON.stringify({
        user_id: getUserId(),
        query,
        limit,
      }),
    })
  }, [fetchWithError])

  const markAsSeen = useCallback(async (mediaId) => {
    return fetchWithError(`${API_BASE}/seen`, {
      method: 'POST',
      body: JSON.stringify({
        user_id: getUserId(),
        media_id: mediaId,
      }),
    })
  }, [fetchWithError])

  const getWatchHistory = useCallback(async () => {
    return fetchWithError(`${API_BASE}/seen?user_id=${encodeURIComponent(getUserId())}`)
  }, [fetchWithError])

  const getHiddenGems = useCallback(async () => {
    return fetchWithError(`${API_BASE}/hidden-gems?user_id=${encodeURIComponent(getUserId())}`)
  }, [fetchWithError])

  const getSimilar = useCallback(async (mediaId) => {
    return fetchWithError(`${API_BASE}/similar/${encodeURIComponent(mediaId)}?user_id=${encodeURIComponent(getUserId())}`)
  }, [fetchWithError])

  return {
    loading,
    error,
    search,
    markAsSeen,
    getWatchHistory,
    getHiddenGems,
    getSimilar,
    clearError: () => setError(null),
  }
}
