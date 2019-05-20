import networkx as nx

def print_G_info(G):
    print("Nodes: ", G.number_of_nodes())
    print("Edges: ", G.number_of_edges())
    print("Average degree <k> ", 2*G.number_of_edges() / G.number_of_nodes())


# Trace taint leading from the root node to all its successors and their successors ...
def trace_taint(G, root):
    S = []
    S.append(root)
    visited = set()
    visited.add(root)
    while S:
        # Pop a vertex from stack to visit next
        n = S.pop()
        # Push all the neighbours of v in stack that are not visited   
        for neighbor in G[n].keys():
            if neighbor not in visited:
                # TODO: Add some taint attribute based on in-links, weights, ...
                S.append(neighbor)
                visited.add(neighbor)
                global_tainted_addresses.append(neighbor)

# Thief = 1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu
G = nx.read_edgelist("../data/bitcoin.edgelist",create_using=nx.DiGraph())
thief_addr = '1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu'
print(nx.is_directed(G))
print_G_info(G)

# Trace taintedness starting from root, going to all neighbors
global_tainted_addresses = [thief_addr]
trace_taint(G, thief_addr)
print("Number of tainted addresses: ", len(global_taint))
