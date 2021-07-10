package service

type CheckService interface {
	HealthCheck() bool
}

type CheckServiceimpl struct {
}

func (svc CheckServiceimpl) HealthCheck() bool {
	return true
}
