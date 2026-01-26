package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"w2w/internal/database"
	"w2w/internal/embeddings"
	"w2w/internal/llm"
	"w2w/internal/models"
)

// SeedEntry contains info for seeding a media entry
type SeedEntry struct {
	Title     string
	MediaType string
	Year      int
	Synopsis  string
	// Pre-written vibe profiles for when no API key is available
	FallbackVibe string
}

// Sample hidden gems and quality titles
var seedData = []SeedEntry{
	{
		Title:     "Pantheon",
		MediaType: "anime",
		Year:      2022,
		Synopsis:  "A bullied teen receives mysterious help from an entity claiming to be her dead father's uploaded consciousness.",
		FallbackVibe: "Cerebral and haunting digital noir. The pacing is deliberate and tense, with moments of existential dread punctuated by raw emotional vulnerability. Aesthetic of cold server rooms and warm human memories colliding.",
	},
	{
		Title:     "Id:Invaded",
		MediaType: "anime",
		Year:      2020,
		Synopsis:  "A detective enters the subconscious minds of serial killers using a device that creates dream worlds from their cognition particles.",
		FallbackVibe: "Mind-bending neo-noir with a kaleidoscopic dreamscape aesthetic. Frenetic psychological energy wrapped in contemplative detective procedural pacing. Feels like falling through fractured mirrors of guilt and identity.",
	},
	{
		Title:     "Odd Taxi",
		MediaType: "anime",
		Year:      2021,
		Synopsis:  "A 41-year-old walrus taxi driver becomes entangled in a missing girl case through his passengers' interconnected stories.",
		FallbackVibe: "Rain-soaked urban melancholy with sharp wit. Deceptively cozy surface hiding razor-sharp tension. Lo-fi jazz atmosphere, the quiet dread of ordinary people carrying extraordinary secrets.",
	},
	{
		Title:     "Akira",
		MediaType: "anime",
		Year:      1988,
		Synopsis:  "A biker gang member gains telekinetic powers after a military experiment goes wrong in post-apocalyptic Neo-Tokyo.",
		FallbackVibe: "Hyperkinetic cyberpunk explosion of neon and chrome. Pulsing with restless adolescent energy and apocalyptic dread. Every frame vibrates with mechanical detail and psychic overflow.",
	},
	{
		Title:     "Perfect Blue",
		MediaType: "anime",
		Year:      1997,
		Synopsis:  "A pop idol turned actress's sense of reality becomes distorted as she's stalked and her virtual image takes on a life of its own.",
		FallbackVibe: "Suffocating psychological thriller wrapped in glossy pop aesthetics. Reality fractures like a broken mirror. Paranoid, claustrophobic, deeply unsettling descent into identity dissolution.",
	},
	{
		Title:     "Paprika",
		MediaType: "anime",
		Year:      2006,
		Synopsis:  "A research psychologist uses a device that allows therapists to enter patients' dreams, but the device is stolen.",
		FallbackVibe: "Kaleidoscopic dream logic explosion. Joyful and terrifying in equal measure. Saturated colors bleeding into impossible geometries. The aesthetic of pure imagination unshackled.",
	},
	{
		Title:     "Serial Experiments Lain",
		MediaType: "anime",
		Year:      1998,
		Synopsis:  "A shy girl becomes obsessed with the Wired, a global communications network, and discovers unsettling truths about reality and identity.",
		FallbackVibe: "Eerie digital liminal spaces and humming power lines. Slow, hypnotic dread. The loneliness of dial-up connections and empty suburban streets. God is in the machine.",
	},
	{
		Title:     "Blade Runner 2049",
		MediaType: "movie",
		Year:      2017,
		Synopsis:  "A blade runner discovers a secret that could destabilize society and his quest leads him to find a former blade runner.",
		FallbackVibe: "Glacial neon-noir meditation. Vast empty spaces that swallow the soul. Every shot a melancholic painting of rain and holographic ghosts. Aching loneliness rendered in brutalist architecture.",
	},
	{
		Title:     "Arrival",
		MediaType: "movie",
		Year:      2016,
		Synopsis:  "A linguist is recruited to help communicate with alien visitors before global tensions escalate.",
		FallbackVibe: "Contemplative sci-fi wrapped in fog and grief. Time bends like light through water. Quiet, mournful, intellectually stimulating. The weight of knowing the future and choosing it anyway.",
	},
	{
		Title:     "The Lighthouse",
		MediaType: "movie",
		Year:      2019,
		Synopsis:  "Two lighthouse keepers try to maintain their sanity while living on a remote island.",
		FallbackVibe: "Claustrophobic maritime madness in crushing black and white. Salt-crusted fever dream. Fog horns and crashing waves as the soundtrack to psychological disintegration. Mythic dread.",
	},
	{
		Title:     "Severance",
		MediaType: "tv",
		Year:      2022,
		Synopsis:  "Employees of a corporation undergo a procedure that divides their work and personal memories.",
		FallbackVibe: "Corporate uncanny valley nightmare in mint green and fluorescent dread. Retrofuturistic unease. The horror of bureaucratic existence rendered in painfully sterile corridors. Kafkaesque with a beating heart.",
	},
	{
		Title:     "Mr. Robot",
		MediaType: "tv",
		Year:      2015,
		Synopsis:  "A cybersecurity engineer with social anxiety and clinical depression becomes involved with an underground hacker group.",
		FallbackVibe: "Paranoid digital-age thriller shot through fractured perspectives. Off-center framing that mirrors psychological instability. Loneliness in the age of connection. The revolution will be debugged.",
	},
	{
		Title:     "Cowboy Bebop",
		MediaType: "anime",
		Year:      1998,
		Synopsis:  "A ragtag crew of bounty hunters aboard a spaceship pursue criminals across the solar system.",
		FallbackVibe: "Cool-jazz-in-space melancholy. Effortlessly stylish ennui. Everyone running from their past through neon-lit corridors and dusty frontier towns. You're gonna carry that weight.",
	},
	{
		Title:     "Ghost in the Shell",
		MediaType: "anime",
		Year:      1995,
		Synopsis:  "A cyborg policewoman hunts a mysterious hacker while questioning the nature of consciousness.",
		FallbackVibe: "Philosophical cyberpunk meditation dripping with rain and neon. Where does the machine end and the soul begin? Lush urban decay and digital transcendence. Haunted by questions of identity.",
	},
	{
		Title:     "Annihilation",
		MediaType: "movie",
		Year:      2018,
		Synopsis:  "A biologist joins an expedition into an environmental disaster zone where the laws of nature don't apply.",
		FallbackVibe: "Iridescent body-horror beauty. Nature reclaiming and remixing human form. Dream-logic dread in prismatic colors. Self-destruction as transformation. Hypnotically unsettling.",
	},
}

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./vibe.db"
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	fmt.Printf("Seeding database: %s\n", dbPath)
	fmt.Printf("OpenAI API Key: %v\n\n", apiKey != "")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize embedding provider
	var embedProvider embeddings.Provider
	var llmClient *llm.Client

	if apiKey != "" {
		embedProvider = embeddings.NewOpenAIProvider(apiKey)
		llmClient = llm.NewClient(apiKey)
		fmt.Println("Using OpenAI for vibe generation and embeddings")
	} else {
		embedProvider = &placeholderEmbedder{}
		fmt.Println("Using placeholder embeddings (no API key)")
		fmt.Println("Vibes will use pre-written fallback profiles")
	}

	// Create a default user
	defaultUser := &models.User{
		ID:        "default",
		Username:  "default",
		CreatedAt: time.Now(),
	}
	db.CreateUser(defaultUser)
	fmt.Println("Created default user")

	// Seed media entries
	for _, entry := range seedData {
		fmt.Printf("Processing: %s (%d)...", entry.Title, entry.Year)

		// Check if already exists
		existing, _ := db.GetMediaByTitle(entry.Title)
		if existing != nil {
			fmt.Println(" already exists, skipping")
			continue
		}

		// Generate or use fallback vibe profile
		var vibeProfile string
		if llmClient != nil {
			vibe, err := llmClient.GenerateVibeProfile(entry.Title, entry.MediaType, entry.Year, entry.Synopsis)
			if err != nil {
				fmt.Printf(" LLM error: %v, using fallback\n", err)
				vibeProfile = entry.FallbackVibe
			} else {
				vibeProfile = vibe
				fmt.Print(" generated vibe...")
			}
		} else {
			vibeProfile = entry.FallbackVibe
		}

		// Create media entry
		media := &models.Media{
			ID:          generateID(entry.Title, entry.MediaType),
			Title:       entry.Title,
			MediaType:   entry.MediaType,
			Year:        entry.Year,
			PlotSummary: entry.Synopsis,
			VibeProfile: vibeProfile,
			QualityScore: 0.8, // Start with decent quality score for known good titles
		}

		if err := db.CreateMedia(media); err != nil {
			fmt.Printf(" DB error: %v\n", err)
			continue
		}

		// Generate and store embedding
		embedding, err := embedProvider.Embed(vibeProfile)
		if err != nil {
			fmt.Printf(" embed error: %v\n", err)
			continue
		}

		if err := db.StoreEmbedding(media.ID, embedding, embedProvider.ModelName()); err != nil {
			fmt.Printf(" store error: %v\n", err)
			continue
		}

		fmt.Println(" done")
	}

	fmt.Println("\nSeed complete!")
	fmt.Println("\nYou can now run the server with: ./vibe-server")
	fmt.Println("Try a query: curl 'http://localhost:8080/vibe?q=mind-bending+psychological+thriller'")
}

func generateID(title, mediaType string) string {
	id := fmt.Sprintf("%s-%s", mediaType, title)
	result := ""
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r == ' ' || r == '-' {
			result += "-"
		}
	}
	return result
}

// placeholderEmbedder for when no API key is available
type placeholderEmbedder struct{}

func (p *placeholderEmbedder) Embed(text string) ([]float32, error) {
	embedding := make([]float32, 1536)
	for i, r := range text {
		idx := i % 1536
		embedding[idx] += float32(r) / 1000.0
	}
	var sum float32
	for _, v := range embedding {
		sum += v * v
	}
	if sum > 0 {
		norm := float32(1.0) / float32(sum)
		for i := range embedding {
			embedding[i] *= norm
		}
	}
	return embedding, nil
}

func (p *placeholderEmbedder) ModelName() string {
	return "placeholder-dev"
}
