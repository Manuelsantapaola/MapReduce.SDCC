package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"

	mapreduce "Prog/prog" // Pacchetto gRPC generato da protoc

	"google.golang.org/grpc"
)

// Worker definisce la struttura del worker con ID e Porta
type Worker struct {
	ID   string `json:"id"`
	Port int    `json:"port"`
}

// Reducer definisce la struttura di un reducer con ID, Porta e Range
type Reducer struct {
	ID    string `json:"id"`
	Port  int    `json:"port"`
	Range []int  `json:"range"` // Range del reducer
}

// Config definisce la struttura del file JSON
type Config struct {
	Workers  []Worker  `json:"workers"`
	Reducers []Reducer `json:"reducers"`
	Data     []int32   `json:"data"`
}

// Variabili globali per la sincronizzazione e coordinamento
var wg sync.WaitGroup      // Sincronizza l'attività di tutti i worker
var mu sync.Mutex          // Mutex per sincronizzare l'accesso alla variabile sharedChunks
var sharedChunks [][]int32 // Variabile condivisa dove ogni worker inserisce il proprio chunk ordinato

// Carica la configurazione dal file JSON
func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Funzione per suddividere i dati in chunk per i worker
func splitDataIntoChunks(data []int32, numChunks int) [][]int32 {
	chunkSize := len(data) / numChunks
	remainder := len(data) % numChunks

	var chunks [][]int32
	start := 0

	for i := 0; i < numChunks; i++ {
		end := start + chunkSize
		if i < remainder {
			end++ // Distribuisce i rimanenti elementi
		}
		chunks = append(chunks, data[start:end])
		start = end
	}
	return chunks
}

// Funzione per eseguire la fase di Map e salvare i dati ordinati
func runMapPhase(workerID string, chunk []int32, index int, wg *sync.WaitGroup) {
	defer wg.Done()

	// Ordinamento del chunk
	log.Printf("Worker %s: Ordinamento in corso per il chunk", workerID)
	sort.Slice(chunk, func(i, j int) bool {
		return chunk[i] < chunk[j]
	})

	// Salva il chunk ordinato nella variabile condivisa
	mu.Lock()
	sharedChunks[index] = chunk
	mu.Unlock()

	log.Printf("Worker %s: Ordinamento completato", workerID)
}

// Funzione per avviare i worker e coordinare la riduzione
func startWorker(config *Config) {
	numWorkers := len(config.Workers)
	sharedChunks = make([][]int32, numWorkers)

	// Suddividi i dati in chunk da inviare ai worker
	chunks := splitDataIntoChunks(config.Data, numWorkers)

	// Per ogni worker invia il proprio chunk da ordinare
	for i, chunk := range chunks {
		wg.Add(1)
		go runMapPhase(config.Workers[i].ID, chunk, i, &wg)
	}

	// Aspetta che tutti i worker completino
	wg.Wait()

	// Uno dei worker diventa il coordinatore per gestire la fase di riduzione
	log.Println("Tutti i worker hanno completato l'ordinamento. Inizio fase di riduzione.")
	coordinateReducers(sharedChunks, config.Reducers)
}

// Funzione per coordinare la fase di riduzione e inviare i chunk ai reducer
func coordinateReducers(chunks [][]int32, reducers []Reducer) {
	// Combina tutti i chunk ordinati in un unico array
	var allData []int32
	for _, chunk := range chunks {
		allData = append(allData, chunk...)
	}

	// Invia i dati combinati ai reducer per la riduzione
	for _, reducer := range reducers {
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(reducer.Port), grpc.WithInsecure())
		if err != nil {
			log.Printf("Errore di connessione al reducer %s: %v", reducer.ID, err)
			continue
		}
		defer conn.Close()

		client := mapreduce.NewMapReduceServiceClient(conn)

		// Invia tutti i dati ordinati al reducer
		_, err = client.Reduce(context.Background(), &mapreduce.MapResult{SortedNumbers: allData})
		if err != nil {
			log.Printf("Errore nell'invio al reducer %s: %v", reducer.ID, err)
		} else {
			log.Printf("Dati inviati al reducer %s", reducer.ID)
		}
	}
}

func main() {
	// Carica la configurazione dal file JSON
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Errore nel caricare la configurazione: %v", err)
	}

	// Avvia la fase di Map per ogni worker
	startWorker(config)
}
