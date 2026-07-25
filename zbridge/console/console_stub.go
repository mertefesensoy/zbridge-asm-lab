//go:build !(zos && s390x)

package console

import "github.com/mertefesensoy/zbridge"

// issueWTO on any platform that is not z/OS on s390x.
//
// This returns a typed error rather than using the UNDEF-stub convention the
// lab exercises use. The distinction is deliberate and is recorded in ADR 0003
// §4: an UNDEF stub is correct for a teaching exercise, where a build failure is
// the lesson, and wrong for a library, which must compile and be importable on a
// developer's laptop so its pure-Go parts can be tested there.
//
// Same principle — fail loudly, never silently — different mechanism, because
// the consumer is different.
func issueWTO(wpl []byte) error {
	return zbridge.Unsupported("console.WTO", "SVC 35")
}
