package v1

import "fmt"

type ReentrantSchedule struct {
	// TODO: write regex validator to cover all the cases
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}

func (in *ReentrantSchedule) String() string {
	if in.End != "" {
		return fmt.Sprintf("Re-entrant Schedule [start: %s, end: %s]", in.Start, in.End)
	} else {
		return fmt.Sprintf("Schedule [%s]", in.Start)
	}
}
