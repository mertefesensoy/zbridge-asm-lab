# zbridge-asm-lab

Lab notebook for the Go assembly to z/OS services bridge project.
Each subdirectory is one self-contained exercise on the path toward
calling MVS macros from Go assembly without C, following the pattern
established by `ibmruntimes/go-recordio`.

## Exercises

- `add/` · toolchain checkpoint: prove Plan 9 assembly compiles, links,
  tests, and benchmarks on amd64.
- `ebcdic/` · table-driven byte translation between ISO-8859-1 and
  EBCDIC IBM-1047, the conversion primitive every MVS macro call
  needs at its parameter list boundary. Includes an s390x stub
  documenting the TR-instruction collapse.
