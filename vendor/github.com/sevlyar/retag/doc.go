// Package retag provides an ability to change tags of structures' fields in runtime
// without copying of the data. It may be helpful in next cases:
//
//  - Automatic tags generation;
//  - Different views of the one data;
//  - Fixing of leaky abstractions with minimal boilerplate code
//    when application has layers of abstractions and model is
//    separated from storages and presentation layers.
//
// Features:
//  - No memory allocations (for cached types);
//  - Fast converting (lookup in table and pointer creation for cached types).
//
// The package requires go1.7+.
//
// The package is still experimental and subject to change. The package can be broken by a next release of go.
//
package retag
