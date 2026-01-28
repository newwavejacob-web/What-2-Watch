// Random vibe prompts for the surprise me feature
export const vibePrompts = [
  "melancholic cyberpunk with jazz undertones",
  "psychological mindfuck with gorgeous cinematography",
  "cozy slice of life that heals the soul",
  "existential dread but make it beautiful",
  "neon-drenched noir with a synth soundtrack",
  "slow burn romance in a dystopian setting",
  "philosophical anime that questions reality",
  "heartwarming found family in space",
  "gritty crime drama with moral ambiguity",
  "surreal dreamscape with emotional depth",
  "action-packed revenge thriller",
  "haunting gothic atmosphere with romance",
  "coming of age in a broken world",
  "time travel paradox that actually makes sense",
  "isolated horror with mounting dread",
  "bittersweet nostalgia trip",
  "cerebral sci-fi that respects your intelligence",
  "dark comedy about the human condition",
  "ethereal fantasy with Studio Ghibli vibes",
  "tense political thriller with conspiracies",
  "wholesome adventure with stakes",
  "artistic film that's actually entertaining",
  "mind-bending mystery that rewards attention",
  "emotional devastation disguised as entertainment",
  "retro aesthetic with modern themes",
  "quiet character study with deep themes",
  "chaotic energy but emotionally resonant",
  "cyberpunk detective story",
  "poetic meditation on loneliness",
  "high concept executed perfectly",
]

export const getRandomVibe = () => {
  return vibePrompts[Math.floor(Math.random() * vibePrompts.length)]
}

export const placeholderVibes = [
  "melancholic cyberpunk with jazz...",
  "psychological mindfuck with gorgeous cinematography...",
  "cozy slice of life that heals the soul...",
  "existential dread but make it beautiful...",
  "neon-drenched noir with synth...",
]

export const getRandomPlaceholder = () => {
  return placeholderVibes[Math.floor(Math.random() * placeholderVibes.length)]
}

// Media type colors
export const mediaTypeColors = {
  anime: '#FF006E',
  movie: '#00FFFF',
  tv: '#39FF14',
}

export const mediaTypeLabels = {
  anime: 'ANIME',
  movie: 'FILM',
  tv: 'SERIES',
}
