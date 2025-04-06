# Semantic Search Proposal

Provide a semantic search capability to the agent directory service. This will allow users and agents to find agents based on their skills, capabilities, and metadata.

**Design Considerations**
* Vector search vs TF-IDF/BM25
* In-memory vs external store (go-bleve vs weaviate, pinecone, etc.)

**Worflow**
1. When an agent record is pushed to an ADS server, the agent's metadata is indexed.
2. When an ADS peer server gets notice of the new record via announcement, it will also index the agent's metadata.
3. When a user or agent wants to search for agents, they will send a search request to an ADS server.
4. The ADS server will search its local index and peers (if network is provided) for matching agents.
5. The ADS server will return the list of matching agents to the user or agent.

Add Search() to  
`https://github.com/agntcy/dir/blob/main/server/routing/routing.go`
`https://github.com/agntcy/dir/blob/main/server/routing/routing_remote.go`
`https://github.com/agntcy/dir/blob/main/server/routing/routing_local.go`

Add SearchIndex to   
`https://github.com/agntcy/dir/tree/main/server/store`

## Extra curriculum
* Python bindings
* Browser UI 