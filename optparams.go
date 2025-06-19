package optparams

// Func should initialize some fields in the passed receiver.
type Func[T any] func(receiver *T) error
