import { motion } from 'framer-motion'
import { useEffect, useState } from 'react'

const glitchText = [
  'SCANNING NEURAL NETWORK...',
  'ANALYZING VIBE PATTERNS...',
  'QUERYING EMBEDDING SPACE...',
  'CROSS-REFERENCING AESTHETICS...',
  'CONSULTING THE ALGORITHM...',
  'DECODING YOUR TASTE...',
  'SEARCHING THE VOID...',
  'PROCESSING SIGNAL...',
  'COMPUTING SIMILARITY MATRIX...',
  'LOADING RECOMMENDATIONS...',
]

const matrixChars = 'アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン0123456789'

function MatrixRain() {
  const columns = 30

  return (
    <div className="absolute inset-0 overflow-hidden opacity-20 pointer-events-none">
      {[...Array(columns)].map((_, i) => (
        <motion.div
          key={i}
          initial={{ y: -100 }}
          animate={{ y: '100vh' }}
          transition={{
            duration: 2 + Math.random() * 3,
            repeat: Infinity,
            delay: Math.random() * 2,
            ease: 'linear',
          }}
          className="absolute text-neon-green font-mono text-sm"
          style={{
            left: `${(i / columns) * 100}%`,
            writingMode: 'vertical-rl',
          }}
        >
          {[...Array(20)].map((_, j) => (
            <span
              key={j}
              style={{ opacity: 1 - j * 0.05 }}
            >
              {matrixChars[Math.floor(Math.random() * matrixChars.length)]}
            </span>
          ))}
        </motion.div>
      ))}
    </div>
  )
}

function GlitchText({ text }) {
  return (
    <div className="relative">
      <span className="text-warm-white font-mono text-lg tracking-wider">{text}</span>
      <motion.span
        animate={{
          x: [-2, 2, -2],
          opacity: [0.8, 0.2, 0.8],
        }}
        transition={{
          duration: 0.2,
          repeat: Infinity,
        }}
        className="absolute inset-0 text-neon-cyan font-mono text-lg tracking-wider"
        style={{ clipPath: 'inset(10% 0 60% 0)' }}
      >
        {text}
      </motion.span>
      <motion.span
        animate={{
          x: [2, -2, 2],
          opacity: [0.8, 0.2, 0.8],
        }}
        transition={{
          duration: 0.2,
          repeat: Infinity,
          delay: 0.1,
        }}
        className="absolute inset-0 text-neon-pink font-mono text-lg tracking-wider"
        style={{ clipPath: 'inset(50% 0 20% 0)' }}
      >
        {text}
      </motion.span>
    </div>
  )
}

export default function LoadingState() {
  const [textIndex, setTextIndex] = useState(0)

  useEffect(() => {
    const interval = setInterval(() => {
      setTextIndex((prev) => (prev + 1) % glitchText.length)
    }, 800)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="relative flex flex-col items-center justify-center py-20">
      {/* Matrix rain background */}
      <MatrixRain />

      {/* Central loading indicator */}
      <div className="relative z-10">
        {/* Spinning border */}
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 3, repeat: Infinity, ease: 'linear' }}
          className="w-32 h-32 border-4 border-transparent"
          style={{
            borderTopColor: '#00FFFF',
            borderRightColor: '#FF006E',
          }}
        />

        {/* Inner pulsing circle */}
        <motion.div
          animate={{
            scale: [1, 1.2, 1],
            opacity: [0.5, 1, 0.5],
          }}
          transition={{ duration: 1.5, repeat: Infinity }}
          className="absolute inset-4 bg-void border border-neon-cyan"
          style={{ boxShadow: '0 0 30px rgba(0, 255, 255, 0.5)' }}
        />

        {/* Center icon */}
        <div className="absolute inset-0 flex items-center justify-center">
          <motion.div
            animate={{
              rotateY: 360,
              scale: [1, 1.1, 1],
            }}
            transition={{
              rotateY: { duration: 2, repeat: Infinity, ease: 'linear' },
              scale: { duration: 1, repeat: Infinity },
            }}
            className="text-3xl"
          >
            <span className="text-glow-cyan">
              {['>', '<', '/', '|'][Math.floor(Date.now() / 200) % 4]}
            </span>
          </motion.div>
        </div>
      </div>

      {/* Status text */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="mt-8 text-center z-10"
      >
        <GlitchText text={glitchText[textIndex]} />

        {/* Progress dots */}
        <div className="flex items-center justify-center gap-2 mt-4">
          {[0, 1, 2].map((i) => (
            <motion.div
              key={i}
              animate={{
                scale: [1, 1.5, 1],
                opacity: [0.3, 1, 0.3],
              }}
              transition={{
                duration: 0.8,
                repeat: Infinity,
                delay: i * 0.2,
              }}
              className="w-2 h-2 bg-neon-cyan"
            />
          ))}
        </div>
      </motion.div>

      {/* Horizontal scan line */}
      <motion.div
        animate={{ y: [-100, 200] }}
        transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
        className="absolute left-0 right-0 h-px bg-gradient-to-r from-transparent via-neon-cyan to-transparent"
        style={{ boxShadow: '0 0 10px #00FFFF' }}
      />
    </div>
  )
}
