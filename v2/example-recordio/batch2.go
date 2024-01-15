package main

import (
	"bufio"
	"os"
	"sync"
	"syscall"

	"github.com/ibmruntimes/go-recordio/v2"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func readFdWriteIO(wg *sync.WaitGroup, fd uintptr, stdio int32) {
	defer wg.Done()
	file := os.NewFile(fd, "tmp")
	scanner := bufio.NewScanner(file)
	stream := zosrecordio.Stdout()
	if stdio == 2 {
		stream = zosrecordio.Stderr()
	}
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		data := []byte(line)
		utils.AtoE(data)
		stream.Fwrite(data)
	}
	stream.Fclose()
}

func redirect(wgp *sync.WaitGroup) {
	wg := *wgp
	r, w, err := os.Pipe()
	if err == nil {
		syscall.Close(2)
		nfd := utils.Dup2(w.Fd(), 2)
		if nfd == 2 {
			syscall.Close(int(w.Fd()))
			wg.Add(1)
			go readFdWriteIO(&wg, r.Fd(), 2)
		} else {
			os.Exit(1)
		}
	} else {
		os.Exit(1)
	}
	r, w, err = os.Pipe()
	if err == nil {
		syscall.Close(1)
		nfd := utils.Dup2(w.Fd(), 1)
		if nfd == 1 {
			syscall.Close(int(w.Fd()))
			wg.Add(1)
			go readFdWriteIO(&wg, r.Fd(), 1)
		} else {
			os.Exit(1)
		}
	} else {
		os.Exit(1)
	}
}
func unredirect(wgp *sync.WaitGroup) {
	wg := *wgp
	syscall.Close(2)
	syscall.Close(1)
	wg.Wait()
}

func main() {
	var wg sync.WaitGroup
	redirect(&wg)
	println("Hello World")
	println("Redirecting fd 2 to SYSPRINT")
	unredirect(&wg)
}
