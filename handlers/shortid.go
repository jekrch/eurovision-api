package handlers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/teris-io/shortid"
)

// Global instance of the shortid generator
var globalShortID *shortid.Shortid

/**
 * initializes the global short id generator
 */
func InitShortID() error {
	seedStr := os.Getenv("SHORT_ID_SEED")
	if seedStr == "" {
		seedStr = "2342" // fallback default
	}

	seed, err := strconv.ParseUint(seedStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid SHORT_ID_SEED: %v", err)
	}

	sid, err := shortid.New(1, shortid.DefaultABC, seed)
	if err != nil {
		return fmt.Errorf("failed to initialize shortid: %v", err)
	}

	globalShortID = sid
	return nil
}

/**
 * generates a new short id
 */
func GenerateShortID() string {
	sid, err := globalShortID.Generate()
	if err != nil {
		panic(fmt.Errorf("failed to generate short id: %v", err))
	}
	return sid
}
