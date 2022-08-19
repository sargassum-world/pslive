package opa

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/pkg/errors"
)

// SQLite Statement AST (in disjunctive normal form)

type Statement interface {
	String() string
	NamedParams() map[string]interface{}
}

type WhereClause struct {
	ID       string
	Column   string // LHS
	Operator string
	Value    interface{} // RHS
}

func (c WhereClause) ParamName() string {
	if c.ID == "" {
		return c.Column
	}
	return fmt.Sprintf("$%s__%s", c.ID, c.Column)
}

func (c WhereClause) String() string {
	return fmt.Sprintf("(%s %s %s)", c.Column, c.Operator, c.ParamName())
}

func (c WhereClause) NamedParams() map[string]interface{} {
	return map[string]interface{}{
		c.ParamName(): c.Value,
	}
}

type SelectionStatement struct {
	Table            string
	Columns          string
	WhereConjunction []WhereClause
}

func (s SelectionStatement) String() string {
	clauses := make([]string, len(s.WhereConjunction))
	for i, clause := range s.WhereConjunction {
		clauses[i] = clause.String()
	}
	return fmt.Sprintf(
		"select %s from %s where %s", s.Columns, s.Table, strings.Join(clauses, " and "),
	)
}

func (s SelectionStatement) NamedParams() map[string]interface{} {
	params := make(map[string]interface{})
	for _, clause := range s.WhereConjunction {
		for name, value := range clause.NamedParams() {
			params[name] = value
		}
	}
	return params
}

type ExistsExpression struct {
	Selection SelectionStatement
}

func (s ExistsExpression) String() string {
	return fmt.Sprintf("exists(%s)", s.Selection)
}

func (s ExistsExpression) NamedParams() map[string]interface{} {
	return s.Selection.NamedParams()
}

type ExistsConjunction []ExistsExpression

func (c ExistsConjunction) String() string {
	expressions := make([]string, len(c))
	for i, expression := range c {
		expressions[i] = expression.String()
	}
	return strings.Join(expressions, " and ")
}

func (c ExistsConjunction) NamedParams() map[string]interface{} {
	params := make(map[string]interface{})
	for _, expression := range c {
		for name, value := range expression.NamedParams() {
			params[name] = value
		}
	}
	return params
}

type ExistsConjunctionDisjunction []ExistsConjunction

func (d ExistsConjunctionDisjunction) String() string {
	expressions := make([]string, len(d))
	for i, expression := range d {
		if len(expression) == 1 {
			expressions[i] = expression.String()
		} else {
			expressions[i] = "(" + expression.String() + ")"
		}
	}
	return strings.Join(expressions, " or ")
}

func (d ExistsConjunctionDisjunction) NamedParams() map[string]interface{} {
	params := make(map[string]interface{})
	for _, expression := range d {
		for name, value := range expression.NamedParams() {
			params[name] = value
		}
	}
	return params
}

type ExistsDNFStatement struct {
	// A select statement in disjunctive normal form
	Disjunction ExistsConjunctionDisjunction
	ResultName  string
}

func (s ExistsDNFStatement) String() string {
	return fmt.Sprintf("select %s as %s", s.Disjunction, s.ResultName)
}

func (s ExistsDNFStatement) NamedParams() map[string]interface{} {
	return s.Disjunction.NamedParams()
}

// Generation of SQL statement AST from Rego AST query bodies

type selectionTerm struct {
	SelectionID string
	Table       string
	Column      string
}

type selectionExpression struct {
	Term  selectionTerm // LHS
	Where WhereClause
}

var operators = map[string]string{
	"eq":    "=",
	"equal": "=",
	"neq":   "!=",
	"lt":    "<",
	"lte":   "<=",
	"gt":    ">",
	"gte":   ">=",
}

var operatorFlips = map[string]string{
	"=":  "=",
	"!=": "!=",
	"<":  ">",
	"<=": ">=",
	">":  "<",
	">=": "<=",
}

func parseSelectionOperator(op string) (operator string, err error) {
	operator, ok := operators[op]
	if !ok {
		return "", errors.Errorf("can't recognize operator: %s", op)
	}
	return operator, nil
}

func parseSelectionTerm(termString, dbName string) (term selectionTerm, err error) {
	if !strings.HasPrefix(termString, dbName+".") {
		return selectionTerm{}, errors.Errorf(
			"can't handle termString which is not a database reference: %s", termString,
		)
	}
	trimmed := strings.TrimPrefix(termString, dbName+".")
	parts := strings.Split(trimmed, "].")
	const numParts = 2
	if len(parts) != numParts {
		return selectionTerm{}, errors.Errorf("can't parse term: %s", trimmed)
	}

	// Table Name & Selection ID
	tableParts := strings.Split(parts[0], "[")
	if len(tableParts) != numParts {
		return selectionTerm{}, errors.Errorf("can't parse table reference: %s", parts[0])
	}
	term.Table = tableParts[0]
	term.SelectionID = tableParts[1]

	// Column Name
	term.Column = parts[1]
	return term, nil
}

func ParseExpression(
	expr *ast.Expr, dbName string,
) (selection selectionExpression, err error) {
	const numOperands = 2
	if len(expr.Operands()) != numOperands {
		return selectionExpression{}, errors.Errorf(
			"can't handle expression with %d operands: %s", len(expr.Operands()), expr,
		)
	}

	if selection.Where.Operator, err = parseSelectionOperator(expr.Operator().String()); err != nil {
		return selectionExpression{}, errors.Wrapf(
			err, "can't parse expression operator: %s", expr.Operator(),
		)
	}

	var valueOnLeft bool
	for i, term := range expr.Operands() {
		if ast.IsConstant(term.Value) {
			if selection.Where.Value, err = ast.JSON(term.Value); err != nil {
				return selectionExpression{}, errors.Wrapf(
					err, "error converting constant to JSON: %s", term.Value,
				)
			}
			valueOnLeft = (i == 0)
			continue
		}
		if selection.Term, err = parseSelectionTerm(term.String(), dbName); err != nil {
			return selectionExpression{}, errors.Wrapf(err, "can't parse term %f", term.Value)
		}
		selection.Where.ID = fmt.Sprintf("%s_%s", selection.Term.SelectionID, selection.Term.Table)
		selection.Where.Column = selection.Term.Column
	}

	if flipped, ok := operatorFlips[selection.Where.Operator]; ok && valueOnLeft {
		// Flip the operator so that the value is on the RHS
		selection.Where.Operator = flipped
	}
	return selection, nil
}

func ParseQuery(
	query ast.Body, dbName string,
) (conjunction ExistsConjunction, err error) {
	expressions := make(map[string]ExistsExpression)
	for _, expr := range query {
		if !expr.IsCall() {
			continue
		}

		selectionExpr, err := ParseExpression(expr, dbName)
		if err != nil {
			return nil, errors.Wrap(err, "can't parse expression")
		}
		key := selectionExpr.Where.ID // includes the table name
		statement := expressions[key]
		statement.Selection.Columns = "1" // select no columns
		statement.Selection.Table = selectionExpr.Term.Table
		statement.Selection.WhereConjunction = append(
			statement.Selection.WhereConjunction, selectionExpr.Where,
		)
		expressions[key] = statement
	}

	conjunction = make([]ExistsExpression, 0, len(expressions))
	for _, expression := range expressions {
		conjunction = append(conjunction, expression)
	}
	return conjunction, nil
}

func ParseQueries(
	queries []ast.Body, dbName string,
) (disjunction ExistsConjunctionDisjunction, err error) {
	disjunction = make([]ExistsConjunction, 0)
	for _, query := range queries {
		// fmt.Println("QUERY:")
		// fmt.Println(query)
		conjunction, err := ParseQuery(query, dbName)
		if err != nil {
			return nil, errors.Wrap(err, "can't parse query")
		}
		// fmt.Println("CONJUNCTION:")
		// fmt.Println(conjunction)
		disjunction = append(disjunction, conjunction)
	}
	return disjunction, nil
}

type SQLiteTranspiler struct {
	DBName string
}

func NewSQLiteTranspiler(dbName string) SQLiteTranspiler {
	return SQLiteTranspiler{
		DBName: dbName,
	}
}

func (t SQLiteTranspiler) Parse(queries []ast.Body) (statement ExistsDNFStatement, err error) {
	statement.ResultName = "result"
	statement.Disjunction, err = ParseQueries(queries, t.DBName)
	if err != nil {
		return ExistsDNFStatement{}, errors.Wrap(err, "can't parse rego queries")
	}
	return statement, nil
}
