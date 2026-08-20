package domain

import (
	"context"
	"sort"
)

// ScheduleOptimizer is a domain service because selecting a schedule spans
// multiple booking entities and is defined entirely by business rules.
type ScheduleOptimizer struct{}

func NewScheduleOptimizer() ScheduleOptimizer { return ScheduleOptimizer{} }

type scheduleState struct {
	occupiedNights int64
	profitCents    ProfitCents
	earliestInput  int
	take           bool
	previous       int
}

type scheduleCandidate struct {
	booking BookingRequest
	input   int
}

func (ScheduleOptimizer) Optimize(ctx context.Context, batch BookingBatch) (Schedule, error) {
	if err := ctx.Err(); err != nil {
		return Schedule{}, err
	}
	if err := batch.Validate(); err != nil {
		return Schedule{}, err
	}

	bookings := batch.Bookings()
	candidates := make([]scheduleCandidate, len(bookings))
	for index, booking := range bookings {
		candidates[index] = scheduleCandidate{booking: booking, input: booking.InputIndex()}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := candidates[i].booking, candidates[j].booking
		if left.CheckOut().Equal(right.CheckOut()) {
			if left.CheckIn().Equal(right.CheckIn()) {
				return left.InputIndex() < right.InputIndex()
			}
			return left.CheckIn().Before(right.CheckIn())
		}
		return left.CheckOut().Before(right.CheckOut())
	})

	predecessors := findPredecessors(candidates)
	states := make([]scheduleState, len(candidates)+1)
	states[0].earliestInput = len(bookings)

	for index, item := range candidates {
		if index%256 == 0 {
			if err := ctx.Err(); err != nil {
				return Schedule{}, err
			}
		}

		previousStateIndex := predecessors[index] + 1
		include := scheduleState{
			occupiedNights: states[previousStateIndex].occupiedNights + int64(item.booking.Nights().Int32()),
			profitCents:    states[previousStateIndex].profitCents + item.booking.Profit(),
			earliestInput:  min(states[previousStateIndex].earliestInput, item.input),
			take:           true,
			previous:       previousStateIndex,
		}
		exclude := states[index]
		if betterSchedule(include, exclude) {
			states[index+1] = include
		} else {
			states[index+1] = scheduleState{
				occupiedNights: exclude.occupiedNights,
				profitCents:    exclude.profitCents,
				earliestInput:  exclude.earliestInput,
				take:           false,
				previous:       index,
			}
		}
	}

	selected := reconstructSchedule(candidates, states)
	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].CheckIn().Equal(selected[j].CheckIn()) {
			return selected[i].InputIndex() < selected[j].InputIndex()
		}
		return selected[i].CheckIn().Before(selected[j].CheckIn())
	})
	return NewSchedule(selected)
}

func findPredecessors(candidates []scheduleCandidate) []int {
	predecessors := make([]int, len(candidates))
	for index, current := range candidates {
		predecessors[index] = sort.Search(index, func(candidateIndex int) bool {
			return current.booking.CheckIn().Before(candidates[candidateIndex].booking.CheckOut())
		}) - 1
	}
	return predecessors
}

func betterSchedule(left, right scheduleState) bool {
	if left.occupiedNights != right.occupiedNights {
		return left.occupiedNights > right.occupiedNights
	}
	if left.profitCents != right.profitCents {
		return left.profitCents > right.profitCents
	}
	return left.earliestInput < right.earliestInput
}

func reconstructSchedule(candidates []scheduleCandidate, states []scheduleState) []BookingRequest {
	selected := make([]BookingRequest, 0)
	for stateIndex := len(states) - 1; stateIndex > 0; {
		state := states[stateIndex]
		if state.take {
			selected = append(selected, candidates[stateIndex-1].booking)
		}
		stateIndex = state.previous
	}
	for left, right := 0, len(selected)-1; left < right; left, right = left+1, right-1 {
		selected[left], selected[right] = selected[right], selected[left]
	}
	return selected
}
