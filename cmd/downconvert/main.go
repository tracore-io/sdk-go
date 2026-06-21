// Command downconvert performs a minimal OpenAPI 3.1 -> 3.0 rewrite so oapi-codegen
// (which does not yet support 3.1, see oapi-codegen/oapi-codegen#373) can consume
// the Tracore spec. Two 3.1 constructs appear in the spec, both expressing
// nullability the JSON-Schema way:
//
//	type: [T, "null"]                      ->  type: T   + nullable: true
//	oneOf: [ {$ref: X}, {type: "null"} ]   ->  $ref: X   + nullable: true  (null member dropped)
//
// It also pins the top-level openapi version to 3.0.3. The rewrite is deliberately
// narrow: it panics on any nullable union it does not recognise, so a future spec
// change cannot silently corrupt codegen. The output is a build artifact
// (gitignored), regenerated on every `go generate`; it is never committed.
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// isNullSchema reports whether a schema node means exactly "the null type" — the
// member dropped from a nullable oneOf/anyOf. Matches both `type: "null"` (scalar)
// and `type: ["null"]` (single-element union).
func isNullSchema(m map[string]interface{}) bool {
	if len(m) != 1 {
		return false
	}
	switch t := m["type"].(type) {
	case string:
		return t == "null"
	case []interface{}:
		return len(t) == 1 && t[0] == "null"
	}
	return false
}

// collapseNullableUnion rewrites `oneOf`/`anyOf` that carry a null member into a
// single schema + `nullable: true`, in place on parent.
func collapseNullableUnion(parent map[string]interface{}, key string) {
	raw, ok := parent[key].([]interface{})
	if !ok {
		return
	}
	var kept []interface{}
	hadNull := false
	for _, e := range raw {
		if m, ok := e.(map[string]interface{}); ok && isNullSchema(m) {
			hadNull = true
			continue
		}
		kept = append(kept, e)
	}
	if !hadNull {
		return
	}
	if len(kept) != 1 {
		panic(fmt.Sprintf("unsupported nullable %s with %d non-null members", key, len(kept)))
	}
	only, ok := kept[0].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("unsupported nullable %s member type %T", key, kept[0]))
	}
	delete(parent, key)
	for k, v := range only {
		parent[k] = v
	}
	parent["nullable"] = true
}

func walk(n interface{}) {
	switch v := n.(type) {
	case map[string]interface{}:
		// type: [T, "null"] -> type: T + nullable: true
		if t, ok := v["type"].([]interface{}); ok {
			var nonNull []interface{}
			hasNull := false
			for _, e := range t {
				if e == "null" {
					hasNull = true
				} else {
					nonNull = append(nonNull, e)
				}
			}
			if !hasNull || len(nonNull) != 1 {
				panic(fmt.Sprintf("unsupported 3.1 type union: %v", t))
			}
			v["type"] = nonNull[0]
			v["nullable"] = true
		}
		collapseNullableUnion(v, "oneOf")
		collapseNullableUnion(v, "anyOf")
		for _, e := range v {
			walk(e)
		}
	case []interface{}:
		for _, e := range v {
			walk(e)
		}
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: downconvert <in.yaml> <out.yaml>")
		os.Exit(2)
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		panic(err)
	}
	doc["openapi"] = "3.0.3"
	walk(doc)
	out, err := yaml.Marshal(doc)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[2], out, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("down-converted %s -> %s\n", os.Args[1], os.Args[2])
}
