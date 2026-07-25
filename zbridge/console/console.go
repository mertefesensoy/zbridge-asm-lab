package console

import "github.com/mertefesensoy/zbridge"

// WTO writes msg to the z/OS operator console via SVC 35.
//
// This is the function the entire project exists to make possible.
//
// # Contract
//
//	Input:        msg, 1..MaxTextLen bytes, no control characters
//	Side effect:  one line appears on the operator console(s) selected by the
//	              route codes
//	Returns:      nil on success; otherwise *zbridge.Error
//
// # Do not read the return code without checking HasCode
//
// Whether a single-line WTO returns anything to map is an OPEN QUESTION on
// z/OS. On the ancestor system it does not: GC28-0683-2 p.210 states that no
// return codes are issued by the WTO service routine when MLWTO is not used, and
// R1 comes back holding the 24-bit message identification number instead. The
// documented 00-14 return-code table is MLWTO-only.
//
// zbridge.Error.HasCode exists for exactly this reason. A caller that reads Code
// without checking HasCode may be reading a register the service never set.
// Research brief 003 Q4 is the resolving question.
//
// # Console etiquette
//
// An operator console is a workplace, not a log file. Messages should be tagged
// identifiably (the project's own test messages are prefixed "ZBRIDGE TEST") and
// kept to the minimum volume that proves what is being proven.
func WTO(msg string, opts ...Option) error {
	if err := validate(msg); err != nil {
		return err
	}

	buf, err := EncodeWPL(msg, opts...)
	if err != nil {
		// Today this is always ErrLayoutUnverified, which is the honest answer:
		// the platform call cannot be attempted because the bytes to hand it
		// are not known. The ordering is deliberate — refuse before trapping,
		// never trap with a guess.
		return err
	}

	return issueWTO(buf)
}

// WTOR writes msg to the operator console and waits for a reply.
//
// WTOR is strictly harder than WTO: it needs an ECB to wait on and a reply
// buffer, which means it needs storage the service can reach, which is U3a. It
// is declared here because ADR 0003 commits to declaring the console family's
// surface, and it returns ErrLayoutUnverified for the same reason WTO does.
//
// The reply buffer length is fixed by the caller's slice.
func WTOR(msg string, reply []byte, opts ...Option) (n int, err error) {
	if err := validate(msg); err != nil {
		return 0, err
	}
	if len(reply) == 0 {
		return 0, &zbridge.Error{
			Op:      "console.WTOR",
			Service: "SVC 35",
			Err:     zbridge.ErrLayoutUnverified,
			Unknown: "U2",
		}
	}
	return 0, zbridge.Unverified("console.WTOR", "SVC 35", "U2")
}
