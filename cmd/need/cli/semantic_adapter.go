package cli

import (
	"fmt"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/semantic"
)

// ----------------------------------------------------------------------------
// Symbol Table Adapters
// ----------------------------------------------------------------------------

// symbolTableAdapter wraps semantic.SymbolTable to implement the SymbolTable interface.
type symbolTableAdapter struct {
	st *semantic.SymbolTable
	// We need to track variables and environments in order for iteration
	varOrder []string
	envOrder []string
}

func (sta *symbolTableAdapter) AddVariable(v interface{}) error {
	astVar, ok := v.(*ast.Variable)
	if !ok {
		return nil // Skip non-variable nodes
	}
	if err := sta.st.AddVariable(astVar); err != nil {
		return err
	}
	sta.varOrder = append(sta.varOrder, astVar.Name)
	return nil
}

func (sta *symbolTableAdapter) AddTarget(t interface{}) error {
	astTarget, ok := t.(*ast.Target)
	if !ok {
		return nil // Skip non-target nodes
	}
	return sta.st.AddTarget(astTarget)
}

func (sta *symbolTableAdapter) AddEnvironment(e interface{}) error {
	astEnv, ok := e.(*ast.Environment)
	if !ok {
		return nil // Skip non-environment nodes
	}
	key := ""
	if astEnv.Name != nil {
		key = *astEnv.Name
	}
	if err := sta.st.AddEnvironment(astEnv); err != nil {
		return err
	}
	sta.envOrder = append(sta.envOrder, key)
	return nil
}

func (sta *symbolTableAdapter) VariableCount() int {
	return len(sta.varOrder)
}

func (sta *symbolTableAdapter) VariableName(i int) string {
	if i < 0 || i >= len(sta.varOrder) {
		return ""
	}
	return sta.varOrder[i]
}

func (sta *symbolTableAdapter) VariableLocation(i int) string {
	if i < 0 || i >= len(sta.varOrder) {
		return ""
	}
	name := sta.varOrder[i]
	if v := sta.st.LookupVariable(name); v != nil {
		return v.Location.String()
	}
	return ""
}

func (sta *symbolTableAdapter) VariableIsLazy(i int) bool {
	if i < 0 || i >= len(sta.varOrder) {
		return false
	}
	name := sta.varOrder[i]
	if v := sta.st.LookupVariable(name); v != nil {
		return v.Lazy
	}
	return false
}

func (sta *symbolTableAdapter) AddConditionalVariable(varDef interface{}, cond interface{}, branchType string, branchIndex int) {
	astVar, ok := varDef.(*ast.Variable)
	if !ok {
		return
	}
	astCond, ok := cond.(*ast.Conditional)
	if !ok {
		return
	}
	def := &semantic.ConditionalVarDef{
		Variable:    astVar,
		Conditional: astCond,
		BranchType:  branchType,
		BranchIndex: branchIndex,
	}
	sta.st.AddConditionalVariable(def)
}

// condVarOrder returns the list of conditional variable names in stable order.
func (sta *symbolTableAdapter) condVarOrder() []string {
	// Build a stable order by iterating the map
	var names []string
	seen := make(map[string]bool)
	for name := range sta.st.ConditionalVars {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	return names
}

func (sta *symbolTableAdapter) ConditionalVarCount() int {
	return len(sta.st.ConditionalVars)
}

func (sta *symbolTableAdapter) ConditionalVarName(i int) string {
	order := sta.condVarOrder()
	if i < 0 || i >= len(order) {
		return ""
	}
	return order[i]
}

func (sta *symbolTableAdapter) ConditionalVarDefCount(i int) int {
	order := sta.condVarOrder()
	if i < 0 || i >= len(order) {
		return 0
	}
	name := order[i]
	return len(sta.st.ConditionalVars[name])
}

func (sta *symbolTableAdapter) ConditionalVarDefLocation(i, j int) string {
	order := sta.condVarOrder()
	if i < 0 || i >= len(order) {
		return ""
	}
	name := order[i]
	defs := sta.st.ConditionalVars[name]
	if j < 0 || j >= len(defs) {
		return ""
	}
	return defs[j].Variable.Location.String()
}

func (sta *symbolTableAdapter) ConditionalVarDefBranch(i, j int) string {
	order := sta.condVarOrder()
	if i < 0 || i >= len(order) {
		return ""
	}
	name := order[i]
	defs := sta.st.ConditionalVars[name]
	if j < 0 || j >= len(defs) {
		return ""
	}
	def := defs[j]
	switch def.BranchType {
	case "if":
		return "if"
	case "elif":
		return fmt.Sprintf("elif[%d]", def.BranchIndex)
	case "else":
		return "else"
	default:
		return def.BranchType
	}
}

func (sta *symbolTableAdapter) TargetCount() int {
	return len(sta.st.Targets)
}

func (sta *symbolTableAdapter) TargetPattern(i int) string {
	if i < 0 || i >= len(sta.st.Targets) {
		return ""
	}
	return semantic.PatternString(&sta.st.Targets[i].Pattern)
}

func (sta *symbolTableAdapter) TargetLocation(i int) string {
	if i < 0 || i >= len(sta.st.Targets) {
		return ""
	}
	return sta.st.Targets[i].Location.String()
}

func (sta *symbolTableAdapter) EnvironmentCount() int {
	return len(sta.envOrder)
}

func (sta *symbolTableAdapter) EnvironmentName(i int) string {
	if i < 0 || i >= len(sta.envOrder) {
		return ""
	}
	return sta.envOrder[i]
}

func (sta *symbolTableAdapter) EnvironmentLocation(i int) string {
	if i < 0 || i >= len(sta.envOrder) {
		return ""
	}
	key := sta.envOrder[i]
	if e, ok := sta.st.Environments[key]; ok {
		return e.Location.String()
	}
	return ""
}

func (sta *symbolTableAdapter) IsAutomatic(name string) bool {
	return sta.st.IsAutomatic(name)
}

func (sta *symbolTableAdapter) IsBuiltin(name string) bool {
	return sta.st.IsBuiltin(name)
}

func (sta *symbolTableAdapter) IsDefined(name string) bool {
	return sta.st.IsDefined(name)
}

// NewSymbolTable creates a new SymbolTable.
func NewSymbolTable() SymbolTable {
	return &symbolTableAdapter{
		st:       semantic.NewSymbolTable(),
		varOrder: make([]string, 0),
		envOrder: make([]string, 0),
	}
}

// ----------------------------------------------------------------------------
// Symbol Collection
// ----------------------------------------------------------------------------

// collectResultAdapter wraps the result of semantic.Collect.
type collectResultAdapter struct {
	st     *semantic.SymbolTable
	errors []error
}

func (cr *collectResultAdapter) SymbolTable() SymbolTable {
	// Build the varOrder and envOrder from the underlying symbol table
	varOrder := make([]string, 0, len(cr.st.Variables))
	for name := range cr.st.Variables {
		varOrder = append(varOrder, name)
	}

	envOrder := make([]string, 0, len(cr.st.Environments))
	for name := range cr.st.Environments {
		envOrder = append(envOrder, name)
	}

	return &symbolTableAdapter{
		st:       cr.st,
		varOrder: varOrder,
		envOrder: envOrder,
	}
}

func (cr *collectResultAdapter) HasErrors() bool {
	return len(cr.errors) > 0
}

func (cr *collectResultAdapter) ErrorCount() int {
	return len(cr.errors)
}

func (cr *collectResultAdapter) GetError(i int) error {
	if i < 0 || i >= len(cr.errors) {
		return nil
	}
	return cr.errors[i]
}

func (cr *collectResultAdapter) Errors() []error {
	return cr.errors
}

// CollectSymbols performs Pass 1: Symbol Collection on the parsed statements.
// It returns a CollectResult that provides access to the symbol table and any errors.
func CollectSymbols(result NeedfileResult) CollectResult {
	// Extract the underlying ast.Statement slice from the result
	stmts := make([]ast.Statement, 0)
	for _, stmt := range result.Statements() {
		if raw, ok := stmt.(interface{ Raw() interface{} }); ok {
			if astStmt, ok := raw.Raw().(ast.Statement); ok {
				stmts = append(stmts, astStmt)
			}
		}
	}

	collectResult := semantic.Collect(stmts)
	return &collectResultAdapter{
		st:     collectResult.SymbolTable,
		errors: collectResult.Errors,
	}
}

// ----------------------------------------------------------------------------
// Capture Validation Adapters
// ----------------------------------------------------------------------------

// captureResultAdapter wraps the result of semantic.ValidateCaptures.
type captureResultAdapter struct {
	cr          *semantic.CaptureResult
	targetOrder []*ast.Target
	interpOrder []*ast.Target
}

func (cra *captureResultAdapter) HasErrors() bool {
	return cra.cr.HasErrors()
}

func (cra *captureResultAdapter) ErrorCount() int {
	return len(cra.cr.Errors)
}

func (cra *captureResultAdapter) GetError(i int) error {
	if i < 0 || i >= len(cra.cr.Errors) {
		return nil
	}
	return cra.cr.Errors[i]
}

func (cra *captureResultAdapter) Errors() []error {
	return cra.cr.Errors
}

func (cra *captureResultAdapter) CaptureCount() int {
	return len(cra.targetOrder)
}

func (cra *captureResultAdapter) TargetPattern(i int) string {
	if i < 0 || i >= len(cra.targetOrder) {
		return ""
	}
	return semantic.PatternString(&cra.targetOrder[i].Pattern)
}

func (cra *captureResultAdapter) CaptureNames(i int) []string {
	if i < 0 || i >= len(cra.targetOrder) {
		return nil
	}
	target := cra.targetOrder[i]
	if info, ok := cra.cr.Captures[target]; ok {
		return info.Names
	}
	return nil
}

func (cra *captureResultAdapter) InterpolationCount() int {
	return len(cra.interpOrder)
}

func (cra *captureResultAdapter) InterpolationTargetPattern(i int) string {
	if i < 0 || i >= len(cra.interpOrder) {
		return ""
	}
	return semantic.PatternString(&cra.interpOrder[i].Pattern)
}

func (cra *captureResultAdapter) InterpolationNames(i int) []string {
	if i < 0 || i >= len(cra.interpOrder) {
		return nil
	}
	target := cra.interpOrder[i]
	if info, ok := cra.cr.Interpolations[target]; ok {
		return info.Names
	}
	return nil
}

// ValidateCaptures performs Pass 2: Capture Validation on the symbol table.
// It returns a CaptureResult that provides access to captures, interpolations, and errors.
func ValidateCaptures(collectResult CollectResult) CaptureResult {
	// Extract the underlying semantic.SymbolTable from the adapter
	var st *semantic.SymbolTable
	if sta, ok := collectResult.SymbolTable().(*symbolTableAdapter); ok {
		st = sta.st
	} else {
		// Shouldn't happen in normal use
		return &captureResultAdapter{
			cr:          &semantic.CaptureResult{},
			targetOrder: make([]*ast.Target, 0),
			interpOrder: make([]*ast.Target, 0),
		}
	}

	cr := semantic.ValidateCaptures(st)

	// Build stable order for targets with captures
	targetOrder := make([]*ast.Target, 0, len(cr.Captures))
	for _, target := range st.Targets {
		if _, ok := cr.Captures[target]; ok {
			targetOrder = append(targetOrder, target)
		}
	}

	// Build stable order for targets with interpolations
	interpOrder := make([]*ast.Target, 0, len(cr.Interpolations))
	for _, target := range st.Targets {
		if _, ok := cr.Interpolations[target]; ok {
			interpOrder = append(interpOrder, target)
		}
	}

	return &captureResultAdapter{
		cr:          cr,
		targetOrder: targetOrder,
		interpOrder: interpOrder,
	}
}

// ----------------------------------------------------------------------------
// Reference Validation Adapters
// ----------------------------------------------------------------------------

// referenceResultAdapter wraps the result of semantic.ValidateReferences.
type referenceResultAdapter struct {
	rr *semantic.ReferenceResult
}

func (rra *referenceResultAdapter) HasErrors() bool {
	return rra.rr.HasErrors()
}

func (rra *referenceResultAdapter) ErrorCount() int {
	return len(rra.rr.Errors)
}

func (rra *referenceResultAdapter) GetError(i int) error {
	if i < 0 || i >= len(rra.rr.Errors) {
		return nil
	}
	return rra.rr.Errors[i]
}

func (rra *referenceResultAdapter) Errors() []error {
	return rra.rr.Errors
}

// ValidateReferences performs Pass 3: Reference Validation on the AST.
// It verifies that all variable references point to defined symbols.
// The captureResult from Pass 2 is needed to recognize captures as valid references.
func ValidateReferences(collectResult CollectResult, stmts []ast.Statement, captureResult CaptureResult) ReferenceResult {
	// Extract the underlying semantic.SymbolTable from the adapter
	var st *semantic.SymbolTable
	if sta, ok := collectResult.SymbolTable().(*symbolTableAdapter); ok {
		st = sta.st
	} else {
		// Shouldn't happen in normal use
		return &referenceResultAdapter{
			rr: &semantic.ReferenceResult{},
		}
	}

	// Extract the underlying semantic.CaptureResult from the adapter
	var cr *semantic.CaptureResult
	if cra, ok := captureResult.(*captureResultAdapter); ok {
		cr = cra.cr
	}

	// Build options
	var opts []semantic.ReferenceOption
	if cr != nil {
		opts = append(opts, semantic.WithCaptures(cr))
	}

	rr := semantic.ValidateReferences(st, stmts, opts...)

	return &referenceResultAdapter{
		rr: rr,
	}
}

// ----------------------------------------------------------------------------
// Dependency Graph Validation Adapters
// ----------------------------------------------------------------------------

// dependencyResultAdapter wraps the result of semantic.ValidateDependencies.
type dependencyResultAdapter struct {
	dr        *semantic.DependencyResult
	nodeOrder []string
	unsatKeys []string
}

func (dra *dependencyResultAdapter) HasErrors() bool {
	return dra.dr.HasErrors()
}

func (dra *dependencyResultAdapter) ErrorCount() int {
	return len(dra.dr.Errors)
}

func (dra *dependencyResultAdapter) GetError(i int) error {
	if i < 0 || i >= len(dra.dr.Errors) {
		return nil
	}
	return dra.dr.Errors[i]
}

func (dra *dependencyResultAdapter) Errors() []error {
	return dra.dr.Errors
}

func (dra *dependencyResultAdapter) NodeCount() int {
	return len(dra.nodeOrder)
}

func (dra *dependencyResultAdapter) NodeName(i int) string {
	if i < 0 || i >= len(dra.nodeOrder) {
		return ""
	}
	return dra.nodeOrder[i]
}

func (dra *dependencyResultAdapter) NodeEdgeCount(i int) int {
	if i < 0 || i >= len(dra.nodeOrder) {
		return 0
	}
	name := dra.nodeOrder[i]
	return len(dra.dr.Graph.Edges[name])
}

func (dra *dependencyResultAdapter) NodeEdge(i, j int) string {
	if i < 0 || i >= len(dra.nodeOrder) {
		return ""
	}
	name := dra.nodeOrder[i]
	edges := dra.dr.Graph.Edges[name]
	if j < 0 || j >= len(edges) {
		return ""
	}
	return edges[j]
}

func (dra *dependencyResultAdapter) PatternTargetCount() int {
	return len(dra.dr.PatternTargets)
}

func (dra *dependencyResultAdapter) PatternTargetPattern(i int) string {
	if i < 0 || i >= len(dra.dr.PatternTargets) {
		return ""
	}
	return semantic.PatternString(&dra.dr.PatternTargets[i].Pattern)
}

func (dra *dependencyResultAdapter) UnsatisfiedDepsCount() int {
	return len(dra.unsatKeys)
}

func (dra *dependencyResultAdapter) UnsatisfiedDepsTarget(i int) string {
	if i < 0 || i >= len(dra.unsatKeys) {
		return ""
	}
	return dra.unsatKeys[i]
}

func (dra *dependencyResultAdapter) UnsatisfiedDepsList(i int) []string {
	if i < 0 || i >= len(dra.unsatKeys) {
		return nil
	}
	key := dra.unsatKeys[i]
	return dra.dr.UnsatisfiedDeps[key]
}

// ValidateDependencies performs Pass 4: Dependency Graph Validation.
// It builds the dependency graph and detects circular dependencies.
func ValidateDependencies(collectResult CollectResult) DependencyResult {
	// Extract the underlying semantic.SymbolTable from the adapter
	var st *semantic.SymbolTable
	if sta, ok := collectResult.SymbolTable().(*symbolTableAdapter); ok {
		st = sta.st
	} else {
		// Shouldn't happen in normal use
		return &dependencyResultAdapter{
			dr:        &semantic.DependencyResult{},
			nodeOrder: make([]string, 0),
			unsatKeys: make([]string, 0),
		}
	}

	dr := semantic.ValidateDependencies(st.Targets)

	// Build stable order for nodes
	nodeOrder := make([]string, 0, len(dr.Graph.Nodes))
	for name := range dr.Graph.Nodes {
		nodeOrder = append(nodeOrder, name)
	}

	// Build stable order for unsatisfied deps keys
	unsatKeys := make([]string, 0, len(dr.UnsatisfiedDeps))
	for key := range dr.UnsatisfiedDeps {
		unsatKeys = append(unsatKeys, key)
	}

	return &dependencyResultAdapter{
		dr:        dr,
		nodeOrder: nodeOrder,
		unsatKeys: unsatKeys,
	}
}
