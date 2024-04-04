package proc

import (
	"testing"
)

const (
	m1Frequency        = 3_200_000_000
	secondToNanosecond = 1_000_000_000

	m1PrimeExecutionTime      = 70.29
	m1SumsExecutionTime       = 1300.
	m1StringCopyExecutionTime = 3232.
)

func primeStats(t *testing.T, cycles int) {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1PrimeExecutionTime
	t.Logf("%.0f ns, %.1f slower\n", ns, slower)
}

func sumStats(t *testing.T, cycles int) {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1SumsExecutionTime
	t.Logf("%.0f ns, %.1f slower\n", ns, slower)
}

func sumStringCopy(t *testing.T, cycles int) {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1StringCopyExecutionTime
	t.Logf("%.0f ns, %.1f slower\n", ns, slower)
}
