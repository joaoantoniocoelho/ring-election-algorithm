package main

import (
	"fmt"
	"sync"
	"time"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmação da eleição, conhecimento do novo coordenador)
	corpo [3]int // conteúdo da mensagem para colocar os ids (usar um tamanho compatível com o número de processos no anel)
}

var (
	chans          = make([]chan mensagem, 4) // vetor de canais para formar o anel de eleição - chan[0], chan[1] e chan[2] ...
	controle       = make(chan int)
	wg             sync.WaitGroup // wg é usado para esperar o programa terminar
	leaderID       int            // ID do líder inicial
	processes      = 4
	externalReq    = make(chan bool) // canal para simular solicitações externas
	failedProcs    = make([]bool, 4) // vetor para simular falhas nos nós do anel
	failedProcsMux sync.Mutex        // mutex para proteger o acesso ao vetor de falhas
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// comandos para o anel iniciam aqui

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isso)
	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isso)
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// matar os outros processos com mensagens não conhecidas (só para consumir a leitura)
	temp.tipo = 4
	chans[1] <- temp
	chans[2] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskID int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variáveis locais que indicam se este processo é o líder e se está ativo
	var actualLeader int
	var bFailed bool = false         // todos iniciam sem falha
	var electionStarted bool = false // indica se a eleição já foi iniciada neste processo
	var electionMsg mensagem         // mensagem de eleição enviada pelo processo

	actualLeader = leader // indicação do líder veio por parâmetro

	temp := <-in // ler mensagem
	fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskID, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

	switch temp.tipo {
	case 2:
		bFailed = true
		fmt.Printf("%2d: falho %v \n", TaskID, bFailed)
		fmt.Printf("%2d: líder atual %d\n", TaskID, actualLeader)
		controle <- -5
	case 1:
		fmt.Printf("%2d: mensagem de eleição recebida\n", TaskID)
		if !electionStarted {
			electionStarted = true
			electionMsg = temp
			electionMsg.corpo[TaskID] = TaskID // incluir ID do processo na mensagem de eleição
			out <- electionMsg
		} else if TaskID == 0 { // processo que iniciou a eleição
			fmt.Printf("%2d: volta completa. Escolhendo novo coordenador...\n", TaskID)
			maxID := -1
			newLeader := -1
			for i := 0; i < processes; i++ {
				if electionMsg.corpo[i] > maxID {
					maxID = electionMsg.corpo[i]
					newLeader = i
				}
			}
			fmt.Printf("%2d: novo coordenador escolhido: %d\n", TaskID, newLeader)
			leaderID = newLeader // atualizar o ID do líder

			// enviar mensagem para todos conhecerem o novo coordenador
			newLeaderMsg := mensagem{
				tipo:  3,
				corpo: [3]int{newLeader, newLeader, newLeader}, // ID do novo coordenador
			}
			out <- newLeaderMsg

			// Simular solicitações externas ao novo coordenador
			go func() {
				for {
					select {
					case <-externalReq:
						fmt.Printf("Solicitação recebida pelo coordenador %d\n", newLeader)
					default:
						time.Sleep(time.Millisecond * 100)
					}
				}
			}()
		}
	case 3:
		fmt.Printf("%2d: mensagem de conhecimento do novo coordenador recebida. Novo coordenador: %d\n", TaskID, temp.corpo[0])
	default:
		fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskID)
		fmt.Printf("%2d: líder atual %d\n", TaskID, actualLeader)
	}

	// verificar se o líder não está mais ativo
	if actualLeader == TaskID && bFailed {
		fmt.Printf("%2d: coordenador não está mais ativo. Iniciando eleição...\n", TaskID)
		// enviar mensagem de eleição para o próximo processo no sentido do anel
		electionMsg := mensagem{
			tipo:  1,
			corpo: [3]int{TaskID, TaskID, TaskID}, // ID do processo está disputando a coordenação
		}

		go func() {
			time.Sleep(time.Second) // Aguardar um pouco para evitar deadlock
			out <- electionMsg
		}()
	}

	fmt.Printf("%2d: terminei \n", TaskID)
}

func main() {
	wg.Add(processes + 1) // Adiciona uma contagem para cada goroutine e para o controlador

	// criar os processos do anel de eleição
	for i := 0; i < processes; i++ {
		chans[i] = make(chan mensagem)
	}

	// definir o líder inicial como processo 0
	leaderID = 0

	// criar as goroutines para os processos do anel
	for i := 0; i < processes; i++ {
		nextProcessID := (i + 1) % processes // índice do próximo processo no sentido do anel
		go ElectionStage(i, chans[i], chans[nextProcessID], leaderID)
	}

	fmt.Println("\n   Anel de processos criado")

	// criar a goroutine para o processo controlador
	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	// Simular falhas nos nós do anel
	go func() {
		for {
			time.Sleep(time.Second * 5) // Aguardar um pouco antes de simular uma falha
			failedProcsMux.Lock()
			for i := 0; i < processes; i++ {
				if !failedProcs[i] {
					fmt.Printf("Falha no processo %d\n", i)
					failedProcs[i] = true
					// Enviar mensagem de falha para o próximo processo no sentido do anel
					failureMsg := mensagem{
						tipo:  2,
						corpo: [3]int{i, i, i},
					}
					chans[i] <- failureMsg
					break
				}
			}
			failedProcsMux.Unlock()
		}
	}()

	// Simular solicitações externas
	go func() {
		for {
			time.Sleep(time.Second * 3) // Aguardar um pouco antes de simular uma solicitação externa
			externalReq <- true
		}
	}()

	wg.Wait() // Esperar as goroutines terminarem
}
