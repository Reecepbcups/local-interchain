package main

// This makes a fake T framework to make ICtest happy

// T is a subset of testing.TB,
// representing only the methods required by the reporter.
type T interface {
	Name() string
	Cleanup(func())

	Skip(...any)

	Parallel()

	Failed() bool
	Skipped() bool
}

// create methods for T

type FakeT struct {
	name string
}

func (t *FakeT) Name() string {
	return t.name
}

// Cleanup(func())
func (t *FakeT) Cleanup(func()) {
}

func (t *FakeT) Skip(...any) {
}

func (t *FakeT) Parallel() {
}

func (t *FakeT) Failed() bool {
	return false
}

func (t *FakeT) Skipped() bool {
	return false
}

// Helper
func (t *FakeT) Helper() {
}
