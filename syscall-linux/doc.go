// Package syscalllinux demonstrates issuing Linux amd64 system calls directly
// from Go assembly. It is a rehearsal for the z/OS SVC path: load the service
// number and parameters into the ABI-defined registers, execute the trap
// instruction, and interpret the returned register value.
package syscalllinux
