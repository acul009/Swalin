// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"
	"rahnit-rmm/ent/device"
	"rahnit-rmm/ent/hostconfig"
	"rahnit-rmm/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// HostConfigQuery is the builder for querying HostConfig entities.
type HostConfigQuery struct {
	config
	ctx        *QueryContext
	order      []hostconfig.OrderOption
	inters     []Interceptor
	predicates []predicate.HostConfig
	withDevice *DeviceQuery
	withFKs    bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the HostConfigQuery builder.
func (hcq *HostConfigQuery) Where(ps ...predicate.HostConfig) *HostConfigQuery {
	hcq.predicates = append(hcq.predicates, ps...)
	return hcq
}

// Limit the number of records to be returned by this query.
func (hcq *HostConfigQuery) Limit(limit int) *HostConfigQuery {
	hcq.ctx.Limit = &limit
	return hcq
}

// Offset to start from.
func (hcq *HostConfigQuery) Offset(offset int) *HostConfigQuery {
	hcq.ctx.Offset = &offset
	return hcq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (hcq *HostConfigQuery) Unique(unique bool) *HostConfigQuery {
	hcq.ctx.Unique = &unique
	return hcq
}

// Order specifies how the records should be ordered.
func (hcq *HostConfigQuery) Order(o ...hostconfig.OrderOption) *HostConfigQuery {
	hcq.order = append(hcq.order, o...)
	return hcq
}

// QueryDevice chains the current query on the "device" edge.
func (hcq *HostConfigQuery) QueryDevice() *DeviceQuery {
	query := (&DeviceClient{config: hcq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := hcq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := hcq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(hostconfig.Table, hostconfig.FieldID, selector),
			sqlgraph.To(device.Table, device.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, hostconfig.DeviceTable, hostconfig.DeviceColumn),
		)
		fromU = sqlgraph.SetNeighbors(hcq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first HostConfig entity from the query.
// Returns a *NotFoundError when no HostConfig was found.
func (hcq *HostConfigQuery) First(ctx context.Context) (*HostConfig, error) {
	nodes, err := hcq.Limit(1).All(setContextOp(ctx, hcq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{hostconfig.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (hcq *HostConfigQuery) FirstX(ctx context.Context) *HostConfig {
	node, err := hcq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first HostConfig ID from the query.
// Returns a *NotFoundError when no HostConfig ID was found.
func (hcq *HostConfigQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = hcq.Limit(1).IDs(setContextOp(ctx, hcq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{hostconfig.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (hcq *HostConfigQuery) FirstIDX(ctx context.Context) int {
	id, err := hcq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single HostConfig entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one HostConfig entity is found.
// Returns a *NotFoundError when no HostConfig entities are found.
func (hcq *HostConfigQuery) Only(ctx context.Context) (*HostConfig, error) {
	nodes, err := hcq.Limit(2).All(setContextOp(ctx, hcq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{hostconfig.Label}
	default:
		return nil, &NotSingularError{hostconfig.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (hcq *HostConfigQuery) OnlyX(ctx context.Context) *HostConfig {
	node, err := hcq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only HostConfig ID in the query.
// Returns a *NotSingularError when more than one HostConfig ID is found.
// Returns a *NotFoundError when no entities are found.
func (hcq *HostConfigQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = hcq.Limit(2).IDs(setContextOp(ctx, hcq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{hostconfig.Label}
	default:
		err = &NotSingularError{hostconfig.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (hcq *HostConfigQuery) OnlyIDX(ctx context.Context) int {
	id, err := hcq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of HostConfigs.
func (hcq *HostConfigQuery) All(ctx context.Context) ([]*HostConfig, error) {
	ctx = setContextOp(ctx, hcq.ctx, "All")
	if err := hcq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*HostConfig, *HostConfigQuery]()
	return withInterceptors[[]*HostConfig](ctx, hcq, qr, hcq.inters)
}

// AllX is like All, but panics if an error occurs.
func (hcq *HostConfigQuery) AllX(ctx context.Context) []*HostConfig {
	nodes, err := hcq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of HostConfig IDs.
func (hcq *HostConfigQuery) IDs(ctx context.Context) (ids []int, err error) {
	if hcq.ctx.Unique == nil && hcq.path != nil {
		hcq.Unique(true)
	}
	ctx = setContextOp(ctx, hcq.ctx, "IDs")
	if err = hcq.Select(hostconfig.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (hcq *HostConfigQuery) IDsX(ctx context.Context) []int {
	ids, err := hcq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (hcq *HostConfigQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, hcq.ctx, "Count")
	if err := hcq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, hcq, querierCount[*HostConfigQuery](), hcq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (hcq *HostConfigQuery) CountX(ctx context.Context) int {
	count, err := hcq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (hcq *HostConfigQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, hcq.ctx, "Exist")
	switch _, err := hcq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (hcq *HostConfigQuery) ExistX(ctx context.Context) bool {
	exist, err := hcq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the HostConfigQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (hcq *HostConfigQuery) Clone() *HostConfigQuery {
	if hcq == nil {
		return nil
	}
	return &HostConfigQuery{
		config:     hcq.config,
		ctx:        hcq.ctx.Clone(),
		order:      append([]hostconfig.OrderOption{}, hcq.order...),
		inters:     append([]Interceptor{}, hcq.inters...),
		predicates: append([]predicate.HostConfig{}, hcq.predicates...),
		withDevice: hcq.withDevice.Clone(),
		// clone intermediate query.
		sql:  hcq.sql.Clone(),
		path: hcq.path,
	}
}

// WithDevice tells the query-builder to eager-load the nodes that are connected to
// the "device" edge. The optional arguments are used to configure the query builder of the edge.
func (hcq *HostConfigQuery) WithDevice(opts ...func(*DeviceQuery)) *HostConfigQuery {
	query := (&DeviceClient{config: hcq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	hcq.withDevice = query
	return hcq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Config []byte `json:"config,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.HostConfig.Query().
//		GroupBy(hostconfig.FieldConfig).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (hcq *HostConfigQuery) GroupBy(field string, fields ...string) *HostConfigGroupBy {
	hcq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &HostConfigGroupBy{build: hcq}
	grbuild.flds = &hcq.ctx.Fields
	grbuild.label = hostconfig.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Config []byte `json:"config,omitempty"`
//	}
//
//	client.HostConfig.Query().
//		Select(hostconfig.FieldConfig).
//		Scan(ctx, &v)
func (hcq *HostConfigQuery) Select(fields ...string) *HostConfigSelect {
	hcq.ctx.Fields = append(hcq.ctx.Fields, fields...)
	sbuild := &HostConfigSelect{HostConfigQuery: hcq}
	sbuild.label = hostconfig.Label
	sbuild.flds, sbuild.scan = &hcq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a HostConfigSelect configured with the given aggregations.
func (hcq *HostConfigQuery) Aggregate(fns ...AggregateFunc) *HostConfigSelect {
	return hcq.Select().Aggregate(fns...)
}

func (hcq *HostConfigQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range hcq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, hcq); err != nil {
				return err
			}
		}
	}
	for _, f := range hcq.ctx.Fields {
		if !hostconfig.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if hcq.path != nil {
		prev, err := hcq.path(ctx)
		if err != nil {
			return err
		}
		hcq.sql = prev
	}
	return nil
}

func (hcq *HostConfigQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*HostConfig, error) {
	var (
		nodes       = []*HostConfig{}
		withFKs     = hcq.withFKs
		_spec       = hcq.querySpec()
		loadedTypes = [1]bool{
			hcq.withDevice != nil,
		}
	)
	if hcq.withDevice != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, hostconfig.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*HostConfig).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &HostConfig{config: hcq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, hcq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := hcq.withDevice; query != nil {
		if err := hcq.loadDevice(ctx, query, nodes, nil,
			func(n *HostConfig, e *Device) { n.Edges.Device = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (hcq *HostConfigQuery) loadDevice(ctx context.Context, query *DeviceQuery, nodes []*HostConfig, init func(*HostConfig), assign func(*HostConfig, *Device)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*HostConfig)
	for i := range nodes {
		if nodes[i].device_configs == nil {
			continue
		}
		fk := *nodes[i].device_configs
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(device.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "device_configs" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (hcq *HostConfigQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := hcq.querySpec()
	_spec.Node.Columns = hcq.ctx.Fields
	if len(hcq.ctx.Fields) > 0 {
		_spec.Unique = hcq.ctx.Unique != nil && *hcq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, hcq.driver, _spec)
}

func (hcq *HostConfigQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(hostconfig.Table, hostconfig.Columns, sqlgraph.NewFieldSpec(hostconfig.FieldID, field.TypeInt))
	_spec.From = hcq.sql
	if unique := hcq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if hcq.path != nil {
		_spec.Unique = true
	}
	if fields := hcq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, hostconfig.FieldID)
		for i := range fields {
			if fields[i] != hostconfig.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := hcq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := hcq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := hcq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := hcq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (hcq *HostConfigQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(hcq.driver.Dialect())
	t1 := builder.Table(hostconfig.Table)
	columns := hcq.ctx.Fields
	if len(columns) == 0 {
		columns = hostconfig.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if hcq.sql != nil {
		selector = hcq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if hcq.ctx.Unique != nil && *hcq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range hcq.predicates {
		p(selector)
	}
	for _, p := range hcq.order {
		p(selector)
	}
	if offset := hcq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := hcq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// HostConfigGroupBy is the group-by builder for HostConfig entities.
type HostConfigGroupBy struct {
	selector
	build *HostConfigQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (hcgb *HostConfigGroupBy) Aggregate(fns ...AggregateFunc) *HostConfigGroupBy {
	hcgb.fns = append(hcgb.fns, fns...)
	return hcgb
}

// Scan applies the selector query and scans the result into the given value.
func (hcgb *HostConfigGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, hcgb.build.ctx, "GroupBy")
	if err := hcgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*HostConfigQuery, *HostConfigGroupBy](ctx, hcgb.build, hcgb, hcgb.build.inters, v)
}

func (hcgb *HostConfigGroupBy) sqlScan(ctx context.Context, root *HostConfigQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(hcgb.fns))
	for _, fn := range hcgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*hcgb.flds)+len(hcgb.fns))
		for _, f := range *hcgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*hcgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := hcgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// HostConfigSelect is the builder for selecting fields of HostConfig entities.
type HostConfigSelect struct {
	*HostConfigQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (hcs *HostConfigSelect) Aggregate(fns ...AggregateFunc) *HostConfigSelect {
	hcs.fns = append(hcs.fns, fns...)
	return hcs
}

// Scan applies the selector query and scans the result into the given value.
func (hcs *HostConfigSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, hcs.ctx, "Select")
	if err := hcs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*HostConfigQuery, *HostConfigSelect](ctx, hcs.HostConfigQuery, hcs, hcs.inters, v)
}

func (hcs *HostConfigSelect) sqlScan(ctx context.Context, root *HostConfigQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(hcs.fns))
	for _, fn := range hcs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*hcs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := hcs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
