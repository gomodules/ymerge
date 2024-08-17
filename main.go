/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: ymerge a.yaml ... z.yaml")
		os.Exit(1)
	}

	cur, err := read(os.Args[1])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to read file %q: %v", os.Args[1], err)
		os.Exit(1)
	}
	for _, filename := range args[2:] {
		override, err := read(filename)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to read file %q: %v", filename, err)
			os.Exit(1)
		}
		cur = mergeMaps(cur, override)
	}
	data, err := yaml.Marshal(cur)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to marshal: %v", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintln(os.Stdout, string(data))
}

func read(filename string) (map[string]any, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func mergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		switch vt := v.(type) {
		case map[string]any:
			if bv, ok := out[k]; ok {
				if bvt, ok := bv.(map[string]any); ok {
					out[k] = mergeMaps(bvt, vt)
					continue
				}
			}
		case []any:
			if bv, ok := out[k]; ok {
				if bvt, ok := bv.([]any); ok {
					out[k] = mergeArray(bvt, vt)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func mergeArray(a, b []any) []any {
	if len(a) == 0 {
		return b
	} else if len(b) == 0 {
		return a
	}

	// https://go.dev/wiki/SliceTricks#extend-capacity
	len_a, len_b := len(a), len(b)
	if cap(a) < len_b {
		a = append(make([]any, 0, len_b), a...)
	}
	i := 0
	for ; i < len_a; i++ {
		at, aObj := a[i].(map[string]any)
		bt, bObj := b[i].(map[string]any)
		if aObj && bObj {
			a[i] = mergeMaps(at, bt)
		} else {
			a[i] = b[i]
		}
	}
	for ; i < len_b; i++ {
		a[i] = b[i]
	}

	return a
}
