package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Command docker docker run -i -t -p 8080:8080 -v $(pwd):/opt/metarank metarank/metarank:latest standalone --config /opt/metarank/config.yml --data /opt/metarank/user_interaction_events_100.jsonl.gz

// Structures pour les données brutes issues d'Elasticsearch
type RawData struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     Hits `json:"hits"`
}

type Hits struct {
	Total    Total       `json:"total"`
	MaxScore float64     `json:"max_score"`
	Hits     []HitDetail `json:"hits"`
}

type Total struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type HitDetail struct {
	Index  string  `json:"_index"`
	Type   string  `json:"_type"`
	ID     string  `json:"_id"`
	Score  float64 `json:"_score"`
	Source Source  `json:"_source"`
}

type Source struct {
	Price       Price    `json:"price"`
	Property    Property `json:"property"`
	Location    Location `json:"location"`
	ID          string   `json:"id"`
	Transaction string   `json:"transaction"`
}

type Price struct {
	Value int `json:"value"`
}

type Property struct {
	EstateType string `json:"estateType"`
}

type Location struct {
	City       string `json:"city"`
	PostalCode string `json:"postalCode"`
}

// Structure pour les événements Metarank
type Event struct {
	Event     string  `json:"event"`
	ID        string  `json:"id"`
	Timestamp string  `json:"timestamp"`
	Item      string  `json:"item"`
	Fields    []Field `json:"fields"`
}

type Field struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func main() {
	// Lire le fichier JSON brut
	inputFile := "classifieds.json"
	outputFile := "formatted_classifieds.jsonl"

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Failed to read input file: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal des données brutes dans une structure Go
	var rawData RawData
	if err := json.Unmarshal(data, &rawData); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Créer un fichier de sortie
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	now := time.Now()
	ts := strconv.FormatInt(now.Unix(), 10)
	// Transformer les données et les écrire au format JSONL
	for _, hit := range rawData.Hits.Hits {
		event := Event{
			Event:     "item",
			ID:        hit.ID,
			Timestamp: ts,
			Item:      hit.ID,
			Fields: []Field{
				{"price", hit.Source.Price.Value},
				{"estateType", hit.Source.Property.EstateType},
				{"city", hit.Source.Location.City},
				{"postalCode", hit.Source.Location.PostalCode},
				{"transaction", hit.Source.Transaction},
			},
		}

		// Convertir l'événement en JSON
		eventData, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Failed to marshal event: %v\n", err)
			continue
		}

		// Écrire l'événement au format JSONL
		output.Write(eventData)
		output.Write([]byte("\n"))
	}

	fmt.Printf("Conversion complete. Output written to %s\n", outputFile)
}
