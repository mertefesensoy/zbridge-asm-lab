package regs

// GetBP returns the value of the hardware base pointer register (BP), which Go
// maintains as the frame pointer on amd64. It links call frames together.
//
// This declaration is amd64-only, by filename constraint. s390x has no
// frame-pointer register; see the package doc in regs.go and the s390x
// counterparts GetLinkRegister and GetGReg in regs_s390x.go.
func GetBP() uintptr
