package formatting

import (
	"math"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/migsalazar/labtop/internal/model"
)

func TestPercentageAndTemperature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "percentage", got: Percentage(model.Some(18.4)), want: "18%"},
		{name: "percentage rounds", got: Percentage(model.Some(18.6)), want: "19%"},
		{name: "missing percentage", got: Percentage(model.None[float64]()), want: "—"},
		{name: "nonfinite percentage", got: Percentage(model.Some(math.NaN())), want: "—"},
		{name: "temperature", got: Temperature(model.Some(50.6)), want: "51°C"},
		{name: "missing temperature", got: Temperature(model.None[float64]()), want: "—"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.got != test.want {
				t.Fatalf("got %q, want %q", test.got, test.want)
			}
		})
	}
}

func TestThroughput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value model.Optional[float64]
		want  string
	}{
		{value: model.Some(812.0), want: "812 B/s"},
		{value: model.Some(12.4 * 1024), want: "12.4 KiB/s"},
		{value: model.Some(3.1 * 1024 * 1024), want: "3.1 MiB/s"},
		{value: model.Some(0.0), want: "0 B/s"},
		{value: model.Some(-1.0), want: "—"},
		{value: model.None[float64](), want: "—"},
	}
	for _, test := range tests {
		if got := Throughput(test.value); got != test.want {
			t.Errorf("Throughput(%#v) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value model.Optional[time.Duration]
		want  string
	}{
		{value: model.Some(42 * time.Second), want: "42s"},
		{value: model.Some(18 * time.Minute), want: "18m"},
		{value: model.Some(3*time.Hour + 12*time.Minute), want: "3h 12m"},
		{value: model.Some(8*24*time.Hour + 14*time.Hour), want: "8d 14h"},
		{value: model.Some(time.Duration(0)), want: "0s"},
		{value: model.Some(-time.Second), want: "—"},
		{value: model.None[time.Duration](), want: "—"},
	}
	for _, test := range tests {
		if got := Duration(test.value); got != test.want {
			t.Errorf("Duration(%#v) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestProbeLatency(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		value model.Optional[time.Duration]
		want  string
	}{
		{value: model.Some(999 * time.Microsecond), want: "<1 ms"},
		{value: model.Some(12 * time.Millisecond), want: "12 ms"},
		{value: model.None[time.Duration](), want: "—"},
	} {
		if got := ProbeLatency(test.value); got != test.want {
			t.Errorf("ProbeLatency(%#v) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestLastSeen(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, test := range []struct {
		value model.Optional[time.Time]
		want  string
	}{
		{value: model.None[time.Time](), want: "never"},
		{value: model.Some(now), want: "now"},
		{value: model.Some(now.Add(time.Second)), want: "now"},
		{value: model.Some(now.Add(-42 * time.Second)), want: "42s"},
		{value: model.Some(now.Add(-4 * time.Minute)), want: "4m"},
		{value: model.Some(now.Add(-3 * time.Hour)), want: "3h"},
		{value: model.Some(now.Add(-2 * 24 * time.Hour)), want: "2d"},
	} {
		if got := LastSeen(test.value, now); got != test.want {
			t.Errorf("LastSeen(%#v) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestSparklinePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		samples []float64
		width   int
		want    string
	}{
		{name: "empty", width: 5, want: "—"},
		{name: "zero width", samples: []float64{50}, width: 0, want: ""},
		{name: "fixed scale", samples: []float64{0, 50, 100}, width: 3, want: "▁▅█"},
		{name: "unfilled", samples: []float64{0, 100}, width: 4, want: "  ▁█"},
		{name: "newest only", samples: []float64{0, 25, 50, 75, 100}, width: 3, want: "▅▆█"},
		{name: "clamped", samples: []float64{-1, 101}, width: 2, want: "▁█"},
		{name: "unavailable sample", samples: []float64{math.NaN()}, width: 1, want: " "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := SparklinePercent(test.samples, test.width)
			if got != test.want {
				t.Fatalf("SparklinePercent(%v, %d) = %q, want %q", test.samples, test.width, got, test.want)
			}
			if test.samples != nil && test.width > 0 && utf8.RuneCountInString(got) != test.width {
				t.Fatalf("rendered width = %d, want %d", utf8.RuneCountInString(got), test.width)
			}
		})
	}
}
