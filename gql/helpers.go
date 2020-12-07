package gql

import (
	apipb "github.com/autom8ter/graphik/gen/go"
	"github.com/autom8ter/graphik/gen/gql/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func protoIPath(path *model.PathInput) *apipb.Path {
	return &apipb.Path{
		Gtype: path.Gtype,
		Gid:   path.Gid,
	}
}

func protoPath(path *model.Path) *apipb.Path {
	return &apipb.Path{
		Gtype: path.Gtype,
		Gid:   path.Gid,
	}
}

func protoPathC(path *model.PathConstructor) *apipb.PathConstructor {
	p := &apipb.PathConstructor{
		Gtype: path.Gtype,
	}
	if path.Gid != nil {
		p.Gid = *path.Gid
	}
	return p
}

func protoMetadata(m *model.Metadata) *apipb.Metadata {
	return &apipb.Metadata{
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		CreatedBy: protoPath(m.CreatedBy),
		UpdatedBy: protoPath(m.UpdatedBy),
		Version:   uint64(m.Version),
	}
}

func protoDoc(d *model.Doc) *apipb.Doc {
	return &apipb.Doc{
		Path:       protoPath(d.Path),
		Attributes: apipb.NewStruct(d.Attributes),
		Metadata:   protoMetadata(d.Metadata),
	}
}

func protoDocC(d *model.DocConstructor) *apipb.DocConstructor {
	return &apipb.DocConstructor{
		Path:       protoPathC(d.Path),
		Attributes: apipb.NewStruct(d.Attributes),
	}
}

func protoEdit(e *model.Edit) *apipb.Edit {
	return &apipb.Edit{
		Path:       protoIPath(e.Path),
		Attributes: apipb.NewStruct(e.Attributes),
	}
}

func protoConnection(d *model.Connection) *apipb.Connection {
	return &apipb.Connection{
		Path:       protoPath(d.Path),
		Attributes: apipb.NewStruct(d.Attributes),
		Metadata:   protoMetadata(d.Metadata),
		From:       protoPath(d.From),
		To:         protoPath(d.To),
		Directed:   d.Directed,
	}
}

func protoConnectionC(d *model.ConnectionConstructor) *apipb.ConnectionConstructor {
	return &apipb.ConnectionConstructor{
		Path:       protoPathC(d.Path),
		Attributes: apipb.NewStruct(d.Attributes),
		From:       protoIPath(d.From),
		To:         protoIPath(d.To),
		Directed:   d.Directed,
	}
}

func protoFilter(filter *model.Filter) *apipb.Filter {
	f := &apipb.Filter{
		Gtype: filter.Gtype,
		Limit: int32(filter.Limit),
	}
	if filter.Expression != nil {
		f.Expression = *filter.Expression
	}
	if filter.Sort != nil {
		f.Sort = *filter.Sort
	}
	if filter.Index != nil {
		f.Index = *filter.Index
	}
	if filter.Seek != nil {
		f.Seek = *filter.Seek
	}
	if filter.Reverse != nil {
		f.Reverse = *filter.Reverse
	}
	return f
}

func protoEditFilter(filter *model.EditFilter) *apipb.EditFilter {
	return &apipb.EditFilter{
		Filter:     protoFilter(filter.Filter),
		Attributes: apipb.NewStruct(filter.Attributes),
	}
}

func gqlPath(p *apipb.Path) *model.Path {
	return &model.Path{
		Gtype: p.GetGtype(),
		Gid:   p.GetGid(),
	}
}

func gqlMetadata(m *apipb.Metadata) *model.Metadata {
	return &model.Metadata{
		CreatedBy: gqlPath(m.GetCreatedBy()),
		UpdatedBy: gqlPath(m.GetUpdatedBy()),
		CreatedAt: m.GetCreatedAt().AsTime(),
		UpdatedAt: m.GetUpdatedAt().AsTime(),
	}
}

func gqlDoc(d *apipb.Doc) *model.Doc {
	return &model.Doc{
		Path:       gqlPath(d.GetPath()),
		Attributes: d.GetAttributes().AsMap(),
		Metadata:   gqlMetadata(d.GetMetadata()),
	}
}

func gqlDocs(d *apipb.Docs) *model.Docs {
	var docs []*model.Doc
	for _, doc := range d.GetDocs() {
		docs = append(docs, gqlDoc(doc))
	}
	return &model.Docs{
		Docs:     docs,
		SeekNext: d.GetSeekNext(),
	}
}

func gqlConnection(d *apipb.Connection) *model.Connection {
	return &model.Connection{
		Path:       gqlPath(d.GetPath()),
		Attributes: d.GetAttributes().AsMap(),
		Metadata:   gqlMetadata(d.GetMetadata()),
		From:       gqlPath(d.GetFrom()),
		To:         gqlPath(d.GetTo()),
		Directed:   d.GetDirected(),
	}
}

func gqlConnections(d *apipb.Connections) *model.Connections {
	var conns []*model.Connection
	for _, c := range d.GetConnections() {
		conns = append(conns, gqlConnection(c))
	}
	return &model.Connections{
		Connections: conns,
		SeekNext:    d.GetSeekNext(),
	}
}

func gqlAuthorizer(val *apipb.Authorizer) *model.Authorizer {
	return &model.Authorizer{
		Name:       val.GetName(),
		Expression: val.GetExpression(),
	}
}

func gqlTypeValidator(val *apipb.TypeValidator) *model.TypeValidator {
	return &model.TypeValidator{
		Name:       val.GetName(),
		Expression: val.GetExpression(),
	}
}

func gqlIndex(val *apipb.Index) *model.Index {
	return &model.Index{
		Name:        val.GetName(),
		Gtype:       val.GetGtype(),
		Expression:  val.GetExpression(),
		Connections: &val.Connections,
		Docs:        &val.Docs,
	}
}

func gqlIndexes(val *apipb.Indexes) *model.Indexes {
	var vals []*model.Index
	for _, v := range val.GetIndexes() {
		vals = append(vals, gqlIndex(v))
	}
	return &model.Indexes{Indexes: vals}
}

func gqlAuthorizers(val *apipb.Authorizers) *model.Authorizers {
	var vals []*model.Authorizer
	for _, v := range val.GetAuthorizers() {
		vals = append(vals, gqlAuthorizer(v))
	}
	return &model.Authorizers{Authorizers: vals}
}

func gqlTypeValidators(val *apipb.TypeValidators) *model.TypeValidators {
	var vals []*model.TypeValidator
	for _, v := range val.GetValidators() {
		vals = append(vals, gqlTypeValidator(v))
	}
	return &model.TypeValidators{Validators: vals}
}

func gqlSchema(s *apipb.Schema) *model.Schema {
	return &model.Schema{
		ConnectionTypes: s.GetConnectionTypes(),
		DocTypes:        s.GetDocTypes(),
		Authorizers:     gqlAuthorizers(s.GetAuthorizers()),
		Validators:      gqlTypeValidators(s.GetValidators()),
		Indexes:         gqlIndexes(s.GetIndexes()),
	}
}

func protoAggFilter(filter *model.AggregateFilter) *apipb.AggregateFilter {
	return &apipb.AggregateFilter{
		Filter:    protoFilter(filter.Filter),
		Aggregate: filter.Aggregate,
		Field:     filter.Field,
	}
}

func protoChannelFilter(filter *model.ChannelFilter) *apipb.ChannelFilter {
	c := &apipb.ChannelFilter{
		Channel: filter.Channel,
	}
	if filter.Expression != nil {
		c.Expression = *filter.Expression
	}
	return c
}

func protoDepthFilter(filter *model.DepthFilter) *apipb.DepthFilter {
	c := &apipb.DepthFilter{
		Root:  protoIPath(filter.Root),
		Limit: int32(filter.Limit),
	}
	if filter.DocExpression != nil {
		c.DocExpression = *filter.DocExpression
	}
	if filter.DocExpression != nil {
		c.DocExpression = *filter.DocExpression
	}
	if filter.Sort != nil {
		c.Sort = *filter.Sort
	}
	return c
}

func protoConnectionFilter(filter *model.ConnectionFilter) *apipb.ConnectionFilter {
	f := &apipb.ConnectionFilter{
		DocPath:    protoIPath(filter.DocPath),
		Gtype:      filter.Gtype,
		Expression: "",
		Limit:      int32(filter.Limit),
		Sort:       "",
		Seek:       "",
		Reverse:    false,
	}
	if filter.Expression != nil {
		f.Expression = *filter.Expression
	}
	if filter.Sort != nil {
		f.Sort = *filter.Sort
	}
	if filter.Seek != nil {
		f.Seek = *filter.Seek
	}
	if filter.Reverse != nil {
		f.Reverse = *filter.Reverse
	}
	return f
}

func protoExpressionFilter(filter *model.ExpressionFilter) *apipb.ExpressionFilter {
	exp := &apipb.ExpressionFilter{}
	if filter.Expression != nil {
		exp.Expression = *filter.Expression
	}
	return exp
}

func gqlPaths(ps *apipb.Paths) *model.Paths {
	var paths []*model.Path
	for _, p := range ps.GetPaths() {
		paths = append(paths, gqlPath(p))
	}
	return &model.Paths{Paths: paths}
}

func protoIndex(index *model.IndexInput) *apipb.Index {
	i := &apipb.Index{
		Name:       index.Name,
		Gtype:      index.Gtype,
		Expression: index.Expression,
	}
	if index.Docs != nil {
		i.Docs = *index.Docs
	}
	if index.Connections != nil {
		i.Connections = *index.Connections
	}
	return i
}

func protoAuthorizer(auth *model.AuthorizerInput) *apipb.Authorizer {
	return &apipb.Authorizer{
		Name:       auth.Name,
		Expression: auth.Expression,
	}
}

func protoTypeValidator(validator *model.TypeValidatorInput) *apipb.TypeValidator {
	return &apipb.TypeValidator{
		Name:       validator.Name,
		Expression: validator.Expression,
	}
}

func gqlTraversal(traversal *apipb.DocTraversal) *model.DocTraversal {
	return &model.DocTraversal{
		Doc:          gqlDoc(traversal.GetDoc()),
		RelativePath: gqlPaths(traversal.GetRelativePath()),
	}
}

func gqlTraversals(traversals *apipb.DocTraversals) *model.DocTraversals {
	var paths []*model.DocTraversal
	for _, p := range traversals.GetTraversals() {
		paths = append(paths, gqlTraversal(p))
	}
	return &model.DocTraversals{Traversals: paths}
}
