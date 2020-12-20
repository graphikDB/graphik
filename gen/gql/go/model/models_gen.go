// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type AggFilter struct {
	Filter    *Filter   `json:"filter"`
	Aggregate Aggregate `json:"aggregate"`
	Field     *string   `json:"field"`
}

type AuthTarget struct {
	User    *Doc                   `json:"user"`
	Target  map[string]interface{} `json:"target"`
	Peer    string                 `json:"peer"`
	Headers map[string]interface{} `json:"headers"`
}

type Authorizer struct {
	Name            string `json:"name"`
	Method          string `json:"method"`
	Expression      string `json:"expression"`
	TargetRequests  bool   `json:"target_requests"`
	TargetResponses bool   `json:"target_responses"`
}

type AuthorizerInput struct {
	Name            string `json:"name"`
	Method          string `json:"method"`
	Expression      string `json:"expression"`
	TargetRequests  bool   `json:"target_requests"`
	TargetResponses bool   `json:"target_responses"`
}

type Authorizers struct {
	Authorizers []*Authorizer `json:"authorizers"`
}

type AuthorizersInput struct {
	Authorizers []*AuthorizerInput `json:"authorizers"`
}

type ConnectFilter struct {
	DocRef     *RefInput `json:"doc_ref"`
	Gtype      string    `json:"gtype"`
	Expression *string   `json:"expression"`
	Limit      int       `json:"limit"`
	Sort       *string   `json:"sort"`
	Seek       *string   `json:"seek"`
	Reverse    *bool     `json:"reverse"`
}

type Connection struct {
	Ref        *Ref                   `json:"ref"`
	Attributes map[string]interface{} `json:"attributes"`
	Directed   bool                   `json:"directed"`
	From       *Ref                   `json:"from"`
	To         *Ref                   `json:"to"`
}

type ConnectionConstructor struct {
	Ref        *RefConstructor        `json:"ref"`
	Directed   bool                   `json:"directed"`
	Attributes map[string]interface{} `json:"attributes"`
	From       *RefInput              `json:"from"`
	To         *RefInput              `json:"to"`
}

type ConnectionConstructors struct {
	Connections []*ConnectionConstructor `json:"connections"`
}

type Connections struct {
	Connections []*Connection `json:"connections"`
	SeekNext    *string       `json:"seek_next"`
}

type Doc struct {
	Ref        *Ref                   `json:"ref"`
	Attributes map[string]interface{} `json:"attributes"`
}

type DocConstructor struct {
	Ref        *RefConstructor        `json:"ref"`
	Attributes map[string]interface{} `json:"attributes"`
}

type DocConstructors struct {
	Docs []*DocConstructor `json:"docs"`
}

type Docs struct {
	Docs     []*Doc  `json:"docs"`
	SeekNext *string `json:"seek_next"`
}

type Edit struct {
	Ref        *RefInput              `json:"ref"`
	Attributes map[string]interface{} `json:"attributes"`
}

type EditFilter struct {
	Filter     *Filter                `json:"filter"`
	Attributes map[string]interface{} `json:"attributes"`
}

type ExistsFilter struct {
	Gtype      string  `json:"gtype"`
	Expression string  `json:"expression"`
	Seek       *string `json:"seek"`
	Reverse    *bool   `json:"reverse"`
	Index      *string `json:"index"`
}

type ExprFilter struct {
	Expression *string `json:"expression"`
}

type Filter struct {
	Gtype      string  `json:"gtype"`
	Expression *string `json:"expression"`
	Limit      int     `json:"limit"`
	Sort       *string `json:"sort"`
	Seek       *string `json:"seek"`
	Reverse    *bool   `json:"reverse"`
	Index      *string `json:"index"`
}

type Index struct {
	Name        string `json:"name"`
	Gtype       string `json:"gtype"`
	Expression  string `json:"expression"`
	Docs        bool   `json:"docs"`
	Connections bool   `json:"connections"`
}

type IndexInput struct {
	Name        string `json:"name"`
	Gtype       string `json:"gtype"`
	Expression  string `json:"expression"`
	Docs        bool   `json:"docs"`
	Connections bool   `json:"connections"`
}

type Indexes struct {
	Indexes []*Index `json:"indexes"`
}

type IndexesInput struct {
	Indexes []*IndexInput `json:"indexes"`
}

type Message struct {
	Channel   string                 `json:"channel"`
	Data      map[string]interface{} `json:"data"`
	User      *Ref                   `json:"user"`
	Timestamp time.Time              `json:"timestamp"`
	Method    string                 `json:"method"`
}

type OutboundMessage struct {
	Channel string                 `json:"channel"`
	Data    map[string]interface{} `json:"data"`
}

type Peer struct {
	NodeID string `json:"node_id"`
	Addr   string `json:"addr"`
}

type PeerInput struct {
	NodeID string `json:"node_id"`
	Addr   string `json:"addr"`
}

type RaftState struct {
	Leader     string                 `json:"leader"`
	Membership Membership             `json:"membership"`
	Peers      []*Peer                `json:"peers"`
	Stats      map[string]interface{} `json:"stats"`
}

type Ref struct {
	Gtype string `json:"gtype"`
	Gid   string `json:"gid"`
}

type RefConstructor struct {
	Gtype string  `json:"gtype"`
	Gid   *string `json:"gid"`
}

type RefInput struct {
	Gtype string `json:"gtype"`
	Gid   string `json:"gid"`
}

type Refs struct {
	Refs []*Ref `json:"refs"`
}

type Schema struct {
	ConnectionTypes []string        `json:"connection_types"`
	DocTypes        []string        `json:"doc_types"`
	Authorizers     *Authorizers    `json:"authorizers"`
	Validators      *TypeValidators `json:"validators"`
	Indexes         *Indexes        `json:"indexes"`
}

type SearchConnectFilter struct {
	Filter     *Filter                `json:"filter"`
	Gtype      string                 `json:"gtype"`
	Attributes map[string]interface{} `json:"attributes"`
	Directed   bool                   `json:"directed"`
	From       *RefInput              `json:"from"`
}

type SearchConnectMeFilter struct {
	Filter     *Filter                `json:"filter"`
	Gtype      string                 `json:"gtype"`
	Attributes map[string]interface{} `json:"attributes"`
	Directed   bool                   `json:"directed"`
}

type StreamFilter struct {
	Channel    string  `json:"channel"`
	Expression *string `json:"expression"`
}

type Traversal struct {
	Doc           *Doc   `json:"doc"`
	TraversalPath []*Ref `json:"traversal_path"`
	Depth         int    `json:"depth"`
	Hops          int    `json:"hops"`
}

type Traversals struct {
	Traversals []*Traversal `json:"traversals"`
}

type TraverseFilter struct {
	Root                 *RefInput  `json:"root"`
	DocExpression        *string    `json:"doc_expression"`
	ConnectionExpression *string    `json:"connection_expression"`
	Limit                int        `json:"limit"`
	Sort                 *string    `json:"sort"`
	Reverse              *bool      `json:"reverse"`
	Algorithm            *Algorithm `json:"algorithm"`
	MaxDepth             int        `json:"max_depth"`
	MaxHops              int        `json:"max_hops"`
}

type TraverseMeFilter struct {
	DocExpression        *string    `json:"doc_expression"`
	ConnectionExpression *string    `json:"connection_expression"`
	Limit                int        `json:"limit"`
	Sort                 *string    `json:"sort"`
	Reverse              *bool      `json:"reverse"`
	Algorithm            *Algorithm `json:"algorithm"`
	MaxDepth             int        `json:"max_depth"`
	MaxHops              int        `json:"max_hops"`
}

type TypeValidator struct {
	Name              string `json:"name"`
	Gtype             string `json:"gtype"`
	Expression        string `json:"expression"`
	TargetDocs        bool   `json:"target_docs"`
	TargetConnections bool   `json:"target_connections"`
}

type TypeValidatorInput struct {
	Name              string `json:"name"`
	Gtype             string `json:"gtype"`
	Expression        string `json:"expression"`
	TargetDocs        bool   `json:"target_docs"`
	TargetConnections bool   `json:"target_connections"`
}

type TypeValidators struct {
	Validators []*TypeValidator `json:"validators"`
}

type TypeValidatorsInput struct {
	Validators []*TypeValidatorInput `json:"validators"`
}

type Aggregate string

const (
	AggregateCount Aggregate = "COUNT"
	AggregateSum   Aggregate = "SUM"
	AggregateAvg   Aggregate = "AVG"
	AggregateMax   Aggregate = "MAX"
	AggregateMin   Aggregate = "MIN"
	AggregateProd  Aggregate = "PROD"
)

var AllAggregate = []Aggregate{
	AggregateCount,
	AggregateSum,
	AggregateAvg,
	AggregateMax,
	AggregateMin,
	AggregateProd,
}

func (e Aggregate) IsValid() bool {
	switch e {
	case AggregateCount, AggregateSum, AggregateAvg, AggregateMax, AggregateMin, AggregateProd:
		return true
	}
	return false
}

func (e Aggregate) String() string {
	return string(e)
}

func (e *Aggregate) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Aggregate(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Aggregate", str)
	}
	return nil
}

func (e Aggregate) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Algorithm string

const (
	AlgorithmBfs Algorithm = "BFS"
	AlgorithmDfs Algorithm = "DFS"
)

var AllAlgorithm = []Algorithm{
	AlgorithmBfs,
	AlgorithmDfs,
}

func (e Algorithm) IsValid() bool {
	switch e {
	case AlgorithmBfs, AlgorithmDfs:
		return true
	}
	return false
}

func (e Algorithm) String() string {
	return string(e)
}

func (e *Algorithm) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Algorithm(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Algorithm", str)
	}
	return nil
}

func (e Algorithm) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Membership string

const (
	MembershipUnknown   Membership = "UNKNOWN"
	MembershipFollower  Membership = "FOLLOWER"
	MembershipCandidate Membership = "CANDIDATE"
	MembershipLeader    Membership = "LEADER"
	MembershipShutdown  Membership = "SHUTDOWN"
)

var AllMembership = []Membership{
	MembershipUnknown,
	MembershipFollower,
	MembershipCandidate,
	MembershipLeader,
	MembershipShutdown,
}

func (e Membership) IsValid() bool {
	switch e {
	case MembershipUnknown, MembershipFollower, MembershipCandidate, MembershipLeader, MembershipShutdown:
		return true
	}
	return false
}

func (e Membership) String() string {
	return string(e)
}

func (e *Membership) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Membership(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Membership", str)
	}
	return nil
}

func (e Membership) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
