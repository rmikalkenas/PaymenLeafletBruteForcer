package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/yeka/zip"
)

var zipPath string

func decrypt(password string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}

	defer r.Close()

	for _, f := range r.File {
		if f.IsEncrypted() {
			f.SetPassword(password)
		}

		fr, err := f.Open()
		if err != nil {
			return "", err
		}

		defer fr.Close()

		_, err = io.ReadAll(fr)
		if err != nil {
			return "", err
		}

		return password, nil
	}

	return "", fmt.Errorf("no files in zip")
}

func padLeft(s string, length int) string {
	for len(s) < length {
		s = "0" + s
	}

	return s
}

func populatePasswords(c chan<- string, ctx context.Context) {
	defer close(c)

	for i := 1; i <= 999999; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			c <- padLeft(fmt.Sprint(i), 6)
		}
	}
}

func brute(pc <-chan string, rc chan<- string) {
	var wg sync.WaitGroup
	for v := range pc {
		wg.Add(1)
		go func(v2 string) {
			p, err := decrypt(v2)
			if err == nil {
				rc <- p
			}

			wg.Done()
		}(v)
	}
	wg.Wait()
	close(rc)
}

func init() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter path to a zip file (Atsiskaitymo lapelis.zip): ")
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	text = strings.TrimSpace(text)

	if text == "" {
		text = "Atsiskaitymo lapelis.zip"
	}

	if _, err := os.Stat(text); err != nil {
		fmt.Printf("Zip file %v does not exist\n", text)
		os.Exit(1)
	}

	zipPath = text
}

func main() {

	pc := make(chan string)
	rc := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())

	go populatePasswords(pc, ctx)
	go brute(pc, rc)

	var password string

	for p := range rc {
		password = p
		cancel()
	}

	if password != "" {
		fmt.Println("Zip archive's password is:", password)
	} else {
		fmt.Println("Unable to find a correct password :(")
	}
}
