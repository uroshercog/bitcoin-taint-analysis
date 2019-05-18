package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Label     string
	Address   string
	ClusterID int
}

type arc struct {
	From   int
	To     *vertex
	Weight uint64
}

func main() {
	clusterCount := 0
	clusters := make(map[int][]*vertex)
	vertices := make(map[string]*vertex)
	var arcs []*arc

	filepath.Walk("../data/bitcoin_1_month", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		// done = true

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

			// Iterate over the outputs
			for _, output := range tx.Outputs {
				if len(output.Addresses) > 1 {
					fmt.Println("Multisig")
					continue
				}

				if _, ok := vertices[output.Addresses[0]]; !ok {
					vertices[output.Addresses[0]] = &vertex{
						Address: output.Addresses[0],
						Label:   strconv.FormatInt(int64(len(vertices)), 10),
					}
				}
			}

			found := false
			currentClusterID := 0
			// Iterate over the inputs. If it's a multisig address skip it.
			for _, input := range tx.Inputs {
				if len(input.Addresses) > 1 {
					fmt.Println("Multisig")
					continue
				}
				if v, ok := vertices[input.Addresses[0]]; !ok {
					vertices[input.Addresses[0]] = &vertex{
						Address: input.Addresses[0],
						Label:   strconv.FormatInt(int64(len(vertices)), 10),
					}
				} else if ok && !found && v.ClusterID > 0 {
					found = true
					currentClusterID = v.ClusterID
				}
			}

			if !found {
				// No clusters exists, create a new one and add all the vertices to this new cluster.
				clusterCount++
				currentClusterID = clusterCount
				clusters[currentClusterID] = []*vertex{}
			}

			// Now go through all the vertices again and change the cluster id
			// Go over all nodes in other clusters and join them to this cluster
			for _, input := range tx.Inputs {
				if v, ok := vertices[input.Addresses[0]]; ok {
					if v.ClusterID > 0 {
						if c, ok := clusters[v.ClusterID]; ok {
							for _, cl := range c {
								cl.ClusterID = currentClusterID
								clusters[currentClusterID] = append(clusters[currentClusterID], cl)
							}
							// Delete the cluster
							delete(clusters, v.ClusterID)
						}
					} else {
						v.ClusterID = currentClusterID
						clusters[currentClusterID] = append(clusters[currentClusterID], v)
					}
				}
			}

			// Lastly we can create an edge between the cluster and all the target outputs. We cannot be sure
			// whether all the outputs belong to the same cluster.
			for _, output := range tx.Outputs {
				arcs = append(arcs, &arc{
					From:   currentClusterID,
					To:     vertices[output.Addresses[0]],
					Weight: output.Value,
				})
			}
		}

		return nil
	})

	filename := "../data/bitcoin_clustered.net"
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

	var verticesBuffer bytes.Buffer

	for cluster, _ := range clusters {
		verticesBuffer.WriteString(fmt.Sprintf("%d \"c_%d\"\n", cluster, cluster))
	}

	c := len(clusters)
	for _, vertex := range vertices {
		if vertex.ClusterID == 0 {
			vertex.ClusterID = c
			verticesBuffer.WriteString(fmt.Sprintf("%d \"v_%s\"\n", c, vertex.Label))
			c++
		}
	}

	fmt.Printf("Vertices: %d\n", c)
	f.WriteString("*vertices " + strconv.Itoa(c) + "\n")
	f.Write(verticesBuffer.Bytes())

	fmt.Printf("Arcs: %d\n", len(arcs))
	f.WriteString("*arcs " + strconv.Itoa(len(arcs)) + "\n")
	for _, arc := range arcs {
		f.WriteString(fmt.Sprintf("%d %d %d\n", arc.From, arc.To.ClusterID, arc.Weight))
		if arc.Weight == 0 {
			fmt.Printf("Arc (%d, %d) has weight 0\n", arc.From, arc.To.ClusterID)
		}
	}

	fmt.Printf("Clusters: %d\n", len(clusters))

	cl := make([]int, clusterCount)
	for k, _ := range clusters {
		cl = append(cl, k)
	}
	sort.Slice(cl, func(i, j int) bool {
		return len(clusters[cl[i]]) < len(clusters[cl[j]])
	})
	f.Close()

	for k := len(cl) - 1; k > len(cl)-10; k-- {
		v := clusters[cl[k]]
		fmt.Printf("Cluster %d => %d vertices\n", k, len(v))
		/*for _, vx := range v {
			fmt.Printf("\t%s\n", vx.Address)
		}*/
	}
}
