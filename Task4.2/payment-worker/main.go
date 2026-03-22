package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/camunda/zeebe/clients/go/v8/pkg/entities"
	"github.com/camunda/zeebe/clients/go/v8/pkg/worker"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
)

func main() {
	gatewayAddress := os.Getenv("ZEEBE_ADDRESS")
	if gatewayAddress == "" {
		gatewayAddress = "localhost:26500" // Значение по умолчанию для локальной разработки
	}

	fmt.Printf("Connecting to Zeebe Gateway at %s...\n", gatewayAddress)

	client, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         gatewayAddress,
		UsePlaintextConnection: true,
	})
	if err != nil {
		log.Fatalf("Failed to create Zeebe client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	topology, err := client.NewTopologyCommand().Send(ctx)
	if err != nil {
		log.Fatalf("Failed to get Zeebe topology: %v", err)
	}
	fmt.Printf("Connected to Zeebe! Cluster size: %d, Partitions: %d\n", topology.ClusterSize, topology.PartitionsCount)

	// Воркер: Списание средств
	worker1 := client.NewJobWorker().JobType("charge-account").Handler(handleChargeAccount).Open()
	defer worker1.Close()

	// Воркер: Антифрод проверка
	worker2 := client.NewJobWorker().JobType("fraud-check").Handler(handleFraudCheck).Open()
	defer worker2.Close()

	// Воркер: Перевод денег контрагенту
	worker3 := client.NewJobWorker().JobType("transfer-funds").Handler(handleTransferFunds).Open()
	defer worker3.Close()

	// Воркер (Компенсация): Возврат средств
	worker4 := client.NewJobWorker().JobType("refund-account").Handler(handleRefundAccount).Open()
	defer worker4.Close()

	fmt.Println("All workers started and waiting for jobs...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Terminating application...")
}

// handleChargeAccount имитирует списание средств.
func handleChargeAccount(client worker.JobClient, job entities.Job) {
	fmt.Printf("[💰 Charge Account] Processing job %d...\n", job.Key)

	time.Sleep(1 * time.Second)

	fmt.Println("[💰 Charge Account] SUCCESS: Funds successfully held.")

	request, err := client.NewCompleteJobCommand().JobKey(job.Key).Send(context.Background())
	if err != nil {
		log.Printf("Failed to complete job %d: %v", job.Key, err)
		return
	}
	_ = request
}

// handleFraudCheck обращается к "антифрод-системе" и возвращает статус.
func handleFraudCheck(client worker.JobClient, job entities.Job) {
	fmt.Printf("[🛡️ Fraud Check] Processing job %d...\n", job.Key)
	time.Sleep(1 * time.Second)

	// Эмулируем ответ от антифрода. Выбираем случайный из 3 исходов:
	// APPROVED(0), REJECTED(1), MANUAL(2)
	rand.Seed(time.Now().UnixNano())
	outcomeIndex := rand.Intn(3)

	var status string
	switch outcomeIndex {
	case 0:
		status = "APPROVED"
	case 1:
		status = "REJECTED"
	case 2:
		status = "MANUAL"
	}

	fmt.Printf("[🛡️ Fraud Check] OUTCOME: %s\n", status)

	variables := make(map[string]interface{})
	variables["fraudStatus"] = status

	// Создаем команду, передаем переменные и отправляем
	cmd, err := client.NewCompleteJobCommand().JobKey(job.Key).VariablesFromMap(variables)
	if err != nil {
		log.Printf("Failed to set variables map: %v", err)
		return
	}

	_, err = cmd.Send(context.Background())
	if err != nil {
		log.Printf("Failed to complete job %d: %v", job.Key, err)
		return
	}
}

// handleTransferFunds переводит деньги продавцу.
func handleTransferFunds(client worker.JobClient, job entities.Job) {
	fmt.Printf("[🏦 Transfer] Processing job %d...\n", job.Key)
	time.Sleep(1 * time.Second)

	fmt.Println("[🏦 Transfer] SUCCESS: Funds transferred to merchant.")

	_, err := client.NewCompleteJobCommand().JobKey(job.Key).Send(context.Background())
	if err != nil {
		log.Printf("Failed to complete job %d: %v", job.Key, err)
	}
}

// handleRefundAccount - компенсирующая транзакция (срабатывает только при ошибках).
func handleRefundAccount(client worker.JobClient, job entities.Job) {
	fmt.Printf("[⏪ Refund (Compensate)] Processing job %d...\n", job.Key)
	time.Sleep(1 * time.Second)

	fmt.Println("[⏪ Refund (Compensate)] SUCCESS: Funds refunded to customer.")

	_, err := client.NewCompleteJobCommand().JobKey(job.Key).Send(context.Background())
	if err != nil {
		log.Printf("Failed to complete job %d: %v", job.Key, err)
	}
}
