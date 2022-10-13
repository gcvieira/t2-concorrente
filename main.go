// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
}

var (
	mutex           sync.Mutex        // mutex is used to define a critical section of code
	simulation_time int        = 0
	chans                      = []chan mensagem { // vetor de canais para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	pacote_eleicao mensagem
	controle       = make(chan int)
	wg             sync.WaitGroup      // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	temp.tipo = 1
	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1

	chans[2] <- temp // pedir eleição para o processo 0
	fmt.Printf("Controle: eleicao enviada \n")

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem) {
	defer wg.Done()

	temp := <-in

	fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])
	temp.corpo[TaskId] = TaskId

	out <- temp
	fmt.Printf("%2d: enviei próximo anel \n", TaskId)

	if TaskId == 0 {
		temp := <-in
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])
		controle <- -5
		fmt.Printf("%2d: enviei confirmação controle \n", TaskId)

	}
	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {

	wg.Add(4) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[2], chans[0])
	go ElectionStage(1, chans[0], chans[1])
	go ElectionStage(2, chans[1], chans[2])

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish
}
