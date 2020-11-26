package vm

import (
	apipb "github.com/autom8ter/graphik/api"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	_expressionFilter = &apipb.ExpressionFilter{}
	_filter           = &apipb.Filter{}
	_patchFilter      = &apipb.PatchFilter{}
	_subGraphFilter   = &apipb.SubGraphFilter{}
	_nodeDetailFilter = &apipb.NodeDetailFilter{}
	_edgeFilter       = &apipb.EdgeFilter{}
	_channelFilter    = &apipb.ChannelFilter{}
	_meFilter         = &apipb.MeFilter{}
	_edge             = &apipb.Edge{}
	_edgeDetails      = &apipb.EdgeDetails{}
	_edgeDetail       = &apipb.EdgeDetail{}
	_node             = &apipb.Node{}
	_nodeDetails      = &apipb.NodeDetails{}
	_graph            = &apipb.Graph{}
	_meta             = &apipb.Metadata{}
	_nodeC            = &apipb.NodeConstructor{}
	_nodeCs           = &apipb.NodeConstructors{}
	_edgeC            = &apipb.EdgeConstructor{}
	_edgeCs           = &apipb.EdgeConstructors{}
	_nodeChange       = &apipb.NodeChange{}
	_edgeChange       = &apipb.EdgeChange{}
	_change           = &apipb.Change{}
	_message          = &apipb.Message{}
	_outboundMessage  = &apipb.Message{}
	_request          = &apipb.Request{}
	_pong             = &apipb.Pong{}
	_path             = &apipb.Path{}
	_structp          = &structpb.Struct{}
)
