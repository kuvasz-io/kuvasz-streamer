package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/ext"
)

func ConvertPGColumnsToEnv(c map[string]PGColumn) map[string]any {
	env := make(map[string]any)

	for key, column := range c {
		var goVal any // Will be nil for NULL values by default

		// Use a type switch to handle different pgtype implementations from pgx/v5
		switch column.ColumnType {
		case "text":
			goVal = string("")
		case "int2":
			goVal = int32(0)
		case "int4":
			goVal = int64(0)
		case "int8":
			goVal = int8(0)
		case "uuid":
			goVal = string("")
		case "float8":
			goVal = float64(0)
		case "bytea":
			goVal = []byte{}
		}
		env[key] = goVal
	}
	log.Debug("Converted pgcolumns to env", "PGColumns", c, "env", env)
	return env
}

func prepareExpression(expression string, variables map[string]any) (cel.Program, error) { //nolint:ireturn // we have no choice here
	envOpts := []cel.EnvOption{
		cel.OptionalTypes(),
		cel.HomogeneousAggregateLiterals(),
		ext.Bindings(),
		ext.Encoders(),
		ext.Lists(),
		ext.Math(),
		ext.Regex(),
		ext.Sets(),
		ext.Strings(),
		ext.TwoVarComprehensions(),
	}
	log.Debug("prepareExpression", "expression", expression, "variables", variables)
	for key, value := range variables {
		// Infer the CEL type from the Go type of the value.
		// cel-go's common.Types provides helpers for common Go types.
		var celType *cel.Type
		switch value.(type) {
		case nil:
			celType = cel.NullType
		case int, int32, int64:
			celType = cel.IntType
		case float32, float64:
			celType = cel.DoubleType
		case string:
			celType = cel.StringType
		case bool:
			celType = cel.BoolType
		case time.Time:
			celType = cel.TimestampType
		case map[string]any:
			// For maps, we can declare it as a dynamic type (cel.DynType)
			// or define a more specific map type if the structure is known.
			celType = cel.DynType // Or cel.MapType(cel.StringType, cel.DynType) for string keys
		case []any:
			// For slices, declare as a dynamic list type.
			celType = cel.ListType(cel.DynType) // Or cel.ListType(cel.DynType)
		case []uint8:
			// For slices, declare as a dynamic list type.
			celType = cel.BytesType // Or cel.ListType(cel.DynType)
		default:
			celType = cel.DynType
			log.Debug("Warning: Could not infer specific CEL type, using cel.DynType", "key", key, "type", reflect.TypeOf(value))
		}
		if key == "type" {
			key = "_type"
		}
		envOpts = append(envOpts, cel.Variable(key, celType))
	}
	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		return nil, fmt.Errorf("can't create env: %w", err)
	}
	ast, iss := env.Parse(expression)
	if iss.Err() != nil {
		return nil, fmt.Errorf("can't parse expression: %s: %w", expression, iss.Err())
	}
	checked, iss := env.Check(ast)
	if iss.Err() != nil {
		return nil, fmt.Errorf("check ast failed for expression: %s: %w", expression, iss.Err())
	}
	program, err := env.Program(checked)
	if err != nil {
		return nil, fmt.Errorf("can't compile expression: %s: %w", expression, err)
	}
	return program, nil
}

func evalExpression(program cel.Program, variables map[string]any) (any, error) {
	output, _, err := program.Eval(variables)
	// change variable types CEL -> Go as CEL type causes stramge formatting of timestamps.
	// others may be needed
	switch v := output.(type) {
	case types.Timestamp:
		return v.Time, nil
	case types.Null:
		return nil, nil //nolint:nilnil // this is a false alert
	}
	if err != nil {
		return output, fmt.Errorf("cannot evaluate expression: %w", err)
	}
	return output, nil
}
