// TPD Fundamentos de processamento paralelo e distribuído (eleicao em anel)
// Guilherme, Lucca e Marcelo

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleicao, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho compativel com o numero de processos no anel)
}

var (
	// mutex is used to define a critical section of code
	mutex sync.Mutex
	// vetor de canais para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
	chans = []chan mensagem{
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	simulation_time int  = 0
	pacote_eleicao mensagem
	controle = make(chan int)
	// wg is used to wait for the program to finish
	wg sync.WaitGroup
)

// variaveis de controle do tipo de mensagem
const (
	NovaEleicao     = 0
	VotarEleicao    = 1
	NovoLiderAchado = 2
	Matar           = 3
	Renascer        = 4
)

func ElectionControler(in chan int) {
	defer wg.Done()

	// eleicao
	var temp mensagem
	temp.tipo = NovaEleicao
	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1
	fmt.Printf("Controle: eleicao enviada \n")
	chans[1] <- temp   // p2 faz eleicao
	fmt.Printf("Controle: confirmacao eleicao %d\n", <-in) // receber e imprimir confirmacao

	// desativa p2
	temp.tipo = Matar
	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1
	fmt.Printf("Controle: desativando p2\n")
	chans[1] <- temp   // desativa p2
	fmt.Printf("Controle: desativado confirmacao %d\n", <-in) // receber e imprimir confirmacao

	// eleicao
	temp.tipo = NovaEleicao
	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1
	fmt.Printf("Controle: eleicao enviada \n")
	chans[2] <- temp   // p0 faz eleicao
	fmt.Printf("Controle: confirmacao eleicao %d\n", <-in) // receber e imprimir confirmacao

	// ativa p2
	temp.tipo = Renascer
	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1
	fmt.Printf("Controle: ativar p2 enviado\n")
	chans[1] <- temp   // renasce p2
	fmt.Printf("Controle: ativado confirmacao %d\n", <-in) // receber e imprimir confirmacao

	// nova eleicao
	temp.tipo = NovaEleicao
	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1
	fmt.Printf("Controle: eleicao enviada \n")
	chans[2] <- temp   // p0 faz eleicao
	fmt.Printf("Controle: confirmacao %d\n", <-in) // receber e imprimir confirmacao

	// termina programa
	temp.tipo = 8
	fmt.Println("Controle: terminando")
	chans[2] <- temp
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem) {
	lider := 0
	estou_vivo := true

	for {

		temp := <-in
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]  - %t (lider=%d)\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], estou_vivo, lider)

		switch temp.tipo {
			case NovaEleicao:
			  {
				  pacote_eleicao.tipo = VotarEleicao
				  pacote_eleicao.corpo[0] = -1
				  pacote_eleicao.corpo[1] = -1
				  pacote_eleicao.corpo[2] = -1
				  pacote_eleicao.corpo[TaskId] = TaskId

				  out <- pacote_eleicao

				  fmt.Printf("%2d: enviei proximo anel\n", TaskId)

				  // le a votacao
				  temp := <-in

				  fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]  - %t (lider=%d)\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], estou_vivo, lider)

				  // decide quem eh o lider e escreve
				  for i := 2 ; i >= 0 ; i-- {
					  if temp.corpo[i] != -1 {
						  lider = temp.corpo[i]
						  break
					  }
				  }
				  fmt.Printf("%2d: achei um novo lider: %d\n", TaskId, lider)
				  temp.tipo = NovoLiderAchado

				  // poe a informacao no ring
				  out <- temp

				  // le pra confirmar que todos sabem do novo lider
				  temp = <-in
				  fmt.Printf("%2d: recebi confirmacao de lider %d, [ %d, %d, %d ]  - %t (lider=%d)\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], estou_vivo, lider)

				  // e avisa o controle
				  controle <- 0
				  fmt.Printf("%2d: enviei confirmacao pro controle\n", TaskId)
			  }
		  case VotarEleicao:
			  {
				  // passa seu id pro ring
				  if estou_vivo {
					  temp.corpo[TaskId] = TaskId
				  } else {
					  fmt.Printf("%2d: estou morto, nao posso fazer nada.\n", TaskId)
				  }
				  out <- temp
				  fmt.Printf("%2d: enviei proximo anel\n", TaskId)
			  }
		  case NovoLiderAchado:
			  {
				  for i := 2 ; i >= 0 ; i-- {
					  if temp.corpo[i] != -1 {
						  lider = temp.corpo[i]
						  break
					  }
				  }
				  fmt.Printf("%2d: entendi que temos um novo lider: %d\n", TaskId, lider)
				  out <- temp
				  fmt.Printf("%2d: enviei proximo anel\n", TaskId)
			  }
		  case Matar:
			  {
				  estou_vivo = false
				  fmt.Printf("%2d: fui morto - %t\n", TaskId, estou_vivo)
				  controle <- Matar
				  fmt.Printf("%2d: enviei confirmacao de kill pro controle\n", TaskId)
			  }
		  case Renascer:
			  {
				  estou_vivo = true
				  fmt.Printf("%2d: renasci - %t\n", TaskId, estou_vivo)
				  controle <- Renascer
				  fmt.Printf("%2d: enviei confirmacao de nascimento pro controle\n", TaskId)
			  }
		  default:
			  {
				  // sair do loop (matar todo mundo e terminar o programa)
				  out <- temp
				  fmt.Printf("%2d: finalizando\n", TaskId)
				  wg.Done()
			  }
		}
	}
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

	fmt.Println("\n   Processo controlador criado")

	wg.Wait() // Wait for the goroutines to finish
}
