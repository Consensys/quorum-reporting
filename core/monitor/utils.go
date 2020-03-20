package monitor

func isClosed(ch <-chan uint64) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
