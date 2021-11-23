package layout

import (
	"fmt"
	"testing"
	"time"

	"github.com/awesome-gocui/gocui"
)

type size struct {
	x0, y0, x1, y1 int
}

func (s *size) String() string {
	return fmt.Sprintf("(%d,%d)-(%d,%d)", s.x0, s.y0, s.x1, s.y1)
}

type sample struct {
	x, y int
	name string
}

type testCase struct {
	desc             string
	layout           *layoutLevel
	size             size
	wantOverlap      map[string]size
	wantNoOverlap    map[string]size
	ignore           []string
	samplesNoOverlap []sample
	samplesOverlap   []sample
	wantErr          bool
}

func (t testCase) wants(overlap bool) map[string]size {
	if overlap {
		return t.wantOverlap
	}
	return t.wantNoOverlap
}

func (t testCase) samples(overlap bool) []sample {
	if overlap {
		return t.samplesOverlap
	}
	return t.samplesNoOverlap
}

var tests = []testCase{
	{
		desc: "simple",
		layout: &layoutLevel{
			direction: LayoutHorizontal,
			items: []*layoutItem{
				{
					ratio: 1,
					name:  "test",
				},
			},
		},
		wantNoOverlap: map[string]size{
			"test": {0, 0, 79, 24},
		},
		wantOverlap: map[string]size{
			"test": {0, 0, 79, 24},
		},
	},
	{
		desc: "column",
		layout: NewLevel(
			LayoutVertical,
			NewRatioItem(1, "test1"),
			NewRatioItem(1, "test2"),
			NewRatioItem(1, "test3"),
		),
		wantNoOverlap: map[string]size{
			"test1": {0, 0, 79, 7},
			"test2": {0, 8, 79, 15},
			"test3": {0, 16, 79, 24},
		},
		wantOverlap: map[string]size{
			"test1": {0, 0, 79, 8},
			"test2": {0, 8, 79, 16},
			"test3": {0, 16, 79, 24},
		},
	},
	{
		desc: "row 2:1",
		layout: NewLevel(
			LayoutHorizontal,
			NewRatioItem(2, "test1"),
			NewRatioItem(1, "test2"),
		),
		wantNoOverlap: map[string]size{
			"test1": {0, 0, 51, 24},
			"test2": {52, 0, 79, 24},
		},
		wantOverlap: map[string]size{
			"test1": {0, 0, 52, 24},
			"test2": {52, 0, 79, 24},
		},
	},
	{
		desc: "grid{11, 12; 21, 22, 23}",
		layout: NewLevel(LayoutHorizontal,
			NewRatioItem(1, "col1", WithInner(
				NewLevel(LayoutVertical,
					NewRatioItem(1, "test11"),
					NewRatioItem(1, "test12"),
				))),
			NewRatioItem(1, "col2", WithInner(
				NewLevel(LayoutVertical,
					NewRatioItem(1, "test21"),
					NewRatioItem(1, "test22"),
					NewRatioItem(1, "test23"),
				))),
		),
		wantNoOverlap: map[string]size{
			"test11": {0, 0, 39, 11},
			"test12": {0, 12, 39, 24},
			"test21": {40, 0, 79, 7},
			"test22": {40, 8, 79, 15},
			"test23": {40, 16, 79, 24},
		},
		wantOverlap: map[string]size{
			"test11": {0, 0, 40, 12},
			"test12": {0, 12, 40, 24},
			"test21": {40, 0, 79, 8},
			"test22": {40, 8, 79, 16},
			"test23": {40, 16, 79, 24},
		},
	},
	{
		desc: "grid{11(h), 12(h); 21, 22, 23}",
		layout: NewLevel(LayoutHorizontal,
			NewRatioItem(1, "col1", WithInner(
				NewLevel(LayoutVertical,
					NewRatioItem(1, "test11", Hidden()),
					NewRatioItem(1, "test12", Hidden()),
				))),
			NewRatioItem(1, "col2", WithInner(
				NewLevel(LayoutVertical,
					NewRatioItem(1, "test21"),
					NewRatioItem(1, "test22"),
					NewRatioItem(1, "test23"),
				))),
		),
		wantNoOverlap: map[string]size{
			"test21": {0, 0, 79, 7},
			"test22": {0, 8, 79, 15},
			"test23": {0, 16, 79, 24},
		},
		wantOverlap: map[string]size{
			"test21": {0, 0, 79, 8},
			"test22": {0, 8, 79, 16},
			"test23": {0, 16, 79, 24},
		},
		ignore: []string{
			"test11", "test12",
		},
	},
	{
		desc: "grid{11, 12(h); 21, 22, 23(h)}",
		layout: NewLevel(LayoutHorizontal,
			NewRatioItem(1, "col1", WithInner(
				NewLevel(LayoutVertical,
					NewRatioItem(1, "test11"),
					NewRatioItem(1, "test12", Hidden()),
				))),
			NewRatioItem(1, "col2", WithInner(
				NewLevel(LayoutVertical,
					NewRatioItem(1, "test21"),
					NewRatioItem(1, "test22"),
					NewRatioItem(1, "test23", Hidden()),
				))),
		),
		wantNoOverlap: map[string]size{
			"test11": {0, 0, 39, 24},
			"test21": {40, 0, 79, 11},
			"test22": {40, 12, 79, 24},
		},
		wantOverlap: map[string]size{
			"test11": {0, 0, 40, 24},
			"test21": {40, 0, 79, 12},
			"test22": {40, 12, 79, 24},
		},
		ignore: []string{
			"test12", "test23",
		},
	},
	{
		desc: "fixed",
		layout: NewLevel(
			LayoutHorizontal,
			NewFixedItem(10, "test1"),
			NewRatioItem(1, "test2"),
			NewRatioItem(4, "test3"),
		),
		wantNoOverlap: map[string]size{
			"test1": {0, 0, 9, 24},
			"test2": {10, 0, 23, 24},
			"test3": {24, 0, 79, 24},
		},
		wantOverlap: map[string]size{
			"test1": {0, 0, 10, 24},
			"test2": {10, 0, 24, 24},
			"test3": {24, 0, 79, 24},
		},
	},
	{
		desc: "fixed w/hidden fixed",
		layout: NewLevel(
			LayoutHorizontal,
			NewFixedItem(10, "test1", Hidden()),
			NewRatioItem(1, "test2"),
			NewRatioItem(4, "test3"),
		),
		wantNoOverlap: map[string]size{
			"test1": {0, 0, 79, 24},
			"test2": {0, 0, 15, 24},
			"test3": {16, 0, 79, 24},
		},
		wantOverlap: map[string]size{
			"test1": {0, 0, 79, 24},
			"test2": {0, 0, 16, 24},
			"test3": {16, 0, 79, 24},
		},
		samplesNoOverlap: []sample{
			{1, 1, "test2"},
			{14, 15, "test2"},
			{15, 15, "test1"},
			{16, 16, "test1"},
			{17, 16, "test3"},
			{78, 23, "test3"},
		},
		samplesOverlap: []sample{
			{1, 1, "test2"},
			{14, 15, "test2"},
			{15, 15, "test2"},
			{16, 16, "test1"},
			{17, 16, "test3"},
			{78, 23, "test3"},
		},
	},
	{
		desc: "fixed w/hidden not fixed",
		layout: NewLevel(
			LayoutHorizontal,
			NewFixedItem(10, "test1"),
			NewRatioItem(1, "test2", Hidden()),
			NewRatioItem(4, "test3"),
		),
		wantNoOverlap: map[string]size{
			"test1": {0, 0, 9, 24},
			"test2": {0, 0, 79, 24},
			"test3": {10, 0, 79, 24},
		},
		wantOverlap: map[string]size{
			"test1": {0, 0, 10, 24},
			"test2": {0, 0, 79, 24},
			"test3": {10, 0, 79, 24},
		},
		samplesNoOverlap: []sample{
			{1, 1, "test1"},
			{9, 15, "test2"},
			{10, 15, "test2"},
			{11, 16, "test3"},
			{78, 23, "test3"},
		},
		samplesOverlap: []sample{
			{1, 1, "test1"},
			{9, 15, "test1"},
			{10, 15, "test2"},
			{11, 16, "test3"},
			{78, 23, "test3"},
		},
	},
}

func TestLayoutNoOverlap(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, false)
	if err != nil {
		t.Fatalf("Can't create gui: %v", err)
	}
	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	runTests(t, g, tests)
	cleanup()
}

func TestLayoutWithOverlap(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("Can't create overlapping gui: %v", err)
	}
	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	runTests(t, g, tests)
	cleanup()
}

func runTests(t *testing.T, g *gocui.Gui, tests []testCase) {
	for _, tc := range tests {
		o := "without"
		if g.SupportOverlaps {
			o = "with"
		}
		t.Run(fmt.Sprintf("%s (%s overlapping)", tc.desc, o), func(t *testing.T) {
			wants := tc.wants(g.SupportOverlaps)
			samples := tc.samples(g.SupportOverlaps)
			g.SetManager(tc.layout)

			<-time.After(50 * time.Millisecond)
			found := make(map[string]bool)
			ignore := make(map[string]bool)
			for _, i := range tc.ignore {
				ignore[i] = true
			}
			for _, v := range g.Views() {
				name := v.Name()
				if _, ig := ignore[name]; ig {
					continue
				}
				found[name] = true
				if s, ok := wants[name]; !ok {
					x0, y0, x1, y1 := v.Dimensions()
					got := size{x0, y0, x1, y1}
					t.Errorf("Found unexpected view %q %s", name, got.String())
				} else {
					x0, y0, x1, y1 := v.Dimensions()
					got := size{x0, y0, x1, y1}
					if got != s {
						t.Errorf("Unexpected size for %q: got %s, want %s", name, got.String(), s.String())
					}
				}
			}

			for w := range wants {
				if !found[w] {
					t.Errorf("Expected view %q not found, views: %v", w, g.Views())
				}
			}

			for _, s := range samples {
				v, err := g.ViewByPosition(s.x, s.y)
				if err != nil {
					t.Errorf("Can't get view at %d,%d: %v", s.x, s.y, err)
					continue
				}
				if v.Name() != s.name {
					t.Errorf("Unexpected view at %d,%d: want %s, got %s", s.x, s.y, s.name, v.Name())
				}
			}
		})
	}
	<-time.After(50 * time.Millisecond)
}
