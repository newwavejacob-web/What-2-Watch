import { motion } from 'framer-motion'
import { Search, Sparkles, Zap } from 'lucide-react'

export default function EmptyState({ type = 'initial' }) {
  if (type === 'no-results') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center py-20"
      >
        <motion.div
          animate={{ rotate: [0, 10, -10, 0] }}
          transition={{ duration: 3, repeat: Infinity }}
          className="inline-block text-6xl mb-6"
        >
          <span className="text-glow-pink">:/</span>
        </motion.div>
        <h3 className="text-2xl font-bold font-mono text-warm-white mb-2">
          NO SIGNAL DETECTED
        </h3>
        <p className="text-muted max-w-md mx-auto">
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
      className="text-center py-16"
    >
      {/* Animated prompt suggestions */}
      <div className="mb-12">
        <p className="text-muted font-mono text-sm mb-6 tracking-wider">
          // TRY SOMETHING LIKE
        </p>
        <div className="flex flex-wrap justify-center gap-3 max-w-2xl mx-auto">
          {[
            'existential dread but make it beautiful',
            'cozy apocalypse vibes',
            'neon noir with heart',
            'psychological slow burn',
            'found family in space',
          ].map((suggestion, i) => (
            <motion.div
              key={suggestion}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.8 + i * 0.1 }}
              className="px-4 py-2 bg-void-lighter border border-muted/30 text-muted-light text-sm font-mono cursor-default hover:border-neon-cyan hover:text-neon-cyan transition-colors duration-300"
            >
              "{suggestion}"
            </motion.div>
          ))}
        </div>
      </div>

      {/* Feature highlights */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-3xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.2 }}
          className="p-6 border border-muted/20 hover:border-neon-cyan/50 transition-colors"
        >
          <Search className="w-8 h-8 text-neon-cyan mx-auto mb-4" />
          <h4 className="font-mono font-bold text-warm-white mb-2">VIBE SEARCH</h4>
          <p className="text-muted text-sm">
            Describe the feeling you want. We'll decode your taste.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.3 }}
          className="p-6 border border-muted/20 hover:border-neon-green/50 transition-colors"
        >
          <Sparkles className="w-8 h-8 text-neon-green mx-auto mb-4" />
          <h4 className="font-mono font-bold text-warm-white mb-2">HIDDEN GEMS</h4>
          <p className="text-muted text-sm">
            Discover high-quality obscure titles you've never heard of.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.4 }}
          className="p-6 border border-muted/20 hover:border-neon-pink/50 transition-colors"
        >
          <Zap className="w-8 h-8 text-neon-pink mx-auto mb-4" />
          <h4 className="font-mono font-bold text-warm-white mb-2">AI POWERED</h4>
          <p className="text-muted text-sm">
            Embeddings + LLM reranking for uncanny accuracy.
          </p>
        </motion.div>
      </div>
    </motion.div>
  )
}
