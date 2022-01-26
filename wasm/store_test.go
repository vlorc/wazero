package wasm

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetModuleInstance(t *testing.T) {
	name := "test"

	// 'jit' and 'wazeroir' package cannot be used because of circular import.
	// Here we will use 'nil' instead. Should we have an Engine for testing?
	s := NewStore(nil)

	m1 := s.getModuleInstance(name)
	require.Equal(t, m1, s.ModuleInstances[name])
	require.NotNil(t, m1.Exports)

	m2 := s.getModuleInstance(name)
	require.Equal(t, m1, m2)
}

func TestBuildFunctionInstances_FunctionNames(t *testing.T) {
	name := "test"
	s := NewStore(nil)
	mi := s.getModuleInstance(name)

	zero := Index(0)
	nopCode := &Code{nil, []byte{OpcodeEnd}}
	m := &Module{
		TypeSection:     []*FunctionType{{}},
		FunctionSection: []Index{zero, zero, zero, zero, zero},
		NameSection: &NameSection{
			FunctionNames: NameMap{
				{Index: Index(1), Name: "two"},
				{Index: Index(3), Name: "four"},
				{Index: Index(4), Name: "five"},
			},
		},
		CodeSection: []*Code{nopCode, nopCode, nopCode, nopCode, nopCode},
	}

	_, err := s.buildFunctionInstances(m, mi)
	require.NoError(t, err)

	var names []string
	for _, f := range mi.Functions {
		names = append(names, f.Name)
	}

	// We expect unknown for any functions missing data in the NameSection
	require.Equal(t, []string{"unknown", "two", "unknown", "four", "five"}, names)
}

func TestValidateAddrRange(t *testing.T) {
	name := "test"
	s := NewStore(nil)
	mi := s.getModuleInstance(name)

	m := &Module{
		MemorySection: []*LimitsType{{Min: 100}},
	}

	_, err := s.buildMemoryInstances(m, mi)
	require.NoError(t, err)

	require.True(t, mi.Memory.ValidateAddrRange(uint32(0), uint64(0)))
	require.True(t, mi.Memory.ValidateAddrRange(uint32(0), uint64(100*PageSize)))
	require.False(t, mi.Memory.ValidateAddrRange(uint32(0), uint64(100*PageSize+1)))
	require.False(t, mi.Memory.ValidateAddrRange(uint32(1), uint64(100*PageSize)))
	require.False(t, mi.Memory.ValidateAddrRange(uint32(100*PageSize), uint64(0)))
}

func TestPutUint32(t *testing.T) {
	name := "test"
	s := NewStore(nil)
	mi := s.getModuleInstance(name)

	m := &Module{
		MemorySection: []*LimitsType{{Min: 100}},
	}

	_, err := s.buildMemoryInstances(m, mi)
	require.NoError(t, err)

	maxUint32 := uint32(math.MaxUint32)
	asymmetryBitsVal := uint32(0xfffffffe)
	require.True(t, mi.Memory.PutUint32(uint32(0), asymmetryBitsVal))
	require.Equal(t, asymmetryBitsVal, binary.LittleEndian.Uint32(mi.Memory.Buffer[0:4]))
	require.True(t, mi.Memory.PutUint32(uint32(0), maxUint32))
	require.Equal(t, maxUint32, binary.LittleEndian.Uint32(mi.Memory.Buffer[0:4]))
	require.True(t, mi.Memory.PutUint32(uint32(100*PageSize-4), asymmetryBitsVal))
	require.Equal(t, asymmetryBitsVal, binary.LittleEndian.Uint32(mi.Memory.Buffer[100*PageSize-4:100*PageSize]))
	require.False(t, mi.Memory.PutUint32(uint32(100*PageSize-3), asymmetryBitsVal))
}
