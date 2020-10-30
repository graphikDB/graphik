# Graphik

A persistant labelled property graph database written in 100% Go

## Example Queries

### create a node
```graphql
mutation createNode {
    createNode(input: {
        type: "user"
        attributes: {
          name: "coleman"
        }
    }) {
        id
    	type
        attributes
    }
}
```

### get nodes

```graphql
query {
    nodes(input:{type:"user" limit: 5}) {
    	id
        type
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
        id: "4d26e673-906e-4e7a-acef-19250d2b0b15"
        type: "user"
        patch: {
          name: "Coleman Word"
          email: "colemanword@gmail.com"
          gender: "male"
        }
    }) {
        id
    	type
        attributes
    }
}
```

### delete a node

```graphql
mutation delNode {
    delNode(input: {
        id: "3cd440ba-136d-d99f-facc-51ac2a06b53a",
        type: "user"
    }) {
      count  
    }
}
```

### createEdge

```graphql
mutation createEdge {
    createEdge(input: {
      	type: "fiance"
      	mutual: true
        from: {
          id: "4d26e673-906e-4e7a-acef-19250d2b0b15"
          type: "user"
        }
        to: {
          id: "68f06b51-f8b0-04e9-485a-3e6005a9728d"
          type: "user"
        }
    }) {
        id
        type
        attributes
    	from {
          attributes
        }
        to {
          attributes
        }
    }
}
```

### get edges

```graphql
query {
    edges(input:{type:"fiance" limit: 10 expressions: [
      {
      	key: "from.attributes.name"
      	operator: NEQ
      	value: "coleman"
      },
      {
      	key: "to.attributes.name"
      	operator: NEQ
      	value: "coleman"
      },
    ]}) {
        id
        type
        attributes
    	from {
          attributes
        }
    	to {
          attributes
        }
    }
}
```

### get single edge

```graphql
query {
    edge(input:{type:"fiance" id: "43d442c6-f129-9c42-3e7f-4ef160a7a079"}) {
        id
    	type
    	attributes
    	from {
          attributes
        }
    	to {
          attributes
        }
    }
}
```