package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type AIModelConnector struct {
	Client *http.Client
}

type Payload struct {
	Inputs string `json:"inputs"`
}

type Inputs struct {
	Table map[string][]string `json:"table"`
	Query string              `json:"query"`
}

type Response struct {
	Answer      string   `json:"answer"`
	Coordinates [][]int  `json:"coordinates"`
	Cells       []string `json:"cells"`
	Aggregator  string   `json:"aggregator"`
}

func CsvToSlice(data string) (map[string][]string, error) {
	reader := csv.NewReader(strings.NewReader(data))

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	result := make(map[string][]string)

	headers := records[0]

	for _, header := range headers {
		result[header] = []string{}
	}

	for _, record := range records[1:] {
		for i, value := range record {
			header := headers[i]
			result[header] = append(result[header], value)
		}
	}

	return result, nil // TODO: replace this
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq", bytes.NewBuffer(jsonData))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Response{}, err
	}

	return response, nil // TODO: replace this
}

func callGPT2Model(inputText string) (string, error) {
	url := "https://api-inference.huggingface.co/models/gpt2"
	token := os.Getenv("HUGGINGFACE_TOKEN")

	payload := Payload{
		Inputs: inputText,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleJawab(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("jawab.html")
	test := r.Header.Get("pertanyaan")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, test)

}

// func handleUpload(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != "POST" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	file, _, err := r.FormFile("picker")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	defer file.Close()

// 	csvData, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	err = godotenv.Load()
// 	if err != nil {
// 		fmt.Println("Error loading .env file")
// 	}

// 	client := &http.Client{}

// 	connector := AIModelConnector{Client: client}

// 	data := string(csvData[:])

// 	// Convert CSV data to a slice
// 	table, err := CsvToSlice(data)
// 	if err != nil {
// 		fmt.Println("Error converting CSV to slice:", err)
// 		return
// 	}

// 	// Prepare the payload for AI model connection
// 	payload := Inputs{
// 		Table: table,
// 		Query: "total energy consumed?",
// 	}

// 	// Token for Huggingface model API (example token, replace with actual token)
// 	token := os.Getenv("HUGGINGFACE_TOKEN")
// 	if token == "" {
// 		fmt.Println("Error: HUGGINGFACE_TOKEN is not set")
// 		return
// 	}

// 	// Connect to AI model and get the response
// 	response, err := connector.ConnectAIModel(payload, token)
// 	if err != nil {
// 		fmt.Println("Error connecting to AI model:", err)
// 		return
// 	}

// 	if response.Aggregator == "SUM" {
// 		sum := 0.0
// 		for _, cell := range response.Cells {
// 			value, err := strconv.ParseFloat(strings.Trim(cell, " "), 64)
// 			if err != nil {
// 				fmt.Println("Error converting cell to integer:", err)
// 				return
// 			}
// 			sum += value
// 		}
// 		q := payload.Query
// 		responseText, err := callGPT2Model("provide recommendations for: " + q + " " + strconv.FormatFloat(sum, 'f', 2, 64) + "kWh")
// 		if err != nil {
// 			fmt.Println("Error calling GPT-2 model:", err)
// 			return
// 		}
// 		fmt.Printf("%s. %.2f\n\n%s", q, sum, responseText)
// 	} else if response.Aggregator == "AVERAGE" {
// 		fmt.Printf("AI Model Response: %+v\n", response.Cells)
// 	} else {
// 		fmt.Printf("AI Model Response: %+v\n", response)
// 	}
// }

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/jawab", handleJawab)
	fmt.Println("Server berjalan di port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Gagal memulai server:", err)
		os.Exit(1)
	}

	// TODO: answer here
}
