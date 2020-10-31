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

### depthSearch

```graphql
query {
    depthSearch(input:{
      path:"user/9e02d1c2-719a-9770-9b41-024c6a90db41" 
      edgeType: "friend" 
      reverse: false
      depth: 1
      limit: 6
    }) {
    	path
        updatedAt
    	createdAt
        attributes
    }
}
```

### createEdge

```graphql
mutation createEdge {
    createEdge(input: {
      	path: "friend"
      	mutual: true
        from: "user/9e02d1c2-719a-9770-9b41-024c6a90db41"
        to: "user/49bbc52e-f5ab-7f8a-81e8-913e380c51b7"
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
    getEdges(input:{type:"friend" limit: 10}) {
    	path
        updatedAt
        attributes
    	mutual
    	from
    	to
    }
}
```

### get single edge

```graphql
query {
    getEdge(input: "friend/8516e0d3-03c0-5b2c-e5b5-fba631c6c1b4") {
        path
        attributes
        from
        to
        createdAt
    }
}
```

