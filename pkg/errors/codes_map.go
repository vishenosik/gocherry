package errors

type ErrorsMap[Type any] struct {
	_map_     map[error]Type
	_default_ Type
}

func NewErrorsMap[Type any](_default_ Type, _map_ map[error]Type) *ErrorsMap[Type] {
	return &ErrorsMap[Type]{
		_default_: _default_,
		_map_:     _map_,
	}
}

func (em *ErrorsMap[Type]) Get(err error) Type {
	for er := range em._map_ {
		if Is(err, er) {
			return em._map_[er]
		}
	}
	return em._default_
}
