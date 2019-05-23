import networkx as nx

def print_G_info(G):
    print("Nodes: ", G.number_of_nodes())
    print("Edges: ", G.number_of_edges())
    print("Average degree <k> ", 2*G.number_of_edges() / G.number_of_nodes())


# Trace taint leading from the root node to all its successors and their successors ...
def trace_taint(G, root):
    S = []
    S.append(root)
    # set of visited nodes
    visited = set()
    visited.add(root)
    G.node[root]['taint_rank'] = 1
    while S:
        # Pop a vertex from stack to visit next
        n = S.pop()
        # Get all in-edges of the node and
        # calculate the funds of 'n' based on in-edges weights
        n_funds = 0
        # link is a tuple (predecessor, n)
        for link in G.in_edges(n):
            n_funds += G.get_edge_data(link[0], link[1])['weight']
        # Push all the neighbours of n that are not visited in stack
        for neighbor in G[n].keys():
            if neighbor not in visited:
                # TODO: find out addresses of exchanges and dont spread taintrank through them
                # TODO: find out addresses of mixers and DO spread taintrank throguh them
                # TODO: test different approaches for spreading taint (e.g. BFS and distance from root, ...)
                # Currently based on amount of tainted funds received -> gets smaller taintrank
                # Check if entry exists and add to it, else initialize it with this taintrank
                if 'taint_rank' in G.node[neighbor]:
                    G.node[neighbor]['taint_rank'] += G.node[n]['taint_rank'] * (G.get_edge_data(n, neighbor)['weight']/n_funds)
                else:
                    G.node[neighbor]['taint_rank'] = G.node[n]['taint_rank'] * (G.get_edge_data(n, neighbor)['weight']/n_funds)

                S.append(neighbor)
                visited.add(neighbor)
    return visited


# Thief = 1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu
G = nx.read_edgelist("../data/bitcoin.edgelist",create_using=nx.DiGraph())
thief_addr = '1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu'
print_G_info(G)

# Returns subgraph, including all addresses that are reachable from thief_addr
# tainted_G = nx.ego_graph(G, thief_addr, radius=G.number_of_nodes())

# Trace taintedness starting from root, going to all neighbors
tainted_addresses = trace_taint(G, thief_addr)
print("Number of tainted addresses: ", len(tainted_addresses))
# # Dictionary of taint scores of each node
taint_scores = nx.get_node_attributes(G,'taint_rank')
sorted_taint_cores = sorted(taint_scores.items(), key=lambda kv: kv[1], reverse=True)
print(sorted_taint_cores[:10])
