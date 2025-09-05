package main

import (
	mapreduce "Prog/prog"
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"google.golang.org/grpc"
)

// Worker definisce la struttura di un worker, con ID e porta
type Worker struct {
	ID   string `json:"id"`
	Port int    `json:"port"`
}

// Config definisce la struttura del file JSON
type Config struct {
	Data    []int32  `json:"data"`
	Workers []Worker `json:"workers"`
}

// MasterServer definisce il server Master
type MasterServer struct {
	mapreduce.UnimplementedMapReduceServiceServer
	workers []Worker // Lista dei worker
}

// Funzione per suddividere i dati in chunk
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

// Funzione per la fase Map
func (s *MasterServer) runMapPhase(config *Config) []mapreduce.MapResult {
	var wg sync.WaitGroup
	chunks := splitDataIntoChunks(config.Data, len(config.Workers))
	var mu sync.Mutex
	var results []mapreduce.MapResult

	log.Println("Inviando chunk ai worker...")

	// Avvia i worker e assegna i chunk
	for i := 0; i < len(config.Workers); i++ {
		wg.Add(1)
		go func(worker Worker, chunk []int32) {
			defer wg.Done()
			workerAddr := "localhost:" + strconv.Itoa(worker.Port)

			conn, err := grpc.Dial(workerAddr, grpc.WithInsecure())
			if err != nil {
				log.Printf("Errore di connessione al worker %v: %v", workerAddr, err)
				return
			}
			defer conn.Close()

			client := mapreduce.NewMapReduceServiceClient(conn)

			// Invio dei chunk al worker per la fase di Map
			dataChunk := &mapreduce.DataChunk{Numbers: chunk}
			res, err := client.Map(context.Background(), dataChunk)
			if err != nil {
				log.Printf("Errore nell'eseguire la mappa con il worker %v: %v", workerAddr, err)
				return
			}

			mu.Lock()
			results = append(results, *res)
			mu.Unlock()
		}(config.Workers[i], chunks[i])
	}

	wg.Wait()
	log.Println("Fase Map completata.")
	return results
}

// Funzione per avviare il server Master
func startMasterServer(port int, workers []Worker) {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Errore nell'aprire la porta %v: %v", port, err)
	}

	masterServer := &MasterServer{workers: workers}

	grpcServer := grpc.NewServer()
	mapreduce.RegisterMapReduceServiceServer(grpcServer, masterServer)

	log.Printf("Master in ascolto sulla porta %v", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Errore nell'avvio del server sulla porta %v: %v", port, err)
	}
}

// Funzione per caricare la configurazione dal file JSON
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

func main() {
	// Carica la configurazione dal file JSON
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Errore nel caricare la configurazione: %v", err)
	}

	// Avvia il server Master su una porta
	go startMasterServer(55000, config.Workers)

	// Crea un'istanza del MasterServer
	masterServer := &MasterServer{workers: config.Workers}

	// Avvia la fase Map
	results := masterServer.runMapPhase(config)

	// Mostra i risultati della fase Map
	log.Printf("Risultati della fase Map: %v", results)
}
