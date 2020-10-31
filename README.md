# Graphik

A persistant labelled property graph database written in 100% Go

## Example Queries

### create a node
```graphql
mutation createNode {
    createNode(input: {
        path: "user"
        attributes: {
          name: "coleman"
        }
    }) {
        path
        attributes
    }
}
```

### get nodes

```graphql
query {
    nodes(input:{type:"user" limit: 100}) {
    	  path
        attributes
    		createdAt
        updatedAt
    }
}
```

### patch node

```graphql
mutation patchNode {
    patchNode(input: {
        path: "user/cac3b142-0e52-2e24-388c-79a9bd50c35f"
        patch: {
          name: "Coleman Word"
          email: "colemanword@gmail.com"
          gender: "male"
        }
    }) {
        path
        attributes
    }
}
```

### delete a node

```graphql
mutation delNode {
    delNode(input: "user/abd70fc5-2520-28c4-6f17-ef3bd5ae6073") {
      count  
    }
}
```

### createEdge

```graphql
mutation createEdge {
    createEdge(input: {
      	path: "friend"
      	direction: UNDIRECTED
        from: "user/cac3b142-0e52-2e24-388c-79a9bd50c35f"
        to: "user/152ceaf3-1f14-898d-efd6-54b18f207386"
    }) {
        path
        attributes
    		from
        to
    }
}
```

### get edges

```graphql
query {
    edges(input:{type:"*" limit: 10}) {
    		path
        updatedAt
        attributes
    		from
    		to
    }
}
```

### get single edge

```graphql
query {
    edge(input: "friend/43a00c20-88e3-8b84-2145-853d59be2456") {
       	path
    		attributes
    		from
    		to
    		createdAt
    }
}
```