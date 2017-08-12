package storage

import (
	"bytes"
	"strconv"
)

// NodeVisitor can be called by Walk to traverse the Node hierarchy.
// The Visit() function is called once per node.
type NodeVisitor interface {
	Visit(*Node) NodeVisitor
}

func (x Node_Type) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(x.String())), nil
}

func (x Node_Logical) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(x.String())), nil
}

func (x Node_Comparison) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(x.String())), nil
}

func walkChildren(v NodeVisitor, node *Node) {
	for _, n := range node.Children {
		WalkNode(v, n)
	}
}

func WalkNode(v NodeVisitor, node *Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	walkChildren(v, node)
}

func PredicateToExprString(p *Predicate) string {
	if p == nil {
		return "<none>"
	}

	var v predicateExpressionPrinter
	WalkNode(&v, p.Root)
	return v.Buffer.String()
}

type predicateExpressionPrinter struct {
	bytes.Buffer
}

func (v *predicateExpressionPrinter) Visit(n *Node) NodeVisitor {
	switch n.NodeType {
	case NodeTypeGroupExpression:
		if len(n.Children) > 0 {
			var op string
			if n.GetLogical() == LogicalAnd {
				op = " AND "
			} else {
				op = " OR "
			}
			v.Buffer.WriteString("( ")
			WalkNode(v, n.Children[0])
			for _, e := range n.Children[1:] {
				v.Buffer.WriteString(op)
				WalkNode(v, e)
			}

			v.Buffer.WriteString(" )")
		}

		return nil

	case NodeTypeBooleanExpression:
		WalkNode(v, n.Children[0])
		v.Buffer.WriteByte(' ')
		switch n.GetComparison() {
		case ComparisonEqual:
			v.Buffer.WriteByte('=')
		case ComparisonNotEqual:
			v.Buffer.WriteString("!=")
		case ComparisonStartsWith:
			v.Buffer.WriteString("startsWith")
		case ComparisonRegex:
			v.Buffer.WriteString("=~")
		case ComparisonNotRegex:
			v.Buffer.WriteString("!~")
		}

		v.Buffer.WriteByte(' ')
		WalkNode(v, n.Children[1])
		return nil

	case NodeTypeRef:
		v.Buffer.WriteByte('\'')
		v.Buffer.WriteString(n.GetRefValue())
		v.Buffer.WriteByte('\'')
		return nil

	case NodeTypeLiteral:
		switch val := n.Value.(type) {
		case *Node_StringValue:
			v.Buffer.WriteString(strconv.Quote(val.StringValue))

		case *Node_RegexValue:
			v.Buffer.WriteByte('/')
			v.Buffer.WriteString(val.RegexValue)
			v.Buffer.WriteByte('/')

		case *Node_IntegerValue:
			v.Buffer.WriteString(strconv.FormatInt(val.IntegerValue, 10))

		case *Node_UnsignedValue:
			v.Buffer.WriteString(strconv.FormatUint(val.UnsignedValue, 10))

		case *Node_FloatValue:
			v.Buffer.WriteString(strconv.FormatFloat(val.FloatValue, 'f', 10, 64))

		case *Node_BooleanValue:
			if val.BooleanValue {
				v.Buffer.WriteString("true")
			} else {
				v.Buffer.WriteString("false")
			}
		}

		return nil

	default:
		return v
	}
}
