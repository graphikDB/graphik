package vm

import (
	"github.com/autom8ter/graphik/gen/go/api"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	_expressionFilter  = &apipb.ExpressionFilter{}
	_filter            = &apipb.Filter{}
	_patchFilter       = &apipb.PatchFilter{}
	_subGraphFilter    = &apipb.SubGraphFilter{}
	_docDetailFilter   = &apipb.DocDetailFilter{}
	_connectionFilter  = &apipb.ConnectionFilter{}
	_channelFilter     = &apipb.ChannelFilter{}
	_meFilter          = &apipb.MeFilter{}
	_connection        = &apipb.Connection{}
	_connectionDetails = &apipb.ConnectionDetails{}
	_connectionDetail  = &apipb.ConnectionDetail{}
	_doc               = &apipb.Doc{}
	_docDetails        = &apipb.DocDetails{}
	_graph             = &apipb.Graph{}
	_meta              = &apipb.Metadata{}
	_docC              = &apipb.DocConstructor{}
	_docCs             = &apipb.DocConstructors{}
	_connectionC       = &apipb.ConnectionConstructor{}
	_connectionCs      = &apipb.ConnectionConstructors{}
	_docChange         = &apipb.DocChange{}
	_connectionChange  = &apipb.ConnectionChange{}
	_change            = &apipb.Change{}
	_message           = &apipb.Message{}
	_outboundMessage   = &apipb.Message{}
	_request           = &apipb.Request{}
	_pong              = &apipb.Pong{}
	_path              = &apipb.Path{}
	_structp           = &structpb.Struct{}
)
