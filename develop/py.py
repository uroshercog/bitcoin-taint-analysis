import networkx as nx
import queue


def print_graph_info(G):
    print("Nodes: ", G.number_of_nodes())
    print("Edges: ", G.number_of_edges())
    print("Average degree <k> ", 2 * G.number_of_edges() / G.number_of_nodes())


# Trace taint leading from the root node to all its successors and their successors ...
def trace_taint(G, root, method='distance'):
    # Use stack for DFS
    # S = []
    # S.append(root)

    # Use queue for BFS
    Q = queue.Queue()
    Q.put(root)

    # set of visited nodes
    visited = set()
    visited.add(root)
    # TaintRank of the root node is 1
    G.node[root]['taint_rank'] = 1

    # Distance from root node
    G.node[root]['distance'] = 0
    while not Q.empty():
        # Pop a vertex from stack/queue to visit next
        # n = S.pop()
        n = Q.get()

        # Get all in-edges of the node and
        # calculate the funds of 'n' based on in-edges weights
        n_funds = 0

        # Link is a tuple (predecessor, n)
        for link in G.in_edges(n):
            n_funds += G.get_edge_data(link[0], link[1])['weight']
        # Push all the neighbours of n that are not visited in stack
        for neighbor in G[n].keys():
            if neighbor not in visited:
                # TODO: find out addresses of exchanges and don't spread taint-rank through them
                # TODO: find out addresses of mixers and DO spread taint-rank through them
                # TODO: test different approaches for spreading taint (e.g. BFS and distance from root, ...)
                # Check if entry exists(can have multiple in links) and add to it else initialize
                # it with this taint-rank.

                if method == 'distance':
                    # Taintrank based on distance from the root node (root starts with distance 0)
                    # Update distance from the root node, based on the predecessor node distance + 1
                    G.node[neighbor]['distance'] = G.node[n]['distance'] + 1
                    if 'taint_rank' in G.node[neighbor]:
                        G.node[neighbor]['taint_rank'] += G.node[n]['taint_rank'] / G.node[neighbor]['distance']
                    else:
                        G.node[neighbor]['taint_rank'] = G.node[n]['taint_rank'] / G.node[neighbor]['distance']

                elif method == 'amount':
                    # Based on amount of tainted funds received -> gets smaller taintrank
                    if 'taint_rank' in G.node[neighbor]:
                        G.node[neighbor]['taint_rank'] += G.node[n]['taint_rank'] * (
                                G.get_edge_data(n, neighbor)['weight'] / n_funds)
                    else:
                        G.node[neighbor]['taint_rank'] = G.node[n]['taint_rank'] * (
                                G.get_edge_data(n, neighbor)['weight'] / n_funds)

                elif method == 'fixed':
                    # Each successor gets same taintrank
                    G.node[neighbor]['taint_rank'] = G.node[n]['taint_rank']

                # Append to stack / queue
                # S.append(neighbor)
                Q.put(neighbor)
                visited.add(neighbor)
    return visited


# Thief = 1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu
G = nx.read_edgelist("../data/bitcoin.edgelist", create_using=nx.DiGraph())
thief_addr = '1MAazCWMydsQB5ynYXqSGQDjNQMN3HFmEu'
print_graph_info(G)

# Returns subgraph, including all addresses that are reachable from thief_addr
# tainted_G = nx.ego_graph(G, thief_addr, radius=1660)

# nodes in tainted graph = 227656
# distance from root to the furthest node reachable is 1660

# Trace taintedness starting from root, going to all neighbors
tainted_addresses = trace_taint(G, thief_addr)
print("Number of tainted addresses: ", len(tainted_addresses))
# Dictionary of taint scores of each node
taint_scores = nx.get_node_attributes(G, 'taint_rank')
sorted_taint_cores = sorted(taint_scores.items(), key=lambda kv: kv[1], reverse=True)
print(sorted_taint_cores[:100])
