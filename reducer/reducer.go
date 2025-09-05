package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"

	mapreduce "Prog/prog" // Pacchetto gRPC generato da protoc

	"google.golang.org/grpc"
)

// Configurazione del reducer
type Config struct {
	Reducers []Reducer `json:"reducers"`
}

// Struttura che rappresenta un reducer
type Reducer struct {
	ID    string `json:"id"`
	Port  int    `json:"port"`
	Range []int  `json:"range"` // Range in cui il reducer opererà
}

// ReducerServer implementa il server gRPC per il reducer
type ReducerServer struct {
	reducer Reducer // Usa il tipo Reducer definito nel pacchetto principale
	mapreduce.UnimplementedMapReduceServiceServer
}

// Reduce è il metodo che il server gRPC espone per ricevere e elaborare i dati dai worker
func (s *ReducerServer) Reduce(ctx context.Context, req *mapreduce.MapResult) (*mapreduce.ReduceResult, error) {
	// Ricevi i dati dal worker
	data := req.GetSortedNumbers()

	// Filtra i dati in base al range del reducer
	filteredData := filterDataByRange(data, s.reducer.Range)

	// Ordina i dati filtrati
	sort.Slice(filteredData, func(i, j int) bool {
		return filteredData[i] < filteredData[j]
	})

	// Scrivi i dati ordinati nel file condiviso
	err := writeToFile(s.reducer.ID, filteredData)
	if err != nil {
		return nil, fmt.Errorf("errore durante la scrittura nel file: %v", err)
	}

	// Restituisci i dati ordinati senza messaggi aggiuntivi
	return &mapreduce.ReduceResult{
		FinalSortedNumbers: filteredData,
	}, nil
}

// filterDataByRange filtra i dati in base al range definito dal reducer
func filterDataByRange(data []int32, rangeBounds []int) []int32 {
	var filteredData []int32
	for _, num := range data {
		if num >= int32(rangeBounds[0]) && num <= int32(rangeBounds[1]) {
			filteredData = append(filteredData, num)
		}
	}
	return filteredData
}

// writeToFile scrive i dati ordinati nel file di output, sincronizzando l'accesso al file
var fileLock sync.Mutex
var outputFileName = "final_output.txt"

// writeToFile scrive i dati nel file di output in modo sicuro (bloccando l'accesso al file)
func writeToFile(reducerID string, data []int32) error {
	// Blocca l'accesso al file per evitare conflitti di scrittura
	fileLock.Lock()
	defer fileLock.Unlock()

	// Apri il file di output in modalità append
	file, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Errore nell'aprire il file di output: %v", err)
		return err
	}
	defer file.Close()

	// Scrivi l'intestazione per la sezione del reducer
	_, err = file.WriteString(fmt.Sprintf("Reducer %s:\n", reducerID))
	if err != nil {
		log.Printf("Errore nell'intestazione del reducer %s: %v", reducerID, err)
		return err
	}

	// Scrivi i dati ordinati
	for _, num := range data {
		_, err = file.WriteString(fmt.Sprintf("%d\n", num))
		if err != nil {
			log.Printf("Errore nella scrittura del numero %d: %v", num, err)
			return err
		}
	}

	return nil
}

// Configurazione del server gRPC per il reducer
func main() {
	// Carica la configurazione dal file JSON
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Errore nel caricare la configurazione: %v", err)
	}

	// Avvia il server per ciascun reducer configurato
	for _, reducer := range config.Reducers {
		go startReducerServer(reducer)
	}

	// Mantieni il server attivo
	select {}
}

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

// Avvia un server gRPC per il reducer specificato
func startReducerServer(reducer Reducer) {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(reducer.Port))
	if err != nil {
		log.Fatalf("Errore nell'ascolto della porta %d: %v", reducer.Port, err)
	}

	server := grpc.NewServer()
	reducerServer := &ReducerServer{reducer: reducer}
	mapreduce.RegisterMapReduceServiceServer(server, reducerServer)

	log.Printf("Reducer %s in ascolto sulla porta %d", reducer.ID, reducer.Port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Errore nel server gRPC per il reducer %s: %v", reducer.ID, err)
	}
}
