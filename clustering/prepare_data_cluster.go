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
	Label   string
	Address string
	Cluster *cluster
}

type arc struct {
	From   *cluster
	To     *cluster
	Weight uint64
}

type cluster struct {
	id       int
	vertices []*vertex
}

func assignCluster(v *int, addr string, clusters map[int]*cluster, vertices map[string]*vertex) {
	// Assign this vertex to a new cluster
	*v++
	clusters[*v] = &cluster{
		id:       *v,
		vertices: []*vertex{},
	}

	vertices[addr] = &vertex{
		Address: addr,
		Label:   strconv.FormatInt(int64(len(vertices)), 10),
		Cluster: clusters[*v],
	}

	clusters[*v].vertices = append(clusters[*v].vertices, vertices[addr])
}

func ensureFile(filename string) *os.File {
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

	return f
}

func main() {
	clusterCount := 0
	clusters := make(map[int]*cluster)
	vertices := make(map[string]*vertex)
	var arcs []*arc
	
	filepath.Walk("../data/bitcoin_1_month", func(path string, info os.FileInfo, err error) error {
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

			if len(tx.Inputs) == 0 {
				// fmt.Printf("Transaction has no inputs: %#v\n", tx)
				continue
			}

			if len(tx.Outputs) == 0 {
				fmt.Printf("Transaction has no outputs: %#v\n", tx)
				continue
			}

			// fmt.Printf("Inputs: %d, Outputs: %d\n", len(tx.Inputs), len(tx.Outputs))

			// Iterate over the outputs
			for _, output := range tx.Outputs {
				if len(output.Addresses) > 1 {
					fmt.Println("Multisig")
					continue
				}
				if _, ok := vertices[output.Addresses[0]]; !ok {
					assignCluster(&clusterCount, output.Addresses[0], clusters, vertices)
				}
			}

			// Iterate over the inputs. If it's a multisig address skip it.
			for _, input := range tx.Inputs {
				if len(input.Addresses) > 1 {
					fmt.Println("Multisig")
					continue
				}
				if _, ok := vertices[input.Addresses[0]]; !ok {
					assignCluster(&clusterCount, input.Addresses[0], clusters, vertices)
				}
			}

			// Join all the clusters of the inputs together into a single cluster (the one of the first vertex)
			cluster := vertices[tx.Inputs[0].Addresses[0]].Cluster
			for _, input := range tx.Inputs {
				v := vertices[input.Addresses[0]]
				if v.Cluster.id != cluster.id {
					clId := v.Cluster.id
					// Move all the vertices in this cluster to the new cluster
					for _, vx := range v.Cluster.vertices {
						// Add vertex to new cluster
						vx.Cluster = cluster
						cluster.vertices = append(cluster.vertices, vx)
					}
					delete(clusters, clId)
				}
			}

			// Lastly we can create an edge between the cluster and all the target outputs. We cannot be sure
			// whether all the outputs belong to the same cluster.
			for _, output := range tx.Outputs {
				arcs = append(arcs, &arc{
					From:   cluster,
					To:     vertices[output.Addresses[0]].Cluster,
					Weight: output.Value,
				})
			}
		}

		return nil
	})

	f := ensureFile("../data/bitcoin_clustered.edgelist")
	// Write out an edge list
	for _, arc := range arcs {
		f.WriteString(fmt.Sprintf("%d %d {'weight': %d}\n", arc.From.id, arc.To.id, arc.Weight))
	}
	f.Close()

	f = ensureFile("../data/bitcoin_clustered_mapping.txt")
	// Write out the mapping of addresses to clusters
	for _, cluster := range clusters {
		for _, vx := range cluster.vertices {
			f.WriteString(fmt.Sprintf("%d %s\n", cluster.id, vx.Address))
		}
	}
	f.Close()
}
