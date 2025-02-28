package utils

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	lock sync.Mutex
)

// Initializes the Snowflake node
func InitSnowflake(machineID int64) error {
	var err error
	node, err = snowflake.NewNode(machineID)
	return err
}

// Generates a Snowflake ID for a click event
func GenerateSnowflakeID(shortURL string) string {
	lock.Lock()
	defer lock.Unlock()

	// Generate a unique Snowflake ID
	snowflakeID := node.Generate().Int64()

	// Format the key as click:shortURL:snowflakeID
	return fmt.Sprintf("click:%s:%d", shortURL, snowflakeID)
}

// Extracts the Snowflake ID from a click key
func DecodeShortURLFromSnowflakeID(clickKey string) (string, error) {
	// Example format: "click:shortURL:snowflakeID"
	parts := strings.Split(clickKey, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid click key format")
	}

	return parts[1], nil // Extract the shortURL
}
