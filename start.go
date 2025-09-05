package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func startMaster() {
	cmd := exec.Command("go", "run", "-tags", "master", "master/master.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Errore nell'avviare il master: %v", err)
	}
	fmt.Println("Master avviato...")
}

func startWorker() {
	cmd := exec.Command("go", "run", "-tags", "worker", "worker/worker.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Errore nell'avviare il worker: %v", err)
	}
	fmt.Println("Worker avviato...")
}
func startReducer() {
	cmd := exec.Command("go", "run", "-tags", "reducer", "reducer/reducer.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Errore nell'avviare il reducer: %v", err)
	}
	fmt.Println("reducer avviato...")
}

func main() {
	// Avvia master e worker in parallelo
	go startMaster()
	go startWorker()
	go startReducer()
	// Aspetta che i processi finiscano (oppure implementa una sincronizzazione adeguata)
	time.Sleep(10 * time.Second) // Puoi usare una sincronizzazione più robusta se necessario
}
