package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moby/moby/client"
)

type AppRouter struct {
	Pool *pgxpool.Pool
}

func NewAppRouter(pool *pgxpool.Pool) *AppRouter {
	return &AppRouter{
		Pool: pool,
	}
}

type ImageDetails struct {
	ImageID   string
	ImageName string
}

func GetContainerImageDetails(imageLoadResponse client.LoadResponse) (ImageDetails, error) {
	var imageID string
	scanner := bufio.NewScanner(imageLoadResponse.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Loaded image ID:") {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(line), &result); err == nil {
				if stream, ok := result["stream"].(string); ok {
					if strings.HasPrefix(stream, "Loaded image ID: ") {
						imageID = strings.TrimPrefix(stream, "Loaded image ID: ")
						imageID = strings.TrimSpace(imageID)
						break
					}
				}
			}
		} else if strings.Contains(line, "Loaded image:") {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(line), &result); err == nil {
				if stream, ok := result["stream"].(string); ok {
					if strings.HasPrefix(stream, "Loaded image: ") {
						imageID = strings.TrimPrefix(stream, "Loaded image: ")
						imageID = strings.TrimSpace(imageID)
						break
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ImageDetails{}, fmt.Errorf("Error reading ImageLoad response: %v", err)
	}

	if imageID == "" {
		return ImageDetails{}, fmt.Errorf("Could not extract image ID from ImageLoad response")
	}

	imageName := strings.Split(imageID, ":")[0]

	return ImageDetails{
		ImageID:   imageID,
		ImageName: imageName,
	}, nil
}
