package main

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"os"
	"strconv"
	"time"
)

const esEndpoint = "https://multi-es-elasticsearch-immo.staging.fcms.io:9201" // Elasticsearch endpoint

/* const esEndpoint = "http://localhost:9201" // Elasticsearch endpoint

const username = ""
const password = "" */

const index = "fi-classified-latest" // Elasticsearch index
var limit = 100000                   // Nombre total de documents souhaités

//go:embed metarank.json
var request []byte

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
	outputFile := "formatted_classifieds.jsonl"

	// Appel à readData pour récupérer les données depuis Elasticsearch
	targetData := readData(index)
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	now := time.Now()
	ts := strconv.FormatInt(now.Unix(), 10)
	for id, hit := range targetData {
		// Vérification et conversion sécurisée du champ _source en map
		source, ok := hit["_source"].(map[string]interface{})
		if !ok {
			fmt.Printf("Warning: _source field missing or not a map for document %s\n", id)
			continue
		}

		event := Event{
			Event:     "item",
			ID:        id,
			Timestamp: ts,
			Item:      id,
			Fields:    []Field{},
		}

		// Gestion du champ "price"
		if price, ok := source["price"].(map[string]interface{}); ok {
			if value, ok := price["value"]; ok {
				event.Fields = append(event.Fields, Field{"price", value})
			} else {
				fmt.Printf("Warning: Missing 'value' in 'price' for document %s\n", id)
			}
		} else {
			fmt.Printf("Warning: Missing 'price' for document %s\n", id)
		}

		// Gestion du champ "estateType"
		if property, ok := source["property"].(map[string]interface{}); ok {
			if estateType, ok := property["estateType"]; ok {
				event.Fields = append(event.Fields, Field{"estateType", estateType})
			} else {
				fmt.Printf("Warning: Missing 'estateType' in 'property' for document %s\n", id)
			}
		} else {
			fmt.Printf("Warning: Missing 'property' for document %s\n", id)
		}

		// Gestion du champ "location"
		if location, ok := source["location"].(map[string]interface{}); ok {
			if city, ok := location["city"]; ok {
				event.Fields = append(event.Fields, Field{"city", city})
			} else {
				fmt.Printf("Warning: Missing 'city' in 'location' for document %s\n", id)
			}
			if postalCode, ok := location["postalCode"]; ok {
				event.Fields = append(event.Fields, Field{"postalCode", postalCode})
			} else {
				fmt.Printf("Warning: Missing 'postalCode' in 'location' for document %s\n", id)
			}
		} else {
			fmt.Printf("Warning: Missing 'location' for document %s\n", id)
		}

		// Gestion du champ "transaction"
		if transaction, ok := source["transaction"]; ok {
			event.Fields = append(event.Fields, Field{"transaction", transaction})
		} else {
			fmt.Printf("Warning: Missing 'transaction' for document %s\n", id)
		}

		eventData, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Failed to marshal event for document %s: %v\n", id, err)
			continue
		}

		output.Write(eventData)
		output.Write([]byte("\n"))
	}

	fmt.Printf("Conversion complete. Output written to %s\n", outputFile)
}

func readData(index string) map[string]map[string]any {
	targetData := make(map[string]map[string]any)
	// Initial search request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/_search?scroll=1m", esEndpoint, index), bytes.NewBuffer(request))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	username := os.Getenv("FI_ES_USERNAME")
	password := os.Getenv("FI_ES_PASSWORD")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error executing request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Parse the response to extract the scroll ID and hits
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	scrollID, ok := response["_scroll_id"].(string)
	if !ok {
		log.Fatalf("Failed to extract scroll ID %s", response)
		log.Fatal("Failed to extract scroll ID")
	}

	hits, ok := response["hits"].(map[string]interface{})["hits"].([]any)
	if !ok || len(hits) == 0 {
		log.Fatal("No documents found.")
	}

	for _, hit := range hits {
		mapHit, ok := hit.(map[string]any)
		if !ok {
			continue
		}
		id, ok := mapHit["_id"].(string)
		if !ok {
			continue
		}
		delete(mapHit, "_index")
		targetData[id] = mapHit
	}

	for scrollID != "" {
		// Prepare the scroll request
		scrollRequestBody := []byte(fmt.Sprintf(`{"scroll":"1m","scroll_id":"%s"}`, scrollID))
		scrollReq, err := http.NewRequest("GET", fmt.Sprintf("%s/_search/scroll", esEndpoint), bytes.NewBuffer(scrollRequestBody))
		if err != nil {
			log.Fatalf("Error creating scroll request: %v", err)
		}
		scrollReq.Header.Add("Content-Type", "application/json")
		scrollReq.SetBasicAuth(username, password)

		scrollResp, err := client.Do(scrollReq)
		if err != nil {
			log.Fatalf("Error executing scroll request: %v", err)
		}
		defer scrollResp.Body.Close()

		scrollBody, err := ioutil.ReadAll(scrollResp.Body)
		if err != nil {
			log.Fatalf("Error reading scroll response body: %v", err)
		}

		if err := json.Unmarshal(scrollBody, &response); err != nil {
			log.Fatalf("Error parsing JSON scroll response: %v", err)
		}

		scrollID, ok = response["_scroll_id"].(string)
		if !ok || scrollID == "" {
			break
		}

		hits, ok = response["hits"].(map[string]interface{})["hits"].([]any)
		if !ok || len(hits) == 0 {
			break
		}
		for _, hit := range hits {
			mapHit, ok := hit.(map[string]any)
			if !ok {
				continue
			}
			id, ok := mapHit["_id"].(string)
			if !ok {
				continue
			}
			delete(mapHit, "_index")
			targetData[id] = mapHit
		}

		if len(targetData) >= limit {
			return targetData
		}
	}
	return targetData
}
