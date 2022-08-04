package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

func askMnemonic() (data.Mnemonic, error) {
	fmt.Println("Enter an Elrond-compatible mnemonic:")
	line, err := readLine()
	if err != nil {
		return "", err
	}

	mnemonic := data.Mnemonic(line)
	return mnemonic, nil
}

func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}
