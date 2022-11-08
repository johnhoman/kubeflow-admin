package eksirsa

import (
	"fmt"
)

type Conditions struct {
	ArnLike                           map[string][]string `json:",omitempty"` // nolint: tagliatelle
	ArnLikeIfExists                   map[string][]string `json:",omitempty"` // nolint: tagliatelle
	ArnNotLike                        map[string][]string `json:",omitempty"` // nolint: tagliatelle
	ArnNotLikeIfExists                map[string][]string `json:",omitempty"` // nolint: tagliatelle
	BinaryEquals                      map[string][]string `json:",omitempty"` // nolint: tagliatelle
	BinaryEqualsIfExists              map[string][]string `json:",omitempty"` // nolint: tagliatelle
	Bool                              map[string][]string `json:",omitempty"` // nolint: tagliatelle
	BoolIfExists                      map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateEquals                        map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateEqualsIfExists                map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateNotEquals                     map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateNotEqualsIfExists             map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateLessThan                      map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateLessThanIfExists              map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateLessThanEquals                map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateLessThanEqualsIfExists        map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateGreaterThan                   map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateGreaterThanIfExists           map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateGreaterThanEquals             map[string][]string `json:",omitempty"` // nolint: tagliatelle
	DateGreaterThanEqualsIfExists     map[string][]string `json:",omitempty"` // nolint: tagliatelle
	IpAddress                         map[string][]string `json:",omitempty"` // nolint: tagliatelle
	IpAddressIfExists                 map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NotIpAddress                      map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NotIpAddressIfExists              map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericEquals                     map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericEqualsIfExists             map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericNotEquals                  map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericNotEqualsIfExists          map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericLessThan                   map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericLessThanIfExists           map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericLessThanEquals             map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericLessThanEqualsIfExists     map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericGreaterThan                map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericGreaterThanIfExists        map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericGreaterThanEquals          map[string][]string `json:",omitempty"` // nolint: tagliatelle
	NumericGreaterThanEqualsIfExists  map[string][]string `json:",omitempty"` // nolint: tagliatelle
	Null                              map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringLike                        map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringLikeIfExists                map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringNotLike                     map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringNotLikeIfExists             map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringEquals                      map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringEqualsIfExists              map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringNotEquals                   map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringNotEqualsIfExists           map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringEqualsIgnoreCase            map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringEqualsIgnoreCaseIfExists    map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringNotEqualsIgnoreCase         map[string][]string `json:",omitempty"` // nolint: tagliatelle
	StringNotEqualsIgnoreCaseIfExists map[string][]string `json:",omitempty"` // nolint: tagliatelle
}

type Principal struct {
	Federated string
}

type Statement struct {
	Sid        string      `json:",omitempty"` // nolint: tagliatelle
	Principal  *Principal  `json:",omitempty"`
	Effect     string      // Allow/Deny
	Action     any         `json:",omitempty"` // nolint: tagliatelle
	Resource   any         `json:",omitempty"`
	Conditions *Conditions `json:"Condition,omitempty"` // nolint: tagliatelle
}

type StatementEffect string

const (
	StatementEffectAllow StatementEffect = "Allow"
	StatementEffectDeny  StatementEffect = "Deny"
)

type StatementOption func(s *Statement)

func WithSid(sid string) StatementOption {
	return func(s *Statement) {
		s.Sid = sid
	}
}

func WithEffect(effect StatementEffect) StatementOption {
	return func(s *Statement) {
		s.Effect = string(effect)
	}
}

func WithEffectAllow() StatementOption { return WithEffect("Allow") }

func WithIssuerArn(arn string) StatementOption {
	return func(s *Statement) {
		if s.Principal == nil {
			s.Principal = &Principal{}
		}
		s.Principal.Federated = arn
	}
}

func WithConditions(cond *Conditions) StatementOption {
	return func(s *Statement) {
		s.Conditions = cond
	}
}

func ForServiceAccount(serviceAccountName string, issuer string) StatementOption {
	return func(s *Statement) {
		s.Conditions = &Conditions{
			StringEquals: map[string][]string{
				fmt.Sprintf("%s:sub", issuer): {serviceAccountName},
				fmt.Sprintf("%s:aud", issuer): {"sts.amazonaws.com"},
			},
		}
	}
}

func WithAction(action string) StatementOption {
	return func(s *Statement) {
		s.Action = action
	}
}

func NewStatement(opts ...StatementOption) *Statement {
	s := &Statement{}
	for _, f := range opts {
		f(s)
	}
	return s
}

type document struct {
	Version    string
	Statements []*Statement `json:"Statement"` // nolint: tagliatelle
}

type DocumentOption func(d *document)

func WithStatement(st *Statement) DocumentOption {
	return func(d *document) {
		if d.Statements == nil {
			d.Statements = make([]*Statement, 0)
		}
		d.Statements = append(d.Statements, st)
	}
}

func WithVersion(version string) DocumentOption {
	return func(d *document) {
		d.Version = version
	}
}

func NewDocument(opts ...DocumentOption) *document {
	d := &document{
		Version:    "2012-10-17",
		Statements: make([]*Statement, 0),
	}
	for _, f := range opts {
		f(d)
	}
	return d
}
