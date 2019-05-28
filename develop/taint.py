import networkx as nx
import queue
import time
import matplotlib.pyplot as plt

ADDRESSES = ["61328"]  # ["1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu"]  #
FILE = "bitcoin_clustered"
TAINT_DELTA = 0.01

start_time = time.time()
# 1. Load the graph into memory
print("=== LOADING GRAPH ===")
G = nx.read_edgelist(f"../data/{FILE}.edgelist", create_using=nx.MultiDiGraph())
print(f"=== GRAPH INFO ({round((time.time() - start_time) * 1000) / 1000}s) ===")
print(f"Number of nodes: {G.number_of_nodes()}")
print(f"Number of edges: {G.number_of_edges()}")
print(f"=== TRAVERSING THIEFS' SUB-GRAPHS ({round((time.time() - start_time) * 1000) / 1000}s) ===")

# 2. Find all the nodes that a specific address we are concerned about has sent funds (to after the hack)
q = queue.Queue()
for addr in ADDRESSES:
    q.put(addr)

while not q.empty():
    n = q.get()
    if 'checked' in G.nodes[n]:
        continue
    G.nodes[n]['checked'] = True

    # 3. Mark all the edges as tainted
    edges = G.edges(n)
    for (a, b) in edges:
        if 'tainted' in G[a][b]:
            continue
        for i in range(0, len(G[a][b])):
            G[a][b][i]['tainted'] = True
        # 4. Follow these new nodes and mark all their outgoing edges as tainted
        q.put(b)
    # 5. Repeat until no more nodes in queue

# 6. We have a tainted sub-graph
# 7. Now go through the whole graph again and calculate the initial taint of every node based on the number of
#    incoming tainted edges (and the number of outgoing tainted edges) (which were created after the hack)
print(f"=== CALCULATING BASE TAINTS ({round((time.time() - start_time) * 1000) / 1000}s) ===")
for n in G.nodes():
    in_edges = G.in_edges(n, data=True)
    out_edges = G.out_edges(n, data=True)

    a, b = 0, 0
    if len(in_edges) > 0:
        t = list(filter(lambda f: 1 if 'tainted' in f[2] and f[2]['tainted'] else 0, in_edges))
        G.nodes[n]['total_out_tainted'] = len(t)

        v = sum(map(lambda x: x[2]['weight'], in_edges))
        # 8. Two ways of calculating the initial taint
        #    - Every edge can add 1/m to the taintness (all tainted transactions add the same to
        #      the total taintness of the node
        a = len(t) / len(in_edges)
        #    - Every edge adds the amount to the taintness (so not all tainted edges add the same amount)
        b = 0 if v == 0 else (sum(map(lambda x: x[2]['weight'], t)) / v)

    G.nodes[n]['taint_average'] = a
    G.nodes[n]['taint_weighted'] = b
    G.nodes[n]['total_out_tainted'] = sum(map(lambda f: 1 if 'tainted' in f[2] and f[2]['tainted'] else 0, out_edges))

# 9. Go over all the nodes again and calculate the actual taint similar to PageRank. Copy the graph so we have the
#    previous state otherwise the calculations wont be correct because we'd be recalculating taints for some nodes
#    with newly assigned taints of neighbour nodes (if that makes sense).
for i in range(0, 5):
    G_copy = G.copy(as_view=True)

    print(f"=== RECALCULATING TAINTS[{i}] ({round((time.time() - start_time) * 1000) / 1000}s) ===")
    for n in G.nodes():
        # Clear any previously set taint
        G.nodes[n]['taint_average'] = 0
        G.nodes[n]['taint_weighted'] = 0
        for (a, b) in G.in_edges(n):
            if G_copy.nodes[a]['total_out_tainted'] > 0:
                G.nodes[n]['taint_average'] += G_copy.nodes[a]['taint_average'] / G_copy.nodes[a]['total_out_tainted']
                G.nodes[n]['taint_weighted'] += G_copy.nodes[a]['taint_weighted'] / G_copy.nodes[a]['total_out_tainted']

# 10. Print out nodes with the highest taint.
nodes = list(G.nodes(data=True))
nodes.sort(reverse=True, key=lambda x: x[1]['taint_average'])

print(f"=== AVERAGE TAINT ({round((time.time() - start_time) * 1000) / 1000}s) ===")
dist_average = {}
total_average = 0
for i in nodes:
    v = round(i[1]['taint_average'] * 100) / 100
    if v < TAINT_DELTA:
        continue

    total_average += 1

    if v not in dist_average:
        dist_average[v] = 1
    else:
        dist_average[v] += 1

nodes.sort(reverse=True, key=lambda x: x[1]['taint_weighted'])
print(f"=== WEIGHTED TAINT ({round((time.time() - start_time) * 1000) / 1000}s) ===")
dist_weighted = {}
total_weighted = 0
for i in nodes:
    v = round(i[1]['taint_weighted'] * 100) / 100

    if v < TAINT_DELTA:
        continue

    total_weighted += 1

    if v not in dist_weighted:
        dist_weighted[v] = 1
    else:
        dist_weighted[v] += 1

print(f"=== DRAWING DISTRIBUTION PLOT ({round((time.time() - start_time) * 1000) / 1000}s) ===")
# 11. Draw the distribution of taints

items = list(dist_average.items())
items.sort(key=lambda x: x[0])
a, b = zip(*items)
b = list(map(lambda x: x / total_average, b))
plt.loglog(a, b, 'ro', label='Average')
print(f"=== TOP 100 AVERAGE ({round((time.time() - start_time) * 1000) / 1000}s) ===")

for i in range(0, 10):
    print(f"{nodes[i][0]} {nodes[i][1]['taint_average']}")

items = list(dist_weighted.items())
items.sort(key=lambda x: x[0])
a, b = zip(*items)
b = list(map(lambda x: x / total_weighted, b))
plt.loglog(a, b, 'bo', label='Weighted')
print(f"=== TOP 100 WEIGHTED ({round((time.time() - start_time) * 1000) / 1000}s) ===")
for i in range(0, 10):
    print(f"{nodes[i][0]} {nodes[i][1]['taint_weighted']}")

plt.legend(loc='upper right')
plt.xlabel("Taint")
plt.ylabel("Probability")

plt.savefig("../data/image_10.png")

print(f"=== DONE ({round((time.time() - start_time) * 1000) / 1000}s) ===")
