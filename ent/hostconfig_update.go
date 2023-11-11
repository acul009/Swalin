// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"rahnit-rmm/ent/device"
	"rahnit-rmm/ent/hostconfig"
	"rahnit-rmm/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// HostConfigUpdate is the builder for updating HostConfig entities.
type HostConfigUpdate struct {
	config
	hooks    []Hook
	mutation *HostConfigMutation
}

// Where appends a list predicates to the HostConfigUpdate builder.
func (hcu *HostConfigUpdate) Where(ps ...predicate.HostConfig) *HostConfigUpdate {
	hcu.mutation.Where(ps...)
	return hcu
}

// SetConfig sets the "config" field.
func (hcu *HostConfigUpdate) SetConfig(b []byte) *HostConfigUpdate {
	hcu.mutation.SetConfig(b)
	return hcu
}

// SetType sets the "type" field.
func (hcu *HostConfigUpdate) SetType(s string) *HostConfigUpdate {
	hcu.mutation.SetType(s)
	return hcu
}

// SetDeviceID sets the "device" edge to the Device entity by ID.
func (hcu *HostConfigUpdate) SetDeviceID(id int) *HostConfigUpdate {
	hcu.mutation.SetDeviceID(id)
	return hcu
}

// SetDevice sets the "device" edge to the Device entity.
func (hcu *HostConfigUpdate) SetDevice(d *Device) *HostConfigUpdate {
	return hcu.SetDeviceID(d.ID)
}

// Mutation returns the HostConfigMutation object of the builder.
func (hcu *HostConfigUpdate) Mutation() *HostConfigMutation {
	return hcu.mutation
}

// ClearDevice clears the "device" edge to the Device entity.
func (hcu *HostConfigUpdate) ClearDevice() *HostConfigUpdate {
	hcu.mutation.ClearDevice()
	return hcu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (hcu *HostConfigUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, hcu.sqlSave, hcu.mutation, hcu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hcu *HostConfigUpdate) SaveX(ctx context.Context) int {
	affected, err := hcu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (hcu *HostConfigUpdate) Exec(ctx context.Context) error {
	_, err := hcu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hcu *HostConfigUpdate) ExecX(ctx context.Context) {
	if err := hcu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (hcu *HostConfigUpdate) check() error {
	if v, ok := hcu.mutation.Config(); ok {
		if err := hostconfig.ConfigValidator(v); err != nil {
			return &ValidationError{Name: "config", err: fmt.Errorf(`ent: validator failed for field "HostConfig.config": %w`, err)}
		}
	}
	if v, ok := hcu.mutation.GetType(); ok {
		if err := hostconfig.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "HostConfig.type": %w`, err)}
		}
	}
	if _, ok := hcu.mutation.DeviceID(); hcu.mutation.DeviceCleared() && !ok {
		return errors.New(`ent: clearing a required unique edge "HostConfig.device"`)
	}
	return nil
}

func (hcu *HostConfigUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := hcu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(hostconfig.Table, hostconfig.Columns, sqlgraph.NewFieldSpec(hostconfig.FieldID, field.TypeInt))
	if ps := hcu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := hcu.mutation.Config(); ok {
		_spec.SetField(hostconfig.FieldConfig, field.TypeBytes, value)
	}
	if value, ok := hcu.mutation.GetType(); ok {
		_spec.SetField(hostconfig.FieldType, field.TypeString, value)
	}
	if hcu.mutation.DeviceCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   hostconfig.DeviceTable,
			Columns: []string{hostconfig.DeviceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(device.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := hcu.mutation.DeviceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   hostconfig.DeviceTable,
			Columns: []string{hostconfig.DeviceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(device.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, hcu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{hostconfig.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	hcu.mutation.done = true
	return n, nil
}

// HostConfigUpdateOne is the builder for updating a single HostConfig entity.
type HostConfigUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *HostConfigMutation
}

// SetConfig sets the "config" field.
func (hcuo *HostConfigUpdateOne) SetConfig(b []byte) *HostConfigUpdateOne {
	hcuo.mutation.SetConfig(b)
	return hcuo
}

// SetType sets the "type" field.
func (hcuo *HostConfigUpdateOne) SetType(s string) *HostConfigUpdateOne {
	hcuo.mutation.SetType(s)
	return hcuo
}

// SetDeviceID sets the "device" edge to the Device entity by ID.
func (hcuo *HostConfigUpdateOne) SetDeviceID(id int) *HostConfigUpdateOne {
	hcuo.mutation.SetDeviceID(id)
	return hcuo
}

// SetDevice sets the "device" edge to the Device entity.
func (hcuo *HostConfigUpdateOne) SetDevice(d *Device) *HostConfigUpdateOne {
	return hcuo.SetDeviceID(d.ID)
}

// Mutation returns the HostConfigMutation object of the builder.
func (hcuo *HostConfigUpdateOne) Mutation() *HostConfigMutation {
	return hcuo.mutation
}

// ClearDevice clears the "device" edge to the Device entity.
func (hcuo *HostConfigUpdateOne) ClearDevice() *HostConfigUpdateOne {
	hcuo.mutation.ClearDevice()
	return hcuo
}

// Where appends a list predicates to the HostConfigUpdate builder.
func (hcuo *HostConfigUpdateOne) Where(ps ...predicate.HostConfig) *HostConfigUpdateOne {
	hcuo.mutation.Where(ps...)
	return hcuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (hcuo *HostConfigUpdateOne) Select(field string, fields ...string) *HostConfigUpdateOne {
	hcuo.fields = append([]string{field}, fields...)
	return hcuo
}

// Save executes the query and returns the updated HostConfig entity.
func (hcuo *HostConfigUpdateOne) Save(ctx context.Context) (*HostConfig, error) {
	return withHooks(ctx, hcuo.sqlSave, hcuo.mutation, hcuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hcuo *HostConfigUpdateOne) SaveX(ctx context.Context) *HostConfig {
	node, err := hcuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (hcuo *HostConfigUpdateOne) Exec(ctx context.Context) error {
	_, err := hcuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hcuo *HostConfigUpdateOne) ExecX(ctx context.Context) {
	if err := hcuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (hcuo *HostConfigUpdateOne) check() error {
	if v, ok := hcuo.mutation.Config(); ok {
		if err := hostconfig.ConfigValidator(v); err != nil {
			return &ValidationError{Name: "config", err: fmt.Errorf(`ent: validator failed for field "HostConfig.config": %w`, err)}
		}
	}
	if v, ok := hcuo.mutation.GetType(); ok {
		if err := hostconfig.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "HostConfig.type": %w`, err)}
		}
	}
	if _, ok := hcuo.mutation.DeviceID(); hcuo.mutation.DeviceCleared() && !ok {
		return errors.New(`ent: clearing a required unique edge "HostConfig.device"`)
	}
	return nil
}

func (hcuo *HostConfigUpdateOne) sqlSave(ctx context.Context) (_node *HostConfig, err error) {
	if err := hcuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(hostconfig.Table, hostconfig.Columns, sqlgraph.NewFieldSpec(hostconfig.FieldID, field.TypeInt))
	id, ok := hcuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "HostConfig.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := hcuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, hostconfig.FieldID)
		for _, f := range fields {
			if !hostconfig.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != hostconfig.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := hcuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := hcuo.mutation.Config(); ok {
		_spec.SetField(hostconfig.FieldConfig, field.TypeBytes, value)
	}
	if value, ok := hcuo.mutation.GetType(); ok {
		_spec.SetField(hostconfig.FieldType, field.TypeString, value)
	}
	if hcuo.mutation.DeviceCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   hostconfig.DeviceTable,
			Columns: []string{hostconfig.DeviceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(device.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := hcuo.mutation.DeviceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   hostconfig.DeviceTable,
			Columns: []string{hostconfig.DeviceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(device.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &HostConfig{config: hcuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, hcuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{hostconfig.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	hcuo.mutation.done = true
	return _node, nil
}
