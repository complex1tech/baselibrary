package async

import "reflect"

// Any awaits and returns the index and the result of any routine or -1 when no more routines.
func Any[T any](
	stop <-chan struct{},
	routines ...Routine[T],
) (index int, result Result[T], err error) {
	if len(routines) == 0 {
		return -1, result, nil
	}

	// make stop case
	cases := make([]reflect.SelectCase, 0, len(routines)+1)
	stop_ := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(stop),
	}
	cases = append(cases, stop_)

	// make routine cases
	for _, routine := range routines {
		wait := routine.Wait()
		wait_ := reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(wait),
		}
		cases = append(cases, wait_)
	}

	// select case
	i, _, _ := reflect.Select(cases)
	if i == 0 {
		return 0, result, Stopped
	}

	// make result
	index = i - 1
	routine := routines[index]
	value, err := routine.Result()

	result = Result[T]{
		Err:   err,
		Value: value,
	}
	return index, result, nil
}

// All awaits and returns the results of all routines in order.
func All[T any](
	stop <-chan struct{},
	routines ...Routine[T],
) ([]Result[T], error) {
	results := make([]Result[T], 0, len(routines))

	for _, r := range routines {
		select {
		case <-r.Wait():
		case <-stop:
			return results, Stopped
		}

		result := NewResult(r.Result())
		results = append(results, result)
	}

	return results, nil
}

// Combine combines multiple routines into one routine and a result channel.
func Combine[T any](routines ...Routine[T]) (Routine[Void], <-chan Result[T]) {
	ch := make(chan Result[T])
	fn := func(stop <-chan struct{}) error {
		defer close(ch)
		defer StopAll(routines...)

		for len(routines) > 0 {
			// await any routine
			index, result, err := Any(stop, routines...)
			if err != nil {
				return err
			}

			// send result
			select {
			case ch <- result:
			case <-stop:
				return Stopped
			}

			// delete routine
			routines = append(routines[:index], routines[index+1:]...)
		}

		return nil
	}

	r := Run(fn)
	return r, ch
}

// StopAll requests all routines to stop.
func StopAll[T any](routines ...Routine[T]) {
	for _, r := range routines {
		r.Stop()
	}
}

// StopWaitAll requests all routines to stop and waits for their results.
func StopWaitAll[T any](routines ...Routine[T]) {
	StopAll(routines...)
	WaitAll(routines...)
}

// WaitAll waits for all routine results.
func WaitAll[T any](routines ...Routine[T]) {
	for _, r := range routines {
		<-r.Wait()
	}
}