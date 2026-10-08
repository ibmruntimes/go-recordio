package main

import (
	"fmt"
	"log"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func main() {
	fmt.Println("--- IFAUSAGE Status Check ---")
	rc, err := utils.IfausageStatus()
	if err != nil {
		log.Fatalf("IfausageStatus error: %v", err)
	}
	fmt.Printf("IFAUSAGE REQUEST=STATUS return code: %d (0x%x)\n", rc, rc)

	fmt.Println("\n--- IFAUSAGE Register ---")
	opts := utils.IfausageOptions{
		ProdOwner:  "IBM",
		ProdName:   "PROJECT TITAN",
		ProdVers:   "01.00.00",
		ProdQual:   "NONE",
		ProdID:     "NONE",
		Domain:     utils.IfausageDomainAddrsp,
		Unauthserv: utils.IfausageUnauthservBase,
	}

	rc, token, err := utils.IfausageRegister(opts)
	if err != nil {
		log.Fatalf("IfausageRegister error: %v", err)
	}
	fmt.Printf("IFAUSAGE REQUEST=REGISTER return code: %d (0x%x)\n", rc, rc)
	fmt.Printf("Product Token (PRTOKEN): %x\n", token)

	if rc == 0 {
		fmt.Println("\n--- IFAUSAGE Deregister ---")
		drc, err := utils.IfausageDeregister(token)
		if err != nil {
			log.Fatalf("IfausageDeregister error: %v", err)
		}
		fmt.Printf("IFAUSAGE REQUEST=DEREGISTER return code: %d (0x%x)\n", drc, drc)
	}
}
