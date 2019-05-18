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

	filename := "./data/bitcoin.net"
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

	fmt.Printf("Vertices: %d\n", len(vertices))
	f.WriteString("*vertices " + strconv.Itoa(len(vertices)) + "\n")

	c := 1
	for _, vertex := range vertices {
		vertex.Index = c
		f.WriteString(fmt.Sprintf("%d \"%s\"\n", c, vertex.Label))
		c++
	}

	fmt.Printf("Arcs: %d\n", len(arcs))
	f.WriteString("*arcs " + strconv.Itoa(len(arcs)) + "\n")
	for _, arc := range arcs {
		f.WriteString(fmt.Sprintf("%d %d %d\n", arc.From.Index, arc.To.Index, arc.Weight))
		if arc.Weight == 0 {
			fmt.Printf("Arc (%d, %d) has weight 0\n", arc.From.Index, arc.To.Index)
		}
	}

	f.Close()
}
