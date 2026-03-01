// Package det provides dynamic JSON unmarshaling with runtime interface type
// determination. It extends encoding/json by allowing callers to specify which
// concrete struct types should be used for interface fields, map values, and
// slice elements during deserialization.
package det
