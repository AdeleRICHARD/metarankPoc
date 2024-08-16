package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

type RecommendRequest struct {
	Count int      `json:"count"`
	User  string   `json:"user,omitempty"`
	Items []string `json:"items,omitempty"`
}

type RecommendResponse struct {
	Took  int            `json:"took"`
	Items []ResponseItem `json:"items"`
}

type ResponseItem struct {
	ID    string  `json:"item"`
	Score float64 `json:"score"`
}

// Fonction pour envoyer une requête de recommandation à Metarank
func sendRecommendRequest(modelName string, request RecommendRequest) (RecommendResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return RecommendResponse{}, err
	}

	url := fmt.Sprintf("http://localhost:8080/recommend/%s", modelName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return RecommendResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RecommendResponse{}, err
	}

	fmt.Println("Response from Metarank:", string(body))

	var recommendResponse RecommendResponse
	err = json.Unmarshal(body, &recommendResponse)
	if err != nil {
		return RecommendResponse{}, err
	}

	return recommendResponse, nil
}

// Fonction de test pour vérifier l'intégration avec le modèle de recommandations
func TestMetarankRecommendIntegration(t *testing.T) {
	request := RecommendRequest{
		Count: 5,
		User:  "user1",
		Items: []string{"69034642", "64835416"}, // Le contexte de recommandation, par ex. un item de référence
	}

	modelName := "similar" // Remplacez par le nom du modèle que vous utilisez pour les recommandations
	response, err := sendRecommendRequest(modelName, request)
	if err != nil {
		t.Fatalf("Failed to send recommend request: %v", err)
	}

	if len(response.Items) == 0 {
		t.Fatalf("Expected recommended items, got none")
	}

	for _, item := range response.Items {
		fmt.Printf("Recommended Item ID: %s, Score: %f\n", item.ID, item.Score)
	}
}
