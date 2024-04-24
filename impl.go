package promises

type impl[T any] struct {
	value T
	err   error
	done  chan struct{}
}

func (p *impl[T]) Wait() (T, error) {
	<-p.done
	return p.value, p.err
}

func (p *impl[T]) Done() <-chan struct{} {
	return p.done
}

func (p *impl[T]) resolve(value T) {
	select {
	case <-p.done:
		break
	default:
		p.value = value
		close(p.done)
	}
}

func (p *impl[T]) reject(err error) {
	select {
	case <-p.done:
		break
	default:
		p.err = err
		close(p.done)
	}
}
