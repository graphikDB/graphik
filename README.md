# Graphik

A persistant labelled property graph database written in 100% Go

## Example Queries

### create a node
```graphql
mutation createNode {
    createNode(input: {
        _type: "user"
        name: "coleman"
    }) {
        attributes
    }
}
```

### get nodes

```graphql
query {
    nodes(input:{_type:"user" limit: 5}) {
        attributes
    }
}
```

### delete a node

```graphql
mutation delNode {
    delNode(input: {
        _id: "1a502cc8-3db5-76bb-9548-7ad5d97057bb",
        _type: "user"
    }) {
        count
    }
}
```

### createEdge

```graphql
mutation createEdge {
    createEdge(input: {
        attributes: {
            _type: "fiance"
            _mutual: true
        }
        from: {
            _id: "b7b38163-28dd-5cca-3a20-9503b9d19bfe"
            _type: "user"
        }
        to: {
            _id: "fb98b4d0-a3bb-592c-6a41-8c8fc7ff3306"
            _type: "user"
        }
    }) {
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
    edges(input:{_type:"fiance" limit: 10 filter: [
        {
            key: "_mutual"
            operator: "=="
            value: true
        },
        {
            key: "from.name"
            operator: "!="
            value: "coleman"
        }
    ]}) {
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
    edge(input:{_type:"fiance" _id: "5705ddfc-081f-bd4f-5d5d-1be4ab446867"}) {
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