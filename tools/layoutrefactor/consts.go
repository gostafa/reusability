// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

const (
	errFmtCommitRunner = "commit runner: %w"
	licenseLine1       = "// Gostafa 2026."
	licenseLine2       = "// SPDX-License-Identifier: Apache-2.0."

	emptyString         = ""
	newline             = "\n"
	doubleNewline       = "\n\n"
	dot                 = "."
	quoteChar           = `"`
	extTestSuffix       = "_ext_test.go"
	goSuffix            = ".go"
	testSuffix          = "_test.go"
	docGoName           = "doc.go"
	constsGoName        = "consts.go"
	typesGoName         = "types.go"
	varsGoName          = "vars.go"
	funcsGoName         = "funcs.go"
	mainTestName        = "main_test.go"
	mainPkgName         = "main"
	constantsTest       = "constants_test.go"
	benchFragment       = "bench"
	testdataSeg         = "/testdata/"
	selfToolPath        = "/tools/layoutrefactor"
	extNameSuffix       = "Ext"
	extPrefix           = "ext"
	goBinaryName        = "go"
	fmtShortWrite       = "%w: wrote %d of %d"
	fmtFormatMergedTest = "format merged tests: %w"
	fmtCommitWrites     = "commit writes: %w"
	dryRunRemove        = "  remove "

	countZero  = 0
	countOne   = 1
	countTwo   = 2
	countThree = 3
	cmpLess    = -1
	filePerm   = 0o644

	kindConst declKind = 10
	kindType  declKind = 11
	kindVar   declKind = 12
	kindFunc  declKind = 13

	constConflictSkip   constConflictAction = 20
	constConflictRename constConflictAction = 21

	testFileInternal testFileMode = 30
	testFileExternal testFileMode = 31
)
