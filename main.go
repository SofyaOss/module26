package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const bufferSize = 5
const updateInterval = 10 * time.Second

func checkPositive(done <-chan int, c <-chan int) <-chan int {
	suitableNumbers := make(chan int)
	go func() {
		for {
			select {
			case <-done:
				close(suitableNumbers)
				return
			case num := <-c:
				if num > 0 {
					select {
					case <-done:
						close(suitableNumbers)
						return
					case suitableNumbers <- num:
					}
				}
			}
		}
	}()
	return suitableNumbers
}

func checkMultiple(done <-chan int, c <-chan int) <-chan int {
	suitableNumbers := make(chan int)
	go func() {
		for {
			select {
			case <-done:
				close(suitableNumbers)
				return
			case num := <-c:
				if num%3 != 0 {
					select {
					case <-done:
						close(suitableNumbers)
						return
					case suitableNumbers <- num:
					}
				}
			}
		}
	}()
	return suitableNumbers
}

type buffer struct {
	m       sync.Mutex
	numbers []int
	size    int
	oldVal  int
}

func NewBuffer(size int) *buffer {
	return &buffer{sync.Mutex{}, make([]int, size), size, 0}
}

func (b *buffer) Extract() []int {
	b.m.Lock()
	defer b.m.Unlock()
	if b.numbers[b.size-1] > 0 {
		returnList := b.numbers
		b.numbers = make([]int, b.size)
		b.oldVal = 0
		return returnList
	} else if b.numbers[b.size-1] == 0 && b.numbers[0] != 0 {
		returnList := b.numbers[:b.oldVal]
		b.numbers = make([]int, b.size)
		b.oldVal = 0
		return returnList
	}
	return nil
}

func (b *buffer) Add(newNum int) {
	b.m.Lock()
	defer b.m.Unlock()
	if b.oldVal < b.size {
		b.numbers[b.oldVal] = newNum
		b.oldVal++
	} else {
		b.oldVal = 0
		b.numbers[b.oldVal] = newNum
	}
	//fmt.Println(b.numbers)
}

func Buffering(done <-chan int, c <-chan int) <-chan int {
	buf := NewBuffer(bufferSize)
	res := make(chan int)
	go func() {
		for {
			select {
			case <-done:
				close(res)
				return
			case num := <-c:
				//fmt.Println(num)
				buf.Add(num)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(updateInterval):
				data := buf.Extract()
				//fmt.Println(data)
				if data != nil {
					for _, el := range data {
						select {
						case <-done:
							return
						case res <- el:
						}
					}
				}
			}
		}
	}()
	return res
}

func Pipeline(done <-chan int, c <-chan int) <-chan int {
	firstCheck := checkPositive(done, c)
	secondCheck := checkMultiple(done, firstCheck)
	output := Buffering(done, secondCheck)
	return output
}

func main() {
	done := make(chan int)
	c := make(chan int)
	scanner := bufio.NewScanner(os.Stdin)
	var line string
	fmt.Println("Чтобы завершить программу введите exit или end. Введите числа в буфер:")
	go func() {
		defer close(done)
		for {
			scanner.Scan()
			line = scanner.Text()
			if strings.EqualFold(line, "exit") || strings.EqualFold(line, "end") {
				fmt.Println("Программа завершила работу!")
				close(c)
				return
			}
			i, err := strconv.Atoi(line)
			if err != nil {
				fmt.Println("Программа обрабатывает только целые числа!")
				continue
			}
			c <- i
		}
	}()
	pipeline := Pipeline(done, c)
	for j := range pipeline {
		fmt.Println("Данные из буфера:", j)
	}
}
