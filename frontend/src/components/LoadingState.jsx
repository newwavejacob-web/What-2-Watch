import { motion } from 'framer-motion'
import { useEffect, useState, useMemo } from 'react'

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

const matrixChars = 'アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン'

function MatrixRain() {
  const columns = 20
  const chars = useMemo(
    () =>
      Array.from({ length: columns }, () =>
        Array.from({ length: 12 }, () =>
          matrixChars[Math.floor(Math.random() * matrixChars.length)]
        ).join('')
      ),
    []
  )

  return (
    <div className="absolute inset-0 overflow-hidden opacity-15 pointer-events-none">
      {chars.map((text, i) => (
        <motion.div
          key={i}
          initial={{ y: -100 }}
          animate={{ y: '100vh' }}
          transition={{
            duration: 3 + Math.random() * 4,
            repeat: Infinity,
            delay: Math.random() * 3,
            ease: 'linear',
          }}
          className="absolute text-neon-green/40 font-mono text-[10px]"
          style={{
            left: `${(i / columns) * 100}%`,
            writingMode: 'vertical-rl',
            willChange: 'transform',
          }}
        >
          {text}
        </motion.div>
      ))}
    </div>
  )
}

function GlitchText({ text }) {
  return (
    <div className="relative inline-block">
      <span className="text-warm-white/90 font-mono text-sm tracking-widest">{text}</span>
      <motion.span
        animate={{
          x: [-1.5, 1.5, -1.5],
          opacity: [0.6, 0.15, 0.6],
        }}
        transition={{
          duration: 0.25,
          repeat: Infinity,
          ease: 'linear',
        }}
        className="absolute inset-0 text-neon-cyan font-mono text-sm tracking-widest"
        style={{ clipPath: 'inset(10% 0 60% 0)' }}
      >
        {text}
      </motion.span>
      <motion.span
        animate={{
          x: [1.5, -1.5, 1.5],
          opacity: [0.6, 0.15, 0.6],
        }}
        transition={{
          duration: 0.25,
          repeat: Infinity,
          ease: 'linear',
          delay: 0.08,
        }}
        className="absolute inset-0 text-neon-pink font-mono text-sm tracking-widest"
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
    }, 900)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="relative flex flex-col items-center justify-center py-24">
      {/* Matrix rain background */}
      <MatrixRain />

      {/* Central loading indicator */}
      <div className="relative z-10 w-28 h-28">
        {/* Outer spinning ring */}
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 3, repeat: Infinity, ease: 'linear' }}
          className="absolute inset-0 rounded-full"
          style={{
            border: '2px solid transparent',
            borderTopColor: '#00FFFF',
            borderRightColor: 'rgba(255, 0, 110, 0.5)',
            willChange: 'transform',
          }}
        />

        {/* Inner counter-spinning ring */}
        <motion.div
          animate={{ rotate: -360 }}
          transition={{ duration: 5, repeat: Infinity, ease: 'linear' }}
          className="absolute inset-3 rounded-full"
          style={{
            border: '1px solid transparent',
            borderBottomColor: 'rgba(57, 255, 20, 0.4)',
            borderLeftColor: 'rgba(0, 255, 255, 0.2)',
            willChange: 'transform',
          }}
        />

        {/* Inner pulsing core */}
        <motion.div
          animate={{
            scale: [1, 1.15, 1],
            opacity: [0.3, 0.7, 0.3],
          }}
          transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
          className="absolute inset-6 rounded-full bg-void border border-neon-cyan/20"
          style={{
            boxShadow: '0 0 20px rgba(0, 255, 255, 0.2), inset 0 0 10px rgba(0, 255, 255, 0.05)',
          }}
        />

        {/* Center dot */}
        <div className="absolute inset-0 flex items-center justify-center">
          <motion.div
            animate={{
              scale: [1, 1.3, 1],
              opacity: [0.5, 1, 0.5],
            }}
            transition={{ duration: 1.5, repeat: Infinity, ease: 'easeInOut' }}
            className="w-2 h-2 rounded-full bg-neon-cyan"
            style={{ boxShadow: '0 0 10px #00FFFF' }}
          />
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
        <div className="flex items-center justify-center gap-1.5 mt-5">
          {[0, 1, 2].map((i) => (
            <motion.div
              key={i}
              animate={{
                scale: [1, 1.4, 1],
                opacity: [0.2, 0.8, 0.2],
              }}
              transition={{
                duration: 0.8,
                repeat: Infinity,
                delay: i * 0.15,
                ease: 'easeInOut',
              }}
              className="w-1.5 h-1.5 rounded-full bg-neon-cyan"
            />
          ))}
        </div>
      </motion.div>

      {/* Horizontal scan line */}
      <motion.div
        animate={{ top: ['0%', '100%'] }}
        transition={{ duration: 2.5, repeat: Infinity, ease: 'linear' }}
        className="absolute left-0 right-0 h-px"
        style={{
          background: 'linear-gradient(90deg, transparent, rgba(0,255,255,0.3), transparent)',
          boxShadow: '0 0 6px rgba(0,255,255,0.2)',
          willChange: 'top',
        }}
      />
    </div>
  )
}
