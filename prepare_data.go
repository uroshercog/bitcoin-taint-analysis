package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type entry struct {
	Addresses          []string `json:"addresses"`
	Value              uint64   `json:"value"`
	RequiredSignatures string   `json:"required_signatures"`
}

type transaction struct {
	Inputs  []entry `json:"inputs"`
	Outputs []entry `json:"outputs"`
}

type vertex struct {
	Index   int
	Label   string
	Address string
}

type arc struct {
	From   *vertex
	To     *vertex
	Weight uint64
}

func main() {
	vertices := make(map[string]*vertex)
	var arcs []*arc

	done := false
	filepath.Walk("./data/bitcoin_1_month", func(path string, info os.FileInfo, err error) error {
		if done {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		br := bufio.NewReader(f)

		fmt.Printf("Chunk: %s\n", path)
		for {
			b, err := br.ReadBytes('\n')
			if len(b) == 0 {
				break
			}
			if err != nil {
				panic(err)
			}

			tx := new(transaction)
			if err := json.Unmarshal(b, tx); err != nil {
				panic(err)
			}

			// Iterate over the inputs
			for _, input := range tx.Inputs {
				// Iterate over the outputs
				for _, output := range tx.Outputs {
					if len(input.Addresses) > 1 || len(output.Addresses) > 1 {
						fmt.Println("Multisig")
						continue
					}

					if _, ok := vertices[input.Addresses[0]]; !ok {
						vertices[input.Addresses[0]] = &vertex{
							Address: input.Addresses[0],
							Index:   len(vertices),
							Label:   strconv.FormatInt(int64(len(vertices)), 10),
						}
					}
					if _, ok := vertices[output.Addresses[0]]; !ok {
						vertices[output.Addresses[0]] = &vertex{
							Address: output.Addresses[0],
							Index:   len(vertices),
							Label:   strconv.FormatInt(int64(len(vertices)), 10),
						}
					}

					arcs = append(arcs, &arc{
						From:   vertices[input.Addresses[0]],
						To:     vertices[output.Addresses[0]],
						Weight: output.Value,
					})
				}
			}
		}

		return nil
	})

	filename := "./data/bitcoin.edgelist"
	if _, err := os.Stat(filename); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		// Exists, remove
		if err := os.Remove(filename); err != nil {
			panic(err)
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	for _, arc := range arcs {
		f.WriteString(fmt.Sprintf("%s %s %s\n", arc.From.Address, arc.To.Address, "{'weight': "+strconv.FormatUint(arc.Weight, 10)+"}"))
	}

	f.Close()
}
