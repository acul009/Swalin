// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"rahnit-rmm/ent/device"
	"rahnit-rmm/ent/hostconfig"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// HostConfig is the model entity for the HostConfig schema.
type HostConfig struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Config holds the value of the "config" field.
	Config []byte `json:"config,omitempty"`
	// Type holds the value of the "type" field.
	Type string `json:"type,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the HostConfigQuery when eager-loading is set.
	Edges          HostConfigEdges `json:"edges"`
	device_configs *int
	selectValues   sql.SelectValues
}

// HostConfigEdges holds the relations/edges for other nodes in the graph.
type HostConfigEdges struct {
	// Device holds the value of the device edge.
	Device *Device `json:"device,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// DeviceOrErr returns the Device value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e HostConfigEdges) DeviceOrErr() (*Device, error) {
	if e.loadedTypes[0] {
		if e.Device == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: device.Label}
		}
		return e.Device, nil
	}
	return nil, &NotLoadedError{edge: "device"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*HostConfig) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case hostconfig.FieldConfig:
			values[i] = new([]byte)
		case hostconfig.FieldID:
			values[i] = new(sql.NullInt64)
		case hostconfig.FieldType:
			values[i] = new(sql.NullString)
		case hostconfig.ForeignKeys[0]: // device_configs
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the HostConfig fields.
func (hc *HostConfig) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case hostconfig.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			hc.ID = int(value.Int64)
		case hostconfig.FieldConfig:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field config", values[i])
			} else if value != nil {
				hc.Config = *value
			}
		case hostconfig.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				hc.Type = value.String
			}
		case hostconfig.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field device_configs", value)
			} else if value.Valid {
				hc.device_configs = new(int)
				*hc.device_configs = int(value.Int64)
			}
		default:
			hc.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the HostConfig.
// This includes values selected through modifiers, order, etc.
func (hc *HostConfig) Value(name string) (ent.Value, error) {
	return hc.selectValues.Get(name)
}

// QueryDevice queries the "device" edge of the HostConfig entity.
func (hc *HostConfig) QueryDevice() *DeviceQuery {
	return NewHostConfigClient(hc.config).QueryDevice(hc)
}

// Update returns a builder for updating this HostConfig.
// Note that you need to call HostConfig.Unwrap() before calling this method if this HostConfig
// was returned from a transaction, and the transaction was committed or rolled back.
func (hc *HostConfig) Update() *HostConfigUpdateOne {
	return NewHostConfigClient(hc.config).UpdateOne(hc)
}

// Unwrap unwraps the HostConfig entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (hc *HostConfig) Unwrap() *HostConfig {
	_tx, ok := hc.config.driver.(*txDriver)
	if !ok {
		panic("ent: HostConfig is not a transactional entity")
	}
	hc.config.driver = _tx.drv
	return hc
}

// String implements the fmt.Stringer.
func (hc *HostConfig) String() string {
	var builder strings.Builder
	builder.WriteString("HostConfig(")
	builder.WriteString(fmt.Sprintf("id=%v, ", hc.ID))
	builder.WriteString("config=")
	builder.WriteString(fmt.Sprintf("%v", hc.Config))
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(hc.Type)
	builder.WriteByte(')')
	return builder.String()
}

// HostConfigs is a parsable slice of HostConfig.
type HostConfigs []*HostConfig
